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
INSERT INTO memory_atoms (tenant_id, subject_id, predicate, value_norm, memory_id, observed_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (tenant_id, subject_id, predicate, value_norm, memory_id) DO UPDATE
SET observed_at = COALESCE(EXCLUDED.observed_at, memory_atoms.observed_at),
    updated_at = EXCLUDED.updated_at
`, tenantID, subjectID, predicate, valueNorm, memoryID, observedAt, now)
	return err
}

// ListAtomMemoryIDs returns memory IDs for a predicate scan (optional value_norm).
func (s *Store) ListAtomMemoryIDs(ctx context.Context, tenantID, subjectID, predicate, valueNorm string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 50
	}
	predicate = strings.TrimSpace(predicate)
	valueNorm = strings.ToLower(strings.TrimSpace(valueNorm))

	var query string
	var args []any
	if valueNorm == "" {
		query = `
SELECT memory_id FROM memory_atoms
WHERE tenant_id = $1 AND subject_id = $2 AND predicate = $3
ORDER BY updated_at DESC
LIMIT $4`
		args = []any{tenantID, subjectID, predicate, limit}
	} else {
		query = `
SELECT memory_id FROM memory_atoms
WHERE tenant_id = $1 AND subject_id = $2 AND predicate = $3 AND value_norm = $4
ORDER BY updated_at DESC
LIMIT $5`
		args = []any{tenantID, subjectID, predicate, valueNorm, limit}
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
