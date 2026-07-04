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
