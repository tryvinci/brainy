package memory

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"brainy/internal/pack"
)

type memoryStoreStub struct {
	records     map[string]MemoryRecord
	jobs        map[string]ExtractionJob
	entityLinks map[string][]string
}

func newMemoryStoreStub() *memoryStoreStub {
	return &memoryStoreStub{
		records:     map[string]MemoryRecord{},
		jobs:        map[string]ExtractionJob{},
		entityLinks: map[string][]string{},
	}
}

func (s *memoryStoreStub) UpsertMemory(_ context.Context, record MemoryRecord) (StoreUpsertResult, error) {
	if existing, ok := s.records[record.DedupeKey]; ok {
		if existing.Status == StatusSuppressed {
			return StoreUpsertResult{Record: existing, State: "deduped"}, nil
		}
		if existing.Content == record.Content && existing.Status == StatusActive &&
			metadataEqual(existing.Metadata, record.Metadata) &&
			existing.LifecycleState == record.LifecycleState &&
			existing.Label == record.Label &&
			existing.Scope == record.Scope {
			return StoreUpsertResult{Record: existing, State: "deduped"}, nil
		}
		record.MemoryID = existing.MemoryID
		record.CreatedAt = existing.CreatedAt
		s.records[record.DedupeKey] = record
		return StoreUpsertResult{Record: record, State: "updated"}, nil
	}
	s.records[record.DedupeKey] = record
	return StoreUpsertResult{Record: record, State: "created"}, nil
}

func metadataEqual(left, right map[string]any) bool {
	if len(left) != len(right) {
		return false
	}
	for key, leftValue := range left {
		rightValue, ok := right[key]
		if !ok {
			return false
		}
		if fmt.Sprint(leftValue) != fmt.Sprint(rightValue) {
			return false
		}
	}
	return true
}

func (s *memoryStoreStub) ListActiveMemories(ctx context.Context, tenantID, subjectID string) ([]MemoryRecord, error) {
	return s.ListMemories(ctx, tenantID, subjectID, false)
}

func (s *memoryStoreStub) ListMemories(_ context.Context, tenantID, subjectID string, includeSuperseded bool) ([]MemoryRecord, error) {
	var out []MemoryRecord
	for _, record := range s.records {
		if record.TenantID != tenantID || record.SubjectID != subjectID || record.Status != StatusActive {
			continue
		}
		if includeSuperseded {
			if record.LifecycleState == LifecycleArchived || record.LifecycleState == LifecycleSuppressed {
				continue
			}
		} else if !IsLifecycleSearchVisible(record.LifecycleState) {
			continue
		}
		out = append(out, record)
	}
	return out, nil
}

func (s *memoryStoreStub) SearchActiveMemories(ctx context.Context, tenantID, subjectID string, patterns []string, limit int) ([]MemoryRecord, error) {
	return s.SearchMemories(ctx, tenantID, subjectID, patterns, limit, false)
}

func (s *memoryStoreStub) SearchMemories(ctx context.Context, tenantID, subjectID string, patterns []string, limit int, includeSuperseded bool) ([]MemoryRecord, error) {
	_ = patterns
	_ = limit
	return s.ListMemories(ctx, tenantID, subjectID, includeSuperseded)
}

func (s *memoryStoreStub) GetMemory(_ context.Context, tenantID, subjectID, memoryID string) (MemoryRecord, error) {
	for _, record := range s.records {
		if record.TenantID == tenantID && record.SubjectID == subjectID && record.MemoryID == memoryID {
			return record, nil
		}
	}
	return MemoryRecord{}, ErrMemoryNotFound
}

func (s *memoryStoreStub) LinkMemoryEntities(_ context.Context, _, _, memoryID string, entities []string) error {
	if s.entityLinks == nil {
		s.entityLinks = map[string][]string{}
	}
	for _, e := range entities {
		s.entityLinks[e] = appendUnique(s.entityLinks[e], memoryID)
	}
	return nil
}

