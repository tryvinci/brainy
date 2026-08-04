package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

// ShadowWriteEvidence persists an immutable evidence row alongside memory_records
// (program Stage 1 transitional migration). Failures are ignored by callers.
func (s *Store) ShadowWriteEvidence(ctx context.Context, tenantID, subjectID, sourceType, sourceRef, sessionID, content, memoryID string, occurredAt *time.Time, metadata map[string]any) error {
	content = strings.TrimSpace(content)
	if tenantID == "" || content == "" {
		return nil
	}
	sum := sha256.Sum256([]byte(tenantID + "\x00" + content))
	hash := hex.EncodeToString(sum[:])
	id := "ev_" + hash[:24]
	metaJSON, _ := json.Marshal(metadata)
	if metaJSON == nil {
		metaJSON = []byte("{}")
	}
	ns := "default"
	_, err := s.pool.Exec(ctx, `
INSERT INTO memory_evidence (
  id, tenant_id, namespace, subject_id, source_type, source_ref, session_id,
  content, content_hash, occurred_at, recorded_at, metadata, memory_id
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,now(),$11::jsonb,$12)
ON CONFLICT (tenant_id, content_hash) DO NOTHING
`, id, tenantID, ns, subjectID, sourceType, sourceRef, sessionID, content, hash, occurredAt, string(metaJSON), memoryID)
	return err
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
    confidence = EXCLUDED.confidence
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
