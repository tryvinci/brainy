package memory

import "strings"

// QueryPlan is the deterministic typed plan for /recall (program Phase 4).
// It does not rewrite SearchOpt; it records what synthesis should do with
// retrieved evidence and temporal reads.
type QueryPlan struct {
	Intents           []string  `json:"intents"`
	PrimaryIntent     string    `json:"primary_intent"`
	NeedsTemporal     bool      `json:"needs_temporal"`
	NeedsEnumeration  bool      `json:"needs_enumeration"`
	NeedsMultiHop     bool      `json:"needs_multi_hop"`
	NeedsAbstention   bool      `json:"needs_abstention"`
	CoverageTargets   []string  `json:"coverage_targets,omitempty"`
	Hops              []HopStep `json:"hops,omitempty"`
	Tools             []string  `json:"tools,omitempty"`
	BudgetPasses      int       `json:"budget_passes"`
	PreferredModeHint string    `json:"preferred_mode_hint,omitempty"`
}

// HopStep is a typed multi-hop subgoal (resolve entity, then fetch predicate).
type HopStep struct {
	Kind      string   `json:"kind"` // resolve_entity | fetch_predicate | follow_relation | answer_slot
	Entity    string   `json:"entity,omitempty"`
	Predicate string   `json:"predicate,omitempty"`
	Probe     string   `json:"probe,omitempty"`
	Output    string   `json:"output,omitempty"`
	DependsOn []string `json:"depends_on,omitempty"`
}

// EvidencePacket is the bounded evidence set handed to synthesis/oracles.
// ContextEvidence is the broad SearchOpt hit set (reader/context assembly).
// ProofChain is hop join evidence (answer-status / hop_join_proven). Hops must
// not replace context.
type EvidencePacket struct {
	MemoryIDs       []string       `json:"memory_ids"`
	Contents        []string       `json:"contents,omitempty"`
	ContextEvidence []PacketItem   `json:"context_evidence,omitempty"`
	ProofChain      []PacketItem   `json:"proof_chain,omitempty"`
	Items           []PacketItem   `json:"items,omitempty"`
	Predicates      []string       `json:"predicates,omitempty"`
	TemporalAnswer  string         `json:"temporal_answer,omitempty"`
	Coverage        map[string]any `json:"coverage,omitempty"`
	Plan            QueryPlan      `json:"plan"`
}

