package memory

import (
	"context"
	"testing"
	"time"
)

func TestShouldReplaceCurrentStateRejectsOlderLateArrival(t *testing.T) {
	newerObs := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	olderObs := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	existing := MemoryRecord{
		MemoryID:   "mem_new",
		ObservedAt: &newerObs,
		CreatedAt:  newerObs,
		Confidence: 0.9,
	}
	incoming := MemoryRecord{
		MemoryID:   "mem_old_late",
		ObservedAt: &olderObs,
		CreatedAt:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), // recorded later
		Confidence: 0.95,
	}
	if shouldReplaceCurrentState(incoming, existing) {
		t.Fatal("older world-valid fact must not replace newer current state")
	}
}

func TestShouldReplaceCurrentStateAllowsCorrection(t *testing.T) {
	obs := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	existing := MemoryRecord{MemoryID: "mem_a", ObservedAt: &obs, CreatedAt: obs, Confidence: 0.9}
	incoming := MemoryRecord{
		MemoryID:   "mem_b",
		ObservedAt: &obs,
		CreatedAt:  obs.Add(time.Hour),
		Confidence: 0.5,
		Explain:    map[string]any{"assertion_kind": "corrective"},
	}
	if !shouldReplaceCurrentState(incoming, existing) {
		t.Fatal("corrective assertion should replace")
	}
}

func TestShouldReplaceCurrentStateExplicitSupersede(t *testing.T) {
	obs := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	existing := MemoryRecord{MemoryID: "mem_a", ObservedAt: &obs, CreatedAt: obs}
	incoming := MemoryRecord{
		MemoryID:  "mem_b",
		CreatedAt: obs.Add(-time.Hour),
		Metadata:  map[string]any{"supersedes_memory_id": "mem_a"},
	}
	if !shouldReplaceCurrentState(incoming, existing) {
		t.Fatal("explicit supersede should replace")
	}
}

func TestParseAsOf(t *testing.T) {
	if _, ok := ParseAsOf("2024-03-15"); !ok {
		t.Fatal("date-only")
	}
	if _, ok := ParseAsOf("2024-03-15T12:00:00Z"); !ok {
		t.Fatal("rfc3339")
	}
	if _, ok := ParseAsOf("not-a-date"); ok {
		t.Fatal("expected fail")
	}
}

func TestProjectCurrentStateRejectsOlderLateArrival(t *testing.T) {
	newerObs := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	olderObs := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	store := &projectStateStub{
		records: map[string]MemoryRecord{
			"mem_new": {
				MemoryID: "mem_new", TenantID: "t", SubjectID: "s",
				Status: StatusActive, LifecycleState: LifecycleActive,
				ObservedAt: &newerObs, CreatedAt: newerObs, Confidence: 0.9,
				Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse"},
			},
		},
		current: map[string]string{"occupation": "mem_new"},
		values:  map[string]string{"occupation": "nurse"},
	}
	late := MemoryRecord{
		MemoryID: "mem_old_late", TenantID: "t", SubjectID: "s",
		Status: StatusActive, LifecycleState: LifecycleActive,
		ObservedAt: &olderObs, CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Confidence: 0.99,
		Metadata:   map[string]any{"predicate": PredicateOccupation, "value_norm": "doctor"},
	}
	ProjectCurrentStateIfApplicable(context.Background(), store, late)
	if store.current["occupation"] != "mem_new" {
		t.Fatalf("late older fact projected: %v", store.current)
	}
}

type projectStateStub struct {
	records map[string]MemoryRecord
	current map[string]string
	values  map[string]string
}

