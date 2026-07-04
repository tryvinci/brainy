package embedding

import "testing"

func TestEmbedParaphraseSimilarity(t *testing.T) {
	original := Embed("Brand voice is warm concise and second-person addressing.")
	paraphrase := Embed("Friendly brief tone speaking directly to the reader.")
	unrelated := Embed("Database migration completed successfully.")

	sim := CosineSimilarity(original, paraphrase)
	if sim < 0.15 {
		t.Fatalf("expected paraphrase similarity >= 0.15, got %v", sim)
	}
	if CosineSimilarity(original, unrelated) >= sim {
		t.Fatalf("expected paraphrase closer than unrelated: paraphrase=%v unrelated=%v", sim, CosineSimilarity(original, unrelated))
	}
}

func TestEmbedDeterministic(t *testing.T) {
	a := Embed("Never mention competitor Acme.")
	b := Embed("Never mention competitor Acme.")
	if CosineSimilarity(a, b) != 1 {
		t.Fatal("expected identical text to produce identical normalized vectors")
	}
}
