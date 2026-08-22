package memory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestHybridReaderReadsReasoningWhenContentEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"finish_reason": "length", "message": map[string]any{
					"content":   nil,
					"reasoning": `prefix {"answer":"ceramics","supporting_memory_ids":[],"unresolved_targets":[],"abstain":false}`,
				}},
			},
		})
	}))
	defer server.Close()
	t.Setenv("BRAINY_RECALL_LLM", "1")
	svc := NewService(newMemoryStoreStub()).WithHybridReader(HybridReaderConfig{
		BaseURL: server.URL, APIKey: "test", Model: "test-model",
	})
	plan := QueryPlan{NeedsMultiHop: true, PrimaryIntent: IntentMultiHop}
	pkt := EvidencePacket{
		Items:    []PacketItem{{MemoryID: "m1", Content: "Melanie loves ceramics"}},
		Contents: []string{"Melanie loves ceramics"},
	}
	res := svc.synthesizeHybridAnswer(context.Background(), "What hobby?", plan, pkt)
	if !res.OK || res.Answer != "ceramics" {
		t.Fatalf("got %#v", res)
	}
}

func TestHybridReaderAcceptsAnswerWithoutSupportIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content": `{"answer":"ceramics","supporting_memory_ids":[],"unresolved_targets":[],"abstain":false}`,
				}},
			},
		})
	}))
	defer server.Close()

	t.Setenv("BRAINY_RECALL_LLM", "1")
	svc := NewService(newMemoryStoreStub()).WithHybridReader(HybridReaderConfig{
		BaseURL: server.URL,
		APIKey:  "test",
		Model:   "test-model",
	})
	plan := QueryPlan{NeedsMultiHop: true, PrimaryIntent: IntentMultiHop, CoverageTargets: []string{"caroline", "hobby"}}
	pkt := EvidencePacket{
		Items: []PacketItem{
			{MemoryID: "m1", Content: "Caroline's friend Melanie loves ceramics", Role: "bridge"},
			{MemoryID: "m2", Content: "Melanie's favorite hobby is pottery", Role: "direct"},
		},
		Contents: []string{"Caroline's friend Melanie loves ceramics", "Melanie's favorite hobby is pottery"},
	}
	res := svc.synthesizeHybridAnswer(context.Background(), "What hobby does Caroline's friend enjoy?", plan, pkt)
	if !res.OK || res.Answer != "ceramics" {
		t.Fatalf("expected soft-grounded answer, got %#v", res)
	}
	if res.Reason != "ok_without_support_ids" {
		t.Fatalf("reason=%q", res.Reason)
	}
}

func TestHybridReaderRecordsFailureReason(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusBadGateway)
	}))
	defer server.Close()

	t.Setenv("BRAINY_RECALL_LLM", "1")
	svc := NewService(newMemoryStoreStub()).WithHybridReader(HybridReaderConfig{
		BaseURL: server.URL,
		APIKey:  "test",
		Model:   "test-model",
	})
	plan := QueryPlan{NeedsMultiHop: true, PrimaryIntent: IntentMultiHop}
	pkt := EvidencePacket{
		Items:    []PacketItem{{MemoryID: "m1", Content: "Alex works with Jordan"}},
		Contents: []string{"Alex works with Jordan"},
	}
	res := svc.synthesizeHybridAnswer(context.Background(), "Who does Alex work with?", plan, pkt)
	if res.OK {
		t.Fatalf("expected failure, got %#v", res)
	}
	if !res.Attempted || res.Reason != "llm_http_status" {
		t.Fatalf("expected attempted llm_http_status, got %#v", res)
	}
}

func TestHybridReaderSkipsPointFactWhenPacketOK(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "1")
	svc := NewService(newMemoryStoreStub()).WithHybridReader(HybridReaderConfig{
		BaseURL: "http://127.0.0.1:1",
		APIKey:  "test",
		Model:   "test-model",
	})
	plan := QueryPlan{
		PrimaryIntent:    IntentPointFact,
		CoverageTargets:  []string{"nurse"},
		NeedsMultiHop:    false,
		NeedsEnumeration: false,
	}
	pkt := EvidencePacket{
		Contents:  []string{"Alex currently works as a nurse in Seattle."},
		MemoryIDs: []string{"m1"},
		Items: []PacketItem{
			{MemoryID: "m1", Content: "Alex currently works as a nurse in Seattle.", Targets: []string{"nurse"}},
		},
	}
	res := svc.synthesizeHybridAnswer(context.Background(), "what does Alex currently do?", plan, pkt)
	if res.OK || res.Attempted {
		t.Fatalf("point-fact should skip hybrid, got %#v", res)
	}
	if res.Reason != "point_fact_packet_ok" {
		t.Fatalf("reason=%q", res.Reason)
	}
}

func TestRecallHybridSetsReaderSource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content": `{"answer":"pottery","supporting_memory_ids":["m2"],"unresolved_targets":[],"abstain":false}`,
				}},
			},
		})
	}))
	defer server.Close()

	t.Setenv("BRAINY_RECALL_LLM", "1")
	store := newMemoryStoreStub()
	svc := NewService(store).WithHybridReader(HybridReaderConfig{
		BaseURL: server.URL,
		APIKey:  "test",
		Model:   "test-model",
	})
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-hyb", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Caroline: my friend Melanie loves pottery and ceramics"},
			{Role: "user", Content: "Caroline: Melanie and I met at school"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-hyb", SubjectID: "u1",
		Query: "What hobby does Caroline's friend Melanie enjoy?",
		Mode:  "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Hybrid may or may not be selected depending on plan.NeedsMultiHop for this query.
	// Force-check synthesize path already covered; here assert OD-like answer is not blanked.
	if strings.TrimSpace(out.Answer) == "" || out.Answer == "not in memory" {
		t.Fatalf("answer blanked: %#v", out)
	}
	if out.Explain["reader_source"] == "hybrid_llm_packet" {
		if out.Explain["hybrid_reader_attempted"] != true {
			t.Fatalf("expected hybrid_reader_attempted, explain=%v", out.Explain)
		}
	}
}

func TestRecallOpenDomainNotBlankedWhenHybridEnabled(t *testing.T) {
	// Hybrid enabled but point-fact packet should stay on evidence_packet.
	t.Setenv("BRAINY_RECALL_LLM", "1")
	_ = os.Setenv("BRAINY_RECALL_LLM", "1")
	store := newMemoryStoreStub()
	svc := NewService(store).WithHybridReader(HybridReaderConfig{
		BaseURL: "http://127.0.0.1:1", // would fail if called
		APIKey:  "test",
		Model:   "test-model",
	})
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-od", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Alex currently works as a nurse in Seattle."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-od", SubjectID: "u1",
		Query: "what does Alex currently do?",
		Mode:  "answer", TopK: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Abstained || strings.TrimSpace(out.Answer) == "" || out.Answer == "not in memory" {
		t.Fatalf("OD answer blanked under hybrid enable: %#v", out)
	}
	if out.Explain["reader_source"] == "hybrid_llm_packet" {
		t.Fatalf("point-fact should not use hybrid, explain=%v", out.Explain)
	}
}

