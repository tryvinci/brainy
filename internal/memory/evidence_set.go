package memory

import (
	"math"
	"strings"
)

// selectEvidenceSet maximizes marginal coverage for list/multi-hop queries
// (program §12.6 greedy set-cover). limit caps selected items.
func selectEvidenceSet(ranked []rankedSearchResult, limit int) []rankedSearchResult {
	if limit <= 0 || len(ranked) <= limit {
		return ranked
	}
	selected := make([]rankedSearchResult, 0, limit)
	covered := map[string]struct{}{}
	used := make([]bool, len(ranked))

	coverageTokens := func(content string) []string {
		return contentBearingTokens(tokenize(content))
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
		used[bestI] = true
		picked := ranked[bestI]
		selected = append(selected, picked)
		for _, t := range coverageTokens(picked.result.Content) {
			covered[t] = struct{}{}
		}
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
