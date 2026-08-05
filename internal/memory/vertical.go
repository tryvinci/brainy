package memory

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	// Always honor explicit request label (event match / supersede v2).
	if label := strings.TrimSpace(req.Label); label != "" {
		record.Label = label
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
	if err := validatePackStateMachine(p, req, ""); err != nil {
		return err
	}
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

// validatePackStateTransition looks up prior ticket_state for the same
// ticket_id and enforces the pack sidecar FSM when present.
func (s *Service) validatePackStateTransition(ctx context.Context, req IngestRequest) error {
	if s.packs == nil {
		return nil
	}
	vertical := strings.TrimSpace(req.Vertical)
	if vertical == "" || vertical == VerticalCore {
		return nil
	}
	p, ok := s.packs.Get(vertical)
	if !ok {
		return nil
	}
	label := strings.TrimSpace(req.Label)
	if p.MachineForLabel(label) == "" {
		return nil
	}
	prior := s.lookupPriorPackState(ctx, req)
	return validatePackStateMachine(p, req, prior)
}

func (s *Service) lookupPriorPackState(ctx context.Context, req IngestRequest) string {
	if s.store == nil {
		return ""
	}
	ticketID := metadataString(req.Metadata, "ticket_id")
	campaignName := metadataString(req.Metadata, "name")
	campaignID := metadataString(req.Metadata, "campaign_id")
	if ticketID == "" && campaignName == "" && campaignID == "" {
		return ""
	}
	listed, err := s.listSubjectCorpus(ctx, req.TenantID, req.SubjectID, true, 200)
	if err != nil {
		return ""
	}
	field := "status"
	if s.packs != nil {
		if p, ok := s.packs.Get(strings.TrimSpace(req.Vertical)); ok {
			field = p.StatusFieldForLabel(req.Label)
		}
	}
	var best string
	var bestTime time.Time
	for _, m := range listed {
		if strings.TrimSpace(m.Label) != strings.TrimSpace(req.Label) {
			continue
		}
		match := false
		if ticketID != "" && metadataString(m.Metadata, "ticket_id") == ticketID {
			match = true
		}
		if campaignName != "" && metadataString(m.Metadata, "name") == campaignName {
			match = true
		}
		if campaignID != "" && (metadataString(m.Metadata, "campaign_id") == campaignID || metadataString(m.Metadata, "name") == campaignID) {
			match = true
		}
		if !match {
			continue
		}
		status := metadataString(m.Metadata, field)
		if status == "" {
			status = metadataString(m.Metadata, "value_norm")
		}
		if status == "" {
			continue
		}
		ts := m.UpdatedAt
		if ts.IsZero() {
			ts = m.CreatedAt
		}
		if best == "" || ts.After(bestTime) {
			best = status
			bestTime = ts
		}
	}
	return best
}
