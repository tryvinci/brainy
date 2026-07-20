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
// thresholds are meaningless. We min-max normalize the candidate distribution so
// the strongest matches keep high signal while the baseline noise floor is
// suppressed toward zero. Model-agnostic; preserves top-match recall.
//
// Small or flat candidate sets are returned unchanged: with few candidates the
// distribution is uninformative and normalizing could erase a lone true match
// (e.g. a single paraphrase in a tiny corpus).
func calibrateSimilarities(raw map[string]float64) map[string]float64 {
	if len(raw) < 6 {
		return raw
	}
	var min, max float64
	first := true
	for _, v := range raw {
		if first {
			min, max, first = v, v, false
			continue
		}
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	spread := max - min
	if spread < 0.05 {
		// Nearly flat: similarities carry no discriminative signal.
		return raw
	}
	out := make(map[string]float64, len(raw))
	for k, v := range raw {
		r := (v - min) / spread
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
