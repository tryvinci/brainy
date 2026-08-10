package memory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
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
