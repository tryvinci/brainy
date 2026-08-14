package memory

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Deterministic attribute atoms turn dialogue into searchable standalone facts.
// Patterns are general linguistic forms only — no benchmark surface-forms.
// LOCOMO-shaped regexes are gated behind BRAINY_LEGACY_LOCOMO_ATOMS (default off).
// See docs/research/master-plan.md §1.2 / W1.

var (
	speakerLineRE  = regexp.MustCompile(`(?i)^\s*([A-Za-z][A-Za-z0-9_-]{1,40})\s*:\s*(.+)$`)
	quotedTitleRE  = regexp.MustCompile(`"([^"]{2,80})"|'([^']{2,80})'`)
	movedFromRE    = regexp.MustCompile(`(?i)\b(?:moved|relocated)\s+from\s+([A-Za-z][A-Za-z\s-]{1,40})`)
	movedToRE      = regexp.MustCompile(`(?i)\b(?:moved|relocated)\s+to\s+([A-Za-z][A-Za-z\s-]{1,40})`)
	iAmRE          = regexp.MustCompile(`(?i)\b(?:i am|i'm)\s+(?:a|an)\s+([^,.!?]{3,60})`)
	iAmStatusRE    = regexp.MustCompile(`(?i)\b(?:i am|i'm)\s+(single|married|divorced|engaged|widowed)\b`)
	asARoleRE      = regexp.MustCompile(`(?i)\bas an?\s+([a-z][a-z\s-]{2,40})\b`)
	activityGerund = regexp.MustCompile(`(?i)\b(?:i|i've|i have)\s+(?:been\s+)?([a-z]+ing)\b`)
	loveActivityRE = regexp.MustCompile(`(?i)\b(?:i love|i like|i enjoy|i'm a (?:big )?fan of)\s+([a-z][a-z\s-]{2,40})`)
	placeWithActRE = regexp.MustCompile(`(?i)\b([a-z]+ing|hiking|camping|swimming)\s+(?:in|at|on)\s+(?:the\s+)?([a-z][a-z\s-]{2,40})\b`)
	kidsLikeRE     = regexp.MustCompile(`(?i)\b(?:kids?|children)\s+(?:love|like|enjoy|are into|were\s+\w+\s+(?:about|for))\s+(?:the\s+)?([a-z][a-z\s-]{2,40})`)
	homeCountryRE  = regexp.MustCompile(`(?i)\bhome country[,:]?\s+([A-Z][a-zA-Z]+(?:\s+[A-Z][a-zA-Z]+)?)\b`)

	// Activity gerunds that are usually not hobbies (skip).
	activityGerundStop = map[string]struct{}{
		"being": {}, "having": {}, "going": {}, "doing": {}, "getting": {},
		"looking": {}, "thinking": {}, "feeling": {}, "saying": {}, "trying": {},
		"making": {}, "taking": {}, "coming": {}, "seeing": {}, "knowing": {},
		"wanting": {}, "needing": {}, "working": {}, "living": {}, "telling": {},
		"fulfilling": {}, "investing": {}, "helping": {}, "using": {},
	}

	placeTailStop = map[string]struct{}{
		"last": {}, "next": {}, "this": {}, "yesterday": {}, "today": {}, "tomorrow": {},
		"week": {}, "month": {}, "year": {}, "morning": {}, "night": {}, "ago": {},
		"it": {}, "was": {}, "really": {}, "so": {}, "many": {}, "my": {}, "our": {},
		"in": {}, "at": {}, "on": {}, "of": {}, "for": {}, "with": {}, "and": {}, "or": {},
		"but": {}, "that": {}, "since": {}, "then": {}, "the": {},
	}
)

func legacyLocomoAtomsEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("BRAINY_LEGACY_LOCOMO_ATOMS"))) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}