func (s *memoryStoreStub) EntityHubBoosts(_ context.Context, _, _ string, queryEntities []string) (map[string]float64, error) {
	out := map[string]float64{}
	for _, e := range queryEntities {
		ids := s.entityLinks[e]
		if len(ids) == 0 || len(ids) > 40 {
			continue
		}
		w := 0.5 / float64(len(ids))
		if w > 0.35 {
			w = 0.35
		}
		for _, id := range ids {
			out[id] += w
			if out[id] > 0.5 {
				out[id] = 0.5
			}
		}
	}
	return out, nil
}

func (s *memoryStoreStub) MarkSuperseded(_ context.Context, tenantID, subjectID, memoryID string) error {
	for key, record := range s.records {
		if record.TenantID == tenantID && record.SubjectID == subjectID && record.MemoryID == memoryID {
			now := time.Now().UTC()
			record.LifecycleState = LifecycleSuperseded
			record.SupersededAt = &now
			record.UpdatedAt = now
			s.records[key] = record
			return nil
		}
	}
	return ErrMemoryNotFound
}

func (s *memoryStoreStub) SuppressMemory(_ context.Context, tenantID, subjectID, memoryID string) error {
	for key, record := range s.records {
		if record.TenantID == tenantID && record.SubjectID == subjectID && record.MemoryID == memoryID {
			record.Status = StatusSuppressed
			record.UpdatedAt = time.Now().UTC()
			s.records[key] = record
			return nil
		}
	}
	return nil
}

func (s *memoryStoreStub) CorrectMemory(_ context.Context, tenantID, subjectID, memoryID, content, sourceText string) (MemoryRecord, error) {
	for key, record := range s.records {
		if record.TenantID == tenantID && record.SubjectID == subjectID && record.MemoryID == memoryID {
			delete(s.records, key)
			record.Content = content
			record.SourceText = sourceText
			record.DedupeKey = DedupeKey(tenantID, subjectID, record.Kind, content)
			record.Status = StatusActive
			record.UpdatedAt = time.Now().UTC()
			s.records[record.DedupeKey] = record
			return record, nil
		}
	}
	return MemoryRecord{}, errors.New("memory not found")
}

func (s *memoryStoreStub) EnqueueIngestJob(_ context.Context, ingestID, jobID, _ string, req IngestRequest) (EnqueueResult, error) {
	s.jobs[jobID] = ExtractionJob{JobID: jobID, IngestID: ingestID, Request: req}
	return EnqueueResult{IngestID: ingestID, JobID: jobID, Accepted: true}, nil
}

func (s *memoryStoreStub) ClaimNextExtractionJob(_ context.Context) (ExtractionJob, bool, error) {
	for jobID, job := range s.jobs {
		delete(s.jobs, jobID)
		return job, true, nil
	}
	return ExtractionJob{}, false, nil
}

func (s *memoryStoreStub) CompleteExtractionJob(_ context.Context, _, _ string) error {
	return nil
}

func (s *memoryStoreStub) FailExtractionJob(_ context.Context, _, _, _ string) error {
	return nil
}

func TestServiceIngestSearchAndSuppress(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)

	result, err := service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "I prefer concise, direct answers."},
		},
	})
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if result.Created != 1 {
		t.Fatalf("expected 1 created memory, got %d", result.Created)
	}

	search, err := service.Search(context.Background(), "t1", "u1", "", "", "How should I respond?")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(search.Results) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(search.Results))
	}

	if err := service.Suppress(context.Background(), "t1", "u1", result.Memories[0].MemoryID); err != nil {
		t.Fatalf("suppress failed: %v", err)
	}

	searchAfterSuppress, err := service.Search(context.Background(), "t1", "u1", "", "", "How should I respond?")
	if err != nil {
		t.Fatalf("search after suppress failed: %v", err)
	}
	if len(searchAfterSuppress.Results) != 0 {
		t.Fatalf("expected 0 results after suppression, got %d", len(searchAfterSuppress.Results))
	}
}

func TestServiceCorrectUpdatesLaterSearchResults(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)

	result, err := service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "I prefer concise answers."},
		},
	})
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}

	_, err = service.Correct(context.Background(), "t1", "u1", result.Memories[0].MemoryID, CorrectionRequest{
		Content: "Prefers detailed answers",
	})
	if err != nil {
		t.Fatalf("correct failed: %v", err)
	}

	search, err := service.Search(context.Background(), "t1", "u1", "", "", "How should I answer?")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(search.Results) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(search.Results))
	}
	if search.Results[0].Content != "Prefers detailed answers" {
		t.Fatalf("expected corrected content, got %q", search.Results[0].Content)
	}
}

