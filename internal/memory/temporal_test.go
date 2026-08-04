package memory

import (
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
}
