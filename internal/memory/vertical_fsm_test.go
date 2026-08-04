package memory

import (
	"context"
	"path/filepath"
	"testing"

	"brainy/internal/pack"
)

func TestSupportTicketStateMachineOnIngest(t *testing.T) {
	reg, err := pack.LoadRegistryFromDir(filepath.Join("..", "..", "packs"))
	if err != nil {
		t.Fatalf("load packs: %v", err)
	}
	store := newMemoryStoreStub()
	svc := NewServiceWithPacks(store, reg)

	base := IngestRequest{
		TenantID:   "t-fsm",
		SubjectID:  "cust-1",
		Vertical:   "support",
		Label:      "ticket_state",
		SourceType: "crm",
		Messages:   []Message{{Role: "user", Content: "Ticket T-9 is resolved"}},
		Metadata:   map[string]any{"status": "resolved", "ticket_id": "T-9"},
	}
	if _, err := svc.Ingest(context.Background(), base); err != nil {
		t.Fatalf("bootstrap resolved: %v", err)
	}

	reopen := base
	reopen.Messages = []Message{{Role: "user", Content: "Ticket T-9 was reopened"}}
	reopen.Metadata = map[string]any{"status": "reopened", "ticket_id": "T-9"}
	if _, err := svc.Ingest(context.Background(), reopen); err != nil {
		t.Fatalf("resolved→reopened: %v", err)
	}

	bad := base
	bad.Messages = []Message{{Role: "user", Content: "illegal jump"}}
	bad.Metadata = map[string]any{"status": "pending", "ticket_id": "T-9", "from_status": "closed"}
	if _, err := svc.Ingest(context.Background(), bad); err == nil {
		t.Fatal("expected closed→pending to fail")
	}
}
