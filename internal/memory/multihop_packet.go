package memory

import (
	"strings"
)

// multiHopTargets derives coverage targets for multi-hop questions:
// named entities (bridge candidates) plus distinctive content cues (answer side).
func multiHopTargets(query string) []string {
	toks := contentBearingTokens(tokenize(query))
	names := nameLikeTokens(toks)
	out := make([]string, 0, 6)
	seen := map[string]struct{}{}
	add := func(t string) {
		t = strings.ToLower(strings.TrimSpace(t))
		if len(t) < 3 {
			return
		}
		if _, ok := seen[t]; ok {
			return
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	for _, n := range names {
		add(n)
	}
	// Prefer longer distinctive cues next (activities, places, titles).
	ranked := append([]string(nil), toks...)
	for i := 0; i < len(ranked); i++ {
		for j := i + 1; j < len(ranked); j++ {
			if len(ranked[j]) > len(ranked[i]) {
				ranked[i], ranked[j] = ranked[j], ranked[i]
			}
		}
	}
	for _, t := range ranked {
		add(t)
		if len(out) >= 6 {
			break
		}
	}
	if len(out) == 0 {
		return []string{"bridge", "answer"}
	}
	if len(out) == 1 {
		out = append(out, "answer")
	}
	return out
}

// bindPacketToTargets assigns bridge/direct roles and records which coverage
// targets each item satisfies. Reorders items so a linked chain surfaces first.
func bindPacketToTargets(pkt *EvidencePacket, results []SearchResult, query string, targets []string) {
	if pkt == nil {
		return
	}
	if len(targets) == 0 {
		targets = multiHopTargets(query)
	}
	qTokens := contentBearingTokens(tokenize(query))
	items := make([]PacketItem, 0, len(results))
	targetHits := map[string]int{}

	for _, r := range results {
		content := strings.TrimSpace(r.Content)
		if content == "" || strings.HasSuffix(content, "?") {
			continue
		}
		lower := strings.ToLower(content)
		covered := make([]string, 0, 2)
		for _, t := range targets {
			if strings.Contains(lower, strings.ToLower(t)) {
				covered = append(covered, t)
				targetHits[t]++
			}
		}
		overlap := 0
		for _, t := range qTokens {
			if strings.Contains(lower, t) {
				overlap++
			}
		}
		role := "context"
		switch {
		case len(covered) >= 2:
			role = "direct"
		case len(covered) == 1 && overlap >= 1:
			// First-target hits tend to be bridges in multi-hop "who/what about X that Y".
			if len(targets) >= 2 && covered[0] == targets[0] {
				role = "bridge"
			} else {
				role = "direct"
			}
		case overlap >= 2:
			role = "direct"
		case overlap == 1:
			role = "bridge"
		}
		pred := ""
		if r.Explain != nil {
			if p, ok := r.Explain["predicate"].(string); ok {
				pred = p
			}
		}
		// Prefer predicate match as direct evidence when query asks for that slot.
		if pred != "" {
			for _, t := range targets {
				if strings.EqualFold(pred, t) || strings.Contains(strings.ToLower(pred), strings.ToLower(t)) {
					if !containsFold(covered, t) {
						covered = append(covered, t)
						targetHits[t]++
					}
					role = "direct"
				}
			}
		}
		items = append(items, PacketItem{
			MemoryID:  r.MemoryID,
			Content:   content,
			Predicate: pred,
			Role:      role,
			Score:     r.Score,
			Targets:   covered,
		})
	}

	// Prefer a bridge+direct pair that jointly covers distinct targets.
	ordered := orderBridgeChain(items, targets)
	pkt.Items = ordered
	pkt.Contents = make([]string, 0, len(ordered))
	pkt.MemoryIDs = make([]string, 0, len(ordered))
	for _, it := range ordered {
		pkt.Contents = append(pkt.Contents, it.Content)
		if it.MemoryID != "" {
			pkt.MemoryIDs = append(pkt.MemoryIDs, it.MemoryID)
		}
	}
	uncovered := make([]string, 0)
	for _, t := range targets {
		if targetHits[t] == 0 {
			uncovered = append(uncovered, t)
		}
	}
	if pkt.Coverage == nil {
		pkt.Coverage = map[string]any{}
	}
	pkt.Coverage["targets"] = len(targets)
	pkt.Coverage["target_list"] = targets
	pkt.Coverage["uncovered"] = uncovered
	pkt.Coverage["bridge_count"] = countRole(ordered, "bridge")
	pkt.Coverage["direct_count"] = countRole(ordered, "direct")
}

func countRole(items []PacketItem, role string) int {
	n := 0
	for _, it := range items {
		if it.Role == role {
			n++
		}
	}
	return n
}

func containsFold(vals []string, needle string) bool {
	needle = strings.ToLower(strings.TrimSpace(needle))
	for _, v := range vals {
		if strings.ToLower(strings.TrimSpace(v)) == needle {
			return true
		}
	}
	return false
}

func orderBridgeChain(items []PacketItem, targets []string) []PacketItem {
	if len(items) <= 1 {
		return items
	}
	var bridges, directs, rest []PacketItem
	for _, it := range items {
		switch it.Role {
		case "bridge":
			bridges = append(bridges, it)
		case "direct":
			directs = append(directs, it)
		default:
			rest = append(rest, it)
		}
	}
	out := make([]PacketItem, 0, len(items))
	// Pick best bridge/direct pair with complementary target coverage.
	bestB, bestD := -1, -1
	bestScore := -1
	for i, b := range bridges {
		bset := map[string]struct{}{}
		for _, t := range b.Targets {
			bset[t] = struct{}{}
		}
		for j, d := range directs {
			union := len(bset)
			for _, t := range d.Targets {
				if _, ok := bset[t]; !ok {
					union++
				}
			}
			score := union*10 + int(d.Score*5) + int(b.Score*3)
			if score > bestScore {
				bestScore = score
				bestB, bestD = i, j
			}
		}
	}
	if bestB >= 0 && bestD >= 0 {
		out = append(out, bridges[bestB], directs[bestD])
		for i, b := range bridges {
			if i != bestB {
				out = append(out, b)
			}
		}
		for j, d := range directs {
			if j != bestD {
				out = append(out, d)
			}
		}
		out = append(out, rest...)
		return out
	}
	out = append(out, bridges...)
	out = append(out, directs...)
	out = append(out, rest...)
	_ = targets
	return out
}

// composeMultiHopAnswer builds a short deterministic answer from a bridge+direct chain.
func composeMultiHopAnswer(pkt EvidencePacket) string {
	if hops := hopResultsFromPacket(pkt); len(hops) > 0 {
		if composed := composeFromHopValues(hops); composed != "" {
			return composed
		}
		// Search-fallback hop Values are not proof. Extract typed slots from
		// hop contents (e.g. "Joanna likes turtles") before slogans.
		if composed := composeFromHopContents(hops); composed != "" {
			return composed
		}
		// Hops planned but unproven: use structured packet values, not slogans
		// or resolve-only mentions.
		if composed := composeFromPacketStructuredValues(pkt, hops); composed != "" {
			return composed
		}
		return ""
	}
	var bridge, direct string
	for _, it := range pkt.Items {
		content := strings.TrimSpace(it.Content)
		if content == "" || looksTitleCaseSlogan(content) {
			continue
		}
		if it.Role == "bridge" && bridge == "" {
			bridge = content
		}
		if it.Role == "direct" && direct == "" {
			direct = content
		}
	}
	switch {
	case bridge != "" && direct != "" && bridge != direct:
		return direct + " (via " + truncateRunes(bridge, 80) + ")"
	case direct != "":
		return direct
	case bridge != "":
		return bridge
	default:
		return ""
	}
}

func composeFromHopContents(hops []HopResult) string {
	if hopFetchEntityCount(hops) >= 2 {
		return joinTitledHopValues(hopSharedContentValues(hops))
	}
	vals := hopContentSlotValues(hops)
	if len(vals) > 4 {
		vals = vals[:4]
	}
	if len(vals) == 0 {
		return ""
	}
	return strings.Join(vals, ", ")
}

func composeFromPacketStructuredValues(pkt EvidencePacket, hops []HopResult) string {
	echo := map[string]struct{}{}
	for _, h := range hops {
		if v := strings.ToLower(strings.TrimSpace(h.Entity)); v != "" {
			echo[v] = struct{}{}
		}
		if v := strings.ToLower(strings.TrimSpace(h.Value)); v != "" && h.Kind == "resolve_entity" {
			echo[v] = struct{}{}
		}
	}
	addable := func(v string) bool {
		v = strings.TrimSpace(v)
		if v == "" || anaphoricSlotValue(v) || looksTitleCaseSlogan(v) {
			return false
		}
		if utf8Len(v) > 80 {
			return false
		}
		if _, ok := echo[strings.ToLower(v)]; ok {
			return false
		}
		return true
	}
	seen := map[string]struct{}{}
	vals := make([]string, 0, 4)
	collect := func(items []PacketItem) {
		for _, it := range items {
			if it.Role == "bridge" {
				continue
			}
			v := strings.TrimSpace(it.Value)
			if v == "" {
				if extracted, ok := slotValueFromMemoryContent(it.Content); ok {
					v = extracted
				}
			}
			if !addable(v) {
				continue
			}
			key := strings.ToLower(v)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			vals = append(vals, v)
			if len(vals) >= 4 {
				return
			}
		}
	}
	collect(pkt.ProofChain)
	if len(vals) == 0 {
		collect(pkt.Items)
	}
	if len(vals) == 0 {
		return ""
	}
	return strings.Join(vals, ", ")
}

func hopResultsFromPacket(pkt EvidencePacket) []HopResult {
	if pkt.Coverage == nil {
		return nil
	}
	raw, ok := pkt.Coverage["hop_results"]
	if !ok {
		return nil
	}
	hops, ok := raw.([]HopResult)
	if !ok {
		return nil
	}
	return hops
}

func uncoveredTargets(pkt EvidencePacket) []string {
	if pkt.Coverage == nil {
		return nil
	}
	raw, ok := pkt.Coverage["uncovered"]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, x := range v {
			if s, ok := x.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

// mergeSearchResults unions two result lists by MemoryID, preferring higher scores, capped at limit.
func mergeSearchResults(a, b []SearchResult, limit int) []SearchResult {
	if limit <= 0 {
		limit = 30
	}
	byID := map[string]SearchResult{}
	order := make([]string, 0, len(a)+len(b))
	add := func(r SearchResult) {
		id := r.MemoryID
		if id == "" {
			id = strings.ToLower(strings.TrimSpace(r.Content))
		}
		if id == "" {
			return
		}
		if existing, ok := byID[id]; ok {
			if r.Score > existing.Score {
				byID[id] = r
			}
			return
		}
		byID[id] = r
		order = append(order, id)
	}
	for _, r := range a {
		add(r)
	}
	for _, r := range b {
		add(r)
	}
	out := make([]SearchResult, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
		if len(out) >= limit {
			break
		}
	}
	return out
}
