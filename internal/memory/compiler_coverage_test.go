package memory

import (
	"context"
	"testing"
)

// TestSemanticCoverageAudit is the S1a one-command gate:
//
//	go test ./internal/memory -run TestSemanticCoverageAudit -count=1
//
// Held-out Jordan/Riley/Dana claims (not a benchmark transcript).
func TestSemanticCoverageAudit(t *testing.T) {
	at := mustParseTime(t, "2023-07-15T12:00:00Z")
	ext := NewDeterministicExtractor()
	memories, err := ext.Extract(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u-s1a", SourceType: "conversation",
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
			{Role: "user", Content: "Riley: Last Fri I finally took my kids to a pottery workshop."},
			{Role: "user", Content: "Dana: Yesterday I took the kids to the museum."},
			{Role: "assistant", Content: "Riley researched wildfire recovery last spring."},
			{Role: "user", Content: "Casey: Riley lives in Portland now."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	claims := []CoverageClaim{
		{Text: "Jordan community organizer", Subject: "Jordan"},
		{Text: "Jordan is single", Subject: "Jordan"},
		{Text: "Jordan from Portugal", Subject: "Jordan"},
		{Text: "Jordan moved Portugal", Subject: "Jordan"},
		{Text: "Jordan counseling career", Subject: "Jordan"},
		{Text: "Jordan children's books", Subject: "Jordan"},
		{Text: "Little Prince", Subject: "Jordan"},
		{Text: "Riley painting", Subject: "Riley"},
		{Text: "Riley camping beach", Subject: "Riley"},
		{Text: "Riley pottery", Subject: "Riley", MustDate: true},
		{Text: "Dana museum", Subject: "Dana", MustDate: true},
		{Text: "Riley researched wildfire", Subject: "Riley"},
		{Text: "Riley lives Portland", Subject: "Riley"},
	}
	rep := ScoreSemanticCoverage(memories, claims, &at)
	if rep.CompiledRate() < 0.85 {
		t.Fatalf("compiled rate %.2f (%d/%d) below 85%% gate; entity=%d dated=%d evidenced=%d",
			rep.CompiledRate(), rep.Compiled, rep.Claims, rep.EntityBound, rep.Dated, rep.Evidenced)
	}
	if rep.EntityBound < int(0.7*float64(rep.Compiled)) {
		t.Fatalf("entity-bound %d/%d compiled is thin", rep.EntityBound, rep.Compiled)
	}
	if rep.Evidenced < int(0.8*float64(rep.Compiled)) {
		t.Fatalf("evidenced %d/%d compiled is thin", rep.Evidenced, rep.Compiled)
	}
}

func TestScoreSemanticCoverageMissesUncompiled(t *testing.T) {
	rep := ScoreSemanticCoverage(nil, []CoverageClaim{{Text: "Alex pottery instructor", Subject: "Alex"}}, nil)
	if rep.Compiled != 0 || rep.Claims != 1 {
		t.Fatalf("%+v", rep)
	}
}
