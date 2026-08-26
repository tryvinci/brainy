package memory

import (
	"context"
	"regexp"
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
	query string,
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
	s.recoverSlotAlignedHops(ctx, tenantID, subjectID, query, out)
	for _, res := range out {
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

// hopDumpsUnproven is true when fetch/follow hops exist but none are typed
// exact (search_fallback activity dumps). Those dumps crowd the hybrid
// prompt without proving a join.
func hopDumpsUnproven(results []HopResult) bool {
	n, typed := 0, 0
	for _, r := range results {
		switch r.Kind {
		case "follow_relation", "fetch_predicate", "answer_slot":
			n++
			if hopResultTypedExact(r) {
				typed++
			}
		}
	}
	return n > 0 && typed == 0
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

type recoveredSlot struct {
	value   string
	content string
	memID   string
}

func (s *Service) recoverSlotAlignedHops(ctx context.Context, tenantID, subjectID, query string, hops []HopResult) {
	if s == nil || s.store == nil || len(hops) == 0 || strings.TrimSpace(query) == "" {
		return
	}
	needLoc := looksLocationListQuery(query) && len(practiceObjectTokens(query)) > 0
	needUnwind := looksUnwindQuery(query)
	needPlay := looksInstrumentQuery(query)
	needTrick := looksTrickQuery(query)
	needItems := looksItemTransferQuery(query)
	needBesides := looksBesidesQuery(query) && itemHitsExclusion(query, []string{"stress", "stressor", "stressors"})
	needNames := looksNameListQuery(query)
	needFood := looksFoodSetQuery(query)
	needBeneficiaries := looksBeneficiarySetQuery(query)
	needParticipation := looksParticipationSetQuery(query)
	needTried := looksTriedPolarQuery(query)
	if !needLoc && !needUnwind && !needPlay && !needTrick && !needItems && !needBesides && !needNames && !needFood && !needBeneficiaries && !needParticipation && !needTried {
		return
	}
	listed, err := s.store.ListMemories(ctx, tenantID, subjectID, false)
	if err != nil || len(listed) == 0 {
		return
	}
	person := ""
	if ents := hopQueryEntities(query); len(ents) > 0 {
		person = ents[0]
	}
	if needLoc {
		prependHopSlots(hops, PredicateActivity, recoverPracticeLocationSlots(query, person, hops, listed))
	}
	if needUnwind {
		prependHopSlots(hops, PredicateActivity, recoverUnwindSlots(person, listed))
	}
	if needPlay {
		prependHopSlots(hops, PredicateSkill, recoverPlayObjectSlots(person, listed))
	}
	if needTrick {
		slots := recoverTrickSlots(person, listed)
		if len(possessedBeingMentions(person, listed)) > 0 && len(slots) > 0 {
			replaceHopSlots(hops, PredicateSkill, slots)
		} else {
			prependHopSlots(hops, PredicateSkill, slots)
		}
	}
	if needItems {
		filterHopSlotsByTransferCue(hops, PredicatePossession)
		prependHopSlots(hops, PredicatePossession, recoverItemTransferSlots(person, listed, query))
	}
	if needBesides {
		prependHopSlots(hops, PredicateActivity, recoverBesidesStressorSlots(person, listed))
	}
	if needNames {
		slots := recoverNamedBeingSlots(person, listed)
		if len(slots) >= 2 {
			replaceHopSlots(hops, PredicatePossession, slots)
		}
	}
	if needFood {
		giver := foodSetRecoverPerson(query)
		if giver == "" {
			giver = person
		}
		slots := recoverFoodSetSlots(giver, listed, query)
		if len(slots) >= 2 {
			replaceHopSlotsOn(hops, hopIndexForPredicateEntity(hops, PredicatePreference, giver), slots)
		}
	}
	if needBeneficiaries {
		slots := recoverBeneficiarySlots(person, listed)
		if len(slots) >= 2 {
			idx := hopIndexForPredicateEntity(hops, PredicateAffiliation, person)
			replaceHopSlotsOn(hops, idx, slots)
			if idx >= 0 && idx < len(hops) && len(hops[idx].Values) >= 2 {
				hops[idx].Predicate = PredicateAffiliation
				if person != "" {
					hops[idx].Entity = person
				}
			}
		}
	}
	if needParticipation {
		slots := recoverParticipationSlots(person, listed)
		if len(slots) >= 2 {
			idx := hopIndexForPredicateEntity(hops, PredicateActivity, person)
			replaceHopSlotsOn(hops, idx, slots)
			if idx >= 0 && idx < len(hops) && len(hops[idx].Values) >= 2 {
				hops[idx].Predicate = PredicateActivity
				if person != "" {
					hops[idx].Entity = person
				}
			}
		}
	}
	if needTried {
		slots := recoverTriedExperienceSlots(query, person, listed)
		if len(slots) > 0 {
			pred := PredicatePreference
			if hopIndexForPredicate(hops, PredicatePreference) < 0 {
				pred = PredicateActivity
			}
			prependHopSlots(hops, pred, slots)
			idx := hopIndexForPredicate(hops, pred)
			if idx >= 0 && idx < len(hops) {
				hops[idx].Predicate = pred
				if person != "" {
					hops[idx].Entity = person
				}
			}
		}
	}
}

func prependHopSlots(hops []HopResult, pred string, slots []recoveredSlot) {
	if len(hops) == 0 || len(slots) == 0 {
		return
	}
	idx := hopIndexForPredicate(hops, pred)
	if idx < 0 {
		return
	}
	h := &hops[idx]
	haveVal := map[string]struct{}{}
	for _, v := range h.Values {
		if k := strings.ToLower(strings.TrimSpace(v)); k != "" {
			haveVal[k] = struct{}{}
		}
	}
	haveContent := map[string]struct{}{}
	for _, c := range h.Contents {
		if k := strings.ToLower(strings.TrimSpace(c)); k != "" {
			haveContent[k] = struct{}{}
		}
	}
	var vals, contents, ids []string
	for _, sl := range slots {
		val := strings.TrimSpace(sl.value)
		if val != "" {
			slogan := looksTitleCaseSlogan(val) || looksTitleCaseSlogan(titleCaseWords(val))
			if anaphoricSlotValue(val) || looksCodedSlotValue(val) || utf8Len(val) > 80 {
				val = ""
			} else if slogan && !itemHasTransferCue(val) && !itemHasTransferCue(sl.content) {
				val = ""
			} else {
				key := strings.ToLower(val)
				if _, ok := haveVal[key]; ok {
					val = ""
				} else {
					haveVal[key] = struct{}{}
				}
			}
		}
		content := strings.TrimSpace(sl.content)
		if content != "" {
			ck := strings.ToLower(content)
			if _, ok := haveContent[ck]; ok {
				content = ""
			} else {
				haveContent[ck] = struct{}{}
			}
		}
		if val == "" && content == "" {
			continue
		}
		vals = append(vals, val)
		contents = append(contents, content)
		ids = append(ids, sl.memID)
	}
	if len(vals) == 0 {
		return
	}
	h.Values = append(vals, h.Values...)
	h.Contents = append(contents, h.Contents...)
	h.MemoryIDs = append(ids, h.MemoryIDs...)
	nonempty := make([]string, 0, len(h.Values))
	for _, v := range h.Values {
		if strings.TrimSpace(v) != "" {
			nonempty = append(nonempty, v)
		}
	}
	if len(nonempty) > 0 {
		h.Value = strings.Join(nonempty, ", ")
	}
	if h.Source == "search_fallback" || h.Source == "unresolved" || h.Source == "" {
		h.Source = "typed_store"
		h.ProofKind = "typed_exact"
	}
}

func replaceHopSlots(hops []HopResult, pred string, slots []recoveredSlot) {
	replaceHopSlotsOn(hops, hopIndexForPredicate(hops, pred), slots)
}

func replaceHopSlotsOn(hops []HopResult, idx int, slots []recoveredSlot) {
	if len(hops) == 0 || len(slots) == 0 || idx < 0 || idx >= len(hops) {
		return
	}
	h := &hops[idx]
	haveVal := map[string]struct{}{}
	var vals, contents, ids []string
	for _, sl := range slots {
		val := strings.TrimSpace(sl.value)
		if val == "" || anaphoricSlotValue(val) || looksCodedSlotValue(val) || utf8Len(val) > 80 {
			continue
		}
		if looksTitleCaseSlogan(val) {
			continue
		}
		key := strings.ToLower(val)
		if _, ok := haveVal[key]; ok {
			continue
		}
		haveVal[key] = struct{}{}
		vals = append(vals, val)
		contents = append(contents, strings.TrimSpace(sl.content))
		ids = append(ids, sl.memID)
		if len(vals) >= 8 {
			break
		}
	}
	if len(vals) == 0 {
		return
	}
	h.Values = vals
	h.Contents = contents
	h.MemoryIDs = ids
	h.Value = strings.Join(vals, ", ")
	h.Source = "typed_store"
	h.ProofKind = "typed_exact"
}

func filterHopSlotsByTransferCue(hops []HopResult, pred string) {
	if len(hops) == 0 {
		return
	}
	idx := hopIndexForPredicate(hops, pred)
	if idx < 0 {
		return
	}
	h := &hops[idx]
	var vals, contents, ids []string
	for i, v := range h.Values {
		content := ""
		if i < len(h.Contents) {
			content = h.Contents[i]
		}
		id := ""
		if i < len(h.MemoryIDs) {
			id = h.MemoryIDs[i]
		}
		if looksCodedSlotValue(v) {
			continue
		}
		if !itemHasTransferCue(v) && !itemHasTransferCue(content) {
			continue
		}
		vals = append(vals, v)
		contents = append(contents, content)
		ids = append(ids, id)
	}
	if len(vals) == 0 {
		return
	}
	h.Values = vals
	h.Contents = contents
	h.MemoryIDs = ids
	h.Value = strings.Join(vals, ", ")
}

func hopIndexForPredicateEntity(hops []HopResult, pred, person string) int {
	if pred != "" && strings.TrimSpace(person) != "" {
		for i, h := range hops {
			if h.Kind == "resolve_entity" {
				continue
			}
			if h.Predicate == pred && strings.EqualFold(strings.TrimSpace(h.Entity), person) {
				return i
			}
		}
	}
	return hopIndexForPredicate(hops, pred)
}

func hopIndexForPredicate(hops []HopResult, pred string) int {
	if pred != "" {
		for i, h := range hops {
			if h.Kind == "resolve_entity" {
				continue
			}
			if h.Predicate == pred {
				return i
			}
		}
	}
	for i, h := range hops {
		if h.Kind != "resolve_entity" && h.Source != "unresolved" {
			return i
		}
	}
	for i, h := range hops {
		if h.Kind != "resolve_entity" {
			return i
		}
	}
	return -1
}

func recordMatchesQueryPerson(rec MemoryRecord, person string, hops []HopResult) bool {
	subj := entitySubjectOf(rec)
	if person != "" && strings.EqualFold(subj, person) {
		return true
	}
	if person != "" && looksKinDestMention(subj) {
		low := strings.ToLower(subj)
		p := strings.ToLower(person)
		if strings.HasPrefix(low, p+"'s ") || strings.HasPrefix(low, p+"’s ") {
			return true
		}
	}
	for _, h := range hops {
		dest := strings.TrimSpace(h.Entity)
		if dest != "" && strings.EqualFold(subj, dest) {
			return true
		}
	}
	return false
}

func recoverPracticeLocationSlots(query, person string, hops []HopResult, listed []MemoryRecord) []recoveredSlot {
	focus := practiceObjectTokens(query)
	if len(focus) == 0 {
		return nil
	}
	var out []recoveredSlot
	seen := map[string]struct{}{}
	for _, rec := range listed {
		content := strings.TrimSpace(rec.Content)
		if content == "" || !recordMatchesQueryPerson(rec, person, hops) {
			continue
		}
		if !itemHitsExclusion(content, focus) {
			continue
		}
		if len(placesFromContent(content)) == 0 && compositionalPracticePlace(content, focus) == "" {
			continue
		}
		key := strings.ToLower(content)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, recoveredSlot{content: content, memID: rec.MemoryID})
	}
	return out
}

func recoverUnwindSlots(person string, listed []MemoryRecord) []recoveredSlot {
	if person == "" {
		return nil
	}
	var out []recoveredSlot
	seen := map[string]struct{}{}
	add := func(sl recoveredSlot) {
		key := strings.ToLower(strings.TrimSpace(sl.value + "\n" + sl.content + "\n" + sl.memID))
		if key == "\n\n" {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, sl)
	}
	for _, rec := range listed {
		content := strings.TrimSpace(rec.Content)
		if content == "" || !strings.EqualFold(entitySubjectOf(rec), person) {
			continue
		}
		if !unwindEvidenceHit(content) {
			continue
		}
		add(recoveredSlot{content: content, memID: rec.MemoryID})
		for _, v := range unwindActivitySlots(content) {
			add(recoveredSlot{value: v, content: content, memID: rec.MemoryID})
		}
	}
	return out
}

func recoverPlayObjectSlots(person string, listed []MemoryRecord) []recoveredSlot {
	if person == "" {
		return nil
	}
	var out []recoveredSlot
	seen := map[string]struct{}{}
	add := func(sl recoveredSlot) {
		key := strings.ToLower(strings.TrimSpace(sl.value + "\n" + sl.content + "\n" + sl.memID))
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, sl)
	}
	for _, rec := range listed {
		content := strings.TrimSpace(rec.Content)
		if content == "" || !strings.EqualFold(entitySubjectOf(rec), person) {
			continue
		}
		slots := playPracticeObjectSlots(content)
		if len(slots) == 0 {
			continue
		}
		add(recoveredSlot{content: content, memID: rec.MemoryID})
		for _, v := range slots {
			add(recoveredSlot{value: v, content: content, memID: rec.MemoryID})
		}
	}
	return out
}

func recoverTrickSlots(person string, listed []MemoryRecord) []recoveredSlot {
	if person == "" {
		return nil
	}
	pets := possessedBeingMentions(person, listed)
	var out []recoveredSlot
	seen := map[string]struct{}{}
	add := func(sl recoveredSlot) {
		key := strings.ToLower(strings.TrimSpace(sl.value + "\n" + sl.content + "\n" + sl.memID))
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, sl)
	}
	for _, rec := range listed {
		content := strings.TrimSpace(rec.Content)
		if content == "" {
			continue
		}
		subj := entitySubjectOf(rec)
		trickLine := queryHasToken(content, "trick", "tricks")
		destHit := possessedBeingNamed(subj, pets) || contentMentionsPossessed(content, pets)
		taughtLine := queryHasToken(content, person) &&
			queryHasToken(content, "taught", "teach", "skilled", "perform") &&
			queryHasToken(content, "dog", "dogs", "pet", "pets", "cat", "cats")
		if !trickLine && !destHit && !taughtLine {
			continue
		}
		if trickLine && len(pets) > 0 {
			if !queryHasToken(content, person) && !strings.EqualFold(subj, person) && !destHit {
				continue
			}
		}
		if destHit && !trickLine && !taughtLine {
			if recordPredicateOf(rec) != PredicateSkill && !destCapabilityLine(content) {
				continue
			}
		}
		add(recoveredSlot{content: content, memID: rec.MemoryID})
		for _, v := range trickObjectSlots(content) {
			add(recoveredSlot{value: v, content: content, memID: rec.MemoryID})
		}
		if vn := recordValueNorm(rec); vn != "" && !looksCodedSlotValue(vn) {
			if destCapabilityLine(content) || queryHasToken(content, "trick", "tricks") || len(trickObjectSlots(content)) > 0 {
				if strings.Contains(vn, ",") {
					rest := strings.ReplaceAll(vn, " and ", ",")
					for _, part := range strings.Split(rest, ",") {
						part = strings.TrimSpace(part)
						if part != "" && !looksCodedSlotValue(part) {
							add(recoveredSlot{value: part, content: content, memID: rec.MemoryID})
						}
					}
				} else {
					add(recoveredSlot{value: vn, content: content, memID: rec.MemoryID})
				}
			}
		}
	}
	return out
}

