package memory

import (
	"strings"

	"brainy/internal/pack"
)

func ApplyVerticalPack(record *MemoryRecord, vertical, kind, content string, reg *pack.Registry) {
	if record.Metadata == nil {
		record.Metadata = map[string]any{}
	}
	vertical = strings.TrimSpace(vertical)
	if vertical == "" {
		vertical = VerticalCore
	}
	record.Vertical = vertical
	record.LifecycleState = LifecycleActive

	if reg == nil {
		return
	}
	p, ok := reg.Get(vertical)
	if !ok {
		return
	}

	label, primitive, ok := p.LabelForKind(kind)
	if ok {
		record.Label = label
		record.Primitive = pack.NormalizePrimitive(primitive)
	}

	if vertical == "marketing" && strings.Contains(strings.ToLower(content), "never ") {
		record.Label = "brand_rule"
		record.Primitive = PrimitivePrinciple
	}
}
