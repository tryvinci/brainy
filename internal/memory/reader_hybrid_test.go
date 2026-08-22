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
		QueryPlan{NeedsMultiHop: true},
		EvidencePacket{Contents: []string{"a"}, Items: []PacketItem{{Content: "a"}}},
	)
	if !ok || reason != "composition_needed" {
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

func TestRecallHybridDoesNotOverwriteDateAnswer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content": `{"answer":"yesterday","supporting_memory_ids":[],"unresolved_targets":[],"abstain":false}`,
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
	if strings.Contains(strings.ToLower(out.Answer), "yesterday") {
		t.Fatalf("hybrid overwrote date_answer: %q explain=%v", out.Answer, out.Explain)
	}
	if out.Explain["date_answer"] == true && out.Explain["hybrid_skipped_lock"] != "date" && out.Explain["reader_source"] == "hybrid_llm_packet" {
		t.Fatalf("date lock should skip hybrid apply, explain=%v", out.Explain)
	}
}