func recordPredicateOf(rec MemoryRecord) string {
	if rec.Metadata != nil {
		if v, ok := rec.Metadata["predicate"].(string); ok {
			return strings.TrimSpace(v)
		}
	}
	if rec.Explain != nil {
		if v, ok := rec.Explain["predicate"].(string); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func possessedBeingMentions(person string, listed []MemoryRecord) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(n string) {
		n = strings.TrimSpace(n)
		n = strings.Trim(n, ".,;:'\"")
		if n == "" || strings.EqualFold(n, person) || !looksHopPerson(n) {
			return
		}
		key := strings.ToLower(n)
		if _, ok := seen[key]; ok {
			return
		}
		if _, stop := hopEntityStop[key]; stop {
			return
		}
		seen[key] = struct{}{}
		out = append(out, n)
	}
	for _, rec := range listed {
		content := strings.TrimSpace(rec.Content)
		if content == "" {
			continue
		}
		subj := entitySubjectOf(rec)
		if !strings.EqualFold(subj, person) && !queryHasToken(content, person) {
			continue
		}
		for _, n := range namedBeings(content + " " + recordValueNorm(rec)) {
			add(n)
		}
		if n := destFromPossessiveClassLine(content, person); n != "" {
			add(n)
		}
	}
	return out
}

func possessedBeingNamed(subj string, pets []string) bool {
	subj = strings.TrimSpace(subj)
	if subj == "" {
		return false
	}
	for _, p := range pets {
		if strings.EqualFold(p, subj) {
			return true
		}
	}
	return false
}

func contentMentionsPossessed(content string, pets []string) bool {
	if content == "" || len(pets) == 0 {
		return false
	}
	for _, p := range pets {
		if queryHasToken(content, p) {
			return true
		}
	}
	return false
}

func namedBeings(value string) []string {
	lower := strings.ToLower(value)
	var out []string
	seen := map[string]struct{}{}
	add := func(name string) {
		name = strings.TrimSpace(strings.Trim(name, ".,;:'\""))
		if name == "" {
			return
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			return
		}
		if !looksHopPerson(name) {
			return
		}
		seen[key] = struct{}{}
		out = append(out, name)
	}
	for _, cue := range []string{" named ", " called "} {
		i := strings.Index(lower, cue)
		if i < 0 {
			continue
		}
		rest := value[i+len(cue):]
		rest = strings.ReplaceAll(rest, " and ", ", ")
		rest = strings.ReplaceAll(rest, " And ", ", ")
		for _, part := range strings.Split(rest, ",") {
			fields := strings.Fields(strings.TrimSpace(part))
			if len(fields) == 0 {
				continue
			}
			add(fields[0])
		}
	}
	return out
}

func possessedClassFollowNames(content string) []string {
	fields := strings.Fields(content)
	classes := map[string]struct{}{
		"dog": {}, "dogs": {}, "pet": {}, "pets": {},
		"cat": {}, "cats": {}, "snake": {}, "snakes": {},
		"puppy": {}, "pup": {},
	}
	var out []string
	seen := map[string]struct{}{}
	add := func(name string) {
		name = strings.TrimSpace(strings.Trim(name, ".,;:'\""))
		if name == "" || !looksHopPerson(name) {
			return
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			return
		}
		if _, stop := hopEntityStop[key]; stop {
			return
		}
		seen[key] = struct{}{}
		out = append(out, name)
	}
	for i, raw := range fields {
		tok := strings.ToLower(strings.Trim(raw, ".,;:'\""))
		if _, ok := classes[tok]; !ok {
			continue
		}
		if i+1 >= len(fields) {
			continue
		}
		next := strings.Trim(fields[i+1], ".,;:'\"")
		if strings.EqualFold(next, "named") || strings.EqualFold(next, "called") {
			rest := fields[i+2:]
			blob := strings.Join(rest, " ")
			if j := strings.IndexAny(blob, ".!?"); j >= 0 {
				blob = blob[:j]
			}
			blob = strings.ReplaceAll(blob, " and ", ", ")
			for _, part := range strings.Split(blob, ",") {
				toks := strings.Fields(strings.TrimSpace(part))
				if len(toks) == 0 {
					continue
				}
				add(toks[0])
			}
			continue
		}
		add(next)
	}
	return out
}

func destFromPossessiveClassLine(content, person string) string {
	lower := strings.ToLower(content)
	p := strings.ToLower(strings.TrimSpace(person))
	if p == "" {
		return ""
	}
	for _, class := range []string{"dog", "pet", "cat", "puppy", "pup"} {
		for _, cue := range []string{" is " + p + "'s " + class, " is " + p + "’s " + class} {
			i := strings.Index(lower, cue)
			if i < 3 {
				continue
			}
			head := strings.TrimSpace(content[:i])
			fields := strings.Fields(head)
			if len(fields) == 0 {
				continue
			}
			name := strings.Trim(fields[len(fields)-1], ".,")
			if looksHopPerson(name) {
				return name
			}
		}
	}
	return ""
}

func destCapabilityLine(content string) bool {
	if queryHasToken(content, "trick", "tricks", "taught", "teach", "skilled", "perform") {
		return true
	}
	return queryHasToken(content, "swim", "swimming", "catch", "catching", "balance", "skateboard")
}

func recoverItemTransferSlots(person string, listed []MemoryRecord, query string) []recoveredSlot {
	if person == "" {
		return nil
	}
	classToks := forClauseTokens(query)
	var out []recoveredSlot
	seen := map[string]struct{}{}
	add := func(sl recoveredSlot) {
		key := strings.ToLower(strings.TrimSpace(sl.value + "\n" + sl.content + "\n" + sl.memID))
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, sl)
	}
	for _, rec := range listed {
		content := strings.TrimSpace(rec.Content)
		if content == "" || !itemHasTransferCue(content) {
			continue
		}
		subj := entitySubjectOf(rec)
		if !strings.EqualFold(subj, person) && !queryHasToken(content, person) {
			continue
		}
		if len(classToks) > 0 && !itemHitsExclusion(content, classToks) {
			continue
		}
		add(recoveredSlot{content: content, memID: rec.MemoryID})
		if obj := transferObjectFromContent(content); obj != "" {
			add(recoveredSlot{value: obj, content: content, memID: rec.MemoryID})
		}
		if vn := recordValueNorm(rec); vn != "" && !looksCodedSlotValue(vn) {
			add(recoveredSlot{value: vn, content: content, memID: rec.MemoryID})
		}
	}
	return out
}

