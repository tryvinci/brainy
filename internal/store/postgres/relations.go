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
	rel.SrcEntityID = memory.SanitizeUTF8(strings.TrimSpace(rel.SrcEntityID))
	rel.DstEntityID = memory.SanitizeUTF8(strings.TrimSpace(rel.DstEntityID))
	if rel.SrcEntityID == "" {
		rel.SrcEntityID = memory.CanonicalEntityID(rel.TenantID, rel.SubjectID, rel.SrcEntity)
	}
	if rel.DstEntityID == "" {
		rel.DstEntityID = memory.CanonicalEntityID(rel.TenantID, rel.SubjectID, rel.DstEntity)
	}
	if rel.TenantID == "" || rel.SubjectID == "" || rel.SrcEntity == "" || rel.Relation == "" || rel.DstEntity == "" || rel.MemoryID == "" {
		return nil
	}
	now := time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
INSERT INTO memory_relations (
  tenant_id, subject_id, src_entity, relation, dst_entity, memory_id, observed_at, updated_at,
  src_entity_id, dst_entity_id, valid_from, valid_to, evidence_span
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT (tenant_id, subject_id, src_entity, relation, dst_entity, memory_id) DO UPDATE
SET observed_at = COALESCE(EXCLUDED.observed_at, memory_relations.observed_at),
    src_entity_id = CASE WHEN EXCLUDED.src_entity_id <> '' THEN EXCLUDED.src_entity_id ELSE memory_relations.src_entity_id END,
    dst_entity_id = CASE WHEN EXCLUDED.dst_entity_id <> '' THEN EXCLUDED.dst_entity_id ELSE memory_relations.dst_entity_id END,
    valid_from = COALESCE(EXCLUDED.valid_from, memory_relations.valid_from),
    valid_to = COALESCE(EXCLUDED.valid_to, memory_relations.valid_to),
    evidence_span = CASE WHEN EXCLUDED.evidence_span <> '' THEN EXCLUDED.evidence_span ELSE memory_relations.evidence_span END,
    updated_at = EXCLUDED.updated_at
`, rel.TenantID, rel.SubjectID, rel.SrcEntity, rel.Relation, rel.DstEntity, rel.MemoryID, rel.ObservedAt, now,
		rel.SrcEntityID, rel.DstEntityID, rel.ValidFrom, rel.ValidTo, memory.SanitizeUTF8(rel.EvidenceSpan))
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
	srcID := srcEntity
	if !strings.HasPrefix(srcEntity, "ent:") {
		srcID = memory.CanonicalEntityID(tenantID, subjectID, srcEntity)
	}
	var (
		query string
		args  []any
	)
	if relation == "" {
		query = `
SELECT src_entity, relation, dst_entity, memory_id, observed_at,
       src_entity_id, dst_entity_id, valid_from, valid_to, evidence_span
FROM memory_relations
WHERE tenant_id = $1 AND subject_id = $2
  AND (src_entity = $3 OR src_entity_id = $3 OR (src_entity_id <> '' AND src_entity_id = $4))
ORDER BY updated_at DESC
LIMIT $5`
		args = []any{tenantID, subjectID, srcEntity, srcID, limit}
	} else {
		query = `
SELECT src_entity, relation, dst_entity, memory_id, observed_at,
       src_entity_id, dst_entity_id, valid_from, valid_to, evidence_span
FROM memory_relations
WHERE tenant_id = $1 AND subject_id = $2
  AND (src_entity = $3 OR src_entity_id = $3 OR (src_entity_id <> '' AND src_entity_id = $4))
  AND relation = $5
ORDER BY updated_at DESC
LIMIT $6`
		args = []any{tenantID, subjectID, srcEntity, srcID, relation, limit}
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
		if err := rows.Scan(
			&row.SrcEntity, &row.Relation, &row.DstEntity, &row.MemoryID, &row.ObservedAt,
			&row.SrcEntityID, &row.DstEntityID, &row.ValidFrom, &row.ValidTo, &row.EvidenceSpan,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
