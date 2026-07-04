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
		_ = writer.UpsertEmbedding(ctx, record.MemoryID, record.TenantID, record.SubjectID, embedding.Embed(record.Content))
	}
}

func (s *Service) embeddingScores(ctx context.Context, tenantID, subjectID string, queryVector []float32) map[string]float64 {
	if searcher, ok := s.store.(embeddingSearcher); ok {
		if scores, err := searcher.SearchByEmbedding(ctx, tenantID, subjectID, queryVector, 50); err == nil && len(scores) > 0 {
			return scores
		}
	}

	allMemories, err := s.store.ListActiveMemories(ctx, tenantID, subjectID)
	if err != nil {
		return map[string]float64{}
	}
	scores := make(map[string]float64, len(allMemories))
	for _, record := range allMemories {
		scores[record.MemoryID] = embedding.CosineSimilarity(queryVector, recordEmbedding(record))
	}
	return scores
}

func recordEmbedding(record MemoryRecord) []float32 {
	if len(record.Embedding) > 0 {
		return record.Embedding
	}
	return embedding.Embed(record.Content)
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
