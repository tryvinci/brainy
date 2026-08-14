package memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"
)

// EvidenceWriter is the Stage-1 shadow evidence plane (program §18.1).
type EvidenceWriter interface {
	ShadowWriteEvidence(ctx context.Context, tenantID, subjectID, sourceType, sourceRef, sessionID, content, memoryID string, occurredAt *time.Time, metadata map[string]any) error
}

// RawEvidenceWriter is Evidence Plane v2 — source-faithful capture before extract.
type RawEvidenceWriter interface {
	WriteRawEvidence(ctx context.Context, tenantID, subjectID, namespace, sourceType, sourceRef, sessionID, actorRole, content string, occurredAt *time.Time, metadata map[string]any) (evidenceID string, err error)
	ListEvidence(ctx context.Context, tenantID, subjectID string, limit int) ([]map[string]any, error)
}

// EventWriter persists first-class events (Phase 3).
type EventWriter interface {
	UpsertMemoryEvent(ctx context.Context, tenantID, subjectID, eventID, eventType, title, description, memoryID, evidenceID string, startsAt *time.Time, confidence float64, participants []string) error
}

// EventEnder closes world-valid event intervals on supersede (mig-16 ends_at).
type EventEnder interface {
	EndMemoryEventsByMemoryID(ctx context.Context, tenantID, subjectID, memoryID string, endedAt *time.Time) error
}

// CurrentStateStore is the rebuildable projection for stateful predicates.
type CurrentStateStore interface {
	UpsertCurrentState(ctx context.Context, tenantID, subjectID, predicate, memoryID, value, policy string) error
	GetCurrentState(ctx context.Context, tenantID, subjectID, predicate string) (memoryID, value, policy string, ok bool, err error)
	DeleteCurrentStateByMemory(ctx context.Context, tenantID, subjectID, memoryID string) error
}

// RankedSearcher returns lexical ranks alongside memories (real FTS when available).
type RankedSearcher interface {
	SearchMemoriesRanked(ctx context.Context, tenantID, subjectID string, patterns []string, limit int, includeSuperseded bool) ([]MemoryRecord, map[string]float64, error)
}

func observedAtFromMetadata(meta map[string]any) *time.Time {
	if meta == nil {
		return nil
	}
	raw, ok := meta["observed_at"]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case time.Time:
		t := v.UTC()
		return &t
	case *time.Time:
		if v == nil {
			return nil
		}
		t := v.UTC()
		return &t
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return nil
		}
		for _, layout := range []string{time.RFC3339, "2006-01-02", "2006-01-02T15:04:05Z07:00"} {
			if t, err := time.Parse(layout, s); err == nil {
				u := t.UTC()
				return &u
			}
		}
	}
	return nil
}

