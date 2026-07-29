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
			{Role: "user", Content: "Alex: I am a transgender woman"},
			{Role: "user", Content: "Alex: I moved from Sweden four years ago"},
			{Role: "user", Content: "Thanks, Alex"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, m := range memories {
		joined += " " + strings.ToLower(m.Content)
	}
	if !strings.Contains(joined, "alex is a transgender woman") && !strings.Contains(joined, "alex is transgender") {
		t.Fatalf("expected identity atom, got %q", joined)
	}
	if !strings.Contains(joined, "alex moved from sweden") {
		t.Fatalf("expected origin atom, got %q", joined)
	}
}

func TestAttributeAtomsTitlesAndActivities(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: `Sam: I loved reading "Charlotte's Web" as a kid`},
			{Role: "user", Content: "Sam: I'm a big fan of pottery - the creativity is awesome"},
			{Role: "user", Content: "Sam: I've been camping in the mountains"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, m := range memories {
		joined += " | " + strings.ToLower(m.Content)
	}
	if !strings.Contains(joined, "charlotte") {
		t.Fatalf("expected titled work atom, got %q", joined)
	}
	if !strings.Contains(joined, "pottery") {
		t.Fatalf("expected pottery activity atom, got %q", joined)
	}
	if !strings.Contains(joined, "camping") {
		t.Fatalf("expected camping activity atom, got %q", joined)
	}
}

func TestAttributeAtomsSkippedForPackLabeledIngest(t *testing.T) {
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation", Label: "brand_rule",
		Messages: []Message{
			{Role: "user", Content: "Alex: I am a transgender woman"},
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
