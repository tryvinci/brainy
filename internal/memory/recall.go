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
		IncludeHistorical: req.IncludeHistorical,
		Limit:             topK,
	})
	if err != nil {
		return RecallResponse{}, err
	}

	out := RecallResponse{
		Mode:     mode,
		Memories: search.Results,
		Intents:  AnalyzeQueryIntents(req.Query),
		Trace:    search.Trace,
		Explain:  map[string]any{"top_k": topK, "budget_tokens": budget},
	}
	if req.OracleMode != "" {
		out.Explain["oracle_mode"] = req.OracleMode
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
	return out, nil
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
