package memory

import "strings"

// QueryPlan is the deterministic typed plan for /recall (program Phase 4).
// It does not rewrite SearchOpt; it records what synthesis should do with
// retrieved evidence and temporal reads.
type QueryPlan struct {
	Intents           []string `json:"intents"`
	PrimaryIntent     string   `json:"primary_intent"`
	NeedsTemporal     bool     `json:"needs_temporal"`
	NeedsEnumeration  bool     `json:"needs_enumeration"`
	NeedsMultiHop     bool     `json:"needs_multi_hop"`
	NeedsAbstention   bool     `json:"needs_abstention"`
	CoverageTargets   []string `json:"coverage_targets,omitempty"`
	Tools             []string `json:"tools,omitempty"`
	BudgetPasses      int      `json:"budget_passes"`
	PreferredModeHint string   `json:"preferred_mode_hint,omitempty"`
}

// EvidencePacket is the bounded evidence set handed to synthesis/oracles.
type EvidencePacket struct {
	MemoryIDs      []string       `json:"memory_ids"`
	Contents       []string       `json:"contents,omitempty"`
	Items          []PacketItem   `json:"items,omitempty"`
	Predicates     []string       `json:"predicates,omitempty"`
	TemporalAnswer string         `json:"temporal_answer,omitempty"`
	Coverage       map[string]any `json:"coverage,omitempty"`
	Plan           QueryPlan      `json:"plan"`
}

// PlanQuery builds a typed plan from the deterministic intent classifier.
func PlanQuery(query string, intents []string) QueryPlan {
	if len(intents) == 0 {
		intents = AnalyzeQueryIntents(query)
	}
	plan := QueryPlan{
		Intents:       append([]string(nil), intents...),
		BudgetPasses:  1,
		CoverageTargets: nil,
	}
	if len(intents) > 0 {
		plan.PrimaryIntent = intents[0]
	}
	for _, intent := range intents {
		switch intent {
		case IntentCurrentState, IntentHistoricalState, IntentTemporalSequence:
			plan.NeedsTemporal = true
		case IntentEnumeration, IntentAggregation:
			plan.NeedsEnumeration = true
		case IntentMultiHop:
			plan.NeedsMultiHop = true
			plan.BudgetPasses = 2
		case IntentAbstentionSens:
			plan.NeedsAbstention = true
		}
	}
	lower := strings.ToLower(query)
	if strings.Contains(lower, "not sure") || strings.Contains(lower, "unknown") ||
		strings.Contains(lower, "do you know") {
		plan.NeedsAbstention = true
	}

	plan.Tools = planTools(plan)
	plan.CoverageTargets = planCoverageTargets(query, plan)
	plan.PreferredModeHint = preferredModeHint(plan)
	return plan
}

func planTools(plan QueryPlan) []string {
	tools := []string{"search"}
	add := func(t string) {
		for _, e := range tools {
			if e == t {
				return
			}
		}
		tools = append(tools, t)
	}
	if plan.NeedsTemporal {
		add("temporal_resolve")
	}
	if plan.NeedsEnumeration {
		add("enumerate")
	}
	if plan.NeedsMultiHop {
		add("evidence_set")
	}
	if plan.NeedsAbstention {
		add("abstain")
	}
	return tools
}

func planCoverageTargets(query string, plan QueryPlan) []string {
	if plan.NeedsMultiHop {
		return multiHopTargets(query)
	}
	if plan.NeedsEnumeration {
		toks := contentBearingTokens(tokenize(query))
		if len(toks) == 0 {
			return []string{"set_complete"}
		}
		out := make([]string, 0, len(toks))
		for _, t := range toks {
			if len(t) >= 3 {
				out = append(out, t)
			}
		}
		if len(out) == 0 {
			return []string{"set_complete"}
		}
		return out
	}
	if plan.NeedsTemporal {
		return predicateHintsFromQuery(query)
	}
	return []string{"point_fact"}
}