func TestShouldAttemptHybridMultiHop(t *testing.T) {
	ok, reason := shouldAttemptHybrid(
		"What hobby does Melanie enjoy?",
		QueryPlan{NeedsMultiHop: true},
		EvidencePacket{Contents: []string{"a"}, Items: []PacketItem{{Content: "a"}}},
	)
	if !ok || reason != "composition_needed" {
		t.Fatalf("ok=%v reason=%q", ok, reason)
	}
}

func TestShouldAttemptHybridWhenEventDespiteTemporalAnswer(t *testing.T) {
	ok, reason := shouldAttemptHybrid(
		"When did Jordan get an ankle injury in 2023?",
		QueryPlan{NeedsTemporal: true},
		EvidencePacket{
			Contents:       []string{"Jordan got an ankle injury"},
			TemporalAnswer: "7 May 2023",
			Items:          []PacketItem{{Content: "Jordan got an ankle injury"}},
		},
	)
	if !ok {
		t.Fatalf("when-event must still attempt hybrid, reason=%q", reason)
	}
}

func TestShouldAttemptHybridSkipsResolvedNonWhenTemporal(t *testing.T) {
	ok, reason := shouldAttemptHybrid(
		"How long ago was the injury?",
		QueryPlan{NeedsTemporal: true},
		EvidencePacket{
			Contents:       []string{"Jordan got an ankle injury"},
			TemporalAnswer: "7 May 2023",
			Items:          []PacketItem{{Content: "Jordan got an ankle injury"}},
		},
	)
	if ok || reason != "temporal_resolved" {
		t.Fatalf("ok=%v reason=%q", ok, reason)
	}
}

func TestHybridAnswerStatusTruthful(t *testing.T) {
	plan := QueryPlan{NeedsMultiHop: true, Hops: []HopStep{{Kind: "resolve_entity"}, {Kind: "fetch_predicate"}}}
	pktOK := EvidencePacket{Coverage: map[string]any{"hop_join_proven": true, "uncovered": []string{}}}
	got := hybridAnswerStatus(hybridReaderResult{Answer: "nurse", OK: true}, plan, pktOK, true)
	if got != AnswerSupported {
		t.Fatalf("supported got %s", got)
	}
	pktPartial := EvidencePacket{Coverage: map[string]any{"hop_join_proven": false, "uncovered": []string{"e1"}}}
	got = hybridAnswerStatus(hybridReaderResult{Answer: "maybe", OK: true, UnresolvedTargets: []string{"where"}}, plan, pktPartial, false)
	if got != AnswerPartiallySupported {
		t.Fatalf("partial got %s", got)
	}
	got = hybridAnswerStatus(hybridReaderResult{Abstain: true}, plan, pktOK, true)
	if got != AnswerInsufficient {
		t.Fatalf("insufficient got %s", got)
	}
	conflictPkt := EvidencePacket{
		Items: []PacketItem{
			{Predicate: "occupation", Content: "Jordan is a nurse"},
			{Predicate: "occupation", Content: "Jordan is a doctor"},
		},
	}
	got = hybridAnswerStatus(hybridReaderResult{Answer: "doctor", OK: true}, plan, conflictPkt, true)
	if got != AnswerConflicted {
		t.Fatalf("conflicted got %s", got)
	}
}

func TestFormatHybridMemoryLinesIncludesHopChain(t *testing.T) {
	pkt := EvidencePacket{
		Coverage: map[string]any{
			"hop_results": []HopResult{
				{HopIndex: 0, Kind: "resolve_entity", OutputKey: "e1", Value: "Melanie", Source: "typed_store", Contents: []string{"Caroline knows Melanie"}, MemoryIDs: []string{"m1"}},
				{HopIndex: 1, Kind: "fetch_predicate", OutputKey: "ans", DependsOn: []string{"e1"}, Value: "nurse", Source: "typed_store", Contents: []string{"Melanie is a nurse"}, MemoryIDs: []string{"m2"}},
			},
		},
	}
	lines := formatHybridMemoryLines(pkt)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "Hop chain:") || !strings.Contains(joined, "depends_on=e1") {
		t.Fatalf("expected hop chain lines, got %q", joined)
	}
}

func TestFormatHybridMemoryLinesLeadsWithStructuredSlots(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Alex lives in New York", Role: "context"},
			{Content: "Alex later moved to Austin", Role: "context"},
		},
		Coverage: map[string]any{
			"hop_results": []HopResult{
				{HopIndex: 0, Kind: "fetch_predicate", Predicate: PredicateResidence, Value: "Austin", Source: "typed_store", Contents: []string{"Alex lives in Austin"}},
			},
		},
	}
	lines := formatHybridMemoryLines(pkt)
	if len(lines) < 2 {
		t.Fatalf("lines=%v", lines)
	}
	if lines[0] != "Structured:" {
		t.Fatalf("structured slots must lead, first=%q", lines[0])
	}
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "New York") {
		t.Fatalf("context must still be present, lines=%v", lines)
	}
}

func TestFormatHybridMemoryLinesSkipsUnprovenHopDumps(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "The filling is strawberry", MemoryID: "m-fill"},
		},
		Coverage: map[string]any{
			"hop_results": []HopResult{
				{HopIndex: 0, Kind: "resolve_entity", OutputKey: "e1", Value: "Alex", Source: "search_fallback"},
				{HopIndex: 1, Kind: "fetch_predicate", Entity: "Alex", Source: "search_fallback", ProofKind: "context",
					Contents: []string{"Alex does pottery", "Alex goes to the beach"}},
			},
		},
	}
	joined := strings.Join(formatHybridMemoryLines(pkt), "\n")
	if strings.Contains(joined, "Hop chain:") || strings.Contains(joined, "pottery") {
		t.Fatalf("unproven hops must not dump into hybrid prompt, got %q", joined)
	}
	if !strings.Contains(joined, "strawberry") {
		t.Fatalf("context filling fact must remain, got %q", joined)
	}
}

func TestFormatHybridMemoryLinesSkipsSlotsMissingQueryTokens(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Maria has a dog named Shadow.", MemoryID: "m-shadow"},
			{Content: "Maria got a puppy named Coco on 28 July 2023.", MemoryID: "m-coco"},
		},
		Coverage: map[string]any{
			"hop_results": []HopResult{
				{HopIndex: 0, Kind: "resolve_entity", Entity: "Maria", Value: "Maria", Source: "search_fallback"},
				{HopIndex: 1, Kind: "follow_relation", Entity: "Maria", Predicate: PredicateIdentity, Source: "typed_store",
					Value: "inspiration, family, team", Values: []string{"inspiration", "family", "team"},
					Contents: []string{"Maria is a inspiration"}},
			},
		},
	}
	joined := strings.Join(formatHybridMemoryLinesForQuery("What is the name of Maria's second puppy?", pkt), "\n")
	if strings.Contains(joined, "Structured:") || strings.Contains(joined, "inspiration") {
		t.Fatalf("unrelated identity slots must not lead hybrid prompt, got %q", joined)
	}
	if !strings.Contains(joined, "Shadow") {
		t.Fatalf("covering puppy memory must remain, got %q", joined)
	}
	if strings.Contains(joined, "inspiration") {
		t.Fatalf("identity slogan leaked into covering-only prompt, got %q", joined)
	}
}

