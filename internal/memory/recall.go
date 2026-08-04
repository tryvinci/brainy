package memory

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

// RecallRequest is the product synthesis surface (master-plan W4 / program §14).
type RecallRequest struct {
	TenantID           string `json:"tenant_id"`
	SubjectID          string `json:"subject_id"`
	Query              string `json:"q"`
	Mode               string `json:"mode"` // context | enumerate | answer
	Vertical           string `json:"vertical,omitempty"`
	TopK               int    `json:"top_k,omitempty"`
	BudgetTokens       int    `json:"budget_tokens,omitempty"`
	IncludeHistorical  bool   `json:"include_historical,omitempty"`
	View               string `json:"view,omitempty"` // current | historical | all
	AsOf               string `json:"as_of,omitempty"`
	IncludeProvenance  bool   `json:"include_provenance,omitempty"`
	MaxEvidenceTokens  int    `json:"max_evidence_tokens,omitempty"`
	OracleMode         string `json:"oracle_mode,omitempty"` // reader | evidence | semantic | ""
}

// RecallItem is one enumerated value with evidence.
type RecallItem struct {
	Value      string   `json:"value"`
	Predicate  string   `json:"predicate,omitempty"`
	Evidence   []string `json:"evidence"`
	ObservedAt string   `json:"observed_at,omitempty"`
}

// RecallResponse is returned by POST /recall.
type RecallResponse struct {
	Mode         string         `json:"mode"`
	AnswerStatus string         `json:"answer_status,omitempty"`
	Intents      []string       `json:"intents,omitempty"`
	ContextBlock string         `json:"context_block,omitempty"`
	Answer       string         `json:"answer,omitempty"`
	Abstained    bool           `json:"abstained,omitempty"`
	Items        []RecallItem   `json:"items,omitempty"`
	Memories     []SearchResult `json:"memories"`
	Coverage     map[string]any `json:"coverage,omitempty"`
	Explain      map[string]any `json:"explain,omitempty"`
	Trace        *SearchTrace   `json:"trace,omitempty"`
}

