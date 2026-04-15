package memory

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Store interface {
	UpsertMemory(ctx context.Context, record MemoryRecord) (StoreUpsertResult, error)
	ListActiveMemories(ctx context.Context, tenantID, subjectID string) ([]MemoryRecord, error)
	SuppressMemory(ctx context.Context, tenantID, subjectID, memoryID string) error
	CorrectMemory(ctx context.Context, tenantID, subjectID, memoryID, content, sourceText string) (MemoryRecord, error)
}

type StoreUpsertResult struct {
	Record MemoryRecord
	State  string
}

type Service struct {
	store     Store
	extractor Extractor
	now       func() time.Time
	id        func(prefix string) string
}

func NewService(store Store) *Service {
	return &Service{
		store:     store,
		extractor: NewExtractor(),
		now:       time.Now().UTC,
		id:        defaultID,
	}
}

func (s *Service) Ingest(ctx context.Context, req IngestRequest) (IngestResult, error) {
	if req.TenantID == "" || req.SubjectID == "" || req.SourceType == "" || len(req.Messages) == 0 {
		return IngestResult{}, errors.New("tenant_id, subject_id, source_type, and messages are required")
	}

	memories := s.extractor.Extract(req)
	if len(memories) == 0 {
		return IngestResult{
			IngestID: s.id("ing"),
			Accepted: true,
		}, nil
	}

	result := IngestResult{
		IngestID: s.id("ing"),
		Accepted: true,
	}

	for _, extracted := range memories {
		now := s.now()
		record := MemoryRecord{
			MemoryID:          s.id("mem"),
			TenantID:          req.TenantID,
			SubjectID:         req.SubjectID,
			Kind:              extracted.Kind,
			Content:           extracted.Content,
			SourceText:        extracted.SourceText,
			SourceType:        req.SourceType,
			DedupeKey:         DedupeKey(req.TenantID, req.SubjectID, extracted.Kind, extracted.Content),
			Status:            StatusActive,
			Confidence:        extracted.Confidence,
			ExtractionVersion: "deterministic-v1",
			Explain:           extracted.Explain,
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		upserted, err := s.store.UpsertMemory(ctx, record)
		if err != nil {
			return IngestResult{}, err
		}

		switch upserted.State {
		case "created":
			result.Created++
		case "updated":
			result.Updated++
		default:
			result.Deduped++
		}

		result.Memories = append(result.Memories, IngestResultMemory{
			MemoryID: upserted.Record.MemoryID,
			Kind:     upserted.Record.Kind,
			Content:  upserted.Record.Content,
			Status:   upserted.Record.Status,
		})
	}

	return result, nil
}

func (s *Service) Search(ctx context.Context, tenantID, subjectID, query string) (SearchResponse, error) {
	if tenantID == "" || subjectID == "" || query == "" {
		return SearchResponse{}, errors.New("tenant_id, subject_id, and q are required")
	}

	memories, err := s.store.ListActiveMemories(ctx, tenantID, subjectID)
	if err != nil {
		return SearchResponse{}, err
	}

	queryTokens := tokenize(query)
	results := make([]SearchResult, 0, len(memories))
	for _, record := range memories {
		score, explain := scoreMemory(record, queryTokens)
		if score <= 0 {
			continue
		}
		results = append(results, SearchResult{
			MemoryID: record.MemoryID,
			Kind:     record.Kind,
			Content:  record.Content,
			Score:    score,
			Explain:  explain,
		})
	}

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].MemoryID < results[j].MemoryID
		}
		return results[i].Score > results[j].Score
	})

	return SearchResponse{Results: results}, nil
}

func (s *Service) Suppress(ctx context.Context, tenantID, subjectID, memoryID string) error {
	if tenantID == "" || subjectID == "" || memoryID == "" {
		return errors.New("tenant_id, subject_id, and memory_id are required")
	}
	return s.store.SuppressMemory(ctx, tenantID, subjectID, memoryID)
}

func (s *Service) Correct(ctx context.Context, tenantID, subjectID, memoryID string, req CorrectionRequest) (MutationResult, error) {
	if tenantID == "" || subjectID == "" || memoryID == "" {
		return MutationResult{}, errors.New("tenant_id, subject_id, and memory_id are required")
	}
	content := NormalizeText(req.Content)
	if content == "" {
		return MutationResult{}, errors.New("content is required")
	}
	sourceText := NormalizeText(req.SourceText)
	if sourceText == "" {
		sourceText = content
	}

	record, err := s.store.CorrectMemory(ctx, tenantID, subjectID, memoryID, content, sourceText)
	if err != nil {
		return MutationResult{}, err
	}

	return MutationResult{
		MemoryID: record.MemoryID,
		Kind:     record.Kind,
		Content:  record.Content,
		Status:   record.Status,
	}, nil
}

func scoreMemory(record MemoryRecord, queryTokens []string) (float64, map[string]any) {
	contentTokens := tokenize(record.Content)
	matched := make([]string, 0, len(queryTokens))
	for _, queryToken := range queryTokens {
		for _, contentToken := range contentTokens {
			if queryToken == contentToken {
				matched = append(matched, queryToken)
				break
			}
		}
	}

	if len(matched) == 0 {
		if preferenceResponseQuery(record, queryTokens) {
			return 0.82, map[string]any{
				"matched_terms": []string{"response_style"},
				"ranking_basis": "deterministic_baseline",
			}
		}
		return 0, nil
	}

	score := float64(len(matched)) / float64(len(queryTokens))
	switch record.Kind {
	case KindPreference:
		score += 0.25
	case KindProfile:
		score += 0.15
	case KindFact:
		score += 0.05
	}

	return score, map[string]any{
		"matched_terms": matched,
		"ranking_basis": "deterministic_baseline",
	}
}

func preferenceResponseQuery(record MemoryRecord, queryTokens []string) bool {
	if record.Kind != KindPreference {
		return false
	}

	needsResponseStyle := false
	for _, token := range queryTokens {
		switch token {
		case "respond", "response", "reply", "answer", "answers", "write":
			needsResponseStyle = true
		}
	}
	if !needsResponseStyle {
		return false
	}

	content := strings.ToLower(record.Content + " " + record.SourceText)
	for _, token := range []string{"prefer", "preference", "concise", "direct", "tone", "style"} {
		if strings.Contains(content, token) {
			return true
		}
	}
	return false
}

func tokenize(value string) []string {
	value = strings.ToLower(value)
	value = strings.NewReplacer(",", " ", ".", " ", "?", " ", "!", " ", ":", " ", ";", " ").Replace(value)
	tokens := strings.Fields(value)
	out := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if token != "" {
			out = append(out, token)
		}
	}
	return out
}

func defaultID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UTC().UnixNano())
}