func TestFormatHybridMemoryLinesSkipsActivityDumpForCountryQuery(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Tim has experiences in the United Kingdom.", MemoryID: "m-uk"},
			{Content: "Tim is researching visa requirements for the countries he wants to visit.", MemoryID: "m-visa"},
		},
		Coverage: map[string]any{
			"hop_results": []HopResult{
				{HopIndex: 0, Kind: "resolve_entity", Entity: "Tim", Value: "Tim", Source: "typed_store"},
				{HopIndex: 1, Kind: "follow_relation", Entity: "Tim", EntityID: "e-tim", Predicate: PredicateActivity, Source: "typed_store",
					Value: "visa requirements", Values: []string{"visa requirements", "explore nature", "traveling"}},
			},
		},
	}
	q := "which country has Tim visited most frequently in his travels?"
	hops, _ := pkt.Coverage["hop_results"].([]HopResult)
	if hopsKeepTypedJoin(hops) {
		t.Fatal("activity dumps must not count as typed skill/possession/preference joins")
	}
	if !skipUnrelatedHopSlots(q, hops, pkt) {
		t.Fatalf("skip=false leftover=%v slots=%v", leftoverNonEntityQueryTokens(q, hops), hopSlotValues(hops))
	}
	joined := strings.Join(formatHybridMemoryLinesForQuery(q, pkt), "\n")
	if strings.Contains(strings.ToLower(joined), "visa requirements") && strings.Contains(joined, "Structured:") {
		t.Fatalf("activity dump must not lead hybrid prompt, got %q", joined)
	}
	if !strings.Contains(joined, "United Kingdom") {
		t.Fatalf("covering country memory must remain, got %q", joined)
	}
}

func TestSkipUnrelatedIdentityDumpLiveShape(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "maria", Source: "search_fallback", Value: "Maria"},
		{Kind: "follow_relation", Entity: "Maria", Predicate: PredicateIdentity, Source: "typed_store",
			Value: "inspiration, family, team", Values: []string{"inspiration", "walk or a song can totally switch up our", "family", "team"}},
	}
	pkt := EvidencePacket{
		Contents:        []string{"Maria has a new puppy.", "Maria has a dog named Shadow."},
		ContextEvidence: []PacketItem{{Content: "Maria has a dog named Shadow."}},
		Coverage:        map[string]any{"hop_results": hops},
	}
	q := "What is the name of Maria's second puppy?"
	if !skipUnrelatedHopSlots(q, hops, pkt) {
		t.Fatalf("skip=false toks=%v slots=%v fetchN=%d", distinctiveQueryTokens(tokenize(q)), hopSlotValues(hops), hopFetchEntityCount(hops))
	}
	joined := strings.Join(formatHybridMemoryLinesForQuery(q, pkt), "\n")
	if strings.Contains(joined, "Structured:") || strings.Contains(joined, "inspiration") {
		t.Fatalf("identity dump must not lead, got %q", joined)
	}
}

func TestFormatHybridMemoryLinesKeepsLeftoverIdentityHopContent(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Maria has a dog named Shadow.", MemoryID: "m-shadow"},
		},
		Coverage: map[string]any{
			"hop_results": []HopResult{
				{HopIndex: 0, Kind: "resolve_entity", Entity: "Maria", Value: "Maria", Source: "search_fallback"},
				{HopIndex: 1, Kind: "follow_relation", Entity: "Maria", Predicate: PredicateIdentity, Source: "typed_store",
					Value: "inspiration, family, team", Values: []string{"inspiration", "family", "team"},
					Contents: []string{
						"Maria is a inspiration",
						"Maria got a puppy named Coco on 28 July 2023.",
					},
					MemoryIDs: []string{"m-insp", "m-coco"},
				},
			},
		},
	}
	q := "What is the name of Maria's second puppy?"
	hops, _ := pkt.Coverage["hop_results"].([]HopResult)
	if !hopsAreIdentityOnly(hops) {
		t.Fatal("fixture must be identity-only")
	}
	if !skipUnrelatedHopSlots(q, hops, pkt) {
		t.Fatal("identity dump must skip")
	}
	joined := strings.Join(formatHybridMemoryLinesForQuery(q, pkt), "\n")
	if strings.Contains(joined, "Structured:") || strings.Contains(joined, "inspiration") {
		t.Fatalf("identity dump must not lead, got %q", joined)
	}
	if !strings.Contains(joined, "Coco") {
		t.Fatalf("leftover-covering puppy hop content must remain, got %q", joined)
	}
}

func TestFormatHybridMemoryLinesKeepsSkillSlotsForInstrumentQuery(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "You play any instruments", MemoryID: "m0"},
			{Content: "Melanie plays the clarinet.", MemoryID: "m1"},
			{Content: "Melanie plays the violin.", MemoryID: "m2"},
		},
		Coverage: map[string]any{
			"hop_results": []HopResult{
				{HopIndex: 0, Kind: "fetch_predicate", Entity: "Melanie", Predicate: PredicateSkill, Source: "typed_store",
					Value: "clarinet, violin", Values: []string{"clarinet", "violin"}},
			},
		},
	}
	joined := strings.Join(formatHybridMemoryLinesForQuery("What instruments does Melanie play?", pkt), "\n")
	if !strings.Contains(joined, "Structured:") || !strings.Contains(joined, "clarinet") {
		t.Fatalf("skill slots must stay for instrument questions, got %q", joined)
	}
}

func TestFormatHybridMemoryLinesKeepsDualEntityJoinSlots(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Caroline plays the clarinet.", MemoryID: "m1"},
			{Content: "Melanie plays the violin.", MemoryID: "m2"},
		},
		Coverage: map[string]any{
			"hop_results": []HopResult{
				{HopIndex: 0, Kind: "fetch_predicate", Entity: "Caroline", Predicate: PredicateSkill, Source: "typed_store", Value: "clarinet", Values: []string{"clarinet"}},
				{HopIndex: 1, Kind: "fetch_predicate", Entity: "Melanie", Predicate: PredicateSkill, Source: "typed_store", Value: "violin", Values: []string{"violin"}},
			},
		},
	}
	joined := strings.Join(formatHybridMemoryLinesForQuery("What instruments do both Caroline and Melanie play?", pkt), "\n")
	if !strings.Contains(joined, "Structured:") || !strings.Contains(joined, "clarinet") {
		t.Fatalf("dual-entity join slots must stay, got %q", joined)
	}
}

