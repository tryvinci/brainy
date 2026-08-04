package memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// EvidenceWriter is the Stage-1 shadow evidence plane (program §18.1).
type EvidenceWriter interface {
	ShadowWriteEvidence(ctx context.Context, tenantID, subjectID, sourceType, sourceRef, sessionID, content, memoryID string, occurredAt *time.Time, metadata map[string]any) error
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

func (s *Service) persistEvidenceShadow(ctx context.Context, record MemoryRecord) {
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
		// Still update current-state for stateful predicates.
		if IsStatefulPredicate(pred) && val != "" {
			if cs, ok := s.store.(CurrentStateStore); ok {
				_ = cs.UpsertCurrentState(ctx, record.TenantID, record.SubjectID, pred, record.MemoryID, val, string(policy))
			}
		}
		return
	}
	sum := sha256.Sum256([]byte(record.TenantID + pred + val + record.MemoryID))
	eventID := "evt_" + hex.EncodeToString(sum[:12])
	parts := ExtractEntities(record.Content)
	_ = writer.UpsertMemoryEvent(ctx, record.TenantID, record.SubjectID, eventID, pred, val, record.Content, record.MemoryID, "", record.ObservedAt, record.Confidence, parts)
}
