package memory

import (
	"context"
	"strings"
)

// HopResult is the typed output of one executed HopStep.
type HopResult struct {
	HopIndex  int      `json:"hop_index"`
	Kind      string   `json:"kind"`
	OutputKey string   `json:"output_key,omitempty"`
	Value     string   `json:"value,omitempty"`
	Values    []string `json:"values,omitempty"` // all typed destinations (enumeration)
	Entity    string   `json:"entity,omitempty"`
	EntityID  string   `json:"entity_id,omitempty"`
	Predicate string   `json:"predicate,omitempty"`
	MemoryIDs []string `json:"memory_ids,omitempty"`
	Contents  []string `json:"contents,omitempty"`
	Source    string   `json:"source,omitempty"`     // typed_store | search_fallback | unresolved
	ProofKind string   `json:"proof_kind,omitempty"` // typed_exact | context
	DependsOn []string `json:"depends_on,omitempty"`
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
		// Wire DependsOn: dest slot (or rewritten kin dest) is the next entity.
		// Source EntityID is kept only when dest is the same person.
		for _, dep := range hop.DependsOn {
			prev, ok := byKey[dep]
			if !ok {
				continue
			}
			applyHopDependency(&res, prev)
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
		s.enrichKinDestSubjectSlots(ctx, tenantID, subjectID, &res)
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
	resolved := ResolveCanonicalEntity(ctx, s.store, tenantID, subjectID, mention)
	if resolved.EntityID != "" {
		res.EntityID = resolved.EntityID
	}
	if strings.TrimSpace(resolved.CanonicalLabel) != "" {
		res.Value = resolved.CanonicalLabel
	} else {
		res.Value = mention
	}
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
			res.ProofKind = "typed_exact"
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
	if res.EntityID == "" && entity != "" {
		res.EntityID = CanonicalEntityID(tenantID, subjectID, entity)
	}
	if indexer, ok := s.store.(RelationIndexer); ok && (entity != "" || res.EntityID != "") {
		src := firstNonEmpty(res.EntityID, strings.ToLower(entity))
		rels, err := indexer.ListRelationsFrom(ctx, tenantID, subjectID, src, pred, topK)
		if err == nil && len(rels) == 0 && src != strings.ToLower(entity) && entity != "" {
			rels, err = indexer.ListRelationsFrom(ctx, tenantID, subjectID, strings.ToLower(entity), pred, topK)
		}
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
				res.ProofKind = "typed_exact"
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
	if res.EntityID == "" && entity != "" {
		res.EntityID = CanonicalEntityID(tenantID, subjectID, entity)
	}

	// Current-state typed read — entity-scoped keys only for typed_exact.
	// Historical / when-event hops need the full atom set, not the latest row.
	if pred != "" {
		if !includeHistorical {
			if cs, ok := s.store.(CurrentStateStore); ok {
				keys := make([]string, 0, 2)
				if res.EntityID != "" {
					keys = append(keys, statePredicateKey(res.EntityID, pred))
				}
				if entity != "" {
					keys = append(keys, statePredicateKey(entity, pred))
				}
				for _, key := range keys {
					memID, val, _, found, err := cs.GetCurrentState(ctx, tenantID, subjectID, key)
					if err != nil || !found {
						continue
					}
					if rec, err := s.store.GetMemory(ctx, tenantID, subjectID, memID); err == nil {
						if !recordMatchesHopEntity(rec, entity, res.EntityID) && entity != "" {
							continue
						}
						res.Contents = append(res.Contents, rec.Content)
						if val == "" {
							val = rec.Content
						}
					}
					res.MemoryIDs = []string{memID}
					res.Value = firstNonEmpty(val, "")
					res.Source = "typed_store"
					res.ProofKind = "typed_exact"
					return
				}
				// Unscoped predicate miss may enrich context, never typed_exact.
				if entity != "" {
					if memID, _, _, found, err := cs.GetCurrentState(ctx, tenantID, subjectID, pred); err == nil && found {
						if rec, err := s.store.GetMemory(ctx, tenantID, subjectID, memID); err == nil {
							res.Contents = append(res.Contents, rec.Content)
						}
						res.MemoryIDs = append(res.MemoryIDs, memID)
						res.ProofKind = "context"
					}
				}
			}
		}
		if indexer, ok := s.store.(AtomIndexer); ok {
			ids, err := indexer.ListAtomMemoryIDs(ctx, tenantID, subjectID, pred, "", topK)
			if err == nil && len(ids) > 0 {
				filtered := make([]string, 0, len(ids))
				filteredContents := make([]string, 0, len(ids))
				filteredNorms := make([]string, 0, len(ids))
				if entity != "" || res.EntityID != "" {
					for _, id := range ids {
						rec, err := s.store.GetMemory(ctx, tenantID, subjectID, id)
						if err != nil {
							continue
						}
						if recordMatchesHopEntity(rec, entity, res.EntityID) {
							filtered = append(filtered, id)
							filteredContents = append(filteredContents, rec.Content)
							filteredNorms = append(filteredNorms, recordValueNorm(rec))
						}
					}
				} else {
					filtered = ids
				}
				if len(filtered) == 0 {
					// Unscoped atom hits stay context-only.
					for _, id := range ids {
						if rec, err := s.store.GetMemory(ctx, tenantID, subjectID, id); err == nil {
							res.Contents = append(res.Contents, rec.Content)
						}
						res.MemoryIDs = append(res.MemoryIDs, id)
					}
					if res.ProofKind == "" {
						res.ProofKind = "context"
					}
				} else {
					res.Contents = append(res.Contents, filteredContents...)
					res.MemoryIDs = filtered
					if len(res.Contents) > 0 {
						seenVal := map[string]struct{}{}
						for i, c := range res.Contents {
							v, ok := slotValueFromMemoryContent(c)
							if !ok || looksTitleCaseSlogan(titleCaseWords(v)) {
								if i < len(filteredNorms) && strings.TrimSpace(filteredNorms[i]) != "" {
									v = filteredNorms[i]
								} else if !ok {
									v = strings.TrimSpace(c)
								}
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
					res.ProofKind = "typed_exact"
					return
				}
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
	res.ProofKind = "context"
	seenVal := map[string]struct{}{}
	for _, r := range search.Results {
		if r.MemoryID != "" {
			res.MemoryIDs = append(res.MemoryIDs, r.MemoryID)
		}
		if c := strings.TrimSpace(r.Content); c != "" {
			res.Contents = append(res.Contents, c)
		}
		if hop.Predicate != "" {
			recPred := searchResultPredicate(r)
			if recPred != "" && !strings.EqualFold(recPred, hop.Predicate) && !hopUsefulForList(recPred, hop.Predicate) {
				continue
			}
		}
		v := structuredValueOf(r)
		if v == "" || anaphoricSlotValue(v) || looksTitleCaseSlogan(v) {
			continue
		}
		if hop.Entity != "" && strings.EqualFold(v, hop.Entity) {
			continue
		}
		key := strings.ToLower(v)
		if _, ok := seenVal[key]; ok {
			continue
		}
		seenVal[key] = struct{}{}
		res.Values = append(res.Values, v)
	}
	if len(res.Values) > 0 && res.Value == "" {
		res.Value = strings.Join(res.Values, ", ")
	}
}

// hopJoinProven is true when a fetch/follow hop produced a typed slot value.
// When a resolve hop exists, the fetch entity must be that resolved ID
// (hop[i].output_entity_id == hop[i+1].input_entity_id).
func hopJoinProven(results []HopResult) bool {
	resolved := ""
	resolvedID := ""
	var fetched, idJoin bool
	for _, r := range results {
		switch r.Kind {
		case "resolve_entity":
			if r.Source == "unresolved" {
				continue
			}
			if strings.TrimSpace(r.EntityID) != "" {
				resolvedID = r.EntityID
			}
			if strings.TrimSpace(r.Value) != "" {
				resolved = strings.ToLower(strings.TrimSpace(r.Value))
			}
		case "follow_relation", "fetch_predicate", "answer_slot":
			if !hopResultTypedExact(r) {
				continue
			}
			if strings.TrimSpace(r.Value) == "" && len(r.Values) == 0 && len(r.Contents) == 0 && len(r.MemoryIDs) == 0 {
				continue
			}
			fetched = true
			if resolvedID != "" && strings.TrimSpace(r.EntityID) == resolvedID {
				idJoin = true
			} else if resolvedID == "" && resolved != "" && strings.ToLower(strings.TrimSpace(r.Entity)) == resolved {
				idJoin = true
			}
		}
	}
	if resolvedID != "" || resolved != "" {
		return idJoin && fetched
	}
	if fetched {
		for _, r := range results {
			if (r.Kind == "fetch_predicate" || r.Kind == "answer_slot" || r.Kind == "follow_relation") && hopResultTypedExact(r) {
				return true
			}
		}
	}
	return false
}

func hopResultTypedExact(r HopResult) bool {
	if r.Source == "unresolved" || r.Source == "search_fallback" {
		return false
	}
	if r.ProofKind == "context" {
		return false
	}
	return r.Source == "typed_store" || r.ProofKind == "typed_exact"
}

func applyHopDependency(res *HopResult, prev HopResult) {
	if res == nil {
		return
	}
	if dest := strings.TrimSpace(kinDestMention(prev)); dest != "" {
		res.Entity = dest
		if strings.EqualFold(dest, strings.TrimSpace(prev.Entity)) {
			if res.EntityID == "" && strings.TrimSpace(prev.EntityID) != "" {
				res.EntityID = prev.EntityID
			}
			return
		}
		// Dest is not the source person (named kin or "Name's mother").
		res.EntityID = ""
		return
	}
	if res.EntityID == "" && strings.TrimSpace(prev.EntityID) != "" {
		res.EntityID = prev.EntityID
	}
}

func kinRoleToken(value string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	v = strings.Trim(v, ".,!?\"'")
	v = strings.TrimPrefix(v, "the ")
	if v == "" {
		return ""
	}
	fields := strings.Fields(v)
	if len(fields) == 2 {
		switch fields[0] {
		case "her", "his", "their", "my", "our":
			v = fields[1]
		}
	}
	for _, role := range kinshipRoles {
		if v == role {
			return role
		}
	}
	return ""
}

func kinRoleLookupForms(role string) []string {
	switch role {
	case "mom", "mother":
		return []string{"mother", "mom"}
	case "dad", "father":
		return []string{"father", "dad"}
	case "grandma", "grandmother":
		return []string{"grandmother", "grandma"}
	case "grandpa", "grandfather":
		return []string{"grandfather", "grandpa"}
	default:
		return []string{role}
	}
}

func kinDestMention(prev HopResult) string {
	raw := strings.TrimSpace(prev.Value)
	role := kinRoleToken(raw)
	if role == "" {
		return raw
	}
	forms := kinRoleLookupForms(role)
	for _, form := range forms {
		for _, c := range prev.Contents {
			if m := possessiveKinMention(c, form); m != "" {
				return m
			}
		}
	}
	src := strings.TrimSpace(prev.Entity)
	if src != "" && kinRoleToken(src) == "" {
		return src + "'s " + forms[0]
	}
	return raw
}

func possessiveKinMention(content, role string) string {
	role = strings.ToLower(strings.TrimSpace(role))
	if role == "" || strings.TrimSpace(content) == "" {
		return ""
	}
	fields := strings.Fields(content)
	for i, f := range fields {
		name := trimPossessiveToken(strings.Trim(f, ".,;:!?\"'"))
		if name == "" || i+1 >= len(fields) {
			continue
		}
		next := strings.Trim(fields[i+1], ".,;:!?\"'")
		if !strings.EqualFold(next, role) {
			continue
		}
		return name + "'s " + role
	}
	return ""
}

func trimPossessiveToken(raw string) string {
	r := []rune(raw)
	if len(r) < 3 {
		return ""
	}
	if r[len(r)-1] != 's' {
		return ""
	}
	prev := r[len(r)-2]
	if prev != '\'' && prev != '’' {
		return ""
	}
	return strings.TrimSpace(string(r[:len(r)-2]))
}

func kinDestSubjectLooks(subj, role string) bool {
	lower := strings.ToLower(strings.TrimSpace(subj))
	role = strings.ToLower(strings.TrimSpace(role))
	if lower == "" || role == "" {
		return false
	}
	return strings.HasSuffix(lower, "'s "+role) || strings.HasSuffix(lower, "’s "+role)
}

func looksKinDestMention(mention string) bool {
	mention = strings.TrimSpace(mention)
	if mention == "" {
		return false
	}
	for _, role := range kinshipRoles {
		for _, form := range kinRoleLookupForms(role) {
			if kinDestSubjectLooks(mention, form) {
				return true
			}
		}
	}
	return false
}

func skipKinDestEnrichPred(pred string) bool {
	switch strings.ToLower(strings.TrimSpace(pred)) {
	case PredicateFamilyMember, PredicateHealth, PredicateRelationshipStatus:
		return true
	}
	return false
}

func attitudeObjectSlot(content string) (string, bool) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", false
	}
	lower := strings.ToLower(content)
	for _, cue := range []string{
		" as one of her hobbies", " as one of his hobbies", " as one of their hobbies",
		" as a hobby", " as her hobby", " as his hobby",
	} {
		i := strings.Index(lower, cue)
		if i <= 0 {
			continue
		}
		head := strings.TrimSpace(content[:i])
		v := afterLastCue(" "+head+" ", []string{" had ", " has "})
		v = strings.Trim(v, ".,;:!?")
		if v != "" && utf8Len(v) >= 3 && utf8Len(v) <= 40 && !anaphoricSlotValue(v) {
			return v, true
		}
	}
	padded := " " + lower + " "
	orig := " " + content + " "
	for _, sep := range []string{
		" passionate about ",
		" interested in ",
		" had a big passion for ",
		" had a passion for ",
		" a passion for ",
	} {
		i := strings.Index(padded, sep)
		if i < 0 {
			continue
		}
		rest := strings.TrimSpace(orig[i+len(sep):])
		rest = strings.Trim(rest, ".,;:!?\"'")
		cut := len(rest)
		restLower := strings.ToLower(rest)
		for _, stop := range []string{",", ";", " and ", " often", " when ", " by ", " that "} {
			if k := strings.Index(restLower, stop); k >= 0 && k < cut {
				cut = k
			}
		}
		v := strings.TrimSpace(rest[:cut])
		v = strings.Trim(v, ".,;:!?\"'")
		if v != "" && utf8Len(v) >= 3 && utf8Len(v) <= 40 && !anaphoricSlotValue(v) {
			return v, true
		}
	}
	return "", false
}

func afterLastCue(s string, cues []string) string {
	low := strings.ToLower(s)
	best := -1
	cueLen := 0
	for _, c := range cues {
		if c == "" {
			continue
		}
		i := strings.LastIndex(low, c)
		if i >= 0 && i >= best {
			best = i
			cueLen = len(c)
		}
	}
	if best < 0 {
		return ""
	}
	return strings.TrimSpace(s[best+cueLen:])
}

func (s *Service) enrichKinDestSubjectSlots(ctx context.Context, tenantID, subjectID string, res *HopResult) {
	if s == nil || s.store == nil || res == nil {
		return
	}
	if res.Kind == "resolve_entity" || skipKinDestEnrichPred(res.Predicate) {
		return
	}
	dest := strings.TrimSpace(res.Entity)
	if !looksKinDestMention(dest) {
		return
	}
	listed, err := s.store.ListMemories(ctx, tenantID, subjectID, false)
	if err != nil || len(listed) == 0 {
		return
	}
	have := map[string]struct{}{}
	for _, v := range res.Values {
		if k := strings.ToLower(strings.TrimSpace(v)); k != "" {
			have[k] = struct{}{}
		}
	}
	added := make([]string, 0, 8)
	add := func(v, memID, content string) {
		v = strings.TrimSpace(v)
		if v == "" || anaphoricSlotValue(v) || looksTitleCaseSlogan(v) || looksTitleCaseSlogan(titleCaseWords(v)) {
			return
		}
		if utf8Len(v) > 80 {
			return
		}
		if strings.EqualFold(v, dest) {
			return
		}
		key := strings.ToLower(v)
		if _, ok := have[key]; ok {
			return
		}
		have[key] = struct{}{}
		added = append(added, v)
		if memID != "" {
			res.MemoryIDs = append(res.MemoryIDs, memID)
		}
		if content != "" {
			res.Contents = append(res.Contents, content)
		}
	}
	for _, rec := range listed {
		if !strings.EqualFold(entitySubjectOf(rec), dest) {
			continue
		}
		pred := ""
		if rec.Metadata != nil {
			if p, ok := rec.Metadata["predicate"].(string); ok {
				pred = p
			}
		}
		if pred == "" && rec.Explain != nil {
			if p, ok := rec.Explain["predicate"].(string); ok {
				pred = p
			}
		}
		if skipKinDestEnrichPred(pred) {
			continue
		}
		att, ok := attitudeObjectSlot(rec.Content)
		if !ok {
			continue
		}
		add(att, rec.MemoryID, rec.Content)
	}
	if len(added) == 0 {
		return
	}
	res.Values = append(added, res.Values...)
	if len(res.Values) > 0 {
		res.Value = strings.Join(res.Values, ", ")
	}
	if res.Source == "" || res.Source == "unresolved" || res.Source == "search_fallback" {
		res.Source = "typed_store"
		res.ProofKind = "typed_exact"
	}
}

func recordMatchesHopEntity(rec MemoryRecord, mention, entityID string) bool {
	if entityID != "" {
		if got := entityIDOf(rec); got != "" {
			return got == entityID
		}
	}
	if mention != "" && strings.EqualFold(entitySubjectOf(rec), mention) {
		return true
	}
	if mention == "" {
		return entityID == ""
	}
	// Bare kin role ("mother" / "her mom") matches dest-subject records,
	// not every source memory that merely mentions the role.
	if role := kinRoleToken(mention); role != "" {
		for _, form := range kinRoleLookupForms(role) {
			if kinDestSubjectLooks(entitySubjectOf(rec), form) {
				return true
			}
			if possessiveKinMention(rec.Content, form) != "" {
				return true
			}
		}
		return false
	}
	return containsEntityMention(rec.Content, mention)
}

func containsEntityMention(content, mention string) bool {
	m := strings.ToLower(strings.TrimSpace(mention))
	c := strings.ToLower(content)
	if m == "" || c == "" {
		return false
	}
	if c == m {
		return true
	}
	padded := " " + c + " "
	needle := " " + m + " "
	return strings.Contains(padded, needle)
}

func recordValueNorm(rec MemoryRecord) string {
	if rec.Metadata != nil {
		if v, ok := rec.Metadata["value_norm"].(string); ok {
			if s := strings.TrimSpace(v); s != "" {
				return s
			}
		}
	}
	if rec.Explain != nil {
		if v, ok := rec.Explain["value_norm"].(string); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
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
	return hopSlotValuesFiltered(results, false)
}

func hopSlotValuesFiltered(results []HopResult, includeFallback bool) []string {
	out := make([]string, 0, 8)
	seen := map[string]struct{}{}
	add := func(v, entity string) {
		v = strings.TrimSpace(v)
		if v == "" || anaphoricSlotValue(v) || looksTitleCaseSlogan(v) || looksTitleCaseSlogan(titleCaseWords(v)) {
			return
		}
		if entity != "" && strings.EqualFold(v, entity) {
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
			if !includeFallback && r.Source == "search_fallback" {
				continue
			}
			if len(r.Values) > 0 {
				for _, v := range r.Values {
					add(v, r.Entity)
				}
				continue
			}
			add(r.Value, r.Entity)
		}
	}
	return out
}

func hopSharedSlotValues(results []HopResult) []string {
	return intersectHopValueGroups(results, hopSlotValues)
}

func hopFetchEntityCount(results []HopResult) int {
	seen := map[string]struct{}{}
	for _, r := range results {
		switch r.Kind {
		case "follow_relation", "fetch_predicate", "answer_slot":
			ent := strings.ToLower(strings.TrimSpace(r.Entity))
			if ent == "" {
				continue
			}
			seen[ent] = struct{}{}
		}
	}
	return len(seen)
}

func intersectHopValueGroups(results []HopResult, valuesOf func([]HopResult) []string) []string {
	groups := map[string][]HopResult{}
	order := make([]string, 0, 2)
	for _, r := range results {
		switch r.Kind {
		case "follow_relation", "fetch_predicate", "answer_slot":
			ent := strings.ToLower(strings.TrimSpace(r.Entity))
			if ent == "" {
				continue
			}
			if _, ok := groups[ent]; !ok {
				order = append(order, ent)
			}
			groups[ent] = append(groups[ent], r)
		}
	}
	if len(order) < 2 {
		return nil
	}
	var inter []string
	for i, ent := range order {
		vals := valuesOf(groups[ent])
		if i == 0 {
			inter = append([]string(nil), vals...)
			continue
		}
		allow := map[string]struct{}{}
		for _, v := range vals {
			allow[strings.ToLower(v)] = struct{}{}
		}
		kept := inter[:0]
		for _, v := range inter {
			if _, ok := allow[strings.ToLower(v)]; ok {
				kept = append(kept, v)
			}
		}
		inter = kept
	}
	return inter
}

func hopEntitySlotGroups(results []HopResult) ([]string, map[string][]string) {
	hopGroups := map[string][]HopResult{}
	order := make([]string, 0, 2)
	for _, r := range results {
		switch r.Kind {
		case "follow_relation", "fetch_predicate", "answer_slot":
			ent := strings.ToLower(strings.TrimSpace(r.Entity))
			if ent == "" {
				continue
			}
			if _, ok := hopGroups[ent]; !ok {
				order = append(order, ent)
			}
			hopGroups[ent] = append(hopGroups[ent], r)
		}
	}
	groups := make(map[string][]string, len(order))
	for _, ent := range order {
		groups[ent] = hopSlotValuesFiltered(hopGroups[ent], true)
	}
	return order, groups
}

func slotValueTokens(v string) []string {
	out := make([]string, 0, 4)
	for _, t := range contentBearingTokens(tokenize(v)) {
		if len(t) < 4 {
			continue
		}
		out = append(out, t)
	}
	return out
}

func slotTokensSubset(small, big []string) bool {
	if len(small) == 0 {
		return false
	}
	for _, s := range small {
		ok := false
		for _, b := range big {
			if tokensMatch(s, b) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

func coveredSlotValue(a, b string) string {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == "" || b == "" {
		return ""
	}
	if strings.EqualFold(a, b) {
		return a
	}
	at, bt := slotValueTokens(a), slotValueTokens(b)
	aSubset := slotTokensSubset(at, bt)
	bSubset := slotTokensSubset(bt, at)
	switch {
	case aSubset && bSubset:
		if utf8Len(a) <= utf8Len(b) {
			return a
		}
		return b
	case aSubset:
		if !joinModifierValue(b) {
			return ""
		}
		return a
	case bSubset:
		if !joinModifierValue(a) {
			return ""
		}
		return b
	}
	return ""
}

func joinModifierValue(s string) bool {
	for _, tok := range tokenize(s) {
		switch strings.ToLower(strings.Trim(tok, "'s")) {
		case "organized", "started", "group":
			return true
		}
	}
	return false
}

// intersectHopValuesByContainment keeps the shorter slot when one entity's
// value is a token subset of the other's (yoga ∩ organized yoga). Exact
// equality is already handled by intersectHopValueGroups.
func intersectHopValuesByContainment(results []HopResult) []string {
	order, groups := hopEntitySlotGroups(results)
	if len(order) < 2 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0)
	add := func(v string) {
		v = strings.TrimSpace(v)
		if v == "" {
			return
		}
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}
	first := groups[order[0]]
	for _, ent := range order[1:] {
		next := groups[ent]
		kept := make([]string, 0)
		for _, va := range first {
			for _, vb := range next {
				if cov := coveredSlotValue(va, vb); cov != "" {
					kept = append(kept, cov)
				}
			}
		}
		first = kept
	}
	for _, v := range first {
		add(v)
	}
	return out
}

// hopValuesMentioningPartner keeps a hop value that names another join
// entity (Casey joined Riley's running group).
func hopValuesMentioningPartner(results []HopResult) []string {
	order, groups := hopEntitySlotGroups(results)
	if len(order) < 2 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0)
	for _, ent := range order {
		for _, other := range order {
			if other == ent {
				continue
			}
			for _, v := range groups[ent] {
				blob := strings.ToLower(v)
				if !strings.Contains(blob, other) {
					continue
				}
				key := strings.ToLower(strings.TrimSpace(v))
				if key == "" {
					continue
				}
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				out = append(out, v)
			}
		}
	}
	return out
}

func hopContentSlotValues(results []HopResult) []string {
	echo := map[string]struct{}{}
	for _, h := range results {
		if v := strings.ToLower(strings.TrimSpace(h.Entity)); v != "" {
			echo[v] = struct{}{}
		}
	}
	seen := map[string]struct{}{}
	vals := make([]string, 0, 4)
	add := func(v string) {
		v = strings.TrimSpace(v)
		if v == "" || anaphoricSlotValue(v) || looksTitleCaseSlogan(v) || utf8Len(v) > 80 {
			return
		}
		if _, ok := echo[strings.ToLower(v)]; ok {
			return
		}
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		vals = append(vals, v)
	}
	for _, h := range results {
		if h.Kind == "resolve_entity" || h.Source == "unresolved" {
			continue
		}
		for _, c := range h.Contents {
			if extracted, ok := slotValueFromMemoryContent(c); ok {
				add(extracted)
			}
		}
	}
	return vals
}

func hopSharedContentValues(results []HopResult) []string {
	return intersectHopValueGroups(results, hopContentSlotValues)
}

func joinTitledHopValues(vals []string) string {
	if len(vals) == 0 {
		return ""
	}
	titled := make([]string, 0, len(vals))
	for _, v := range vals {
		titled = append(titled, titleCaseWords(v))
	}
	return strings.Join(titled, ", ")
}

func composeFromHopValues(results []HopResult) string {
	if hopFetchEntityCount(results) >= 2 {
		vals := hopSharedSlotValues(results)
		if len(vals) == 0 {
			vals = hopSharedContentValues(results)
		}
		return joinTitledHopValues(vals)
	}
	vals := hopSlotValues(results)
	if len(vals) > 6 {
		vals = vals[:6]
	}
	return joinTitledHopValues(vals)
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
		pkt.ContextEvidence = packetItemsFromContents(pkt.Contents, pkt.MemoryIDs)
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
	for _, it := range pkt.ContextEvidence {
		c := strings.TrimSpace(it.Content)
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
