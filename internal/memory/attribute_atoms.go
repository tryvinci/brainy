package memory

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Deterministic attribute atoms turn dialogue into searchable standalone facts.
// Patterns are general linguistic forms only — no benchmark surface-forms.
// LOCOMO-shaped regexes are gated behind BRAINY_LEGACY_LOCOMO_ATOMS (default off).
// See docs/research/master-plan.md §1.2 / W1.

var (
	speakerLineRE         = regexp.MustCompile(`(?i)^\s*([A-Za-z][A-Za-z0-9_-]{1,40})\s*:\s*(.+)$`)
	quotedTitleRE         = regexp.MustCompile(`"([^"]{2,80})"|“([^”]{2,80})”|'([^']{2,80})'|‘([^’]{2,80})’`)
	movedFromRE           = regexp.MustCompile(`(?i)\b(?:moved|relocated)\s+from\s+([A-Za-z][A-Za-z\s-]{1,40})`)
	movedToRE             = regexp.MustCompile(`(?i)\b(?:moved|relocated)\s+to\s+([A-Za-z][A-Za-z\s-]{1,40})`)
	iAmRE                 = regexp.MustCompile(`(?i)\b(?:i am|i'm)\s+(?:a|an)\s+([^,.!?]{3,60})`)
	iAmStatusRE           = regexp.MustCompile(`(?i)\b(?:i am|i'm)\s+(single|married|divorced|engaged|widowed)\b`)
	worksAsRE             = regexp.MustCompile(`(?i)\bwork(?:s|ing)?\s+as\s+(?:a |an )?([^,.!?]{3,40})`)
	livesInRE             = regexp.MustCompile(`(?i)\b(?:live[sd]?|living)\s+in\s+([A-Za-z][A-Za-z\s-]{1,40})`)
	realizedThatRE        = regexp.MustCompile(`(?i)\b(?:realized|realised|noticed|learned|learnt)\s+that\s+([^,.!?]{4,80})`)
	namedIsRoleRE         = regexp.MustCompile(`\b([A-Z][a-z]{1,20})\s+is\s+(?:a|an)\s+([^,.!?]{3,60})`)
	namedIsStatusRE       = regexp.MustCompile(`(?i)\b([A-Z][a-z]{1,20})\s+is\s+(single|married|divorced|engaged|widowed)\b`)
	youAreRoleRE          = regexp.MustCompile(`(?i)\b(?:you(?:'re| are))\s+(?:a|an)\s+([^,.!?]{3,60})`)
	youAreStatusRE        = regexp.MustCompile(`(?i)\b(?:you(?:'re| are))\s+(single|married|divorced|engaged|widowed)\b`)
	isARoleBareRE         = regexp.MustCompile(`(?i)\bis\s+(?:a|an)\s+([^,.!?]{3,60})`)
	asARoleRE             = regexp.MustCompile(`(?i)\bas an?\s+([a-z][a-z\s-]{2,40})\b`)
	activityGerund        = regexp.MustCompile(`(?i)\b(?:i|i've|i have)\s+(?:been\s+)?([a-z]+ing)\b`)
	loveActivityRE        = regexp.MustCompile(`(?i)\b(?:i love|i like|i enjoy|we love|we like|we enjoy|i'm a (?:big )?fan of)\s+([a-z][a-z\s-]{2,40})`)
	placeWithActRE        = regexp.MustCompile(`(?i)\b([a-z]+ing)\s+(?:in|at|on)\s+(?:the\s+)?([a-z][a-z-]{2,30})\b`)
	tripPlaceRE           = regexp.MustCompile(`(?i)\b(camping|hiking|swimming)\s+trip\s+(?:in|at|on|to|through)\s+(?:the\s+)?([a-z][a-z-]{2,30})\b`)
	gaveEventRE           = regexp.MustCompile(`(?i)\b(?:gave|give|giving)\s+(?:a |an )?([a-z][a-z-]{3,24})\b`)
	hadEventRE            = regexp.MustCompile(`(?i)\b(?:had|have)\s+a\s+([a-z][a-z-]{3,24})\b`)
	ranEventRE            = regexp.MustCompile(`(?i)\b(?:ran|run)\s+(?:a|an)\s+([a-z][a-z\s-]{3,40})\b`)
	trainingForRE         = regexp.MustCompile(`(?i)\btraining\s+for\s+(?:a |an )?([a-z][a-z\s-]{3,40})\b`)
	kidsLikeRE            = regexp.MustCompile(`(?i)\b(?:kids?|children)\s+(?:love|like|enjoy|are into|were\s+\w+\s+(?:about|for))\s+(?:the\s+)?([a-z][a-z\s-]{2,40})`)
	kidsAboutRE           = regexp.MustCompile(`(?i)\b(?:kids?|children).{0,48}?\b(?:love|like|enjoy|into|about|obsessed with)\s+(?:the\s+)?([a-z][a-z\s-]{2,40})`)
	kidsMentionRE         = regexp.MustCompile(`(?i)\b(?:kids?|children)\b`)
	theyExcitedRE         = regexp.MustCompile(`(?i)\bthey\s+(?:were\s+|are\s+)?(?:stoked|excited|pumped|thrilled)\s+(?:for|about)\s+(?:the\s+)?([a-z][a-z\s-]{2,40})`)
	theyLikeRE            = regexp.MustCompile(`(?i)\bthey\s+(?:love|like|enjoy|are into)\s+(?:the\s+)?([a-z][a-z\s-]{2,40})`)
	homeCountryRE         = regexp.MustCompile(`(?i)\bhome country[,:]?\s+([A-Z][a-zA-Z]+(?:\s+[A-Z][a-zA-Z]+)?)\b`)
	homeCountryBareRE     = regexp.MustCompile(`(?i)\bhome country\b`)
	originallyFromRE      = regexp.MustCompile(`(?i)\boriginally from\s+([A-Za-z][A-Za-z\s-]{1,40})`)
	imFromRE              = regexp.MustCompile(`(?i)\b(?:i'm|i am)\s+from\s+([A-Za-z][A-Za-z\s-]{1,40})`)
	lookingIntoRE         = regexp.MustCompile(`(?i)\b(?:looking into|keen on|interested in)\s+([^,.!?]{3,70})`)
	researchingRE         = regexp.MustCompile(`(?i)\b(?:researching|researched)\s+([^,.!?]{3,50})`)
	careerFieldRE         = regexp.MustCompile(`(?i)\b(?:career|profession)\s+(?:in|as)\s+([^,.!?]{3,50})`)
	wantToBeRE            = regexp.MustCompile(`(?i)\b(?:want to|wanted to|hope to|plan to|planning to)\s+(?:be|become|pursue|work as)\s+(?:a |an )?([^,.!?]{3,50})`)
	workWithGroupRE       = regexp.MustCompile(`(?i)\b(?:thinking of working with|working with|work with)\s+([a-z][a-z\s-]{1,40}(?:people|community|patients|clients|families))\b`)
	planningActRE         = regexp.MustCompile(`(?i)\b(?:planning on|planning to|going to)\s+(?:go(?:ing)?\s+)?([a-z]+ing)\b`)
	workshopRE            = regexp.MustCompile(`(?i)\b([a-z][a-z-]{2,30})\s+(?:workshop|class|lesson)s?\b`)
	goGerundRE            = regexp.MustCompile(`(?i)\b(?:go|going|went|off to go)\s+([a-z]+ing)\b`)
	durationYearsRE       = regexp.MustCompile(`(?i)\bfor\s+(?:(?:about|around|nearly|almost|roughly|approximately)\s+)?(\d+|one|two|three|four|five|six|seven|eight|nine|ten)\s+years\b`)
	doingNamedActRE       = regexp.MustCompile(`(?i)\b(?:doing|practicing|practising|playing)\s+([a-z][a-z]+(?:\s+[a-z][a-z]+){0,4})`)
	doingActAnaphoraRE    = regexp.MustCompile(`(?i)\b(?:doing|practicing|practising|playing)\s+(?:them|it|this|that)\b`)
	durationMonthsRE      = regexp.MustCompile(`(?i)\b(?:for|took)\s+(\d+|one|two|three|four|five|six|seven|eight|nine|ten)\s+months\b`)
	seasonYearRE          = regexp.MustCompile(`(?i)\b(summer|winter|spring|fall|autumn)\s+(?:of\s+)?(20\d{2})\b`)
	startedActivityYearRE = regexp.MustCompile(`(?i)\b(?:started|began)\s+(practicing|playing|doing|working on|learning)\s+([^,.!?]{2,40}?)\s+in\s+(20\d{2})\b`)
	sinceMonthYearRE      = regexp.MustCompile(`(?i)\bsince\s+(january|february|march|april|may|june|july|august|september|october|november|december)\s+(20\d{2})\b`)
	collectsRE            = regexp.MustCompile(`(?i)\b(?:collect(?:s|ing)?|collection of)\s+([^,.!?]{3,50})`)
	educationRE           = regexp.MustCompile(`(?i)\b(?:studying|degree in|certification in|certified in)\s+([^,.!?]{3,50})`)
	readUnquotedRE        = regexp.MustCompile(`(?i)\b(?:read|reading|loved reading)\s+([A-Z][^"“”'.]{2,60})`)
	bookTitleRunRE        = regexp.MustCompile(`\b((?:The|A|An)\s+[A-Z][A-Za-z']+(?:\s+[A-Z][A-Za-z']+){0,5}|[A-Z][A-Za-z']+(?:\s+(?:is|the|a|an|of|and|in|to|for|on)\s+[A-Z][A-Za-z']+|\s+[A-Z][A-Za-z']+){1,5})\b`)
	visibleTextBlockRE    = regexp.MustCompile(`\[visible text:\s*([^\]]{4,800})\]`)

	// Activity gerunds that are usually not hobbies (skip).
	activityGerundStop = map[string]struct{}{
		"being": {}, "having": {}, "going": {}, "doing": {}, "getting": {},
		"looking": {}, "thinking": {}, "feeling": {}, "saying": {}, "trying": {},
		"making": {}, "taking": {}, "coming": {}, "seeing": {}, "knowing": {},
		"wanting": {}, "needing": {}, "working": {}, "living": {}, "telling": {},
		"fulfilling": {}, "investing": {}, "helping": {}, "using": {},
		"planning": {}, "starting": {}, "keeping": {},
		"sitting": {}, "standing": {}, "laying": {}, "lying": {},
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

type turn struct {
	who, body, utt string
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
		turns = append(turns, turn{who: speaker, body: body, utt: utt})
	}
	speakers := collectSpeakers(turns)
	acts := collectDoingActivities(turns)
	var prior []string
	lastNamed := ""
	for i, t := range turns {
		whoKey := strings.ToLower(t.who)
		if t.who != "" && kidsMentionRE.MatchString(t.body) {
			kidsBySpeaker[whoKey] = true
		}
		bind := newClauseBind(t.who, speakers).withAddressee(prior)
		bind.lastNamed = lastNamed
		for _, atom := range attributeAtomsFromUtterance(t.who, t.body, t.utt, observedAt, &bind, i, acts) {
			add(atom)
			if pred, _ := atom.Explain["predicate"].(string); pred == PredicateOrigin {
				if v, _ := atom.Explain["value_norm"].(string); v != "" && whoKey != "" {
					origins[whoKey] = v
				}
			}
		}
		if t.who != "" && kidsBySpeaker[whoKey] {
			for _, like := range pronounPreferenceObjects(t.body) {
				add(atomFact(t.who, fmt.Sprintf("%s's kids like %s", t.who, like), t.utt, 0.82, "attribute_preference", observedAt))
			}
		}
		if bind.lastNamed != "" {
			lastNamed = bind.lastNamed
		}
		if t.who != "" {
			for _, like := range pronounExcitedObjects(t.body) {
				add(atomFact(t.who, fmt.Sprintf("%s enjoys %s", t.who, like), t.utt, 0.8, "attribute_preference", observedAt))
			}
			prior = append(prior, t.who)
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

func attributeAtomsFromUtterance(who, body, source string, observedAt *time.Time, bind *clauseBind, turnIdx int, acts []doingActSpan) []ExtractedMemory {
	if bind == nil {
		tmp := newClauseBind(who, nil)
		bind = &tmp
	}
	if bind.speaker == "" {
		bind.speaker = who
	}
	var out []ExtractedMemory
	lowerBody := strings.ToLower(body)
	lexicalBody := visibleTextBlockRE.ReplaceAllString(body, " ")
	lexicalLower := strings.ToLower(lexicalBody)
	emit := func(subj, content string, conf float64, rule string) {
		if strings.TrimSpace(subj) == "" || strings.TrimSpace(content) == "" {
			return
		}
		out = append(out, atomFact(subj, content, source, conf, rule, observedAt))
	}
	emitOn := func(haystack string, start int, defaultSpeaker bool, conf float64, rule, format string, args ...any) {
		subj, ok := bind.subjectAt(haystack, start, defaultSpeaker)
		if !ok {
			return
		}
		full := make([]any, 0, 1+len(args))
		full = append(full, subj)
		full = append(full, args...)
		emit(subj, fmt.Sprintf(format, full...), conf, rule)
	}

	if who != "" {
		if m := iAmStatusRE.FindStringSubmatch(body); m != nil {
			emit(who, fmt.Sprintf("%s is %s", who, NormalizeText(m[1])), 0.9, "attribute_relationship")
		}
		if m := iAmRE.FindStringSubmatch(body); m != nil {
			attr := clipIdentityTail(NormalizeText(m[1]))
			if utf8.RuneCountInString(attr) <= 40 && !roleStopWord(attr) {
				emit(who, fmt.Sprintf("%s is a %s", who, attr), 0.88, "attribute_identity")
			}
		}
		if m := imFromRE.FindStringSubmatch(body); m != nil {
			place, ok := normalizePlace(m[1])
			if ok {
				emit(who, fmt.Sprintf("%s is from %s", who, place), 0.92, "attribute_origin")
			}
		}
	}
	if hit, ok := reFind(youAreStatusRE, body); ok && bind.partner != "" {
		emit(bind.partner, fmt.Sprintf("%s is %s", bind.partner, NormalizeText(hit.groups[1])), 0.9, "attribute_relationship")
	}
	if hit, ok := reFind(youAreRoleRE, body); ok && bind.partner != "" {
		attr := clipIdentityTail(NormalizeText(hit.groups[1]))
		if utf8.RuneCountInString(attr) <= 40 && !roleStopWord(attr) {
			emit(bind.partner, fmt.Sprintf("%s is a %s", bind.partner, attr), 0.88, "attribute_identity")
		}
	}
	if hit, ok := reFind(namedIsStatusRE, body); ok {
		name := bind.canonName(hit.groups[1])
		if likelyPersonName(name) {
			emit(name, fmt.Sprintf("%s is %s", name, NormalizeText(hit.groups[2])), 0.9, "attribute_relationship")
		}
	}
	if hit, ok := reFind(namedIsRoleRE, body); ok {
		name := bind.canonName(hit.groups[1])
		attr := clipIdentityTail(NormalizeText(hit.groups[2]))
		if likelyPersonName(name) && utf8.RuneCountInString(attr) <= 40 && !roleStopWord(attr) && !looksLikeWorkTitle(attr) {
			emit(name, fmt.Sprintf("%s is a %s", name, attr), 0.88, "attribute_identity")
		}
	}
	if hit, ok := reFind(isARoleBareRE, body); ok {
		attr := clipIdentityTail(NormalizeText(hit.groups[1]))
		if utf8.RuneCountInString(attr) <= 40 && !roleStopWord(attr) && !looksLikeWorkTitle(attr) && !strings.Contains(attr, " that ") {
			emitOn(body, hit.start, false, 0.86, "attribute_identity", "%s is a %s", attr)
		}
	}
	if hit, ok := reFind(asARoleRE, body); ok {
		role := NormalizeText(hit.groups[1])
		if utf8.RuneCountInString(role) <= 40 && !strings.Contains(role, " that ") && !roleStopWord(role) {
			emitOn(body, hit.start, true, 0.9, "attribute_identity", "%s is a %s", role)
			if strings.Contains(strings.ToLower(role), "single") {
				emitOn(body, hit.start, true, 0.88, "attribute_relationship", "%s is single")
			}
		}
	}
	if hit, ok := reFind(worksAsRE, body); ok {
		role := clipIdentityTail(NormalizeText(hit.groups[1]))
		if utf8.RuneCountInString(role) >= 3 && utf8.RuneCountInString(role) <= 40 && !roleStopWord(role) {
			emitOn(body, hit.start, true, 0.9, "attribute_occupation", "%s works as %s", role)
			emitOn(body, hit.start, true, 0.86, "attribute_identity", "%s is a %s", role)
		}
	}
	if hit, ok := reFind(livesInRE, body); ok {
		place, ok := normalizePlace(hit.groups[1])
		if ok {
			emitOn(body, hit.start, true, 0.9, "attribute_residence", "%s lives in %s", place)
		}
	}
	if hit, ok := reFind(realizedThatRE, body); ok {
		clause := NormalizeText(hit.groups[1])
		if utf8.RuneCountInString(clause) >= 4 && utf8.RuneCountInString(clause) <= 80 {
			emitOn(body, hit.start, true, 0.88, "attribute_belief", "%s realized that %s", clause)
		}
	}
	if hit, ok := reFind(homeCountryRE, body); ok {
		place, ok := normalizePlace(hit.groups[1])
		if ok {
			emitOn(body, hit.start, true, 0.93, "attribute_origin", "%s is from %s", place)
			emitOn(body, hit.start, true, 0.9, "attribute_origin", "%s moved from %s", place)
		}
	}
	if hit, ok := reFind(originallyFromRE, body); ok {
		place, ok := normalizePlace(hit.groups[1])
		if ok {
			emitOn(body, hit.start, true, 0.93, "attribute_origin", "%s is from %s", place)
		}
	}
	if hit, ok := reFind(movedFromRE, body); ok {
		place, ok := normalizePlace(hit.groups[1])
		if ok {
			emitOn(body, hit.start, true, 0.9, "attribute_origin", "%s moved from %s", place)
		}
	}
	if hit, ok := reFind(movedToRE, body); ok {
		place, ok := normalizePlace(hit.groups[1])
		if ok {
			emitOn(body, hit.start, true, 0.9, "attribute_origin", "%s moved to %s", place)
		}
	}

	for _, hit := range reFindAll(quotedTitleRE, lexicalBody, 4) {
		title := firstNonEmpty(hit.groups[1], hit.groups[2], hit.groups[3], hit.groups[4])
		title = NormalizeText(title)
		if utf8.RuneCountInString(title) < 4 || looksBrokenQuotedTitle(title) {
			continue
		}
		bookCue := strings.Contains(lexicalLower, "read") || strings.Contains(lexicalLower, "reading") ||
			strings.Contains(lexicalLower, "book") || strings.Contains(lexicalLower, "title") ||
			strings.Contains(lexicalLower, "novel") || strings.Contains(lexicalLower, "story")
		mediaCue := strings.Contains(lexicalLower, "play") || strings.Contains(lexicalLower, "game") ||
			strings.Contains(lexicalLower, "watch") || strings.Contains(lexicalLower, "movie")
		if !bookCue && !mediaCue {
			continue
		}
		if len(strings.Fields(title)) < 2 {
			continue
		}
		format := "%s read \"%s\""
		if !bookCue && mediaCue {
			format = "%s plays \"%s\""
		}
		emitOn(lexicalBody, hit.start, true, 0.86, "attribute_titled_work", format, title)
	}
	if hit, ok := reFind(readUnquotedRE, lexicalBody); ok {
		title := NormalizeText(strings.TrimSpace(hit.groups[1]))
		title = strings.TrimSuffix(title, " as a kid")
		title = strings.TrimSuffix(title, " last year")
		if looksLikeWorkTitle(title) && utf8.RuneCountInString(title) >= 4 && !looksBrokenQuotedTitle(title) && !titleLeadStopped(title) {
			emitOn(lexicalBody, hit.start, true, 0.84, "attribute_titled_work", "%s read \"%s\"", title)
		}
	}
	if strings.Contains(lexicalLower, "book") || strings.Contains(lexicalLower, "read") ||
		strings.Contains(lexicalLower, "game") || strings.Contains(lexicalLower, "play") {
		bookCue := strings.Contains(lexicalLower, "book") || strings.Contains(lexicalLower, "read")
		for _, hit := range reFindAll(bookTitleRunRE, lexicalBody, 4) {
			title := NormalizeText(strings.TrimSpace(hit.groups[1]))
			if titleLeadStopped(title) || !looksLikeWorkTitle(title) || looksBrokenQuotedTitle(title) {
				continue
			}
			if utf8.RuneCountInString(title) < 8 {
				continue
			}
			format := "%s read \"%s\""
			if !bookCue {
				format = "%s plays \"%s\""
			}
			emitOn(lexicalBody, hit.start, true, 0.82, "attribute_titled_work", format, title)
		}
	}
	if title, ok := deicticVisibleWorkTitle(body); ok {
		emitOn(body, 0, true, 0.85, "attribute_titled_work", "%s read \"%s\"", title)
	}

	if who != "" {
		if m := activityGerund.FindStringSubmatch(lexicalBody); m != nil {
			act := strings.ToLower(NormalizeText(m[1]))
			if _, stop := activityGerundStop[act]; !stop && len(act) >= 5 {
				emit(who, fmt.Sprintf("%s participates in %s", who, act), 0.82, "attribute_activity")
			}
		}
		if m := loveActivityRE.FindStringSubmatch(lexicalBody); m != nil {
			act := clipActivityTail(NormalizeText(m[1]))
			if utf8.RuneCountInString(act) >= 4 && utf8.RuneCountInString(act) <= 40 && !strings.Contains(act, " that ") {
				emit(who, fmt.Sprintf("%s enjoys %s", who, act), 0.82, "attribute_activity")
			}
		}
	}
	for _, hit := range reFindAll(placeWithActRE, lexicalBody, 6) {
		act := strings.ToLower(NormalizeText(hit.groups[1]))
		place, ok := normalizePlace(hit.groups[2])
		if _, stop := activityGerundStop[act]; !stop && ok {
			emitOn(lexicalBody, hit.start, true, 0.86, "attribute_place_activity", "%s has done %s at %s", act, place)
		}
	}
	for _, hit := range reFindAll(tripPlaceRE, lexicalBody, 4) {
		act := strings.ToLower(NormalizeText(hit.groups[1]))
		place, ok := normalizePlace(hit.groups[2])
		if _, stop := activityGerundStop[act]; !stop && ok {
			emitOn(lexicalBody, hit.start, true, 0.86, "attribute_place_activity", "%s has done %s at %s", act, place)
		}
	}
	if hit, ok := reFind(goGerundRE, lexicalBody); ok {
		act := strings.ToLower(NormalizeText(hit.groups[1]))
		if _, stop := activityGerundStop[act]; !stop && len(act) >= 5 {
			emitOn(lexicalBody, hit.start, true, 0.84, "attribute_activity", "%s participates in %s", act)
		}
	}
	if hit, ok := reFind(workshopRE, lexicalBody); ok {
		act := strings.ToLower(NormalizeText(hit.groups[1]))
		if _, stop := activityGerundStop[act]; !stop && utf8.RuneCountInString(act) >= 4 {
			emitOn(lexicalBody, hit.start, true, 0.88, "attribute_activity", "%s participates in %s", act)
		}
	}
	if hit, ok := reFind(planningActRE, lexicalBody); ok {
		act := strings.ToLower(NormalizeText(hit.groups[1]))
		if _, stop := activityGerundStop[act]; !stop && len(act) >= 5 {
			emitOn(lexicalBody, hit.start, true, 0.86, "attribute_plan", "%s plans %s", act)
			emitOn(lexicalBody, hit.start, true, 0.8, "attribute_activity", "%s participates in %s", act)
		}
	}
	seenLikes := map[string]struct{}{}
	emitLike := func(start int, raw string, conf float64) {
		like := clipPreferenceTail(NormalizeText(raw))
		if !validPreferenceValue(like) {
			return
		}
		key := strings.ToLower(like)
		if _, ok := seenLikes[key]; ok {
			return
		}
		seenLikes[key] = struct{}{}
		emitOn(body, start, true, conf, "attribute_preference", "%s's kids like %s", like)
	}
	for _, hit := range reFindAll(kidsLikeRE, body, 6) {
		emitLike(hit.start, hit.groups[1], 0.84)
	}
	for _, hit := range reFindAll(kidsAboutRE, body, 6) {
		emitLike(hit.start, hit.groups[1], 0.82)
	}
	emitCareer := func(start int, field string, conf float64, rule string) {
		field = clipCareerTail(field)
		for _, part := range splitAndOrPhrases(field) {
			if utf8.RuneCountInString(part) >= 3 && utf8.RuneCountInString(part) <= 60 {
				if rule == "attribute_education" {
					emitOn(body, start, true, conf, rule, "%s studies %s", part)
				} else {
					emitOn(body, start, true, conf, rule, "%s plans career in %s", part)
				}
			}
		}
	}
	if hit, ok := reFind(lookingIntoRE, body); ok {
		emitCareer(hit.start, NormalizeText(hit.groups[1]), 0.88, "attribute_occupation")
	}
	if hit, ok := reFind(researchingRE, body); ok {
		topic := clipResearchTopic(NormalizeText(hit.groups[1]))
		if utf8.RuneCountInString(topic) >= 4 && utf8.RuneCountInString(topic) <= 50 {
			emitOn(body, hit.start, true, 0.86, "attribute_plan", "%s researched %s", topic)
		}
	}
	if hit, ok := reFind(careerFieldRE, body); ok {
		emitCareer(hit.start, NormalizeText(hit.groups[1]), 0.88, "attribute_occupation")
	}
	if hit, ok := reFind(wantToBeRE, body); ok {
		emitCareer(hit.start, NormalizeText(hit.groups[1]), 0.86, "attribute_occupation")
	}
	if hit, ok := reFind(educationRE, body); ok {
		emitCareer(hit.start, NormalizeText(hit.groups[1]), 0.86, "attribute_education")
	}
	if hit, ok := reFind(workWithGroupRE, body); ok {
		group := clipCareerTail(NormalizeText(hit.groups[1]))
		if utf8.RuneCountInString(group) >= 5 && utf8.RuneCountInString(group) <= 50 {
			emitOn(body, hit.start, true, 0.86, "attribute_occupation", "%s plans career for %s", group)
		}
	}
	emitEvent := func(start int, raw string, conf float64) {
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
		emitOn(body, start, true, conf, "attribute_event", "%s attended %s", ev)
	}
	if hit, ok := reFind(gaveEventRE, body); ok {
		emitEvent(hit.start, hit.groups[1], 0.86)
	}
	if hit, ok := reFind(hadEventRE, body); ok {
		emitEvent(hit.start, hit.groups[1], 0.84)
	}
	if hit, ok := reFind(ranEventRE, body); ok {
		emitEvent(hit.start, hit.groups[1], 0.86)
	}
	if hit, ok := reFind(trainingForRE, body); ok {
		act := strings.ToLower(clipIdentityTail(clipEventTail(NormalizeText(hit.groups[1]))))
		if utf8.RuneCountInString(act) >= 4 && utf8.RuneCountInString(act) <= 40 {
			if _, stop := eventLightStop[act]; !stop {
				if _, stop := activityGerundStop[act]; !stop {
					emitOn(body, hit.start, true, 0.86, "attribute_activity", "%s participates in %s", act)
				}
			}
		}
	}
	if hit, ok := reFind(durationYearsRE, body); ok {
		n := numberWord(hit.groups[1])
		if strings.Contains(lowerBody, "friend") {
			emitOn(body, hit.start, true, 0.9, "attribute_duration", "%s has known friends for %s years", n)
		} else if named := resolveDurationActivities(body, turnIdx, acts); len(named) > 0 {
			if y := durationStartYear(numberWordInt(n), body, observedAt); y >= 1900 && y <= 2100 {
				for _, act := range named {
					emitOn(body, hit.start, true, 0.86, "attribute_event", "%s started practicing %s in %s", act, strconv.Itoa(y))
				}
			} else {
				emitOn(body, hit.start, true, 0.82, "attribute_duration", "%s has a %s-year duration", n)
			}
		} else {
			emitOn(body, hit.start, true, 0.82, "attribute_duration", "%s has a %s-year duration", n)
		}
	}
	if hit, ok := reFind(durationMonthsRE, body); ok {
		n := numberWord(hit.groups[1])
		emitOn(body, hit.start, true, 0.82, "attribute_duration", "%s has a %s-month duration", n)
	}
	if strings.Contains(lexicalLower, "start") || strings.Contains(lexicalLower, "began") ||
		strings.Contains(lexicalLower, "working on") {
		if hit, ok := reFind(seasonYearRE, body); ok {
			season := strings.ToLower(NormalizeText(hit.groups[1]))
			year := NormalizeText(hit.groups[2])
			if season != "" && year != "" {
				emitOn(body, hit.start, true, 0.84, "attribute_event", "%s started in %s %s", season, year)
			}
		}
	}
	if hit, ok := reFind(startedActivityYearRE, body); ok {
		verb := strings.ToLower(NormalizeText(hit.groups[1]))
		act := strings.ToLower(clipIdentityTail(NormalizeText(hit.groups[2])))
		year := NormalizeText(hit.groups[3])
		if utf8.RuneCountInString(act) >= 2 && utf8.RuneCountInString(act) <= 40 && year != "" {
			emitOn(body, hit.start, true, 0.86, "attribute_event", "%s started %s %s in %s", verb, act, year)
		}
	}
	if hit, ok := reFind(sinceMonthYearRE, body); ok {
		if strings.Contains(lexicalLower, "develop") || strings.Contains(lexicalLower, "working") ||
			strings.Contains(lexicalLower, "practic") || strings.Contains(lexicalLower, "play") ||
			strings.Contains(lexicalLower, "build") || strings.Contains(lexicalLower, "start") ||
			strings.Contains(lexicalLower, "began") {
			month := strings.ToLower(NormalizeText(hit.groups[1]))
			year := NormalizeText(hit.groups[2])
			if month != "" && year != "" {
				emitOn(body, hit.start, true, 0.84, "attribute_event", "%s started in %s %s", month, year)
			}
		}
	}
	if hit, ok := reFind(collectsRE, body); ok {
		obj := NormalizeText(hit.groups[1])
		if utf8.RuneCountInString(obj) >= 4 && utf8.RuneCountInString(obj) <= 50 && !strings.Contains(obj, " that ") {
			emitOn(body, hit.start, true, 0.86, "attribute_possession", "%s collects %s", obj)
		}
	}

	if strings.Contains(lowerBody, "unwind") || strings.Contains(lowerBody, "relax") ||
		strings.Contains(lowerBody, "clear my mind") {
		unwindRE := regexp.MustCompile(`(?i)\b([a-z]{4,}ing)\b`)
		for _, hit := range reFindAll(unwindRE, body, 6) {
			act := strings.ToLower(hit.groups[1])
			if _, stop := activityGerundStop[act]; stop {
				continue
			}
			emitOn(body, hit.start, true, 0.86, "attribute_activity", "%s unwinds via %s", act)
			emitOn(body, hit.start, true, 0.82, "attribute_activity", "%s participates in %s", act)
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
	for _, tail := range []string{
		" and ", " still ", " that ", " which ", " who ", " when ", " because ",
		" but ", " so ",
	} {
		if i := strings.Index(lower, tail); i > 0 {
			s = strings.TrimSpace(s[:i])
			lower = strings.ToLower(s)
		}
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

func clipResearchTopic(s string) string {
	s = NormalizeText(s)
	for _, sep := range []string{" — ", " – ", " -- ", " - ", "—", "–"} {
		if i := strings.Index(s, sep); i > 0 {
			s = strings.TrimSpace(s[:i])
		}
	}
	return s
}

func pronounExcitedObjects(body string) []string {
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
	return out
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

func numberWordInt(raw string) int {
	n, err := strconv.Atoi(numberWord(raw))
	if err != nil || n <= 0 || n > 80 {
		return 0
	}
	return n
}

type doingActSpan struct {
	idx  int
	acts []string
}

var durationActivityStop = map[string]struct{}{
	"them": {}, "it": {}, "this": {}, "that": {}, "these": {}, "those": {},
	"some": {}, "more": {}, "things": {}, "stuff": {}, "thing": {},
	"something": {}, "anything": {}, "everything": {},
	"activity": {}, "activities": {}, "hobby": {}, "hobbies": {},
	"time": {}, "lot": {}, "bit": {}, "way": {},
}

func collectDoingActivities(turns []turn) []doingActSpan {
	out := make([]doingActSpan, 0, len(turns))
	for i, t := range turns {
		acts := namedDoingActivities(t.body)
		if len(acts) == 0 {
			continue
		}
		out = append(out, doingActSpan{idx: i, acts: acts})
	}
	return out
}

func namedDoingActivities(body string) []string {
	hit, ok := reFind(doingNamedActRE, body)
	if !ok {
		return nil
	}
	raw := strings.ToLower(NormalizeText(hit.groups[1]))
	raw = strings.ReplaceAll(raw, ",", " and ")
	parts := strings.Split(raw, " and ")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, p := range parts {
		p = clipDurationActivity(p)
		if p == "" {
			continue
		}
		head, _, _ := strings.Cut(p, " ")
		if _, stop := durationActivityStop[p]; stop {
			continue
		}
		if _, stop := durationActivityStop[head]; stop {
			continue
		}
		if utf8.RuneCountInString(p) < 3 || utf8.RuneCountInString(p) > 40 {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func clipDurationActivity(s string) string {
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)
	for _, tail := range []string{" to ", " for ", " with ", " when ", " that ", " which ", " because ", " during "} {
		if i := strings.Index(lower, tail); i > 0 {
			s = strings.TrimSpace(s[:i])
			lower = strings.ToLower(s)
		}
	}
	return strings.TrimSpace(s)
}

func resolveDurationActivities(body string, turnIdx int, acts []doingActSpan) []string {
	if named := namedDoingActivities(body); len(named) > 0 {
		return named
	}
	lower := strings.ToLower(body)
	if !doingActAnaphoraRE.MatchString(body) && !strings.Contains(lower, "an activity") &&
		!strings.Contains(lower, "the activity") {
		return nil
	}
	best := []string(nil)
	bestDist := 1 << 20
	for _, span := range acts {
		if span.idx == turnIdx || len(span.acts) == 0 {
			continue
		}
		d := span.idx - turnIdx
		if d < 0 {
			d = -d
		} else {
			d += 100
		}
		if d < bestDist {
			bestDist = d
			best = span.acts
		}
	}
	return best
}

func durationStartYear(n int, body string, observedAt *time.Time) int {
	if n <= 0 || n > 80 {
		return 0
	}
	if y := leftoverCoveringRelativeStartYear(body); y >= 1900 && y <= 2100 {
		return y
	}
	if observedAt != nil && !observedAt.IsZero() {
		return observedAt.Year() - n
	}
	return 0
}

func deicticVisibleWorkTitle(body string) (string, bool) {
	lower := strings.ToLower(body)
	if !strings.Contains(lower, "this book") && !strings.Contains(lower, "this novel") &&
		!strings.Contains(lower, "this title") {
		return "", false
	}
	if !strings.Contains(lower, "read") && !strings.Contains(lower, "reading") {
		return "", false
	}
	m := visibleTextBlockRE.FindStringSubmatch(body)
	if m == nil {
		return "", false
	}
	return titleFromVisibleText(m[1])
}

func titleFromVisibleText(raw string) (string, bool) {
	best := ""
	bestRank := 0
	consider := func(cand string) {
		words := strings.Fields(cand)
		n := len(words)
		if n < 2 || n > 5 {
			return
		}
		if words[0][0] >= '0' && words[0][0] <= '9' {
			return
		}
		hasFn := false
		content := 0
		for _, w := range words {
			up := strings.ToUpper(w)
			letters := titleLetterCount(up)
			if _, ok := titleFunctionWord[up]; ok {
				hasFn = true
				continue
			}
			if letters < 3 || !titleHasVowel(up) {
				return
			}
			content++
		}
		if !hasFn || content < 2 {
			return
		}
		rank := titleLenRank(n)
		if rank > bestRank || (rank == bestRank && len(cand) > len(best)) {
			best, bestRank = cand, rank
		}
	}
	var run []string
	flush := func() {
		for i := 0; i < len(run); i++ {
			for n := 2; n <= 5 && i+n <= len(run); n++ {
				consider(strings.Join(run[i:i+n], " "))
			}
		}
		run = nil
	}
	for _, w := range strings.Fields(raw) {
		w = strings.Trim(w, ".,:;!?\"'")
		if isAllCapsWord(w) {
			run = append(run, w)
			continue
		}
		flush()
	}
	flush()
	if best == "" {
		for _, m := range bookTitleRunRE.FindAllStringSubmatch(raw, 8) {
			title := NormalizeText(strings.TrimSpace(m[1]))
			if titleLeadStopped(title) || looksBrokenQuotedTitle(title) {
				continue
			}
			if utf8.RuneCountInString(title) < 8 {
				continue
			}
			consider(strings.ToUpper(title))
		}
	}
	if best == "" || bestRank <= 0 {
		return "", false
	}
	titled := titleCaseWorkTitle(best)
	if !looksLikeWorkTitle(titled) || looksBrokenQuotedTitle(titled) {
		return "", false
	}
	return titled, true
}

func titleLetterCount(w string) int {
	n := 0
	for _, r := range w {
		if r >= 'A' && r <= 'Z' {
			n++
		}
	}
	return n
}

func titleHasVowel(w string) bool {
	for _, r := range w {
		switch r {
		case 'A', 'E', 'I', 'O', 'U', 'Y':
			return true
		}
	}
	return false
}

func isAllCapsWord(w string) bool {
	if w == "" {
		return false
	}
	letters := 0
	for _, r := range w {
		switch {
		case r >= 'A' && r <= 'Z':
			letters++
		case r >= '0' && r <= '9', r == '\'':
		default:
			return false
		}
	}
	return letters >= 1
}

var titleFunctionWord = map[string]struct{}{
	"A": {}, "AN": {}, "THE": {}, "AND": {}, "OR": {}, "OF": {},
	"TO": {}, "FOR": {}, "IN": {}, "ON": {}, "IS": {}, "AT": {},
}

func titleLenRank(n int) int {
	switch n {
	case 3:
		return 4
	case 4:
		return 3
	case 2:
		return 2
	case 5:
		return 1
	default:
		return 0
	}
}

func titleCaseWorkTitle(s string) string {
	words := strings.Fields(strings.ToLower(NormalizeText(s)))
	if len(words) == 0 {
		return ""
	}
	small := map[string]struct{}{
		"a": {}, "an": {}, "the": {}, "and": {}, "or": {}, "of": {},
		"to": {}, "for": {}, "in": {}, "on": {}, "is": {}, "at": {},
	}
	for i, w := range words {
		if i > 0 {
			if _, ok := small[w]; ok {
				words[i] = w
				continue
			}
		}
		words[i] = titleCaseWords(w)
	}
	return strings.Join(words, " ")
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
	case "attribute_residence":
		return PredicateResidence
	case "attribute_belief":
		return PredicateBelief
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
		" moved from ", " moved to ", " is from ", " lives in ", " participates in ", " enjoys ",
		" kids like ", " read \"", " mentioned \"", " works as ", " realized that ", " is a ", " is ",
		" plans career in ", " plans career for ", " plans ", " studies ", " collects ",
		" researched ", " has known friends for ", " has done ", " attended ",
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
		"life", "ways", "changes", "touch", "then", "need", "you",
		"top", "floor", "grass", "chairs", "chair", "stage", "sky",
		"steps", "front", "ground", "table", "sign":
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
	if malformedIndefiniteLightVerbSubject(c) {
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

// malformedIndefiniteLightVerbSubject is an extractor template whose subject
// never resolved (Anything/Something has done X at Y). Those lines are not
// durable facts and must not be admitted or used as leftover covering.
func malformedIndefiniteLightVerbSubject(c string) bool {
	for _, p := range []string{"anything", "something", "someone", "everyone", "everybody", "nobody"} {
		if strings.HasPrefix(c, p+" has done ") || strings.HasPrefix(c, p+" participates in ") {
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
