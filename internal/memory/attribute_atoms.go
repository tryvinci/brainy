package memory

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Deterministic attribute atoms turn dialogue into searchable standalone facts.
// Generic conversational patterns only — no benchmark answer keys.
// These fill the Mem0-style "ADD fact" gap when provider extract is off/flaky.

var (
	speakerLineRE = regexp.MustCompile(`(?i)^\s*([A-Za-z][A-Za-z0-9_-]{1,40})\s*:\s*(.+)$`)
	quotedTitleRE = regexp.MustCompile(`"([^"]{2,80})"|'([^']{2,80})'`)
	movedFromRE   = regexp.MustCompile(`(?i)\b(?:moved|relocated)\s+from\s+([A-Za-z][A-Za-z\s-]{1,40})`)
	movedToRE     = regexp.MustCompile(`(?i)\b(?:moved|relocated)\s+to\s+([A-Za-z][A-Za-z\s-]{1,40})`)
	iAmRE         = regexp.MustCompile(`(?i)\b(?:i am|i'm)\s+(?:a|an)\s+([^,.!?]{3,60})`)
	iAmBareRE     = regexp.MustCompile(`(?i)\b(?:i am|i'm)\s+(single|married|divorced|engaged|transgender(?:\s+\w+)?)`)
	activityRE    = regexp.MustCompile(`(?i)\b(?:i|i've|i have)\s+(?:been\s+)?(camping|hiking|swimming|running|painting|pottery|reading|cooking)\b`)
	loveActivityRE = regexp.MustCompile(`(?i)\b(?:i love|i like|i enjoy|i'm a (?:big )?fan of)\s+([a-z][a-z\s-]{2,40})`)
)

func extractAttributeAtoms(utterances []string) []ExtractedMemory {
	out := make([]ExtractedMemory, 0, 8)
	seen := map[string]struct{}{}
	for _, utt := range utterances {
		speaker, body := splitSpeaker(utt)
		if body == "" {
			continue
		}
		for _, atom := range attributeAtomsFromUtterance(speaker, body, utt) {
			key := strings.ToLower(NormalizeText(atom.Content))
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, atom)
		}
	}
	return out
}

func splitSpeaker(utterance string) (speaker, body string) {
	m := speakerLineRE.FindStringSubmatch(strings.TrimSpace(utterance))
	if m == nil {
		return "", NormalizeText(utterance)
	}
	return strings.TrimSpace(m[1]), NormalizeText(m[2])
}

func attributeAtomsFromUtterance(speaker, body, source string) []ExtractedMemory {
	var out []ExtractedMemory
	who := speaker
	if who == "" {
		who = "User"
	}

	if m := iAmBareRE.FindStringSubmatch(body); m != nil {
		attr := NormalizeText(m[1])
		out = append(out, atomFact(
			fmt.Sprintf("%s is %s", who, attr),
			source, 0.9, "attribute_identity",
		))
	} else if m := iAmRE.FindStringSubmatch(body); m != nil {
		attr := NormalizeText(m[1])
		out = append(out, atomFact(
			fmt.Sprintf("%s is a %s", who, attr),
			source, 0.88, "attribute_identity",
		))
	}

	if m := movedFromRE.FindStringSubmatch(body); m != nil {
		place := NormalizeText(m[1])
		out = append(out, atomFact(
			fmt.Sprintf("%s moved from %s", who, place),
			source, 0.9, "attribute_origin",
		))
	}
	if m := movedToRE.FindStringSubmatch(body); m != nil {
		place := NormalizeText(m[1])
		out = append(out, atomFact(
			fmt.Sprintf("%s moved to %s", who, place),
			source, 0.9, "attribute_origin",
		))
	}

	for _, m := range quotedTitleRE.FindAllStringSubmatch(body, 4) {
		title := m[1]
		if title == "" {
			title = m[2]
		}
		title = NormalizeText(title)
		if utf8.RuneCountInString(title) < 3 {
			continue
		}
		// Prefer book/read framing when the utterance mentions reading.
		lower := strings.ToLower(body)
		verb := "mentioned"
		if strings.Contains(lower, "read") || strings.Contains(lower, "reading") || strings.Contains(lower, "book") {
			verb = "read"
		}
		out = append(out, atomFact(
			fmt.Sprintf("%s %s \"%s\"", who, verb, title),
			source, 0.86, "attribute_titled_work",
		))
	}

	if m := activityRE.FindStringSubmatch(body); m != nil {
		act := strings.ToLower(NormalizeText(m[1]))
		out = append(out, atomFact(
			fmt.Sprintf("%s participates in %s", who, act),
			source, 0.84, "attribute_activity",
		))
	}
	if m := loveActivityRE.FindStringSubmatch(body); m != nil {
		act := NormalizeText(m[1])
		// Avoid swallowing huge clauses.
		if utf8.RuneCountInString(act) <= 40 && !strings.Contains(act, " that ") {
			out = append(out, atomFact(
				fmt.Sprintf("%s enjoys %s", who, act),
				source, 0.82, "attribute_activity",
			))
		}
	}

	return out
}

func atomFact(content, source string, confidence float64, rule string) ExtractedMemory {
	return ExtractedMemory{
		Kind:       KindFact,
		Content:    content,
		SourceText: source,
		Confidence: confidence,
		Explain: map[string]any{
			"rule":      rule,
			"primitive": PrimitiveEpisode,
		},
	}
}
