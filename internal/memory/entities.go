package memory

import (
	"context"
	"regexp"
	"strings"
)

// Entity linking is a generic, domain-agnostic memory technique shared by SOTA
// systems (Mem0 entity linking, Zep/Graphiti entity resolution, A-MEM notes):
// extract salient entities from each memory and boost retrieval when a query
// mentions the same entities. No dataset- or benchmark-specific vocabulary.

var (
	quotedSpanRe   = regexp.MustCompile(`"([^"]{2,80})"|“([^”]{2,80})”|'([^']{2,80})'`)
	properNounRe   = regexp.MustCompile(`\b([A-Z][a-z]+(?:\s+[A-Z][a-z]+){0,3})\b`)
	multiDigitYear = regexp.MustCompile(`\b(1[0-9]{3}|2[0-9]{3})\b`)
)

// speakerPrefix strips a leading "Name:" dialogue speaker label so the speaker
// is treated as an entity but not double-counted as content.
var speakerPrefixRe = regexp.MustCompile(`^\s*([A-Z][a-z]+):\s*`)

// entityStopwords are capitalized words that start sentences but are not entities.
var entityStopwords = map[string]struct{}{
	"the": {}, "a": {}, "an": {}, "i": {}, "we": {}, "you": {}, "he": {}, "she": {},
	"they": {}, "it": {}, "this": {}, "that": {}, "these": {}, "those": {}, "my": {},
	"our": {}, "your": {}, "his": {}, "her": {}, "their": {}, "yes": {}, "no": {},
	"ok": {}, "okay": {}, "wow": {}, "oh": {}, "hey": {}, "hi": {}, "hello": {},
	"so": {}, "well": {}, "now": {}, "then": {}, "there": {}, "here": {}, "when": {},
	"what": {}, "where": {}, "who": {}, "why": {}, "how": {}, "if": {}, "but": {},
	"and": {}, "or": {}, "for": {}, "to": {}, "of": {}, "in": {}, "on": {}, "at": {},
	"congrats": {}, "thanks": {}, "sounds": {}, "guess": {}, "cool": {}, "great": {},
}

// ExtractEntities returns normalized (lowercased) salient entities from text:
// quoted spans, proper-noun phrases, and 4-digit years. Deduplicated, order-stable.
func ExtractEntities(text string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, 8)
	add := func(raw string) {
		e := strings.ToLower(strings.TrimSpace(raw))
		e = strings.Trim(e, ".,!?;:\"'")
		if len(e) < 2 {
			return
		}
		if _, stop := entityStopwords[e]; stop {
			return
		}
		if _, ok := seen[e]; ok {
			return
		}
		seen[e] = struct{}{}
		out = append(out, e)
	}

	// Quoted spans (titles, names, exact phrases).
	for _, m := range quotedSpanRe.FindAllStringSubmatch(text, -1) {
		for _, g := range m[1:] {
			if g != "" {
				add(g)
			}
		}
	}

	// Proper-noun phrases, ignoring a leading speaker label.
	body := speakerPrefixRe.ReplaceAllString(text, "")
	if loc := speakerPrefixRe.FindStringSubmatch(text); loc != nil {
		add(loc[1]) // speaker is a real entity
	}
	for _, m := range properNounRe.FindAllString(body, -1) {
		// Skip if the whole phrase is a single common stopword-ish capitalized word.
		words := strings.Fields(m)
		if len(words) == 1 {
			if _, stop := entityStopwords[strings.ToLower(words[0])]; stop {
				continue
			}
		}
		add(m)
	}

	// Years.
	for _, y := range multiDigitYear.FindAllString(text, -1) {
		add(y)
	}

	return out
}

// entityOverlapBoost rewards candidates whose stored entities intersect the
// query entities. Generic across domains; strengthens person/place/title recall.
func entityOverlapBoost(queryEntities []string, recordEntities []string) float64 {
	if len(queryEntities) == 0 || len(recordEntities) == 0 {
		return 0
	}
	rset := make(map[string]struct{}, len(recordEntities))
	for _, e := range recordEntities {
		rset[e] = struct{}{}
	}
	hits := 0
	for _, qe := range queryEntities {
		if _, ok := rset[qe]; ok {
			hits++
			continue
		}
		// Sub-phrase match: query entity contained in a record entity or vice versa.
		for re := range rset {
			if len(qe) >= 4 && (strings.Contains(re, qe) || strings.Contains(qe, re)) {
				hits++
				break
			}
		}
	}
	if hits == 0 {
		return 0
	}
	boost := 0.2 * float64(hits)
	if boost > 0.6 {
		return 0.6
	}
	return boost
}

// entityDocFrequencies counts, over the subject's active memories, how many
// memories mention each entity. Used to down-weight ubiquitous entities.
func (s *Service) entityDocFrequencies(ctx context.Context, tenantID, subjectID string) (map[string]int, int) {
	df := map[string]int{}
	all, err := s.store.ListActiveMemories(ctx, tenantID, subjectID)
	if err != nil {
		return df, 0
	}
	for _, record := range all {
		seen := map[string]struct{}{}
		for _, e := range recordEntities(record) {
			if _, ok := seen[e]; ok {
				continue
			}
			seen[e] = struct{}{}
			df[e]++
		}
	}
	return df, len(all)
}

// isDistinctiveEntity keeps entities that are not ubiquitous across the subject's
// memories. An entity present in >40% of memories (e.g. a dialogue's speakers) is
// treated as non-distinctive and excluded from entity boosting/admission.
func isDistinctiveEntity(entity string, df map[string]int, total int) bool {
	if total <= 3 {
		return true
	}
	count := df[entity]
	if count == 0 {
		return true
	}
	return float64(count)/float64(total) <= 0.4
}

// recordEntities returns entities persisted on a record (from ingest), falling
// back to on-the-fly extraction from content for older records.
func recordEntities(record MemoryRecord) []string {
	if record.Metadata != nil {
		if raw, ok := record.Metadata["entities"]; ok {
			switch v := raw.(type) {
			case []string:
				return v
			case []any:
				out := make([]string, 0, len(v))
				for _, item := range v {
					if s, ok := item.(string); ok {
						out = append(out, s)
					}
				}
				if len(out) > 0 {
					return out
				}
			}
		}
	}
	return ExtractEntities(record.Content + " " + record.SourceText)
}
