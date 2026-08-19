package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"brainy/internal/embedding"
)

func (s *Store) UpsertEmbedding(ctx context.Context, memoryID, tenantID, subjectID string, values []float32) error {
	return s.WriteEmbedding(ctx, embedding.Record{
		MemoryID:  memoryID,
		TenantID:  tenantID,
		SubjectID: subjectID,
		Values:    values,
	})
}

func (s *Store) WriteEmbedding(ctx context.Context, rec embedding.Record) error {
	if rec.MemoryID == "" || len(rec.Values) == 0 {
		return nil
	}
	floats := make([]float64, len(rec.Values))
	for i, value := range rec.Values {
		floats[i] = float64(value)
	}
	dims := rec.Dimensions
	if dims <= 0 {
		dims = len(floats)
	}
	now := time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
INSERT INTO memory_embeddings (
    memory_id, tenant_id, subject_id, embedding, updated_at,
    embedding_provider, embedding_model, embedding_dimensions, embedding_version
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (memory_id) DO UPDATE
SET embedding = EXCLUDED.embedding,
    updated_at = EXCLUDED.updated_at,
    embedding_provider = EXCLUDED.embedding_provider,
    embedding_model = EXCLUDED.embedding_model,
    embedding_dimensions = EXCLUDED.embedding_dimensions,
    embedding_version = EXCLUDED.embedding_version
`, rec.MemoryID, rec.TenantID, rec.SubjectID, floats, now, rec.Provider, rec.Model, dims, rec.Version)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, `
UPDATE memory_records SET embedding = $2 WHERE memory_id = $1
`, rec.MemoryID, floats)
	if err != nil {
		return err
	}

	ready, err := s.vectorANNReady(ctx)
	if err != nil || !ready {
		return nil
	}

	if len(floats) != embedding.ProviderDim {
		_, _ = s.pool.Exec(ctx, `
UPDATE memory_embeddings SET embedding_vec_768 = NULL WHERE memory_id = $1
`, rec.MemoryID)
		return nil
	}

	_, _ = s.pool.Exec(ctx, `
UPDATE memory_embeddings
SET embedding_vec_768 = $2::vector(768)
WHERE memory_id = $1
`, rec.MemoryID, vectorLiteral(floats))
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
		vec := make([]float32, len(values))
		for i, value := range values {
			vec[i] = float32(value)
		}
		out[memoryID] = vec
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

	ready, err := s.vectorANNReady(ctx)
	if err != nil {
		ready = false
	}

	out := map[string]float64{}
	if ready && len(query) == embedding.ProviderDim {
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
			return out, nil
		}
	}

	// ANN inactive or unusable: do not silently scan the last 256 writes.
	// Callers fall through to a full in-process cosine over ListActiveMemories.
	return map[string]float64{}, nil
}

type ANNStatus struct {
	HasPGVector     bool             `json:"has_pgvector"`
	HasVec768Column bool             `json:"has_embedding_vec_768"`
	Active          bool             `json:"active"`
	DimHistogram    map[string]int64 `json:"dim_histogram"`
	ModelHistogram  map[string]int64 `json:"model_histogram"`
	MixedDimensions bool             `json:"mixed_dimensions"`
}

func (s *Store) ANNStatus(ctx context.Context) (ANNStatus, error) {
	st := ANNStatus{
		DimHistogram:   map[string]int64{},
		ModelHistogram: map[string]int64{},
	}
	if err := s.pool.QueryRow(ctx, `
SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector')
`).Scan(&st.HasPGVector); err != nil {
		return st, err
	}
	if err := s.pool.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1 FROM information_schema.columns
  WHERE table_name = 'memory_embeddings' AND column_name = 'embedding_vec_768'
)
`).Scan(&st.HasVec768Column); err != nil {
		return st, err
	}
	st.Active = st.HasPGVector && st.HasVec768Column

	rows, err := s.pool.Query(ctx, `
SELECT cardinality(embedding)::text AS dim, COALESCE(NULLIF(embedding_model, ''), '(unset)') AS model, count(*)
FROM memory_embeddings
GROUP BY 1, 2
`)
	if err == nil {
		defer rows.Close()
		dims := map[string]struct{}{}
		for rows.Next() {
			var dim, model string
			var n int64
			if err := rows.Scan(&dim, &model, &n); err != nil {
				return st, err
			}
			st.DimHistogram[dim] += n
			st.ModelHistogram[model] += n
			dims[dim] = struct{}{}
		}
		if err := rows.Err(); err != nil {
			return st, err
		}
		st.MixedDimensions = len(dims) > 1
	}
	return st, nil
}

func (s *Store) RequireANN(ctx context.Context) error {
	st, err := s.ANNStatus(ctx)
	if err != nil {
		return err
	}
	if !st.Active {
		return fmt.Errorf("hosted 768-d embedder requires pgvector and memory_embeddings.embedding_vec_768 (pgvector=%v column=%v)", st.HasPGVector, st.HasVec768Column)
	}
	return nil
}

func (s *Store) vectorANNReady(ctx context.Context) (bool, error) {
	switch s.annReady.Load() {
	case 1:
		return true, nil
	case -1:
		return false, nil
	}
	var hasVector, hasCol bool
	if err := s.pool.QueryRow(ctx, `
SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector')
`).Scan(&hasVector); err != nil {
		return false, err
	}
	if !hasVector {
		s.annReady.Store(-1)
		return false, nil
	}
	if err := s.pool.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1 FROM information_schema.columns
  WHERE table_name = 'memory_embeddings' AND column_name = 'embedding_vec_768'
)
`).Scan(&hasCol); err != nil {
		return false, err
	}
	if hasCol {
		s.annReady.Store(1)
		return true, nil
	}
	s.annReady.Store(-1)
	return false, nil
}

func vectorLiteral(values []float64) string {
	if len(values) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteByte('[')
	for i, value := range values {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(fmt.Sprintf("%g", value))
	}
	b.WriteByte(']')
	return b.String()
}
