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
