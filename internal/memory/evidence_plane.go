package memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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

// CurrentStateStore is the rebuildable projection for stateful predicates.
type CurrentStateStore interface {
	UpsertCurrentState(ctx context.Context, tenantID, subjectID, predicate, memoryID, value, policy string) error
	GetCurrentState(ctx context.Context, tenantID, subjectID, predicate string) (memoryID, value, policy string, ok bool, err error)
}

// RankedSearcher returns lexical ranks alongside memories (real FTS when available).
type RankedSearcher interface {
	SearchMemoriesRanked(ctx context.Context, tenantID, subjectID string, patterns []string, limit int, includeSuperseded bool) ([]MemoryRecord, map[string]float64, error)
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
	for i, msg := range req.Messages {
		content := msg.Content
		if content == "" {
			continue
		}
		ref := fmt.Sprintf("msg:%d", i)
		id, err := writer.WriteRawEvidence(ctx, req.TenantID, req.SubjectID, "default", req.SourceType, ref, session, msg.Role, content, nil, req.Metadata)
		if err != nil {
			// Surface via explain later; do not silently drop without metric path.
			continue
		}
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
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
	_ = writer.UpsertMemoryEvent(ctx, record.TenantID, record.SubjectID, eventID, pred, val, record.Content, record.MemoryID, "", record.ObservedAt, record.Confidence, parts)
}

func (s *Service) projectCurrentStateIfApplicable(ctx context.Context, record MemoryRecord) {
	if record.Metadata == nil {
		return
	}
	pred, _ := record.Metadata["predicate"].(string)
	val, _ := record.Metadata["value_norm"].(string)
	if !IsStatefulPredicate(pred) || val == "" {
		return
	}
	cs, ok := s.store.(CurrentStateStore)
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
	_ = cs.UpsertCurrentState(ctx, record.TenantID, record.SubjectID, pred, record.MemoryID, val, string(PredicatePolicy(pred)))
}
