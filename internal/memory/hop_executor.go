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
		case "resolve_entity", "follow_relation":
			s.resolveEntityHop(ctx, tenantID, subjectID, vertical, includeHistorical, hop, &res, topK)
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
					if subj := entitySubjectOf(rec); subj != "" {
						res.Value = subj
					} else {
						res.Value = mention
					}
				} else {
					res.Value = mention
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
					res.Value = res.Contents[0]
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

// hopJoinProven is true when resolve produced an entity/value and fetch produced
// evidence that consumes it (typed or fallback content present).
func hopJoinProven(results []HopResult) bool {
	var resolved, fetched bool
	for _, r := range results {
		switch r.Kind {
		case "resolve_entity", "follow_relation":
			if r.Source != "unresolved" && (r.Value != "" || len(r.MemoryIDs) > 0) {
				resolved = true
			}
		case "fetch_predicate", "answer_slot":
			if r.Source != "unresolved" && (r.Value != "" || len(r.Contents) > 0 || len(r.MemoryIDs) > 0) {
				fetched = true
			}
		}
	}
	if resolved && fetched {
		return true
	}
	// Single fetch hop with typed store hit still counts as grounded join.
	if !resolved && fetched {
		for _, r := range results {
			if (r.Kind == "fetch_predicate" || r.Kind == "answer_slot") && r.Source == "typed_store" {
				return true
			}
		}
	}
	return false
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
		case "resolve_entity", "follow_relation":
			role = "bridge"
		case "fetch_predicate", "answer_slot":
			role = "direct"
		}
		targets := []string{r.OutputKey}
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