func TestServiceIngestAsyncEnqueuesJob(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)

	result, err := service.IngestAsync(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "I prefer concise answers."},
		},
	})
	if err != nil {
		t.Fatalf("async ingest failed: %v", err)
	}
	if !result.Accepted || result.JobID == "" || result.IngestID == "" {
		t.Fatalf("expected accepted async ingest with ids, got %+v", result)
	}
	if len(store.jobs) != 1 {
		t.Fatalf("expected 1 enqueued job, got %d", len(store.jobs))
	}
}

func TestVerticalPackPrincipleRanksAbovePreference(t *testing.T) {
	reg, err := pack.LoadRegistryFromDir(filepath.Join("..", "..", "packs"))
	if err != nil {
		t.Fatalf("load packs: %v", err)
	}
	store := newMemoryStoreStub()
	service := NewServiceWithPacks(store, reg)

	_, err = service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "brand",
		Vertical:   "marketing",
		SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "I prefer warm casual copy."},
		},
	})
	if err != nil {
		t.Fatalf("preference ingest: %v", err)
	}

	_, err = service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "brand",
		Vertical:   "marketing",
		SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Never mention competitor X in any copy."},
		},
	})
	if err != nil {
		t.Fatalf("principle ingest: %v", err)
	}

	search, err := service.Search(context.Background(), "t1", "brand", "marketing", "", "copy")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(search.Results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(search.Results))
	}
	if !strings.Contains(strings.ToLower(search.Results[0].Content), "never") {
		t.Fatalf("expected principle first, got %q", search.Results[0].Content)
	}
}

func TestLifecycleArchivedCampaignExcludedFromSearch(t *testing.T) {
	reg, err := pack.LoadRegistryFromDir(filepath.Join("..", "..", "packs"))
	if err != nil {
		t.Fatalf("load packs: %v", err)
	}
	store := newMemoryStoreStub()
	service := NewServiceWithPacks(store, reg)

	_, err = service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "brand",
		Vertical:   "marketing",
		Label:      "campaign",
		Metadata:   map[string]any{"name": "Summer Sale", "status": "archived"},
		SourceType: "campaign",
		Messages: []Message{
			{Role: "user", Content: "Summer Sale campaign headline is Save 20% today."},
		},
	})
	if err != nil {
		t.Fatalf("ingest archived campaign: %v", err)
	}

	search, err := service.Search(context.Background(), "t1", "brand", "marketing", "", "Summer Sale headline")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(search.Results) != 0 {
		t.Fatalf("expected archived campaign hidden, got %d results", len(search.Results))
	}
}

func TestLifecycleActiveCampaignRanksAboveCompleted(t *testing.T) {
	reg, err := pack.LoadRegistryFromDir(filepath.Join("..", "..", "packs"))
	if err != nil {
		t.Fatalf("load packs: %v", err)
	}
	store := newMemoryStoreStub()
	service := NewServiceWithPacks(store, reg)

	_, err = service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "brand",
		Vertical:   "marketing",
		Label:      "campaign",
		Metadata:   map[string]any{"name": "Winter", "status": "completed"},
		SourceType: "campaign",
		Messages: []Message{
			{Role: "user", Content: "Winter splash sale is active."},
		},
	})
	if err != nil {
		t.Fatalf("ingest completed campaign: %v", err)
	}

	_, err = service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "brand",
		Vertical:   "marketing",
		Label:      "campaign",
		Metadata:   map[string]any{"name": "Summer", "status": "active"},
		SourceType: "campaign",
		Messages: []Message{
			{Role: "user", Content: "Summer splash sale is active."},
		},
	})
	if err != nil {
		t.Fatalf("ingest active campaign: %v", err)
	}

	search, err := service.Search(context.Background(), "t1", "brand", "marketing", "", "splash sale")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(search.Results) < 2 {
		t.Fatalf("expected 2 results, got %d", len(search.Results))
	}
	if !strings.Contains(search.Results[0].Content, "Summer") {
		t.Fatalf("expected active Summer campaign first, got %q", search.Results[0].Content)
	}
	if mult, ok := search.Results[0].Explain["lifecycle_rank_multiplier"].(float64); !ok || mult != 1.5 {
		t.Fatalf("expected active lifecycle multiplier 1.5, got %v", search.Results[0].Explain["lifecycle_rank_multiplier"])
	}
}

