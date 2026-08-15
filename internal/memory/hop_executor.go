package memory

import (
	"context"
	"strings"
)

// HopResult is the typed output of one executed HopStep.
type HopResult struct {
	HopIndex   int      `json:"hop_index"`
	Kind       string   `json:"kind"`
	OutputKey  string   `json:"output_key,omitempty"`
	Value      string   `json:"value,omitempty"`
	Values     []string `json:"values,omitempty"` // all typed destinations (enumeration)
	Entity     string   `json:"entity,omitempty"`
	Predicate  string   `json:"predicate,omitempty"`
	MemoryIDs  []string `json:"memory_ids,omitempty"`
	Contents   []string `json:"contents,omitempty"`
	Source     string   `json:"source,omitempty"` // typed_store | search_fallback | unresolved
	DependsOn  []string `json:"depends_on,omitempty"`
}

// executeTypedHops runs plan hops against typed stores, falling back to SearchOpt
// only when typed backends miss. Results are keyed by HopStep.Output.
func (s *Service) executeTypedHops(
	ctx context.Context,
	tenantID, subjectID, vertical string,
	includeHistorical bool,
	plan QueryPlan,
	topK int,
) ([]HopResult, map[string]HopResult) {
	out := make([]HopResult, 0, len(plan.Hops))
	byKey := map[string]HopResult{}
	if len(plan.Hops) == 0 {
		return out, byKey
	}
	for i, hop := range plan.Hops {
		res := HopResult{
			HopIndex:  i,
			Kind:      hop.Kind,
			OutputKey: firstNonEmpty(hop.Output, hopOutputDefault(i, hop.Kind)),
			Entity:    hop.Entity,
			Predicate: hop.Predicate,
			DependsOn: append([]string(nil), hop.DependsOn...),
			Source:    "unresolved",
		}
		// Wire DependsOn values into entity when prior hop resolved one.
		if res.Entity == "" {
			for _, dep := range hop.DependsOn {
				if prev, ok := byKey[dep]; ok && strings.TrimSpace(prev.Value) != "" {
					res.Entity = prev.Value
					break
				}
			}
		} else {
			for _, dep := range hop.DependsOn {
				if prev, ok := byKey[dep]; ok && strings.TrimSpace(prev.Value) != "" {
					// Prefer resolved canonical entity over raw mention.
					res.Entity = prev.Value
					break
				}
			}
		}

		switch hop.Kind {
		case "resolve_entity":
			s.resolveEntityHop(ctx, tenantID, subjectID, vertical, includeHistorical, hop, &res, topK)
		case "follow_relation":
			s.followRelationHop(ctx, tenantID, subjectID, vertical, includeHistorical, hop, &res, topK)
		case "fetch_predicate", "answer_slot":
			s.fetchPredicateHop(ctx, tenantID, subjectID, vertical, includeHistorical, hop, &res, topK)
		default:
			s.searchFallbackHop(ctx, tenantID, subjectID, vertical, includeHistorical, hop, &res, topK)
		}
		out = append(out, res)
		if res.OutputKey != "" {
			byKey[res.OutputKey] = res
		}
	}
	return out, byKey
}

