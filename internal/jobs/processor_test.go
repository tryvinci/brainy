package jobs

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"brainy/internal/memory"
	"brainy/internal/observability"
)

type storeStub struct {
	records    map[string]memory.MemoryRecord
	jobs       map[string]memory.ExtractionJob
	failedJobs map[string]string
	embeddings map[string][]float32
	relations  []memory.MemoryRelation
}

func newStoreStub() *storeStub {
	return &storeStub{
		records:    map[string]memory.MemoryRecord{},
		jobs:       map[string]memory.ExtractionJob{},
		failedJobs: map[string]string{},
		embeddings: map[string][]float32{},
	}
}

func (s *storeStub) UpsertEmbedding(_ context.Context, memoryID, _, _ string, values []float32) error {
	if len(values) == 0 {
		return nil
	}
	copied := make([]float32, len(values))
	copy(copied, values)
	s.embeddings[memoryID] = copied
	return nil
}

func (s *storeStub) UpsertMemory(_ context.Context, record memory.MemoryRecord) (memory.StoreUpsertResult, error) {
	s.records[record.DedupeKey] = record
	return memory.StoreUpsertResult{Record: record, State: "created"}, nil
}

func (s *storeStub) ListActiveMemories(ctx context.Context, tenantID, subjectID string) ([]memory.MemoryRecord, error) {
	return s.ListMemories(ctx, tenantID, subjectID, false)
}

func (s *storeStub) ListMemories(_ context.Context, tenantID, subjectID string, includeSuperseded bool) ([]memory.MemoryRecord, error) {
	_ = includeSuperseded
	var out []memory.MemoryRecord
	for _, record := range s.records {
		if record.TenantID == tenantID && record.SubjectID == subjectID && record.Status == memory.StatusActive {
			out = append(out, record)
		}
	}
	return out, nil
}

func (s *storeStub) SearchActiveMemories(ctx context.Context, tenantID, subjectID string, patterns []string, limit int) ([]memory.MemoryRecord, error) {
	return s.SearchMemories(ctx, tenantID, subjectID, patterns, limit, false)
}

func (s *storeStub) SearchMemories(ctx context.Context, tenantID, subjectID string, patterns []string, limit int, includeSuperseded bool) ([]memory.MemoryRecord, error) {
	_ = patterns
	_ = limit
	return s.ListMemories(ctx, tenantID, subjectID, includeSuperseded)
}

func (s *storeStub) GetMemory(_ context.Context, _, _, _ string) (memory.MemoryRecord, error) {
	return memory.MemoryRecord{}, memory.ErrMemoryNotFound
}

func (s *storeStub) MarkSuperseded(_ context.Context, _, _, _ string) error { return nil }

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

func (s *storeStub) FailExtractionJob(_ context.Context, jobID, _, reason string) error {
	s.failedJobs[jobID] = reason
	return nil
}

func (s *storeStub) UpsertMemoryRelation(_ context.Context, rel memory.MemoryRelation) error {
	s.relations = append(s.relations, rel)
	return nil
}

