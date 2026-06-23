package memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"brainy/internal/pack"
)

type Store interface {
	UpsertMemory(ctx context.Context, record MemoryRecord) (StoreUpsertResult, error)
	ListActiveMemories(ctx context.Context, tenantID, subjectID string) ([]MemoryRecord, error)
	SearchActiveMemories(ctx context.Context, tenantID, subjectID string, patterns []string, limit int) ([]MemoryRecord, error)
	SuppressMemory(ctx context.Context, tenantID, subjectID, memoryID string) error
	CorrectMemory(ctx context.Context, tenantID, subjectID, memoryID, content, sourceText string) (MemoryRecord, error)
	EnqueueIngestJob(ctx context.Context, ingestID, jobID, idempotencyKey string, req IngestRequest) (EnqueueResult, error)
	ClaimNextExtractionJob(ctx context.Context) (ExtractionJob, bool, error)
	CompleteExtractionJob(ctx context.Context, jobID, ingestID string) error
	FailExtractionJob(ctx context.Context, jobID, ingestID, reason string) error
}

type StoreUpsertResult struct {
	Record MemoryRecord
	State  string
}

type Service struct {
	store     Store
	extractor Extractor
	packs     *pack.Registry
	now       func() time.Time
	id        func(prefix string) string
}

func NewService(store Store) *Service {
	reg, _ := pack.LoadRegistryFromDir("packs")
	return NewServiceWithPacks(store, reg)
}

func NewServiceWithPacks(store Store, packs *pack.Registry) *Service {
	if packs == nil {
		packs = pack.NewRegistry()
	}
	return &Service{
		store:     store,
		extractor: NewExtractor(),
		packs:     packs,
		now:       time.Now().UTC,
		id:        defaultID,
	}
}

func (s *Service) applyPackFields(record *MemoryRecord, vertical, kind, content string) {
	ApplyVerticalPack(record, vertical, kind, content, s.packs)
}

func (s *Service) Ingest(ctx context.Context, req IngestRequest) (IngestResult, error) {
	if err := validateIngestRequest(req); err != nil {
		return IngestResult{}, err
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
		s.applyPackFields(&record, req.Vertical, extracted.Kind, extracted.Content)

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

func (s *Service) IngestAsync(ctx context.Context, req IngestRequest) (AsyncIngestResult, error) {
	if err := validateIngestRequest(req); err != nil {
		return AsyncIngestResult{}, err
	}

	result := AsyncIngestResult{
		IngestID: s.id("ing"),
		JobID:    s.id("job"),
		Accepted: true,
	}
	idempotencyKey := s.idempotencyKey(req)
	enqueueResult, err := s.store.EnqueueIngestJob(ctx, result.IngestID, result.JobID, idempotencyKey, req)
	if err != nil {
		return AsyncIngestResult{}, err
	}
	if enqueueResult.Duplicate {
		return AsyncIngestResult{
			IngestID: enqueueResult.IngestID,
			JobID:    enqueueResult.JobID,
			Accepted: true,
		}, nil
	}
	return result, nil
}

func (s *Service) idempotencyKey(req IngestRequest) string {
	parts := []string{req.TenantID, req.SubjectID, req.SourceType, req.Vertical}
	for _, m := range req.Messages {
		parts = append(parts, m.Role, m.Content)
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])
}

func (s *Service) Search(ctx context.Context, tenantID, subjectID, vertical, query string) (SearchResponse, error) {
	if tenantID == "" || subjectID == "" || query == "" {
		return SearchResponse{}, errors.New("tenant_id, subject_id, and q are required")
	}

	queryTokens := tokenize(query)

	patterns := make([]string, 0, len(queryTokens))
	for _, t := range queryTokens {
		patterns = append(patterns, "%"+t+"%")
	}

	memories, err := s.store.SearchActiveMemories(ctx, tenantID, subjectID, patterns, 100)
	if err != nil {
		return SearchResponse{}, err
	}

	if len(memories) < 10 && hasResponseKeyword(queryTokens) {
		allMemories, err := s.store.ListActiveMemories(ctx, tenantID, subjectID)
		if err != nil {
			return SearchResponse{}, err
		}
		preferenceIDs := make(map[string]struct{})
		for _, m := range memories {
			preferenceIDs[m.MemoryID] = struct{}{}
		}
		for _, m := range allMemories {
			if m.Kind == KindPreference {
				if _, ok := preferenceIDs[m.MemoryID]; !ok {
					memories = append(memories, m)
				}
			}
		}
	}

	var packWeights map[string]int
	if p, ok := s.packs.Get(vertical); ok {
		packWeights = p.RankPolicy.PrimitiveWeights
	}

	results := make([]SearchResult, 0, len(memories))
	for _, record := range memories {
		if vertical != "" && vertical != VerticalCore && record.Vertical != vertical && record.Vertical != VerticalCore {
			continue
		}
		score, explain := scoreMemory(record, queryTokens, packWeights)
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

func hasResponseKeyword(tokens []string) bool {
	for _, token := range tokens {
		switch token {
		case "respond", "response", "reply", "answer", "answers", "write":
			return true
		}
	}
	return false
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

func scoreMemory(record MemoryRecord, queryTokens []string, primitiveWeights map[string]int) (float64, map[string]any) {
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
		if !preferenceResponseQuery(record, queryTokens) {
			return 0, nil
		}
		score := 0.82
		explain := map[string]any{
			"matched_terms": []string{"response_style"},
			"ranking_basis": "deterministic_baseline",
		}
		if record.Primitive != "" {
			explain["primitive"] = record.Primitive
		}
		applyPrimitiveBonus(&score, explain, record, primitiveWeights)
		if record.CorrectedAt != nil {
			score += 0.3
			explain["corrected"] = true
			explain["corrected_at"] = record.CorrectedAt.Format(time.RFC3339)
		}
		return score, explain
	}

	score := float64(len(matched)) / float64(len(queryTokens))
	if record.Primitive != PrimitivePrinciple && record.Primitive != PrimitiveIdentityPrior {
		switch record.Kind {
		case KindPreference:
			score += 0.25
		case KindProfile:
			score += 0.15
		case KindFact:
			score += 0.05
		}
	}

	explain := map[string]any{
		"matched_terms": matched,
		"ranking_basis": "deterministic_baseline",
	}
	if record.Primitive != "" {
		explain["primitive"] = record.Primitive
	}

	applyPrimitiveBonus(&score, explain, record, primitiveWeights)

	if record.CorrectedAt != nil {
		score += 0.3
		explain["corrected"] = true
		explain["corrected_at"] = record.CorrectedAt.Format(time.RFC3339)
	}

	return score, explain
}

func applyPrimitiveBonus(score *float64, explain map[string]any, record MemoryRecord, weights map[string]int) {
	if record.Primitive == "" || len(weights) == 0 {
		return
	}
	w, ok := weights[record.Primitive]
	if !ok || w <= 0 {
		return
	}
	bonus := float64(w) / 50.0
	*score += bonus
	explain["primitive_bonus"] = bonus
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

func validateIngestRequest(req IngestRequest) error {
	if req.TenantID == "" || req.SubjectID == "" || req.SourceType == "" || len(req.Messages) == 0 {
		return errors.New("tenant_id, subject_id, source_type, and messages are required")
	}
	return nil
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