func transferObjectFromContent(content string) string {
	lower := strings.ToLower(content)
	cues := []string{" to buy ", " bought ", " buy ", " acquired ", " purchased ", " got ", " made "}
	for _, cue := range cues {
		i := strings.Index(lower, cue)
		if i < 0 {
			continue
		}
		rest := strings.TrimSpace(content[i+len(cue):])
		if j := strings.Index(strings.ToLower(rest), " for "); j >= 0 {
			rest = strings.TrimSpace(rest[:j])
		}
		if k := strings.IndexAny(rest, ".!?"); k >= 0 {
			rest = strings.TrimSpace(rest[:k])
		}
		rest = strings.TrimSpace(rest)
		rest = strings.TrimPrefix(rest, "a ")
		rest = strings.TrimPrefix(rest, "an ")
		rest = strings.TrimPrefix(rest, "the ")
		if rest == "" || utf8Len(rest) < 3 || utf8Len(rest) > 40 {
			continue
		}
		low := strings.ToLower(rest)
		if strings.Contains(low, "store") || strings.Contains(low, "shop") || anaphoricSlotValue(rest) {
			continue
		}
		return rest
	}
	return ""
}

func looksCodedSlotValue(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" || strings.Contains(v, " ") {
		return false
	}
	return strings.Contains(v, "_")
}

