package memory

import "testing"

func TestLooksMultiHopRequiresThreeBearingOrName(t *testing.T) {
	// Ordinary two-token wh-question should NOT force multi-hop corpus scan.
	toks := tokenize("what is launch date")
	if looksMultiHopQuery(toks) {
		t.Fatalf("expected non-multihop for simple wh question, tokens=%v", toks)
	}
	toks = tokenize("what activities does jordan enjoy this year")
	if !looksMultiHopQuery(toks) {
		t.Fatalf("expected multihop for richer question, tokens=%v", toks)
	}
}

func TestAnalyzeQueryIntentsStillWorks(t *testing.T) {
	ints := AnalyzeQueryIntents("what is the current ticket status")
	if len(ints) == 0 {
		t.Fatal("expected intents")
	}
}
