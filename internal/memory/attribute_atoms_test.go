package memory

import (
	"context"
	"strings"
	"testing"
)

func TestAttributeAtomsIdentityAndOrigin(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Alex: I am a software engineer"},
			{Role: "user", Content: "Alex: I moved from Canada four years ago"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, m := range memories {
		joined += " | " + strings.ToLower(m.Content)
	}
	if !strings.Contains(joined, "alex is a software engineer") {
		t.Fatalf("expected identity atom, got %q", joined)
	}
	if !strings.Contains(joined, "alex moved from canada") {
		t.Fatalf("expected origin atom, got %q", joined)
	}
	for _, m := range memories {
		rule, _ := m.Explain["rule"].(string)
		if !strings.HasPrefix(rule, "attribute_") {
			continue
		}
		if p, _ := m.Explain["primitive"].(string); p == PrimitiveEpisode {
			t.Fatalf("attribute atoms must not be provenance episodes: %+v", m)
		}
	}
}

func TestAttributeAtomsTitlesAndActivities(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: `Sam: I loved reading "The Little Prince" as a kid`},
			{Role: "user", Content: "Sam: I'm a big fan of ceramics - the creativity is awesome"},
			{Role: "user", Content: "Sam: I've been hiking in the mountains"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, m := range memories {
		joined += " | " + strings.ToLower(m.Content)
	}
	if !strings.Contains(joined, "little prince") {
		t.Fatalf("expected titled work atom, got %q", joined)
	}
	if !strings.Contains(joined, "ceramics") {
		t.Fatalf("expected ceramics activity atom, got %q", joined)
	}
	if !strings.Contains(joined, "hiking") {
		t.Fatalf("expected hiking activity atom, got %q", joined)
	}
}

func TestSpeakerCarryForwardForOriginAtoms(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Alex: Thanks, Sam!"},
			{Role: "user", Content: "This necklace is from my grandma in my home country, Canada."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, m := range memories {
		joined += " | " + m.Content
	}
	if !strings.Contains(strings.ToLower(joined), "alex") || !strings.Contains(strings.ToLower(joined), "canada") {
		t.Fatalf("expected Alex+Canada atom with carry-forward, got %q", joined)
	}
	if strings.Contains(joined, "User moved") || strings.Contains(joined, "Someone moved") {
		t.Fatalf("must not invent User/Someone speaker, got %q", joined)
	}
}

func TestAttributeAtomsSkippedForPackLabeledIngest(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation", Label: "brand_rule",
		Messages: []Message{
			{Role: "user", Content: "Alex: I am a software engineer"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range memories {
		if rule, _ := m.Explain["rule"].(string); strings.HasPrefix(rule, "attribute_") {
			t.Fatalf("pack-labeled ingest must not emit attribute atoms, got %+v", m)
		}
	}
}

func TestAttributeAtomsRejectMalformedTemplates(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Alex: I've been going at this in my life since then"},
			{Role: "user", Content: "Alex: Getting in touch with ourselves is the point"},
			{Role: "user", Content: "Alex: I've been running every morning"},
			{Role: "user", Content: `Alex: I loved reading "The Little Prince" as a kid`},
			{Role: "user", Content: "Alex: I've been hiking in the mountains"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, m := range memories {
		joined += " | " + m.Content
		if malformedCompilerFact(m.Content) {
			t.Fatalf("extractor persisted malformed atom: %q", m.Content)
		}
	}
	lower := strings.ToLower(joined)
	if strings.Contains(lower, "has done going at") || strings.Contains(lower, "has done getting at") {
		t.Fatalf("light-verb place atoms must not persist, got %q", joined)
	}
	if strings.Contains(lower, "participates in runn") && !strings.Contains(lower, "participates in running") {
		t.Fatalf("failed gerund stem must not persist, got %q", joined)
	}
	if !strings.Contains(lower, "participates in running") {
		t.Fatalf("expected gerund activity atom, got %q", joined)
	}
	if !strings.Contains(lower, "little prince") {
		t.Fatalf("expected well-formed title atom, got %q", joined)
	}
	if !strings.Contains(lower, "hiking") {
		t.Fatalf("expected hiking place/activity atom, got %q", joined)
	}
}

func TestMalformedCompilerFactDetector(t *testing.T) {
	bad := []string{
		"Caroline has done going at since then",
		"Caroline has done going at in my life",
		"Melanie participates in runn",
		"Caroline mentioned \"ve got lots of kids\"",
		"Melanie has done taking at kids in need - you",
	}
	for _, c := range bad {
		if !malformedCompilerFact(c) {
			t.Fatalf("expected malformed: %q", c)
		}
	}
	good := []string{
		"Alex participates in running",
		"Alex has done hiking at mountains",
		"Alex mentioned \"The Little Prince\"",
		"Alex is from Canada",
		"The launch date is June 3.",
		"Alex participates in pottery (14 July 2023; the Friday before 15 July 2023)",
	}
	for _, c := range good {
		if malformedCompilerFact(c) {
			t.Fatalf("expected well-formed: %q", c)
		}
	}
}

func TestAsARoleEmitsIdentityAtom(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Jordan: I made this piece to show my journey as a community organizer"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, m := range memories {
		joined += " | " + strings.ToLower(m.Content)
	}
	if !strings.Contains(joined, "jordan is a community organizer") {
		t.Fatalf("expected as-a-role identity atom, got %q", joined)
	}
}

func TestKidsPreferenceFindAllSkipsTemporalJunk(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Riley: The kids were wild about fossils at the museum."},
			{Role: "user", Content: "Riley: The kids love nature."},
			{Role: "user", Content: "Riley: The kids were talking about our last one over summer break."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, m := range memories {
		joined += " | " + strings.ToLower(m.Content)
	}
	if !strings.Contains(joined, "fossils") {
		t.Fatalf("expected fossils preference, got %q", joined)
	}
	if !strings.Contains(joined, "nature") {
		t.Fatalf("expected nature preference, got %q", joined)
	}
	if strings.Contains(joined, "kids like our last") || strings.Contains(joined, "kids like summer") {
		t.Fatalf("junk preference compiled: %q", joined)
	}
}

func TestKidsPronounBindsAfterMention(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Dana: Yesterday I took the kids to the museum."},
			{Role: "user", Content: "Dana: They were stoked for the fossils exhibit! They love nature."},
			{Role: "user", Content: "Sam: They were excited about jazz night."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, m := range memories {
		joined += " | " + strings.ToLower(m.Content)
	}
	if !strings.Contains(joined, "fossils") {
		t.Fatalf("expected pronoun-bound kids like, got %q", joined)
	}
	if !strings.Contains(joined, "nature") {
		t.Fatalf("expected they-love kids like, got %q", joined)
	}
	if strings.Contains(joined, "sam's kids like jazz") {
		t.Fatalf("unmentioned kids should not bind they-likes: %q", joined)
	}
}

func TestPronounExcitedEmitsSpeakerPreferenceWithoutKids(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Dana: They were stoked for the fossils exhibit!"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, m := range memories {
		joined += " | " + strings.ToLower(m.Content)
	}
	if !strings.Contains(joined, "fossils") {
		t.Fatalf("expected speaker enjoys exhibit noun, got %q", joined)
	}
}

func TestKidsPronounBindsStokedExhibitInSameBatch(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Dana: Congrats on following your dreams. Yesterday I took the kids to the museum - it was so cool spending time with them!"},
			{Role: "user", Content: "Sam: What were they so stoked about?"},
			{Role: "user", Content: "Dana: They were stoked for the fossils exhibit! They love learning about animals and the bones were so cool."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, m := range memories {
		joined += " | " + strings.ToLower(m.Content)
	}
	if !strings.Contains(joined, "fossils") {
		t.Fatalf("expected exhibit noun from they-stoked, got %q", joined)
	}
}

func TestWorkWithGroupEmitsOccupationQualifier(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Jordan: I'm thinking of working with elderly patients."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, m := range memories {
		joined += " | " + strings.ToLower(m.Content)
	}
	if !strings.Contains(joined, "elderly patients") {
		t.Fatalf("expected career-for group, got %q", joined)
	}
}

func TestResearchingTopicEmitsPlanAtom(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Jordan: Researching scholarship programs — it's been a dream to help."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, m := range memories {
		joined += " | " + strings.ToLower(m.Content)
	}
	if !strings.Contains(joined, "researched scholarship programs") {
		t.Fatalf("expected researched topic, got %q", joined)
	}
}

func TestUnquotedTitleRunAndOneWordQuote(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Riley: This book I read last year, The Hidden Garden, still stays with me."},
			{Role: "user", Content: `Riley: I read "Perfect"`},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, m := range memories {
		rule, _ := m.Explain["rule"].(string)
		if !strings.HasPrefix(rule, "attribute_") {
			continue
		}
		joined += " | " + strings.ToLower(m.Content)
	}
	if !strings.Contains(joined, "hidden garden") {
		t.Fatalf("expected title-case run, got %q", joined)
	}
	if strings.Contains(joined, `"perfect"`) {
		t.Fatalf("one-word quote should not compile, got %q", joined)
	}
}