func (s *storeStub) ListRelationsFrom(_ context.Context, tenantID, subjectID, srcEntity, relation string, limit int) ([]memory.MemoryRelation, error) {
	srcEntity = strings.ToLower(strings.TrimSpace(srcEntity))
	out := make([]memory.MemoryRelation, 0)
	for _, rel := range s.relations {
		if rel.TenantID != tenantID || rel.SubjectID != subjectID {
			continue
		}
		if srcEntity != "" &&
			strings.ToLower(rel.SrcEntity) != srcEntity &&
			rel.SrcEntityID != srcEntity &&
			rel.SrcEntityID != memory.CanonicalEntityID(tenantID, subjectID, srcEntity) {
			continue
		}
		if relation != "" && rel.Relation != relation {
			continue
		}
		out = append(out, rel)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

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
	if len(store.embeddings) != 1 {
		t.Fatalf("expected 1 embedding upsert, got %d", len(store.embeddings))
	}
	for memoryID, values := range store.embeddings {
		if len(values) == 0 {
			t.Fatalf("expected non-empty embedding for %s", memoryID)
		}
	}
}

func TestProcessorProviderFailureLeavesNoMemories(t *testing.T) {
	store := newStoreStub()
	failing := failingExtractor{err: context.DeadlineExceeded}
	processor := NewProcessorWithExtractor(store, observability.NewMetrics(), failing)

	_, err := store.EnqueueIngestJob(context.Background(), "ing_fail", "job_fail", "", memory.IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []memory.Message{{Role: "user", Content: "Caroline went on 7 May 2023"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	processed, err := processor.ProcessNext(context.Background())
	if err == nil {
		t.Fatal("expected extract failure")
	}
	if !processed {
		t.Fatal("expected job claim")
	}
	if len(store.records) != 0 {
		t.Fatalf("provider failure must not upsert memories, got %d", len(store.records))
	}
	if store.failedJobs["job_fail"] == "" {
		t.Fatal("expected FailExtractionJob called")
	}
}

func TestProcessorProjectsRelationsFromTypedFacts(t *testing.T) {
	store := newStoreStub()
	processor := NewProcessor(store, observability.NewMetrics())
	_, err := store.EnqueueIngestJob(context.Background(), "ing_rel", "job_rel", "", memory.IngestRequest{
		TenantID:   "t-rel",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages: []memory.Message{
			{Role: "user", Content: "Jordan: I moved from Portugal four years ago."},
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
		t.Fatal("expected job")
	}
	found := false
	for _, rel := range store.relations {
		if rel.Relation == memory.PredicateOrigin && strings.Contains(strings.ToLower(rel.DstEntity), "portugal") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected origin relation from async extract, relations=%#v records=%d", store.relations, len(store.records))
	}
}

type failingExtractor struct {
	err error
}

func (f failingExtractor) Extract(_ context.Context, _ memory.IngestRequest) ([]memory.ExtractedMemory, error) {
	return nil, f.err
}

// fencedStoreStub implements memory.LeaseFencer to prove the processor uses the
// fenced completion path and propagates ErrLeaseLost.
type fencedStoreStub struct {
	storeStub
	mu             sync.Mutex
	heartbeats     int
	completeFenced bool
	failFenced     bool
	failErr        error
}

func (s *fencedStoreStub) HeartbeatExtractionJob(_ context.Context, _, _ string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.heartbeats++
	return nil
}

func (s *fencedStoreStub) heartbeatCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.heartbeats
}

func (s *fencedStoreStub) CompleteExtractionJobFenced(_ context.Context, _, _, _ string) error {
	s.completeFenced = true
	return nil
}

func (s *fencedStoreStub) FailExtractionJobFenced(_ context.Context, _, _, _, _ string) error {
	s.failFenced = true
	return s.failErr
}

func TestProcessorUsesFencedLeasePath(t *testing.T) {
	store := &fencedStoreStub{}
	store.records = map[string]memory.MemoryRecord{}
	store.jobs = map[string]memory.ExtractionJob{}
	store.failedJobs = map[string]string{}
	store.embeddings = map[string][]float32{}

	processor := NewProcessor(store, observability.NewMetrics())
	_, err := store.EnqueueIngestJob(context.Background(), "ing_f", "job_f", "", memory.IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []memory.Message{{Role: "user", Content: "I prefer concise answers."}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := processor.ProcessNext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !store.completeFenced {
		t.Fatal("expected fenced completion to be used")
	}
	if store.heartbeatCount() == 0 {
		t.Fatal("expected lease heartbeat during processing")
	}
}

type delayedExtractor struct {
	delay time.Duration
}

func (d delayedExtractor) Extract(ctx context.Context, _ memory.IngestRequest) ([]memory.ExtractedMemory, error) {
	timer := time.NewTimer(d.delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		return nil, nil
	}
}

func TestProcessorRenewsLeaseDuringSlowExtract(t *testing.T) {
	orig := extractionLeaseRenewInterval
	extractionLeaseRenewInterval = 25 * time.Millisecond
	t.Cleanup(func() { extractionLeaseRenewInterval = orig })

	store := &fencedStoreStub{}
	store.records = map[string]memory.MemoryRecord{}
	store.jobs = map[string]memory.ExtractionJob{}
	store.failedJobs = map[string]string{}
	store.embeddings = map[string][]float32{}

	processor := NewProcessorWithExtractor(store, observability.NewMetrics(), delayedExtractor{delay: 140 * time.Millisecond})
	_, err := store.EnqueueIngestJob(context.Background(), "ing_slow", "job_slow", "", memory.IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []memory.Message{{Role: "user", Content: "I prefer concise answers."}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := processor.ProcessNext(context.Background()); err != nil {
		t.Fatal(err)
	}
	got := store.heartbeatCount()
	if got < 3 {
		t.Fatalf("expected initial heartbeat plus ticker renewals during slow extract, got %d", got)
	}
}

func TestProcessorPropagatesLeaseLostOnFailure(t *testing.T) {
	store := &fencedStoreStub{failErr: memory.ErrLeaseLost}
	store.records = map[string]memory.MemoryRecord{}
	store.jobs = map[string]memory.ExtractionJob{}
	store.failedJobs = map[string]string{}
	store.embeddings = map[string][]float32{}

	failing := failingExtractor{err: context.DeadlineExceeded}
	processor := NewProcessorWithExtractor(store, observability.NewMetrics(), failing)
	_, err := store.EnqueueIngestJob(context.Background(), "ing_fl", "job_fl", "", memory.IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []memory.Message{{Role: "user", Content: "Caroline went on 7 May 2023"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := processor.ProcessNext(context.Background()); err == nil {
		t.Fatal("expected extract failure")
	}
	if !store.failFenced {
		t.Fatal("expected fenced failure to be used")
	}
}
