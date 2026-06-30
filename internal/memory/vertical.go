package memory

import (
	"fmt"
	"strings"

	"brainy/internal/pack"
)

func ApplyVerticalPack(record *MemoryRecord, req IngestRequest, kind, content string, reg *pack.Registry) error {
	if record.Metadata == nil {
		record.Metadata = map[string]any{}
	}
	for key, value := range req.Metadata {
		record.Metadata[key] = value
	}

	vertical := strings.TrimSpace(req.Vertical)
	if vertical == "" {
		vertical = VerticalCore
	}
	record.Vertical = vertical
	record.LifecycleState = LifecycleActive
	if scope := strings.TrimSpace(req.Scope); scope != "" {
		record.Scope = scope
	}

	if reg == nil {
		return nil
	}
	p, ok := reg.Get(vertical)
	if !ok {
		return nil
	}

	label := strings.TrimSpace(req.Label)
	if label != "" {
		record.Label = label
		if entry, exists := p.Vocabulary[label]; exists {
			record.Primitive = pack.NormalizePrimitive(entry.Primitive)
		}
		if err := p.ValidateMetadata(label, record.Metadata); err != nil {
			return fmt.Errorf("pack metadata validation failed: %w", err)
		}
	} else {
		packLabel, primitive, labelOK := p.LabelForKind(kind)
		if labelOK {
			record.Label = packLabel
			record.Primitive = pack.NormalizePrimitive(primitive)
		}
	}

	if vertical == "marketing" && strings.Contains(strings.ToLower(content), "never ") {
		record.Label = "brand_rule"
		record.Primitive = PrimitivePrinciple
	}

	applyPackLifecycle(record, p)
	return nil
}

func applyPackLifecycle(record *MemoryRecord, p *pack.Pack) {
	if p == nil || record == nil {
		return
	}
	effect := p.LifecycleEffectFor(record.Label, record.Metadata)
	if effect == nil {
		return
	}
	if effect.LifecycleState != "" {
		record.LifecycleState = effect.LifecycleState
	}
}

func LifecycleRankMultiplier(reg *pack.Registry, record MemoryRecord) float64 {
	if reg == nil {
		return 1
	}
	p, ok := reg.Get(record.Vertical)
	if !ok {
		return 1
	}
	effect := p.LifecycleEffectFor(record.Label, record.Metadata)
	if effect == nil || effect.RankMultiplier <= 0 {
		return 1
	}
	return effect.RankMultiplier
}

func IsLifecycleSearchVisible(state string) bool {
	switch state {
	case LifecycleArchived, LifecycleSuperseded, LifecycleSuppressed:
		return false
	default:
		return true
	}
}