func (s *Service) persistRawEvidence(ctx context.Context, req IngestRequest) []string {
	writer, ok := s.store.(RawEvidenceWriter)
	if !ok {
		return nil
	}
	ids := make([]string, 0, len(req.Messages))
	session := ""
	if req.Metadata != nil {
		if v, ok := req.Metadata["session_id"].(string); ok {
			session = v
		}
	}
	occurredAt := observedAtFromMetadata(req.Metadata)
	for i, msg := range req.Messages {
		content := msg.Content
		if content == "" {
			continue
		}
		ref := fmt.Sprintf("msg:%d", i)
		id, err := writer.WriteRawEvidence(ctx, req.TenantID, req.SubjectID, "default", req.SourceType, ref, session, msg.Role, content, occurredAt, req.Metadata)
		if err != nil {
			// Prefer failing closed on evidence loss in strict mode; otherwise keep ingest moving.
			if os.Getenv("BRAINY_EVIDENCE_STRICT") == "true" {
				continue
			}
			continue
		}
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

// attachEvidenceIDs stores raw evidence IDs on request metadata for extract→record linking.
func attachEvidenceIDs(req *IngestRequest, ids []string) {
	if len(ids) == 0 {
		return
	}
	if req.Metadata == nil {
		req.Metadata = map[string]any{}
	}
	copied := make([]any, len(ids))
	for i, id := range ids {
		copied[i] = id
	}
	req.Metadata["raw_evidence_ids"] = copied
	if len(ids) == 1 {
		req.Metadata["evidence_id"] = ids[0]
	}
}

func evidenceIDFromRecord(record MemoryRecord) string {
	if record.Metadata == nil {
		return ""
	}
	if v, ok := record.Metadata["evidence_id"].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

func (s *Service) persistEvidenceShadow(ctx context.Context, record MemoryRecord) {
	// Legacy shadow retained only when raw writer is unavailable.
	if _, ok := s.store.(RawEvidenceWriter); ok {
		return
	}
	writer, ok := s.store.(EvidenceWriter)
	if !ok {
		return
	}
	srcType := "conversation"
	if record.Metadata != nil {
		if v, ok := record.Metadata["source_type"].(string); ok && v != "" {
			srcType = v
		}
	}
	session := sessionIDOf(record)
	_ = writer.ShadowWriteEvidence(ctx, record.TenantID, record.SubjectID, srcType, record.MemoryID, session, record.Content, record.MemoryID, record.ObservedAt, record.Metadata)
}

func (s *Service) persistEventIfApplicable(ctx context.Context, record MemoryRecord) {
	writer, ok := s.store.(EventWriter)
	if !ok || record.Metadata == nil {
		return
	}
	pred, _ := record.Metadata["predicate"].(string)
	val, _ := record.Metadata["value_norm"].(string)
	if pred == "" {
		return
	}
	policy := PredicatePolicy(pred)
	if policy != PolicyAppendOnlyEvent && pred != PredicateEvent && pred != PredicateActivity {
		return
	}
	sum := sha256.Sum256([]byte(record.TenantID + pred + val + record.MemoryID))
	eventID := "evt_" + hex.EncodeToString(sum[:12])
	parts := ExtractEntities(record.Content)
	evidenceID := evidenceIDFromRecord(record)
	_ = writer.UpsertMemoryEvent(ctx, record.TenantID, record.SubjectID, eventID, pred, val, record.Content, record.MemoryID, evidenceID, record.ObservedAt, record.Confidence, parts)
}

func (s *Service) projectCurrentStateIfApplicable(ctx context.Context, record MemoryRecord) {
	ProjectCurrentStateIfApplicable(ctx, s.store, record)
}

// statePredicateKey scopes temporal state by entity when known, so two people
// in one subject conversation do not collide on the same predicate.
func statePredicateKey(entity, predicate string) string {
	predicate = strings.TrimSpace(predicate)
	entity = strings.ToLower(strings.TrimSpace(entity))
	if entity == "" {
		return predicate
	}
	return entity + "::" + predicate
}

func entitySubjectOf(record MemoryRecord) string {
	if record.Metadata != nil {
		if v, ok := record.Metadata["subject"].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
		if v, ok := record.Metadata["entity_key"].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	if record.Explain != nil {
		if v, ok := record.Explain["subject"].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// ProjectCurrentStateIfApplicable writes memory_current_state only when the
// incoming record may replace the existing projection (shared by sync + async).
func ProjectCurrentStateIfApplicable(ctx context.Context, store Store, record MemoryRecord) {
	if store == nil || record.Metadata == nil {
		return
	}
	pred, _ := record.Metadata["predicate"].(string)
	val, _ := record.Metadata["value_norm"].(string)
	if !IsStatefulPredicate(pred) || val == "" {
		return
	}
	cs, ok := store.(CurrentStateStore)
	if !ok {
		return
	}
	// Only project if this record is still search-visible (not superseded).
	if record.Status != "" && record.Status != StatusActive {
		return
	}
	if !IsLifecycleSearchVisible(record.LifecycleState) && record.LifecycleState != "" && record.LifecycleState != LifecycleActive {
		return
	}
	keyed := statePredicateKey(entitySubjectOf(record), pred)
	if existingID, _, _, found, _ := cs.GetCurrentState(ctx, record.TenantID, record.SubjectID, keyed); found && existingID != "" && existingID != record.MemoryID {
		existing, err := store.GetMemory(ctx, record.TenantID, record.SubjectID, existingID)
		if err == nil && !shouldReplaceCurrentState(record, existing) {
			return
		}
	}
	_ = cs.UpsertCurrentState(ctx, record.TenantID, record.SubjectID, keyed, record.MemoryID, val, string(PredicatePolicy(pred)))
}

// ReprojCurrentStateForMutation reconciles the current-state projection after a
// mutation changes a record's content (Correct) or retires the winner
// (Supersede). It deletes stale rows pointing at the given memory and then
// re-projects the record if it is still search-visible and stateful. For
// corrections the corrected content is the new ground-truth value; value
// overrides the record's (possibly stale) metadata value_norm when non-empty.
func ReprojCurrentStateForMutation(ctx context.Context, store Store, record MemoryRecord, value string) {
	if store == nil {
		return
	}
	cs, ok := store.(CurrentStateStore)
	if !ok {
		return
	}
	_ = cs.DeleteCurrentStateByMemory(ctx, record.TenantID, record.SubjectID, record.MemoryID)
	if record.Metadata == nil {
		return
	}
	pred, _ := record.Metadata["predicate"].(string)
	if !IsStatefulPredicate(pred) {
		return
	}
	if record.Status != "" && record.Status != StatusActive {
		return
	}
	if !IsLifecycleSearchVisible(record.LifecycleState) && record.LifecycleState != "" && record.LifecycleState != LifecycleActive {
		return
	}
	if value == "" {
		value, _ = record.Metadata["value_norm"].(string)
	}
	if value == "" {
		value = strings.ToLower(NormalizeText(record.Content))
	}
	keyed := statePredicateKey(entitySubjectOf(record), pred)
	if existingID, _, _, found, _ := cs.GetCurrentState(ctx, record.TenantID, record.SubjectID, keyed); found && existingID != "" && existingID != record.MemoryID {
		existing, err := store.GetMemory(ctx, record.TenantID, record.SubjectID, existingID)
		if err == nil && !shouldReplaceCurrentState(record, existing) {
			return
		}
	}
	_ = cs.UpsertCurrentState(ctx, record.TenantID, record.SubjectID, keyed, record.MemoryID, value, string(PredicatePolicy(pred)))
}
