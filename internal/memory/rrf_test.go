package memory

import (
	"math"
	"testing"
)

func TestReciprocalRankFusionBasic(t *testing.T) {
	dense := []string{"a", "b", "c"}
	lex := []string{"b", "a", "d"}
	got := ReciprocalRankFusion([][]string{dense, lex}, 60)
	// a: 1/(60+1) + 1/(60+2)
	wantA := 1.0/61.0 + 1.0/62.0
	wantB := 1.0/62.0 + 1.0/61.0
	wantC := 1.0 / 63.0
	wantD := 1.0 / 63.0
	if math.Abs(got["a"]-wantA) > 1e-12 {
		t.Fatalf("a=%v want %v", got["a"], wantA)
	}
	if math.Abs(got["b"]-wantB) > 1e-12 {
		t.Fatalf("b=%v want %v", got["b"], wantB)
	}
	if math.Abs(got["c"]-wantC) > 1e-12 || math.Abs(got["d"]-wantD) > 1e-12 {
		t.Fatalf("c=%v d=%v", got["c"], got["d"])
	}
	if got["a"] != got["b"] {
		t.Fatalf("symmetric ranks should tie a=%v b=%v", got["a"], got["b"])
	}
}

func TestReciprocalRankFusionDedupsSameList(t *testing.T) {
	got := ReciprocalRankFusion([][]string{{"a", "a", "b"}}, 60)
	if math.Abs(got["a"]-1.0/61.0) > 1e-12 {
		t.Fatalf("duplicate id counted twice: %v", got["a"])
	}
	if math.Abs(got["b"]-1.0/62.0) > 1e-12 {
		t.Fatalf("b rank should be 2 after dedup: %v", got["b"])
	}
}

func TestReciprocalRankFusionEmptyAndDefaultK(t *testing.T) {
	if len(ReciprocalRankFusion(nil, 0)) != 0 {
		t.Fatal("empty lists")
	}
	got := ReciprocalRankFusion([][]string{{"x"}}, 0)
	if math.Abs(got["x"]-1.0/61.0) > 1e-12 {
		t.Fatalf("k<=0 should default to 60: %v", got["x"])
	}
}

func TestRankIDsByScore(t *testing.T) {
	ids := RankIDsByScore(map[string]float64{"b": 0.2, "a": 0.9, "z": 0, "c": 0.9})
	want := []string{"a", "c", "b"}
	if len(ids) != 3 || ids[0] != "a" || ids[1] != "c" || ids[2] != "b" {
		t.Fatalf("got %v want %v", ids, want)
	}
}

func TestRRFToSearchScoreOrderPreserved(t *testing.T) {
	low := rrfToSearchScore(1.0 / 61.0)
	high := rrfToSearchScore(1.0/61.0 + 1.0/61.0)
	if !(high > low && low > 1) {
		t.Fatalf("low=%v high=%v", low, high)
	}
	if rrfToSearchScore(0) != 0 {
		t.Fatal("zero")
	}
}

func TestRetrievalRRFEnabledDefaultOff(t *testing.T) {
	t.Setenv("BRAINY_RETRIEVAL_RRF", "")
	if RetrievalRRFEnabled() {
		t.Fatal("default must be off")
	}
	t.Setenv("BRAINY_RETRIEVAL_RRF", "1")
	if !RetrievalRRFEnabled() {
		t.Fatal("1 must enable")
	}
	t.Setenv("BRAINY_RETRIEVAL_RRF", "false")
	if RetrievalRRFEnabled() {
		t.Fatal("false must disable")
	}
}
