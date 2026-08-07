package memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"brainy/internal/embedding"
	"brainy/internal/pack"
)

type Store interface {
	UpsertMemory(ctx context.Context, record MemoryRecord) (StoreUpsertResult, error)
	GetMemory(ctx context.Context, tenantID, subjectID, memoryID string) (MemoryRecord, error)
	ListActiveMemories(ctx context.Context, tenantID, subjectID string) ([]MemoryRecord, error)
	SearchActiveMemories(ctx context.Context, tenantID, subjectID string, patterns []string, limit int) ([]MemoryRecord, error)
	// List/Search with includeSuperseded=true return lifecycle=superseded rows
	// (historical query). Default list/search pass false.
	ListMemories(ctx context.Context, tenantID, subjectID string, includeSuperseded bool) ([]MemoryRecord, error)
	SearchMemories(ctx context.Context, tenantID, subjectID string, patterns []string, limit int, includeSuperseded bool) ([]MemoryRecord, error)
	SuppressMemory(ctx context.Context, tenantID, subjectID, memoryID string) error
	MarkSuperseded(ctx context.Context, tenantID, subjectID, memoryID string) error
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
	// entityRankingEnabled gates entity-overlap retrieval boosting. Extraction/
	// persistence always runs; ranking integration is opt-in until proven
	// non-regressing on same-pin conversational measurement.
	entityRankingEnabled bool
	// idfRankingEnabled gates IDF-weighted lexical coverage (BM25 intuition).
	// Off by default: it regressed same-pin smoke without re-tuning the additive
	// boost stack. Opt-in for staging experiments.
	idfRankingEnabled bool
	hybridReader      HybridReaderConfig
}

// WithIDFRanking toggles IDF-weighted lexical coverage (default off).
func (s *Service) WithIDFRanking(enabled bool) *Service {
	s.idfRankingEnabled = enabled
	return s
}

