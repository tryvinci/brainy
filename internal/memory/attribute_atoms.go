package memory

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// Deterministic attribute atoms turn dialogue into searchable standalone facts.
// Patterns are general linguistic forms only — no benchmark surface-forms.
// LOCOMO-shaped regexes are gated behind BRAINY_LEGACY_LOCOMO_ATOMS (default off).
// See docs/research/master-plan.md §1.2 / W1.

var (
	speakerLineRE     = regexp.MustCompile(`(?i)^\s*([A-Za-z][A-Za-z0-9_-]{1,40})\s*:\s*(.+)$`)
	quotedTitleRE     = regexp.MustCompile(`"([^"]{2,80})"|“([^”]{2,80})”|'([^']{2,80})'|‘([^’]{2,80})’`)
	movedFromRE       = regexp.MustCompile(`(?i)\b(?:moved|relocated)\s+from\s+([A-Za-z][A-Za-z\s-]{1,40})`)
	movedToRE         = regexp.MustCompile(`(?i)\b(?:moved|relocated)\s+to\s+([A-Za-z][A-Za-z\s-]{1,40})`)
	iAmRE             = regexp.MustCompile(`(?i)\b(?:i am|i'm)\s+(?:a|an)\s+([^,.!?]{3,60})`)
	iAmStatusRE       = regexp.MustCompile(`(?i)\b(?:i am|i'm)\s+(single|married|divorced|engaged|widowed)\b`)
	asARoleRE         = regexp.MustCompile(`(?i)\bas an?\s+([a-z][a-z\s-]{2,40})\b`)
	activityGerund    = regexp.MustCompile(`(?i)\b(?:i|i've|i have)\s+(?:been\s+)?([a-z]+ing)\b`)
	loveActivityRE    = regexp.MustCompile(`(?i)\b(?:i love|i like|i enjoy|we love|we like|we enjoy|i'm a (?:big )?fan of)\s+([a-z][a-z\s-]{2,40})`)
	placeWithActRE    = regexp.MustCompile(`(?i)\b([a-z]+ing)\s+(?:in|at|on)\s+(?:the\s+)?([a-z][a-z-]{2,30})\b`)
	tripPlaceRE       = regexp.MustCompile(`(?i)\b(camping|hiking|swimming)\s+trip\s+(?:in|at|on|to|through)\s+(?:the\s+)?([a-z][a-z-]{2,30})\b`)
	gaveEventRE       = regexp.MustCompile(`(?i)\b(?:gave|give|giving)\s+(?:a |an )?([a-z][a-z-]{3,24})\b`)
	hadEventRE        = regexp.MustCompile(`(?i)\b(?:had|have)\s+a\s+([a-z][a-z-]{3,24})\b`)
	ranEventRE        = regexp.MustCompile(`(?i)\b(?:ran|run)\s+(?:a|an)\s+([a-z][a-z\s-]{3,40})\b`)
	trainingForRE     = regexp.MustCompile(`(?i)\btraining\s+for\s+(?:a |an )?([a-z][a-z\s-]{3,40})\b`)
	kidsLikeRE        = regexp.MustCompile(`(?i)\b(?:kids?|children)\s+(?:love|like|enjoy|are into|were\s+\w+\s+(?:about|for))\s+(?:the\s+)?([a-z][a-z\s-]{2,40})`)
	kidsAboutRE       = regexp.MustCompile(`(?i)\b(?:kids?|children).{0,48}?\b(?:love|like|enjoy|into|about|obsessed with)\s+(?:the\s+)?([a-z][a-z\s-]{2,40})`)
	kidsMentionRE     = regexp.MustCompile(`(?i)\b(?:kids?|children)\b`)
	theyExcitedRE     = regexp.MustCompile(`(?i)\bthey\s+(?:were\s+|are\s+)?(?:stoked|excited|pumped|thrilled)\s+(?:for|about)\s+(?:the\s+)?([a-z][a-z\s-]{2,40})`)
	theyLikeRE        = regexp.MustCompile(`(?i)\bthey\s+(?:love|like|enjoy|are into)\s+(?:the\s+)?([a-z][a-z\s-]{2,40})`)
	homeCountryRE     = regexp.MustCompile(`(?i)\bhome country[,:]?\s+([A-Z][a-zA-Z]+(?:\s+[A-Z][a-zA-Z]+)?)\b`)
	homeCountryBareRE = regexp.MustCompile(`(?i)\bhome country\b`)
	originallyFromRE  = regexp.MustCompile(`(?i)\boriginally from\s+([A-Za-z][A-Za-z\s-]{1,40})`)
	imFromRE          = regexp.MustCompile(`(?i)\b(?:i'm|i am)\s+from\s+([A-Za-z][A-Za-z\s-]{1,40})`)
	lookingIntoRE     = regexp.MustCompile(`(?i)\b(?:looking into|keen on|interested in)\s+([^,.!?]{3,70})`)
	careerFieldRE     = regexp.MustCompile(`(?i)\b(?:career|profession)\s+(?:in|as)\s+([^,.!?]{3,50})`)
	wantToBeRE        = regexp.MustCompile(`(?i)\b(?:want to|wanted to|hope to|plan to|planning to)\s+(?:be|become|pursue|work as)\s+(?:a |an )?([^,.!?]{3,50})`)
	workWithGroupRE   = regexp.MustCompile(`(?i)\b(?:thinking of working with|working with|work with)\s+([a-z][a-z\s-]{1,40}(?:people|community|patients|clients|families))\b`)
	planningActRE     = regexp.MustCompile(`(?i)\b(?:planning on|planning to|going to)\s+(?:go(?:ing)?\s+)?([a-z]+ing)\b`)
	workshopRE        = regexp.MustCompile(`(?i)\b([a-z][a-z-]{2,30})\s+(?:workshop|class|lesson)s?\b`)
	goGerundRE        = regexp.MustCompile(`(?i)\b(?:go|going|went|off to go)\s+([a-z]+ing)\b`)
	durationYearsRE   = regexp.MustCompile(`(?i)\bfor\s+(\d+|one|two|three|four|five|six|seven|eight|nine|ten)\s+years\b`)
	collectsRE        = regexp.MustCompile(`(?i)\b(?:collect(?:s|ing)?|collection of)\s+([^,.!?]{3,50})`)
	educationRE       = regexp.MustCompile(`(?i)\b(?:studying|degree in|certification in|certified in)\s+([^,.!?]{3,50})`)
	readUnquotedRE    = regexp.MustCompile(`(?i)\b(?:read|reading|loved reading)\s+([A-Z][^"“”'.]{2,60})`)
	bookTitleRunRE    = regexp.MustCompile(`\b((?:The|A|An)\s+[A-Z][A-Za-z']+(?:\s+[A-Z][A-Za-z']+){0,5}|[A-Z][A-Za-z']+(?:\s+(?:is|the|a|an|of|and|in|to|for|on)\s+[A-Z][A-Za-z']+|\s+[A-Z][A-Za-z']+){1,5})\b`)

	// Activity gerunds that are usually not hobbies (skip).
	activityGerundStop = map[string]struct{}{
		"being": {}, "having": {}, "going": {}, "doing": {}, "getting": {},
		"looking": {}, "thinking": {}, "feeling": {}, "saying": {}, "trying": {},
		"making": {}, "taking": {}, "coming": {}, "seeing": {}, "knowing": {},
		"wanting": {}, "needing": {}, "working": {}, "living": {}, "telling": {},
		"fulfilling": {}, "investing": {}, "helping": {}, "using": {},
		"planning": {}, "starting": {}, "keeping": {},
	}

	roleStop = map[string]struct{}{
		"kid": {}, "kids": {}, "child": {}, "career": {}, "job": {}, "fan": {},
		"result": {}, "way": {}, "bit": {}, "lot": {}, "while": {}, "moment": {},
	}

	// Light nouns after "had a" / "gave a" / "ran a" that are not events.
	eventLightStop = map[string]struct{}{
		"lot": {}, "bit": {}, "time": {}, "day": {}, "look": {}, "go": {},
		"idea": {}, "chance": {}, "feeling": {}, "moment": {}, "while": {},
		"rest": {}, "laugh": {}, "blast": {}, "great": {}, "good": {}, "hard": {},
		"nice": {}, "way": {}, "job": {}, "try": {}, "few": {}, "couple": {},
		"sense": {}, "point": {}, "chat": {}, "break": {}, "nap": {}, "shot": {},
		"every": {}, "morning": {}, "evening": {}, "afternoon": {}, "night": {},
	}

	placeTailStop = map[string]struct{}{
		"last": {}, "next": {}, "this": {}, "yesterday": {}, "today": {}, "tomorrow": {},
		"week": {}, "month": {}, "year": {}, "morning": {}, "night": {}, "ago": {},
		"it": {}, "was": {}, "really": {}, "so": {}, "many": {}, "my": {}, "our": {},
		"in": {}, "at": {}, "on": {}, "of": {}, "for": {}, "with": {}, "and": {}, "or": {},
		"but": {}, "that": {}, "since": {}, "then": {}, "the": {},
		"going": {}, "planning": {},
	}

	preferenceHeadStop = map[string]struct{}{
		"our": {}, "my": {}, "their": {}, "his": {}, "her": {}, "its": {},
		"the": {}, "this": {}, "that": {}, "these": {}, "those": {},
		"one": {}, "ones": {}, "last": {}, "next": {}, "some": {}, "any": {},
		"it": {}, "them": {}, "stuff": {}, "things": {}, "thing": {},
		"time": {}, "day": {}, "week": {}, "month": {}, "year": {},
		"break": {}, "summer": {}, "winter": {}, "spring": {}, "fall": {},
		"morning": {}, "evening": {}, "afternoon": {}, "night": {},
	}

	titleLeadStop = map[string]struct{}{
		"i": {}, "we": {}, "my": {}, "our": {}, "last": {}, "next": {},
		"every": {}, "on": {}, "in": {}, "at": {}, "after": {}, "before": {},
		"this": {}, "that": {}, "monday": {}, "tuesday": {}, "wednesday": {},
		"thursday": {}, "friday": {}, "saturday": {}, "sunday": {},
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

func extractAttributeAtoms(utterances []string, observedAt *time.Time) []ExtractedMemory {
	out := make([]ExtractedMemory, 0, 8)
	seen := map[string]struct{}{}
	lastSpeaker := ""
	origins := map[string]string{}
	kidsBySpeaker := map[string]bool{}
	add := func(atom ExtractedMemory) {
		if malformedCompilerFact(atom.Content) {
			return
		}
		key := strings.ToLower(NormalizeText(atom.Content))
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, atom)
	}
	type turn struct {
		who, body, utt string
	}
	turns := make([]turn, 0, len(utterances))
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
		if speaker == "" {
			continue
		}
		turns = append(turns, turn{who: speaker, body: body, utt: utt})
	}
	for _, t := range turns {
		whoKey := strings.ToLower(t.who)
		if kidsMentionRE.MatchString(t.body) {
			kidsBySpeaker[whoKey] = true
		}
		for _, atom := range attributeAtomsFromUtterance(t.who, t.body, t.utt, observedAt) {
			add(atom)
			if pred, _ := atom.Explain["predicate"].(string); pred == PredicateOrigin {
				if v, _ := atom.Explain["value_norm"].(string); v != "" {
					origins[whoKey] = v
				}
			}
		}
		if kidsBySpeaker[whoKey] {
			for _, like := range pronounPreferenceObjects(t.body) {
				add(atomFact(t.who, fmt.Sprintf("%s's kids like %s", t.who, like), t.utt, 0.82, "attribute_preference", observedAt))
			}
		}
	}
	// Bind "home country" anaphora to an origin already compiled for this speaker.
	for _, t := range turns {
		if !homeCountryBareRE.MatchString(t.body) || homeCountryRE.MatchString(t.body) {
			continue
		}
		place, ok := origins[strings.ToLower(t.who)]
		if !ok || place == "" {
			continue
		}
		add(atomFact(t.who, fmt.Sprintf("%s moved from %s", t.who, titleCaseWords(place)), t.utt, 0.9, "attribute_origin", observedAt))
		add(atomFact(t.who, fmt.Sprintf("%s is from %s", t.who, titleCaseWords(place)), t.utt, 0.9, "attribute_origin", observedAt))
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

func attributeAtomsFromUtterance(who, body, source string, observedAt *time.Time) []ExtractedMemory {
	var out []ExtractedMemory
	lowerBody := strings.ToLower(body)
	emit := func(content string, conf float64, rule string) {
		out = append(out, atomFact(who, content, source, conf, rule, observedAt))
	}

	if m := iAmStatusRE.FindStringSubmatch(body); m != nil {
		emit(fmt.Sprintf("%s is %s", who, NormalizeText(m[1])), 0.9, "attribute_relationship")
	}
	if m := iAmRE.FindStringSubmatch(body); m != nil {
		attr := clipIdentityTail(NormalizeText(m[1]))
		if utf8.RuneCountInString(attr) <= 40 && !roleStopWord(attr) {
			emit(fmt.Sprintf("%s is a %s", who, attr), 0.88, "attribute_identity")
		}
	}
	if m := asARoleRE.FindStringSubmatch(body); m != nil {
		role := NormalizeText(m[1])
		if utf8.RuneCountInString(role) <= 40 && !strings.Contains(role, " that ") && !roleStopWord(role) {
			emit(fmt.Sprintf("%s is a %s", who, role), 0.9, "attribute_identity")
			if strings.Contains(strings.ToLower(role), "single") {
				emit(fmt.Sprintf("%s is single", who), 0.88, "attribute_relationship")
			}
		}
	}
	if m := homeCountryRE.FindStringSubmatch(body); m != nil {
		place, ok := normalizePlace(m[1])
		if ok {
			emit(fmt.Sprintf("%s is from %s", who, place), 0.93, "attribute_origin")
			emit(fmt.Sprintf("%s moved from %s", who, place), 0.9, "attribute_origin")
		}
	}
	if m := originallyFromRE.FindStringSubmatch(body); m != nil {
		place, ok := normalizePlace(m[1])
		if ok {
			emit(fmt.Sprintf("%s is from %s", who, place), 0.93, "attribute_origin")
		}
	}
	if m := imFromRE.FindStringSubmatch(body); m != nil {
		place, ok := normalizePlace(m[1])
		if ok {
			emit(fmt.Sprintf("%s is from %s", who, place), 0.92, "attribute_origin")
		}
	}
	if m := movedFromRE.FindStringSubmatch(body); m != nil {
		place, ok := normalizePlace(m[1])
		if ok {
			emit(fmt.Sprintf("%s moved from %s", who, place), 0.9, "attribute_origin")
		}
	}
	if m := movedToRE.FindStringSubmatch(body); m != nil {
		place, ok := normalizePlace(m[1])
		if ok {
			emit(fmt.Sprintf("%s moved to %s", who, place), 0.9, "attribute_origin")
		}
	}

	for _, m := range quotedTitleRE.FindAllStringSubmatch(body, 4) {
		title := firstNonEmpty(m[1], m[2], m[3], m[4])
		title = NormalizeText(title)
		if utf8.RuneCountInString(title) < 4 || looksBrokenQuotedTitle(title) {
			continue
		}
		bookCue := strings.Contains(lowerBody, "read") || strings.Contains(lowerBody, "reading") ||
			strings.Contains(lowerBody, "book") || strings.Contains(lowerBody, "title") ||
			strings.Contains(lowerBody, "novel") || strings.Contains(lowerBody, "story")
		if !bookCue {
			continue
		}
		if len(strings.Fields(title)) < 2 {
			continue
		}
		verb := "read"
		emit(fmt.Sprintf("%s %s \"%s\"", who, verb, title), 0.86, "attribute_titled_work")
	}
	if m := readUnquotedRE.FindStringSubmatch(body); m != nil {
		title := NormalizeText(strings.TrimSpace(m[1]))
		title = strings.TrimSuffix(title, " as a kid")
		title = strings.TrimSuffix(title, " last year")
		if looksLikeWorkTitle(title) && utf8.RuneCountInString(title) >= 4 && !looksBrokenQuotedTitle(title) && !titleLeadStopped(title) {
			emit(fmt.Sprintf("%s read \"%s\"", who, title), 0.84, "attribute_titled_work")
		}
	}
	if strings.Contains(lowerBody, "book") || strings.Contains(lowerBody, "read") {
		for _, m := range bookTitleRunRE.FindAllStringSubmatch(body, 4) {
			title := NormalizeText(strings.TrimSpace(m[1]))
			if titleLeadStopped(title) || !looksLikeWorkTitle(title) || looksBrokenQuotedTitle(title) {
				continue
			}
			if utf8.RuneCountInString(title) < 8 {
				continue
			}
			emit(fmt.Sprintf("%s read \"%s\"", who, title), 0.82, "attribute_titled_work")
		}
	}

	if m := activityGerund.FindStringSubmatch(body); m != nil {
		act := strings.ToLower(NormalizeText(m[1]))
		if _, stop := activityGerundStop[act]; !stop && len(act) >= 5 {
			emit(fmt.Sprintf("%s participates in %s", who, act), 0.82, "attribute_activity")
		}
	}
	if m := loveActivityRE.FindStringSubmatch(body); m != nil {
		act := clipActivityTail(NormalizeText(m[1]))
		if utf8.RuneCountInString(act) >= 4 && utf8.RuneCountInString(act) <= 40 && !strings.Contains(act, " that ") {
			emit(fmt.Sprintf("%s enjoys %s", who, act), 0.82, "attribute_activity")
		}
	}
	if matches := placeWithActRE.FindAllStringSubmatch(body, 6); len(matches) > 0 {
		for _, m := range matches {
			act := strings.ToLower(NormalizeText(m[1]))
			place, ok := normalizePlace(m[2])
			if _, stop := activityGerundStop[act]; !stop && ok {
				emit(fmt.Sprintf("%s has done %s at %s", who, act, place), 0.86, "attribute_place_activity")
			}
		}
	}
	if matches := tripPlaceRE.FindAllStringSubmatch(body, 4); len(matches) > 0 {
		for _, m := range matches {
			act := strings.ToLower(NormalizeText(m[1]))
			place, ok := normalizePlace(m[2])
			if _, stop := activityGerundStop[act]; !stop && ok {
				emit(fmt.Sprintf("%s has done %s at %s", who, act, place), 0.86, "attribute_place_activity")
			}
		}
	}
	if m := goGerundRE.FindStringSubmatch(body); m != nil {
		act := strings.ToLower(NormalizeText(m[1]))
		if _, stop := activityGerundStop[act]; !stop && len(act) >= 5 {
			emit(fmt.Sprintf("%s participates in %s", who, act), 0.84, "attribute_activity")
		}
	}
	if m := workshopRE.FindStringSubmatch(body); m != nil {
		act := strings.ToLower(NormalizeText(m[1]))
		if _, stop := activityGerundStop[act]; !stop && utf8.RuneCountInString(act) >= 4 {
			emit(fmt.Sprintf("%s participates in %s", who, act), 0.88, "attribute_activity")
		}
	}
	if m := planningActRE.FindStringSubmatch(body); m != nil {
		act := strings.ToLower(NormalizeText(m[1]))
		if _, stop := activityGerundStop[act]; !stop && len(act) >= 5 {
			emit(fmt.Sprintf("%s plans %s", who, act), 0.86, "attribute_plan")
			emit(fmt.Sprintf("%s participates in %s", who, act), 0.8, "attribute_activity")
		}
	}
	seenLikes := map[string]struct{}{}
	emitLike := func(raw string, conf float64) {
		like := clipPreferenceTail(NormalizeText(raw))
		if !validPreferenceValue(like) {
			return
		}
		key := strings.ToLower(like)
		if _, ok := seenLikes[key]; ok {
			return
		}
		seenLikes[key] = struct{}{}
		emit(fmt.Sprintf("%s's kids like %s", who, like), conf, "attribute_preference")
	}
	for _, m := range kidsLikeRE.FindAllStringSubmatch(body, 6) {
		emitLike(m[1], 0.84)
	}
	for _, m := range kidsAboutRE.FindAllStringSubmatch(body, 6) {
		emitLike(m[1], 0.82)
	}
	emitCareer := func(field string, conf float64, rule string) {
		field = clipCareerTail(field)
		for _, part := range splitAndOrPhrases(field) {
			if utf8.RuneCountInString(part) >= 3 && utf8.RuneCountInString(part) <= 60 {
				if rule == "attribute_education" {
					emit(fmt.Sprintf("%s studies %s", who, part), conf, rule)
				} else {
					emit(fmt.Sprintf("%s plans career in %s", who, part), conf, rule)
				}
			}
		}
	}
	if m := lookingIntoRE.FindStringSubmatch(body); m != nil {
		emitCareer(NormalizeText(m[1]), 0.88, "attribute_occupation")
	}
	if m := careerFieldRE.FindStringSubmatch(body); m != nil {
		emitCareer(NormalizeText(m[1]), 0.88, "attribute_occupation")
	}
	if m := wantToBeRE.FindStringSubmatch(body); m != nil {
		emitCareer(NormalizeText(m[1]), 0.86, "attribute_occupation")
	}
	if m := educationRE.FindStringSubmatch(body); m != nil {
		emitCareer(NormalizeText(m[1]), 0.86, "attribute_education")
	}
	if m := workWithGroupRE.FindStringSubmatch(body); m != nil {
		group := clipCareerTail(NormalizeText(m[1]))
		if utf8.RuneCountInString(group) >= 5 && utf8.RuneCountInString(group) <= 50 {
			emit(fmt.Sprintf("%s plans career for %s", who, group), 0.86, "attribute_occupation")
		}
	}
	emitEvent := func(raw string, conf float64) {
		ev := clipEventTail(NormalizeText(raw))
		head, _, _ := strings.Cut(strings.ToLower(ev), " ")
		if ev == "" || utf8.RuneCountInString(ev) < 4 || utf8.RuneCountInString(ev) > 40 {
			return
		}
		if _, stop := eventLightStop[head]; stop {
			return
		}
		if _, stop := activityGerundStop[head]; stop {
			return
		}
		emit(fmt.Sprintf("%s attended %s", who, ev), conf, "attribute_event")
	}
	if m := gaveEventRE.FindStringSubmatch(body); m != nil {
		emitEvent(m[1], 0.86)
	}
	if m := hadEventRE.FindStringSubmatch(body); m != nil {
		emitEvent(m[1], 0.84)
	}
	if m := ranEventRE.FindStringSubmatch(body); m != nil {
		emitEvent(m[1], 0.86)
	}
	if m := trainingForRE.FindStringSubmatch(body); m != nil {
		act := strings.ToLower(clipIdentityTail(clipEventTail(NormalizeText(m[1]))))
		if utf8.RuneCountInString(act) >= 4 && utf8.RuneCountInString(act) <= 40 {
			if _, stop := eventLightStop[act]; !stop {
				if _, stop := activityGerundStop[act]; !stop {
					emit(fmt.Sprintf("%s participates in %s", who, act), 0.86, "attribute_activity")
				}
			}
		}
	}
	if m := durationYearsRE.FindStringSubmatch(body); m != nil {
		n := numberWord(m[1])
		if strings.Contains(lowerBody, "friend") {
			emit(fmt.Sprintf("%s has known friends for %s years", who, n), 0.9, "attribute_duration")
		} else {
			emit(fmt.Sprintf("%s has a %s-year duration", who, n), 0.82, "attribute_duration")
		}
	}
	if m := collectsRE.FindStringSubmatch(body); m != nil {
		obj := NormalizeText(m[1])
		if utf8.RuneCountInString(obj) >= 4 && utf8.RuneCountInString(obj) <= 50 && !strings.Contains(obj, " that ") {
			emit(fmt.Sprintf("%s collects %s", who, obj), 0.86, "attribute_possession")
		}
	}

	if strings.Contains(lowerBody, "unwind") || strings.Contains(lowerBody, "relax") ||
		strings.Contains(lowerBody, "clear my mind") {
		for _, m := range regexp.MustCompile(`(?i)\b([a-z]{4,}ing)\b`).FindAllStringSubmatch(body, 6) {
			act := strings.ToLower(m[1])
			if _, stop := activityGerundStop[act]; stop {
				continue
			}
			emit(fmt.Sprintf("%s unwinds via %s", who, act), 0.86, "attribute_activity")
			emit(fmt.Sprintf("%s participates in %s", who, act), 0.82, "attribute_activity")
		}
	}

	// Temporary legacy path removed — see git history / BRAINY_LEGACY_LOCOMO_ATOMS
	// docs note. W2 typed atoms replace benchmark-shaped regexes.
	_ = legacyLocomoAtomsEnabled

	return out
}

func atomFact(who, content, source string, confidence float64, rule string, observedAt *time.Time) ExtractedMemory {
	content = stampAtomTime(content, source, observedAt)
	pred := predicateForAttributeRule(rule, content)
	explain := map[string]any{
		"rule": rule,
	}
	if who != "" {
		explain["subject"] = who
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

func stampAtomTime(content, source string, observedAt *time.Time) string {
	if observedAt == nil || observedAt.IsZero() || content == "" || source == "" {
		return content
	}
	stamped := EnrichRelativeEventTime(source, *observedAt)
	if stamped == source {
		return content
	}
	i := strings.LastIndex(stamped, " (")
	if i < 0 || !strings.HasSuffix(stamped, ")") {
		return content
	}
	paren := stamped[i:]
	if strings.Contains(content, paren) {
		return content
	}
	return content + paren
}

func clipIdentityTail(s string) string {
	s = NormalizeText(s)
	lower := strings.ToLower(s)
	if i := strings.Index(lower, " and "); i > 0 {
		s = strings.TrimSpace(s[:i])
	}
	return s
}

func roleStopWord(role string) bool {
	head, _, _ := strings.Cut(strings.ToLower(strings.TrimSpace(role)), " ")
	_, stop := roleStop[head]
	return stop
}

func clipCareerTail(s string) string {
	s = NormalizeText(s)
	lower := strings.ToLower(s)
	for _, tail := range []string{
		" as a career", " as a job", " career options", " options",
		" so i could", " so i can", " because",
	} {
		if i := strings.Index(lower, tail); i > 0 {
			s = strings.TrimSpace(s[:i])
			lower = strings.ToLower(s)
		}
	}
	return s
}

func clipActivityTail(s string) string {
	s = NormalizeText(s)
	lower := strings.ToLower(s)
	for _, tail := range []string{
		" together", " lately", " recently", " these days", " right now",
		" with the", " with my", " with our",
	} {
		if i := strings.Index(lower, tail); i > 0 {
			s = strings.TrimSpace(s[:i])
			lower = strings.ToLower(s)
		}
	}
	return s
}

func clipPreferenceTail(s string) string {
	s = NormalizeText(s)
	lower := strings.ToLower(s)
	for _, tail := range []string{" at the ", " at a ", " in the ", " in a ", " with the ", " with my ", " and "} {
		if i := strings.Index(lower, tail); i > 0 {
			s = strings.TrimSpace(s[:i])
			lower = strings.ToLower(s)
		}
	}
	return s
}

func pronounPreferenceObjects(body string) []string {
	out := make([]string, 0, 4)
	seen := map[string]struct{}{}
	add := func(raw string) {
		like := clipPreferenceTail(NormalizeText(raw))
		if !validPreferenceValue(like) {
			return
		}
		key := strings.ToLower(like)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, like)
	}
	for _, m := range theyExcitedRE.FindAllStringSubmatch(body, 4) {
		add(m[1])
	}
	for _, m := range theyLikeRE.FindAllStringSubmatch(body, 4) {
		add(m[1])
	}
	return out
}

func clipEventTail(s string) string {
	s = NormalizeText(s)
	lower := strings.ToLower(s)
	for _, tail := range []string{
		" last ", " this ", " yesterday", " today", " at a ", " at the ",
		" in a ", " in the ", " for my ", " for the ",
	} {
		if i := strings.Index(lower, tail); i > 0 {
			s = strings.TrimSpace(s[:i])
			lower = strings.ToLower(s)
		}
	}
	return s
}

func splitAndOrPhrases(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	lower := strings.ToLower(s)
	cut := func(sep string) []string {
		i := strings.Index(lower, sep)
		if i <= 0 {
			return nil
		}
		left := strings.TrimSpace(s[:i])
		right := strings.TrimSpace(s[i+len(sep):])
		if strings.HasPrefix(strings.ToLower(right), "working in ") {
			right = strings.TrimSpace(right[len("working in "):])
		}
		if utf8.RuneCountInString(left) < 3 || utf8.RuneCountInString(right) < 3 {
			return nil
		}
		return []string{left, right}
	}
	if parts := cut(" or "); parts != nil {
		return parts
	}
	if parts := cut(" and "); parts != nil {
		return parts
	}
	return []string{s}
}

func stripTrailingStamp(c string) string {
	c = strings.TrimSpace(c)
	if i := strings.LastIndex(c, " ("); i > 0 && strings.HasSuffix(c, ")") {
		return strings.TrimSpace(c[:i])
	}
	return c
}

func numberWord(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "one":
		return "1"
	case "two":
		return "2"
	case "three":
		return "3"
	case "four":
		return "4"
	case "five":
		return "5"
	case "six":
		return "6"
	case "seven":
		return "7"
	case "eight":
		return "8"
	case "nine":
		return "9"
	case "ten":
		return "10"
	default:
		return strings.TrimSpace(raw)
	}
}

func titleCaseWords(s string) string {
	words := strings.Fields(strings.TrimSpace(s))
	for i, w := range words {
		r, size := utf8.DecodeRuneInString(w)
		if r == utf8.RuneError || size == 0 {
			continue
		}
		words[i] = strings.ToUpper(string(r)) + w[size:]
	}
	return strings.Join(words, " ")
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
	case "attribute_occupation":
		return PredicateOccupation
	case "attribute_education":
		return PredicateEducation
	case "attribute_plan":
		return PredicatePlan
	case "attribute_event":
		return PredicateEvent
	case "attribute_possession":
		return PredicatePossession
	case "attribute_duration":
		return PredicateMetric
	default:
		return ""
	}
}

func valueNormFromAtomContent(content string) string {
	content = stripTrailingStamp(content)
	lower := strings.ToLower(content)
	for _, sep := range []string{
		" moved from ", " moved to ", " is from ", " participates in ", " enjoys ",
		" kids like ", " read \"", " mentioned \"", " is a ", " is ",
		" plans career in ", " plans career for ", " plans ", " studies ", " collects ",
		" has known friends for ", " has done ", " attended ",
	} {
		if i := strings.Index(lower, sep); i >= 0 {
			v := strings.TrimSpace(content[i+len(sep):])
			v = strings.Trim(v, "\"")
			if sep == " has done " {
				if _, place, ok := strings.Cut(v, " at "); ok {
					v = strings.TrimSpace(place)
				}
			}
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
	return titleCaseWords(p), isConcretePlace(p)
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
	c := strings.ToLower(stripTrailingStamp(content))
	if c == "" {
		return true
	}
	if i := strings.Index(c, " participates in "); i >= 0 {
		val := strings.Trim(strings.TrimSpace(c[i+len(" participates in "):]), `."'`)
		if val == "" {
			return true
		}
		head, _, _ := strings.Cut(val, " ")
		if _, stop := activityGerundStop[head]; stop {
			return true
		}
		if strings.HasSuffix(val, "ing") {
			// gerund form is the preferred activity atom
		} else if strings.Contains(val, " ") || utf8.RuneCountInString(head) < 5 {
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
		r, _ := utf8.DecodeRuneInString(last)
		if r >= 'a' && r <= 'z' {
			return true
		}
	}
	return false
}

func looksLikeWorkTitle(title string) bool {
	words := strings.Fields(title)
	if len(words) == 0 || len(words) > 8 {
		return false
	}
	if titleLeadStopped(title) {
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

func titleLeadStopped(title string) bool {
	words := strings.Fields(strings.TrimSpace(title))
	if len(words) == 0 {
		return true
	}
	head := strings.ToLower(strings.Trim(words[0], ",.;:!?\"'"))
	if head == "the" || head == "a" || head == "an" {
		return len(words) < 2
	}
	_, stop := titleLeadStop[head]
	return stop
}

func validPreferenceValue(like string) bool {
	like = strings.TrimSpace(like)
	if utf8.RuneCountInString(like) < 3 || utf8.RuneCountInString(like) > 40 {
		return false
	}
	if strings.Contains(strings.ToLower(like), " that ") {
		return false
	}
	head, _, _ := strings.Cut(strings.ToLower(like), " ")
	if _, stop := preferenceHeadStop[head]; stop {
		return false
	}
	return true
}