func looksNameListQuery(query string) bool {
	if !queryHasToken(query, "name", "names") {
		return false
	}
	return wantsHistoricalAtomScan(query)
}

// looksFoodSetQuery is a typed meal/suggestion set ("what kind of meals",
// "what kind of food suggestions"). Bare "what kind of food" leftover and
// un- snack leftover stay on covering.
func foodSetRecoverPerson(query string) string {
	ents := hopQueryEntities(query)
	if recip := transferRecipient(query); recip != "" {
		for _, e := range ents {
			if !strings.EqualFold(e, recip) {
				return e
			}
		}
	}
	if len(ents) > 0 {
		return ents[0]
	}
	return ""
}

func looksFoodSetQuery(query string) bool {
	if !looksWhatKindQuery(query) {
		return false
	}
	return queryHasToken(query, "meals", "meal", "suggestions", "suggestion")
}

func looksBeneficiarySetQuery(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return false
	}
	if !queryHasToken(query, "beneficiaries", "beneficiary") {
		return false
	}
	return strings.HasPrefix(q, "who") || strings.HasPrefix(q, "which") || strings.HasPrefix(q, "what")
}

func looksParticipationSetQuery(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" || !strings.Contains(q, "community") {
		return false
	}
	return strings.Contains(q, "ways") || strings.HasPrefix(q, "in what")
}

func recoverTriedExperienceSlots(query, person string, listed []MemoryRecord) []recoveredSlot {
	if person == "" {
		return nil
	}
	claim := polarClaimTokens(query)
	if len(claim) == 0 {
		return nil
	}
	var out []recoveredSlot
	seen := map[string]struct{}{}
	for _, rec := range listed {
		content := strings.TrimSpace(rec.Content)
		if content == "" || strings.HasPrefix(content, "[") {
			continue
		}
		if !strings.EqualFold(entitySubjectOf(rec), person) && !queryHasToken(content, person) &&
			!strings.HasPrefix(strings.ToLower(content), strings.ToLower(person)+":") {
			continue
		}
		if !polarPieceHasClaim(content, claim) || !polarExperienceCue(content) {
			continue
		}
		val := polarTriedClaimValue(content, rec, claim)
		if val == "" || anaphoricSlotValue(val) || looksCodedSlotValue(val) {
			continue
		}
		key := strings.ToLower(val)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, recoveredSlot{value: val, content: content, memID: rec.MemoryID})
		if len(out) >= 4 {
			break
		}
	}
	return out
}

func polarTriedClaimValue(content string, rec MemoryRecord, claim []string) string {
	if v := strings.ToLower(strings.TrimSpace(recordValueNorm(rec))); v != "" {
		for _, t := range claim {
			if t != "" && strings.Contains(v, t) {
				return t
			}
		}
	}
	lower := strings.ToLower(content)
	for _, t := range claim {
		if t != "" && strings.Contains(lower, t) {
			return t
		}
	}
	return ""
}

func looksThinParticipation(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return true
	}
	switch v {
	case "community", "rights", "difference", "painting", "sidewalk", "month",
		"city", "neighborhood", "students", "folks", "adoption", "freedom",
		"pride", "work", "people", "group", "events", "campaigns", "meetings",
		"back", "those", "chasing", "not-so-great", "not so great":
		return true
	}
	if strings.HasPrefix(v, "community ") && !strings.Contains(v, "garden") {
		return true
	}
	return false
}

func recoverParticipationSlots(person string, listed []MemoryRecord) []recoveredSlot {
	if person == "" {
		return nil
	}
	var primary, attend []recoveredSlot
	seen := map[string]struct{}{}
	add := func(dst *[]recoveredSlot, sl recoveredSlot) {
		val := strings.TrimSpace(sl.value)
		val = strings.Trim(val, ".,;: ")
		if val == "" || anaphoricSlotValue(val) || looksCodedSlotValue(val) || utf8Len(val) > 48 || looksThinParticipation(val) {
			return
		}
		key := strings.ToLower(val)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		*dst = append(*dst, sl)
	}
	for _, rec := range listed {
		content := strings.TrimSpace(rec.Content)
		if content == "" || strings.HasPrefix(content, "[") {
			continue
		}
		if !strings.EqualFold(entitySubjectOf(rec), person) && !queryHasToken(content, person) &&
			!strings.HasPrefix(strings.ToLower(content), strings.ToLower(person)+":") {
			continue
		}
		for _, v := range participationObjectsFromCues(content, participationPrimaryCues) {
			add(&primary, recoveredSlot{value: v, content: content, memID: rec.MemoryID})
		}
		for _, v := range participationObjectsFromCues(content, participationAttendCues) {
			add(&attend, recoveredSlot{value: v, content: content, memID: rec.MemoryID})
		}
	}
	out := append(primary, attend...)
	if len(out) > 8 {
		out = out[:8]
	}
	return out
}

var participationPrimaryCues = []string{
	" joined a ", " joined an ", " joined the ", " joined ",
	" organizing an ", " organizing a ", " host an ", " host a ", " hosting an ", " hosting a ",
	" participating in a ", " participating in an ", " participating in the ",
	" participates in a ", " participates in an ", " participates in the ",
	"mentorship program",
}

var participationAttendCues = []string{
	" attended a ", " attended an ", " attended the ", " attended ",
}

func participationObjectsFromContent(content string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, v := range append(participationObjectsFromCues(content, participationPrimaryCues),
		participationObjectsFromCues(content, participationAttendCues)...) {
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}
	return out
}

func participationObjectsFromCues(content string, cues []string) []string {
	lower := strings.ToLower(content)
	var out []string
	seen := map[string]struct{}{}
	add := func(v string) {
		v = strings.TrimSpace(v)
		v = strings.Trim(v, ".,;: ")
		if j := strings.IndexAny(v, ".!?"); j >= 0 {
			v = strings.TrimSpace(v[:j])
		}
		for _, tail := range []string{
			" which ", " during ", " on ", " last ", " after ", " and we ", " and ",
			" for ", " since ", " scheduled ", " featuring ", " in ",
		} {
			if k := strings.Index(strings.ToLower(v), tail); k >= 3 {
				v = strings.TrimSpace(v[:k])
			}
		}
		if low := strings.ToLower(v); strings.HasSuffix(low, " scheduled") {
			v = strings.TrimSpace(v[:len(v)-len(" scheduled")])
		}
		v = strings.TrimPrefix(v, "a ")
		v = strings.TrimPrefix(v, "an ")
		v = strings.TrimPrefix(v, "the ")
		v = strings.TrimPrefix(v, "new ")
		if v == "" || utf8Len(v) < 5 || utf8Len(v) > 48 || looksThinParticipation(v) {
			return
		}
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}
	for _, cue := range cues {
		start := 0
		for {
			i := strings.Index(lower[start:], cue)
			if i < 0 {
				break
			}
			i += start
			if cue == "mentorship program" {
				add("mentorship program")
				start = i + len(cue)
				continue
			}
			rest := strings.TrimSpace(content[i+len(cue):])
			add(rest)
			start = i + len(cue)
		}
	}
	return out
}

func looksThinBeneficiary(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return true
	}
	switch v {
	case "charity", "fun", "game", "cs:go", "csgo", "kids", "money", "leftover money",
		"the game", "the tournament", "charity events", "charity event":
		return true
	}
	if strings.Contains(v, "cs:go") || strings.Contains(v, "csgo") || strings.Contains(v, "tournament") {
		return true
	}
	if strings.HasPrefix(v, "the game") || strings.HasPrefix(v, "game ") || strings.HasPrefix(v, "charity") {
		return true
	}
	return false
}

