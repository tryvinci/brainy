package jobs

import (
	"context"
	"testing"

	"brainy/internal/memory"
	"brainy/internal/observability"
)

type storeStub struct {
	records map[string]memory.MemoryRecord
	jobs    map[string]memory.ExtractionJob
}

func newStoreStub() *storeStub {
	return &storeStub{
		records: map[string]memory.MemoryRecord{},
		jobs:    map[string]memory.ExtractionJob{},
	}
}

func (s *storeStub) UpsertMemory(_ context.Context, record memory.MemoryRecord) (memory.StoreUpsertResult, error) {
	s.records[record.DedupeKey] = record
	return memory.StoreUpsertResult{Record: record, State: "created"}, nil
}

func (s *storeStub) ListActiveMemories(_ context.Context, tenantID, subjectID string) ([]memory.MemoryRecord, error) {
	var out []memory.MemoryRecord
	for _, record := range s.records {
		if record.TenantID == tenantID && record.SubjectID == subjectID && record.Status == memory.StatusActive {
			out = append(out, record)
		}
	}
	return out, nil
}

func (s *storeStub) SuppressMemory(_ context.Context, _, _, _ string) error { return nil }

func (s *storeStub) CorrectMemory(_ context.Context, _, _, _, _, _ string) (memory.MemoryRecord, error) {
	return memory.MemoryRecord{}, nil
}

func (s *storeStub) EnqueueIngestJob(_ context.Context, ingestID, jobID, _ string, req memory.IngestRequest) (memory.EnqueueResult, error) {
	s.jobs[jobID] = memory.ExtractionJob{JobID: jobID, IngestID: ingestID, Request: req}
	return memory.EnqueueResult{IngestID: ingestID, JobID: jobID, Accepted: true}, nil
}

func (s *storeStub) ClaimNextExtractionJob(_ context.Context) (memory.ExtractionJob, bool, error) {
	for jobID, job := range s.jobs {
		delete(s.jobs, jobID)
		return job, true, nil
	}
	return memory.ExtractionJob{}, false, nil
}

func (s *storeStub) CompleteExtractionJob(_ context.Context, _, _ string) error { return nil }

func (s *storeStub) FailExtractionJob(_ context.Context, _, _, _ string) error { return nil }

func TestProcessorProcessesQueuedJob(t *testing.T) {
	store := newStoreStub()
	processor := NewProcessor(store, observability.NewMetrics())

	_, err := store.EnqueueIngestJob(context.Background(), "ing_1", "job_1", "", memory.IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages: []memory.Message{
			{Role: "user", Content: "I prefer concise answers."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	processed, err := processor.ProcessNext(context.Background())
	if err != nil {
		t.Fatalf("process next failed: %v", err)
	}
	if !processed {
		t.Fatalf("expected a queued job to be processed")
	}
	if len(store.records) != 1 {
		t.Fatalf("expected 1 created memory, got %d", len(store.records))
	}
}
