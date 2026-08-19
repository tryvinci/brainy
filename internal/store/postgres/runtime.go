package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *Store) UpsertProviderRuntime(ctx context.Context, role string, payload []byte) error {
	if role == "" {
		return nil
	}
	if len(payload) == 0 {
		payload = []byte(`{}`)
	}
	_, err := s.pool.Exec(ctx, `
INSERT INTO provider_runtime (role, payload, updated_at)
VALUES ($1, $2::jsonb, $3)
ON CONFLICT (role) DO UPDATE
SET payload = EXCLUDED.payload, updated_at = EXCLUDED.updated_at
`, role, payload, time.Now().UTC())
	return err
}

func (s *Store) GetProviderRuntime(ctx context.Context, role string) (json.RawMessage, time.Time, bool, error) {
	var payload []byte
	var updated time.Time
	err := s.pool.QueryRow(ctx, `
SELECT payload, updated_at FROM provider_runtime WHERE role = $1
`, role).Scan(&payload, &updated)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, time.Time{}, false, nil
	}
	if err != nil {
		return nil, time.Time{}, false, err
	}
	return json.RawMessage(payload), updated, true, nil
}

type EmbeddingTarget struct {
	MemoryID  string
	TenantID  string
	SubjectID string
	Content   string
}

func (s *Store) ListEmbeddingTargets(ctx context.Context, limit, offset int) ([]EmbeddingTarget, error) {
	if limit <= 0 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := s.pool.Query(ctx, `
SELECT memory_id, tenant_id, subject_id, content
FROM memory_records
WHERE status = 'active'
ORDER BY memory_id
LIMIT $1 OFFSET $2
`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmbeddingTarget
	for rows.Next() {
		var row EmbeddingTarget
		if err := rows.Scan(&row.MemoryID, &row.TenantID, &row.SubjectID, &row.Content); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (s *Store) RuntimeParts(ctx context.Context) (map[string]any, error) {
	st, err := s.ANNStatus(ctx)
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"ann": st,
	}
	if payload, updated, ok, err := s.GetProviderRuntime(ctx, "worker"); err == nil && ok {
		var worker any
		if json.Unmarshal(payload, &worker) == nil {
			out["worker"] = worker
		} else {
			out["worker"] = map[string]any{"raw": string(payload)}
		}
		out["worker_updated_at"] = updated.UTC().Format(time.RFC3339)
	}
	return out, nil
}