func recoverBeneficiarySlots(person string, listed []MemoryRecord) []recoveredSlot {
	if person == "" {
		return nil
	}
	var out []recoveredSlot
	seen := map[string]struct{}{}
	add := func(sl recoveredSlot) {
		val := strings.TrimSpace(sl.value)
		val = strings.Trim(val, ".,;: ")
		if val == "" || anaphoricSlotValue(val) || looksCodedSlotValue(val) || utf8Len(val) > 40 || looksThinBeneficiary(val) {
			return
		}
		key := strings.ToLower(val)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, sl)
	}
	for _, rec := range listed {
		content := strings.TrimSpace(rec.Content)
		if content == "" || strings.HasPrefix(content, "[") {
			continue
		}
		if !strings.EqualFold(entitySubjectOf(rec), person) && !queryHasToken(content, person) &&
			!strings.HasPrefix(strings.ToLower(content), strings.ToLower(person)+":") {
			continue
		}
		body := " " + strings.ToLower(content) + " "
		if !strings.Contains(body, " charity ") && !strings.Contains(body, " raise ") &&
			!strings.Contains(body, " raised ") && !strings.Contains(body, " leftover money ") &&
			!strings.Contains(body, " shelter ") && !strings.Contains(body, " homeless ") &&
			!strings.Contains(body, " hospital ") {
			continue
		}
		for _, v := range beneficiaryObjectsFromContent(content) {
			add(recoveredSlot{value: v, content: content, memID: rec.MemoryID})
		}
	}
	if len(out) > 6 {
		out = out[:6]
	}
	return out
}

func beneficiaryObjectsFromContent(content string) []string {
	lower := strings.ToLower(content)
	var out []string
	seen := map[string]struct{}{}
	add := func(v string) {
		v = strings.TrimSpace(v)
		v = strings.Trim(v, ".,;: ")
		if j := strings.IndexAny(v, ".!?"); j >= 0 {
			v = strings.TrimSpace(v[:j])
		}
		for _, tail := range []string{" which ", " during ", " on ", " last ", " after ", " and we "} {
			if k := strings.Index(strings.ToLower(v), tail); k >= 3 {
				v = strings.TrimSpace(v[:k])
			}
		}
		v = strings.TrimPrefix(v, "a ")
		v = strings.TrimPrefix(v, "an ")
		v = strings.TrimPrefix(v, "the ")
		if v == "" || utf8Len(v) < 4 || utf8Len(v) > 40 || looksThinBeneficiary(v) {
			return
		}
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}
	for _, cue := range []string{" for a ", " for an ", " for the ", " for "} {
		start := 0
		for {
			i := strings.Index(lower[start:], cue)
			if i < 0 {
				break
			}
			i += start
			rest := strings.TrimSpace(content[i+len(cue):])
			add(rest)
			start = i + len(cue)
		}
	}
	return out
}

func itemHasSuggestionCue(s string) bool {
	body := " " + strings.ToLower(s) + " "
	for _, cue := range []string{
		" suggest ", " suggested ", " suggests ", " suggestion ",
		" recommend ", " recommended ", " recommends ",
		" how about ", " swap ", " replace ", " instead of ",
		" given to ", " gave to ",
		" check out ", " recipe for ", " prepared ",
	} {
		if strings.Contains(body, cue) {
			return true
		}
	}
	return false
}

func itemHasMealCue(s string) bool {
	body := " " + strings.ToLower(s) + " "
	for _, cue := range []string{
		" prepared ", " started eating ", " recipe is ", " recipe for ",
		" made this ", " made a ",
	} {
		if strings.Contains(body, cue) {
			return true
		}
	}
	return false
}

func looksThinFoodObject(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return false
	}
	if strings.HasPrefix(v, "started ") && !strings.Contains(v, "eating") {
		return true
	}
	if strings.Contains(v, " ") {
		return false
	}
	switch v {
	case "healthier", "healthy", "yummy", "tasty", "good", "great",
		"recipes", "recipe", "meals", "meal", "snacks", "snack", "food":
		return true
	}
	return false
}

func recoverFoodSetSlots(person string, listed []MemoryRecord, query string) []recoveredSlot {
	if person == "" {
		return nil
	}
	wantSuggest := queryHasToken(query, "suggestions", "suggestion")
	wantMeal := queryHasToken(query, "meals", "meal")
	var out []recoveredSlot
	seen := map[string]struct{}{}
	add := func(sl recoveredSlot) {
		val := strings.TrimSpace(sl.value)
		if val == "" || anaphoricSlotValue(val) || looksCodedSlotValue(val) || utf8Len(val) > 60 || looksThinFoodObject(val) {
			return
		}
		key := strings.ToLower(val)
		if _, ok := seen[key]; ok {
			return
		}
		for i, existing := range out {
			ex := strings.ToLower(strings.TrimSpace(existing.value))
			if utf8Len(ex) < 4 || utf8Len(key) < 4 {
				continue
			}
			if !strings.Contains(ex, key) && !strings.Contains(key, ex) {
				continue
			}
			if utf8Len(key) <= utf8Len(ex) {
				return
			}
			delete(seen, ex)
			out[i] = sl
			seen[key] = struct{}{}
			return
		}
		seen[key] = struct{}{}
		out = append(out, sl)
	}
	for _, rec := range listed {
		content := strings.TrimSpace(rec.Content)
		if content == "" || strings.HasPrefix(content, "[") {
			continue
		}
		if !strings.EqualFold(entitySubjectOf(rec), person) && !queryHasToken(content, person) &&
			!strings.HasPrefix(strings.ToLower(content), strings.ToLower(person)+":") {
			continue
		}
		cueOK := (wantSuggest && (itemHasSuggestionCue(content) || itemHasMealCue(content))) ||
			(wantMeal && itemHasMealCue(content))
		if !cueOK {
			continue
		}
		for _, v := range foodObjectsFromContent(content) {
			add(recoveredSlot{value: v, content: content, memID: rec.MemoryID})
		}
		if len(foodObjectsFromContent(content)) > 0 {
			continue
		}
		if vn := recordValueNorm(rec); vn != "" && !looksCodedSlotValue(vn) && !strings.Contains(vn, ",") && !looksThinFoodObject(vn) {
			add(recoveredSlot{value: vn, content: content, memID: rec.MemoryID})
		}
	}
	if len(out) > 6 {
		out = out[:6]
	}
	return out
}

func foodObjectsFromContent(content string) []string {
	lower := strings.ToLower(content)
	takeAfter := func(cue string, splitList bool) []string {
		i := strings.Index(lower, cue)
		if i < 0 {
			return nil
		}
		rest := strings.TrimSpace(content[i+len(cue):])
		rest = trimFoodObjectTail(rest)
		rest = clipFoodObjectRest(rest)
		if rest == "" {
			return nil
		}
		if splitList {
			return splitFoodListParts(rest)
		}
		if utf8Len(rest) < 3 || utf8Len(rest) > 60 || anaphoricSlotValue(rest) {
			return nil
		}
		return []string{rest}
	}
	if parts := takeAfter(" how about some ", true); len(parts) > 0 {
		return parts
	}
	if parts := takeAfter(" how about ", true); len(parts) > 0 {
		return parts
	}
	if i := strings.Index(lower, " swap "); i >= 0 {
		rest := lower[i:]
		if j := strings.Index(rest, " for "); j >= 0 {
			got := strings.TrimSpace(content[i+j+len(" for "):])
			got = trimFoodObjectTail(got)
			if got != "" && utf8Len(got) >= 3 && utf8Len(got) <= 60 {
				return []string{got}
			}
		}
	}
	if i := strings.Index(lower, " replace "); i >= 0 {
		rest := lower[i:]
		if j := strings.Index(rest, " with "); j >= 0 {
			got := strings.TrimSpace(content[i+j+len(" with "):])
			got = trimFoodObjectTail(got)
			if got != "" && utf8Len(got) >= 3 && utf8Len(got) <= 60 {
				return []string{got}
			}
		}
	}
	for _, cue := range []string{
		" prepared a ", " prepared ",
		" started eating ",
		" recipe is ", " recipe for these ", " recipe for ",
		" suggested ", " recommends that ",
	} {
		if parts := takeAfter(cue, false); len(parts) > 0 {
			return parts
		}
	}
	if strings.Contains(lower, "recipe") {
		if parts := takeAfter(" for these ", false); len(parts) > 0 {
			return parts
		}
	}
	return nil
}