// Recall assembles product-side context / enumeration / answer.
func (s *Service) Recall(ctx context.Context, req RecallRequest) (RecallResponse, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.SubjectID) == "" || strings.TrimSpace(req.Query) == "" {
		return RecallResponse{}, errors.New("tenant_id, subject_id, and q are required")
	}
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "context"
	}
	topK := req.TopK
	if topK <= 0 {
		topK = 30
	}
	budget := req.BudgetTokens
	if budget <= 0 {
		budget = 4000
	}

	search, err := s.SearchOpt(ctx, req.TenantID, req.SubjectID, req.Vertical, "", req.Query, SearchOptions{
		IncludeHistorical: req.IncludeHistorical || strings.EqualFold(req.View, "historical") || strings.EqualFold(req.View, "all"),
		Limit:             topK,
	})
	if err != nil {
		return RecallResponse{}, err
	}

	intents := AnalyzeQueryIntents(req.Query)
	plan := PlanQuery(req.Query, intents)
	out := RecallResponse{
		Mode:     mode,
		Memories: search.Results,
		Intents:  intents,
		Trace:    search.Trace,
		Explain: map[string]any{
			"top_k":         topK,
			"budget_tokens": budget,
			"query_plan":    plan,
		},
	}
	if req.View != "" {
		out.Explain["view"] = req.View
	}
	asOfTime, asOfOK := ParseAsOf(req.AsOf)
	if req.AsOf != "" {
		out.Explain["as_of"] = req.AsOf
		out.Explain["as_of_applied"] = asOfOK
		if !asOfOK {
			out.Explain["as_of_error"] = "unparseable"
		}
	}
	s.applyTemporalResolution(ctx, req, &out, asOfTime, asOfOK)
	pkt := BuildEvidencePacket(plan, search.Results, out.Explain)
	out.Explain["evidence_packet"] = pkt
	if out.Coverage == nil && pkt.Coverage != nil {
		out.Coverage = pkt.Coverage
	}
	oracle := strings.ToLower(strings.TrimSpace(req.OracleMode))
	if oracle != "" {
		out.Explain["oracle_mode"] = oracle
		switch oracle {
		case "evidence":
			if raw, ok := s.store.(RawEvidenceWriter); ok {
				rows, err := raw.ListEvidence(ctx, req.TenantID, req.SubjectID, topK)
				if err != nil {
					return RecallResponse{}, err
				}
				out.Explain["oracle_evidence_count"] = len(rows)
				out.ContextBlock = formatEvidenceOracle(rows, budget)
				out.AnswerStatus = AnswerSupported
				if len(rows) == 0 {
					out.AnswerStatus = AnswerNotFound
					out.Abstained = true
				}
				out.Coverage = map[string]any{"targets": 1, "satisfied": len(rows) > 0, "oracle": "evidence"}
				return out, nil
			}
			out.Explain["oracle_unsupported"] = true
			out.AnswerStatus = AnswerInsufficient
			out.Abstained = true
			out.Answer = "oracle_unsupported"
			return out, nil
		case "semantic":
			// Semantic-object-present: prefer atom-indexed memories; fall back to
			// any active search hit as weak semantic presence.
			atomCount := 0
			if indexer, ok := s.store.(AtomIndexer); ok {
				ids, err := indexer.ListAtomMemoryIDs(ctx, req.TenantID, req.SubjectID, "", "", topK)
				if err == nil {
					atomCount = len(ids)
				}
			}
			out.Explain["oracle_semantic"] = true
			out.Explain["oracle_atom_count"] = atomCount
			out.Explain["oracle_memory_count"] = len(search.Results)
			present := atomCount > 0 || len(search.Results) > 0
			out.Coverage = map[string]any{"targets": 1, "satisfied": present, "oracle": "semantic"}
			if !present {
				out.AnswerStatus = AnswerNotFound
				out.Abstained = true
				out.Answer = "semantic_absent"
				return out, nil
			}
			out.ContextBlock = assembleContextBlock(search.Results, budget)
			out.AnswerStatus = AnswerSupported
			out.Answer = firstStatement(search.Results)
			return out, nil
		case "retrieval":
			out.Explain["oracle_retrieval"] = true
			out.Explain["oracle_memory_count"] = len(search.Results)
			out.Coverage = map[string]any{"targets": 1, "satisfied": len(search.Results) > 0, "oracle": "retrieval"}
			if len(search.Results) == 0 {
				out.AnswerStatus = AnswerNotFound
				out.Abstained = true
				out.Answer = "retrieval_miss"
				return out, nil
			}
			out.ContextBlock = assembleContextBlock(search.Results, budget)
			out.AnswerStatus = AnswerSupported
			out.Answer = firstStatement(search.Results)
			return out, nil
		case "coverage":
			items := s.enumerateFromSearch(ctx, req, search.Results)
			out.Items = items
			out.Explain["oracle_coverage"] = true
			out.Explain["oracle_item_count"] = len(items)
			satisfied := len(items) > 0
			out.Coverage = map[string]any{"targets": 1, "satisfied": satisfied, "oracle": "coverage"}
			if !satisfied {
				out.AnswerStatus = AnswerInsufficient
				out.Abstained = true
				out.Answer = "coverage_miss"
				return out, nil
			}
			out.ContextBlock = formatEnumerateContext(items)
			out.AnswerStatus = AnswerSupported
			out.Answer = out.ContextBlock
			return out, nil
		case "reader":
			// Reader oracle = product answer path (same as mode=answer) but tagged.
			out.Explain["oracle_reader"] = true
			req.Mode = "answer"
			mode = "answer"
		default:
			out.Explain["oracle_unsupported"] = true
			out.AnswerStatus = AnswerInsufficient
			out.Abstained = true
			out.Answer = "oracle_unsupported"
			return out, nil
		}
	}

	switch mode {
	case "enumerate":
		items := s.enumerateFromSearch(ctx, req, search.Results)
		out.Items = items
		out.ContextBlock = formatEnumerateContext(items)
		out.Explain["item_count"] = len(items)
		out.AnswerStatus = AnswerSupported
		if len(items) == 0 {
			out.AnswerStatus = AnswerNotFound
			out.Abstained = true
		}
		out.Coverage = map[string]any{"targets": 1, "satisfied": len(items) > 0}
	case "answer":
		items := s.enumerateFromSearch(ctx, req, search.Results)
		out.Items = items
		out.ContextBlock = assembleContextBlock(search.Results, budget)
		if len(items) > 0 {
			vals := make([]string, 0, len(items))
			for _, it := range items {
				vals = append(vals, it.Value)
			}
			out.Answer = strings.Join(vals, ", ")
			out.AnswerStatus = AnswerSupported
			if len(items) == 1 && looksListQuery(tokenize(req.Query)) {
				out.AnswerStatus = AnswerPartiallySupported
			}
		} else if out.ContextBlock != "" {
			// Deterministic extractive fallback: first non-question statement.
			out.Answer = firstStatement(search.Results)
			out.AnswerStatus = AnswerPartiallySupported
		}
		if strings.TrimSpace(out.Answer) == "" {
			out.Abstained = true
			out.Answer = "not in memory"
			out.AnswerStatus = AnswerInsufficient
		}
		out.Coverage = map[string]any{
			"targets":   1,
			"satisfied": !out.Abstained,
		}
	case "context":
		out.ContextBlock = assembleContextBlock(search.Results, budget)
		if out.ContextBlock == "" {
			out.AnswerStatus = AnswerNotFound
			out.Abstained = true
		} else {
			out.AnswerStatus = AnswerSupported
		}
	default:
		return RecallResponse{}, fmt.Errorf("unsupported recall mode %q", mode)
	}
	out.Explain["memory_count"] = len(search.Results)
	out.Explain["context_runes"] = utf8.RuneCountInString(out.ContextBlock)
	if temporalAnswer, _ := out.Explain["temporal_answer"].(string); temporalAnswer != "" {
		if mode == "answer" && (out.Abstained || out.AnswerStatus == AnswerPartiallySupported || strings.TrimSpace(out.Answer) == "" || strings.TrimSpace(out.Answer) == "not in memory") {
			out.Answer = temporalAnswer
			out.Abstained = false
			out.AnswerStatus = AnswerSupported
			out.Coverage = map[string]any{"targets": 1, "satisfied": true, "source": "temporal_resolver"}
		}
		if mode == "context" && out.ContextBlock == "" {
			out.ContextBlock = temporalAnswer
			out.Abstained = false
			out.AnswerStatus = AnswerSupported
		}
	}
	return out, nil
}

