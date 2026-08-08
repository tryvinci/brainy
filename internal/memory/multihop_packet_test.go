package memory

import "testing"

func TestMultiHopTargetsPreferNames(t *testing.T) {
	targets := multiHopTargets("What activities does Melanie partake in?")
	if len(targets) < 2 {
		t.Fatalf("expected >=2 targets, got %v", targets)
	}
	found := false
	for _, x := range targets {
		if x == "melanie" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected melanie in targets, got %v", targets)
	}
}

func TestBindPacketToTargetsBridgeDirect(t *testing.T) {
	plan := PlanQuery("What activities does Melanie partake in with Caroline?", nil)
	if !plan.NeedsMultiHop {
		t.Fatalf("expected multi-hop plan, intents=%v", plan.Intents)
	}
	results := []SearchResult{
		{MemoryID: "b1", Content: "Melanie and Caroline went camping together last week", Score: 0.9},
		{MemoryID: "d1", Content: "Melanie enjoys pottery and painting", Score: 0.8},
		{MemoryID: "c1", Content: "Weather was nice", Score: 0.1},
	}
	pkt := BuildEvidencePacket(plan, results, nil)
	bindPacketToTargets(&pkt, results, "What activities does Melanie partake in with Caroline?", plan.CoverageTargets)
	if countRole(pkt.Items, "bridge") < 1 && countRole(pkt.Items, "direct") < 1 {
		t.Fatalf("expected bridge/direct roles, items=%+v", pkt.Items)
	}
	if !packetCoverageSatisfied(plan, pkt) {
		t.Fatalf("expected coverage satisfied, coverage=%v items=%+v", pkt.Coverage, pkt.Items)
	}
	ans := composeMultiHopAnswer(pkt)
	if ans == "" {
		t.Fatal("expected composed answer")
	}
}

func TestMergeSearchResultsDedupesByID(t *testing.T) {
	a := []SearchResult{{MemoryID: "m1", Content: "a", Score: 0.5}, {MemoryID: "m2", Content: "b", Score: 0.4}}
	b := []SearchResult{{MemoryID: "m1", Content: "a-better", Score: 0.9}, {MemoryID: "m3", Content: "c", Score: 0.3}}
	out := mergeSearchResults(a, b, 10)
	if len(out) != 3 {
		t.Fatalf("len=%d out=%+v", len(out), out)
	}
	for _, r := range out {
		if r.MemoryID == "m1" && r.Score != 0.9 {
			t.Fatalf("expected higher score for m1, got %+v", r)
		}
	}
}

func TestUncoveredTargetsFromCoverage(t *testing.T) {
	pkt := EvidencePacket{Coverage: map[string]any{"uncovered": []string{"sweden", "moved"}}}
	got := uncoveredTargets(pkt)
	if len(got) != 2 || got[0] != "sweden" {
		t.Fatalf("got %v", got)
	}
}
