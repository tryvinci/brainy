package postgres

import (
	"context"
	"time"
)

// LinkMemoryEntities appends memoryID to each entity's linked_memory_ids
// (Mem0-style hub-and-spoke). Idempotent per entity+memory.
func (s *Store) LinkMemoryEntities(ctx context.Context, tenantID, subjectID, memoryID string, entities []string) error {
	if memoryID == "" || len(entities) == 0 {
		return nil
	}
	now := time.Now().UTC()
	for _, entity := range entities {
		if entity == "" {
			continue
		}
		_, err := s.pool.Exec(ctx, `
INSERT INTO memory_entity_links (tenant_id, subject_id, entity_key, linked_memory_ids, updated_at)
VALUES ($1, $2, $3, ARRAY[$4]::text[], $5)
ON CONFLICT (tenant_id, subject_id, entity_key) DO UPDATE
SET linked_memory_ids = (
        SELECT ARRAY(SELECT DISTINCT unnest(memory_entity_links.linked_memory_ids || EXCLUDED.linked_memory_ids))
    ),
    updated_at = EXCLUDED.updated_at
`, tenantID, subjectID, entity, memoryID, now)
		if err != nil {
			return err
		}
	}
	return nil
}

// EntityHubBoosts returns boosts for memories linked to query entities.
// Ubiquitous entities (many linked memories) are attenuated or skipped.
func (s *Store) EntityHubBoosts(ctx context.Context, tenantID, subjectID string, queryEntities []string) (map[string]float64, error) {
	out := map[string]float64{}
	if len(queryEntities) == 0 {
		return out, nil
	}
	rows, err := s.pool.Query(ctx, `
SELECT entity_key, linked_memory_ids
FROM memory_entity_links
WHERE tenant_id = $1 AND subject_id = $2 AND entity_key = ANY($3::text[])
`, tenantID, subjectID, queryEntities)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var entity string
		var ids []string
		if err := rows.Scan(&entity, &ids); err != nil {
			return nil, err
		}
		n := len(ids)
		if n == 0 || n > 40 {
			// Skip speaker-scale hubs that would flood ranking.
			continue
		}
		weight := 0.5 / float64(n)
		if weight > 0.35 {
			weight = 0.35
		}
		if weight < 0.04 {
			weight = 0.04
		}
		for _, id := range ids {
			if id == "" {
				continue
			}
			out[id] += weight
			if out[id] > 0.5 {
				out[id] = 0.5
			}
		}
	}
	return out, rows.Err()
}
