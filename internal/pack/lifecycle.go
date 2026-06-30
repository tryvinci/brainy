package pack

import (
	"fmt"
	"strings"
)

type LifecycleRule struct {
	Match map[string]string `yaml:"match"`
	Set   LifecycleEffect   `yaml:"set"`
}

type LifecycleEffect struct {
	LifecycleState    string  `yaml:"lifecycle_state"`
	ExcludeFromSearch bool    `yaml:"exclude_from_search"`
	RankMultiplier    float64 `yaml:"rank_multiplier"`
}

func (r LifecycleRule) Matches(label string, metadata map[string]any) bool {
	if len(r.Match) == 0 {
		return false
	}
	for key, want := range r.Match {
		got, ok := matchValue(key, label, metadata)
		if !ok || got != want {
			return false
		}
	}
	return true
}

func matchValue(key, label string, metadata map[string]any) (string, bool) {
	switch key {
	case "label":
		return label, true
	default:
		if path, ok := strings.CutPrefix(key, "metadata."); ok {
			return metadataString(metadata, path), true
		}
		return "", false
	}
}

func metadataString(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	value, ok := metadata[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}

func (p *Pack) LifecycleEffectFor(label string, metadata map[string]any) *LifecycleEffect {
	if p == nil {
		return nil
	}
	for _, rule := range p.LifecycleRules {
		if rule.Matches(label, metadata) {
			effect := rule.Set
			return &effect
		}
	}
	return nil
}