// applyTemporalResolution fills explain + optional temporal_answer from typed
// current-state / as-of / history reads when the store supports TemporalStore.
func (s *Service) applyTemporalResolution(ctx context.Context, req RecallRequest, out *RecallResponse, asOf time.Time, asOfOK bool) {
	ts, ok := s.store.(TemporalStore)
	if !ok {
		// CurrentStateStore alone still helps view=current.
		if cs, ok := s.store.(CurrentStateStore); ok {
			s.applyCurrentStateOnly(ctx, req, out, cs)
		}
		return
	}
	view := strings.ToLower(strings.TrimSpace(req.View))
	wantCurrent := view == "current" || view == "" && hasIntent(out.Intents, IntentCurrentState)
	wantHistory := req.IncludeHistorical || view == "historical" || view == "all" || hasIntent(out.Intents, IntentHistoricalState) || hasIntent(out.Intents, IntentTemporalSequence)
	preds := predicateHintsFromQuery(req.Query)
	resolved := make([]map[string]any, 0, len(preds))
	var answerParts []string

	for _, pred := range preds {
		entry := map[string]any{"predicate": pred}
		switch {
		case asOfOK && strings.EqualFold(view, "as_known_at"):
			id, val, found, _ := ts.GetStateAsKnownAt(ctx, req.TenantID, req.SubjectID, pred, asOf)
			entry["mode"] = "as_known_at"
			if found {
				entry["memory_id"] = id
				entry["value"] = val
				answerParts = append(answerParts, pred+": "+val)
			}
		case asOfOK:
			id, val, found, _ := ts.GetStateAsOf(ctx, req.TenantID, req.SubjectID, pred, asOf)
			entry["mode"] = "as_of"
			if found {
				entry["memory_id"] = id
				entry["value"] = val
				answerParts = append(answerParts, pred+": "+val)
			}
		case wantCurrent || view == "current":
			id, val, policy, found, _ := ts.GetCurrentState(ctx, req.TenantID, req.SubjectID, pred)
			entry["mode"] = "current"
			entry["policy"] = policy
			if found {
				entry["memory_id"] = id
				entry["value"] = val
				answerParts = append(answerParts, pred+": "+val)
			}
		}
		if wantHistory {
			hist, err := ts.ListStateHistory(ctx, req.TenantID, req.SubjectID, pred, 8)
			if err == nil && len(hist) > 0 {
				entry["history"] = hist
			}
		}
		if entry["value"] != nil || entry["history"] != nil {
			resolved = append(resolved, entry)
		}
	}
	if len(resolved) > 0 {
		out.Explain["temporal"] = resolved
		out.Explain["temporal_applied"] = true
	}
	if len(answerParts) > 0 {
		out.Explain["temporal_answer"] = strings.Join(answerParts, "; ")
	}
}

func (s *Service) applyCurrentStateOnly(ctx context.Context, req RecallRequest, out *RecallResponse, cs CurrentStateStore) {
	view := strings.ToLower(strings.TrimSpace(req.View))
	if view != "current" && !hasIntent(out.Intents, IntentCurrentState) {
		return
	}
	preds := predicateHintsFromQuery(req.Query)
	resolved := make([]map[string]any, 0, len(preds))
	var parts []string
	for _, pred := range preds {
		id, val, policy, found, _ := cs.GetCurrentState(ctx, req.TenantID, req.SubjectID, pred)
		if !found {
			continue
		}
		resolved = append(resolved, map[string]any{
			"predicate": pred, "memory_id": id, "value": val, "policy": policy, "mode": "current",
		})
		parts = append(parts, pred+": "+val)
	}
	if len(resolved) > 0 {
		out.Explain["temporal"] = resolved
		out.Explain["temporal_applied"] = true
		out.Explain["temporal_answer"] = strings.Join(parts, "; ")
	}
}