func TestFormatHybridMemoryLinesSkipsDualEntityActivityDump(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Andrew and Buddy enjoy walking together.", MemoryID: "m-walk"},
			{Content: "Andrew is curious about trying sushi at a new restaurant.", MemoryID: "m-sushi"},
		},
		Coverage: map[string]any{
			"hop_results": []HopResult{
				{HopIndex: 0, Kind: "resolve_entity", Entity: "Andrew", Value: "Andrew", Source: "search_fallback"},
				{HopIndex: 1, Kind: "resolve_entity", Entity: "Buddy", Value: "Buddy", Source: "typed_store"},
				{HopIndex: 2, Kind: "follow_relation", Entity: "Andrew", Predicate: PredicateActivity, Source: "typed_store",
					Value: "sushi", Values: []string{"creating safe and fun space for scout", "sushi", "a new sushi restaurant in town"}},
				{HopIndex: 3, Kind: "follow_relation", Entity: "Buddy", Predicate: PredicateActivity, Source: "search_fallback",
					Value: "sushi", Values: []string{"discover new places to eat around town", "sushi"}},
			},
		},
	}
	q := "What activity do Andrew and Buddy enjoy doing together?"
	hops, _ := pkt.Coverage["hop_results"].([]HopResult)
	if hopsKeepTypedJoin(hops) {
		t.Fatal("activity dumps must not count as typed skill joins")
	}
	if !skipUnrelatedHopSlots(q, hops, pkt) {
		t.Fatalf("dual-entity activity dump must skip, leftover=%v slots=%v fetchN=%d ents=%v",
			leftoverNonEntityQueryTokens(q, hops), hopSlotValues(hops), hopFetchEntityCount(hops), hopQueryEntities(q))
	}
	joined := strings.Join(formatHybridMemoryLinesForQuery(q, pkt), "\n")
	if strings.Contains(joined, "Structured:") {
		t.Fatalf("activity dump must not lead hybrid prompt, got %q", joined)
	}
	if !strings.Contains(joined, "walking") {
		t.Fatalf("covering walking memory must remain, got %q", joined)
	}
}

func TestTypedAnswerIsHopDump(t *testing.T) {
	if typedAnswerIsHopDump("clarinet, violin") {
		t.Fatal("short skill list")
	}
	if typedAnswerIsHopDump("Oliver, Luna, Bailey") {
		t.Fatal("short name list")
	}
	if typedAnswerIsHopDump("jersey") {
		t.Fatal("single typed value")
	}
	if !typedAnswerIsHopDump("Way, Road Trip, McGee's Bar, Playing Cyberpunk 2077, Notebook, First, Simple Dishes, Tried Cyberpunk 2077") {
		t.Fatal("title-case where dump")
	}
	if !typedAnswerIsHopDump("sushi, sushi before, curious about trying sushi, a new sushi restaurant in town in mid-October 2023") {
		t.Fatal("dual-entity sushi dump")
	}
}

func TestWhereAnswerFromHopsUsesLocativeCover(t *testing.T) {
	hops := []HopResult{
		{Kind: "follow_relation", Entity: "Jolene", Predicate: PredicateActivity, Source: "typed_store",
			Contents: []string{
				"Jolene and her partner attended a yoga retreat in South America.",
				"Jolene attended a meditation retreat in Phuket with her partner, starting on 9 September 2023.",
				"Jolene and her partner found a cool diving spot in Bali.",
			}},
	}
	q := "Where did Jolene and her partner find a cool diving spot?"
	ans := whereAnswerFromHops(q, hops)
	low := strings.ToLower(ans)
	if !strings.Contains(low, "bali") {
		t.Fatalf("locative leftover should keep the diving-spot place, got %q", ans)
	}
	if strings.Contains(low, "phuket") || strings.Contains(low, "south") {
		t.Fatalf("unrelated retreats must not fill a locative where-answer, got %q", ans)
	}
}

func TestPrioritizeHybridMemoryLinesKeepsHyphenatedEvents(t *testing.T) {
	lines := make([]string, 0, 20)
	for i := 0; i < 15; i++ {
		lines = append(lines, "- Maria volunteered at a homeless shelter during a kids' event.")
	}
	lines = append(lines, "- Maria's fundraiser will feature a chili cook-off event.")
	lines = append(lines, "- Maria is planning a ring-toss tournament fundraiser for the homeless shelter.")
	hops := []HopResult{
		{Kind: "follow_relation", Entity: "Maria", Predicate: PredicatePlan, Source: "typed_store",
			Value: "ring-toss", Values: []string{"ring-toss tournament fundraiser"}},
	}
	got := prioritizeHybridMemoryLines("What events is Maria planning for the homeless shelter funraiser?", hops, lines)
	joined := strings.ToLower(strings.Join(got, "\n"))
	if !strings.Contains(joined, "chili") {
		t.Fatalf("chili cook-off must survive leftover-cover crowding, got %q", joined)
	}
	if !strings.Contains(joined, "ring-toss") && !strings.Contains(joined, "ring") {
		t.Fatalf("ring-toss must remain, got %q", joined)
	}
}

func TestFormatHybridMemoryLinesKeepsPlaceHopContentWhenSkipping(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Jolene participates in diving.", MemoryID: "m-div"},
		},
		Coverage: map[string]any{
			"hop_results": []HopResult{
				{HopIndex: 0, Kind: "resolve_entity", Entity: "Jolene", Value: "Jolene", Source: "search_fallback"},
				{HopIndex: 1, Kind: "follow_relation", Entity: "Jolene", Predicate: PredicateActivity, Source: "typed_store",
					Value:  "hiking, yoga, meditation retreat in phuket, diving",
					Values: []string{"hiking", "yoga", "meditation retreat in phuket", "diving"},
					Contents: []string{
						"Jolene attended a meditation retreat in Phuket with her partner.",
						"Jolene went hiking last weekend.",
					},
					MemoryIDs: []string{"m-phu", "m-hike"},
				},
			},
		},
	}
	q := "Where did Jolene and her partner find a cool diving spot?"
	hops, _ := pkt.Coverage["hop_results"].([]HopResult)
	if hopsKeepTypedJoin(hops) {
		t.Fatal("activity dump without typed kin must not count as typed join")
	}
	if !skipUnrelatedHopSlots(q, hops, pkt) {
		t.Fatal("expected skip")
	}
	joined := strings.Join(formatHybridMemoryLinesForQuery(q, pkt), "\n")
	if strings.Contains(joined, "Structured:") {
		t.Fatalf("activity dump must not lead, got %q", joined)
	}
	if !strings.Contains(joined, "Phuket") {
		t.Fatalf("place hop content must remain when skipping dumps, got %q", joined)
	}
}

