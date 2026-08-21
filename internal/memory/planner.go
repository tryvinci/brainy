package memory

import (
	"strings"
	"unicode"
)

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
	if (plan.NeedsMultiHop || plan.NeedsEnumeration || looksWhenEventQuery(query)) && hopPlanAllowed(query) {
		plan.Hops = buildTypedHops(query)
	}
	plan.PreferredModeHint = preferredModeHint(plan)
	return plan
}

// buildTypedHops emits resolve_entity → follow_relation/fetch_predicate subgoals.
func buildTypedHops(query string) []HopStep {
	toks := contentBearingTokens(tokenize(query))
	entities := hopQueryEntities(query)
	preds := predicateHintsFromQuery(query)
	hops := make([]HopStep, 0, 6)
	for i, ent := range entities {
		out := "e1"
		if i > 0 {
			out = "e" + itoa(i+1)
		}
		hops = append(hops, HopStep{
			Kind:   "resolve_entity",
			Entity: ent,
			Probe:  ent,
			Output: out,
		})
	}
	entity := ""
	bridge := ""
	if len(entities) > 0 {
		entity = entities[0]
	}
	if len(entities) > 1 {
		bridge = entities[1]
	}
	if role := possessiveKinshipRole(query); role != "" && entity != "" && bridge == "" {
		return buildKinshipHops(query, entity, role, preds)
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
	usePreds := preds
	if len(usePreds) > 3 {
		usePreds = usePreds[:3]
	}
	// Multi-entity joins prove one predicate; extra hints (activity on a
	// like-query) crowd search-fallback lists over the shared fact.
	if len(entities) >= 2 && len(usePreds) > 0 {
		usePreds = usePreds[:1]
	}
	targets := entities
	if len(targets) == 0 {
		targets = []string{""}
	}
	n := 0
	for _, pred := range usePreds {
		for ti, ent := range targets {
			n++
			outKey := "ans"
			if n > 1 {
				outKey = "ans" + itoa(n)
			}
			fetch := HopStep{
				Kind:      "fetch_predicate",
				Entity:    firstNonEmpty(ent, entity, bridge),
				Predicate: pred,
				Probe:     hopProbe(toks, firstNonEmpty(ent, entity), pred),
				Output:    outKey,
			}
			if ent != "" {
				dep := "e1"
				if ti > 0 {
					dep = "e" + itoa(ti+1)
				}
				fetch.DependsOn = []string{dep}
			}
			if relationFollowPredicate(pred) {
				fetch.Kind = "follow_relation"
			} else if looksPlaceOrPersonSlot(query) {
				fetch.Kind = "answer_slot"
			}
			hops = append(hops, fetch)
		}
	}
	return hops
}

func hopQueryEntities(query string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, 2)
	recipient := strings.ToLower(transferRecipient(query))
	add := func(n string) bool {
		n = strings.TrimSpace(n)
		n = strings.TrimSuffix(n, "'s")
		n = strings.TrimSuffix(n, "’s")
		n = strings.Trim(n, "'\"?,.")
		if n == "" {
			return false
		}
		key := strings.ToLower(n)
		if recipient != "" && key == recipient {
			return false
		}
		if _, stop := hopEntityStop[key]; stop {
			return false
		}
		if _, ok := seen[key]; ok {
			return false
		}
		seen[key] = struct{}{}
		out = append(out, key)
		return true
	}
	// Coordinated people ("Tim and John") are the join, not the first topic noun.
	for _, n := range coordinatedPersonMentions(query) {
		add(n)
	}
	if n := personAfterCue(query, "with"); n != "" {
		add(n)
	}
	join := hopJoinCue(query) || len(out) >= 2
	limit := 1
	if join {
		limit = 2
	}
	if n := personAfterAuxiliary(query); n != "" {
		add(n)
	}
	if len(out) >= limit {
		return out[:limit]
	}
	for _, n := range capitalizedMentionTokens(query) {
		if skipHowManyHead(query, n) {
			continue
		}
		add(n)
		if len(out) >= limit {
			return out
		}
	}
	if len(out) == 0 {
		for _, n := range nameLikeTokens(contentBearingTokens(tokenize(query))) {
			add(n)
			if len(out) >= limit {
				return out
			}
		}
	}
	return out
}