func hasIntent(intents []string, want string) bool {
	for _, i := range intents {
		if i == want {
			return true
		}
	}
	return false
}

func (s *Service) enumerateFromSearch(ctx context.Context, req RecallRequest, results []SearchResult) []RecallItem {
	tokens := tokenize(req.Query)
	pred := predicateFromListQuery(tokens)
	seen := map[string]RecallItem{}
	order := make([]string, 0)

	add := func(value, predicate, memoryID, observed string) {
		v := strings.TrimSpace(value)
		if v == "" {
			return
		}
		key := strings.ToLower(v)
		if existing, ok := seen[key]; ok {
			existing.Evidence = appendUnique(existing.Evidence, memoryID)
			seen[key] = existing
			return
		}
		seen[key] = RecallItem{
			Value:      v,
			Predicate:  predicate,
			Evidence:   []string{memoryID},
			ObservedAt: observed,
		}
		order = append(order, key)
	}

	// Prefer indexed atoms when available.
	if pred != "" {
		if indexer, ok := s.store.(AtomIndexer); ok {
			ids, err := indexer.ListAtomMemoryIDs(ctx, req.TenantID, req.SubjectID, pred, "", 40)
			if err == nil {
				byID := map[string]SearchResult{}
				for _, r := range results {
					byID[r.MemoryID] = r
				}
				for _, id := range ids {
					if r, ok := byID[id]; ok {
						val := valueFromMemoryContent(r.Content)
						obs := ""
						if r.ObservedAt != nil {
							obs = r.ObservedAt.UTC().Format(time.RFC3339)
						}
						add(val, pred, id, obs)
					}
				}
			}
		}
	}

	for _, r := range results {
		obs := ""
		if r.ObservedAt != nil {
			obs = r.ObservedAt.UTC().Format(time.RFC3339)
		}
		val := valueFromMemoryContent(r.Content)
		p := pred
		if p == "" {
			p = "fact"
		}
		add(val, p, r.MemoryID, obs)
	}

	items := make([]RecallItem, 0, len(order))
	for _, key := range order {
		items = append(items, seen[key])
	}
	return items
}

func formatEvidenceOracle(rows []map[string]any, budgetTokens int) string {
	if budgetTokens <= 0 {
		budgetTokens = 4000
	}
	budget := budgetTokens * 4
	var b strings.Builder
	for _, row := range rows {
		content, _ := row["content"].(string)
		content = strings.TrimSpace(content)
		if content == "" {
			continue
		}
		line := "- " + content + "\n"
		if b.Len()+len(line) > budget {
			break
		}
		b.WriteString(line)
	}
	return b.String()
}

func valueFromMemoryContent(content string) string {
	lower := strings.ToLower(content)
	for _, sep := range []string{" participates in ", " enjoys ", " moved from ", " is from ", " kids like ", " read \"", " has done ", " is a ", " is "} {
		if i := strings.Index(lower, sep); i >= 0 {
			v := strings.TrimSpace(content[i+len(sep):])
			v = strings.Trim(v, "\"")
			if j := strings.IndexAny(v, ".(["); j > 0 {
				v = strings.TrimSpace(v[:j])
			}
			return v
		}
	}
	// Fall back to full statement without speaker prefix.
	v := speakerPrefixRe.ReplaceAllString(content, "")
	return strings.TrimSpace(v)
}

func assembleContextBlock(results []SearchResult, budgetTokens int) string {
	if budgetTokens <= 0 {
		budgetTokens = 4000
	}
	// Rough rune budget ≈ 4 chars/token.
	budget := budgetTokens * 4
	var b strings.Builder
	seen := map[string]struct{}{}
	for _, r := range results {
		content := strings.TrimSpace(r.Content)
		if content == "" || strings.HasSuffix(content, "?") {
			continue
		}
		key := strings.ToLower(content)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		line := "- " + content
		if r.ObservedAt != nil {
			line += " [event_time=" + r.ObservedAt.UTC().Format(time.RFC3339) + "]"
		}
		line += "\n"
		if b.Len()+len(line) > budget {
			break
		}
		b.WriteString(line)
	}
	return strings.TrimSpace(b.String())
}

func formatEnumerateContext(items []RecallItem) string {
	if len(items) == 0 {
		return ""
	}
	parts := make([]string, 0, len(items))
	for _, it := range items {
		parts = append(parts, it.Value)
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

func firstStatement(results []SearchResult) string {
	for _, r := range results {
		c := strings.TrimSpace(r.Content)
		if c != "" && !strings.HasSuffix(c, "?") {
			return c
		}
	}
	return ""
}

func appendUnique(ids []string, id string) []string {
	for _, existing := range ids {
		if existing == id {
			return ids
		}
	}
	return append(ids, id)
}
