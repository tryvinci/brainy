package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// WriteRawEvidence persists a source-faithful evidence row (Evidence Plane v2).
// Dedupe key includes tenant, namespace, subject, source_ref, and content hash
// so identical text for different subjects cannot collapse.
func (s *Store) WriteRawEvidence(ctx context.Context, tenantID, subjectID, namespace, sourceType, sourceRef, sessionID, actorRole, content string, occurredAt *time.Time, metadata map[string]any) (evidenceID string, err error) {
	content = strings.TrimSpace(content)
	if tenantID == "" || subjectID == "" || content == "" {
		return "", nil
	}
	if namespace == "" {
		namespace = "default"
	}
	if sourceType == "" {
		sourceType = "conversation"
	}
	if sourceRef == "" {
		sourceRef = "msg"
	}
	sum := sha256.Sum256([]byte(strings.Join([]string{tenantID, namespace, subjectID, sourceRef, content}, "\x00")))
	hash := hex.EncodeToString(sum[:])
	id := "ev_" + hash[:24]
	metaJSON, _ := json.Marshal(metadata)
	if metaJSON == nil {
		metaJSON = []byte("{}")
	}
	_, err = s.pool.Exec(ctx, `
INSERT INTO memory_evidence (
  id, tenant_id, namespace, subject_id, source_type, source_ref, session_id,
  content, content_hash, occurred_at, recorded_at, metadata, memory_id, actor_role
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,now(),$11::jsonb,NULL,$12)
ON CONFLICT (tenant_id, namespace, subject_id, content_hash, source_ref) DO UPDATE
SET session_id = COALESCE(EXCLUDED.session_id, memory_evidence.session_id),
    occurred_at = COALESCE(EXCLUDED.occurred_at, memory_evidence.occurred_at)
`, id, tenantID, namespace, subjectID, sourceType, sourceRef, sessionID, content, hash, occurredAt, string(metaJSON), actorRole)
	if err != nil {
		return "", fmt.Errorf("write raw evidence: %w", err)
	}
	return id, nil
}

// ShadowWriteEvidence is retained for transitional callers; prefers subject-safe dedupe.
// Prefer WriteRawEvidence for new ingest paths.
func (s *Store) ShadowWriteEvidence(ctx context.Context, tenantID, subjectID, sourceType, sourceRef, sessionID, content, memoryID string, occurredAt *time.Time, metadata map[string]any) error {
	content = strings.TrimSpace(content)
	if tenantID == "" || content == "" {
		return nil
	}
	if sourceRef == "" {
		sourceRef = memoryID
	}
	sum := sha256.Sum256([]byte(strings.Join([]string{tenantID, "default", subjectID, sourceRef, content}, "\x00")))
	hash := hex.EncodeToString(sum[:])
	id := "ev_" + hash[:24]
	metaJSON, _ := json.Marshal(metadata)
	if metaJSON == nil {
		metaJSON = []byte("{}")
	}
	_, err := s.pool.Exec(ctx, `
INSERT INTO memory_evidence (
  id, tenant_id, namespace, subject_id, source_type, source_ref, session_id,
  content, content_hash, occurred_at, recorded_at, metadata, memory_id
) VALUES ($1,$2,'default',$3,$4,$5,$6,$7,$8,$9,now(),$10::jsonb,$11)
ON CONFLICT (tenant_id, namespace, subject_id, content_hash, source_ref) DO NOTHING
`, id, tenantID, subjectID, sourceType, sourceRef, sessionID, content, hash, occurredAt, string(metaJSON), memoryID)
	return err
}

