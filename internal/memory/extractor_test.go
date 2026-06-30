package memory

import "testing"

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

func TestDedupeKeyIsStable(t *testing.T) {
	left := DedupeKey("t1", "u1", KindPreference, "Prefers concise answers")
	right := DedupeKey("t1", "u1", KindPreference, "Prefers concise answers")
	if left != right {
		t.Fatalf("expected dedupe key to be stable")
	}
}