func TestFormatHybridMemoryLinesKeepsHyphenatedHopContentWhenSkipping(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Maria volunteered at a homeless shelter during a kids' event.", MemoryID: "m-vol"},
		},
		Coverage: map[string]any{
			"hop_results": []HopResult{
				{HopIndex: 0, Kind: "follow_relation", Entity: "Maria", Predicate: PredicateEvent, Source: "typed_store",
					Value:     "picnic, chili cook-off, charity event",
					Values:    []string{"picnic", "chili cook-off", "charity event"},
					Contents:  []string{"Maria's fundraiser will feature a chili cook-off event."},
					MemoryIDs: []string{"m-chili"},
				},
			},
		},
	}
	q := "What events is Maria planning for the homeless shelter funraiser?"
	hops, _ := pkt.Coverage["hop_results"].([]HopResult)
	if !skipUnrelatedHopSlots(q, hops, pkt) {
		t.Fatal("event dump with leftover tokens should skip")
	}
	joined := strings.Join(formatHybridMemoryLinesForQuery(q, pkt), "\n")
	if strings.Contains(joined, "Structured:") {
		t.Fatalf("dump must not lead, got %q", joined)
	}
	if !strings.Contains(strings.ToLower(joined), "chili") {
		t.Fatalf("chili hop content must remain, got %q", joined)
	}
}

func TestHopsKeepTypedJoinKinshipActivity(t *testing.T) {
	hops := []HopResult{
		{Kind: "follow_relation", Entity: "Deborah", Predicate: PredicateFamilyMember, Source: "typed_store",
			Value: "mother", Values: []string{"mother"}},
		{Kind: "follow_relation", Entity: "Deborah's mother", Predicate: PredicateActivity, Source: "typed_store",
			Value: "reading", Values: []string{"reading", "travel", "art", "cooking"}},
	}
	if !hopsKeepTypedJoin(hops) {
		t.Fatal("typed kinship dest activity is a join")
	}
	pkt := EvidencePacket{Contents: []string{"Deborah's mother had reading as one of her hobbies."}}
	if skipUnrelatedHopSlots("What were Deborah's mother's hobbies?", hops, pkt) {
		t.Fatal("kinship activity join must not skip")
	}
}

func TestPrioritizeHybridMemoryLinesPrefersCovering(t *testing.T) {
	lines := make([]string, 0, 40)
	for i := 0; i < 40; i++ {
		lines = append(lines, "- filler slogan about cars and tours "+strings.Repeat("x", i%3))
	}
	lines = append(lines, "- Evan took his family on a road trip to Jasper National Park.")
	hops := []HopResult{
		{Kind: "follow_relation", Entity: "Evan", Predicate: PredicateActivity, Source: "typed_store",
			Value: "car spin", Values: []string{"car spin", "city tour", "setback"}},
	}
	got := prioritizeHybridMemoryLines("Where did Evan take his family for a road trip to Jasper?", hops, lines)
	if len(got) > hybridMemoryLineLimit {
		t.Fatalf("cap %d got %d", hybridMemoryLineLimit, len(got))
	}
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, "Jasper") {
		t.Fatalf("covering place must survive cap, got %q", joined)
	}
	if idx := strings.Index(joined, "Jasper"); idx < 0 || (strings.Contains(joined, "filler slogan") && idx > strings.Index(joined, "filler slogan")) {
		t.Fatalf("covering line should rank before filler, got %q", joined)
	}
}

func TestPrioritizeHybridMemoryLinesKeepsSpecificPlaceFacts(t *testing.T) {
	lines := make([]string, 0, 42)
	for i := 0; i < 40; i++ {
		lines = append(lines, "- Tim participates in hobby")
	}
	lines = append(lines, "- Tim has experiences in the United Kingdom.")
	hops := []HopResult{
		{Kind: "follow_relation", Entity: "Tim", Predicate: PredicateActivity, Source: "typed_store",
			Value: "visa requirements", Values: []string{"visa requirements", "explore nature", "traveling"}},
	}
	got := prioritizeHybridMemoryLines("which country has Tim visited most frequently in his travels?", hops, lines)
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, "United Kingdom") {
		t.Fatalf("specific country fact must survive thin participates-in crowding, got %q", joined)
	}
}

func TestPrioritizeHybridMemoryLinesKeepsUKAmongVisaCover(t *testing.T) {
	lines := make([]string, 0, 42)
	for i := 0; i < 40; i++ {
		lines = append(lines, "- Tim is researching visa requirements for the countries he wants to visit.")
	}
	lines = append(lines, "- Tim has experiences in the United Kingdom.")
	hops := []HopResult{
		{Kind: "follow_relation", Entity: "Tim", Predicate: PredicateActivity, Source: "typed_store",
			Value: "visa requirements", Values: []string{"visa requirements", "explore nature", "traveling"}},
	}
	got := prioritizeHybridMemoryLines("which country has Tim visited most frequently in his travels?", hops, lines)
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, "United Kingdom") {
		t.Fatalf("UK fact must survive leftover-covering visa crowding, got %q", joined)
	}
}

func TestPrioritizeHybridMemoryLinesKeepsGymAmongWorkCover(t *testing.T) {
	lines := make([]string, 0, 42)
	for i := 0; i < 40; i++ {
		lines = append(lines, "- Maria does community work with church friends.")
	}
	lines = append(lines, "- Maria joined a gym on 2023-06-09.")
	hops := []HopResult{
		{Kind: "follow_relation", Entity: "Maria", Predicate: PredicateActivity, Source: "typed_store",
			Value: "community work", Values: []string{"community work with church friends", "volunteered at shelter"}},
	}
	got := prioritizeHybridMemoryLines("Where does Maria go to work out?", hops, lines)
	joined := strings.Join(got, "\n")
	if !strings.Contains(strings.ToLower(joined), "gym") {
		t.Fatalf("gym fact must survive leftover work covering, got %q", joined)
	}
}

func TestIsHybridGarbageAnswer(t *testing.T) {
	if !isHybridGarbageAnswer(strings.Repeat("!", 80)) {
		t.Fatal("repeated punctuation")
	}
	if !isHybridGarbageAnswer("none") {
		t.Fatal("none")
	}
	if isHybridGarbageAnswer("Shadow") {
		t.Fatal("name")
	}
	if isHybridGarbageAnswer("Luna, Oliver, Bailey") {
		t.Fatal("list")
	}
	if isHybridGarbageAnswer("20 May 2023") {
		t.Fatal("date")
	}
	if !isHybridGarbageAnswer(`We need to answer: "What habits does Jolene practice?" Search memories: - [mem_1] yoga`) {
		t.Fatal("leaked hybrid prompt")
	}
}

