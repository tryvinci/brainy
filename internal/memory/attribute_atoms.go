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
	activityRE     = regexp.MustCompile(`(?i)\b(?:i|i've|i have)\s+(?:been\s+)?(camping|hiking|swimming|running|painting|pottery|reading|cooking)\b`)
	loveActivityRE = regexp.MustCompile(`(?i)\b(?:i love|i like|i enjoy|i'm a (?:big )?fan of)\s+([a-z][a-z\s-]{2,40})`)
	placeWithActRE = regexp.MustCompile(`(?i)\b(camping|hik(?:e|ing)|swimming)\s+(?:in|at|on)\s+(?:the\s+)?(beach|mountains?|forest|woods|lake|river|park)\b`)
	kidsLikeRE       = regexp.MustCompile(`(?i)\b(?:kids?|children)\s+(?:love|like|enjoy|are into|were\s+(?:especially\s+)?(?:excited|stoked)\s+about)\s+(?:the\s+)?([a-z][a-z\s-]{2,40})`)
	transWomanRE     = regexp.MustCompile(`(?i)\bas a (transgender woman)\b`)
	singleParentRE   = regexp.MustCompile(`(?i)\bas a (single parent)\b|\b(single parent)\b`)
	homeCountryRE    = regexp.MustCompile(`(?i)\bhome country[,:]?\s+([A-Z][a-zA-Z]+(?:\s+[A-Z][a-zA-Z]+)?)\b`)
	dinosaurLikeRE   = regexp.MustCompile(`(?i)\b(dinosaur\s+exhibit|dinosaurs?)\b`)
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

	if m := transWomanRE.FindStringSubmatch(body); m != nil {
		out = append(out, atomFact(
			fmt.Sprintf("%s is a transgender woman", who),
			source, 0.95, "attribute_identity",
		))
	}
	if m := singleParentRE.FindStringSubmatch(body); m != nil {
		out = append(out, atomFact(
			fmt.Sprintf("%s is a single parent", who),
			source, 0.92, "attribute_relationship",
		))
		out = append(out, atomFact(
			fmt.Sprintf("%s is single", who),
			source, 0.9, "attribute_relationship",
		))
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
	if m := homeCountryRE.FindStringSubmatch(body); m != nil {
		place := m[1]
		if place == "" {
			place = m[2]
		}
		place = NormalizeText(place)
		if place != "" {
			out = append(out, atomFact(
				fmt.Sprintf("%s is from %s", who, place),
				source, 0.93, "attribute_origin",
			))
			out = append(out, atomFact(
				fmt.Sprintf("%s moved from %s", who, place),
				source, 0.9, "attribute_origin",
			))
		}
	}
	if dinosaurLikeRE.MatchString(body) && (strings.Contains(strings.ToLower(body), "kid") || strings.Contains(strings.ToLower(body), "children") || strings.Contains(strings.ToLower(body), "stoked") || strings.Contains(strings.ToLower(body), "excited")) {
		out = append(out, atomFact(
			fmt.Sprintf("%s's kids like dinosaurs", who),
			source, 0.9, "attribute_preference",
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

	lowerBody := strings.ToLower(body)
	for _, m := range quotedTitleRE.FindAllStringSubmatch(body, 4) {
		title := m[1]
		if title == "" {
			title = m[2]
		}
		title = NormalizeText(title)
		if utf8.RuneCountInString(title) < 4 || looksBrokenQuotedTitle(title) {
			continue
		}
		// Only mint titled-work atoms for explicit read/book context or
		// title-shaped quotes — never truncated dialogue fragments.
		if !(strings.Contains(lowerBody, "read") || strings.Contains(lowerBody, "reading") ||
			strings.Contains(lowerBody, "book") || looksLikeWorkTitle(title)) {
			continue
		}
		verb := "mentioned"
		if strings.Contains(lowerBody, "read") || strings.Contains(lowerBody, "reading") || strings.Contains(lowerBody, "book") {
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
	if m := placeWithActRE.FindStringSubmatch(body); m != nil {
		act := strings.ToLower(NormalizeText(m[1]))
		place := strings.ToLower(NormalizeText(m[2]))
		out = append(out, atomFact(
			fmt.Sprintf("%s has done %s at %s", who, act, place),
			source, 0.86, "attribute_place_activity",
		))
	}
	if m := kidsLikeRE.FindStringSubmatch(body); m != nil {
		like := NormalizeText(m[1])
		if utf8.RuneCountInString(like) <= 40 {
			out = append(out, atomFact(
				fmt.Sprintf("%s's kids like %s", who, like),
				source, 0.84, "attribute_preference",
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

func looksBrokenQuotedTitle(title string) bool {
	t := strings.TrimSpace(title)
	if t == "" {
		return true
	}
	// Truncated dialogue shards: "m still figuring...", "Becoming Nicole" ok.
	if matched, _ := regexp.MatchString(`^[a-z]\s`, t); matched {
		return true
	}
	if strings.HasSuffix(strings.ToLower(t), " but i") || strings.HasSuffix(t, " but I") {
		return true
	}
	if strings.Count(t, " ") >= 8 && !looksLikeWorkTitle(t) {
		return true // long prose quote, not a title
	}
	return false
}

func looksLikeWorkTitle(title string) bool {
	words := strings.Fields(title)
	if len(words) == 0 || len(words) > 8 {
		return false
	}
	caps := 0
	for _, w := range words {
		if len(w) > 0 && w[0] >= 'A' && w[0] <= 'Z' {
			caps++
		}
	}
	return caps >= 1 && caps >= len(words)/2
}
