package memory

import "testing"

func TestScoreAndRankV2Additive(t *testing.T) {
	combined, parts := ScoreAndRankV2(0.8, 0.6, 0.25, 0.12)
	if combined <= 0 || combined > 1 {
		t.Fatalf("combined out of range: %v %#v", combined, parts)
	}
	if parts["max_possible"] != 2.5 {
		t.Fatalf("expected max_possible 2.5 got %v", parts["max_possible"])
	}
}

func TestScoreAndRankV2SemanticGate(t *testing.T) {
	combined, _ := ScoreAndRankV2(0.05, 0, 0, 0.12)
	if combined != 0 {
		t.Fatalf("weak semantic alone should gate to 0, got %v", combined)
	}
	// Template false-friend range must not pass semantic-only.
	combined, details := ScoreAndRankV2(0.55, 0, 0, 0.12)
	if combined != 0 {
		t.Fatalf("mid semantic-only should gate to 0, got %v %#v", combined, details)
	}
	combined, _ = ScoreAndRankV2(0.85, 0, 0, 0.12)
	if combined <= 0 {
		t.Fatalf("strong semantic-only should pass, got %v", combined)
	}
}

func TestNormalizeBM25Sigmoid(t *testing.T) {
	if NormalizeBM25Sigmoid(0, 3) != 0 {
		t.Fatal("zero raw")
	}
	hi := NormalizeBM25Sigmoid(20, 3)
	lo := NormalizeBM25Sigmoid(1, 3)
	if !(hi > lo && hi <= 1) {
		t.Fatalf("sigmoid order hi=%v lo=%v", hi, lo)
	}
}

func TestCandidateOverfetch(t *testing.T) {
	if CandidateOverfetch(10) != 60 {
		t.Fatalf("got %d", CandidateOverfetch(10))
	}
	if CandidateOverfetch(40) != 160 {
		t.Fatalf("got %d", CandidateOverfetch(40))
	}
}

func TestPredicatePolicyStateful(t *testing.T) {
	if !IsStatefulPredicate(PredicateResidence) {
		t.Fatal("residence should be stateful")
	}
	if IsStatefulPredicate(PredicateActivity) {
		t.Fatal("activity should accumulate, not supersede")
	}
	if PredicatePolicy(PredicateBelief) != PolicyDerivedBelief {
		t.Fatal("belief policy")
	}
}

func TestAnalyzeQueryIntents(t *testing.T) {
	ints := AnalyzeQueryIntents("what activities does Alex currently enjoy?")
	found := map[string]bool{}
	for _, i := range ints {
		found[i] = true
	}
	if !found[IntentCurrentState] && !found[IntentEnumeration] && !found[IntentPointFact] {
		t.Fatalf("expected useful intents, got %v", ints)
	}
}

func TestSelectEvidenceSetCoversDistinct(t *testing.T) {
	ranked := []rankedSearchResult{
		{result: SearchResult{MemoryID: "1", Content: "Alex plays tennis weekly", Score: 2.0}},
		{result: SearchResult{MemoryID: "2", Content: "Alex plays tennis often", Score: 1.9}},
		{result: SearchResult{MemoryID: "3", Content: "Alex paints watercolor landscapes", Score: 1.5}},
		{result: SearchResult{MemoryID: "4", Content: "Alex volunteers at the library", Score: 1.4}},
	}
	got := selectEvidenceSet(ranked, 3)
	if len(got) != 3 {
		t.Fatalf("want 3 got %d", len(got))
	}
	ids := map[string]bool{}
	for _, g := range got {
		ids[g.result.MemoryID] = true
	}
	// Redundant tennis copy should usually lose to distinct activities.
	if ids["1"] && ids["2"] && !ids["3"] {
		t.Fatalf("expected diversity away from duplicate tennis, got %v", ids)
	}
}

func TestFailureTaxonomyConstants(t *testing.T) {
	if FailRetrievalMiss == "" || AnswerInsufficient == "" {
		t.Fatal("empty constants")
	}
}