// WithEntityRanking toggles entity-overlap retrieval boosting (default off).
func (s *Service) WithEntityRanking(enabled bool) *Service {
	s.entityRankingEnabled = enabled
	return s
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
	if err := s.validatePackStateTransition(ctx, req); err != nil {
		return IngestResult{}, err
	}

	// Evidence Plane v2: capture raw messages before extraction.
	ids := s.persistRawEvidence(ctx, req)
	attachEvidenceIDs(&req, ids)

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
		s.persistEntityLinks(ctx, upserted.Record)
		s.persistEvidenceShadow(ctx, upserted.Record)
		s.persistEventIfApplicable(ctx, upserted.Record)
		_ = s.autoSupersedePriorState(ctx, upserted.Record)
		// Project current state only after supersession decisions.
		fresh := upserted.Record
		if got, err := s.store.GetMemory(ctx, fresh.TenantID, fresh.SubjectID, fresh.MemoryID); err == nil {
			fresh = got
		}
		s.projectCurrentStateIfApplicable(ctx, fresh)
		if err := s.applyIngestSupersession(ctx, upserted.Record); err != nil {
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
		s.persistEntityLinks(ctx, upserted.Record)
		s.persistEvidenceShadow(ctx, upserted.Record)
		s.persistEventIfApplicable(ctx, upserted.Record)
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
	if err := validatePackMetadata(s.packs, req); err != nil {
		return AsyncIngestResult{}, err
	}
	if err := s.validatePackStateTransition(ctx, req); err != nil {
		return AsyncIngestResult{}, err
	}

	result := AsyncIngestResult{
		IngestID: s.id("ing"),
		JobID:    s.id("job"),
		Accepted: true,
	}
	// Capture raw evidence before async enrichment so source survives extract failure.
	ids := s.persistRawEvidence(ctx, req)
	attachEvidenceIDs(&req, ids)
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

func (s *Service) GetJob(ctx context.Context, jobID string) (JobStatusInfo, bool, error) {
	q, ok := s.store.(JobQuerier)
	if !ok {
		return JobStatusInfo{}, false, errors.New("job status not supported by store")
	}
	return q.GetExtractionJob(ctx, jobID)
}

func (s *Service) SubjectJobCounts(ctx context.Context, tenantID, subjectID string) (SubjectJobCounts, error) {
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(subjectID) == "" {
		return SubjectJobCounts{}, errors.New("tenant_id and subject_id are required")
	}
	q, ok := s.store.(JobQuerier)
	if !ok {
		return SubjectJobCounts{}, errors.New("job status not supported by store")
	}
	return q.CountSubjectJobs(ctx, tenantID, subjectID)
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
	return s.SearchOpt(ctx, tenantID, subjectID, vertical, scope, query, SearchOptions{})
}

func (s *Service) SearchOpt(ctx context.Context, tenantID, subjectID, vertical, scope, query string, opts SearchOptions) (SearchResponse, error) {
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

	includeSuperseded := opts.IncludeHistorical
	fusionV2 := FusionV2Enabled()
	intents := AnalyzeQueryIntents(query)
	overfetch := CandidateOverfetch(opts.Limit)
	trace := &SearchTrace{
		CandidateOverfetch: overfetch,
		FusionV2:           fusionV2,
		Intents:            intents,
	}

	// Lexical search and dense scoring run in parallel — same signals, lower p95.
	var (
		memories    []MemoryRecord
		lexErr      error
		embedScores map[string]float64
		wg          sync.WaitGroup
	)
	wg.Add(2)
	var lexRanks map[string]float64
	go func() {
		defer wg.Done()
		if ranked, ok := s.store.(RankedSearcher); ok {
			memories, lexRanks, lexErr = ranked.SearchMemoriesRanked(ctx, tenantID, subjectID, patterns, overfetch, includeSuperseded)
		} else {
			memories, lexErr = s.store.SearchMemories(ctx, tenantID, subjectID, patterns, overfetch, includeSuperseded)
		}
	}()
	go func() {
		defer wg.Done()
		queryVector, _ := s.embed(ctx, query)
		embedScores = s.embeddingScores(ctx, tenantID, subjectID, queryVector)
	}()
	wg.Wait()
	if lexErr != nil {
		return SearchResponse{}, lexErr
	}
	trace.LexicalHits = len(memories)

	// Bounded subject corpus for multi-hop / list / thin lexical fill.
	// Prefer ListMemoriesLimited to avoid unbounded hot-path full scans (Phase 1).
	// Name-like tokens alone must not force a 400-row subject scan.
	var allMemories []MemoryRecord
	listQuery := looksListQuery(queryTokens)
	needCorpus := looksMultiHopQuery(queryTokens) ||
		(len(memories) < 10 && hasResponseKeyword(queryTokens)) ||
		listQuery
	if needCorpus {
		listed, err := s.listSubjectCorpus(ctx, tenantID, subjectID, includeSuperseded, 400)
		if err != nil {
			return SearchResponse{}, err
		}
		allMemories = listed
		trace.ListedSubject = true
	}
	// Dense admit: fetch only high-similarity IDs via GetMemory when no corpus.
	if len(embedScores) > 0 && len(allMemories) == 0 {
		for memID, score := range embedScores {
			if score < 0.15 {
				continue
			}
			rec, err := s.store.GetMemory(ctx, tenantID, subjectID, memID)
			if err != nil || (rec.Status != "" && rec.Status != StatusActive) {
				continue
			}
			if !includeSuperseded && !IsLifecycleSearchVisible(rec.LifecycleState) {
				continue
			}
			allMemories = append(allMemories, rec)
			if len(allMemories) >= 48 {
				break
			}
		}
	}

	if len(memories) < 10 && hasResponseKeyword(queryTokens) {
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

	candidates := make(map[string]MemoryRecord, len(memories)+32)
	for _, record := range memories {
		candidates[record.MemoryID] = record
	}
	denseAdmitted := 0
	if len(embedScores) > 0 {
		for _, record := range allMemories {
			if score := embedScores[record.MemoryID]; score >= 0.15 {
				if _, ok := candidates[record.MemoryID]; !ok {
					candidates[record.MemoryID] = record
					denseAdmitted++
				}
			}
		}
	}
	trace.DenseAdmitted = denseAdmitted

	// Mem0-style entity hub: admit memories linked to query entities even when
	// they lack surface-verb overlap (multi-hop attribute recall).
	hubBoosts := s.entityHubBoostMap(ctx, tenantID, subjectID, query)
	if len(hubBoosts) > 0 {
		byID := indexMemoriesByID(allMemories)
		admitted := 0
		for memID := range hubBoosts {
			if _, ok := candidates[memID]; ok {
				continue
			}
			var record MemoryRecord
			var ok bool
			if record, ok = byID[memID]; !ok {
				fetched, err := s.store.GetMemory(ctx, tenantID, subjectID, memID)
				if err != nil {
					continue
				}
				record = fetched
			}
			if admitRecord(candidates, record, includeSuperseded) {
				admitted++
			}
			if admitted >= 24 {
				break
			}
		}
		trace.EntityHubAdmitted = admitted
	}

	// Subject-content bridge: admit content-dense memories that mention a
	// queried person/topic even when they lack the question's surface verbs.
	// List-shaped queries get a larger, diversity-aware admit set so multi-fact
	// answers are not starved by one dense theme.
	if len(allMemories) > 0 {
		subLimit := 20
		if listQuery {
			subLimit = 48
		}
		expandSubjectContentMemories(candidates, contentQueryTokens, allMemories, subLimit, listQuery)
	}
	// Predicate enumeration (W2/W3): admit all atoms for the list intent.
	if listQuery {
		if indexer, ok := s.store.(AtomIndexer); ok {
			if pred := predicateFromListQuery(queryTokens); pred != "" {
				ids, err := indexer.ListAtomMemoryIDs(ctx, tenantID, subjectID, pred, "", 40)
				if err == nil && len(ids) > 0 {
					byID := indexMemoriesByID(allMemories)
					atomAdmitted := 0
					for _, id := range ids {
						if _, ok := candidates[id]; ok {
							continue
						}
						var record MemoryRecord
						var ok bool
						if record, ok = byID[id]; !ok {
							fetched, err := s.store.GetMemory(ctx, tenantID, subjectID, id)
							if err != nil {
								continue
							}
							record = fetched
						}
						if admitRecord(candidates, record, includeSuperseded) {
							atomAdmitted++
						}
					}
					trace.AtomScanAdmitted = atomAdmitted
				}
			}
		}
	}

	// For multi-hop shaped questions, admit a capped set of same-session neighbors
	// and a second-pass of fact-like memories related to first-hit content tokens.
	if looksMultiHopQuery(queryTokens) {
		if len(allMemories) > 0 {
			expandSessionNeighbors(candidates, memories, allMemories, 16)
		}
		// Seed the related-fact pass from lexical hits AND subject-bridge admits
		// so supporting facts without the question verbs still expand.
		relatedSeeds := make([]MemoryRecord, 0, len(memories)+16)
		relatedSeeds = append(relatedSeeds, memories...)
		for _, record := range candidates {
			content := strings.TrimSpace(record.Content)
			if content == "" || strings.HasSuffix(content, "?") {
				continue
			}
			if len(contentBearingTokens(tokenize(content))) < 5 {
				continue
			}
			relatedSeeds = append(relatedSeeds, record)
			if len(relatedSeeds) >= 40 {
				break
			}
		}
		relLimit := 12
		if listQuery {
			relLimit = 24
		}
		if related, err := s.relatedFactMemories(ctx, tenantID, subjectID, queryTokens, relatedSeeds, relLimit); err == nil {
			for _, record := range related {
				if _, ok := candidates[record.MemoryID]; !ok {
					candidates[record.MemoryID] = record
				}
			}
		}
	}

	// Entity linking: entities are extracted and persisted on ingest (used for
	// provenance and the planned graph layer). Applying entity overlap as a
	// retrieval boost/recall-expander regressed conversational ranking in
	// same-pin smoke measurement (distinctive-entity mentions are not answers),
	// so the ranking integration is gated off by default until a version proves
	// non-regressing. See entityRankingEnabled.
	var distinctiveQueryEntities []string
	var entityDF map[string]int
	var totalMemories int
	if s.entityRankingEnabled {
		queryEntities := ExtractEntities(query)
		entityDF, totalMemories = s.entityDocFrequencies(ctx, tenantID, subjectID)
		for _, e := range queryEntities {
			if isDistinctiveEntity(e, entityDF, totalMemories) {
				distinctiveQueryEntities = append(distinctiveQueryEntities, e)
			}
		}
	}

	var packWeights map[string]int
	if p, ok := s.packs.Get(vertical); ok {
		packWeights = p.RankPolicy.PrimitiveWeights
	}

	// IDF weights over the subject's corpus so distinctive query terms dominate
	// lexical scoring (BM25 intuition). Gated: same-pin smoke showed it regresses
	// without re-tuning the additive-boost stack, so it is opt-in
	// (BRAINY_IDF_RANKING) pending a staging re-tune. Computed once per query.
	var idf map[string]float64
	if s.idfRankingEnabled {
		corpus := allMemories
		if len(corpus) == 0 {
			if listed, err := s.store.ListMemories(ctx, tenantID, subjectID, includeSuperseded); err == nil {
				corpus = listed
			}
		}
		if len(corpus) > 0 {
			idf = computeQueryIDF(corpus, contentBearingTokens(queryTokens))
		}
	}

	ranked := make([]rankedSearchResult, 0, len(candidates))
	entitiesByID := make(map[string][]string, len(candidates))
	for _, record := range candidates {
		if record.Status != "" && record.Status != StatusActive {
			continue
		}
		if includeSuperseded {
			if record.LifecycleState == LifecycleArchived || record.LifecycleState == LifecycleSuppressed {
				continue
			}
		} else if !IsLifecycleSearchVisible(record.LifecycleState) {
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
		score, explain := scoreMemoryIDF(record, queryTokens, packWeights, idf)
		if explain == nil {
			explain = map[string]any{}
		}
		// Calibrated semantic + Mem0-style entity-hub boost.
		embedScore := embedScores[record.MemoryID]
		hub := 0.0
		if hubBoosts != nil {
			hub = hubBoosts[record.MemoryID]
		}
		if fusionV2 {
			// Prefer real FTS rank when the ranked searcher returned lexRanks.
			// Do not pretend token coverage is BM25 when FTS ranks exist.
			bm25 := 0.0
			if lexRanks != nil {
				if r, ok := lexRanks[record.MemoryID]; ok {
					bm25 = r
				}
				if bm25 > 1 {
					bm25 = NormalizeBM25Sigmoid(bm25, len(contentQueryTokens))
				}
				explain["lexical_channel"] = "fts"
			} else {
				if cov, ok := explain["coverage"].(float64); ok {
					bm25 = cov
				} else if score > 0 {
					bm25 = math.Min(score/3.0, 1.0)
				}
				explain["lexical_channel"] = "coverage"
			}
			semOnlyFloor := 0.42
			if len(contentQueryTokens) <= 1 {
				// Short entity-probe queries: block template false-friends.
				semOnlyFloor = 0.78
			}
			combined, parts := ScoreAndRankV2(embedScore, bm25, hub, 0.12, semOnlyFloor)
			if combined <= 0 && score <= 0 {
				continue
			}
			if combined > 0 {
				// Rescale into Brainy score units so downstream boosts stay calibrated.
				score = math.Max(score, 1.0) * (0.35 + 0.65*combined)
				for k, v := range parts {
					explain["signal_"+k] = v
				}
				explain["fusion"] = "v2"
			} else if score <= 0 {
				continue
			}
		} else if score > 0 || embedScore >= 0.15 || hub > 0 {
			if score <= 0 && embedScore >= 0.15 {
				score = embedScore * 0.9
				explain["ranking_basis"] = "hybrid_embedding"
			}
			combined, parts := combineRetrievalSignals(score, embedScore, hub)
			score = combined
			for k, v := range parts {
				explain["signal_"+k] = v
			}
			if hub > 0 {
				explain["entity_hub_boost"] = hub
			}
			if embedScore > 0 {
				explain["embedding_similarity"] = embedScore
			}
		}
		if s.entityRankingEnabled {
			if bonus := entityOverlapBoost(distinctiveQueryEntities, recordEntities(record)); bonus > 0 {
				score += bonus
				explain["entity_overlap_boost"] = bonus
			}
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
		if s.entityRankingEnabled {
			entitiesByID[record.MemoryID] = recordEntities(record)
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

	if s.entityRankingEnabled {
		propagateEntityGraph(ranked, entitiesByID, entityDF, totalMemories)
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

	// List / multi-hop: evidence-set selection (coverage-aware) instead of flat top-k.
	if listQuery || looksMultiHopQuery(queryTokens) {
		limit := opts.Limit
		if limit <= 0 {
			limit = 32
		}
		ranked = selectEvidenceSet(ranked, limit)
	} else if opts.Limit > 0 && len(ranked) > opts.Limit {
		ranked = ranked[:opts.Limit]
	}

	results := make([]SearchResult, len(ranked))
	for i, item := range ranked {
		results[i] = item.result
	}

	return SearchResponse{Results: results, Trace: trace}, nil
}

func admitRecord(candidates map[string]MemoryRecord, record MemoryRecord, includeSuperseded bool) bool {
	if record.MemoryID == "" {
		return false
	}
	if _, ok := candidates[record.MemoryID]; ok {
		return false
	}
	if record.Status != "" && record.Status != StatusActive {
		return false
	}
	if includeSuperseded {
		if record.LifecycleState == LifecycleArchived || record.LifecycleState == LifecycleSuppressed {
			return false
		}
	} else if !IsLifecycleSearchVisible(record.LifecycleState) {
		return false
	}
	candidates[record.MemoryID] = record
	return true
}

func indexMemoriesByID(records []MemoryRecord) map[string]MemoryRecord {
	out := make(map[string]MemoryRecord, len(records))
	for _, r := range records {
		out[r.MemoryID] = r
	}
	return out
}

// LimitedSubjectLister caps hot-path subject scans (Phase 1 / MEM-015).
type LimitedSubjectLister interface {
	ListMemoriesLimited(ctx context.Context, tenantID, subjectID string, includeSuperseded bool, limit int) ([]MemoryRecord, error)
}

func (s *Service) listSubjectCorpus(ctx context.Context, tenantID, subjectID string, includeSuperseded bool, limit int) ([]MemoryRecord, error) {
	if limit <= 0 {
		limit = 400
	}
	if limited, ok := s.store.(LimitedSubjectLister); ok {
		return limited.ListMemoriesLimited(ctx, tenantID, subjectID, includeSuperseded, limit)
	}
	all, err := s.store.ListMemories(ctx, tenantID, subjectID, includeSuperseded)
	if err != nil {
		return nil, err
	}
	if len(all) > limit {
		return all[:limit], nil
	}
	return all, nil
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

// looksMultiHopQuery detects questions that typically need multiple supporting
// memories. Kept strict to avoid ordinary wh-questions triggering subject scans.
func looksMultiHopQuery(tokens []string) bool {
	hasAsk := false
	for _, token := range tokens {
		switch token {
		case "what", "which", "who", "where", "how", "when", "why":
			hasAsk = true
		}
	}
	if !hasAsk {
		return false
	}
	bearing := contentBearingTokens(tokens)
	// Require at least three content-bearing tokens (name-like alone is too broad).
	return len(bearing) >= 3
}

// nameLikeTokens returns longer non-stop content tokens that often name people
// or topics in conversational queries (heuristic, language-agnostic-ish).
func nameLikeTokens(tokens []string) []string {
	out := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if len(token) >= 4 && !isQueryStopword(token) {
			out = append(out, token)
		}
	}
	return out
}

// looksListQuery detects questions that ask for multiple supporting items
// (activities, books, places, likes). Generic conversational cues only —
// keep this narrow so recency/preference ranking is not disturbed.
func looksListQuery(tokens []string) bool {
	for _, token := range tokens {
		switch token {
		case "activities", "activity", "hobbies", "hobby", "books", "book",
			"places", "place", "stress", "camping", "camped":
			return true
		}
	}
	return false
}

func predicateFromListQuery(tokens []string) string {
	for _, token := range tokens {
		switch token {
		case "activities", "activity", "hobbies", "hobby", "stress":
			return PredicateActivity
		case "books", "book", "read", "reading":
			return PredicateMediaConsumed
		case "camped", "camping", "places", "place":
			return PredicateEvent
		case "kids", "children":
			return PredicateFamilyMember
		}
	}
	return ""
}

// expandSubjectContentMemories admits content-dense memories that mention a
// queried name/topic so profile and multi-fact questions are not drowned out by
// short acknowledgment turns that only share the name token.
// When diversify is true, admits by novel content tokens (MMR) instead of density alone.
func expandSubjectContentMemories(candidates map[string]MemoryRecord, queryTokens []string, all []MemoryRecord, limit int, diversify bool) {
	subjects := nameLikeTokens(queryTokens)
	if len(subjects) == 0 || limit <= 0 {
		return
	}
	subjectSet := map[string]struct{}{}
	for _, s := range subjects {
		subjectSet[s] = struct{}{}
	}
	type ranked struct {
		record MemoryRecord
		dens   int
		toks   []string
	}
	pool := make([]ranked, 0, limit*2)
	for _, record := range all {
		if _, exists := candidates[record.MemoryID]; exists {
			continue
		}
		content := strings.TrimSpace(record.Content)
		if content == "" || strings.HasSuffix(content, "?") {
			continue
		}
		bearing := contentBearingTokens(tokenize(content))
		if len(bearing) < 5 {
			continue
		}
		hit := false
		for _, tok := range bearing {
			if _, ok := subjectSet[tok]; ok {
				hit = true
				break
			}
		}
		if !hit {
			continue
		}
		pool = append(pool, ranked{record: record, dens: len(bearing), toks: bearing})
	}
	sort.SliceStable(pool, func(i, j int) bool {
		if pool[i].dens == pool[j].dens {
			return pool[i].record.MemoryID > pool[j].record.MemoryID
		}
		return pool[i].dens > pool[j].dens
	})
	if !diversify {
		for i, item := range pool {
			if i >= limit {
				break
			}
			candidates[item.record.MemoryID] = item.record
		}
		return
	}
	covered := map[string]struct{}{}
	added := 0
	used := map[string]struct{}{}
	for added < limit && len(used) < len(pool) {
		best := -1
		bestGain := -1
		for i, item := range pool {
			if _, ok := used[item.record.MemoryID]; ok {
				continue
			}
			novel := 0
			for _, tok := range item.toks {
				if len(tok) < 4 {
					continue
				}
				if _, ok := covered[tok]; !ok {
					novel++
				}
			}
			gain := novel*10 + item.dens
			if best < 0 || gain > bestGain {
				best = i
				bestGain = gain
			}
		}
		if best < 0 {
			break
		}
		item := pool[best]
		used[item.record.MemoryID] = struct{}{}
		candidates[item.record.MemoryID] = item.record
		added++
		for _, tok := range item.toks {
			if len(tok) >= 4 {
				covered[tok] = struct{}{}
			}
		}
	}
}

// diversifyByContentTokens reorders ranked results so the head of the list covers
// distinct content-bearing tokens (MMR-style). Remaining items keep score order.
func diversifyByContentTokens(ranked []rankedSearchResult, keep int) []rankedSearchResult {
	if len(ranked) <= 1 {
		return ranked
	}
	if keep <= 0 || keep > len(ranked) {
		keep = len(ranked)
	}
	selected := make([]rankedSearchResult, 0, len(ranked))
	used := make(map[string]struct{}, keep)
	covered := map[string]struct{}{}
	for len(selected) < keep {
		best := -1
		bestGain := -1.0
		for i, item := range ranked {
			if _, ok := used[item.result.MemoryID]; ok {
				continue
			}
			novel := 0
			for _, tok := range contentBearingTokens(tokenize(item.result.Content)) {
				if len(tok) < 4 {
					continue
				}
				if _, ok := covered[tok]; !ok {
					novel++
				}
			}
			gain := float64(novel) + 0.02*item.result.Score
			if best < 0 || gain > bestGain {
				best = i
				bestGain = gain
			}
		}
		if best < 0 {
			break
		}
		item := ranked[best]
		used[item.result.MemoryID] = struct{}{}
		selected = append(selected, item)
		if explain := item.result.Explain; explain != nil {
			explain["content_diversity"] = true
		}
		for _, tok := range contentBearingTokens(tokenize(item.result.Content)) {
			if len(tok) >= 4 {
				covered[tok] = struct{}{}
			}
		}
	}
	for _, item := range ranked {
		if _, ok := used[item.result.MemoryID]; ok {
			continue
		}
		selected = append(selected, item)
	}
	return selected
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
	found, err := s.store.SearchMemories(ctx, tenantID, subjectID, patterns, 50, false)
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
		case "activities", "activity", "hobby", "hobbies":
			return []string{"hobby", "hobbies", "activity", "activities"}
		case "camping", "camp":
			return []string{"camp", "camping", "camped"}
		case "books", "read", "reading":
			return []string{"reading", "book", "books", "library"}
		case "stress", "relax":
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
	s.persistEntityLinks(ctx, record)

	return MutationResult{
		MemoryID: record.MemoryID,
		Kind:     record.Kind,
		Content:  record.Content,
		Status:   record.Status,
	}, nil
}

// Supersede creates a new active memory that replaces priorID, then marks the
// prior record lifecycle=superseded. Search excludes the old record by default.
func (s *Service) Supersede(ctx context.Context, tenantID, subjectID, priorID string, req SupersedeRequest) (MutationResult, error) {
	if tenantID == "" || subjectID == "" || priorID == "" {
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

	prior, err := s.store.GetMemory(ctx, tenantID, subjectID, priorID)
	if err != nil {
		return MutationResult{}, err
	}
	if !IsLifecycleSearchVisible(prior.LifecycleState) && prior.LifecycleState != LifecycleActive {
		// Allow superseding deprioritized; reject already-terminal states.
		if prior.LifecycleState == LifecycleSuperseded || prior.LifecycleState == LifecycleArchived || prior.LifecycleState == LifecycleSuppressed || prior.Status == StatusSuppressed {
			return MutationResult{}, fmt.Errorf("memory %s is not active (lifecycle=%s status=%s)", priorID, prior.LifecycleState, prior.Status)
		}
	}

	now := s.now()
	replacement := prior
	replacement.MemoryID = s.id("mem")
	replacement.Content = content
	replacement.SourceText = sourceText
	replacement.DedupeKey = DedupeKey(tenantID, subjectID, prior.Kind, content)
	replacement.Status = StatusActive
	replacement.LifecycleState = LifecycleActive
	replacement.SupersedesID = priorID
	replacement.SupersededAt = nil
	replacement.CorrectedAt = &now
	replacement.CreatedAt = now
	replacement.UpdatedAt = now
	if replacement.Explain == nil {
		replacement.Explain = map[string]any{}
	}
	replacement.Explain["supersedes"] = priorID

	upserted, err := s.store.UpsertMemory(ctx, replacement)
	if err != nil {
		return MutationResult{}, err
	}
	s.persistEmbedding(ctx, upserted.Record)
	s.persistEntityLinks(ctx, upserted.Record)

	if err := s.store.MarkSuperseded(ctx, tenantID, subjectID, priorID); err != nil {
		return MutationResult{}, err
	}

	return MutationResult{
		MemoryID: upserted.Record.MemoryID,
		Kind:     upserted.Record.Kind,
		Content:  upserted.Record.Content,
		Status:   upserted.Record.Status,
	}, nil
}

// ApplyDomainEvent marks listed memories superseded (batch invalidation).
// With Match set, selects active memories by label/kind/metadata (pack-style
// triggers without requiring callers to know memory IDs).
func (s *Service) ApplyDomainEvent(ctx context.Context, req DomainEventRequest) (DomainEventResult, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.SubjectID) == "" {
		return DomainEventResult{}, errors.New("tenant_id and subject_id are required")
	}
	if strings.TrimSpace(req.EventType) == "" {
		return DomainEventResult{}, errors.New("event_type is required")
	}
	ids := append([]string{}, req.SupersedeMemoryIDs...)
	if req.Match != nil {
		matched, err := s.matchMemoriesForEvent(ctx, req.TenantID, req.SubjectID, req.Match)
		if err != nil {
			return DomainEventResult{}, err
		}
		ids = append(ids, matched...)
	}
	out := DomainEventResult{EventType: req.EventType, Superseded: make([]string, 0, len(ids))}
	seen := map[string]struct{}{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		if err := s.store.MarkSuperseded(ctx, req.TenantID, req.SubjectID, id); err != nil {
			if errors.Is(err, ErrMemoryNotFound) {
				continue
			}
			return DomainEventResult{}, err
		}
		out.Superseded = append(out.Superseded, id)
	}
	return out, nil
}

func (s *Service) matchMemoriesForEvent(ctx context.Context, tenantID, subjectID string, match *DomainEventMatch) ([]string, error) {
	if match == nil {
		return nil, nil
	}
	all, err := s.store.ListMemories(ctx, tenantID, subjectID, false)
	if err != nil {
		return nil, err
	}
	wantLabel := strings.TrimSpace(match.Label)
	wantKind := strings.TrimSpace(match.Kind)
	out := make([]string, 0)
	for _, record := range all {
		if wantLabel != "" && record.Label != wantLabel {
			continue
		}
		if wantKind != "" && record.Kind != wantKind {
			continue
		}
		ok := true
		for key, want := range match.Metadata {
			got := ""
			if record.Metadata != nil {
				if raw, exists := record.Metadata[key]; exists && raw != nil {
					got = strings.TrimSpace(fmt.Sprint(raw))
				}
			}
			if got != want {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, record.MemoryID)
		}
	}
	return out, nil
}

// applyIngestSupersession honors metadata.supersedes_memory_id on a newly
// written record: mark the prior memory superseded and ensure lineage is set.
func (s *Service) applyIngestSupersession(ctx context.Context, record MemoryRecord) error {
	priorID := supersedesMemoryIDFromMetadata(record.Metadata)
	if priorID == "" {
		priorID = strings.TrimSpace(record.SupersedesID)
	}
	if priorID == "" || priorID == record.MemoryID {
		return nil
	}
	return s.store.MarkSuperseded(ctx, record.TenantID, record.SubjectID, priorID)
}

// autoSupersedePriorState marks older same-(subject,predicate) state atoms
// superseded when a newer atom arrives (master-plan W5).
func (s *Service) autoSupersedePriorState(ctx context.Context, record MemoryRecord) error {
	return AutoSupersedePriorState(ctx, s.store, record)
}

// AutoSupersedePriorState marks older same-(subject,predicate) state atoms
// superseded when a newer atom arrives (shared by sync + async).
func AutoSupersedePriorState(ctx context.Context, store Store, record MemoryRecord) error {
	if store == nil || record.Metadata == nil {
		return nil
	}
	pred, _ := record.Metadata["predicate"].(string)
	val, _ := record.Metadata["value_norm"].(string)
	if pred == "" || val == "" {
		return nil
	}
	// Only auto-supersede stateful predicates (not events/media lists).
	if !IsStatefulPredicate(pred) {
		return nil
	}
	indexer, ok := store.(AtomIndexer)
	if !ok {
		return nil
	}
	ids, err := indexer.ListAtomMemoryIDs(ctx, record.TenantID, record.SubjectID, pred, "", 20)
	if err != nil {
		return nil
	}
	for _, id := range ids {
		if id == record.MemoryID {
			continue
		}
		prior, err := store.GetMemory(ctx, record.TenantID, record.SubjectID, id)
		if err != nil {
			continue
		}
		// Same predicate, different value → supersede older.
		pval, _ := prior.Metadata["value_norm"].(string)
		if pval == "" || pval == val {
			continue
		}
		if prior.ObservedAt != nil && record.ObservedAt != nil && !record.ObservedAt.After(*prior.ObservedAt) {
			continue
		}
		_ = store.MarkSuperseded(ctx, record.TenantID, record.SubjectID, id)
		if retirer, ok := indexer.(AtomRetirer); ok {
			_ = retirer.RetireMemoryAtom(ctx, record.TenantID, record.SubjectID, pred, pval, id, record.ObservedAt)
		}
	}
	return nil
}

func supersedesMemoryIDFromMetadata(metadata map[string]any) string {
	if metadata == nil {
		return ""
	}
	raw, ok := metadata["supersedes_memory_id"]
	if !ok {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func scoreMemory(record MemoryRecord, queryTokens []string, primitiveWeights map[string]int) (float64, map[string]any) {
	return scoreMemoryIDF(record, queryTokens, primitiveWeights, nil)
}

// coverageScore returns matched/total query coverage, IDF-weighted when idf is
// available so distinctive terms dominate over common ones (BM25 intuition).
// Result is normalized to [0,1] to preserve downstream boost calibration.
func coverageScore(matched, bearingQuery []string, idf map[string]float64) float64 {
	if len(bearingQuery) == 0 {
		return 0
	}
	if len(idf) == 0 {
		return float64(len(matched)) / float64(len(bearingQuery))
	}
	var total, hit float64
	for _, t := range bearingQuery {
		w := idf[t]
		if w <= 0 {
			w = 1.0 // unseen term: neutral weight
		}
		total += w
	}
	for _, t := range matched {
		w := idf[t]
		if w <= 0 {
			w = 1.0
		}
		hit += w
	}
	if total <= 0 {
		return float64(len(matched)) / float64(len(bearingQuery))
	}
	return hit / total
}

// computeQueryIDF returns inverse document frequency for each content-bearing
// query term over the subject's active memories. idf = log(1 + N/(1+df)).
func computeQueryIDF(all []MemoryRecord, bearingQuery []string) map[string]float64 {
	if len(all) == 0 || len(bearingQuery) == 0 {
		return nil
	}
	n := float64(len(all))
	df := make(map[string]int, len(bearingQuery))
	for _, record := range all {
		contentTokens := tokenize(record.Content)
		seen := map[string]struct{}{}
		for _, ct := range contentTokens {
			seen[ct] = struct{}{}
		}
		for _, q := range bearingQuery {
			if _, done := df[q]; false {
				_ = done
			}
			for ct := range seen {
				if tokensMatch(q, ct) {
					df[q]++
					break
				}
			}
		}
	}
	idf := make(map[string]float64, len(bearingQuery))
	for _, q := range bearingQuery {
		idf[q] = math.Log(1.0 + n/(1.0+float64(df[q])))
	}
	return idf
}

// scoreMemoryIDF scores a memory against the query. When idf is provided, the
// base lexical coverage is IDF-weighted (BM25-style): matching rare/distinctive
// query terms counts more than common ones. idf==nil falls back to plain
// match-count coverage (used by direct unit tests / no-corpus paths).
func scoreMemoryIDF(record MemoryRecord, queryTokens []string, primitiveWeights map[string]int, idf map[string]float64) (float64, map[string]any) {
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

	score := coverageScore(matched, bearingQuery, idf)
	explain := map[string]any{
		"matched_terms": matched,
		"ranking_basis": "deterministic_baseline",
		"coverage":      score,
	}
	if len(idf) > 0 {
		explain["ranking_basis"] = "idf_weighted"
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
	if penalty := lowInformationPenalty(record.Content, matched, bearingQuery); penalty != 0 {
		score += penalty
		explain["low_information_penalty"] = penalty
	}
	if bonus := attributeAtomBoost(record); bonus > 0 {
		score += bonus
		explain["attribute_atom_boost"] = bonus
	}
	if bonus := queryAttributeIntentBoost(queryTokens, record); bonus > 0 {
		score += bonus
		explain["attribute_intent_boost"] = bonus
	}
	// Broken quote shards ("Name mentioned \"m still…\"") must not dominate.
	lowerContent := strings.ToLower(record.Content)
	if strings.Contains(lowerContent, " mentioned \"") &&
		(strings.Contains(lowerContent, " but i") || strings.Contains(lowerContent, "mentioned \"m ") ||
			strings.Contains(lowerContent, "mentioned \"i ")) {
		score -= 0.7
		explain["broken_quote_penalty"] = -0.7
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

// subjectMentionBoost rewards content-dense memories that mention a named
// subject from the query. Short acknowledgments ("Yeah, Alice") do not qualify —
// they previously flooded person-centric recall.
func subjectMentionBoost(queryTokens, contentTokens []string) float64 {
	nameLike := nameLikeTokens(queryTokens)
	if len(nameLike) == 0 {
		return 0
	}
	bearingContent := contentBearingTokens(contentTokens)
	if len(bearingContent) < 5 {
		return 0
	}
	contentSet := map[string]struct{}{}
	for _, token := range bearingContent {
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

// attributeAtomBoost prefers deterministic/provider atomic facts so
// multi-hop attribute questions (identity, activities, titles) surface.
func attributeAtomBoost(record MemoryRecord) float64 {
	if record.Explain == nil {
		return 0
	}
	rule, _ := record.Explain["rule"].(string)
	switch {
	case rule == "attribute_identity" || rule == "attribute_relationship" || rule == "attribute_origin":
		return 0.4
	case strings.HasPrefix(rule, "attribute_"):
		return 0.28
	case rule == "provider_extract":
		return 0.12
	}
	return 0
}

// queryAttributeIntentBoost aligns atom type with question intent
// (moved-from → origin atoms, relationship status → relationship atoms, etc.).
func queryAttributeIntentBoost(queryTokens []string, record MemoryRecord) float64 {
	if record.Explain == nil {
		return 0
	}
	rule, _ := record.Explain["rule"].(string)
	if rule == "" {
		return 0
	}
	qset := map[string]struct{}{}
	for _, t := range queryTokens {
		qset[t] = struct{}{}
	}
	has := func(words ...string) bool {
		for _, w := range words {
			if _, ok := qset[w]; ok {
				return true
			}
		}
		return false
	}
	switch {
	case rule == "attribute_origin" && has("moved", "from", "where", "country", "live", "lived"):
		return 0.55
	case rule == "attribute_relationship" && has("relationship", "status", "single", "married", "partner"):
		return 0.55
	case rule == "attribute_identity" && has("identity", "who", "gender"):
		return 0.45
	case (rule == "attribute_activity" || rule == "attribute_place_activity") &&
		has("activities", "activity", "hobbies", "hobby", "camping", "camped", "stress"):
		return 0.45
	case rule == "attribute_titled_work" && has("books", "book", "read", "reading"):
		return 0.5
	case rule == "attribute_preference" && has("kids", "children", "like", "likes"):
		return 0.5
	}
	// Content fallback when explain.rule missing (older rows / provider atoms).
	content := strings.ToLower(record.Content)
	if has("moved", "from") && (strings.Contains(content, " moved from ") || strings.Contains(content, " is from ")) {
		return 0.45
	}
	if has("relationship", "status", "single") && strings.Contains(content, " single") {
		return 0.45
	}
	return 0
}

// lowInformationPenalty downranks greeting/ack turns and name-only matches so
// fact-bearing memories surface for person and multi-fact questions.
func lowInformationPenalty(content string, matched, bearingQuery []string) float64 {
	bearing := contentBearingTokens(tokenize(content))
	if len(bearing) == 0 {
		return -0.8
	}
	if len(bearing) <= 2 {
		return -0.75
	}
	if len(bearing) <= 4 {
		return -0.4
	}
	if len(matched) == 0 {
		return 0
	}
	nameSet := map[string]struct{}{}
	for _, token := range nameLikeTokens(bearingQuery) {
		nameSet[token] = struct{}{}
	}
	if len(nameSet) == 0 {
		return 0
	}
	onlyNames := true
	for _, m := range matched {
		if _, ok := nameSet[m]; !ok {
			onlyNames = false
			break
		}
	}
	if onlyNames && len(bearing) < 8 {
		return -0.45
	}
	return 0
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