func (s *projectStateStub) UpsertMemory(context.Context, MemoryRecord) (StoreUpsertResult, error) {
	return StoreUpsertResult{}, nil
}
func (s *projectStateStub) ListActiveMemories(ctx context.Context, tenantID, subjectID string) ([]MemoryRecord, error) {
	return s.ListMemories(ctx, tenantID, subjectID, false)
}
func (s *projectStateStub) ListMemories(context.Context, string, string, bool) ([]MemoryRecord, error) {
	return nil, nil
}
func (s *projectStateStub) SearchActiveMemories(ctx context.Context, tenantID, subjectID string, patterns []string, limit int) ([]MemoryRecord, error) {
	return nil, nil
}
func (s *projectStateStub) SearchMemories(context.Context, string, string, []string, int, bool) ([]MemoryRecord, error) {
	return nil, nil
}
func (s *projectStateStub) GetMemory(_ context.Context, _, _, memoryID string) (MemoryRecord, error) {
	if r, ok := s.records[memoryID]; ok {
		return r, nil
	}
	return MemoryRecord{}, ErrMemoryNotFound
}
func (s *projectStateStub) MarkSuperseded(context.Context, string, string, string) error { return nil }
func (s *projectStateStub) SuppressMemory(context.Context, string, string, string) error {
	return nil
}
func (s *projectStateStub) CorrectMemory(context.Context, string, string, string, string, string) (MemoryRecord, error) {
	return MemoryRecord{}, nil
}
func (s *projectStateStub) EnqueueIngestJob(context.Context, string, string, string, IngestRequest) (EnqueueResult, error) {
	return EnqueueResult{}, nil
}
func (s *projectStateStub) ClaimNextExtractionJob(context.Context) (ExtractionJob, bool, error) {
	return ExtractionJob{}, false, nil
}
func (s *projectStateStub) CompleteExtractionJob(context.Context, string, string) error { return nil }
func (s *projectStateStub) FailExtractionJob(context.Context, string, string, string) error {
	return nil
}
func (s *projectStateStub) GetCurrentState(_ context.Context, _, _, predicate string) (memoryID, value, policy string, ok bool, err error) {
	id, ok := s.current[predicate]
	return id, s.values[predicate], "", ok, nil
}
func (s *projectStateStub) UpsertCurrentState(_ context.Context, _, _, predicate, memoryID, value, _ string) error {
	s.current[predicate] = memoryID
	s.values[predicate] = value
	return nil
}
func (s *projectStateStub) DeleteCurrentStateByMemory(_ context.Context, _, _, memoryID string) error {
	for pred, id := range s.current {
		if id == memoryID {
			delete(s.current, pred)
			delete(s.values, pred)
		}
	}
	return nil
}

func TestPredicateHintsFromQuery(t *testing.T) {
	h := predicateHintsFromQuery("where does she live now")
	found := false
	for _, p := range h {
		if p == PredicateResidence {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected residence hint, got %#v", h)
	}
	pref := predicateHintsFromQuery("What animal do both Nate and Joanna like?")
	foundPref := false
	for _, p := range pref {
		if p == PredicatePreference {
			foundPref = true
		}
	}
	if !foundPref {
		t.Fatalf("expected preference hint for like-query, got %#v", pref)
	}
	own := predicateHintsFromQuery("How many cars does Calvin own?")
	foundPoss := false
	for _, p := range own {
		if p == PredicatePossession {
			foundPoss = true
		}
	}
	if !foundPoss {
		t.Fatalf("expected possession hint, got %#v", own)
	}
}

func TestTemporalScorePrefersSupersededOnHistorical(t *testing.T) {
	ny := MemoryRecord{
		LifecycleState: LifecycleSuperseded,
		Metadata:       map[string]any{"memory_type": "state", "predicate": PredicateResidence},
	}
	austin := MemoryRecord{
		LifecycleState: LifecycleActive,
		Metadata:       map[string]any{"memory_type": "state", "predicate": PredicateResidence},
	}
	hist := []string{IntentHistoricalState}
	if TemporalScore(ny, hist, true) <= TemporalScore(austin, hist, true) {
		t.Fatal("superseded prior should outrank current on historical intent")
	}
	cur := []string{IntentCurrentState}
	if TemporalScore(austin, cur, false) <= TemporalScore(ny, cur, false) {
		t.Fatal("current-state should prefer active over superseded")
	}
}
