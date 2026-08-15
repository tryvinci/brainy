package memory

import (
	"context"
	"strings"
	"testing"
	"time"
)

// Held-out representation audit (R1b exit): a conversation that is not a
// benchmark transcript must compile durable claims into well-formed atoms
// with speaker, predicate, value, and provenance.
func TestHeldOutCompilerCoverageAudit(t *testing.T) {
	at := mustParseTime(t, "2023-07-15T12:00:00Z")
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u-audit", SourceType: "conversation",
		Metadata: map[string]any{"observed_at": "2023-07-15T12:00:00Z"},
		Messages: []Message{
			{Role: "user", Content: "Jordan: I am a community organizer and I'm single."},
			{Role: "user", Content: "Jordan: This necklace is from my home country, Portugal."},
			{Role: "user", Content: "Jordan: I've known these friends for four years, since I moved from my home country."},
			{Role: "user", Content: "Jordan: I'm looking into counseling and mental health as a career."},
			{Role: "user", Content: "Jordan: I collect classic children's books."},
			{Role: "user", Content: `Jordan: I loved reading "The Little Prince" as a kid.`},
			{Role: "user", Content: "Riley: We love painting together lately."},
			{Role: "user", Content: "Riley: I've been camping at the beach and later camping at mountains."},
			{Role: "user", Content: "Riley: I'm off to go swimming with the kids."},
			{Role: "user", Content: "Riley: Last Fri I finally took my kids to a pottery workshop."},
			{Role: "user", Content: "Riley: Been running longer — a great way to destress and clear my mind."},
			{Role: "user", Content: "Riley: The kids were wild about dinosaurs at the museum."},
			{Role: "user", Content: "Dana: Yesterday I took the kids to the museum."},
			{Role: "user", Content: "Dana: They were stoked for the fossils exhibit!"},
			{Role: "user", Content: "Riley: I'm planning on going camping in June."},
			{Role: "user", Content: "Riley: We went on another camping trip in the forest."},
			{Role: "user", Content: "Jordan: Last week I gave a speech at a school."},
			{Role: "user", Content: "Jordan: I'm thinking of working with elderly patients."},
			{Role: "user", Content: "Riley: This book I read last year, The Hidden Garden, still stays with me."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	atoms := 0
	preds := map[string]int{}
	withProvenance := 0
	for _, m := range memories {
		joined += " | " + strings.ToLower(m.Content)
		rule, _ := m.Explain["rule"].(string)
		if !strings.HasPrefix(rule, "attribute_") {
			continue
		}
		atoms++
		if malformedCompilerFact(m.Content) {
			t.Fatalf("malformed atom persisted: %q", m.Content)
		}
		if strings.TrimSpace(m.SourceText) == "" {
			t.Fatalf("atom missing provenance: %+v", m)
		}
		withProvenance++
		if p, _ := m.Explain["predicate"].(string); p != "" {
			preds[p]++
		}
		if m.Explain["subject"] == nil {
			t.Fatalf("atom missing subject: %+v", m)
		}
	}
	mustContain := []string{
		"jordan is a community organizer",
		"jordan is single",
		"jordan is from portugal",
		"jordan moved from portugal",
		"4 years",
		"counseling",
		"classic children's books",
		"little prince",
		"painting",
		"camping",
		"beach",
		"mountains",
		"swimming",
		"pottery",
		"14 july 2023",
		"the friday before 15 july 2023",
		"running",
		"dinosaurs",
		"fossils",
		"forest",
		"speech",
		"the week before 15 july 2023",
		"elderly patients",
		"hidden garden",
	}
	for _, needle := range mustContain {
		if !strings.Contains(joined, needle) {
			t.Errorf("coverage miss %q in %q", needle, joined)
		}
	}
	for _, pred := range []string{
		PredicateOrigin, PredicateActivity, PredicateOccupation,
		PredicateMediaConsumed, PredicateFamilyMember, PredicateRelationshipStatus,
	} {
		if preds[pred] == 0 {
			t.Errorf("expected predicate %s in audit atoms, got %v", pred, preds)
		}
	}
	if atoms < 12 {
		t.Fatalf("expected dense compiler coverage, got %d atoms: %q", atoms, joined)
	}
	if withProvenance != atoms {
		t.Fatalf("every atom needs source_text")
	}
	_ = at
}

func TestHomeCountryAnaphoraBindsPriorOrigin(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Alex: This necklace is from my home country, Sweden."},
			{Role: "user", Content: "Alex: I've known these friends for 4 years, since I moved from my home country."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, m := range memories {
		joined += " | " + strings.ToLower(m.Content)
	}
	if !strings.Contains(joined, "alex moved from sweden") {
		t.Fatalf("expected anaphora bind Sweden, got %q", joined)
	}
	if !strings.Contains(joined, "4 years") {
		t.Fatalf("expected duration atom, got %q", joined)
	}
}

func TestRelativeFridayStampsWorkshopAtom(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Metadata: map[string]any{"observed_at": "15 July 2023"},
		Messages: []Message{
			{Role: "user", Content: "Sam: Last Fri I finally took my kids to a pottery workshop."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, m := range memories {
		joined += " | " + m.Content
	}
	lower := strings.ToLower(joined)
	if !strings.Contains(lower, "pottery") {
		t.Fatalf("expected pottery activity, got %q", joined)
	}
	if !strings.Contains(lower, "14 july 2023") {
		t.Fatalf("expected last Friday absolute date, got %q", joined)
	}
	if !strings.Contains(lower, "friday before 15 july 2023") {
		t.Fatalf("expected session-relative Friday phrase, got %q", joined)
	}
}

func TestProjectMemoryRelationFromOriginAtom(t *testing.T) {
	rec := MemoryRecord{
		TenantID: "t1", SubjectID: "u1", MemoryID: "m1",
		Content: "Alex moved from Sweden",
		Explain: map[string]any{
			"subject":    "Alex",
			"predicate":  PredicateOrigin,
			"value_norm": "sweden",
			"rule":       "attribute_origin",
		},
	}
	rel, ok := projectMemoryRelation(rec)
	if !ok {
		t.Fatal("expected relation projection")
	}
	if rel.SrcEntity != "alex" || rel.Relation != PredicateOrigin || rel.DstEntity != "sweden" {
		t.Fatalf("got %+v", rel)
	}
}

func TestEnrichLastFriAbbreviation(t *testing.T) {
	at := time.Date(2023, 7, 15, 12, 0, 0, 0, time.UTC) // Saturday
	got := EnrichRelativeEventTime("Last Fri I went to a workshop", at)
	if !strings.Contains(got, "14 July 2023") {
		t.Fatalf("expected last Fri → 14 July, got %q", got)
	}
	if !strings.Contains(strings.ToLower(got), "friday before 15 july 2023") {
		t.Fatalf("expected session-relative phrase, got %q", got)
	}
}

func TestPlaceActivityFindAllDoesNotSwallowNextPlace(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Sam: I've been camping at the beach and later camping at mountains."},
			{Role: "user", Content: "Sam: We went on a camping trip in the forest."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, m := range memories {
		joined += " | " + strings.ToLower(m.Content)
	}
	for _, needle := range []string{"beach", "mountains", "forest"} {
		if !strings.Contains(joined, needle) {
			t.Errorf("expected place %q in %q", needle, joined)
		}
	}
}

func TestLastWeekSpeechStampsSessionRelativePhrase(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Metadata: map[string]any{"observed_at": "9 June 2023"},
		Messages: []Message{
			{Role: "user", Content: "Sam: Last week I gave a speech at a school."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, m := range memories {
		joined += " | " + strings.ToLower(m.Content)
	}
	if !strings.Contains(joined, "speech") {
		t.Fatalf("expected speech event atom, got %q", joined)
	}
	if !strings.Contains(joined, "the week before 9 june 2023") {
		t.Fatalf("expected week-before stamp, got %q", joined)
	}
}

func TestValueNormStripsStampAndPrefersPlace(t *testing.T) {
	got := valueNormFromAtomContent("Riley participates in pottery (14 July 2023; the Friday before 15 July 2023)")
	if got != "pottery" {
		t.Fatalf("pottery value_norm=%q", got)
	}
	place := valueNormFromAtomContent("Riley has done camping at Beach")
	if place != "beach" {
		t.Fatalf("place value_norm=%q", place)
	}
}