func TestLifecycleMetadataUpdateChangesState(t *testing.T) {
	reg, err := pack.LoadRegistryFromDir(filepath.Join("..", "..", "packs"))
	if err != nil {
		t.Fatalf("load packs: %v", err)
	}
	store := newMemoryStoreStub()
	service := NewServiceWithPacks(store, reg)

	req := IngestRequest{
		TenantID:   "t1",
		SubjectID:  "brand",
		Vertical:   "marketing",
		Label:      "campaign",
		Metadata:   map[string]any{"name": "Launch", "status": "active"},
		SourceType: "campaign",
		Messages: []Message{
			{Role: "user", Content: "Launch campaign offer is free shipping."},
		},
	}
	_, err = service.Ingest(context.Background(), req)
	if err != nil {
		t.Fatalf("ingest active campaign: %v", err)
	}

	searchBefore, err := service.Search(context.Background(), "t1", "brand", "marketing", "", "Launch offer")
	if err != nil {
		t.Fatalf("search before archive: %v", err)
	}
	if len(searchBefore.Results) != 1 {
		t.Fatalf("expected active campaign searchable, got %d", len(searchBefore.Results))
	}

	req.Metadata = map[string]any{"name": "Launch", "status": "archived"}
	result, err := service.Ingest(context.Background(), req)
	if err != nil {
		t.Fatalf("re-ingest archived campaign: %v", err)
	}
	if result.Updated != 1 {
		t.Fatalf("expected metadata update, got created=%d updated=%d deduped=%d", result.Created, result.Updated, result.Deduped)
	}

	searchAfter, err := service.Search(context.Background(), "t1", "brand", "marketing", "", "Launch offer")
	if err != nil {
		t.Fatalf("search after archive: %v", err)
	}
	if len(searchAfter.Results) != 0 {
		t.Fatalf("expected archived campaign hidden after update, got %d", len(searchAfter.Results))
	}
}

func TestSuppressedMemoryNotResurrectedOnReingest(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)

	ingest, err := service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []Message{{Role: "user", Content: "Never share the door code with vendors."}},
	})
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if err := service.Suppress(context.Background(), "t1", "u1", ingest.Memories[0].MemoryID); err != nil {
		t.Fatalf("suppress failed: %v", err)
	}

	reingest, err := service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []Message{{Role: "user", Content: "Never share the door code with vendors."}},
	})
	if err != nil {
		t.Fatalf("re-ingest failed: %v", err)
	}
	if reingest.Created != 0 {
		t.Fatalf("expected no new memories after re-ingesting suppressed content, got created=%d", reingest.Created)
	}

	search, err := service.Search(context.Background(), "t1", "u1", "", "", "door code")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(search.Results) != 0 {
		t.Fatalf("expected suppressed memory to stay hidden after re-ingest, got %d results", len(search.Results))
	}
}

func TestSearchPrefersNewerConflictingPreference(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)

	if _, err := service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []Message{{Role: "user", Content: "I prefer email updates."}},
	}); err != nil {
		t.Fatalf("first ingest failed: %v", err)
	}
	if _, err := service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []Message{{Role: "user", Content: "I prefer SMS updates."}},
	}); err != nil {
		t.Fatalf("second ingest failed: %v", err)
	}

	search, err := service.Search(context.Background(), "t1", "u1", "", "", "updates")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(search.Results) == 0 {
		t.Fatal("expected at least one search result")
	}
	if !strings.Contains(strings.ToLower(search.Results[0].Content), "sms") {
		t.Fatalf("expected newer SMS preference to rank first, got %q", search.Results[0].Content)
	}
}

