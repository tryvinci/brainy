package memory

import (
	"strings"
	"time"
	"unicode"
)

// CoverageClaim is one durable fact a compiler should emit from dialogue.
// Claims are held-out fixtures, never LoCoMo/LME surface forms.
type CoverageClaim struct {
	Text      string
	Subject   string
	MustDate  bool
	Predicate string
}

// CoverageReport is the R0/S1a semantic-coverage score.
type CoverageReport struct {
	Claims      int
	Compiled    int
	EntityBound int
	Dated       int
	Evidenced   int
}

// CompiledRate is compiled / claims.
func (r CoverageReport) CompiledRate() float64 {
	if r.Claims == 0 {
		return 0
	}
	return float64(r.Compiled) / float64(r.Claims)
}

// ScoreSemanticCoverage scores % durable claims compiled / entity-bound / dated / evidenced.
func ScoreSemanticCoverage(memories []ExtractedMemory, claims []CoverageClaim, observedAt *time.Time) CoverageReport {
	out := CoverageReport{Claims: len(claims)}
	blobs := make([]string, 0, len(memories))
	for _, m := range memories {
		blobs = append(blobs, strings.ToLower(m.Content+" "+m.SourceText))
	}
	joined := strings.Join(blobs, " | ")
	for _, claim := range claims {
		needles := coverageNeedles(claim.Text)
		if !coverageContainsAll(joined, needles) {
			continue
		}
		out.Compiled++
		hits := findCoverageMemories(memories, needles)
		if len(hits) == 0 {
			continue
		}
		bound, dated, evidenced := false, false, false
		for _, hit := range hits {
			if coverageEntityBound(hit, claim.Subject) {
				bound = true
			}
			if coverageDated(hit, observedAt) {
				dated = true
			}
			if coverageEvidenced(hit) {
				evidenced = true
			}
		}
		if bound {
			out.EntityBound++
		}
		if dated {
			out.Dated++
		}
		if evidenced {
			out.Evidenced++
		}
	}
	return out
}

func coverageNeedles(text string) []string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(text)))
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.Trim(f, ".,;:!?\"'`")
		if len(f) < 3 {
			continue
		}
		if coverageStop[f] {
			continue
		}
		out = append(out, f)
	}
	return out
}

func coverageContainsAll(blob string, needles []string) bool {
	if len(needles) == 0 {
		return false
	}
	for _, n := range needles {
		if !strings.Contains(blob, n) {
			return false
		}
	}
	return true
}

func findCoverageMemories(memories []ExtractedMemory, needles []string) []ExtractedMemory {
	out := make([]ExtractedMemory, 0, 2)
	for i := range memories {
		blob := strings.ToLower(memories[i].Content + " " + memories[i].SourceText)
		hit := 0
		for _, n := range needles {
			if strings.Contains(blob, n) {
				hit++
			}
		}
		if hit == len(needles) || hit >= 2 {
			out = append(out, memories[i])
		}
	}
	return out
}

func coverageEntityBound(m ExtractedMemory, subject string) bool {
	got, _ := m.Explain["subject"].(string)
	got = strings.TrimSpace(got)
	if got == "" {
		return false
	}
	if subject == "" {
		return true
	}
	return strings.EqualFold(got, subject) || strings.Contains(strings.ToLower(got), strings.ToLower(subject))
}

func coverageDated(m ExtractedMemory, observedAt *time.Time) bool {
	if strings.TrimSpace(m.When) != "" {
		return true
	}
	if m.Explain != nil {
		if v, _ := m.Explain["when"].(string); strings.TrimSpace(v) != "" {
			return true
		}
		if v, _ := m.Explain["event_start"].(string); strings.TrimSpace(v) != "" {
			return true
		}
	}
	if observedAt != nil && !observedAt.IsZero() && hasDateToken(m.Content) {
		return true
	}
	return hasDateToken(m.Content)
}

func coverageEvidenced(m ExtractedMemory) bool {
	return strings.TrimSpace(m.SourceText) != ""
}

func hasDateToken(s string) bool {
	for _, tok := range strings.Fields(s) {
		digits := 0
		for _, r := range tok {
			if unicode.IsDigit(r) {
				digits++
			}
		}
		if digits >= 2 {
			return true
		}
	}
	low := strings.ToLower(s)
	for _, m := range []string{"january", "february", "march", "april", "june", "july", "august", "september", "october", "november", "december"} {
		if strings.Contains(low, m) {
			return true
		}
	}
	return false
}

var coverageStop = map[string]bool{
	"the": true, "and": true, "for": true, "from": true, "with": true,
	"that": true, "this": true, "has": true, "have": true, "was": true,
	"are": true, "his": true, "her": true, "their": true, "been": true,
}
