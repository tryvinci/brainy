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
	"brainy/internal/embedding"
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
		extractor: NewDeterministicExtractor(),
		packs:     packs,
		now:       time.Now().UTC,
		id:        defaultID,
	}
}

// WithExtractor overrides the sync ingest extractor (tests / advanced wiring).
func (s *Service) WithExtractor(extractor Extractor) *Service {
	if extractor != nil {
		s.extractor = extractor
	}
	return s
}

func (s *Service) Ingest(ctx context.Context, req IngestRequest) (IngestResult, error) {
	if err := validateIngestRequest(req); err != nil {
		return IngestResult{}, err
	}
	if err := validatePackMetadata(s.packs, req); err != nil {
		return IngestResult{}, err
	}

	memories := s.extractOrLabel(req)
	if len(memories) == 0 && strings.TrimSpace(req.Label) != "performance_outcome" {
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
		record, err := BuildMemoryRecord(s.id("mem"), s.now(), req, extracted, s.packs)
		if err != nil {
			return IngestResult{}, err
		}

		upserted, err := s.store.UpsertMemory(ctx, record)
		if err != nil {
			return IngestResult{}, err
		}
		s.persistEmbedding(ctx, upserted.Record)

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

	if belief, ok := synthesizeBeliefFromOutcome(req); ok {
		now := s.now()
		belief.MemoryID = s.id("mem")
		belief.TenantID = req.TenantID
		belief.SubjectID = req.SubjectID
		belief.SourceType = req.SourceType
		belief.DedupeKey = DedupeKey(req.TenantID, req.SubjectID, belief.Kind, belief.Content)
		belief.Status = StatusActive
		belief.Confidence = 0.9
		belief.ExtractionVersion = "outcome-belief-v1"
		belief.LifecycleState = LifecycleActive
		belief.Vertical = strings.TrimSpace(req.Vertical)
		if belief.Vertical == "" {
			belief.Vertical = VerticalCore
		}
		if scope := strings.TrimSpace(req.Scope); scope != "" {
			belief.Scope = scope
		}
		belief.CreatedAt = now
		belief.UpdatedAt = now
		belief.Explain = map[string]any{"rule": "outcome_to_belief"}

		upserted, err := s.store.UpsertMemory(ctx, belief)
		if err != nil {
			return IngestResult{}, err
		}
		s.persistEmbedding(ctx, upserted.Record)
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

type rankedSearchResult struct {
	result    SearchResult
	eventTime time.Time
}

func (s *Service) Search(ctx context.Context, tenantID, subjectID, vertical, scope, query string) (SearchResponse, error) {
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

	queryVector := embedding.Embed(query)
	embedScores := s.embeddingScores(ctx, tenantID, subjectID, queryVector)
	candidates := make(map[string]MemoryRecord, len(memories))
	for _, record := range memories {
		candidates[record.MemoryID] = record
	}
	if len(embedScores) > 0 {
		allMemories, err := s.store.ListActiveMemories(ctx, tenantID, subjectID)
		if err != nil {
			return SearchResponse{}, err
		}
		for _, record := range allMemories {
			if score := embedScores[record.MemoryID]; score >= 0.15 {
				if _, ok := candidates[record.MemoryID]; !ok {
					candidates[record.MemoryID] = record
				}
			}
		}
	}

	var packWeights map[string]int
	if p, ok := s.packs.Get(vertical); ok {
		packWeights = p.RankPolicy.PrimitiveWeights
	}

	ranked := make([]rankedSearchResult, 0, len(candidates))
	for _, record := range candidates {
		if !IsLifecycleSearchVisible(record.LifecycleState) {
			continue
		}
		if vertical != "" && vertical != VerticalCore && record.Vertical != vertical && record.Vertical != VerticalCore {
			continue
		}
		if p, ok := s.packs.Get(record.Vertical); ok {
			if effect := p.LifecycleEffectFor(record.Label, record.Metadata); effect != nil && effect.ExcludeFromSearch {
				continue
			}
		}
		score, explain := scoreMemory(record, queryTokens, packWeights)
		if explain == nil {
			explain = map[string]any{}
		}
		embedScore := embedScores[record.MemoryID]
		if embedScore == 0 {
			embedScore = embedding.CosineSimilarity(queryVector, recordEmbedding(record))
		}
		score = applyHybridScore(score, explain, embedScore)
		applyConvictionBoost(&score, explain, record)
		applyTasteSignalBoost(&score, explain, record, queryTokens)
		if mult := LifecycleRankMultiplier(s.packs, record); mult != 1 {
			score *= mult
			explain["lifecycle_rank_multiplier"] = mult
			if state := record.LifecycleState; state != "" && state != LifecycleActive {
				explain["lifecycle_state"] = state
			}
		}
		applyScopeBoost(&score, explain, record, scope, s.packs)
		if score <= 0 {
			continue
		}
		ranked = append(ranked, rankedSearchResult{
			result: SearchResult{
				MemoryID: record.MemoryID,
				Kind:     record.Kind,
				Content:  record.Content,
				Score:    score,
				Explain:  explain,
			},
			eventTime: EventTime(record),
		})
	}

	applyRelativeRecencyBoost(ranked)

	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].result.Score == ranked[j].result.Score {
			if ranked[i].eventTime.Equal(ranked[j].eventTime) {
				return ranked[i].result.MemoryID > ranked[j].result.MemoryID
			}
			return ranked[i].eventTime.After(ranked[j].eventTime)
		}
		return ranked[i].result.Score > ranked[j].result.Score
	})

	results := make([]SearchResult, len(ranked))
	for i, item := range ranked {
		results[i] = item.result
	}

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

func preferenceQuery(tokens []string) bool {
	if hasResponseKeyword(tokens) {
		return true
	}
	for _, token := range tokens {
		switch token {
		case "prefer", "prefers", "preference", "like", "likes", "love", "loves", "hate", "hates", "style", "tone", "concise", "detailed":
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
	s.persistEmbedding(ctx, record)

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
	explain := map[string]any{
		"matched_terms": matched,
		"ranking_basis": "deterministic_baseline",
	}
	if record.Primitive != "" {
		explain["primitive"] = record.Primitive
	}

	if record.Primitive != PrimitivePrinciple && record.Primitive != PrimitiveIdentityPrior {
		switch record.Kind {
		case KindPreference:
			// Preferential kind boost only when the query is actually about tastes/style.
			if preferenceQuery(queryTokens) {
				score += 0.25
			} else {
				score += 0.05
			}
		case KindProfile:
			score += 0.15
		case KindFact:
			score += 0.05
			overlap := float64(len(matched)) / float64(len(queryTokens))
			if overlap >= 0.5 {
				score += 0.2
				explain["dense_overlap_boost"] = 0.2
			}
			if record.Primitive == PrimitiveEpisode {
				score += 0.1
				explain["episode_boost"] = 0.1
			}
		}
	}

	if bonus := exactSpanBoost(queryTokens, record.Content); bonus > 0 {
		score += bonus
		explain["exact_span_boost"] = bonus
	}
	if bonus := dateTokenBoost(queryTokens, contentTokens, record); bonus > 0 {
		score += bonus
		explain["date_token_boost"] = bonus
	}

	applyPrimitiveBonus(&score, explain, record, primitiveWeights)

	if record.CorrectedAt != nil {
		score += 0.3
		explain["corrected"] = true
		explain["corrected_at"] = record.CorrectedAt.Format(time.RFC3339)
	}

	return score, explain
}

func exactSpanBoost(queryTokens []string, content string) float64 {
	if len(queryTokens) < 2 {
		return 0
	}
	contentNorm := " " + strings.Join(tokenize(content), " ") + " "
	// Longest contiguous query token run present in content.
	best := 0
	for i := 0; i < len(queryTokens); i++ {
		run := 0
		for j := i; j < len(queryTokens); j++ {
			phrase := " " + strings.Join(queryTokens[i:j+1], " ") + " "
			if !strings.Contains(contentNorm, phrase) {
				break
			}
			run = j - i + 1
		}
		if run > best {
			best = run
		}
	}
	if best >= 3 {
		return 0.25
	}
	if best >= 2 {
		return 0.12
	}
	return 0
}

func dateTokenBoost(queryTokens, contentTokens []string, record MemoryRecord) float64 {
	queryHasWhen := whenQuery(queryTokens)
	queryDates := filterDateTokens(queryTokens)
	contentDates := filterDateTokens(contentTokens)
	if len(contentDates) == 0 {
		// Also check source text / metadata when slots.
		contentDates = filterDateTokens(tokenize(record.SourceText))
		if when, ok := record.Metadata["when"].(string); ok {
			contentDates = append(contentDates, filterDateTokens(tokenize(when))...)
		}
	}
	if len(contentDates) == 0 {
		return 0
	}
	if len(queryDates) > 0 {
		for _, qd := range queryDates {
			for _, cd := range contentDates {
				if qd == cd {
					return 0.2
				}
			}
		}
	}
	if queryHasWhen {
		return 0.15
	}
	return 0
}

func whenQuery(tokens []string) bool {
	for _, token := range tokens {
		switch token {
		case "when", "before", "after", "during", "date", "dated", "yesterday", "today", "ago":
			return true
		}
	}
	return false
}

func filterDateTokens(tokens []string) []string {
	out := make([]string, 0)
	for _, token := range tokens {
		if looksLikeDateToken(token) {
			out = append(out, token)
		}
	}
	return out
}

func looksLikeDateToken(token string) bool {
	if token == "" {
		return false
	}
	switch token {
	case "january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december",
		"jan", "feb", "mar", "apr", "jun", "jul", "aug", "sep", "sept", "oct", "nov", "dec":
		return true
	}
	// year
	if len(token) == 4 {
		allDigits := true
		for _, r := range token {
			if r < '0' || r > '9' {
				allDigits = false
				break
			}
		}
		if allDigits && (token[0] == '1' || token[0] == '2') {
			return true
		}
	}
	// day number 1-31
	if len(token) <= 2 {
		allDigits := true
		for _, r := range token {
			if r < '0' || r > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			return true
		}
	}
	return false
}

const relativeRecencyBoostMax = 0.05

func applyRelativeRecencyBoost(ranked []rankedSearchResult) {
	if len(ranked) == 0 {
		return
	}
	minEvent := ranked[0].eventTime
	maxEvent := ranked[0].eventTime
	for _, item := range ranked[1:] {
		if item.eventTime.Before(minEvent) {
			minEvent = item.eventTime
		}
		if item.eventTime.After(maxEvent) {
			maxEvent = item.eventTime
		}
	}

	span := maxEvent.Sub(minEvent)
	if span == 0 {
		ids := make([]string, len(ranked))
		for i, item := range ranked {
			ids[i] = item.result.MemoryID
		}
		sort.Strings(ids)
		idRank := make(map[string]int, len(ids))
		for i, id := range ids {
			idRank[id] = i
		}
		maxRank := len(ids) - 1
		for i := range ranked {
			var bonus float64
			if maxRank == 0 {
				bonus = relativeRecencyBoostMax / 2
			} else {
				bonus = relativeRecencyBoostMax * float64(idRank[ranked[i].result.MemoryID]) / float64(maxRank)
			}
			ranked[i].result.Score += bonus
			if ranked[i].result.Explain == nil {
				ranked[i].result.Explain = map[string]any{}
			}
			ranked[i].result.Explain["recency_bonus"] = bonus
		}
		return
	}

	for i := range ranked {
		bonus := relativeRecencyBoostMax * float64(ranked[i].eventTime.Sub(minEvent)) / float64(span)
		ranked[i].result.Score += bonus
		if ranked[i].result.Explain == nil {
			ranked[i].result.Explain = map[string]any{}
		}
		ranked[i].result.Explain["recency_bonus"] = bonus
	}
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

func (s *Service) extractOrLabel(req IngestRequest) []ExtractedMemory {
	memories, err := s.extractor.Extract(context.Background(), req)
	if err != nil || len(memories) == 0 {
		if err != nil {
			memories = nil
		}
	}
	if len(memories) > 0 {
		return memories
	}
	label := strings.TrimSpace(req.Label)
	if label == "" || s.packs == nil {
		return nil
	}
	vertical := strings.TrimSpace(req.Vertical)
	if vertical == "" {
		vertical = VerticalCore
	}
	p, ok := s.packs.Get(vertical)
	if !ok {
		return nil
	}
	entry, ok := p.Vocabulary[label]
	if !ok {
		return nil
	}
	content := firstMessageContent(req)
	if content == "" {
		return nil
	}
	return []ExtractedMemory{{
		Kind:       entry.Kind,
		Content:    titleSentence(content),
		SourceText: content,
		Confidence: 0.9,
		Explain: map[string]any{
			"rule": "pack_label_direct",
		},
	}}
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
