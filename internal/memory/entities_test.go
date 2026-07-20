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
	raw := map[string]float64{"a": 0.9, "b": 0.5, "c": 0.5, "d": 0.5} // mean 0.6
	out := calibrateSimilarities(raw)
	if out["a"] <= 0 {
		t.Fatalf("top match should stay positive, got %v", out["a"])
	}
	for _, k := range []string{"b", "c", "d"} {
		if out[k] != 0 {
			t.Fatalf("below-mean %s should calibrate to 0, got %v", k, out[k])
		}
	}
	flat := calibrateSimilarities(map[string]float64{"x": 0.5, "y": 0.5})
	if flat["x"] != 0 || flat["y"] != 0 {
		t.Fatalf("flat distribution should zero out, got %v", flat)
	}
}