// PlanQuery builds a typed plan from the deterministic intent classifier.
func PlanQuery(query string, intents []string) QueryPlan {
	if len(intents) == 0 {
		intents = AnalyzeQueryIntents(query)
	}
	plan := QueryPlan{
		Intents:         append([]string(nil), intents...),
		BudgetPasses:    1,
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
	if (plan.NeedsMultiHop || plan.NeedsEnumeration) && hopComposeAllowed(query) {
		plan.Hops = buildTypedHops(query)
	}
	plan.PreferredModeHint = preferredModeHint(plan)
	return plan
}

// buildTypedHops emits resolve_entity → follow_relation/fetch_predicate subgoals.
func buildTypedHops(query string) []HopStep {
	toks := contentBearingTokens(tokenize(query))
	names := nameLikeTokens(toks)
	preds := predicateHintsFromQuery(query)
	hops := make([]HopStep, 0, 4)
	entity := hopEntityName(names)
	if entity != "" {
		hops = append(hops, HopStep{
			Kind:   "resolve_entity",
			Entity: entity,
			Probe:  entity,
			Output: "e1",
		})
	}
	bridge := ""
	if len(names) > 1 {
		bridge = names[1]
	}
	if len(preds) == 0 {
		probeParts := make([]string, 0, 2)
		if bridge != "" {
			probeParts = append(probeParts, bridge)
		} else if entity != "" {
			probeParts = append(probeParts, entity)
		}
		if len(probeParts) == 0 {
			for _, t := range toks {
				if t == entity {
					continue
				}
				probeParts = append(probeParts, t)
				break
			}
		}
		if len(probeParts) > 0 {
			fetch := HopStep{
				Kind:   "fetch_predicate",
				Entity: firstNonEmpty(bridge, entity),
				Probe:  strings.Join(probeParts, " "),
				Output: "ans",
			}
			if entity != "" {
				fetch.DependsOn = []string{"e1"}
			}
			if looksPlaceOrPersonSlot(query) {
				fetch.Kind = "answer_slot"
			}
			hops = append(hops, fetch)
		}
		return hops
	}
	for i, pred := range preds {
		if i >= 3 {
			break
		}
		outKey := "ans"
		if i > 0 {
			outKey = "ans" + itoa(i+1)
		}
		fetch := HopStep{
			Kind:      "fetch_predicate",
			Entity:    firstNonEmpty(entity, bridge),
			Predicate: pred,
			Probe:     hopProbe(toks, entity, pred),
			Output:    outKey,
		}
		if entity != "" {
			fetch.DependsOn = []string{"e1"}
		}
		if relationFollowPredicate(pred) {
			fetch.Kind = "follow_relation"
		} else if looksPlaceOrPersonSlot(query) {
			fetch.Kind = "answer_slot"
		}
		hops = append(hops, fetch)
	}
	return hops
}

func hopProbe(toks []string, entity string, extra ...string) string {
	seen := map[string]struct{}{}
	parts := make([]string, 0, 4)
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		key := strings.ToLower(s)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		parts = append(parts, s)
	}
	add(entity)
	for _, e := range extra {
		add(e)
	}
	entKey := strings.ToLower(strings.TrimSpace(entity))
	for _, t := range toks {
		if strings.ToLower(t) == entKey {
			continue
		}
		add(t)
		if len(parts) >= 4 {
			break
		}
	}
	return strings.Join(parts, " ")
}

func hopEntityName(names []string) string {
	for _, n := range names {
		n = strings.TrimSpace(n)
		n = strings.TrimSuffix(n, "'s")
		n = strings.TrimSuffix(n, "’s")
		n = strings.Trim(n, "'\"")
		if n == "" {
			continue
		}
		if _, stop := hopEntityStop[strings.ToLower(n)]; stop {
			continue
		}
		return n
	}
	return ""
}

var hopEntityStop = map[string]struct{}{
	"career": {}, "path": {}, "occupation": {}, "job": {}, "work": {},
	"books": {}, "book": {}, "activities": {}, "activity": {},
	"hobbies": {}, "hobby": {}, "kids": {}, "children": {}, "child": {},
	"country": {}, "years": {}, "year": {}, "identity": {},
	"relationship": {}, "status": {}, "likes": {}, "like": {},
	"places": {}, "place": {}, "enjoy": {}, "enjoys": {},
	"read": {}, "reading": {}, "moved": {}, "move": {},
	"pursue": {}, "decided": {},
	"long": {}, "ago": {}, "current": {}, "currently": {},
	"group": {}, "friends": {}, "friend": {},
	"research": {}, "researched": {}, "researching": {},
}

func relationFollowPredicate(pred string) bool {
	switch pred {
	case PredicateOrigin, PredicateResidence, PredicateActivity,
		PredicateMediaConsumed, PredicateOccupation, PredicateFamilyMember,
		PredicateEducation, PredicatePlan, PredicateEvent, PredicateIdentity,
		PredicatePreference, PredicateRelationshipStatus:
		return true
	}
	return false
}

func looksPlaceOrPersonSlot(query string) bool {
	q := strings.ToLower(query)
	for _, cue := range []string{"where ", "who ", "which "} {
		if strings.Contains(q, cue) {
			return true
		}
	}
	return false
}

