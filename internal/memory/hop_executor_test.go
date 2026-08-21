package memory

import (
	"context"
	"strings"
	"testing"
)

func TestHopJoinProvenRequiresTypedIDJoin(t *testing.T) {
	alex := CanonicalEntityID("t1", "u1", "Alex")
	dana := CanonicalEntityID("t1", "u1", "Dana")
	hops := []HopResult{
		{Kind: "resolve_entity", OutputKey: "e1", Value: "Alex", EntityID: alex, Source: "typed_store", ProofKind: "typed_exact"},
		{Kind: "fetch_predicate", OutputKey: "ans", Entity: "Alex", EntityID: alex, Predicate: PredicateOccupation, Value: "nurse", Source: "typed_store", ProofKind: "typed_exact", DependsOn: []string{"e1"}},
	}
	if !hopJoinProven(hops) {
		t.Fatal("matching entity IDs must prove the join")
	}
	hops[1].EntityID = dana
	if hopJoinProven(hops) {
		t.Fatal("mismatched entity IDs must not prove the join")
	}
}

func TestComposeFromHopValuesIntersectsEntities(t *testing.T) {
	ans := composeFromHopValues([]HopResult{
		{Kind: "fetch_predicate", Entity: "Nate", Value: "turtles", Values: []string{"turtles", "nintendo games"}, Source: "typed_store"},
		{Kind: "fetch_predicate", Entity: "Joanna", Value: "turtles", Values: []string{"turtles", "writing"}, Source: "typed_store"},
	})
	low := strings.ToLower(ans)
	if !strings.Contains(low, "turtle") {
		t.Fatalf("expected shared turtles, got %q", ans)
	}
	if strings.Contains(low, "nintendo") || strings.Contains(low, "writing") {
		t.Fatalf("union leaked into join answer: %q", ans)
	}
}

func TestComposeFromHopValuesCapsSingleEntityDump(t *testing.T) {
	vals := make([]string, 0, 12)
	for i := 0; i < 12; i++ {
		vals = append(vals, "item"+strings.Repeat("x", i+1))
	}
	ans := composeFromHopValues([]HopResult{
		{Kind: "fetch_predicate", Entity: "Riley", Values: vals, Source: "typed_store"},
	})
	if strings.Count(ans, ",") > 5 {
		t.Fatalf("single-entity hop compose must bound the dump, got %q", ans)
	}
}

func TestComposeFromHopValuesJoinDoesNotFallBackToUnion(t *testing.T) {
	ans := composeFromHopValues([]HopResult{
		{Kind: "fetch_predicate", Entity: "Nate", Value: "turtles", Values: []string{"turtles"}, Source: "typed_store"},
		{Kind: "fetch_predicate", Entity: "Joanna", Value: "writing", Values: []string{"writing"}, Source: "typed_store"},
	})
	if strings.TrimSpace(ans) != "" {
		t.Fatalf("disjoint join must not dump the union, got %q", ans)
	}
}

func TestComposeFromHopValuesContainmentAndPartner(t *testing.T) {
	ans := composeFromHopValues([]HopResult{
		{Kind: "follow_relation", Entity: "Riley", Predicate: PredicateActivity, Values: []string{"yoga", "relaxing", "pottery"}, Source: "typed_store"},
		{Kind: "follow_relation", Entity: "Casey", Predicate: PredicateActivity, Values: []string{"organized yoga", "riley's running group", "ceramics"}, Source: "typed_store"},
	})
	low := strings.ToLower(ans)
	if !strings.Contains(low, "yoga") {
		t.Fatalf("expected yoga containment join, got %q", ans)
	}
	if !strings.Contains(low, "run") {
		t.Fatalf("expected partner-mentioned running group, got %q", ans)
	}
	if strings.Contains(low, "relax") {
		t.Fatalf("unwind slogan leaked into join: %q", ans)
	}
	if strings.Contains(low, "ceram") || strings.Contains(low, "potter") {
		t.Fatalf("private activity leaked into join: %q", ans)
	}
}

func TestComposeFromHopContentsIntersectsEntities(t *testing.T) {
	ans := composeFromHopContents([]HopResult{
		{
			Kind: "fetch_predicate", Entity: "Tim", Source: "search_fallback",
			Contents: []string{"Tim owns a jersey.", "Tim owns a signed baseball."},
		},
		{
			Kind: "fetch_predicate", Entity: "John", Source: "search_fallback",
			Contents: []string{"John owns a jersey.", "John owns vintage posters."},
		},
	})
	low := strings.ToLower(ans)
	if !strings.Contains(low, "jersey") {
		t.Fatalf("expected shared jersey, got %q", ans)
	}
	if strings.Contains(low, "baseball") || strings.Contains(low, "poster") {
		t.Fatalf("union leaked into join contents: %q", ans)
	}
}

func TestHopSlotValuesIgnoreSearchFallback(t *testing.T) {
	vals := hopSlotValues([]HopResult{
		{Kind: "fetch_predicate", Value: "watching pets play", Values: []string{"watching pets play"}, Source: "search_fallback"},
		{Kind: "fetch_predicate", Value: "turtles", Source: "typed_store"},
	})
	if len(vals) != 1 || !strings.EqualFold(vals[0], "turtles") {
		t.Fatalf("typed hop value only, got %#v", vals)
	}
}

func TestHopJoinProvenSearchFallbackIsNotTypedExact(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", OutputKey: "e1", Value: "Alex", Entity: "Alex", Source: "typed_store"},
		{Kind: "fetch_predicate", OutputKey: "ans", Entity: "Alex", Value: "nurse", Source: "search_fallback", ProofKind: "context", DependsOn: []string{"e1"}},
	}
	if hopJoinProven(hops) {
		t.Fatal("search_fallback must not yield hop_join_proven")
	}
}

func TestUnscopedPredicateIsNotTypedExact(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	_ = store.UpsertCurrentState(context.Background(), "t1", "u1", PredicateOccupation, "m-dana", "nurse", "replace")
	_, _ = store.UpsertMemory(context.Background(), MemoryRecord{
		MemoryID: "m-dana", TenantID: "t1", SubjectID: "u1",
		Content: "Dana works as a nurse", DedupeKey: "d1", Status: StatusActive,
		Explain:  map[string]any{"subject": "Dana", "predicate": PredicateOccupation, "value_norm": "nurse"},
		Metadata: map[string]any{"subject": "Dana", "predicate": PredicateOccupation, "value_norm": "nurse", "entity_id": CanonicalEntityID("t1", "u1", "Dana")},
	})
	res := HopResult{Entity: "Alex", EntityID: CanonicalEntityID("t1", "u1", "Alex"), Predicate: PredicateOccupation, Source: "unresolved"}
	svc.fetchPredicateHop(context.Background(), "t1", "u1", "", false, HopStep{
		Kind: "fetch_predicate", Entity: "Alex", Predicate: PredicateOccupation,
	}, &res, 10)
	if res.Source == "typed_store" || res.ProofKind == "typed_exact" {
		t.Fatalf("unscoped occupation must not be typed_exact, got source=%s proof=%s value=%q", res.Source, res.ProofKind, res.Value)
	}
}

func TestContainsEntityMentionDoesNotStealFirstNamePrefix(t *testing.T) {
	if containsEntityMention("John Doe works as a nurse", "John Smith") {
		t.Fatal("full-name mention must not match a different John")
	}
	if !containsEntityMention("John Doe works as a nurse", "John Doe") {
		t.Fatal("full-name mention should match")
	}
}
