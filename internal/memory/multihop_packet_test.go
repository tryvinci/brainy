package memory

import (
	"strings"
	"testing"
)

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
	// Lexical bridge/direct is context only when typed hops exist — not proven.
	if packetCoverageSatisfied(plan, pkt) {
		t.Fatalf("lexical bridge must not satisfy coverage when hops exist, coverage=%v", pkt.Coverage)
	}
	ans := composeMultiHopAnswer(pkt)
	if ans == "" {
		t.Fatal("expected composed answer")
	}
}

func TestHopJoinBindingSatisfiesCoverage(t *testing.T) {
	plan := PlanQuery("What is Melanie's occupation according to Caroline?", nil)
	if len(plan.Hops) == 0 {
		t.Fatal("expected hops")
	}
	pkt := EvidencePacket{Plan: plan, Coverage: map[string]any{}}
	hops := []HopResult{
		{HopIndex: 0, Kind: "resolve_entity", OutputKey: "e1", Value: "Melanie", MemoryIDs: []string{"m1"}, Contents: []string{"Caroline knows Melanie"}, Source: "typed_store"},
		{HopIndex: 1, Kind: "fetch_predicate", OutputKey: "ans", Entity: "Melanie", Predicate: "occupation", Value: "nurse", MemoryIDs: []string{"m2"}, Contents: []string{"Melanie works as a nurse"}, Source: "typed_store", DependsOn: []string{"e1"}},
	}
	bindPacketFromHopResults(&pkt, hops, map[string]HopResult{"e1": hops[0], "ans": hops[1]})
	if !hopJoinProven(hops) {
		t.Fatal("expected hop join proven")
	}
	if !packetCoverageSatisfied(plan, pkt) {
		t.Fatalf("expected join coverage, coverage=%v items=%+v", pkt.Coverage, pkt.Items)
	}
	if countRole(pkt.Items, "bridge") < 1 || countRole(pkt.Items, "direct") < 1 {
		t.Fatalf("expected bridge+direct from hops, items=%+v", pkt.Items)
	}
}

func TestHopBindKeepsContextEvidence(t *testing.T) {
	plan := PlanQuery("What is Melanie's occupation according to Caroline?", nil)
	pkt := EvidencePacket{
		Plan:      plan,
		Contents:  []string{"Caroline mentioned Melanie last week", "Weather was nice in Seattle"},
		MemoryIDs: []string{"c1", "c2"},
		Coverage:  map[string]any{},
	}
	hops := []HopResult{
		{HopIndex: 0, Kind: "resolve_entity", OutputKey: "e1", Value: "Melanie", MemoryIDs: []string{"m1"}, Contents: []string{"Caroline knows Melanie"}, Source: "typed_store"},
		{HopIndex: 1, Kind: "fetch_predicate", OutputKey: "ans", Entity: "Melanie", Predicate: "occupation", Value: "nurse", MemoryIDs: []string{"m2"}, Contents: []string{"Melanie works as a nurse"}, Source: "typed_store", DependsOn: []string{"e1"}},
	}
	bindPacketFromHopResults(&pkt, hops, nil)
	if len(pkt.ContextEvidence) != 2 {
		t.Fatalf("context_evidence=%v", pkt.ContextEvidence)
	}
	joined := strings.Join(pkt.Contents, " | ")
	if !strings.Contains(joined, "Weather was nice") {
		t.Fatalf("hops must not replace search context, contents=%q", joined)
	}
	if !strings.Contains(joined, "Melanie works as a nurse") {
		t.Fatalf("proof should still be present, contents=%q", joined)
	}
	if len(pkt.ProofChain) < 2 {
		t.Fatalf("proof_chain=%+v", pkt.ProofChain)
	}
	if proven, _ := pkt.Coverage["hop_join_proven"].(bool); !proven {
		t.Fatalf("expected hop_join_proven, coverage=%v", pkt.Coverage)
	}
	if !packetCoverageSatisfied(plan, pkt) {
		t.Fatalf("coverage should stay hop-proven, coverage=%v", pkt.Coverage)
	}
}

