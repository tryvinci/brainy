package memory

import (
	"context"

	"brainy/internal/embedding"
)

type embeddingWriter interface {
	UpsertEmbedding(ctx context.Context, memoryID, tenantID, subjectID string, values []float32) error
}

type embeddingSearcher interface {
	SearchByEmbedding(ctx context.Context, tenantID, subjectID string, query []float32, limit int) (map[string]float64, error)
}

func (s *Service) persistEmbedding(ctx context.Context, record MemoryRecord) {
	if writer, ok := s.store.(embeddingWriter); ok {
		values, err := s.embed(ctx, record.Content)
		if err != nil || len(values) == 0 {
			return
		}
		_ = writer.UpsertEmbedding(ctx, record.MemoryID, record.TenantID, record.SubjectID, values)
	}
}

func (s *Service) embed(ctx context.Context, text string) ([]float32, error) {
	embedder := s.embedder
	if embedder == nil {
		embedder = embedding.Default()
	}
	return embedder.Embed(ctx, text)
}

func (s *Service) embeddingScores(ctx context.Context, tenantID, subjectID string, queryVector []float32) map[string]float64 {
	var raw map[string]float64
	if searcher, ok := s.store.(embeddingSearcher); ok {
		if scores, err := searcher.SearchByEmbedding(ctx, tenantID, subjectID, queryVector, 50); err == nil && len(scores) > 0 {
			raw = scores
		}
	}
	if raw == nil {
		allMemories, err := s.store.ListActiveMemories(ctx, tenantID, subjectID)
		if err != nil {
			return map[string]float64{}
		}
		raw = make(map[string]float64, len(allMemories))
		for _, record := range allMemories {
			raw[record.MemoryID] = embedding.CosineSimilarity(queryVector, s.recordEmbedding(ctx, record))
		}
	}
	return calibrateSimilarities(raw)
}

// calibrateSimilarities rescales raw cosine similarities to a per-query relative
// scale. Modern embedding models have a high, model-specific baseline similarity
// for arbitrary text (e.g. bge-small ~0.5 for unrelated English), so absolute
// thresholds are meaningless. We map the candidate distribution so only
// above-baseline matches carry positive signal: rescaled = (v - floor)/(1-floor)
// with floor = mean similarity, clamped to [0,1]. Model-agnostic, self-calibrating.
func calibrateSimilarities(raw map[string]float64) map[string]float64 {
	if len(raw) == 0 {
		return raw
	}
	var sum, max float64
	for _, v := range raw {
		sum += v
		if v > max {
			max = v
		}
	}
	mean := sum / float64(len(raw))
	// If the spread is tiny, similarities are uninformative — zero them out.
	floor := mean
	denom := 1.0 - floor
	if denom < 1e-6 || max-mean < 0.02 {
		out := make(map[string]float64, len(raw))
		for k := range raw {
			out[k] = 0
		}
		return out
	}
	out := make(map[string]float64, len(raw))
	for k, v := range raw {
		r := (v - floor) / denom
		if r < 0 {
			r = 0
		}
		if r > 1 {
			r = 1
		}
		out[k] = r
	}
	return out
}

func (s *Service) recordEmbedding(ctx context.Context, record MemoryRecord) []float32 {
	if len(record.Embedding) > 0 {
		return record.Embedding
	}
	values, err := s.embed(ctx, record.Content)
	if err != nil {
		return nil
	}
	return values
}

func applyHybridScore(tokenScore float64, explain map[string]any, embeddingSimilarity float64) float64 {
	if embeddingSimilarity <= 0 {
		return tokenScore
	}
	if tokenScore <= 0 && embeddingSimilarity >= 0.15 {
		score := embeddingSimilarity * 0.9
		explain["ranking_basis"] = "hybrid_embedding"
		explain["embedding_similarity"] = embeddingSimilarity
		return score
	}
	if tokenScore > 0 {
		score := tokenScore + embeddingSimilarity*0.35
		explain["embedding_similarity"] = embeddingSimilarity
		if explain["ranking_basis"] == "deterministic_baseline" {
			explain["ranking_basis"] = "hybrid"
		}
		return score
	}
	return tokenScore
}