func TestIngestRetainsDialogueAndRanksDatedFact(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)

	_, err := service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Metadata: map[string]any{
			"session_id":  "sess-1",
			"observed_at": "2023-05-07T18:00:00Z",
		},
		Messages: []Message{
			{Role: "user", Content: "Alex: I went to the community support group on 7 May 2023"},
			{Role: "user", Content: "Sam: Can't wait to see your show - the community needs more platforms"},
		},
	})
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}

	var hasEpisode bool
	var hasDate bool
	for _, record := range store.records {
		if record.Primitive == PrimitiveEpisode {
			hasEpisode = true
		}
		if strings.Contains(strings.ToLower(record.Content), "7 may 2023") {
			hasDate = true
		}
		if record.Metadata["session_id"] != "sess-1" {
			t.Fatalf("expected session_id metadata copied, got %#v", record.Metadata)
		}
	}
	if !hasDate {
		t.Fatal("expected dated support-group turn retained as a memory")
	}
	if !hasEpisode {
		t.Fatal("expected conversation_episode primitive on free dialogue")
	}
	for _, record := range store.records {
		if strings.Contains(strings.ToLower(record.Content), "7 may 2023") {
			if record.ObservedAt == nil {
				t.Fatal("expected ObservedAt from metadata.observed_at")
			}
		}
	}

	search, err := service.Search(context.Background(), "t1", "u1", "", "", "When did Alex go to the community support group")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(search.Results) == 0 {
		t.Fatal("expected search hits")
	}
	top := strings.ToLower(search.Results[0].Content)
	if !strings.Contains(top, "7 may 2023") && !strings.Contains(top, "support group") {
		t.Fatalf("expected dated fact to outrank topical preference neighbor, top=%q", search.Results[0].Content)
	}
	if search.Results[0].Explain["date_token_boost"] == nil && search.Results[0].Explain["exact_span_boost"] == nil && search.Results[0].Explain["episode_penalty"] == nil {
		t.Fatalf("expected ranking explain boosts, got %#v", search.Results[0].Explain)
	}
}

func TestExactSpanOutranksTopicalNeighbor(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	now := service.now()

	dated := MemoryRecord{
		MemoryID:  "mem_dated",
		TenantID:  "t1",
		SubjectID: "u1",
		Kind:      KindFact,
		Content:   "Alex went to the community support group on 7 May 2023",
		DedupeKey: "dated",
		Status:    StatusActive,
		UpdatedAt: now,
	}
	topical := MemoryRecord{
		MemoryID:  "mem_topical",
		TenantID:  "t1",
		SubjectID: "u1",
		Kind:      KindPreference,
		Content:   "Prefers LGBTQ community platforms and shows",
		DedupeKey: "topical",
		Status:    StatusActive,
		UpdatedAt: now,
	}
	store.records["dated"] = dated
	store.records["topical"] = topical

	search, err := service.Search(context.Background(), "t1", "u1", "", "", "When did Alex go to the community support group")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(search.Results) == 0 {
		t.Fatal("expected results")
	}
	if !strings.Contains(strings.ToLower(search.Results[0].Content), "7 may 2023") {
		t.Fatalf("expected dated fact first, got %q", search.Results[0].Content)
	}
}

func TestEpisodePenaltyPrefersTypedFact(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	now := service.now()
	store.records["ep"] = MemoryRecord{
		MemoryID: "mem_ep", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "Congratulations on sticking to your daily tidying routine for 3 weeks",
		DedupeKey: "ep", Status: StatusActive, UpdatedAt: now,
	}
	store.records["fact"] = MemoryRecord{
		MemoryID: "mem_f", TenantID: "t1", SubjectID: "u1",
		Kind:      KindFact,
		Content:   "Alex has been sticking to a daily tidying routine for 4 weeks",
		DedupeKey: "fact", Status: StatusActive, UpdatedAt: now,
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "how long have I been sticking to my daily tidying routine")
	if err != nil {
		t.Fatal(err)
	}
	if len(search.Results) == 0 {
		t.Fatal("expected hits")
	}
	if !strings.Contains(search.Results[0].Content, "4 weeks") {
		t.Fatalf("typed fact should outrank congratulation episode, top=%q", search.Results[0].Content)
	}
}

