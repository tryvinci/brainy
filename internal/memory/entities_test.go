package memory

import (
	"context"
	"testing"
)

func TestExtractEntities(t *testing.T) {
	ents := ExtractEntities(`Caroline: I loved reading "Charlotte's Web" in 2013 with Melanie`)
	want := map[string]bool{"caroline": true, "charlotte's web": true, "melanie": true, "2013": true}
	got := map[string]bool{}
	for _, e := range ents {
		got[e] = true
	}
	for w := range want {
		if !got[w] {
			t.Fatalf("expected entity %q in %v", w, ents)
		}
	}
	if got["i"] || got["the"] {
		t.Fatalf("stopwords leaked into entities: %v", ents)
	}
}

func TestEntityOverlapBoost(t *testing.T) {
	q := ExtractEntities("What did Melanie read?")
	r := ExtractEntities(`Melanie: I loved "Charlotte's Web"`)
	if entityOverlapBoost(q, r) <= 0 {
		t.Fatalf("expected entity overlap boost for shared Melanie, q=%v r=%v", q, r)
	}
	if entityOverlapBoost(q, ExtractEntities("Caroline went hiking")) != 0 {
		t.Fatal("expected no boost when no shared entity")
	}
}

func TestEntityRankingToggleCovered(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store).WithEntityRanking(true)
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Metadata: map[string]any{"session_id": "s1"},
		Messages: []Message{{Role: "user", Content: `Melanie: I loved "Charlotte's Web"`}},
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := svc.Search(context.Background(), "t1", "u1", "", "", `What book did Melanie love?`)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Results) == 0 {
		t.Fatal("expected results with entity ranking enabled")
	}
}

func TestCalibrateSimilarities(t *testing.T) {
	// >=6 candidates with a clear top match: min-max normalizes, top stays 1.0,
	// baseline noise floor suppressed toward 0.
	raw := map[string]float64{"a": 0.9, "b": 0.5, "c": 0.5, "d": 0.5, "e": 0.5, "f": 0.48}
	out := calibrateSimilarities(raw)
	if out["a"] <= 0.9 {
		t.Fatalf("top match should stay near 1.0, got %v", out["a"])
	}
	if out["f"] > 0.2 {
		t.Fatalf("noise floor should be suppressed, got %v", out["f"])
	}
	// Small sets are returned unchanged (avoid erasing a lone true match).
	small := map[string]float64{"x": 0.75, "y": 0.5, "z": 0.48}
	got := calibrateSimilarities(small)
	if got["x"] != 0.75 {
		t.Fatalf("small set should be unchanged, got %v", got)
	}
}
