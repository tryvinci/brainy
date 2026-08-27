package memory

import (
	"math"
	"strings"
)

// selectEvidenceSet maximizes marginal coverage for list/multi-hop queries
// (program §12.6 greedy set-cover). limit caps selected items.
func selectEvidenceSet(ranked []rankedSearchResult, limit int) []rankedSearchResult {
	return selectEvidenceSetCovering(ranked, limit, nil)
}

// selectEvidenceSetCovering first admits memories that cover leftover query
// tokens (rare nouns starved by name/recency pooling), then the existing
// content-diversity greedy fill.
func selectEvidenceSetCovering(ranked []rankedSearchResult, limit int, queryTokens []string) []rankedSearchResult {
	if limit <= 0 || len(ranked) <= limit {
		return ranked
	}
	selected := make([]rankedSearchResult, 0, limit)
	covered := map[string]struct{}{}
	used := make([]bool, len(ranked))

	coverageTokens := func(content string) []string {
		return contentBearingTokens(tokenize(content))
	}
	markUsed := func(i int) {
		used[i] = true
		picked := ranked[i]
		selected = append(selected, picked)
		for _, t := range coverageTokens(picked.result.Content) {
			covered[t] = struct{}{}
		}
	}

	for _, tok := range queryTokensByRarity(queryTokens, ranked) {
		if len(selected) >= limit {
			break
		}
		if queryTokenCoveredInRanked(selected, tok) {
			continue
		}
		bestI := -1
		bestScore := -1.0
		for i, item := range ranked {
			if used[i] {
				continue
			}
			if !contentCoversQueryToken(item.result.Content, tok) {
				continue
			}
			if item.result.Score > bestScore {
				bestScore = item.result.Score
				bestI = i
			}
		}
		if bestI >= 0 {
			markUsed(bestI)
		}
	}

	marginal := func(content string) float64 {
		toks := coverageTokens(content)
		if len(toks) == 0 {
			return 0
		}
		newCount := 0
		for _, t := range toks {
			if _, ok := covered[t]; !ok {
				newCount++
			}
		}
		return float64(newCount) / float64(len(toks))
	}

	for len(selected) < limit {
		bestI := -1
		bestScore := -1.0
		for i, item := range ranked {
			if used[i] {
				continue
			}
			content := strings.TrimSpace(item.result.Content)
			if content == "" || strings.HasSuffix(content, "?") {
				continue
			}
			marg := marginal(content)
			// Prefer relevance, then coverage, lightly penalize redundancy.
			score := item.result.Score*(0.55+0.45*marg) + marg*0.35
			if marg < 0.05 && len(selected) > 0 {
				score *= 0.7
			}
			if score > bestScore {
				bestScore = score
				bestI = i
			}
		}
		if bestI < 0 || bestScore <= 0 {
			break
		}
		markUsed(bestI)
		if math.IsNaN(bestScore) {
			break
		}
	}
	// Fill remaining slots by original rank if coverage pass under-filled.
	if len(selected) < limit {
		for i, item := range ranked {
			if used[i] {
				continue
			}
			selected = append(selected, item)
			if len(selected) >= limit {
				break
			}
		}
	}
	return selected
}

func contentCoversQueryToken(content, token string) bool {
	token = strings.ToLower(strings.TrimSpace(token))
	if token == "" {
		return false
	}
	return recordTokensCover(tokenize(content), token)
}

func queryTokenCoveredInRanked(items []rankedSearchResult, token string) bool {
	for _, item := range items {
		if contentCoversQueryToken(item.result.Content, token) {
			return true
		}
	}
	return false
}

func queryTokenCoveredInResults(results []SearchResult, token string) bool {
	for _, r := range results {
		if contentCoversQueryToken(r.Content, token) {
			return true
		}
	}
	return false
}

func distinctiveQueryTokens(tokens []string) []string {
	out := make([]string, 0, len(tokens))
	seen := map[string]struct{}{}
	for _, t := range contentBearingTokens(tokens) {
		t = strings.ToLower(strings.TrimSpace(t))
		if len(t) < 4 {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func queryTokensByRarity(queryTokens []string, ranked []rankedSearchResult) []string {
	toks := distinctiveQueryTokens(queryTokens)
	if len(toks) <= 1 || len(ranked) == 0 {
		return toks
	}
	df := make(map[string]int, len(toks))
	for _, item := range ranked {
		for _, tok := range toks {
			if contentCoversQueryToken(item.result.Content, tok) {
				df[tok]++
			}
		}
	}
	for i := 0; i < len(toks); i++ {
		for j := i + 1; j < len(toks); j++ {
			di, dj := df[toks[i]], df[toks[j]]
			if dj < di || (dj == di && len(toks[j]) > len(toks[i])) {
				toks[i], toks[j] = toks[j], toks[i]
			}
		}
	}
	return toks
}

func uncoveredQueryTokensFromResults(query string, results []SearchResult) []string {
	toks := distinctiveQueryTokens(tokenize(query))
	out := make([]string, 0, len(toks))
	for _, tok := range toks {
		if !queryTokenCoveredInResults(results, tok) {
			out = append(out, tok)
		}
	}
	return out
}

func distinctiveProbeToken(tokens []string) string {
	best := ""
	for _, t := range tokens {
		t = strings.ToLower(strings.TrimSpace(t))
		if len(t) < 4 {
			continue
		}
		if len(t) > len(best) {
			best = t
		}
	}
	return best
}

func candidatesCoverQueryToken(candidates map[string]MemoryRecord, token string) bool {
	for _, rec := range candidates {
		if contentCoversQueryToken(rec.Content, token) {
			return true
		}
	}
	return false
}

func uncoveredQueryTokensInCandidates(query string, candidates map[string]MemoryRecord, queryTokens []string) []string {
	toks := distinctiveQueryTokens(queryTokens)
	out := make([]string, 0, len(toks))
	for _, tok := range toks {
		if candidatesCoverQueryTokenNamed(query, candidates, tok) {
			continue
		}
		out = append(out, tok)
	}
	return out
}

// candidatesCoverQueryTokenNamed is true when some candidate both contains
// token and names a query person. A first-person leftover that only has the
// event token must not count as coverage for a named-person when-event, or
// compiled "Name went X" facts never enter the pool.
func candidatesCoverQueryTokenNamed(query string, candidates map[string]MemoryRecord, token string) bool {
	ents := hopQueryEntities(query)
	if len(ents) == 0 {
		return candidatesCoverQueryToken(candidates, token)
	}
	for _, rec := range candidates {
		if leftoverCoveringQueryEntityHits(query, rec.Content) == 0 {
			continue
		}
		if contentCoversQueryToken(rec.Content, token) {
			return true
		}
	}
	return false
}

func coverQueryTokensThenCap(ranked []rankedSearchResult, limit int, queryTokens []string) []rankedSearchResult {
	if limit <= 0 || len(ranked) <= limit {
		return ranked
	}
	if len(distinctiveQueryTokens(queryTokens)) == 0 {
		return ranked[:limit]
	}
	return selectEvidenceSetCovering(ranked, limit, queryTokens)
}
