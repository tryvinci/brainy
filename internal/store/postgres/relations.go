package postgres

import (
	"context"
	"strings"
	"time"

	"brainy/internal/memory"
)

// UpsertMemoryRelation projects an entity-valued atomic fact as a Postgres edge.
func (s *Store) UpsertMemoryRelation(ctx context.Context, rel memory.MemoryRelation) error {
	rel.SrcEntity = memory.SanitizeUTF8(strings.ToLower(strings.TrimSpace(rel.SrcEntity)))
	rel.DstEntity = memory.SanitizeUTF8(strings.ToLower(strings.TrimSpace(rel.DstEntity)))
	rel.Relation = memory.SanitizeUTF8(strings.TrimSpace(rel.Relation))
	if rel.TenantID == "" || rel.SubjectID == "" || rel.SrcEntity == "" || rel.Relation == "" || rel.DstEntity == "" || rel.MemoryID == "" {
		return nil
	}
	now := time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
INSERT INTO memory_relations (
  tenant_id, subject_id, src_entity, relation, dst_entity, memory_id, observed_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (tenant_id, subject_id, src_entity, relation, dst_entity, memory_id) DO UPDATE
SET observed_at = COALESCE(EXCLUDED.observed_at, memory_relations.observed_at),
    updated_at = EXCLUDED.updated_at
`, rel.TenantID, rel.SubjectID, rel.SrcEntity, rel.Relation, rel.DstEntity, rel.MemoryID, rel.ObservedAt, now)
	return err
}

// ListRelationsFrom returns edges out of srcEntity, optionally filtered by relation.
func (s *Store) ListRelationsFrom(ctx context.Context, tenantID, subjectID, srcEntity, relation string, limit int) ([]memory.MemoryRelation, error) {
	srcEntity = strings.ToLower(strings.TrimSpace(srcEntity))
	relation = strings.TrimSpace(relation)
	if tenantID == "" || subjectID == "" || srcEntity == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 40
	}
	var (
		query string
		args  []any
	)
	if relation == "" {
		query = `
SELECT src_entity, relation, dst_entity, memory_id, observed_at
FROM memory_relations
WHERE tenant_id = $1 AND subject_id = $2 AND src_entity = $3
ORDER BY updated_at DESC
LIMIT $4`
		args = []any{tenantID, subjectID, srcEntity, limit}
	} else {
		query = `
SELECT src_entity, relation, dst_entity, memory_id, observed_at
FROM memory_relations
WHERE tenant_id = $1 AND subject_id = $2 AND src_entity = $3 AND relation = $4
ORDER BY updated_at DESC
LIMIT $5`
		args = []any{tenantID, subjectID, srcEntity, relation, limit}
	}
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]memory.MemoryRelation, 0, limit)
	for rows.Next() {
		var row memory.MemoryRelation
		row.TenantID = tenantID
		row.SubjectID = subjectID
		if err := rows.Scan(&row.SrcEntity, &row.Relation, &row.DstEntity, &row.MemoryID, &row.ObservedAt); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
