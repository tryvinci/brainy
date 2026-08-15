package memory

import (
	"context"
	"errors"
	"testing"
	"time"
)

// evidenceStoreStub is a memoryStoreStub that also implements RawEvidenceWriter
// and can be told to fail raw-evidence writes per source ref (msg:%d).
type evidenceStoreStub struct {
	*memoryStoreStub
	evidence map[string]string
	failRefs map[string]error
	writes   int
}

func newEvidenceStoreStub() *evidenceStoreStub {
	return &evidenceStoreStub{
		memoryStoreStub: newMemoryStoreStub(),
		evidence:        map[string]string{},
		failRefs:        map[string]error{},
	}
}

func (s *evidenceStoreStub) WriteRawEvidence(_ context.Context, _, _, _, _, ref, _, _, _ string, _ *time.Time, _ map[string]any) (string, error) {
	s.writes++
	if err, ok := s.failRefs[ref]; ok {
		return "", err
	}
	id := "evid_" + ref
	s.evidence[ref] = id
	return id, nil
}

func (s *evidenceStoreStub) ListEvidence(_ context.Context, _, _ string, _ int) ([]map[string]any, error) {
	return nil, nil
}

func strictEvidenceIngestRequest() IngestRequest {
	return IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "I prefer concise, direct answers."},
			{Role: "user", Content: "I prefer detailed, thorough answers."},
		},
	}
}

// TestEvidenceStrictSyncIngestFailsClosedOnPartialEvidenceFailure guards the P1
// invariant that BRAINY_EVIDENCE_STRICT=true fails closed: when any source
// message's raw evidence cannot be recorded, sync ingest aborts before any
// semantic persistence (no memory is created), even if other messages succeed.
func TestEvidenceStrictSyncIngestFailsClosedOnPartialEvidenceFailure(t *testing.T) {
	t.Setenv("BRAINY_EVIDENCE_STRICT", "true")
	store := newEvidenceStoreStub()
	store.failRefs["msg:1"] = errors.New("evidence backend unavailable")
	service := NewService(store)

	_, err := service.Ingest(context.Background(), strictEvidenceIngestRequest())
	if err == nil {
		t.Fatal("expected strict ingest to fail on evidence write failure")
	}
	if len(store.records) != 0 {
		t.Fatalf("strict ingest aborted but %d memory records were persisted", len(store.records))
	}
	if len(store.evidence) != 1 {
		t.Fatalf("expected only the succeeding message's evidence to be captured, got %d", len(store.evidence))
	}
}

// TestEvidenceStrictAsyncIngestDoesNotEnqueueOnEvidenceFailure guards the P1
// invariant that strict async ingest does not enqueue a job after evidence
// failure: the failure surfaces as an error and no job is created.
func TestEvidenceStrictAsyncIngestDoesNotEnqueueOnEvidenceFailure(t *testing.T) {
	t.Setenv("BRAINY_EVIDENCE_STRICT", "true")
	store := newEvidenceStoreStub()
	store.failRefs["msg:0"] = errors.New("evidence backend unavailable")
	service := NewService(store)

	_, err := service.IngestAsync(context.Background(), strictEvidenceIngestRequest())
	if err == nil {
		t.Fatal("expected strict async ingest to fail on evidence write failure")
	}
	if len(store.jobs) != 0 {
		t.Fatalf("strict async ingest aborted but %d jobs were enqueued", len(store.jobs))
	}
}

// TestEvidenceNonStrictBestEffortContinuesOnEvidenceFailure guards that the
// non-strict path keeps its existing best-effort behavior: an evidence write
// failure is recorded and skipped, ingest proceeds, and memory is created.
func TestEvidenceNonStrictBestEffortContinuesOnEvidenceFailure(t *testing.T) {
	t.Setenv("BRAINY_EVIDENCE_STRICT", "false")
	store := newEvidenceStoreStub()
	store.failRefs["msg:1"] = errors.New("evidence backend unavailable")
	service := NewService(store)

	result, err := service.Ingest(context.Background(), strictEvidenceIngestRequest())
	if err != nil {
		t.Fatalf("non-strict ingest should proceed on evidence failure, got error: %v", err)
	}
	if !result.Accepted {
		t.Fatal("expected non-strict ingest to be accepted")
	}
	if len(store.records) == 0 {
		t.Fatal("expected non-strict ingest to persist memories despite evidence failure")
	}
	if len(store.evidence) != 1 {
		t.Fatalf("expected only the succeeding message's evidence to be captured, got %d", len(store.evidence))
	}
}