func trimFoodObjectTail(rest string) string {
	rest = strings.TrimSpace(rest)
	if j := strings.IndexAny(rest, ".!?"); j >= 0 {
		rest = strings.TrimSpace(rest[:j])
	}
	low := strings.ToLower(rest)
	for _, tail := range []string{" on ", " as a ", " for a ", " that has ", " after "} {
		if k := strings.Index(low, tail); k >= 8 {
			rest = strings.TrimSpace(rest[:k])
			low = strings.ToLower(rest)
		}
	}
	if k := strings.LastIndex(low, " to "); k >= 3 {
		tail := strings.TrimSpace(rest[k+4:])
		if tail != "" && !strings.Contains(tail, " ") {
			rest = strings.TrimSpace(rest[:k])
			low = strings.ToLower(rest)
		}
	}
	rest = strings.TrimSpace(rest)
	rest = strings.TrimPrefix(rest, "a ")
	rest = strings.TrimPrefix(rest, "an ")
	rest = strings.TrimPrefix(rest, "the ")
	rest = strings.TrimPrefix(rest, "these ")
	rest = strings.TrimPrefix(rest, "some ")
	return strings.TrimSpace(rest)
}

func clipFoodObjectRest(rest string) string {
	rest = strings.TrimSpace(rest)
	if rest == "" || utf8Len(rest) <= 60 {
		return rest
	}
	parts := strings.Split(rest, ",")
	acc := strings.TrimSpace(parts[0])
	for _, p := range parts[1:] {
		next := acc + ", " + strings.TrimSpace(p)
		if utf8Len(next) > 60 {
			break
		}
		acc = next
		if strings.Count(acc, ",") >= 1 {
			break
		}
	}
	return acc
}

func splitFoodListParts(rest string) []string {
	body := strings.ToLower(rest)
	body = strings.ReplaceAll(body, " with some ", ",")
	body = strings.ReplaceAll(body, " or ", ",")
	body = strings.ReplaceAll(body, " and ", ",")
	out := make([]string, 0, 4)
	seen := map[string]struct{}{}
	for _, part := range strings.Split(body, ",") {
		part = strings.TrimSpace(part)
		part = strings.TrimPrefix(part, "a ")
		part = strings.TrimPrefix(part, "an ")
		part = strings.TrimPrefix(part, "the ")
		part = strings.TrimPrefix(part, "some ")
		part = strings.TrimSpace(part)
		if part == "" || utf8Len(part) < 3 || utf8Len(part) > 40 || anaphoricSlotValue(part) {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		out = append(out, part)
	}
	return out
}

func recoverNamedBeingSlots(person string, listed []MemoryRecord) []recoveredSlot {
	if person == "" {
		return nil
	}
	var out []recoveredSlot
	seen := map[string]struct{}{}
	add := func(sl recoveredSlot) {
		val := strings.TrimSpace(sl.value)
		if val == "" || strings.EqualFold(val, person) || !looksHopPerson(val) {
			return
		}
		key := strings.ToLower(val)
		if _, ok := seen[key]; ok {
			return
		}
		if _, stop := hopEntityStop[key]; stop {
			return
		}
		seen[key] = struct{}{}
		out = append(out, sl)
	}
	for _, rec := range listed {
		content := strings.TrimSpace(rec.Content)
		if content == "" {
			continue
		}
		if !strings.EqualFold(entitySubjectOf(rec), person) && !queryHasToken(content, person) {
			continue
		}
		if n := destFromPossessiveClassLine(content, person); n != "" {
			add(recoveredSlot{value: n, content: content, memID: rec.MemoryID})
		}
		for _, n := range possessedClassFollowNames(content) {
			add(recoveredSlot{value: n, content: content, memID: rec.MemoryID})
		}
	}
	return out
}

func recoverBesidesStressorSlots(person string, listed []MemoryRecord) []recoveredSlot {
	if person == "" {
		return nil
	}
	var out []recoveredSlot
	seen := map[string]struct{}{}
	add := func(sl recoveredSlot) {
		key := strings.ToLower(strings.TrimSpace(sl.value + "\n" + sl.content + "\n" + sl.memID))
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, sl)
	}
	for _, rec := range listed {
		content := strings.TrimSpace(rec.Content)
		if content == "" || !strings.EqualFold(entitySubjectOf(rec), person) {
			continue
		}
		lower := strings.ToLower(content)
		if !strings.Contains(lower, "stress") && !itemHitsExclusion(content, []string{"stress", "stressed"}) {
			continue
		}
		if itemHitsExclusion(content, []string{"hike", "hiking"}) {
			continue
		}
		add(recoveredSlot{content: content, memID: rec.MemoryID})
		if strings.Contains(lower, "work") {
			add(recoveredSlot{value: "work", content: content, memID: rec.MemoryID})
		}
	}
	return out
}

var (
	playObjectRE     = regexp.MustCompile(`(?i)\bplays(?:\s+the)?\s+([a-z][a-z-]+)\b`)
	playingRE        = regexp.MustCompile(`(?i)\bplaying(?:\s+the)?\s+([a-z][a-z-]+)\b`)
	practiceRE       = regexp.MustCompile(`(?i)\b([a-z][a-z-]+)\s+practice\b`)
	unwindToRE       = regexp.MustCompile(`(?i)\b([a-z][a-z-]+)\s+to\s+(?:[a-z-]*stress|unwind|relax|calm)\b`)
	unwindToStressRE = regexp.MustCompile(`(?i)\bto\s+[a-z-]*stress\b`)
	makingRE         = regexp.MustCompile(`(?i)\b(?:making|practicing|practising)\s+([a-z][a-z-]+)\b`)
	tricksLikeRE     = regexp.MustCompile(`(?i)tricks(?:\s+\w+){0,3}\s+(?:like|such as)\s+(.+)`)
)

func unwindEvidenceTokens() []string {
	return []string{"unwind", "unwinds", "relax", "relaxes", "calming", "calm"}
}

func unwindEvidenceHit(s string) bool {
	if itemHitsExclusion(s, unwindEvidenceTokens()) {
		return true
	}
	return unwindToStressRE.MatchString(s)
}