func hopOutputDefault(i int, kind string) string {
	switch kind {
	case "resolve_entity", "follow_relation":
		return "e" + itoa(i+1)
	default:
		return "ans"
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func (s *Service) resolveEntityHop(
	ctx context.Context,
	tenantID, subjectID, vertical string,
	includeHistorical bool,
	hop HopStep,
	res *HopResult,
	topK int,
) {
	mention := firstNonEmpty(res.Entity, hop.Entity, hop.Probe)
	res.Entity = mention
	// R4: resolved entity ID is the mention, not a random hub record's subject.
	res.Value = mention
	if linker, ok := s.store.(EntityLinker); ok && mention != "" {
		boosts, err := linker.EntityHubBoosts(ctx, tenantID, subjectID, []string{mention})
		if err == nil && len(boosts) > 0 {
			ids := make([]string, 0, len(boosts))
			for id := range boosts {
				ids = append(ids, id)
			}
			res.MemoryIDs = ids
			if len(ids) > 0 {
				if rec, err := s.store.GetMemory(ctx, tenantID, subjectID, ids[0]); err == nil {
					res.Contents = append(res.Contents, rec.Content)
				}
			}
			res.Source = "typed_store"
			return
		}
	}
	s.searchFallbackHop(ctx, tenantID, subjectID, vertical, includeHistorical, HopStep{
		Kind: hop.Kind, Entity: mention, Probe: firstNonEmpty(hop.Probe, mention),
	}, res, topK)
	if res.Value == "" {
		res.Value = mention
	}
}

func (s *Service) followRelationHop(
	ctx context.Context,
	tenantID, subjectID, vertical string,
	includeHistorical bool,
	hop HopStep,
	res *HopResult,
	topK int,
) {
	entity := firstNonEmpty(res.Entity, hop.Entity)
	pred := firstNonEmpty(hop.Predicate, "")
	res.Entity = entity
	res.Predicate = pred
	if indexer, ok := s.store.(RelationIndexer); ok && entity != "" {
		rels, err := indexer.ListRelationsFrom(ctx, tenantID, subjectID, strings.ToLower(entity), pred, topK)
		if err == nil && len(rels) > 0 {
			seenDst := map[string]struct{}{}
			for _, rel := range rels {
				dst := strings.TrimSpace(rel.DstEntity)
				if dst == "" || anaphoricSlotValue(dst) {
					continue
				}
				key := strings.ToLower(dst)
				if _, ok := seenDst[key]; !ok {
					seenDst[key] = struct{}{}
					res.Values = append(res.Values, dst)
				}
				if rel.MemoryID != "" {
					res.MemoryIDs = append(res.MemoryIDs, rel.MemoryID)
				}
				if rec, err := s.store.GetMemory(ctx, tenantID, subjectID, rel.MemoryID); err == nil {
					if strings.Contains(strings.ToLower(rec.Content), key) || !anaphoricSlotValue(rec.Content) {
						res.Contents = append(res.Contents, rec.Content)
					}
				}
			}
			if len(res.Values) > 0 {
				res.Value = strings.Join(res.Values, ", ")
				res.Source = "typed_store"
				return
			}
		}
	}
	s.fetchPredicateHop(ctx, tenantID, subjectID, vertical, includeHistorical, hop, res, topK)
}

func (s *Service) fetchPredicateHop(
	ctx context.Context,
	tenantID, subjectID, vertical string,
	includeHistorical bool,
	hop HopStep,
	res *HopResult,
	topK int,
) {
	entity := firstNonEmpty(res.Entity, hop.Entity)
	pred := firstNonEmpty(hop.Predicate, "")
	res.Entity = entity
	res.Predicate = pred

	// Current-state typed read.
	if pred != "" {
		if cs, ok := s.store.(CurrentStateStore); ok {
			key := statePredicateKey(entity, pred)
			memID, val, _, found, err := cs.GetCurrentState(ctx, tenantID, subjectID, key)
			if err == nil && found {
				res.MemoryIDs = []string{memID}
				res.Value = firstNonEmpty(val, "")
				if rec, err := s.store.GetMemory(ctx, tenantID, subjectID, memID); err == nil {
					res.Contents = append(res.Contents, rec.Content)
					if res.Value == "" {
						res.Value = rec.Content
					}
				}
				res.Source = "typed_store"
				return
			}
			// Unscoped predicate fallback.
			if entity != "" {
				memID, val, _, found, err := cs.GetCurrentState(ctx, tenantID, subjectID, pred)
				if err == nil && found {
					res.MemoryIDs = []string{memID}
					res.Value = firstNonEmpty(val, "")
					if rec, err := s.store.GetMemory(ctx, tenantID, subjectID, memID); err == nil {
						res.Contents = append(res.Contents, rec.Content)
						if res.Value == "" {
							res.Value = rec.Content
						}
					}
					res.Source = "typed_store"
					return
				}
			}
		}
		if indexer, ok := s.store.(AtomIndexer); ok {
			ids, err := indexer.ListAtomMemoryIDs(ctx, tenantID, subjectID, pred, "", topK)
			if err == nil && len(ids) > 0 {
				// Prefer atoms whose content mentions the resolved entity.
				picked := ids
				if entity != "" {
					filtered := make([]string, 0, len(ids))
					for _, id := range ids {
						rec, err := s.store.GetMemory(ctx, tenantID, subjectID, id)
						if err != nil {
							continue
						}
						if strings.Contains(strings.ToLower(rec.Content), strings.ToLower(entity)) ||
							strings.EqualFold(entitySubjectOf(rec), entity) {
							filtered = append(filtered, id)
							res.Contents = append(res.Contents, rec.Content)
						}
					}
					if len(filtered) > 0 {
						picked = filtered
					}
				}
				if len(res.Contents) == 0 {
					for _, id := range picked {
						if rec, err := s.store.GetMemory(ctx, tenantID, subjectID, id); err == nil {
							res.Contents = append(res.Contents, rec.Content)
						}
					}
				}
				res.MemoryIDs = picked
				if len(res.Contents) > 0 {
					seenVal := map[string]struct{}{}
					for _, c := range res.Contents {
						v, ok := slotValueFromMemoryContent(c)
						if !ok {
							v = strings.TrimSpace(c)
						}
						if v == "" || anaphoricSlotValue(v) {
							continue
						}
						key := strings.ToLower(v)
						if _, dup := seenVal[key]; dup {
							continue
						}
						seenVal[key] = struct{}{}
						res.Values = append(res.Values, v)
					}
					if len(res.Values) > 0 {
						res.Value = strings.Join(res.Values, ", ")
					} else {
						res.Value = res.Contents[0]
					}
				}
				res.Source = "typed_store"
				return
			}
		}
	}
	probe := firstNonEmpty(hop.Probe, strings.TrimSpace(entity+" "+pred))
	s.searchFallbackHop(ctx, tenantID, subjectID, vertical, includeHistorical, HopStep{
		Kind: hop.Kind, Entity: entity, Predicate: pred, Probe: probe,
	}, res, topK)
}

func (s *Service) searchFallbackHop(
	ctx context.Context,
	tenantID, subjectID, vertical string,
	includeHistorical bool,
	hop HopStep,
	res *HopResult,
	topK int,
) {
	probe := strings.TrimSpace(hop.Probe)
	if probe == "" {
		probe = strings.TrimSpace(hop.Entity + " " + hop.Predicate)
	}
	if probe == "" {
		return
	}
	search, err := s.SearchOpt(ctx, tenantID, subjectID, vertical, "", probe, SearchOptions{
		IncludeHistorical: includeHistorical,
		Limit:             topK,
	})
	if err != nil || len(search.Results) == 0 {
		return
	}
	res.Source = "search_fallback"
	for _, r := range search.Results {
		if r.MemoryID != "" {
			res.MemoryIDs = append(res.MemoryIDs, r.MemoryID)
		}
		if c := strings.TrimSpace(r.Content); c != "" {
			res.Contents = append(res.Contents, c)
		}
	}
	if len(res.Contents) > 0 && res.Value == "" {
		res.Value = res.Contents[0]
	}
}

// hopJoinProven is true when a fetch/follow hop produced a typed slot value.
// When a resolve hop exists, the fetch entity must be that resolved ID
// (hop[i].output_entity_id == hop[i+1].input_entity_id).
func hopJoinProven(results []HopResult) bool {
	resolved := ""
	var fetched, idJoin bool
	for _, r := range results {
		switch r.Kind {
		case "resolve_entity":
			if r.Source != "unresolved" && strings.TrimSpace(r.Value) != "" {
				resolved = strings.ToLower(strings.TrimSpace(r.Value))
			}
		case "follow_relation", "fetch_predicate", "answer_slot":
			if r.Source == "unresolved" {
				continue
			}
			if strings.TrimSpace(r.Value) == "" && len(r.Values) == 0 && len(r.Contents) == 0 && len(r.MemoryIDs) == 0 {
				continue
			}
			fetched = true
			if resolved != "" && strings.ToLower(strings.TrimSpace(r.Entity)) == resolved {
				idJoin = true
			}
		}
	}
	if resolved != "" {
		return idJoin && fetched
	}
	if fetched {
		for _, r := range results {
			if (r.Kind == "fetch_predicate" || r.Kind == "answer_slot" || r.Kind == "follow_relation") && r.Source == "typed_store" {
				return true
			}
		}
	}
	return false
}

func anaphoricSlotValue(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	if lower == "" {
		return true
	}
	switch {
	case strings.Contains(lower, "home country"), strings.Contains(lower, "my country"),
		lower == "there", lower == "here", lower == "home":
		return true
	}
	return false
}

// hopSlotValues returns typed destinations from answer hops (not resolve).
func hopSlotValues(results []HopResult) []string {
	out := make([]string, 0, 8)
	seen := map[string]struct{}{}
	add := func(v string) {
		v = strings.TrimSpace(v)
		if v == "" || anaphoricSlotValue(v) {
			return
		}
		if utf8Len(v) > 80 {
			return
		}
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}
	for _, r := range results {
		switch r.Kind {
		case "follow_relation", "fetch_predicate", "answer_slot":
			if r.Source == "unresolved" {
				continue
			}
			if len(r.Values) > 0 {
				for _, v := range r.Values {
					add(v)
				}
				continue
			}
			add(r.Value)
		}
	}
	return out
}

func composeFromHopValues(results []HopResult) string {
	vals := hopSlotValues(results)
	if len(vals) == 0 {
		return ""
	}
	titled := make([]string, 0, len(vals))
	for _, v := range vals {
		titled = append(titled, titleCaseWords(v))
	}
	return strings.Join(titled, ", ")
}

func groundToHopValues(answer string, results []HopResult) string {
	vals := hopSlotValues(results)
	if len(vals) == 0 {
		return strings.TrimSpace(answer)
	}
	lower := strings.ToLower(answer)
	missing := false
	for _, v := range vals {
		if !strings.Contains(lower, strings.ToLower(v)) {
			missing = true
			break
		}
	}
	composed := composeFromHopValues(results)
	if missing || strings.TrimSpace(answer) == "" {
		return composed
	}
	return strings.TrimSpace(answer)
}

// bindPacketFromHopResults assigns bridge/direct roles from hop bindings.
// Lexical CoverageTargets are not used for MH coverage when hops exist.
// ContextEvidence (broad search hits) is preserved; hops become ProofChain only.
func bindPacketFromHopResults(pkt *EvidencePacket, hopResults []HopResult, byKey map[string]HopResult) {
	if pkt == nil || len(hopResults) == 0 {
		return
	}
	if len(pkt.ContextEvidence) == 0 {
		pkt.ContextEvidence = append([]string{}, pkt.Contents...)
	}
	items := make([]PacketItem, 0, len(hopResults)*2)
	seen := map[string]struct{}{}
	add := func(role, memID, content, pred string, targets []string) {
		key := memID + "|" + content + "|" + role
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		items = append(items, PacketItem{
			MemoryID:  memID,
			Content:   content,
			Predicate: pred,
			Role:      role,
			Targets:   targets,
			Score:     1,
		})
	}
	for _, r := range hopResults {
		role := "context"
		switch r.Kind {
		case "resolve_entity":
			role = "bridge"
		case "follow_relation", "fetch_predicate", "answer_slot":
			role = "direct"
		}
		targets := []string{r.OutputKey}
		if r.Value != "" && (r.Kind == "follow_relation" || r.Kind == "fetch_predicate" || r.Kind == "answer_slot") {
			add(role, firstNonEmpty(firstID(r.MemoryIDs), ""), r.Value, r.Predicate, targets)
		}
		if len(r.Contents) == 0 && r.Value != "" {
			add(role, firstNonEmpty(firstID(r.MemoryIDs), ""), r.Value, r.Predicate, targets)
			continue
		}
		for i, c := range r.Contents {
			id := ""
			if i < len(r.MemoryIDs) {
				id = r.MemoryIDs[i]
			} else if len(r.MemoryIDs) > 0 {
				id = r.MemoryIDs[0]
			}
			add(role, id, c, r.Predicate, targets)
		}
	}
	if len(items) == 0 {
		return
	}
	pkt.ProofChain = items
	// Keep broad context contents; append proof text that is not already present.
	seenContent := map[string]struct{}{}
	contents := make([]string, 0, len(pkt.ContextEvidence)+len(items))
	for _, c := range pkt.ContextEvidence {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		key := strings.ToLower(c)
		if _, ok := seenContent[key]; ok {
			continue
		}
		seenContent[key] = struct{}{}
		contents = append(contents, c)
	}
	seenID := map[string]struct{}{}
	ids := make([]string, 0, len(pkt.MemoryIDs)+len(items))
	for _, id := range pkt.MemoryIDs {
		if id == "" {
			continue
		}
		if _, ok := seenID[id]; ok {
			continue
		}
		seenID[id] = struct{}{}
		ids = append(ids, id)
	}
	for _, it := range items {
		if c := strings.TrimSpace(it.Content); c != "" {
			key := strings.ToLower(c)
			if _, ok := seenContent[key]; !ok {
				seenContent[key] = struct{}{}
				contents = append(contents, c)
			}
		}
		if it.MemoryID != "" {
			if _, ok := seenID[it.MemoryID]; !ok {
				seenID[it.MemoryID] = struct{}{}
				ids = append(ids, it.MemoryID)
			}
		}
	}
	pkt.Contents = contents
	pkt.MemoryIDs = ids
	merged := make([]PacketItem, 0, len(pkt.Items)+len(items))
	merged = append(merged, pkt.Items...)
	merged = append(merged, items...)
	pkt.Items = merged
	proven := hopJoinProven(hopResults)
	if pkt.Coverage == nil {
		pkt.Coverage = map[string]any{}
	}
	pkt.Coverage["hop_join_proven"] = proven
	pkt.Coverage["hop_results"] = hopResults
	pkt.Coverage["bridge_count"] = countRole(items, "bridge")
	pkt.Coverage["direct_count"] = countRole(items, "direct")
	pkt.Coverage["context_count"] = len(pkt.ContextEvidence)
	pkt.Coverage["proof_count"] = len(pkt.ProofChain)
	if proven {
		pkt.Coverage["uncovered"] = []string{}
	} else {
		unc := make([]string, 0)
		for _, r := range hopResults {
			if r.Source == "unresolved" {
				unc = append(unc, firstNonEmpty(r.OutputKey, r.Kind))
			}
		}
		pkt.Coverage["uncovered"] = unc
	}
	_ = byKey
}

func firstID(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	return ids[0]
}
