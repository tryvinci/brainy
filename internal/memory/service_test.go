package memory

import (
	"context"
	"errors"
	"testing"
	"time"
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
		if existing.Content == record.Content && existing.Status == StatusActive {
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

func (s *memoryStoreStub) ListActiveMemories(_ context.Context, tenantID, subjectID string) ([]MemoryRecord, error) {
	var out []MemoryRecord
	for _, record := range s.records {
		if record.TenantID == tenantID && record.SubjectID == subjectID && record.Status == StatusActive {
			out = append(out, record)
		}
	}
	return out, nil
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

func (s *memoryStoreStub) EnqueueIngestJob(_ context.Context, ingestID, jobID string, req IngestRequest) error {
	s.jobs[jobID] = ExtractionJob{JobID: jobID, IngestID: ingestID, Request: req}
	return nil
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

	search, err := service.Search(context.Background(), "t1", "u1", "How should I respond?")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(search.Results) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(search.Results))
	}

	if err := service.Suppress(context.Background(), "t1", "u1", result.Memories[0].MemoryID); err != nil {
		t.Fatalf("suppress failed: %v", err)
	}

	searchAfterSuppress, err := service.Search(context.Background(), "t1", "u1", "How should I respond?")
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

	search, err := service.Search(context.Background(), "t1", "u1", "How should I answer?")
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
