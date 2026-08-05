package pack

import (
	"path/filepath"
	"testing"
)

func TestSupportV2SidecarsLoaded(t *testing.T) {
	root := filepath.Join("..", "..", "packs")
	reg, err := LoadRegistryFromDir(root)
	if err != nil {
		t.Fatalf("load registry: %v", err)
	}
	p, ok := reg.Get("support")
	if !ok {
		t.Fatal("support pack not registered")
	}
	if p.Version != "2" {
		t.Fatalf("expected support v2 preferred, got version %q", p.Version)
	}
	if len(p.Entities) < 5 {
		t.Fatalf("expected entities sidecar, got %d", len(p.Entities))
	}
	sm, ok := p.StateMachines["ticket_status"]
	if !ok || len(sm.States) < 5 {
		t.Fatalf("expected ticket_status FSM, got %+v", p.StateMachines)
	}
	if err := p.ValidateStateTransition("ticket_status", "resolved", "reopened"); err != nil {
		t.Fatalf("resolved→reopened should be allowed: %v", err)
	}
	if err := p.ValidateStateTransition("ticket_status", "closed", "pending"); err == nil {
		t.Fatal("closed→pending should be rejected")
	}
	if err := p.ValidateStateTransition("ticket_status", "", "open"); err != nil {
		t.Fatalf("bootstrap open should be allowed: %v", err)
	}
}

func TestMarketingV2StateMachinesLoaded(t *testing.T) {
	root := filepath.Join("..", "..", "packs")
	reg, err := LoadRegistryFromDir(root)
	if err != nil {
		t.Fatalf("load registry: %v", err)
	}
	p, ok := reg.Get("marketing")
	if !ok {
		t.Fatal("marketing pack not registered")
	}
	if len(p.Entities) < 5 {
		t.Fatalf("expected marketing entities, got %d", len(p.Entities))
	}
	if p.MachineForLabel("campaign") != "campaign_status" {
		t.Fatalf("campaign machine=%q", p.MachineForLabel("campaign"))
	}
	if err := p.ValidateStateTransition("campaign_status", "active", "completed"); err != nil {
		t.Fatalf("active→completed: %v", err)
	}
	if err := p.ValidateStateTransition("campaign_status", "archived", "active"); err == nil {
		t.Fatal("archived→active should be rejected")
	}
	if err := p.ValidateStateTransition("creative_approval", "draft", "approved"); err != nil {
		t.Fatalf("draft→approved: %v", err)
	}
}
