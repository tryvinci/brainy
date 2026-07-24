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
	records map[string]MemoryRecord
	jobs    map[string]ExtractionJob
}

func newMemoryStoreStub() *memoryStoreStub {
	return &memoryStoreStub{
		records: map[string]MemoryRecord{},
		jobs:    map[string]ExtractionJob{},
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

func (s *memoryStoreStub) ListActiveMemories(_ context.Context, tenantID, subjectID string) ([]MemoryRecord, error) {
	var out []MemoryRecord
	for _, record := range s.records {
		if record.TenantID == tenantID && record.SubjectID == subjectID && record.Status == StatusActive &&
			IsLifecycleSearchVisible(record.LifecycleState) {
			out = append(out, record)
		}
	}
	return out, nil
}

func (s *memoryStoreStub) SearchActiveMemories(_ context.Context, tenantID, subjectID string, patterns []string, limit int) ([]MemoryRecord, error) {
	_ = patterns
	return s.ListActiveMemories(context.Background(), tenantID, subjectID)
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
			{Role: "user", Content: "Caroline: I went to the LGBTQ support group on 7 May 2023"},
			{Role: "user", Content: "Melanie: Can't wait to see your show - the LGBTQ community needs more platforms like this"},
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

	search, err := service.Search(context.Background(), "t1", "u1", "", "", "When did Caroline go to the LGBTQ support group")
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
	if search.Results[0].Explain["date_token_boost"] == nil && search.Results[0].Explain["exact_span_boost"] == nil && search.Results[0].Explain["episode_boost"] == nil {
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
		Primitive: PrimitiveEpisode,
		Content:   "Caroline went to the LGBTQ support group on 7 May 2023",
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

	search, err := service.Search(context.Background(), "t1", "u1", "", "", "When did Caroline go to the LGBTQ support group")
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

func TestSessionNeighborExpansion(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)
	_, err := service.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Metadata: map[string]any{"session_id": "sess-a", "observed_at": "2023-05-08T12:00:00Z"},
		Messages: []Message{
			{Role: "user", Content: "Caroline: I am a transgender woman"},
			{Role: "user", Content: "Melanie: That takes courage"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "What is Caroline identity")
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
	if !strings.Contains(strings.ToLower(joined), "transgender") {
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
			{Role: "user", Content: "What did Caroline research last week?"},
			{Role: "user", Content: "Caroline: I researched adoption agencies this week"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "What did Caroline research?")
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
	search, err := service.Search(context.Background(), "t1", "u1", "", "", "What hobbies does Bob partake in?")
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

