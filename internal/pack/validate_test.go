package pack

import (
	"path/filepath"
	"testing"
)

func TestValidateMarketingCampaignMetadata(t *testing.T) {
	reg, err := LoadRegistryFromDir(filepath.Join("..", "..", "packs"))
	if err != nil {
		t.Fatalf("load registry: %v", err)
	}
	p, ok := reg.Get("marketing")
	if !ok {
		t.Fatal("marketing pack missing")
	}

	if err := p.ValidateMetadata("campaign", map[string]any{
		"name":   "Summer",
		"status": "active",
	}); err != nil {
		t.Fatalf("valid campaign metadata rejected: %v", err)
	}

	if err := p.ValidateMetadata("campaign", map[string]any{
		"name": "Summer",
	}); err == nil {
		t.Fatal("expected missing status to fail validation")
	}

	if err := p.ValidateMetadata("campaign", map[string]any{
		"name":   "Summer",
		"status": "invalid",
	}); err == nil {
		t.Fatal("expected invalid status enum to fail validation")
	}
}