// ListEvidence returns raw evidence rows for a subject (oracle / provenance).
func (s *Store) ListEvidence(ctx context.Context, tenantID, subjectID string, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
SELECT id, content, source_type, source_ref, occurred_at, memory_id
FROM memory_evidence
WHERE tenant_id = $1 AND subject_id = $2 AND suppression_status = 'active'
ORDER BY occurred_at DESC NULLS LAST, recorded_at DESC
LIMIT $3
`, tenantID, subjectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0, limit)
	for rows.Next() {
		var id, content, sourceType, sourceRef string
		var occurredAt *time.Time
		var memoryID *string
		if err := rows.Scan(&id, &content, &sourceType, &sourceRef, &occurredAt, &memoryID); err != nil {
			return nil, err
		}
		row := map[string]any{
			"evidence_id": id,
			"content":     content,
			"source_type": sourceType,
			"source_ref":  sourceRef,
		}
		if occurredAt != nil {
			row["occurred_at"] = occurredAt.UTC().Format(time.RFC3339)
		}
		if memoryID != nil {
			row["memory_id"] = *memoryID
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// UpsertMemoryEvent stores a first-class event (Phase 3 minimal).
func (s *Store) UpsertMemoryEvent(ctx context.Context, tenantID, subjectID, eventID, eventType, title, description, memoryID, evidenceID string, startsAt *time.Time, confidence float64, participants []string) error {
	if tenantID == "" || eventID == "" || eventType == "" {
		return nil
	}
	if confidence <= 0 {
		confidence = 0.5
	}
	_, err := s.pool.Exec(ctx, `
INSERT INTO memory_events (
  id, tenant_id, namespace, subject_id, event_type, title, description,
  starts_at, confidence, recorded_at, evidence_id, memory_id
) VALUES ($1,$2,'default',$3,$4,$5,$6,$7,$8,now(),$9,$10)
ON CONFLICT (id) DO UPDATE
SET title = EXCLUDED.title,
    description = EXCLUDED.description,
    starts_at = COALESCE(EXCLUDED.starts_at, memory_events.starts_at),
    confidence = EXCLUDED.confidence,
    evidence_id = COALESCE(EXCLUDED.evidence_id, memory_events.evidence_id)
`, eventID, tenantID, subjectID, eventType, title, description, startsAt, confidence, evidenceID, memoryID)
	if err != nil {
		return err
	}
	for _, p := range participants {
		p = strings.ToLower(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		_, _ = s.pool.Exec(ctx, `
INSERT INTO memory_event_participants (event_id, entity_key, participant_role)
VALUES ($1, $2, 'participant')
ON CONFLICT DO NOTHING
`, eventID, p)
	}
	return nil
}

// UpsertCurrentState writes a rebuildable current-state projection row.
// Callers must ensure the winning assertion is temporally valid before calling.
func (s *Store) UpsertCurrentState(ctx context.Context, tenantID, subjectID, predicate, memoryID, value, policy string) error {
	predicate = strings.TrimSpace(predicate)
	value = strings.TrimSpace(value)
	if tenantID == "" || subjectID == "" || predicate == "" || memoryID == "" || value == "" {
		return nil
	}
	if policy == "" {
		policy = "TEMPORAL_STATE"
	}
	_, err := s.pool.Exec(ctx, `
INSERT INTO memory_current_state (
  tenant_id, namespace, subject_id, predicate,
  winning_memory_id, resolved_value, resolution_policy, resolved_at
) VALUES ($1,'default',$2,$3,$4,$5,$6,now())
ON CONFLICT (tenant_id, namespace, subject_id, predicate) DO UPDATE
SET winning_memory_id = EXCLUDED.winning_memory_id,
    resolved_value = EXCLUDED.resolved_value,
    resolution_policy = EXCLUDED.resolution_policy,
    resolved_at = EXCLUDED.resolved_at
`, tenantID, subjectID, predicate, memoryID, value, policy)
	return err
}

// GetCurrentState returns the projected current value for a predicate, if any.
func (s *Store) GetCurrentState(ctx context.Context, tenantID, subjectID, predicate string) (memoryID, value, policy string, ok bool, err error) {
	err = s.pool.QueryRow(ctx, `
SELECT winning_memory_id, resolved_value, resolution_policy
FROM memory_current_state
WHERE tenant_id = $1 AND namespace = 'default' AND subject_id = $2 AND predicate = $3
`, tenantID, subjectID, predicate).Scan(&memoryID, &value, &policy)
	if err != nil {
		return "", "", "", false, nil
	}
	return memoryID, value, policy, true, nil
}
