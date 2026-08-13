package memory

import (
	"math"
	"os"
	"strings"
)

// Fusion V2: Mem0-style additive multi-signal scoring (program §12.5).
// semantic and bm25 are expected in [0,1]; entityBoost in [0, entityBoostWeightV2].

const entityBoostWeightV2 = 0.5

// FusionV2Enabled defaults ON; set BRAINY_FUSION_V2=false to disable.
func FusionV2Enabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("BRAINY_FUSION_V2"))) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

// bm25SigmoidParams mirrors Mem0 get_bm25_params (query-length adaptive).
func bm25SigmoidParams(numTerms int) (midpoint, steepness float64) {
	switch {
	case numTerms <= 3:
		return 5.0, 0.7
	case numTerms <= 6:
		return 7.0, 0.6
	case numTerms <= 9:
		return 9.0, 0.5
	case numTerms <= 15:
		return 10.0, 0.5
	default:
		return 12.0, 0.5
	}
}

// NormalizeBM25Sigmoid maps an unbounded BM25-like score into [0,1].
func NormalizeBM25Sigmoid(raw float64, numQueryTerms int) float64 {
	if raw <= 0 {
		return 0
	}
	mid, steep := bm25SigmoidParams(numQueryTerms)
	return 1.0 / (1.0 + math.Exp(-steep*(raw-mid)))
}

// ScoreAndRankV2 fuses signals the Mem0 way: threshold-gate semantic first,
// then (semantic + bm25 + entity) / max_possible, capped at 1.
//
// semanticOnlyFloor gates pure-semantic candidates (no lexical/entity evidence).
// Callers should pass a higher floor for short entity-probe queries.
func ScoreAndRankV2(semantic, bm25, entityBoost, semanticThreshold, semanticOnlyFloor float64) (combined float64, details map[string]float64) {
	details = map[string]float64{
		"semantic":      semantic,
		"bm25":          bm25,
		"entity_boost":  entityBoost,
		"sem_threshold": semanticThreshold,
	}
	if semanticOnlyFloor <= 0 {
		semanticOnlyFloor = 0.42
	}
	if bm25 <= 0 && entityBoost <= 0 {
		if semantic < semanticOnlyFloor {
			details["combined"] = 0
			details["semantic_only_floor"] = semanticOnlyFloor
			return 0, details
		}
	} else if semantic > 0 && semantic < semanticThreshold && bm25 <= 0 && entityBoost <= 0 {
		details["combined"] = 0
		return 0, details
	}
	hasBM25 := bm25 > 0
	hasEntity := entityBoost > 0
	maxPossible := 1.0
	if hasBM25 {
		maxPossible += 1.0
	}
	if hasEntity {
		maxPossible += entityBoostWeightV2
	}
	// When semantic is zero but lexical/entity fire, still allow a result.
	sem := semantic
	if sem <= 0 && (hasBM25 || hasEntity) {
		sem = 0
		maxPossible = 0
		if hasBM25 {
			maxPossible += 1.0
		}
		if hasEntity {
			maxPossible += entityBoostWeightV2
		}
		if maxPossible == 0 {
			return 0, details
		}
		raw := bm25 + entityBoost
		combined = math.Min(raw/maxPossible, 1.0)
		details["combined"] = combined
		details["max_possible"] = maxPossible
		return combined, details
	}
	raw := sem + bm25 + entityBoost
	combined = math.Min(raw/maxPossible, 1.0)
	details["combined"] = combined
	details["max_possible"] = maxPossible
	return combined, details
}

// ScoreAndRankV2Temporal is ScoreAndRankV2 plus a [0,1] temporal_score channel.
// Zero temporal does not change max_possible (not a recency tie-break).
func ScoreAndRankV2Temporal(semantic, bm25, entityBoost, temporal, semanticThreshold, semanticOnlyFloor float64) (combined float64, details map[string]float64) {
	combined, details = ScoreAndRankV2(semantic, bm25, entityBoost, semanticThreshold, semanticOnlyFloor)
	if temporal <= 0 || combined <= 0 {
		details["temporal"] = temporal
		return combined, details
	}
	maxPossible := details["max_possible"]
	if maxPossible <= 0 {
		maxPossible = 1
	}
	raw := combined*maxPossible + temporal
	maxPossible += 1.0
	combined = math.Min(raw/maxPossible, 1.0)
	details["temporal"] = temporal
	details["combined"] = combined
	details["max_possible"] = maxPossible
	return combined, details
}

// CandidateOverfetch returns Mem0-style internal limit: max(limit*4, 60).
func CandidateOverfetch(limit int) int {
	if limit <= 0 {
		limit = 30
	}
	n := limit * 4
	if n < 60 {
		n = 60
	}
	if n > 200 {
		n = 200
	}
	return n
}

// CandidatePoolSize is the retrieval candidate pool at a fixed context budget.
// Explicit CandidateLimit is qualified at 30/50/100/200; default stays overfetch.
func CandidatePoolSize(opts SearchOptions) int {
	if opts.CandidateLimit > 0 {
		n := opts.CandidateLimit
		if n > 200 {
			n = 200
		}
		return n
	}
	return CandidateOverfetch(opts.Limit)
}