func TestDefaultSearchDropsEpisodesWhenFactsExist(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	now := service.now()
	store.records["ep"] = MemoryRecord{
		MemoryID: "mem_ep", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "Yeah, Caroline, Yep, Melanie, Hey Caroline",
		DedupeKey: "ep", Status: StatusActive, UpdatedAt: now,
	}
	store.records["fact"] = MemoryRecord{
		MemoryID: "mem_f", TenantID: "t1", SubjectID: "u1",
		Kind:      KindFact,
		Content:   "Caroline is from Sweden",
		DedupeKey: "fact", Status: StatusActive, UpdatedAt: now,
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "Where did Caroline move from")
	if err != nil {
		t.Fatal(err)
	}
	if search.Trace == nil || search.Trace.EpisodesDropped < 1 {
		t.Fatalf("expected episodes dropped, trace=%+v", search.Trace)
	}
	for _, r := range search.Results {
		if strings.Contains(strings.ToLower(r.Content), "yeah") {
			t.Fatalf("provenance episode leaked into default search: %q", r.Content)
		}
	}
	if len(search.Results) == 0 || !strings.Contains(strings.ToLower(search.Results[0].Content), "sweden") {
		t.Fatalf("expected fact hit, got %#v", search.Results)
	}

	withEp, err := service.SearchOpt(context.Background(), "t1", "u1", "", "", "Where did Caroline move from", SearchOptions{IncludeEpisodes: true, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	sawEp := false
	for _, r := range withEp.Results {
		if strings.Contains(strings.ToLower(r.Content), "yeah") {
			sawEp = true
		}
	}
	if !sawEp {
		t.Fatal("IncludeEpisodes should keep provenance turns")
	}
}

func TestEpisodeOnlyPoolFallsBack(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	now := service.now()
	store.records["ep"] = MemoryRecord{
		MemoryID: "mem_ep", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "Alex went to the community support group on 7 May 2023",
		DedupeKey: "ep", Status: StatusActive, UpdatedAt: now,
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "When did Alex go to the community support group")
	if err != nil {
		t.Fatal(err)
	}
	if search.Trace == nil || !search.Trace.EpisodeFallback {
		t.Fatalf("expected episode fallback when no facts exist, trace=%+v", search.Trace)
	}
	if len(search.Results) == 0 {
		t.Fatal("fallback must not empty the pool")
	}
}

func TestCandidateLimitRecordedOnTrace(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	_, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t-pool", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{{Role: "user", Content: "Alex lives in Austin"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := service.SearchOpt(context.Background(), "t-pool", "u1", "", "", "where does Alex live", SearchOptions{Limit: 10, CandidateLimit: 50})
	if err != nil {
		t.Fatal(err)
	}
	if out.Trace == nil || out.Trace.CandidateOverfetch != 50 {
		t.Fatalf("expected candidate pool 50, trace=%+v", out.Trace)
	}
}

func TestSessionNeighborExpansion(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	_, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Metadata: map[string]any{"session_id": "sess-a", "observed_at": "2023-05-08T12:00:00Z"},
		Messages: []Message{
			{Role: "user", Content: "Alex: I am a community organizer"},
			{Role: "user", Content: "Sam: That takes courage"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "What is Alex identity")
	if err != nil {
		t.Fatal(err)
	}
	if len(search.Results) == 0 {
		t.Fatal("expected results")
	}
	joined := ""
	for _, r := range search.Results {
		joined += " " + r.Content
	}
	if !strings.Contains(strings.ToLower(joined), "community organizer") {
		t.Fatalf("expected session content in results, got %q", joined)
	}
}

func TestQuestionMemoriesDownrankedForFactQueries(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	_, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Metadata: map[string]any{"session_id": "s1", "observed_at": "2023-05-08T12:00:00Z"},
		Messages: []Message{
			{Role: "user", Content: "What did Alex research last week?"},
			{Role: "user", Content: "Alex: I researched adoption agencies this week"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "What did Alex research?")
	if err != nil {
		t.Fatal(err)
	}
	if len(search.Results) == 0 {
		t.Fatal("expected results")
	}
	top := search.Results[0].Content
	if strings.Contains(top, "?") && !strings.Contains(strings.ToLower(top), "adoption") {
		t.Fatalf("expected factual adoption memory on top, got %q", top)
	}
	if !strings.Contains(strings.ToLower(top), "adoption") {
		t.Fatalf("expected adoption fact ranked first, got %q explain=%v", top, search.Results[0].Explain)
	}
}

func TestLowInfoNameOnlyDownranked(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	_, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Metadata: map[string]any{"session_id": "s1", "observed_at": "2023-05-08T12:00:00Z"},
		Messages: []Message{
			{Role: "user", Content: "Yeah, Alice"},
			{Role: "user", Content: "Thanks, Alice"},
			{Role: "user", Content: "Alice: I am training for a marathon and run every morning before work"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "What activities does Alice enjoy?")
	if err != nil {
		t.Fatal(err)
	}
	if len(search.Results) == 0 {
		t.Fatal("expected results")
	}
	top := strings.ToLower(search.Results[0].Content)
	if !strings.Contains(top, "marathon") && !strings.Contains(top, "run") {
		t.Fatalf("expected content-dense activity memory first, got %q explain=%v", search.Results[0].Content, search.Results[0].Explain)
	}
}

func TestSubjectContentExpansionSurfacesProfile(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	_, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Metadata: map[string]any{"session_id": "s-profile", "observed_at": "2023-05-08T12:00:00Z"},
		Messages: []Message{
			{Role: "user", Content: "Hey Bob"},
			{Role: "user", Content: "Bob: pottery keeps me grounded after long weeks at the studio"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Query verbs do not appear in the pottery memory; subject bridge must admit it.
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "What hobbies does Bob enjoy?")
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, r := range search.Results {
		joined += " " + strings.ToLower(r.Content)
	}
	if !strings.Contains(joined, "pottery") {
		t.Fatalf("expected subject-content expansion to surface pottery, got %q", joined)
	}
}

func TestSupersedeHidesPriorFromDefaultSearch(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	ingested, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "note",
		Messages: []Message{{Role: "user", Content: "Door code is 1111"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ingested.Memories) == 0 {
		t.Fatal("expected memory")
	}
	priorID := ingested.Memories[0].MemoryID
	replaced, err := service.Supersede(context.Background(), "t1", "u1", priorID, SupersedeRequest{
		Content: "Door code is 2222",
	})
	if err != nil {
		t.Fatal(err)
	}
	if replaced.MemoryID == priorID {
		t.Fatal("expected a new memory id for superseding record")
	}

	search, err := service.Search(context.Background(), "t1", "u1", "", "", "door code")
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range search.Results {
		if strings.Contains(r.Content, "1111") {
			t.Fatalf("superseded prior leaked into default search: %q", r.Content)
		}
		if r.MemoryID == priorID {
			t.Fatalf("superseded id %s visible in default search", priorID)
		}
	}
	joined := ""
	for _, r := range search.Results {
		joined += " " + r.Content
	}
	if !strings.Contains(joined, "2222") {
		t.Fatalf("expected replacement in search, got %q", joined)
	}

	hist, err := service.SearchOpt(context.Background(), "t1", "u1", "", "", "door code", SearchOptions{IncludeHistorical: true})
	if err != nil {
		t.Fatal(err)
	}
	foundPrior := false
	for _, r := range hist.Results {
		if r.MemoryID == priorID || strings.Contains(r.Content, "1111") {
			foundPrior = true
		}
	}
	if !foundPrior {
		t.Fatal("expected include_historical to surface superseded prior")
	}
}

func TestHistoricalIntentRetrievesPriorResidence(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	now := service.now()
	store.records["ny"] = MemoryRecord{
		MemoryID: "mem_ny", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Content: "Alex lives in New York",
		DedupeKey: "ny", Status: StatusActive, LifecycleState: LifecycleSuperseded,
		Metadata:  map[string]any{"predicate": PredicateResidence, "value_norm": "new york", "memory_type": "state"},
		CreatedAt: now, UpdatedAt: now,
	}
	store.records["au"] = MemoryRecord{
		MemoryID: "mem_au", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Content: "Alex lives in Austin",
		DedupeKey: "au", Status: StatusActive, LifecycleState: LifecycleActive,
		Metadata:  map[string]any{"predicate": PredicateResidence, "value_norm": "austin", "memory_type": "state"},
		CreatedAt: now, UpdatedAt: now,
	}

	cur, err := service.SearchOpt(context.Background(), "t1", "u1", "", "", "where does Alex currently live", SearchOptions{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, r := range cur.Results {
		joined += " " + r.Content
		if strings.Contains(r.Content, "New York") {
			t.Fatalf("current-state search leaked superseded NY: %q", r.Content)
		}
	}
	if !strings.Contains(joined, "Austin") {
		t.Fatalf("current-state should prefer Austin, got %q", joined)
	}

	hist, err := service.SearchOpt(context.Background(), "t1", "u1", "", "", "where did Alex live before", SearchOptions{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	foundNY := false
	for _, r := range hist.Results {
		if strings.Contains(r.Content, "New York") {
			foundNY = true
		}
	}
	if !foundNY {
		t.Fatalf("historical intent should retrieve NY, results=%+v", hist.Results)
	}
}

func TestListQueryDiversifiesThemes(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	_, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Metadata: map[string]any{"session_id": "s-div", "observed_at": "2023-05-08T12:00:00Z"},
		Messages: []Message{
			{Role: "user", Content: "Yeah, Dana"},
			{Role: "user", Content: "Dana: I love pottery and spend weekends at the studio shaping bowls"},
			{Role: "user", Content: "Dana: Camping in the mountains clears my head after long weeks"},
			{Role: "user", Content: "Dana: Swimming at the community pool is my tuesday habit"},
			{Role: "user", Content: "Dana: Wow that sounds great"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "What activities does Dana enjoy?")
	if err != nil {
		t.Fatal(err)
	}
	if len(search.Results) < 3 {
		t.Fatalf("expected diversified candidates, got %d", len(search.Results))
	}
	head := ""
	for i, r := range search.Results {
		if i >= 6 {
			break
		}
		head += " " + strings.ToLower(r.Content)
	}
	for _, need := range []string{"pottery", "camping", "swimming"} {
		if !strings.Contains(head, need) {
			t.Fatalf("expected %q in diversified head results, got %q", need, head)
		}
	}
}

func TestDomainEventMatchByLabel(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	ingested, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "note",
		Label:    "promo",
		Metadata: map[string]any{"season": "summer"},
		// Fact-shaped so deterministic extract still fires when Label is set
		// (labeled ingest skips conversation_episode retention).
		Messages: []Message{{Role: "user", Content: "Summer splash promo is the active seasonal campaign."}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ingested.Memories) == 0 {
		t.Fatal("expected memory")
	}
	// Ensure label stuck on the stored record (core pack may ignore unknown labels).
	got, err := store.GetMemory(context.Background(), "t1", "u1", ingested.Memories[0].MemoryID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Label == "" {
		// Force label for match path coverage when pack vocabulary omits it.
		got.Label = "promo"
		got.Metadata = map[string]any{"season": "summer"}
		store.records[got.DedupeKey] = got
	}
	res, err := service.ApplyDomainEvent(context.Background(), DomainEventRequest{
		TenantID: "t1", SubjectID: "u1", EventType: "promo_ended",
		Match: &DomainEventMatch{Label: "promo"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Superseded) == 0 {
		t.Fatalf("expected match-based supersede, record=%+v", got)
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "Summer splash promo")
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range search.Results {
		if r.MemoryID == ingested.Memories[0].MemoryID {
			t.Fatalf("matched memory still searchable: %q", r.Content)
		}
	}
}

func TestDomainEventBatchSupersede(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	ingested, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "note",
		Messages: []Message{
			{Role: "user", Content: "Campaign splash is live"},
			{Role: "user", Content: "Campaign splash headline is Ready"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 0, len(ingested.Memories))
	for _, m := range ingested.Memories {
		ids = append(ids, m.MemoryID)
	}
	res, err := service.ApplyDomainEvent(context.Background(), DomainEventRequest{
		TenantID: "t1", SubjectID: "u1", EventType: "campaign_ended",
		SupersedeMemoryIDs: ids,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Superseded) == 0 {
		t.Fatal("expected superseded ids")
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "Campaign splash")
	if err != nil {
		t.Fatal(err)
	}
	if len(search.Results) != 0 {
		t.Fatalf("expected empty search after batch supersede, got %#v", search.Results)
	}
}
