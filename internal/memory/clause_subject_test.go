package memory

import "testing"

func TestClauseBindNamedSubjectNotReporter(t *testing.T) {
	speakers := map[string]string{"riley": "Riley", "casey": "Casey"}
	bind := newClauseBind("Riley", speakers)
	body := "Casey researched wildfire recovery last spring."
	subj, ok := bind.subjectAt(body, 6, true) // "researched" starts at index 6
	if !ok || subj != "Casey" {
		t.Fatalf("named subject, got %q ok=%v", subj, ok)
	}
}

func TestClauseBindYouIsAddresseeInTwoParty(t *testing.T) {
	speakers := map[string]string{"riley": "Riley", "casey": "Casey"}
	bind := newClauseBind("Riley", speakers)
	if bind.partner != "Casey" {
		t.Fatalf("partner=%q", bind.partner)
	}
	body := "You realized that rest is part of training."
	subj, ok := bind.subjectAt(body, 4, true)
	if !ok || subj != "Casey" {
		t.Fatalf("addressee, got %q ok=%v", subj, ok)
	}
}

func TestClauseBindYouFallsBackToPriorSpeaker(t *testing.T) {
	speakers := map[string]string{"riley": "Riley", "casey": "Casey", "morgan": "Morgan"}
	bind := newClauseBind("Riley", speakers).withAddressee([]string{"Casey", "Riley"})
	if bind.partner != "Casey" {
		t.Fatalf("prior addressee, partner=%q", bind.partner)
	}
}

func TestClauseBindSheIsLastNamedPerson(t *testing.T) {
	speakers := map[string]string{"riley": "Riley", "casey": "Casey"}
	bind := newClauseBind("Riley", speakers)
	bind.lastNamed = "Casey"
	body := "she researched wildfire recovery"
	subj, ok := bind.subjectAt(body, 4, true)
	if !ok || subj != "Casey" {
		t.Fatalf("she → last named, got %q ok=%v", subj, ok)
	}
}

func TestClauseBindSheWithoutAntecedentSkips(t *testing.T) {
	speakers := map[string]string{"riley": "Riley"}
	bind := newClauseBind("Riley", speakers)
	body := "she researched wildfire recovery"
	subj, ok := bind.subjectAt(body, 4, true)
	if ok || subj != "" {
		t.Fatalf("must not steal she-clause without antecedent, got %q ok=%v", subj, ok)
	}
}

func TestClauseBindTheyDoesNotStealLastNamed(t *testing.T) {
	speakers := map[string]string{"riley": "Riley", "casey": "Casey"}
	bind := newClauseBind("Riley", speakers)
	bind.lastNamed = "Casey"
	body := "they researched wildfire recovery"
	subj, ok := bind.subjectAt(body, 5, true)
	if ok || subj != "" {
		t.Fatalf("they must skip, got %q ok=%v", subj, ok)
	}
}

func TestClauseBindGerundDefaultsToSpeaker(t *testing.T) {
	speakers := map[string]string{"jordan": "Jordan"}
	bind := newClauseBind("Jordan", speakers)
	body := "Researching scholarship programs"
	subj, ok := bind.subjectAt(body, 0, true)
	if !ok || subj != "Jordan" {
		t.Fatalf("gerund default, got %q ok=%v", subj, ok)
	}
}

func TestClauseBindSkipsPrepositionalComplement(t *testing.T) {
	speakers := map[string]string{"morgan": "Morgan"}
	bind := newClauseBind("Morgan", speakers)
	body := "Dana lives in Portland and is a carpenter"
	// "is a carpenter" — walk past "and", PP "in Portland", verb "lives"
	idx := indexAt(body, "is a carpenter")
	subj, ok := bind.subjectAt(body, idx, false)
	if !ok || subj != "Dana" {
		t.Fatalf("coord subject, got %q ok=%v", subj, ok)
	}
}

func TestLikelyPersonNameRejectsCalendar(t *testing.T) {
	if likelyPersonName("June") || likelyPersonName("Monday") || likelyPersonName("The") {
		t.Fatal("calendar/stop names are not people")
	}
	if !likelyPersonName("Dana") {
		t.Fatal("Dana should look like a person")
	}
}

func indexAt(s, needle string) int {
	i := 0
	for i+len(needle) <= len(s) {
		if s[i:i+len(needle)] == needle {
			return i
		}
		i++
	}
	return 0
}
