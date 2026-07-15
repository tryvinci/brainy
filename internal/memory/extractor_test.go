package memory

import (
	"strings"
	"testing"
)

func TestExtractorExtractsPreferenceProfileAndFact(t *testing.T) {
	extractor := NewExtractor()
	req := IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "I prefer concise answers. I work at Acme. Launch date is May 12."},
		},
	}

	memories := extractor.Extract(req)
	if len(memories) != 3 {
		t.Fatalf("expected 3 memories, got %d", len(memories))
	}
	if memories[0].Kind != KindPreference {
		t.Fatalf("expected first memory to be preference, got %s", memories[0].Kind)
	}
	if memories[1].Kind != KindProfile {
		t.Fatalf("expected second memory to be profile, got %s", memories[1].Kind)
	}
	if memories[2].Kind != KindFact {
		t.Fatalf("expected third memory to be fact, got %s", memories[2].Kind)
	}
}

func TestExtractorRetainsFreeDialogueEpisodes(t *testing.T) {
	extractor := NewExtractor()
	req := IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Caroline: Hey Mel! Good to see you!"},
			{Role: "user", Content: "Caroline: I went to the LGBTQ support group on 7 May 2023"},
			{Role: "user", Content: "Melanie: Can't wait to see your show - the LGBTQ community needs more platforms like this"},
		},
	}

	memories := extractor.Extract(req)
	if len(memories) < 3 {
		t.Fatalf("expected at least 3 memories from free dialogue, got %d: %+v", len(memories), memories)
	}

	joined := ""
	hasEpisode := false
	for _, m := range memories {
		joined += " " + strings.ToLower(m.Content)
		if rule, _ := m.Explain["rule"].(string); rule == "conversation_episode" {
			hasEpisode = true
		}
	}
	if !strings.Contains(joined, "7 may 2023") && !strings.Contains(joined, "lgbtq support group") {
		t.Fatalf("expected LGBTQ date fact retained, got contents: %s", joined)
	}
	if !hasEpisode {
		t.Fatalf("expected at least one conversation_episode rule, got %+v", memories)
	}
}

func TestSplitUtterancesKeepsNewlinesAtomic(t *testing.T) {
	units := splitUtterances("Caroline: went on 7 May 2023\nMelanie: cool show about LGBTQ")
	if len(units) != 2 {
		t.Fatalf("expected 2 utterance units, got %d %v", len(units), units)
	}
}

func TestExtractorSkipsEpisodesWhenPackLabelSet(t *testing.T) {
	extractor := NewExtractor()
	req := IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Label:      "campaign",
		Messages: []Message{
			{Role: "user", Content: "Caroline: I went to the LGBTQ support group on 7 May 2023"},
		},
	}

	memories := extractor.Extract(req)
	for _, m := range memories {
		if rule, _ := m.Explain["rule"].(string); rule == "conversation_episode" {
			t.Fatalf("pack-labeled ingest must not create conversation_episode, got %+v", memories)
		}
	}
	if len(memories) != 0 {
		t.Fatalf("expected keyword miss + no episodes (label path owns packing), got %+v", memories)
	}
}

func TestDedupeKeyIsStable(t *testing.T) {
	left := DedupeKey("t1", "u1", KindPreference, "Prefers concise answers")
	right := DedupeKey("t1", "u1", KindPreference, "Prefers concise answers")
	if left != right {
		t.Fatalf("expected dedupe key to be stable")
	}
}