func TestHybridReaderRejectsPunctuationFreeform(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content": strings.Repeat("!", 200),
				}},
			},
		})
	}))
	defer server.Close()
	t.Setenv("BRAINY_RECALL_LLM", "1")
	svc := NewService(newMemoryStoreStub()).WithHybridReader(HybridReaderConfig{
		BaseURL: server.URL, APIKey: "test", Model: "test-model",
	})
	plan := QueryPlan{NeedsEnumeration: true}
	pkt := EvidencePacket{
		Items:    []PacketItem{{MemoryID: "m1", Content: "Jolene has a pet snake named Seraphim."}},
		Contents: []string{"Jolene has a pet snake named Seraphim."},
	}
	res := svc.synthesizeHybridAnswer(context.Background(), "What are the names of Jolene's snakes?", plan, pkt)
	if res.OK || res.Answer != "" {
		t.Fatalf("garbage freeform must not be accepted, got %#v", res)
	}
}

func TestHybridReaderStripsThinkAndExtractsJSON(t *testing.T) {
	raw := "<think>plan</think>\n{\"answer\":\"pottery\",\"supporting_memory_ids\":[],\"unresolved_targets\":[],\"abstain\":false}"
	got := stripThinkTags(raw)
	if strings.Contains(strings.ToLower(got), "<think>") {
		t.Fatalf("think remains: %q", got)
	}
	obj := extractJSONObject("intro " + got + " trailing")
	var out struct {
		Answer string `json:"answer"`
	}
	if err := json.Unmarshal([]byte(obj), &out); err != nil || out.Answer != "pottery" {
		t.Fatalf("obj=%q out=%#v err=%v", obj, out, err)
	}
}

func TestItemsFromCommaAnswer(t *testing.T) {
	got := itemsFromCommaAnswer("running, pottery, running")
	if len(got) != 2 || got[0].Value != "running" || got[1].Value != "pottery" {
		t.Fatalf("got %#v", got)
	}
	if itemsFromCommaAnswer("not in memory") != nil {
		t.Fatal("empty sentinel should yield no items")
	}
}

func TestRecallHybridReplacesEnumerateDump(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content": `{"answer":"horseback riding","supporting_memory_ids":[],"unresolved_targets":[],"abstain":false}`,
				}},
			},
		})
	}))
	defer server.Close()

	t.Setenv("BRAINY_RECALL_LLM", "1")
	store := newMemoryStoreStub()
	svc := NewService(store).WithHybridReader(HybridReaderConfig{
		BaseURL: server.URL,
		APIKey:  "test",
		Model:   "test-model",
	})
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-enum-hyb", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Caroline used to go horseback riding with her dad."},
			{Role: "user", Content: "Caroline also likes camping trips and quality time in nature."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-enum-hyb", SubjectID: "u1",
		Query: "What activities does Caroline enjoy?",
		Mode:  "enumerate", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Explain["reader_source"] != "hybrid_llm_packet" {
		t.Fatalf("enumerate must be hybrid-reachable, explain=%v answer=%q", out.Explain, out.Answer)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "horseback") {
		t.Fatalf("hybrid enumerate answer=%q items=%#v", out.Answer, out.Items)
	}
	if len(out.Items) == 0 || !strings.Contains(strings.ToLower(out.Items[0].Value), "horseback") {
		t.Fatalf("harness reads items; got %#v", out.Items)
	}
}

func TestRecallHybridMayOverwriteDateAnswer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content": `{"answer":"the sunday before 25 May 2023","supporting_memory_ids":[],"unresolved_targets":[],"abstain":false}`,
				}},
			},
		})
	}))
	defer server.Close()

	t.Setenv("BRAINY_RECALL_LLM", "1")
	store := newMemoryStoreStub()
	svc := NewService(store).WithHybridReader(HybridReaderConfig{
		BaseURL: server.URL,
		APIKey:  "test",
		Model:   "test-model",
	})
	now := svc.now()
	d := time.Date(2023, 5, 7, 0, 0, 0, 0, time.UTC)
	store.records["ja"] = MemoryRecord{
		MemoryID: "mem_ja3", TenantID: "t-hyb-when", SubjectID: "u1",
		Kind: KindFact, Content: "Jordan got an ankle injury",
		DedupeKey: "ja", Status: StatusActive, UpdatedAt: now, ObservedAt: &d,
		Metadata: map[string]any{"predicate": PredicateHealth, "value_norm": "ankle", "subject": "Jordan"},
		Explain:  map[string]any{"predicate": PredicateHealth, "value_norm": "ankle", "subject": "Jordan"},
	}
	store.atoms = append(store.atoms, stubAtom{pred: PredicateHealth, val: "ankle", memID: "mem_ja3"})
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-hyb-when", SubjectID: "u1",
		Query: "When did Jordan get an ankle injury in 2023?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(out.Answer), "sunday") {
		t.Fatalf("date lock must not block hybrid relative dates: %q explain=%v", out.Answer, out.Explain)
	}
}

func TestRecallHybridDoesNotOverwriteCountAnswer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content": `{"answer":"one","supporting_memory_ids":[],"unresolved_targets":[],"abstain":false}`,
				}},
			},
		})
	}))
	defer server.Close()

	t.Setenv("BRAINY_RECALL_LLM", "1")
	store := newMemoryStoreStub()
	svc := NewService(store).WithHybridReader(HybridReaderConfig{
		BaseURL: server.URL, APIKey: "test", Model: "test-model",
	})
	now := svc.now()
	store.records["cal-car1"] = MemoryRecord{
		MemoryID: "mem_c1", TenantID: "t-hyb-count", SubjectID: "u1",
		Kind: KindFact, Content: "Calvin owns a red coupe",
		DedupeKey: "c1", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "red coupe", "subject": "Calvin"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "red coupe", "subject": "Calvin"},
	}
	store.records["cal-car2"] = MemoryRecord{
		MemoryID: "mem_c2", TenantID: "t-hyb-count", SubjectID: "u1",
		Kind: KindFact, Content: "Calvin owns a black sedan",
		DedupeKey: "c2", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "black sedan", "subject": "Calvin"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "black sedan", "subject": "Calvin"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicatePossession, val: "red coupe", memID: "mem_c1"},
		stubAtom{pred: PredicatePossession, val: "black sedan", memID: "mem_c2"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-hyb-count", SubjectID: "u1",
		Query: "How many cars does Calvin own?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Answer != "2" {
		t.Fatalf("count_answer should stay locked, got %q explain=%v", out.Answer, out.Explain)
	}
	if out.Explain["hybrid_skipped_lock"] != "count" {
		t.Fatalf("expected count lock, explain=%v", out.Explain)
	}
}