func hopJoinCue(query string) bool {
	lower := strings.ToLower(query)
	if strings.Contains(lower, "both") || strings.Contains(lower, "together") ||
		strings.Contains(lower, " in common") {
		return true
	}
	if len(coordinatedPersonMentions(query)) >= 2 {
		return true
	}
	return personAfterCue(query, "with") != ""
}

func coordinatedPersonMentions(query string) []string {
	fields := strings.Fields(query)
	out := make([]string, 0, 2)
	for i := 0; i < len(fields)-2; i++ {
		a := strings.Trim(fields[i], "?,.!\"'")
		conj := strings.ToLower(strings.Trim(fields[i+1], "?,.!\"'"))
		b := strings.Trim(fields[i+2], "?,.!\"'")
		if conj != "and" && conj != "or" {
			continue
		}
		if !looksHopPerson(a) || !looksHopPerson(b) {
			continue
		}
		out = append(out, a, b)
		return out
	}
	return out
}

func personAfterCue(query, cue string) string {
	cue = strings.ToLower(strings.TrimSpace(cue))
	if cue == "" {
		return ""
	}
	fields := strings.Fields(query)
	for i, raw := range fields {
		if !strings.EqualFold(strings.Trim(raw, "?,.!\""), cue) || i+1 >= len(fields) {
			continue
		}
		next := strings.Trim(fields[i+1], "?,.!\"'")
		if looksHopPerson(next) {
			return next
		}
	}
	return ""
}

func personAfterAuxiliary(query string) string {
	fields := strings.Fields(query)
	for i := 0; i < len(fields)-1; i++ {
		w := strings.ToLower(strings.Trim(fields[i], "?,.!\""))
		switch w {
		case "does", "did", "do", "has", "have":
		default:
			continue
		}
		next := strings.Trim(fields[i+1], "?,.!\"'")
		if looksHopPerson(next) {
			return next
		}
	}
	return ""
}

func skipHowManyHead(query, name string) bool {
	lower := strings.ToLower(query)
	i := strings.Index(lower, "how many ")
	if i < 0 {
		return false
	}
	rest := strings.TrimSpace(query[i+len("how many "):])
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return false
	}
	head := strings.Trim(fields[0], "?,.!'\"")
	return strings.EqualFold(head, name)
}

func looksHopPerson(w string) bool {
	w = strings.TrimSpace(w)
	w = strings.TrimSuffix(w, "'s")
	w = strings.TrimSuffix(w, "’s")
	w = strings.Trim(w, "'\"?,.")
	if w == "" {
		return false
	}
	r := []rune(w)
	if len(r) < 3 || !unicode.IsUpper(r[0]) {
		return false
	}
	switch strings.ToLower(w) {
	case "what", "which", "who", "where", "when", "why", "how":
		return false
	}
	_, stop := hopEntityStop[strings.ToLower(w)]
	return !stop
}

var kinshipRoles = []string{
	"mother", "father", "mom", "dad", "parent", "parents",
	"partner", "spouse", "wife", "husband",
	"brother", "sister", "sibling", "family",
}

func possessiveKinshipRole(query string) string {
	lower := strings.ToLower(query)
	for _, role := range kinshipRoles {
		if strings.Contains(lower, "'s "+role) || strings.Contains(lower, "’s "+role) ||
			strings.Contains(lower, " her "+role) || strings.Contains(lower, " his "+role) ||
			strings.Contains(lower, " their "+role) {
			return role
		}
	}
	return ""
}

