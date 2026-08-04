package postgres

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"brainy/internal/embedding"
)

type scoredMemory struct {
	id    string
	score float64
}

func (s *Store) UpsertEmbedding(ctx context.Context, memoryID, tenantID, subjectID string, values []float32) error {
	if memoryID == "" || len(values) == 0 {
		return nil
	}
	floats := make([]float64, len(values))
	for i, value := range values {
		floats[i] = float64(value)
	}
	now := time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
INSERT INTO memory_embeddings (memory_id, tenant_id, subject_id, embedding, updated_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (memory_id) DO UPDATE
SET embedding = EXCLUDED.embedding, updated_at = EXCLUDED.updated_at
`, memoryID, tenantID, subjectID, floats, now)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, `
UPDATE memory_records SET embedding = $2 WHERE memory_id = $1
`, memoryID, floats)
	if err != nil {
		return err
	}

	// pgvector ANN uses embedding_vec_768 for hosted dims; hash/128 stays on float[] / legacy embedding_vec.
	if len(floats) != embedding.ProviderDim {
		return nil
	}

	var hasVector bool
	if err := s.pool.QueryRow(ctx, `
SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector')
`).Scan(&hasVector); err != nil || !hasVector {
		return nil
	}

	_, _ = s.pool.Exec(ctx, `
UPDATE memory_embeddings
SET embedding_vec_768 = $2::vector(768)
WHERE memory_id = $1
  AND EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'memory_embeddings' AND column_name = 'embedding_vec_768'
  )
`, memoryID, vectorLiteral(floats))
	return nil
}

func (s *Store) LoadEmbeddings(ctx context.Context, tenantID, subjectID string) (map[string][]float32, error) {
	return s.LoadEmbeddingsLimited(ctx, tenantID, subjectID, 0)
}

// LoadEmbeddingsLimited caps in-process cosine fallbacks (never unbounded hot path).
func (s *Store) LoadEmbeddingsLimited(ctx context.Context, tenantID, subjectID string, limit int) (map[string][]float32, error) {
	query := `
SELECT memory_id, embedding
FROM memory_embeddings
WHERE tenant_id = $1 AND subject_id = $2
ORDER BY updated_at DESC`
	args := []any{tenantID, subjectID}
	if limit > 0 {
		query += `
LIMIT $3`
		args = append(args, limit)
	}
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string][]float32{}
	for rows.Next() {
		var memoryID string
		var values []float64
		if err := rows.Scan(&memoryID, &values); err != nil {
			return nil, err
		}
		embedding := make([]float32, len(values))
		for i, value := range values {
			embedding[i] = float32(value)
		}
		out[memoryID] = embedding
	}
	return out, rows.Err()
}

func (s *Store) SearchByEmbedding(ctx context.Context, tenantID, subjectID string, query []float32, limit int) (map[string]float64, error) {
	if limit <= 0 {
		limit = 20
	}
	floats := make([]float64, len(query))
	for i, value := range query {
		floats[i] = float64(value)
	}

	var hasVector bool
	if err := s.pool.QueryRow(ctx, `
SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector')
`).Scan(&hasVector); err != nil {
		hasVector = false
	}

	out := map[string]float64{}
	if hasVector && len(query) == embedding.ProviderDim {
		rows, err := s.pool.Query(ctx, `
SELECT e.memory_id, 1 - (e.embedding_vec_768 <=> $3::vector(768)) AS similarity
FROM memory_embeddings e
JOIN memory_records m ON m.memory_id = e.memory_id
WHERE e.tenant_id = $1 AND e.subject_id = $2
  AND m.status = 'active'
  AND e.embedding_vec_768 IS NOT NULL
ORDER BY e.embedding_vec_768 <=> $3::vector(768)
LIMIT $4
`, tenantID, subjectID, vectorLiteral(floats), limit)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var memoryID string
				var similarity float64
				if err := rows.Scan(&memoryID, &similarity); err != nil {
					return nil, err
				}
				out[memoryID] = similarity
			}
			if err := rows.Err(); err != nil {
				return nil, err
			}
			if len(out) > 0 {
				return out, nil
			}
		}
	}

	// Bounded fallback: never load unbounded subject embeddings on the hot path.
	capN := limit * 8
	if capN < 64 {
		capN = 64
	}
	if capN > 256 {
		capN = 256
	}
	embeddings, err := s.LoadEmbeddingsLimited(ctx, tenantID, subjectID, capN)
	if err != nil {
		return nil, err
	}
	ranked := make([]scoredMemory, 0, len(embeddings))
	for memoryID, values := range embeddings {
		ranked = append(ranked, scoredMemory{
			id:    memoryID,
			score: cosineSimilarity(query, values),
		})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].score == ranked[j].score {
			return ranked[i].id < ranked[j].id
		}
		return ranked[i].score > ranked[j].score
	})
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}
	for _, item := range ranked {
		out[item.id] = item.score
	}
	return out, nil
}

func vectorLiteral(values []float64) string {
	if len(values) == 0 {
		return "[]"
	}
	out := "["
	for i, value := range values {
		if i > 0 {
			out += ","
		}
		out += fmt.Sprintf("%g", value)
	}
	out += "]"
	return out
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