func preferredModeHint(plan QueryPlan) string {
	switch {
	case plan.NeedsEnumeration:
		return "enumerate"
	case plan.NeedsTemporal && plan.PrimaryIntent == IntentCurrentState:
		return "answer"
	case plan.NeedsMultiHop:
		return "answer"
	default:
		return "context"
	}
}

// BuildEvidencePacket projects search hits + temporal explain into a packet.
func BuildEvidencePacket(plan QueryPlan, results []SearchResult, explain map[string]any) EvidencePacket {
	pkt := EvidencePacket{
		Plan:      plan,
		MemoryIDs: make([]string, 0, len(results)),
		Contents:  make([]string, 0, len(results)),
	}
	for _, r := range results {
		if r.MemoryID != "" {
			pkt.MemoryIDs = append(pkt.MemoryIDs, r.MemoryID)
		}
		if c := strings.TrimSpace(r.Content); c != "" && !strings.HasSuffix(c, "?") {
			pkt.Contents = append(pkt.Contents, c)
		}
	}
	if explain != nil {
		if ta, _ := explain["temporal_answer"].(string); ta != "" {
			pkt.TemporalAnswer = ta
		}
		switch preds := explain["temporal"].(type) {
		case []map[string]any:
			for _, p := range preds {
				if pred, _ := p["predicate"].(string); pred != "" {
					pkt.Predicates = append(pkt.Predicates, pred)
				}
			}
		case []any:
			for _, raw := range preds {
				p, _ := raw.(map[string]any)
				if pred, _ := p["predicate"].(string); pred != "" {
					pkt.Predicates = append(pkt.Predicates, pred)
				}
			}
		}
	}
	satisfied := packetCoverageSatisfied(plan, pkt)
	pkt.Coverage = map[string]any{
		"targets":   len(plan.CoverageTargets),
		"satisfied": satisfied,
		"hit_count": len(pkt.MemoryIDs),
	}
	return pkt
}

func packetCoverageSatisfied(plan QueryPlan, pkt EvidencePacket) bool {
	if len(pkt.Contents) == 0 && pkt.TemporalAnswer == "" {
		return false
	}
	if plan.NeedsTemporal && pkt.TemporalAnswer != "" && !plan.NeedsMultiHop {
		return true
	}
	if plan.NeedsMultiHop {
		// Require a bridge+direct pair (or temporal+content), not merely any two strings.
		if len(pkt.Contents) >= 1 && pkt.TemporalAnswer != "" {
			return true
		}
		bridges := countRole(pkt.Items, "bridge")
		directs := countRole(pkt.Items, "direct")
		if bridges >= 1 && directs >= 1 {
			return true
		}
		if len(pkt.Contents) < 2 {
			return false
		}
		return packetLooksLinked(pkt.Contents[0], pkt.Contents[1])
	}
	if plan.NeedsEnumeration {
		return len(pkt.Contents) >= 1
	}
	return len(pkt.MemoryIDs) > 0 || pkt.TemporalAnswer != ""
}

func packetLooksLinked(a, b string) bool {
	ta := contentBearingTokens(tokenize(a))
	tb := map[string]struct{}{}
	for _, t := range contentBearingTokens(tokenize(b)) {
		tb[t] = struct{}{}
	}
	overlap := 0
	for _, t := range ta {
		if _, ok := tb[t]; ok {
			overlap++
		}
	}
	return overlap >= 1
}

// ExecutedPlanTools returns tools that actually ran for this recall synthesis.
func ExecutedPlanTools(plan QueryPlan, temporalApplied, enumerated, abstained bool) []string {
	tools := []string{"search"}
	add := func(t string) {
		for _, e := range tools {
			if e == t {
				return
			}
		}
		tools = append(tools, t)
	}
	if temporalApplied {
		add("temporal_resolve")
	}
	if enumerated {
		add("enumerate")
	}
	if plan.NeedsMultiHop {
		add("evidence_set")
	}
	if abstained {
		add("abstain")
	}
	return tools
}
