package memory

import "testing"

func TestScoreAndRankV2TemporalAddsChannel(t *testing.T) {
	base, _ := ScoreAndRankV2(0.8, 0.6, 0.25, 0.12, 0.42)
	withT, parts := ScoreAndRankV2Temporal(0.8, 0.6, 0.25, 1.0, 0.12, 0.42)
	if withT <= 0 || withT > 1 {
		t.Fatalf("combined=%v", withT)
	}
	if parts["temporal"] != 1 {
		t.Fatalf("temporal=%v", parts["temporal"])
	}
	zero, zparts := ScoreAndRankV2Temporal(0.8, 0.6, 0.25, 0, 0.12, 0.42)
	if zero != base {
		t.Fatalf("zero temporal must not change fusion, got %v want %v %#v", zero, base, zparts)
	}
}

func TestScoreAndRankV2Additive(t *testing.T) {
	combined, parts := ScoreAndRankV2(0.8, 0.6, 0.25, 0.12, 0.42)
	if combined <= 0 || combined > 1 {
		t.Fatalf("combined out of range: %v %#v", combined, parts)
	}
	if parts["max_possible"] != 2.5 {
		t.Fatalf("expected max_possible 2.5 got %v", parts["max_possible"])
	}
}

func TestScoreAndRankV2SemanticGate(t *testing.T) {
	combined, _ := ScoreAndRankV2(0.05, 0, 0, 0.12, 0.42)
	if combined != 0 {
		t.Fatalf("weak semantic alone should gate to 0, got %v", combined)
	}
	// Entity-probe floor must block mid false-friends.
	combined, details := ScoreAndRankV2(0.55, 0, 0, 0.12, 0.78)
	if combined != 0 {
		t.Fatalf("mid semantic-only should gate at 0.78 floor, got %v %#v", combined, details)
	}
	// Paraphrase floor allows mid-high semantic.
	combined, _ = ScoreAndRankV2(0.55, 0, 0, 0.12, 0.42)
	if combined <= 0 {
		t.Fatalf("paraphrase semantic-only should pass at 0.42 floor, got %v", combined)
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

func TestCandidatePoolSizeMatrix(t *testing.T) {
	for _, n := range []int{30, 50, 100, 200} {
		got := CandidatePoolSize(SearchOptions{Limit: 10, CandidateLimit: n})
		if got != n {
			t.Fatalf("candidate_limit %d got %d", n, got)
		}
	}
	if CandidatePoolSize(SearchOptions{CandidateLimit: 500}) != 200 {
		t.Fatal("pool must stay capped at 200")
	}
	if CandidatePoolSize(SearchOptions{Limit: 10}) != CandidateOverfetch(10) {
		t.Fatal("default pool is overfetch")
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
	if !WantsHistoricalRetrieval(AnalyzeQueryIntents("where did they live before")) {
		t.Fatal("before should request historical retrieval")
	}
	if WantsHistoricalRetrieval(AnalyzeQueryIntents("where do they currently live")) {
		t.Fatal("current-state should not force historical")
	}
	if !WantsHistoricalRetrieval(AnalyzeQueryIntents("how long have I been doing this")) {
		t.Fatal("how long should be temporal")
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

func TestSelectEvidenceSetCoversRareQueryTokens(t *testing.T) {
	ranked := make([]rankedSearchResult, 0, 12)
	for i := 0; i < 10; i++ {
		ranked = append(ranked, rankedSearchResult{result: SearchResult{
			MemoryID: "dump" + itoa(i),
			Content:  "Alex made a cake for the party last week",
			Score:    3.0,
		}})
	}
	ranked = append(ranked, rankedSearchResult{result: SearchResult{
		MemoryID: "fill",
		Content:  "The filling is strawberry",
		Score:    0.4,
	}})
	got := selectEvidenceSetCovering(ranked, 4, tokenize("What filling did Alex use in the cake"))
	found := false
	for _, g := range got {
		if g.result.MemoryID == "fill" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected filling fact in evidence set, got %+v", got)
	}
}

func TestFailureTaxonomyConstants(t *testing.T) {
	if FailRetrievalMiss == "" || AnswerInsufficient == "" {
		t.Fatal("empty constants")
	}
}

func TestDistinctiveQueryTokensKeepsDigitShort(t *testing.T) {
	toks := distinctiveQueryTokens(tokenize("When did Riley start working on his 2D Adventure mobile game?"))
	has2d := false
	for _, tok := range toks {
		if tok == "2d" {
			has2d = true
		}
	}
	if !has2d {
		t.Fatalf("expected 2d in distinctive tokens, got %v", toks)
	}
	plain := distinctiveQueryTokens(tokenize("When did Riley go to the gym?"))
	for _, tok := range plain {
		if len(tok) < 4 {
			t.Fatalf("short non-digit token leaked: %v", plain)
		}
	}
	dated := distinctiveQueryTokens(tokenize("Where did Riley take his family for a road trip on 24 May, 2023?"))
	for _, tok := range dated {
		if tok == "24" {
			t.Fatalf("calendar day token must not be distinctive, got %v", dated)
		}
	}
}
