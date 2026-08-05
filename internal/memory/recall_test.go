package memory

import (
	"context"
	"strings"
	"testing"
)

func TestRecallEnumerateDistinctValues(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Dana: I've been hiking every weekend"},
			{Role: "user", Content: "Dana: I'm a big fan of ceramics"},
			{Role: "user", Content: "Dana: I've been swimming at the lake"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t1", SubjectID: "u1", Query: "What activities does Dana enjoy?", Mode: "enumerate", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Items) < 2 {
		t.Fatalf("expected enumerated items, got %#v context=%q", out.Items, out.ContextBlock)
	}
	joined := strings.ToLower(out.ContextBlock + " " + out.Answer)
	for _, need := range []string{"hik", "ceram", "swim"} {
		if !strings.Contains(joined, need) && !itemHas(out.Items, need) {
			t.Fatalf("expected %q in enumerate output %#v", need, out.Items)
		}
	}
}

func TestRecallUsesEvidencePacketAndModeHint(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-pkt", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Alex currently works as a nurse in Seattle."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-pkt", SubjectID: "u1", Query: "what does Alex currently do?", Mode: "answer", TopK: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Explain["reader_source"] != "evidence_packet" {
		t.Fatalf("expected evidence_packet reader, explain=%v", out.Explain)
	}
	tools, _ := out.Explain["tools_executed"].([]string)
	if len(tools) == 0 || tools[0] != "search" {
		t.Fatalf("tools_executed=%v", tools)
	}
	if out.Abstained && out.ContextBlock == "" && out.Answer == "not in memory" {
		// Empty store path is ok only if search truly missed.
		t.Fatalf("unexpected total abstain with ingest: answer=%q context=%q", out.Answer, out.ContextBlock)
	}
}

func TestRecallAbstainsWhenPacketEmpty(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-empty", SubjectID: "u1", Query: "what is the secret code?", Mode: "answer", TopK: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Abstained {
		t.Fatalf("expected abstain on empty evidence, got %#v", out)
	}
}

func itemHas(items []RecallItem, needle string) bool {
	for _, it := range items {
		if strings.Contains(strings.ToLower(it.Value), needle) {
			return true
		}
	}
	return false
}
