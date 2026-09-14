package memory

import (
	"os"
	"sort"
	"strings"
)

// RRFK is the standard reciprocal-rank constant (Cormack et al.).
const RRFK = 60

// RetrievalRRFEnabled is the versioned retrieval switch for generic RRF
// of dense + lexical ranks. Default off so existing fusion/leftover ranking
// stays unchanged until qualification boots BRAINY_RETRIEVAL_RRF=1.
func RetrievalRRFEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("BRAINY_RETRIEVAL_RRF"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// ReciprocalRankFusion merges rank lists. Rank is 1-based. Duplicate ids in
// the same list count once. Empty lists are skipped.
func ReciprocalRankFusion(lists [][]string, k int) map[string]float64 {
	if k <= 0 {
		k = RRFK
	}
	out := map[string]float64{}
	for _, list := range lists {
		seen := make(map[string]struct{}, len(list))
		rank := 0
		for _, id := range list {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			rank++
			out[id] += 1.0 / (float64(k) + float64(rank))
		}
	}
	return out
}

// RankIDsByScore returns ids with score > 0, highest first. Ties break by id.
func RankIDsByScore(scores map[string]float64) []string {
	if len(scores) == 0 {
		return nil
	}
	type pair struct {
		id string
		s  float64
	}
	ps := make([]pair, 0, len(scores))
	for id, s := range scores {
		if s <= 0 || strings.TrimSpace(id) == "" {
			continue
		}
		ps = append(ps, pair{id: id, s: s})
	}
	sort.Slice(ps, func(i, j int) bool {
		if ps[i].s == ps[j].s {
			return ps[i].id < ps[j].id
		}
		return ps[i].s > ps[j].s
	})
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.id
	}
	return out
}

// rrfToSearchScore maps an RRF mass into Brainy search units so downstream
// boosts and top-k cuts stay in the same numeric band as fusion v2.
func rrfToSearchScore(rrf float64) float64 {
	if rrf <= 0 {
		return 0
	}
	return 1.0 + 60.0*rrf
}
