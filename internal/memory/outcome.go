package memory

import (
	"fmt"
	"math"
	"strings"

	"brainy/internal/pack"
)

func validatePackMetadata(reg *pack.Registry, req IngestRequest) error {
	if reg == nil {
		return nil
	}
	vertical := strings.TrimSpace(req.Vertical)
	if vertical == "" || vertical == VerticalCore {
		return nil
	}
	p, ok := reg.Get(vertical)
	if !ok {
		return nil
	}
	label := strings.TrimSpace(req.Label)
	if label == "" {
		return nil
	}
	if err := p.ValidateMetadata(label, req.Metadata); err != nil {
		return fmt.Errorf("pack metadata validation failed: %w", err)
	}
	if err := validatePackStateMachine(p, req, ""); err != nil {
		return err
	}
	return nil
}

func validatePackStateMachine(p *pack.Pack, req IngestRequest, priorStatus string) error {
	if p == nil {
		return nil
	}
	label := strings.TrimSpace(req.Label)
	machine := p.MachineForLabel(label)
	if machine == "" {
		return nil
	}
	status := metadataString(req.Metadata, "status")
	if status == "" {
		status = metadataString(req.Metadata, "value_norm")
	}
	if status == "" {
		return nil
	}
	from := metadataString(req.Metadata, "from_status")
	if from == "" {
		from = metadataString(req.Metadata, "previous_status")
	}
	if from == "" {
		from = priorStatus
	}
	if err := p.ValidateStateTransition(machine, from, status); err != nil {
		return fmt.Errorf("pack state machine validation failed: %w", err)
	}
	return nil
}

func metadataString(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	raw, ok := metadata[key]
	if !ok {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func synthesizeBeliefFromOutcome(req IngestRequest) (MemoryRecord, bool) {
	if strings.TrimSpace(req.Label) != "performance_outcome" {
		return MemoryRecord{}, false
	}
	content := outcomeBeliefContent(req)
	if content == "" {
		return MemoryRecord{}, false
	}

	conviction := outcomeConviction(req.Metadata)
	metadata := map[string]any{
		"conviction": conviction,
		"source":     "performance_outcome",
	}
	if variant, ok := req.Metadata["variant"].(string); ok && variant != "" {
		metadata["variant"] = variant
	}

	return MemoryRecord{
		Kind:       KindFact,
		Content:    content,
		SourceText: firstMessageContent(req),
		SourceType: req.SourceType,
		Label:      "content_belief",
		Primitive:  PrimitiveBelief,
		Metadata:   metadata,
	}, true
}

func outcomeBeliefContent(req IngestRequest) string {
	if text := firstMessageContent(req); text != "" {
		return titleSentence(NormalizeText(text))
	}
	metric, _ := req.Metadata["metric"].(string)
	variant, _ := req.Metadata["variant"].(string)
	if metric != "" && variant != "" {
		return fmt.Sprintf("%s variant %s performed best", titleSentence(metric), variant)
	}
	return ""
}

func outcomeConviction(metadata map[string]any) float64 {
	if metadata == nil {
		return 0.5
	}
	if raw, ok := metadata["conviction"].(float64); ok && raw > 0 {
		return math.Min(raw, 1)
	}
	switch value := metadata["value"].(type) {
	case float64:
		if value > 1 {
			return math.Min(value/100, 1)
		}
		return math.Min(math.Max(value, 0.1), 1)
	case int:
		if value > 1 {
			return math.Min(float64(value)/100, 1)
		}
		return math.Min(math.Max(float64(value), 0.1), 1)
	}
	return 0.75
}

func firstMessageContent(req IngestRequest) string {
	for _, message := range req.Messages {
		if content := NormalizeText(message.Content); content != "" {
			return content
		}
	}
	return ""
}

func applyScopeBoost(score *float64, explain map[string]any, record MemoryRecord, scope string, reg *pack.Registry) {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		return
	}
	if record.Scope == scope {
		mult := 1.5
		if reg != nil {
			if p, ok := reg.Get(record.Vertical); ok {
				if boost, ok := p.RankPolicy.ScopeBoost["active_scope"]; ok && boost > 0 {
					mult = boost
				}
			}
		}
		*score *= mult
		explain["scope_boost"] = mult
		explain["scope"] = scope
	}
}

func applyConvictionBoost(score *float64, explain map[string]any, record MemoryRecord) {
	if record.Primitive != PrimitiveBelief || record.Metadata == nil {
		return
	}
	raw, ok := record.Metadata["conviction"]
	if !ok {
		return
	}
	conviction, ok := toFloat(raw)
	if !ok || conviction <= 0 {
		return
	}
	bonus := conviction * 2
	*score += bonus
	explain["conviction"] = conviction
	explain["conviction_bonus"] = bonus
}

func applyTasteSignalBoost(score *float64, explain map[string]any, record MemoryRecord, queryTokens []string) {
	if record.Primitive != PrimitiveTasteSignal || record.Metadata == nil {
		return
	}
	tags, ok := record.Metadata["style_tags"].([]any)
	if !ok {
		return
	}
	matched := make([]string, 0)
	for _, tag := range tags {
		tagStr, ok := tag.(string)
		if !ok {
			continue
		}
		tagToken := strings.ToLower(tagStr)
		for _, queryToken := range queryTokens {
			if queryToken == tagToken || strings.Contains(tagToken, queryToken) || strings.Contains(queryToken, tagToken) {
				matched = append(matched, tagStr)
				break
			}
		}
	}
	if len(matched) == 0 {
		return
	}
	bonus := float64(len(matched)) * 0.5
	*score += bonus
	explain["taste_signal_matched"] = matched
	explain["taste_signal_bonus"] = bonus
}

func toFloat(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}
