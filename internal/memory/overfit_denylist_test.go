package memory_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Anti-MemPalace clause (docs/research/master-plan.md §1.2):
// no LOCOMO / LongMemEval / BEAM surface-forms in product code or prompts.
// *_test.go / test_*.py are exempt.

var overfitDenylist = []string{
	"caroline",
	"melanie",
	"sweden",
	"transgender woman",
	"dinosaur exhibit",
	"charlotte's web",
	"nothing is impossible",
	"becoming nicole",
	"partake",
	"destress",
	"de-stress",
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// internal/memory/overfit_denylist_test.go -> repo root
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestNoBenchmarkSurfaceFormsInProductCode(t *testing.T) {
	root := repoRoot(t)
	roots := []string{
		filepath.Join(root, "internal"),
		filepath.Join(root, "evals", "public"),
	}
	var offenders []string
	for _, scanRoot := range roots {
		_ = filepath.Walk(scanRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			base := info.Name()
			if strings.HasSuffix(base, "_test.go") || strings.HasSuffix(base, "_test.py") || strings.HasPrefix(base, "test_") {
				return nil
			}
			ext := filepath.Ext(base)
			if ext != ".go" && ext != ".py" {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			lower := strings.ToLower(string(data))
			rel, _ := filepath.Rel(root, path)
			for _, needle := range overfitDenylist {
				if strings.Contains(lower, needle) {
					offenders = append(offenders, rel+": "+needle)
				}
			}
			return nil
		})
	}
	if len(offenders) > 0 {
		t.Fatalf("benchmark surface-forms in product code (master-plan W1):\n  %s", strings.Join(offenders, "\n  "))
	}
}