func TestRecallHybridDoesNotOverwriteMHListAnswer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content": `{"answer":"pottery, beach","supporting_memory_ids":[],"unresolved_targets":[],"abstain":false}`,
				}},
			},
		})
	}))
	defer server.Close()

	t.Setenv("BRAINY_RECALL_LLM", "1")
	store := newMemoryStoreStub()
	svc := NewService(store).WithHybridReader(HybridReaderConfig{
		BaseURL: server.URL, APIKey: "test", Model: "test-model",
	})
	now := svc.now()
	store.records["tim-jersey"] = MemoryRecord{
		MemoryID: "mem_tj", TenantID: "t-hyb-mh", SubjectID: "u1",
		Kind: KindFact, Content: "Tim owns a jersey",
		DedupeKey: "tj", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "jersey", "subject": "Tim"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "jersey", "subject": "Tim"},
	}
	store.records["john-jersey"] = MemoryRecord{
		MemoryID: "mem_jj", TenantID: "t-hyb-mh", SubjectID: "u1",
		Kind: KindFact, Content: "John owns a jersey",
		DedupeKey: "jj", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "jersey", "subject": "John"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "jersey", "subject": "John"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicatePossession, val: "jersey", memID: "mem_tj"},
		stubAtom{pred: PredicatePossession, val: "jersey", memID: "mem_jj"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-hyb-mh", SubjectID: "u1",
		Query: "What similar collectible do Tim and John own?", Mode: "enumerate", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "jersey") {
		t.Fatalf("MH list should stay locked, got %q explain=%v", out.Answer, out.Explain)
	}
	if strings.Contains(got, "pottery") || strings.Contains(got, "beach") {
		t.Fatalf("hybrid dump overwrote MH list: %q explain=%v", out.Answer, out.Explain)
	}
}

func TestRecallHybridEnumerateSkipsHopGrounding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content": `{"answer":"clarinet, violin","supporting_memory_ids":[],"unresolved_targets":[],"abstain":false}`,
				}},
			},
		})
	}))
	defer server.Close()

	t.Setenv("BRAINY_RECALL_LLM", "1")
	store := newMemoryStoreStub()
	svc := NewService(store).WithHybridReader(HybridReaderConfig{
		BaseURL: server.URL, APIKey: "test", Model: "test-model",
	})
	now := svc.now()
	facts := []struct {
		key, pred, val, content string
	}{
		{"clar", PredicateSkill, "clarinet", "Riley plays the clarinet."},
		{"viol", PredicateSkill, "violin", "Riley plays the violin."},
		{"pot", PredicateActivity, "pottery", "Riley also does pottery on weekends."},
		{"beach", PredicateActivity, "beach", "Riley likes the beach."},
	}
	for _, f := range facts {
		id := "mem_" + f.key
		store.records[f.key] = MemoryRecord{
			MemoryID: id, TenantID: "t-hyb-ground", SubjectID: "u1",
			Kind: KindFact, Content: f.content,
			DedupeKey: f.key, Status: StatusActive, UpdatedAt: now,
			Metadata: map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": "Riley"},
			Explain:  map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": "Riley"},
		}
		store.atoms = append(store.atoms, stubAtom{pred: f.pred, val: f.val, memID: id})
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-hyb-ground", SubjectID: "u1",
		Query: "What instruments does Riley play?", Mode: "enumerate", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "clarinet") || !strings.Contains(got, "violin") {
		t.Fatalf("hybrid instruments lost: %q explain=%v", out.Answer, out.Explain)
	}
	if strings.Contains(got, "pottery") || strings.Contains(got, "beach") {
		t.Fatalf("hop grounding re-expanded list: %q explain=%v", out.Answer, out.Explain)
	}
	if out.Explain["hybrid_grounded_to_hops"] == true {
		t.Fatalf("enumerated hybrid should not ground to hops: explain=%v", out.Explain)
	}
}

func TestRecallHybridDoesNotExpandShortTypedList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content": `{"answer":"clarinet, violin, pottery, beach","supporting_memory_ids":[],"unresolved_targets":[],"abstain":false}`,
				}},
			},
		})
	}))
	defer server.Close()

	t.Setenv("BRAINY_RECALL_LLM", "1")
	store := newMemoryStoreStub()
	svc := NewService(store).WithHybridReader(HybridReaderConfig{
		BaseURL: server.URL, APIKey: "test", Model: "test-model",
	})
	now := svc.now()
	facts := []struct {
		key, pred, val, content string
	}{
		{"clar", PredicateSkill, "plays clarinet", "Riley plays the clarinet."},
		{"viol", PredicateActivity, "playing violin", "Riley does daily violin practice after work."},
	}
	for _, f := range facts {
		id := "mem_" + f.key
		store.records[f.key] = MemoryRecord{
			MemoryID: id, TenantID: "t-hyb-short", SubjectID: "u1",
			Kind: KindFact, Content: f.content,
			DedupeKey: f.key, Status: StatusActive, UpdatedAt: now,
			Metadata: map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": "Riley"},
			Explain:  map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": "Riley"},
		}
		store.atoms = append(store.atoms, stubAtom{pred: f.pred, val: f.val, memID: id})
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-hyb-short", SubjectID: "u1",
		Query: "What instruments does Riley play?", Mode: "enumerate", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "clarinet") || !strings.Contains(got, "violin") {
		t.Fatalf("typed instruments lost: %q explain=%v", out.Answer, out.Explain)
	}
	if strings.Contains(got, "pottery") || strings.Contains(got, "beach") {
		t.Fatalf("hybrid expanded short list: %q explain=%v", out.Answer, out.Explain)
	}
	if out.Explain["hybrid_skipped_lock"] != "list" && out.Explain["hybrid_skipped_lock"] != "mh_list" {
		t.Fatalf("expected list lock, explain=%v answer=%q", out.Explain, out.Answer)
	}
}

func TestUncoveredHybridItemCount(t *testing.T) {
	typed := []RecallItem{{Value: "clarinet"}, {Value: "violin"}, {Value: "quiet weekend with kids"}}
	if uncoveredHybridItemCount(typed, "clarinet, violin, pottery, beach") < 2 {
		t.Fatal("pottery/beach should count as extras")
	}
	if uncoveredHybridItemCount(typed, "clarinet, violin") != 0 {
		t.Fatal("subset should have no extras")
	}
	if uncoveredHybridItemCount(nil, "pottery") != 1 {
		t.Fatal("empty typed cannot cover hybrid")
	}
}

