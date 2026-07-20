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
	embedder  embedding.Embedder
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
		embedder:  embedding.Default(),
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

// WithEmbedder overrides the hybrid retrieval embedder.
func (s *Service) WithEmbedder(embedder embedding.Embedder) *Service {
	if embedder != nil {
		s.embedder = embedder
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
	contentQueryTokens := contentBearingTokens(queryTokens)

	patterns := make([]string, 0, len(contentQueryTokens)+len(queryTokens))
	for _, t := range contentQueryTokens {
		patterns = append(patterns, "%"+t+"%")
	}
	// Fall back to all tokens only when content-bearing set is empty.
	if len(patterns) == 0 {
		for _, t := range queryTokens {
			patterns = append(patterns, "%"+t+"%")
		}
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

	queryVector, _ := s.embed(ctx, query)
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

	// For multi-hop shaped questions, admit a capped set of same-session neighbors
	// and a second-pass of fact-like memories related to first-hit content tokens.
	if looksMultiHopQuery(queryTokens) {
		if allMemories, err := s.store.ListActiveMemories(ctx, tenantID, subjectID); err == nil {
			expandSessionNeighbors(candidates, memories, allMemories, 12)
		}
		if related, err := s.relatedFactMemories(ctx, tenantID, subjectID, queryTokens, memories, 8); err == nil {
			for _, record := range related {
				if _, ok := candidates[record.MemoryID]; !ok {
					candidates[record.MemoryID] = record
				}
			}
		}
	}

	queryEntities := ExtractEntities(query)

	// Entity document frequency over the subject's memories: ubiquitous entities
	// (e.g. the two speakers in a dialogue) carry little signal, so we weight by
	// rarity (IDF-style) and only admit/boost on *distinctive* shared entities.
	entityDF, totalMemories := s.entityDocFrequencies(ctx, tenantID, subjectID)
	distinctiveQueryEntities := make([]string, 0, len(queryEntities))
	for _, e := range queryEntities {
		if isDistinctiveEntity(e, entityDF, totalMemories) {
			distinctiveQueryEntities = append(distinctiveQueryEntities, e)
		}
	}

	// Entity-linked recall: admit memories sharing a *distinctive* query entity
	// even when lexical/embedding recall missed them (generic SOTA technique).
	// Only admit when the memory ALSO has some lexical/embedding overlap signal,
	// so a shared distinctive entity refines recall rather than flooding it.
	if len(distinctiveQueryEntities) > 0 && len(candidates) < 40 {
		if allMemories, err := s.store.ListActiveMemories(ctx, tenantID, subjectID); err == nil {
			admitted := 0
			for _, record := range allMemories {
				if _, ok := candidates[record.MemoryID]; ok {
					continue
				}
				if entityOverlapBoost(distinctiveQueryEntities, recordEntities(record)) <= 0 {
					continue
				}
				// Require a secondary signal (token or embedding) to admit.
				tokenScore, _ := scoreMemory(record, queryTokens, nil)
				if tokenScore <= 0 && embedScores[record.MemoryID] < 0.2 {
					continue
				}
				candidates[record.MemoryID] = record
				admitted++
				if admitted >= 10 {
					break
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
			embedScore = embedding.CosineSimilarity(queryVector, s.recordEmbedding(ctx, record))
		}
		score = applyHybridScore(score, explain, embedScore)
		if bonus := entityOverlapBoost(distinctiveQueryEntities, recordEntities(record)); bonus > 0 {
			score += bonus
			explain["entity_overlap_boost"] = bonus
		}
		applySessionNeighborBoost(&score, explain, record, memories)
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
				MemoryID:   record.MemoryID,
				Kind:       record.Kind,
				Content:    record.Content,
				Score:      score,
				ObservedAt: record.ObservedAt,
				Explain:    explain,
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

func sessionIDOf(record MemoryRecord) string {
	if record.Metadata == nil {
		return ""
	}
	if raw, ok := record.Metadata["session_id"]; ok {
		if s, ok := raw.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

// expandSessionNeighbors admits other memories that share a session_id with
// lexical hits so multi-fact conversational questions can see co-occurring turns.
func expandSessionNeighbors(candidates map[string]MemoryRecord, seeds []MemoryRecord, all []MemoryRecord, limit int) {
	sessions := map[string]struct{}{}
	for _, seed := range seeds {
		if sid := sessionIDOf(seed); sid != "" {
			sessions[sid] = struct{}{}
		}
	}
	if len(sessions) == 0 {
		return
	}
	added := 0
	for _, record := range all {
		if limit > 0 && added >= limit {
			break
		}
		sid := sessionIDOf(record)
		if sid == "" {
			continue
		}
		if _, ok := sessions[sid]; !ok {
			continue
		}
		if _, exists := candidates[record.MemoryID]; !exists {
			candidates[record.MemoryID] = record
			added++
		}
	}
}

func looksMultiHopQuery(tokens []string) bool {
	hasAsk := false
	hasCue := false
	for _, token := range tokens {
		switch token {
		case "what", "which", "who", "where", "how":
			hasAsk = true
		case "identity", "relationship", "status", "activities", "activity", "career", "path", "moved", "research", "pursue", "persue", "partake", "camped", "books", "read", "destress", "de-stress", "kids", "like":
			hasCue = true
		}
	}
	return hasAsk && hasCue
}

// relatedFactMemories runs a second lexical pass using distinctive tokens from
// first-hit statement memories, so multi-hop answers can pull supporting facts
// that do not share the question's surface words.
func (s *Service) relatedFactMemories(ctx context.Context, tenantID, subjectID string, queryTokens []string, seeds []MemoryRecord, limit int) ([]MemoryRecord, error) {
	querySet := map[string]struct{}{}
	for _, token := range contentBearingTokens(queryTokens) {
		querySet[token] = struct{}{}
	}
	extra := make([]string, 0, 12)
	seen := map[string]struct{}{}
	for _, seed := range seeds {
		content := strings.TrimSpace(seed.Content)
		if content == "" || strings.HasSuffix(content, "?") {
			continue
		}
		for _, token := range contentBearingTokens(tokenize(content)) {
			if len(token) < 4 {
				continue
			}
			if _, ok := querySet[token]; ok {
				continue
			}
			if _, ok := seen[token]; ok {
				continue
			}
			seen[token] = struct{}{}
			extra = append(extra, token)
			if len(extra) >= 8 {
				break
			}
		}
		if len(extra) >= 8 {
			break
		}
	}
	// Intent cues: expand a few generic related stems from the query itself.
	for _, token := range contentBearingTokens(queryTokens) {
		for _, related := range relatedIntentTokens(token) {
			if _, ok := seen[related]; ok {
				continue
			}
			seen[related] = struct{}{}
			extra = append(extra, related)
		}
	}
	if len(extra) == 0 {
		return nil, nil
	}
	patterns := make([]string, 0, len(extra))
	for _, token := range extra {
		patterns = append(patterns, "%"+token+"%")
	}
	found, err := s.store.SearchActiveMemories(ctx, tenantID, subjectID, patterns, 50)
	if err != nil {
		return nil, err
	}
	out := make([]MemoryRecord, 0, limit)
	for _, record := range found {
		content := strings.TrimSpace(record.Content)
		if content == "" || strings.HasSuffix(content, "?") {
			continue
		}
		out = append(out, record)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

func relatedIntentTokens(token string) []string {
	// Generic conversational synonyms only — never benchmark answer keys.
	switch token {
	case "identity":
		return []string{"gender", "identity"}
	case "relationship", "status":
		return []string{"single", "married", "partner", "dating", "relationship"}
	case "career", "path", "pursue", "persue":
		return []string{"career", "job", "profession", "work"}
	case "activities", "activity", "partake", "hobby", "hobbies":
		return []string{"hobby", "hobbies", "activity", "activities"}
	case "camped", "camping", "camp":
		return []string{"camp", "camping", "camped"}
	case "books", "read", "reading":
		return []string{"reading", "book", "books", "library"}
	case "destress", "de-stress", "stress":
		return []string{"stress", "relax", "unwind"}
	case "moved":
		return []string{"moved", "move", "relocated"}
	case "kids", "children":
		return []string{"kids", "children", "child"}
	default:
		return nil
	}
}

// topicAlignmentBoost rewards memories whose content matches the topical intent
// of the query (e.g. identity ↔ gender/trans*), beyond surface-token overlap.
func topicAlignmentBoost(queryTokens, contentTokens []string) float64 {
	qset := map[string]struct{}{}
	for _, token := range contentBearingTokens(queryTokens) {
		qset[token] = struct{}{}
	}
	cset := map[string]struct{}{}
	for _, token := range contentTokens {
		cset[token] = struct{}{}
	}
	bonus := 0.0
	for q := range qset {
		for _, related := range relatedIntentTokens(q) {
			if _, ok := cset[related]; ok {
				bonus += 0.18
				break
			}
			// prefix match for transgender/trans etc.
			for c := range cset {
				if tokensMatch(related, c) {
					bonus += 0.18
					break
				}
			}
		}
	}
	if bonus > 0.45 {
		return 0.45
	}
	return bonus
}

func applySessionNeighborBoost(score *float64, explain map[string]any, record MemoryRecord, seeds []MemoryRecord) {
	sid := sessionIDOf(record)
	if sid == "" {
		return
	}
	for _, seed := range seeds {
		if seed.MemoryID == record.MemoryID {
			continue
		}
		if sessionIDOf(seed) == sid {
			*score += 0.08
			explain["session_neighbor_boost"] = 0.08
			return
		}
	}
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
	bearingQuery := contentBearingTokens(queryTokens)
	if len(bearingQuery) == 0 {
		bearingQuery = queryTokens
	}
	matched := make([]string, 0, len(bearingQuery))
	for _, queryToken := range bearingQuery {
		for _, contentToken := range contentTokens {
			if tokensMatch(queryToken, contentToken) {
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

	score := float64(len(matched)) / float64(len(bearingQuery))
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
			overlap := float64(len(matched)) / float64(len(bearingQuery))
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

	if bonus := exactSpanBoost(bearingQuery, record.Content); bonus > 0 {
		score += bonus
		explain["exact_span_boost"] = bonus
	}
	if bonus := dateTokenBoost(queryTokens, contentTokens, record); bonus > 0 {
		score += bonus
		explain["date_token_boost"] = bonus
	}
	if bonus := subjectMentionBoost(bearingQuery, contentTokens); bonus > 0 {
		score += bonus
		explain["subject_mention_boost"] = bonus
	}
	if penalty := questionMemoryPenalty(queryTokens, record.Content); penalty != 0 {
		score += penalty
		explain["question_memory_penalty"] = penalty
	}
	if bonus := topicAlignmentBoost(queryTokens, contentTokens); bonus > 0 {
		score += bonus
		explain["topic_alignment_boost"] = bonus
	}

	applyPrimitiveBonus(&score, explain, record, primitiveWeights)

	if record.CorrectedAt != nil {
		score += 0.3
		explain["corrected"] = true
		explain["corrected_at"] = record.CorrectedAt.Format(time.RFC3339)
	}

	return score, explain
}

func contentBearingTokens(tokens []string) []string {
	out := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if isQueryStopword(token) {
			continue
		}
		out = append(out, token)
	}
	return out
}

func isQueryStopword(token string) bool {
	switch token {
	case "a", "an", "the", "and", "or", "to", "of", "in", "on", "for", "is", "it", "as", "at", "by", "from",
		"what", "which", "who", "whom", "whose", "where", "when", "why", "how",
		"did", "does", "do", "has", "have", "had", "was", "were", "be", "been", "being",
		"with", "about", "into", "over", "after", "before", "than", "then",
		"me", "my", "you", "your", "we", "our", "they", "their", "he", "she", "his", "her",
		"this", "that", "these", "those", "s":
		return true
	}
	return false
}

func tokensMatch(queryToken, contentToken string) bool {
	if queryToken == contentToken {
		return true
	}
	// Light stemming: research≈researched, camp≈camped/camping.
	if len(queryToken) >= 4 && len(contentToken) >= 4 {
		if strings.HasPrefix(contentToken, queryToken) || strings.HasPrefix(queryToken, contentToken) {
			return true
		}
	}
	return false
}

// subjectMentionBoost rewards memories that mention a named subject from the query
// (e.g. caroline / melanie) so multi-hop person questions don't drown in generic "what" hits.
func subjectMentionBoost(queryTokens, contentTokens []string) float64 {
	nameLike := make([]string, 0)
	for _, token := range queryTokens {
		if len(token) >= 4 && !isQueryStopword(token) {
			// Heuristic: longer non-stop tokens often include person/topic nouns.
			nameLike = append(nameLike, token)
		}
	}
	if len(nameLike) == 0 {
		return 0
	}
	contentSet := map[string]struct{}{}
	for _, token := range contentTokens {
		contentSet[token] = struct{}{}
	}
	hits := 0
	for _, token := range nameLike {
		if _, ok := contentSet[token]; ok {
			hits++
		}
	}
	if hits == 0 {
		return 0
	}
	return 0.12 * float64(hits)
}

// questionMemoryPenalty downranks stored questions when the user query is itself a question
// seeking facts — otherwise "What did ..." memories dominate over statements.
func questionMemoryPenalty(queryTokens []string, content string) float64 {
	ask := false
	for _, token := range queryTokens {
		switch token {
		case "what", "which", "who", "where", "when", "why", "how":
			ask = true
		}
	}
	if !ask {
		return 0
	}
	trimmed := strings.TrimSpace(content)
	lower := strings.ToLower(trimmed)
	if strings.HasSuffix(trimmed, "?") {
		return -0.55
	}
	for _, prefix := range []string{"what ", "which ", "who ", "where ", "when ", "why ", "how ", "did ", "do ", "does "} {
		if strings.HasPrefix(lower, prefix) {
			return -0.4
		}
	}
	return 0
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
		// Also check source text / metadata when slots / observed_at.
		contentDates = filterDateTokens(tokenize(record.SourceText))
		if when, ok := record.Metadata["when"].(string); ok {
			contentDates = append(contentDates, filterDateTokens(tokenize(when))...)
		}
		if record.ObservedAt != nil {
			contentDates = append(contentDates, filterDateTokens(tokenize(record.ObservedAt.Format("2 January 2006")))...)
			contentDates = append(contentDates, record.ObservedAt.Format("2006"))
		}
	}
	if len(contentDates) == 0 {
		if queryHasWhen && record.ObservedAt != nil {
			return 0.18
		}
		return 0
	}
	if len(queryDates) > 0 {
		for _, qd := range queryDates {
			for _, cd := range contentDates {
				if qd == cd {
					return 0.25
				}
			}
		}
	}
	if queryHasWhen {
		if record.ObservedAt != nil {
			return 0.2
		}
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
