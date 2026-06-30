package pack

import (
	"path/filepath"
	"testing"
)

func TestLifecycleRuleMatchesMetadata(t *testing.T) {
	rule := LifecycleRule{
		Match: map[string]string{
			"label":           "campaign",
			"metadata.status": "archived",
		},
		Set: LifecycleEffect{
			LifecycleState:    "archived",
			ExcludeFromSearch: true,
		},
	}
	if !rule.Matches("campaign", map[string]any{"status": "archived"}) {
		t.Fatal("expected archived campaign rule to match")
	}
	if rule.Matches("campaign", map[string]any{"status": "active"}) {
		t.Fatal("expected active campaign to miss archived rule")
	}
}

func TestMarketingPackLifecycleRules(t *testing.T) {
	root := filepath.Join("..", "..", "packs")
	reg, err := LoadRegistryFromDir(root)
	if err != nil {
		t.Fatalf("load registry: %v", err)
	}
	p, ok := reg.Get("marketing")
	if !ok {
		t.Fatal("marketing pack not registered")
	}
	if len(p.LifecycleRules) < 4 {
		t.Fatalf("expected lifecycle rules in marketing pack, got %d", len(p.LifecycleRules))
	}

	archived := p.LifecycleEffectFor("campaign", map[string]any{"status": "archived"})
	if archived == nil || !archived.ExcludeFromSearch || archived.LifecycleState != "archived" {
		t.Fatalf("archived campaign effect = %+v", archived)
	}

	active := p.LifecycleEffectFor("campaign", map[string]any{"status": "active"})
	if active == nil || active.RankMultiplier != 1.5 || active.LifecycleState != "active" {
		t.Fatalf("active campaign effect = %+v", active)
	}

	rejected := p.LifecycleEffectFor("creative_asset", map[string]any{"approval_status": "rejected"})
	if rejected == nil || !rejected.ExcludeFromSearch || rejected.LifecycleState != "suppressed" {
		t.Fatalf("rejected creative asset effect = %+v", rejected)
	}
}
