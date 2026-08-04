package postgres

import (
	"context"
	"strings"
	"time"
)

// UpsertMemoryAtom indexes a typed (predicate, value) atom for enumeration scans.
func (s *Store) UpsertMemoryAtom(ctx context.Context, tenantID, subjectID, predicate, value, memoryID string, observedAt *time.Time) error {
	predicate = strings.TrimSpace(predicate)
	valueNorm := strings.ToLower(strings.TrimSpace(value))
	if tenantID == "" || subjectID == "" || predicate == "" || valueNorm == "" || memoryID == "" {
		return nil
	}
	now := time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
INSERT INTO memory_atoms (
  tenant_id, subject_id, predicate, value_norm, memory_id,
  observed_at, valid_from, recorded_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $6, $7, $7)
ON CONFLICT (tenant_id, subject_id, predicate, value_norm, memory_id) DO UPDATE
SET observed_at = COALESCE(EXCLUDED.observed_at, memory_atoms.observed_at),
    valid_from = COALESCE(memory_atoms.valid_from, EXCLUDED.valid_from),
    recorded_at = COALESCE(memory_atoms.recorded_at, EXCLUDED.recorded_at),
    retired_at = NULL,
    valid_to = NULL,
    updated_at = EXCLUDED.updated_at
`, tenantID, subjectID, predicate, valueNorm, memoryID, observedAt, now)
	return err
}

// RetireMemoryAtom marks an atom as historically ended (bitemporal retirement).
func (s *Store) RetireMemoryAtom(ctx context.Context, tenantID, subjectID, predicate, value, memoryID string, endedAt *time.Time) error {
	predicate = strings.TrimSpace(predicate)
	valueNorm := strings.ToLower(strings.TrimSpace(value))
	if tenantID == "" || subjectID == "" || predicate == "" || memoryID == "" {
		return nil
	}
	now := time.Now().UTC()
	end := now
	if endedAt != nil {
		end = endedAt.UTC()
	}
	_, err := s.pool.Exec(ctx, `
UPDATE memory_atoms
SET valid_to = COALESCE(valid_to, $6),
    retired_at = $7,
    updated_at = $7
WHERE tenant_id = $1 AND subject_id = $2 AND predicate = $3
  AND ($4 = '' OR value_norm = $4)
  AND memory_id = $5
  AND retired_at IS NULL
`, tenantID, subjectID, predicate, valueNorm, memoryID, end, now)
	return err
}

// ListAtomMemoryIDs returns memory IDs for a predicate scan (optional value_norm).
// Empty predicate lists any current atoms for the subject (semantic oracle).
// By default excludes retired / valid_to-in-past atoms (current-state view).
func (s *Store) ListAtomMemoryIDs(ctx context.Context, tenantID, subjectID, predicate, valueNorm string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 50
	}
	predicate = strings.TrimSpace(predicate)
	valueNorm = strings.ToLower(strings.TrimSpace(valueNorm))
	now := time.Now().UTC()

	var query string
	var args []any
	switch {
	case predicate == "":
		query = `
SELECT memory_id FROM memory_atoms
WHERE tenant_id = $1 AND subject_id = $2
  AND retired_at IS NULL
  AND (valid_to IS NULL OR valid_to > $4)
ORDER BY updated_at DESC
LIMIT $3`
		args = []any{tenantID, subjectID, limit, now}
	case valueNorm == "":
		query = `
SELECT memory_id FROM memory_atoms
WHERE tenant_id = $1 AND subject_id = $2 AND predicate = $3
  AND retired_at IS NULL
  AND (valid_to IS NULL OR valid_to > $5)
ORDER BY updated_at DESC
LIMIT $4`
		args = []any{tenantID, subjectID, predicate, limit, now}
	default:
		query = `
SELECT memory_id FROM memory_atoms
WHERE tenant_id = $1 AND subject_id = $2 AND predicate = $3 AND value_norm = $4
  AND retired_at IS NULL
  AND (valid_to IS NULL OR valid_to > $6)
ORDER BY updated_at DESC
LIMIT $5`
		args = []any{tenantID, subjectID, predicate, valueNorm, limit, now}
	}
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]string, 0, limit)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