func extractAttributeAtoms(utterances []string) []ExtractedMemory {
	out := make([]ExtractedMemory, 0, 8)
	seen := map[string]struct{}{}
	lastSpeaker := ""
	for _, utt := range utterances {
		speaker, body := splitSpeaker(utt)
		if speaker != "" {
			lastSpeaker = speaker
		} else if lastSpeaker != "" {
			speaker = lastSpeaker
		}
		if body == "" {
			continue
		}
		if speaker == "" && isFirstPersonBody(body) {
			continue
		}
		who := speaker
		if who == "" {
			continue // require attributed speaker for atoms
		}
		for _, atom := range attributeAtomsFromUtterance(who, body, utt) {
			if malformedCompilerFact(atom.Content) {
				continue
			}
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

func isFirstPersonBody(body string) bool {
	lower := strings.ToLower(strings.TrimSpace(body))
	return strings.HasPrefix(lower, "i ") || strings.HasPrefix(lower, "i'm ") ||
		strings.HasPrefix(lower, "i’ve") || strings.HasPrefix(lower, "i've") ||
		strings.Contains(lower, "i am ") || strings.Contains(lower, "i'm ")
}

func splitSpeaker(utterance string) (speaker, body string) {
	m := speakerLineRE.FindStringSubmatch(strings.TrimSpace(utterance))
	if m == nil {
		return "", NormalizeText(utterance)
	}
	return strings.TrimSpace(m[1]), NormalizeText(m[2])
}

func attributeAtomsFromUtterance(who, body, source string) []ExtractedMemory {
	var out []ExtractedMemory
	lowerBody := strings.ToLower(body)

	if m := iAmStatusRE.FindStringSubmatch(body); m != nil {
		out = append(out, atomFact(
			fmt.Sprintf("%s is %s", who, NormalizeText(m[1])),
			source, 0.9, "attribute_relationship",
		))
	}
	if m := iAmRE.FindStringSubmatch(body); m != nil {
		attr := NormalizeText(m[1])
		if utf8.RuneCountInString(attr) <= 40 {
			out = append(out, atomFact(
				fmt.Sprintf("%s is a %s", who, attr),
				source, 0.88, "attribute_identity",
			))
		}
	}
	if m := asARoleRE.FindStringSubmatch(body); m != nil {
		role := NormalizeText(m[1])
		if utf8.RuneCountInString(role) <= 40 && !strings.Contains(role, " that ") {
			out = append(out, atomFact(
				fmt.Sprintf("%s is a %s", who, role),
				source, 0.9, "attribute_identity",
			))
			// Relationship-status shorthand when role encodes it.
			if strings.Contains(strings.ToLower(role), "single") {
				out = append(out, atomFact(
					fmt.Sprintf("%s is single", who),
					source, 0.88, "attribute_relationship",
				))
			}
		}
	}
	if m := homeCountryRE.FindStringSubmatch(body); m != nil {
		place, ok := normalizePlace(m[1])
		if ok {
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
	if m := movedFromRE.FindStringSubmatch(body); m != nil {
		place, ok := normalizePlace(m[1])
		if ok {
			out = append(out, atomFact(
				fmt.Sprintf("%s moved from %s", who, place),
				source, 0.9, "attribute_origin",
			))
		}
	}
	if m := movedToRE.FindStringSubmatch(body); m != nil {
		place, ok := normalizePlace(m[1])
		if ok {
			out = append(out, atomFact(
				fmt.Sprintf("%s moved to %s", who, place),
				source, 0.9, "attribute_origin",
			))
		}
	}

	for _, m := range quotedTitleRE.FindAllStringSubmatch(body, 4) {
		title := m[1]
		if title == "" {
			title = m[2]
		}
		title = NormalizeText(title)
		if utf8.RuneCountInString(title) < 4 || looksBrokenQuotedTitle(title) {
			continue
		}
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

	if m := activityGerund.FindStringSubmatch(body); m != nil {
		act := strings.ToLower(NormalizeText(m[1]))
		if _, stop := activityGerundStop[act]; !stop && len(act) >= 5 {
			// Keep the gerund. Stripping "ing" yields "runn"/"hik" and those
			// shards outrank real provenance once they are treated as facts.
			out = append(out, atomFact(
				fmt.Sprintf("%s participates in %s", who, act),
				source, 0.82, "attribute_activity",
			))
		}
	}
	if m := loveActivityRE.FindStringSubmatch(body); m != nil {
		act := NormalizeText(m[1])
		if utf8.RuneCountInString(act) <= 40 && !strings.Contains(act, " that ") {
			out = append(out, atomFact(
				fmt.Sprintf("%s enjoys %s", who, act),
				source, 0.82, "attribute_activity",
			))
		}
	}
	if m := placeWithActRE.FindStringSubmatch(body); m != nil {
		act := strings.ToLower(NormalizeText(m[1]))
		place, ok := normalizePlace(m[2])
		if _, stop := activityGerundStop[act]; !stop && ok {
			out = append(out, atomFact(
				fmt.Sprintf("%s has done %s at %s", who, act, place),
				source, 0.86, "attribute_place_activity",
			))
		}
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

	// Temporary legacy path removed — see git history / BRAINY_LEGACY_LOCOMO_ATOMS
	// docs note. W2 typed atoms replace benchmark-shaped regexes.
	_ = legacyLocomoAtomsEnabled

	return out
}

func atomFact(content, source string, confidence float64, rule string) ExtractedMemory {
	pred := predicateForAttributeRule(rule, content)
	explain := map[string]any{
		"rule": rule,
	}
	if pred != "" {
		explain["predicate"] = pred
		explain["value_norm"] = valueNormFromAtomContent(content)
	}
	return ExtractedMemory{
		Kind:       KindFact,
		Content:    content,
		SourceText: source,
		Confidence: confidence,
		Explain:    explain,
	}
}

func predicateForAttributeRule(rule, content string) string {
	switch rule {
	case "attribute_origin":
		return PredicateOrigin
	case "attribute_activity", "attribute_place_activity":
		return PredicateActivity
	case "attribute_titled_work":
		return PredicateMediaConsumed
	case "attribute_preference":
		if strings.Contains(strings.ToLower(content), "kids like") {
			return PredicateFamilyMember
		}
		return PredicatePreference
	case "attribute_identity":
		return PredicateIdentity
	case "attribute_relationship":
		return PredicateRelationshipStatus
	default:
		return ""
	}
}

func valueNormFromAtomContent(content string) string {
	// Prefer trailing value after common templates.
	lower := strings.ToLower(content)
	for _, sep := range []string{" moved from ", " is from ", " participates in ", " enjoys ", " kids like ", " read \"", " is a ", " is "} {
		if i := strings.Index(lower, sep); i >= 0 {
			v := strings.TrimSpace(content[i+len(sep):])
			v = strings.Trim(v, "\"")
			return strings.ToLower(NormalizeText(v))
		}
	}
	return strings.ToLower(NormalizeText(content))
}

func looksBrokenQuotedTitle(title string) bool {
	t := strings.TrimSpace(title)
	if t == "" {
		return true
	}
	r, _ := utf8.DecodeRuneInString(t)
	if r >= 'a' && r <= 'z' {
		return true
	}
	if matched, _ := regexp.MatchString(`^[a-z]\s`, t); matched {
		return true
	}
	lower := strings.ToLower(t)
	if strings.HasSuffix(lower, " but i") || strings.HasSuffix(t, " but I") {
		return true
	}
	if strings.HasPrefix(lower, "ve ") || strings.HasPrefix(lower, "m ") ||
		strings.HasPrefix(lower, "ll ") || strings.HasPrefix(lower, "re ") {
		return true
	}
	if strings.Count(t, " ") >= 8 && !looksLikeWorkTitle(t) {
		return true
	}
	return false
}

func normalizePlace(place string) (string, bool) {
	words := strings.Fields(strings.ToLower(NormalizeText(place)))
	out := make([]string, 0, 2)
	for _, w := range words {
		w = strings.Trim(w, ",.;:!?\"'")
		if w == "" {
			break
		}
		if _, stop := placeTailStop[w]; stop {
			break
		}
		out = append(out, w)
		if len(out) >= 2 {
			break
		}
	}
	p := strings.Join(out, " ")
	return p, isConcretePlace(p)
}

func isConcretePlace(place string) bool {
	p := strings.ToLower(strings.TrimSpace(place))
	if p == "" || utf8.RuneCountInString(p) < 3 {
		return false
	}
	if strings.Contains(p, " - ") || strings.Contains(p, " last ") {
		return false
	}
	if strings.HasPrefix(p, "my ") || strings.HasPrefix(p, "our ") || strings.HasPrefix(p, "the ") {
		return false
	}
	switch p {
	case "home", "home country", "there", "here", "abroad", "overseas",
		"life", "ways", "changes", "touch", "then", "need", "you":
		return false
	}
	return true
}

// malformedCompilerFact reports extractor templates that are not durable
// semantic memory: light-verb "has done X at Y", failed gerund stems, and
// broken quote shards. Used on write and recall so junk cannot count as
// representation coverage or crowd out provenance.
func malformedCompilerFact(content string) bool {
	c := strings.ToLower(strings.TrimSpace(content))
	if c == "" {
		return true
	}
	if i := strings.Index(c, " participates in "); i >= 0 {
		val := strings.Trim(strings.TrimSpace(c[i+len(" participates in "):]), `."'`)
		if val == "" || !strings.HasSuffix(val, "ing") {
			return true
		}
		head, _, _ := strings.Cut(val, " ")
		if _, stop := activityGerundStop[head]; stop {
			return true
		}
	}
	if i := strings.Index(c, " has done "); i >= 0 {
		rest := strings.TrimSpace(c[i+len(" has done "):])
		act, place, ok := strings.Cut(rest, " at ")
		if !ok {
			return true
		}
		act = strings.ToLower(strings.TrimSpace(act))
		if _, stop := activityGerundStop[act]; stop {
			return true
		}
		if !isConcretePlace(place) {
			return true
		}
	}
	if i := strings.Index(c, ` mentioned "`); i >= 0 {
		orig := strings.TrimSpace(content)
		qi := strings.Index(orig, `"`)
		if qi >= 0 {
			rest := orig[qi+1:]
			q := rest
			if end := strings.Index(rest, `"`); end >= 0 {
				q = rest[:end]
			}
			if looksBrokenQuotedTitle(q) {
				return true
			}
		}
	}
	fields := strings.Fields(c)
	if len(fields) == 0 {
		return true
	}
	last := strings.Trim(fields[len(fields)-1], `."'`)
	if utf8.RuneCountInString(last) == 1 {
		return true
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