func TestLexicalOverlapWithoutJoinStaysUnsatisfied(t *testing.T) {
	plan := PlanQuery("What is Melanie's occupation according to Caroline?", nil)
	results := []SearchResult{
		{MemoryID: "x", Content: "Melanie and Caroline and occupation were mentioned vaguely", Score: 0.9},
	}
	pkt := BuildEvidencePacket(plan, results, nil)
	bindPacketToTargets(&pkt, results, "What is Melanie's occupation according to Caroline?", plan.CoverageTargets)
	// Even if lexical roles fire, without hop_join_proven coverage stays open → second pass.
	if packetCoverageSatisfied(plan, pkt) {
		t.Fatalf("expected unsatisfied without hop join, coverage=%v", pkt.Coverage)
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

func TestBuildTypedHopsResolveThenFetch(t *testing.T) {
	plan := PlanQuery("What is Melanie's occupation according to Caroline?", nil)
	if !plan.NeedsMultiHop {
		t.Fatalf("expected multi-hop, intents=%v", plan.Intents)
	}
	if len(plan.Hops) == 0 {
		t.Fatal("expected typed hops")
	}
	if plan.Hops[0].Kind != "resolve_entity" {
		t.Fatalf("first hop=%+v", plan.Hops[0])
	}
	if plan.Hops[0].Output != "e1" {
		t.Fatalf("expected e1 output, got %+v", plan.Hops[0])
	}
	if len(plan.Hops) > 1 && (len(plan.Hops[1].DependsOn) == 0 || plan.Hops[1].DependsOn[0] != "e1") {
		t.Fatalf("expected fetch DependsOn e1, got %+v", plan.Hops[1])
	}
	probe := nextHopProbe(plan, EvidencePacket{Items: nil, Coverage: map[string]any{"uncovered": []string{"melanie"}}})
	if probe == "" {
		t.Fatal("expected probe")
	}
}

func TestComposeMultiHopAnswerPrefersHopSlotValues(t *testing.T) {
	pkt := EvidencePacket{
		Items: []PacketItem{
			{Role: "direct", Content: "I've known these friends for 4 years, since I moved from my home country"},
			{Role: "bridge", Content: "Jordan plans career in counseling"},
		},
		Coverage: map[string]any{
			"hop_results": []HopResult{
				{Kind: "resolve_entity", OutputKey: "e1", Value: "Jordan", Source: "typed_store"},
				{Kind: "follow_relation", OutputKey: "ans", Entity: "Jordan", Predicate: PredicateOrigin, Value: "portugal", Values: []string{"portugal"}, Source: "typed_store", DependsOn: []string{"e1"}},
			},
		},
	}
	ans := composeMultiHopAnswer(pkt)
	if !strings.Contains(strings.ToLower(ans), "portugal") {
		t.Fatalf("expected hop destination, got %q", ans)
	}
	if strings.Contains(strings.ToLower(ans), "home country") {
		t.Fatalf("anaphora must not win over hop value, got %q", ans)
	}
}

func TestComposeMultiHopAnswerIgnoresResolveOnlyMention(t *testing.T) {
	pkt := EvidencePacket{
		Items: []PacketItem{
			{Role: "bridge", Content: "melanie"},
			{Role: "direct", Content: "We Can Really Accept Who We Are And Be Content"},
		},
		Coverage: map[string]any{
			"hop_results": []HopResult{
				{Kind: "resolve_entity", OutputKey: "e1", Value: "melanie", Source: "search_fallback"},
				{Kind: "fetch_predicate", OutputKey: "ans", Entity: "melanie", Source: "unresolved"},
			},
		},
	}
	ans := composeMultiHopAnswer(pkt)
	if ans != "" {
		t.Fatalf("resolve-only / slogan packet must not compose a factual answer, got %q", ans)
	}
}

func TestGroundHybridToHopValues(t *testing.T) {
	hops := []HopResult{
		{Kind: "follow_relation", Predicate: PredicateOrigin, Value: "portugal", Values: []string{"portugal"}, Source: "typed_store"},
	}
	got := groundToHopValues("I moved from my home country", hops)
	if !strings.Contains(strings.ToLower(got), "portugal") {
		t.Fatalf("got %q", got)
	}
}
