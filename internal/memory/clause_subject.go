package memory

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// clauseBind attributes a verb match to the speaker, the two-party addressee,
// or a named person in the clause. First-person patterns stay speaker-bound.
type clauseBind struct {
	speaker   string
	partner   string
	known     map[string]string // lowercased display name → canonical
	lastNamed string            // last named person for she/he coref
}

func (b *clauseBind) rememberNamed(name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	b.lastNamed = name
}

func (b clauseBind) canonName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	if b.known != nil {
		if c, ok := b.known[strings.ToLower(name)]; ok {
			return c
		}
	}
	return titleCaseWords(name)
}

func newClauseBind(speaker string, speakers map[string]string) clauseBind {
	b := clauseBind{speaker: speaker, known: speakers}
	if len(speakers) != 2 {
		return b
	}
	self := strings.ToLower(speaker)
	for k, v := range speakers {
		if k != self {
			b.partner = v
			break
		}
	}
	return b
}

// withAddressee sets partner from two-party dialogue, else the most recent
// other speaker (so "you …" binds in longer threads).
func (b clauseBind) withAddressee(prior []string) clauseBind {
	if b.partner != "" {
		return b
	}
	self := strings.ToLower(b.speaker)
	for i := len(prior) - 1; i >= 0; i-- {
		who := strings.TrimSpace(prior[i])
		if who == "" || strings.ToLower(who) == self {
			continue
		}
		b.partner = who
		return b
	}
	return b
}

func collectSpeakers(turns []turn) map[string]string {
	out := make(map[string]string, 4)
	for _, t := range turns {
		who := strings.TrimSpace(t.who)
		if who == "" {
			continue
		}
		k := strings.ToLower(who)
		if _, ok := out[k]; !ok {
			out[k] = who
		}
	}
	return out
}

var (
	firstPersonSubj = map[string]struct{}{
		"i": {}, "i'm": {}, "i’m": {}, "i've": {}, "i’ve": {},
		"i'd": {}, "i’d": {}, "i'll": {}, "i’ll": {}, "im": {},
		"we": {}, "we're": {}, "we’re": {}, "we've": {}, "we’ve": {},
		"us": {}, "our": {},
	}
	secondPersonSubj = map[string]struct{}{
		"you": {}, "you're": {}, "you’re": {}, "you've": {}, "you’ve": {},
		"you'd": {}, "you’d": {}, "you'll": {}, "you’ll": {},
	}
	// Singular third-person pronouns bind to the last named person (R6/R7).
	singularCorefPronoun = map[string]struct{}{
		"he": {}, "she": {}, "him": {}, "her": {}, "his": {}, "hers": {},
	}
	// Plural / demonstrative pronouns stay unbound (kids, groups, objects).
	skipAnaphoraPronoun = map[string]struct{}{
		"they": {}, "them": {}, "their": {},
		"it": {}, "this": {}, "that": {}, "these": {}, "those": {},
	}

	// Walk left past auxiliaries, conjunctions, and the verb that was matched.
	subjectWalkSkip = map[string]struct{}{
		"been": {}, "have": {}, "has": {}, "had": {}, "is": {}, "was": {},
		"are": {}, "were": {}, "am": {}, "be": {}, "being": {},
		"currently": {}, "still": {}, "also": {}, "just": {}, "recently": {},
		"really": {}, "already": {}, "now": {}, "always": {}, "actually": {},
		"and": {}, "or": {}, "but": {}, "then": {}, "so": {},
		"a": {}, "an": {}, "the": {}, "my": {}, "your": {},
		"researched": {}, "researching": {}, "works": {}, "work": {}, "working": {},
		"lives": {}, "live": {}, "lived": {}, "living": {}, "moved": {}, "relocated": {},
		"enjoys": {}, "enjoy": {}, "enjoyed": {}, "loves": {}, "love": {}, "loved": {},
		"likes": {}, "like": {}, "liked": {}, "collects": {}, "collect": {}, "collected": {},
		"studied": {}, "studying": {}, "studies": {}, "realized": {}, "realised": {},
		"noticed": {}, "learned": {}, "learnt": {}, "went": {}, "goes": {}, "going": {},
		"gave": {}, "give": {}, "giving": {}, "ran": {}, "run": {}, "running": {},
		"read": {}, "reading": {}, "planning": {}, "plans": {}, "looking": {},
		"volunteers": {}, "volunteer": {}, "volunteering": {},
		"participates": {}, "participate": {},
	}
	prepSkip = map[string]struct{}{
		"in": {}, "at": {}, "from": {}, "to": {}, "on": {}, "of": {}, "for": {},
		"with": {}, "as": {}, "into": {}, "about": {}, "through": {}, "over": {},
		"after": {}, "before": {}, "by": {},
	}
	calendarNameStop = map[string]struct{}{
		"january": {}, "february": {}, "march": {}, "april": {}, "may": {}, "june": {},
		"july": {}, "august": {}, "september": {}, "october": {}, "november": {}, "december": {},
		"monday": {}, "tuesday": {}, "wednesday": {}, "thursday": {}, "friday": {},
		"saturday": {}, "sunday": {},
	}
)