func buildKinshipHops(query, entity, role string, preds []string) []HopStep {
	hops := []HopStep{
		{Kind: "resolve_entity", Entity: entity, Probe: entity, Output: "e1"},
		{
			Kind:      "follow_relation",
			Entity:    entity,
			Predicate: PredicateFamilyMember,
			Probe:     hopProbe(contentBearingTokens(tokenize(query)), entity, role),
			Output:    "e_rel",
			DependsOn: []string{"e1"},
		},
	}
	fetchPred := ""
	for _, p := range preds {
		if p != PredicateFamilyMember && p != PredicateRelationshipStatus {
			fetchPred = p
			break
		}
	}
	if fetchPred == "" && looksPlaceOrPersonSlot(query) {
		fetchPred = PredicateEvent
	}
	fetch := HopStep{
		Kind:      "fetch_predicate",
		Predicate: fetchPred,
		Probe:     hopProbe(contentBearingTokens(tokenize(query)), role, fetchPred),
		Output:    "ans",
		DependsOn: []string{"e_rel"},
	}
	if fetchPred != "" && relationFollowPredicate(fetchPred) {
		fetch.Kind = "follow_relation"
	} else if looksPlaceOrPersonSlot(query) {
		fetch.Kind = "answer_slot"
	}
	return append(hops, fetch)
}

func capitalizedMentionTokens(query string) []string {
	fields := strings.Fields(query)
	out := make([]string, 0, 2)
	for i, raw := range fields {
		w := strings.Trim(raw, "?,.!\"'")
		if w == "" {
			continue
		}
		if i == 0 {
			switch strings.ToLower(w) {
			case "what", "which", "who", "where", "when", "why", "how":
				continue
			}
		}
		r := []rune(w)
		if len(r) < 3 || !unicode.IsUpper(r[0]) {
			continue
		}
		out = append(out, w)
	}
	return out
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
	"animal": {}, "animals": {}, "both": {},
	"they": {}, "them": {}, "their": {}, "theirs": {},
	"partner": {}, "mother": {}, "father": {}, "family": {},
	"collectible": {}, "collectibles": {},
	"dogs": {}, "dog": {}, "pets": {}, "pet": {}, "names": {}, "name": {},
	"instruments": {}, "instrument": {}, "items": {}, "item": {},
	"locations": {}, "location": {}, "tricks": {}, "trick": {},
	"events": {}, "event": {}, "stressor": {}, "stressors": {},
	"organizations": {}, "organization": {}, "colleagues": {},
	"travels": {}, "travel": {},
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

func hopPlanAllowed(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if strings.HasPrefix(q, "how long") || strings.HasPrefix(q, "how old") {
		return false
	}
	return true
}

func hopComposeAllowed(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if strings.HasPrefix(q, "when ") || strings.HasPrefix(q, "how long") || strings.HasPrefix(q, "how old") {
		return false
	}
	return true
}

func looksWhenEventQuery(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	return strings.HasPrefix(q, "when ")
}

func transferRecipient(query string) string {
	if looksWhoQuery(query) {
		return ""
	}
	lower := strings.ToLower(query)
	if !queryHasToken(query, "given", "gave", "give") && !strings.Contains(lower, "suggest") {
		return ""
	}
	return personAfterCue(query, "to")
}

func looksCountQuery(query string) bool {
	q := strings.ToLower(query)
	return strings.Contains(q, "how many") || strings.Contains(q, "how much")
}

func looksPolarQuery(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if looksCountQuery(q) {
		return false
	}
	for _, p := range []string{"has ", "have ", "had ", "did ", "does ", "do ", "is ", "was ", "were ", "can "} {
		if strings.HasPrefix(q, p) {
			return true
		}
	}
	return false
}

func looksUnwindQuery(query string) bool {
	q := strings.ToLower(query)
	return strings.Contains(q, "do to") || strings.Contains(q, "unwind") || strings.Contains(q, "relax")
}

func looksWhoQuery(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	return strings.HasPrefix(q, "who ") || strings.Contains(q, " who ")
}

func looksSuperlativeQuery(query string) bool {
	q := strings.ToLower(query)
	return strings.Contains(q, "most frequently") || strings.Contains(q, "most often") ||
		strings.Contains(q, "the biggest") || strings.Contains(q, "biggest") ||
		strings.Contains(q, "most common")
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
	case plan.PrimaryIntent == IntentAggregation:
		return "answer"
	case plan.NeedsEnumeration:
		return "enumerate"
	case len(plan.Hops) > 0:
		return "answer"
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