func hopComposeAllowed(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if strings.HasPrefix(q, "when ") || strings.HasPrefix(q, "how long") || strings.HasPrefix(q, "how old") {
		return false
	}
	return true
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// nextHopProbe returns a single-subgoal second-pass probe from typed hops
// when the prior packet has not yet satisfied coverage.
func nextHopProbe(plan QueryPlan, pkt EvidencePacket) string {
	if len(plan.Hops) == 0 {
		return strings.Join(uncoveredTargets(pkt), " ")
	}
	bridges := countRole(pkt.Items, "bridge")
	directs := countRole(pkt.Items, "direct")
	for _, hop := range plan.Hops {
		probe := strings.TrimSpace(hop.Probe)
		if probe == "" {
			continue
		}
		switch hop.Kind {
		case "resolve_entity":
			if bridges == 0 {
				return probe
			}
		case "fetch_predicate", "follow_relation", "answer_slot":
			if directs == 0 || len(uncoveredTargets(pkt)) > 0 {
				return probe
			}
		default:
			return probe
		}
	}
	if probe := strings.TrimSpace(plan.Hops[len(plan.Hops)-1].Probe); probe != "" {
		return probe
	}
	return strings.Join(uncoveredTargets(pkt), " ")
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
// ContextEvidence is typed (R5B); Contents is a compatibility projection.
func BuildEvidencePacket(plan QueryPlan, results []SearchResult, explain map[string]any) EvidencePacket {
	pkt := EvidencePacket{
		Plan:            plan,
		MemoryIDs:       make([]string, 0, len(results)),
		Contents:        make([]string, 0, len(results)),
		ContextEvidence: make([]PacketItem, 0, len(results)),
	}
	for _, r := range results {
		item := packetItemFromSearch(r)
		if item.MemoryID != "" {
			pkt.MemoryIDs = append(pkt.MemoryIDs, item.MemoryID)
		}
		if item.Content != "" {
			pkt.Contents = append(pkt.Contents, item.Content)
			pkt.ContextEvidence = append(pkt.ContextEvidence, item)
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

func packetItemFromSearch(r SearchResult) PacketItem {
	content := strings.TrimSpace(r.Content)
	if strings.HasSuffix(content, "?") {
		content = ""
	}
	pred := searchResultPredicate(r)
	val := searchResultValueNorm(r)
	if val == "" {
		val = structuredValueOf(r)
	}
	subj := ""
	eid := ""
	if r.Explain != nil {
		if s, ok := r.Explain["subject"].(string); ok {
			subj = strings.TrimSpace(s)
		}
		if id, ok := r.Explain["entity_id"].(string); ok {
			eid = strings.TrimSpace(id)
		}
	}
	return PacketItem{
		EvidenceID: r.MemoryID,
		MemoryID:   r.MemoryID,
		FactID:     r.MemoryID,
		Content:    content,
		Predicate:  pred,
		Subject:    subj,
		Value:      val,
		EntityID:   eid,
		Span:       content,
		Role:       "context",
		Score:      r.Score,
	}
}

func packetItemsFromContents(contents, ids []string) []PacketItem {
	out := make([]PacketItem, 0, len(contents))
	for i, c := range contents {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		id := ""
		if i < len(ids) {
			id = ids[i]
		}
		out = append(out, PacketItem{
			EvidenceID: id,
			MemoryID:   id,
			FactID:     id,
			Content:    c,
			Span:       c,
			Role:       "context",
		})
	}
	return out
}

func packetCoverageSatisfied(plan QueryPlan, pkt EvidencePacket) bool {
	if len(pkt.Contents) == 0 && pkt.TemporalAnswer == "" {
		return false
	}
	if plan.NeedsTemporal && pkt.TemporalAnswer != "" && !plan.NeedsMultiHop {
		return true
	}
	if plan.NeedsMultiHop {
		// Typed hop join is the only proven MH coverage path when hops exist.
		if proven, _ := pkt.Coverage["hop_join_proven"].(bool); proven {
			return true
		}
		if len(plan.Hops) > 0 {
			// Lexical bridge/direct is context only — not a proven chain.
			return len(pkt.Contents) >= 1 && pkt.TemporalAnswer != ""
		}
		// Legacy (no typed hops): require bridge+direct or temporal+content.
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