// subjectAt returns the person the verb at matchStart belongs to.
// defaultSpeaker is true for clause-initial gerunds ("Researching X").
func (b *clauseBind) subjectAt(body string, matchStart int, defaultSpeaker bool) (string, bool) {
	if b == nil {
		return "", false
	}
	if matchStart < 0 {
		matchStart = 0
	}
	if matchStart > len(body) {
		matchStart = len(body)
	}
	left := strings.TrimSpace(body[:matchStart])
	left = speakerPrefixRe.ReplaceAllString(left, "")
	tokens := strings.Fields(left)
	i := len(tokens) - 1
	for i >= 0 {
		tok := strings.Trim(tokens[i], ",.;:!?\"'")
		if tok == "" {
			i--
			continue
		}
		low := strings.ToLower(tok)
		if _, ok := subjectWalkSkip[low]; ok {
			i--
			continue
		}
		if i > 0 {
			prev := strings.ToLower(strings.Trim(tokens[i-1], ",.;:!?\"'"))
			if _, ok := prepSkip[prev]; ok {
				i -= 2
				continue
			}
		}
		break
	}
	if i < 0 {
		if defaultSpeaker && strings.TrimSpace(b.speaker) != "" {
			return b.speaker, true
		}
		return "", false
	}
	tok := strings.Trim(tokens[i], ",.;:!?\"'")
	tok = strings.TrimSuffix(tok, "'s")
	tok = strings.TrimSuffix(tok, "’s")
	low := strings.ToLower(tok)
	if _, ok := firstPersonSubj[low]; ok {
		if strings.TrimSpace(b.speaker) == "" {
			return "", false
		}
		return b.speaker, true
	}
	if _, ok := secondPersonSubj[low]; ok {
		if b.partner == "" {
			return "", false
		}
		return b.partner, true
	}
	if _, ok := skipAnaphoraPronoun[low]; ok {
		return "", false
	}
	if _, ok := singularCorefPronoun[low]; ok {
		if strings.TrimSpace(b.lastNamed) == "" {
			return "", false
		}
		return b.lastNamed, true
	}
	if b.known != nil {
		if canon, ok := b.known[low]; ok {
			b.rememberNamed(canon)
			return canon, true
		}
	}
	if likelyPersonName(tok) {
		name := titleCaseWords(tok)
		b.rememberNamed(name)
		return name, true
	}
	if defaultSpeaker && strings.TrimSpace(b.speaker) != "" {
		return b.speaker, true
	}
	return "", false
}

func likelyPersonName(tok string) bool {
	tok = strings.TrimSpace(tok)
	if tok == "" {
		return false
	}
	low := strings.ToLower(tok)
	if _, ok := entityStopwords[low]; ok {
		return false
	}
	if _, ok := calendarNameStop[low]; ok {
		return false
	}
	if _, ok := titleLeadStop[low]; ok {
		return false
	}
	r, size := utf8.DecodeRuneInString(tok)
	if r == utf8.RuneError || !unicode.IsUpper(r) {
		return false
	}
	rest := tok[size:]
	if rest == "" {
		return false
	}
	for _, c := range rest {
		if !unicode.IsLetter(c) && c != '-' {
			return false
		}
	}
	n := utf8.RuneCountInString(tok)
	return n >= 2 && n <= 20
}

type reHit struct {
	groups []string
	start  int
}

func reFind(re *regexp.Regexp, s string) (reHit, bool) {
	loc := re.FindStringSubmatchIndex(s)
	if loc == nil {
		return reHit{}, false
	}
	return hitFromLoc(s, loc), true
}

func reFindAll(re *regexp.Regexp, s string, n int) []reHit {
	locs := re.FindAllStringSubmatchIndex(s, n)
	if len(locs) == 0 {
		return nil
	}
	out := make([]reHit, 0, len(locs))
	for _, loc := range locs {
		out = append(out, hitFromLoc(s, loc))
	}
	return out
}

func hitFromLoc(s string, loc []int) reHit {
	groups := make([]string, len(loc)/2)
	for i := 0; i < len(groups); i++ {
		if loc[2*i] >= 0 {
			groups[i] = s[loc[2*i]:loc[2*i+1]]
		}
	}
	return reHit{groups: groups, start: loc[0]}
}
