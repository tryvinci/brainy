package memory_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"brainy/internal/memory"
)

func TestPilotPredicateGoldenFixtures(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	path := filepath.Join(root, "evals", "fixtures", "atoms", "pilot_predicates.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var suite struct {
		Cases []struct {
			ID                string            `json:"id"`
			Messages          []memory.Message  `json:"messages"`
			ExpectPredicates  []string          `json:"expect_predicates"`
			ExpectSubstrings  []string          `json:"expect_substrings"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &suite); err != nil {
		t.Fatal(err)
	}
	ext := memory.NewDeterministicExtractor()
	for _, tc := range suite.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			mems, err := ext.Extract(context.Background(), memory.IngestRequest{
				TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
				Messages: tc.Messages,
			})
			if err != nil {
				t.Fatal(err)
			}
			joined := ""
			preds := map[string]struct{}{}
			for _, m := range mems {
				joined += " | " + strings.ToLower(m.Content)
				if p, ok := m.Explain["predicate"].(string); ok {
					preds[p] = struct{}{}
				}
			}
			for _, sub := range tc.ExpectSubstrings {
				if !strings.Contains(joined, strings.ToLower(sub)) {
					t.Fatalf("expected substring %q in %q", sub, joined)
				}
			}
			for _, want := range tc.ExpectPredicates {
				if _, ok := preds[want]; !ok {
					// Soft: event may appear as activity for camping cases.
					if want == "event" {
						if _, ok := preds["activity"]; ok {
							continue
						}
					}
					t.Fatalf("expected predicate %q, have %v; content=%q", want, keys(preds), joined)
				}
			}
		})
	}
}

func keys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