func playPracticeObjectSlots(content string) []string {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	var out []string
	seen := map[string]struct{}{}
	add := func(v string) {
		v = strings.TrimSpace(strings.ToLower(v))
		if v == "" || isQueryStopword(v) || v == "daily" || v == "practice" || v == "practicing" {
			return
		}
		if utf8Len(v) < 3 || utf8Len(v) > 40 {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	for _, re := range []*regexp.Regexp{playObjectRE, playingRE, practiceRE} {
		for _, m := range re.FindAllStringSubmatch(content, -1) {
			if len(m) >= 2 {
				add(m[1])
			}
		}
	}
	return out
}

func unwindActivitySlots(content string) []string {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	var out []string
	seen := map[string]struct{}{}
	add := func(v string) {
		v = strings.TrimSpace(v)
		if v == "" || anaphoricSlotValue(v) || utf8Len(v) < 3 || utf8Len(v) > 40 {
			return
		}
		key := strings.ToLower(v)
		if isQueryStopword(key) {
			return
		}
		switch key {
		case "way", "ways", "great", "good", "best", "thing", "things", "lot", "lots", "farther", "further":
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}
	low := strings.ToLower(content)
	if strings.Contains(low, "unwinds via") || strings.Contains(low, "enjoys practicing") || strings.Contains(low, "enjoys practising") {
		if v, ok := slotValueFromMemoryContent(content); ok {
			add(v)
		}
	}
	if v, ok := attitudeObjectSlot(content); ok {
		add(v)
	}
	for _, m := range unwindToRE.FindAllStringSubmatch(content, -1) {
		if len(m) >= 2 {
			add(m[1])
		}
	}
	for _, m := range makingRE.FindAllStringSubmatch(content, -1) {
		if len(m) >= 2 {
			add(m[1])
		}
	}
	return out
}

func trickObjectSlots(content string) []string {
	content = strings.TrimSpace(content)
	if content == "" || !queryHasToken(content, "trick", "tricks") {
		return nil
	}
	var out []string
	seen := map[string]struct{}{}
	add := func(v string) {
		v = strings.TrimSpace(strings.Trim(v, " .,"))
		v = strings.TrimPrefix(strings.ToLower(v), "like ")
		v = strings.TrimPrefix(v, "such as ")
		v = strings.TrimSpace(v)
		if v == "" || utf8Len(v) < 3 || utf8Len(v) > 40 || anaphoricSlotValue(v) {
			return
		}
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}
	if m := tricksLikeRE.FindStringSubmatch(content); len(m) == 2 {
		rest := m[1]
		if j := strings.IndexAny(rest, ".!?"); j >= 0 {
			rest = rest[:j]
		}
		rest = strings.ReplaceAll(rest, " and ", ",")
		for _, part := range strings.Split(rest, ",") {
			add(part)
		}
	}
	lower := strings.ToLower(content)
	if i := strings.Index(lower, "tricks "); i >= 0 {
		rest := content[i+len("tricks "):]
		if j := strings.IndexAny(rest, ".!?"); j >= 0 {
			rest = rest[:j]
		}
		if strings.Contains(rest, ",") {
			rest = strings.ReplaceAll(rest, " and ", ",")
			for _, part := range strings.Split(rest, ",") {
				add(part)
			}
		}
	}
	return out
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
		if hopValueIsAttendedEvent(v) || hopValueHasForeignPossessive(v, entity, results) {
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

func hopValueIsAttendedEvent(v string) bool {
	lower := strings.ToLower(strings.TrimSpace(v))
	return strings.HasPrefix(lower, "attended ")
}

func hopValueHasForeignPossessive(v, entity string, hops []HopResult) bool {
	lower := strings.ToLower(strings.TrimSpace(v))
	if lower == "" {
		return false
	}
	// Visiting / checking out someone else's place is a destination, not a
	// stolen possession. Keep those slots so plan answers can name the stop.
	if hopValueIsCoParticipantVisit(lower) {
		return false
	}
	self := strings.ToLower(strings.TrimSpace(entity))
	seen := map[string]struct{}{}
	for _, h := range hops {
		other := strings.ToLower(strings.TrimSpace(h.Entity))
		if other == "" || other == self || utf8Len(other) < 3 {
			continue
		}
		if _, ok := seen[other]; ok {
			continue
		}
		seen[other] = struct{}{}
		if strings.Contains(lower, other+"'s") || strings.Contains(lower, other+"’s") {
			return true
		}
	}
	return false
}

// hopValueIsCoParticipantVisit is true when a slot is a visit / check-out of
// someone else's place ("visit dana's studio"), not a copied possession.
func hopValueIsCoParticipantVisit(v string) bool {
	padded := " " + strings.ToLower(strings.TrimSpace(v)) + " "
	if padded == "  " {
		return false
	}
	if !strings.Contains(padded, "'s") && !strings.Contains(padded, "’s") {
		return false
	}
	for _, cue := range []string{
		" visit ", " visiting ",
		" check out ", " checking out ",
		" stop by ", " stop in ",
		" drop by ",
	} {
		if strings.Contains(padded, cue) {
			return true
		}
	}
	return false
}

func hopHasCoParticipantVisit(hops []HopResult) bool {
	return firstCoParticipantVisitValue(hops) != ""
}

func firstCoParticipantVisitValue(hops []HopResult) string {
	pick := func(v string) string {
		v = strings.TrimSpace(v)
		if v == "" || typedAnswerIsHopDump(v) || !hopValueIsCoParticipantVisit(v) {
			return ""
		}
		return v
	}
	for _, h := range hops {
		for _, v := range h.Values {
			if got := pick(v); got != "" {
				return got
			}
		}
		if got := pick(h.Value); got != "" {
			return got
		}
	}
	return ""
}

func visitDestinationPlace(dest string) string {
	lower := strings.ToLower(strings.TrimSpace(dest))
	idx := strings.Index(lower, "'s ")
	if idx < 0 {
		idx = strings.Index(lower, "’s ")
	}
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(lower[idx+3:])
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return ""
	}
	return strings.Trim(fields[len(fields)-1], ".,;:!?")
}

func contentCoversVisitDestination(answer, dest string) bool {
	ans := strings.ToLower(strings.TrimSpace(answer))
	if ans == "" {
		return false
	}
	if place := visitDestinationPlace(dest); place != "" && strings.Contains(ans, place) {
		return true
	}
	d := strings.ToLower(strings.TrimSpace(dest))
	return d != "" && strings.Contains(ans, d)
}

func looksVisitPlanQuery(query string) bool {
	return queryHasToken(query, "plan", "plans", "planning", "visit", "visits", "visiting")
}

// preferCoParticipantVisitDestination keeps a visit/check-out stop when the
// reader named a related activity but dropped the actual place.
func preferCoParticipantVisitDestination(query string, hops []HopResult, answer string, pkt EvidencePacket) string {
	cur := strings.TrimSpace(answer)
	if !looksVisitPlanQuery(query) {
		return cur
	}
	dest := firstCoParticipantVisitValue(hops)
	if dest == "" {
		return cur
	}
	enriched := enrichVisitDestinationFromPacket(dest, pkt)
	if contentCoversVisitDestination(cur, dest) && !typedAnswerIsHopDump(cur) {
		if visitAnswerIsCompressedHopSlot(cur) && extraContentBeyondVisitDest(enriched, dest) > extraContentBeyondVisitDest(cur, dest) {
			return enriched
		}
		return cur
	}
	if extraContentBeyondVisitDest(enriched, dest) > extraContentBeyondVisitDest(dest, dest) {
		return enriched
	}
	return dest
}

func visitAnswerIsCompressedHopSlot(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || typedAnswerIsHopDump(s) {
		return false
	}
	if len(strings.Fields(s)) > 8 {
		return false
	}
	return hopValueIsCoParticipantVisit(s)
}

func looksVisitPlanLine(s string) bool {
	padded := " " + strings.ToLower(s) + " "
	for _, cue := range []string{
		" visit ", " visiting ", " visits ",
		" check out ", " checking out ",
		" stop by ", " stop in ",
		" drop by ",
		" plans to ", " plan to ", " planning to ",
	} {
		if strings.Contains(padded, cue) {
			return true
		}
	}
	return false
}

func extraContentBeyondVisitDest(line, dest string) int {
	place := visitDestinationPlace(dest)
	stop := map[string]struct{}{
		"the": {}, "and": {}, "for": {}, "with": {}, "from": {},
		"visit": {}, "visiting": {}, "visits": {},
		"check": {}, "checking": {}, "out": {},
		"stop": {}, "drop": {}, "by": {}, "in": {},
		"plans": {}, "plan": {}, "planning": {},
		"will": {}, "they": {}, "that": {},
	}
	if place != "" {
		stop[place] = struct{}{}
	}
	for _, tok := range tokenize(dest) {
		t := strings.ToLower(strings.Trim(tok, "'\".,;:"))
		if t != "" {
			stop[t] = struct{}{}
		}
	}
	n := 0
	seen := map[string]struct{}{}
	for _, tok := range tokenize(line) {
		t := strings.ToLower(strings.Trim(tok, "'\".,;:"))
		if utf8Len(t) < 3 {
			continue
		}
		if _, ok := stop[t]; ok {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		n++
	}
	return n
}

// enrichVisitDestinationFromPacket expands a compressed visit hop slot using a
// short packet source line that names the same place plus leftover purpose.
func enrichVisitDestinationFromPacket(dest string, pkt EvidencePacket) string {
	place := visitDestinationPlace(dest)
	if place == "" {
		return dest
	}
	best := dest
	bestExtra := extraContentBeyondVisitDest(dest, dest)
	bestN := utf8Len(dest)
	for _, line := range packetContentLines(pkt) {
		line = strings.TrimSpace(line)
		n := utf8Len(line)
		if n < 24 || n > 180 {
			continue
		}
		if typedAnswerIsHopDump(line) {
			continue
		}
		if !strings.Contains(strings.ToLower(line), place) {
			continue
		}
		if !hopValueIsCoParticipantVisit(line) && !looksVisitPlanLine(line) {
			continue
		}
		extra := extraContentBeyondVisitDest(line, dest)
		if extra <= 0 {
			continue
		}
		if extra > bestExtra || (extra == bestExtra && n < bestN) {
			best = line
			bestExtra = extra
			bestN = n
		}
	}
	return best
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

func hopEntityRawValueGroups(results []HopResult) ([]string, map[string][]string) {
	order := make([]string, 0, 2)
	groups := map[string][]string{}
	for _, r := range results {
		switch r.Kind {
		case "follow_relation", "fetch_predicate", "answer_slot":
			ent := strings.ToLower(strings.TrimSpace(r.Entity))
			if ent == "" {
				continue
			}
			vals := r.Values
			if len(vals) == 0 && strings.TrimSpace(r.Value) != "" {
				vals = []string{r.Value}
			}
			if len(vals) == 0 {
				continue
			}
			if _, ok := groups[ent]; !ok {
				order = append(order, ent)
			}
			for _, v := range vals {
				v = strings.TrimSpace(v)
				if v == "" || utf8Len(v) > 80 {
					continue
				}
				groups[ent] = append(groups[ent], v)
			}
		}
	}
	return order, groups
}

// intersectHopValuesByRareSharedToken keeps values that share a rare 8+ letter
// token across every hop entity (signed basketball ∩ basketball trophy).
// Generic high-df tokens (collection, photo) lose to rarer ones. When typed
// Values omit the shared object, fall back to hop-content snippets.
func intersectHopValuesByRareSharedToken(results []HopResult) []string {
	if got := rareSharedFromGroups(hopEntityOmittedContentGroups(results)); len(got) > 0 {
		return got
	}
	return rareSharedFromGroups(hopEntityRawValueGroups(results))
}

func rareSharedFromGroups(order []string, groups map[string][]string) []string {
	if len(order) < 2 {
		return nil
	}
	type hit struct {
		ents     map[string]struct{}
		df       int
		minWords int
	}
	toks := map[string]*hit{}
	for _, ent := range order {
		seenTok := map[string]struct{}{}
		for _, v := range groups[ent] {
			nWords := len(strings.Fields(v))
			for _, tok := range slotValueTokens(v) {
				if len(tok) < 8 {
					continue
				}
				h := toks[tok]
				if h == nil {
					h = &hit{ents: map[string]struct{}{}, minWords: 1 << 20}
					toks[tok] = h
				}
				h.df++
				if nWords < h.minWords {
					h.minWords = nWords
				}
				if _, ok := seenTok[tok]; !ok {
					h.ents[ent] = struct{}{}
					seenTok[tok] = struct{}{}
				}
			}
		}
	}
	best, bestDF, bestWords, bestLen := "", 1<<20, 1<<20, 0
	for tok, h := range toks {
		if len(h.ents) < len(order) {
			continue
		}
		better := h.minWords < bestWords ||
			(h.minWords == bestWords && h.df < bestDF) ||
			(h.minWords == bestWords && h.df == bestDF && len(tok) > bestLen)
		if better {
			best, bestDF, bestWords, bestLen = tok, h.df, h.minWords, len(tok)
		}
	}
	if best == "" {
		return nil
	}
	_ = bestDF
	_ = bestLen
	seen := map[string]struct{}{}
	out := make([]string, 0, len(order))
	for _, ent := range order {
		bestVal, bestN := "", 1<<20
		for _, v := range groups[ent] {
			if !contentCoversQueryToken(v, best) {
				continue
			}
			n := utf8Len(v)
			if n < bestN {
				bestVal, bestN = v, n
			}
		}
		key := strings.ToLower(strings.TrimSpace(bestVal))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, strings.TrimSpace(bestVal))
	}
	return out
}

func hopContentSnippet(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	if v, ok := slotValueFromMemoryContent(content); ok {
		v = strings.TrimSpace(v)
		if v != "" && utf8Len(v) <= 80 && !looksCrowdedHopDump(v) {
			return v
		}
	}
	lower := strings.ToLower(content)
	for _, sep := range []string{" possesses ", " possessed "} {
		i := strings.Index(lower, sep)
		if i < 0 {
			continue
		}
		v := strings.TrimSpace(content[i+len(sep):])
		if j := strings.IndexAny(v, ".(["); j > 0 {
			v = strings.TrimSpace(v[:j])
		}
		if v != "" && utf8Len(v) <= 80 && !looksCrowdedHopDump(v) {
			return v
		}
	}
	if utf8Len(content) <= 80 && !looksCrowdedHopDump(content) {
		return content
	}
	return ""
}

func snippetCoveredByValues(snip string, vals []string) bool {
	snip = strings.TrimSpace(snip)
	if snip == "" {
		return false
	}
	low := strings.ToLower(snip)
	snipToks := slotValueTokens(snip)
	for _, v := range vals {
		v = strings.ToLower(strings.TrimSpace(v))
		if v == "" || utf8Len(v) < 4 {
			continue
		}
		if strings.Contains(low, v) || strings.Contains(v, low) {
			return true
		}
		vt := slotValueTokens(v)
		if len(vt) >= 2 && slotTokensSubset(vt, snipToks) {
			return true
		}
	}
	return false
}

func hopEntityOmittedContentGroups(results []HopResult) ([]string, map[string][]string) {
	valOrder, valGroups := hopEntityRawValueGroups(results)
	order := append([]string(nil), valOrder...)
	seenEnt := map[string]struct{}{}
	for _, ent := range order {
		seenEnt[ent] = struct{}{}
	}
	groups := map[string][]string{}
	for _, r := range results {
		switch r.Kind {
		case "follow_relation", "fetch_predicate", "answer_slot":
			ent := strings.ToLower(strings.TrimSpace(r.Entity))
			if ent == "" {
				continue
			}
			if _, ok := seenEnt[ent]; !ok {
				order = append(order, ent)
				seenEnt[ent] = struct{}{}
			}
			for _, c := range r.Contents {
				if looksCrowdedHopDump(c) {
					continue
				}
				snip := hopContentSnippet(c)
				if snip == "" {
					continue
				}
				if snippetCoveredByValues(snip, valGroups[ent]) {
					continue
				}
				groups[ent] = append(groups[ent], snip)
			}
		}
	}
	filtered := make([]string, 0, len(order))
	for _, ent := range order {
		if len(groups[ent]) > 0 {
			filtered = append(filtered, ent)
		}
	}
	return filtered, groups
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
		if len(vals) == 0 && hopsKeepTypedJoin(results) {
			vals = intersectHopValuesByRareSharedToken(results)
		}
		if typedAnswerIsHopDump(strings.Join(vals, ", ")) && hopsKeepTypedJoin(results) {
			if rare := intersectHopValuesByRareSharedToken(results); len(rare) > 0 && len(rare) < len(vals) {
				vals = rare
			}
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
	prevUnc := uncoveredTargets(*pkt)
	pkt.Coverage["hop_join_proven"] = proven
	pkt.Coverage["hop_results"] = hopResults
	pkt.Coverage["bridge_count"] = countRole(items, "bridge")
	pkt.Coverage["direct_count"] = countRole(items, "direct")
	pkt.Coverage["context_count"] = len(pkt.ContextEvidence)
	pkt.Coverage["proof_count"] = len(pkt.ProofChain)
	if proven {
		pkt.Coverage["uncovered"] = []string{}
	} else {
		seen := map[string]struct{}{}
		unc := make([]string, 0, len(prevUnc)+2)
		add := func(s string) {
			s = strings.ToLower(strings.TrimSpace(s))
			if s == "" {
				return
			}
			if _, ok := seen[s]; ok {
				return
			}
			seen[s] = struct{}{}
			unc = append(unc, s)
		}
		for _, t := range prevUnc {
			add(t)
		}
		for _, r := range hopResults {
			if r.Source == "unresolved" {
				add(firstNonEmpty(r.OutputKey, r.Kind))
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