func TestLockHybridListExtras(t *testing.T) {
	cases := []struct {
		name                    string
		enumerated              bool
		typedN, hybridN, extras int
		typedDump               bool
		want                    bool
	}{
		{name: "equal-length dump", enumerated: true, typedN: 6, hybridN: 6, extras: 4, want: true},
		{name: "shortened dump stays locked", enumerated: true, typedN: 8, hybridN: 6, extras: 6, want: true},
		{name: "short list expansion", enumerated: true, typedN: 2, hybridN: 4, extras: 2, want: true},
		{name: "1-item dump replacement", enumerated: true, typedN: 1, hybridN: 1, extras: 1, want: false},
		{name: "long dump to short hybrid", enumerated: true, typedN: 8, hybridN: 1, extras: 1, want: false},
		{name: "empty typed fills", enumerated: true, typedN: 0, hybridN: 3, extras: 3, want: false},
		{name: "shortened real list", enumerated: true, typedN: 6, hybridN: 4, extras: 0, want: true},
		{name: "dump shortened without extras", enumerated: true, typedN: 6, hybridN: 1, extras: 0, typedDump: true, want: false},
		{name: "not enumerated", enumerated: false, typedN: 6, hybridN: 6, extras: 4, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := lockHybridListExtras(tc.enumerated, tc.typedN, tc.hybridN, tc.extras, tc.typedDump)
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestShouldGroundHybridToHopsSkipsIdentityDumps(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Maria", EntityID: "e-maria", Value: "Maria", Source: "typed_store"},
		{Kind: "follow_relation", Entity: "Maria", EntityID: "e-maria", Predicate: PredicateIdentity, Source: "typed_store",
			Value: "inspiration", Values: []string{"inspiration", "family", "team"}},
	}
	pkt := EvidencePacket{
		Contents:        []string{"Maria has a dog named Shadow.", "Maria has a new puppy."},
		ContextEvidence: []PacketItem{{Content: "Maria has a dog named Shadow."}},
		Coverage:        map[string]any{"hop_results": hops},
	}
	q := "What is the name of Maria's second puppy?"
	if !hopJoinProven(hops) {
		t.Fatalf("fixture must be hop-join proven")
	}
	if shouldGroundHybridToHops(q, hops, pkt, false) {
		t.Fatal("must not hop-ground identity dumps over covering memories")
	}
	if shouldGroundHybridToHops(q, hops, pkt, true) {
		t.Fatal("enumerated answers already skip hop-ground")
	}
}

func TestShouldGroundHybridToHopsSkipsUncoveredTravelDumps(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Tim", EntityID: "e-tim", Value: "Tim", Source: "typed_store"},
		{Kind: "follow_relation", Entity: "Tim", EntityID: "e-tim", Predicate: PredicateEvent, Source: "typed_store",
			Value: "visa requirements", Values: []string{"visa requirements", "explore nature", "traveling"}},
	}
	pkt := EvidencePacket{
		Contents: []string{"Tim visits the United Kingdom most often."},
		Coverage: map[string]any{"hop_results": hops},
	}
	q := "which country has Tim visited most frequently in his travels?"
	if !hopJoinProven(hops) {
		t.Fatalf("fixture must be hop-join proven")
	}
	if hopsAreIdentityOnly(hops) {
		t.Fatal("fixture must not be identity-only; that path is covered elsewhere")
	}
	if shouldGroundHybridToHops(q, hops, pkt, false) {
		t.Fatal("must not hop-ground travel dumps that miss leftover query tokens")
	}
}

func TestListItemCount(t *testing.T) {
	if listItemCount("clarinet, violin", nil) != 2 {
		t.Fatal("comma split")
	}
	if listItemCount("", []RecallItem{{Value: "a"}, {Value: "b"}, {Value: "c"}}) != 3 {
		t.Fatal("items win")
	}
	if listItemCount("not in memory", nil) != 0 {
		t.Fatal("sentinel")
	}
}

func TestRecallHybridKeepsFillingWhenHopsUnproven(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content": `{"answer":"strawberry","supporting_memory_ids":[],"unresolved_targets":[],"abstain":false}`,
				}},
			},
		})
	}))
	defer server.Close()

	t.Setenv("BRAINY_RECALL_LLM", "1")
	store := newMemoryStoreStub()
	svc := NewService(store).WithHybridReader(HybridReaderConfig{
		BaseURL: server.URL, APIKey: "test", Model: "test-model",
	})
	now := svc.now()
	facts := []struct {
		key, content string
	}{
		{"fill", "Alex made a dairy-free vanilla cake with strawberry filling recently."},
		{"pot", "Alex does pottery on weekends."},
		{"beach", "Jordan likes the beach."},
		{"camp", "Alex and Jordan went camping together."},
	}
	for _, f := range facts {
		store.records[f.key] = MemoryRecord{
			MemoryID: "mem_" + f.key, TenantID: "t-hyb-fill", SubjectID: "u1",
			Kind: KindFact, Content: f.content,
			DedupeKey: f.key, Status: StatusActive, UpdatedAt: now,
		}
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-hyb-fill", SubjectID: "u1",
		Query: "What filling did Alex use in the cake she made for Jordan?", Mode: "answer", TopK: 8,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "strawberry") {
		t.Fatalf("expected strawberry filling, got %q explain=%v", out.Answer, out.Explain)
	}
	if strings.Contains(got, "pottery") || strings.Contains(got, "beach") || strings.Contains(got, "camping") {
		t.Fatalf("unproven hops replaced hybrid filling: %q explain=%v", out.Answer, out.Explain)
	}
	if out.Explain["hybrid_grounded_to_hops"] == true {
		t.Fatalf("unproven hops must not ground hybrid: explain=%v", out.Explain)
	}
}

func TestRecallHybridUnlocksMHListWhenSkipSlots(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content": `{"answer":"at a music festival","supporting_memory_ids":[],"unresolved_targets":[],"abstain":false}`,
				}},
			},
		})
	}))
	defer server.Close()

	t.Setenv("BRAINY_RECALL_LLM", "1")
	store := newMemoryStoreStub()
	svc := NewService(store).WithHybridReader(HybridReaderConfig{
		BaseURL: server.URL, APIKey: "test", Model: "test-model",
	})
	now := svc.now()
	facts := []struct {
		key, sub, pred, val, content string
	}{
		{"c-rock", "Calvin", PredicateActivity, "rocks", "Calvin has done crashing at Rocks"},
		{"d-rock", "Dave", PredicateActivity, "rocks", "Dave has done crashing at Rocks"},
		{"fest", "Calvin", "", "", "I had the opportunity to meet Frank Ocean at a music festival in Tokyo and we clicked"},
	}
	for _, f := range facts {
		rec := MemoryRecord{
			MemoryID: "mem_" + f.key, TenantID: "t-hyb-fest", SubjectID: "u1",
			Kind: KindFact, Content: f.content,
			DedupeKey: f.key, Status: StatusActive, UpdatedAt: now,
		}
		if f.pred != "" {
			rec.Metadata = map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": f.sub}
			rec.Explain = map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": f.sub}
			store.atoms = append(store.atoms, stubAtom{pred: f.pred, val: f.val, memID: rec.MemoryID})
		}
		store.records[f.key] = rec
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-hyb-fest", SubjectID: "u1",
		Query: "Where did Calvin and Dave meet Frank Ocean to start collaborating?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "festival") {
		t.Fatalf("skip-unrelated mh_list must not lock Rocks over hybrid festival: %q explain=%v", out.Answer, out.Explain)
	}
	if strings.EqualFold(strings.TrimSpace(out.Answer), "Rocks") {
		t.Fatalf("short unrelated slot leaked: %q", out.Answer)
	}
	if out.Explain["hybrid_skipped_lock"] == "mh_list" {
		t.Fatalf("mh_list must unlock when hop slots skip leftover tokens: explain=%v", out.Explain)
	}
}
