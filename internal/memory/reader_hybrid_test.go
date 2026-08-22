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
		want                    bool
	}{
		{name: "equal-length dump", enumerated: true, typedN: 6, hybridN: 6, extras: 4, want: true},
		{name: "shortened dump stays locked", enumerated: true, typedN: 8, hybridN: 6, extras: 6, want: true},
		{name: "short list expansion", enumerated: true, typedN: 2, hybridN: 4, extras: 2, want: true},
		{name: "1-item dump replacement", enumerated: true, typedN: 1, hybridN: 1, extras: 1, want: false},
		{name: "long dump to short hybrid", enumerated: true, typedN: 8, hybridN: 1, extras: 1, want: false},
		{name: "empty typed fills", enumerated: true, typedN: 0, hybridN: 3, extras: 3, want: false},
		{name: "no extras", enumerated: true, typedN: 6, hybridN: 4, extras: 0, want: false},
		{name: "not enumerated", enumerated: false, typedN: 6, hybridN: 6, extras: 4, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := lockHybridListExtras(tc.enumerated, tc.typedN, tc.hybridN, tc.extras)
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
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
