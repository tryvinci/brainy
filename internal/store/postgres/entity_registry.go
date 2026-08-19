package postgres

import (
	"context"
	"strings"
	"time"

	"brainy/internal/memory"
)

// UpsertMemoryEntity dual-writes a canonical identity (R7). Aliases are merged.
func (s *Store) UpsertMemoryEntity(ctx context.Context, ent memory.MemoryEntity) error {
	ent.EntityID = memory.SanitizeUTF8(strings.TrimSpace(ent.EntityID))
	ent.CanonicalLabel = memory.SanitizeUTF8(strings.TrimSpace(ent.CanonicalLabel))
	if ent.TenantID == "" || ent.SubjectID == "" || ent.EntityID == "" || ent.CanonicalLabel == "" {
		return nil
	}
	aliases := make([]string, 0, len(ent.Aliases)+1)
	seen := map[string]struct{}{}
	for _, a := range append(memory.EntityAliases(ent.CanonicalLabel), ent.Aliases...) {
		a = memory.NormalizeEntityLabel(a)
		if a == "" {
			continue
		}
		if _, ok := seen[a]; ok {
			continue
		}
		seen[a] = struct{}{}
		aliases = append(aliases, a)
	}
	now := time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
INSERT INTO memory_entities (
  tenant_id, subject_id, entity_id, canonical_label, aliases, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (tenant_id, subject_id, entity_id) DO UPDATE
SET canonical_label = EXCLUDED.canonical_label,
    aliases = (
        SELECT ARRAY(SELECT DISTINCT unnest(memory_entities.aliases || EXCLUDED.aliases))
    ),
    updated_at = EXCLUDED.updated_at
`, ent.TenantID, ent.SubjectID, ent.EntityID, ent.CanonicalLabel, aliases, now)
	return err
}

// ResolveMemoryEntity ranked-resolves a mention inside one tenant/subject.
func (s *Store) ResolveMemoryEntity(ctx context.Context, tenantID, subjectID, mention string) (memory.MemoryEntity, bool, error) {
	mention = strings.TrimSpace(mention)
	if tenantID == "" || subjectID == "" || mention == "" {
		return memory.MemoryEntity{}, false, nil
	}
	norm := memory.NormalizeEntityLabel(mention)
	rows, err := s.pool.Query(ctx, `
SELECT entity_id, canonical_label, aliases
FROM memory_entities
WHERE tenant_id = $1 AND subject_id = $2
  AND (
    entity_id = $3
    OR lower(canonical_label) = $4
    OR $4 = ANY(aliases)
  )
`, tenantID, subjectID, mention, norm)
	if err != nil {
		return memory.MemoryEntity{}, false, err
	}
	defer rows.Close()
	cands := make([]memory.MemoryEntity, 0, 4)
	for rows.Next() {
		var e memory.MemoryEntity
		e.TenantID = tenantID
		e.SubjectID = subjectID
		if err := rows.Scan(&e.EntityID, &e.CanonicalLabel, &e.Aliases); err != nil {
			return memory.MemoryEntity{}, false, err
		}
		cands = append(cands, e)
	}
	if err := rows.Err(); err != nil {
		return memory.MemoryEntity{}, false, err
	}
	got, ok := memory.RankEntityResolution(cands, mention)
	return got, ok, nil
}
