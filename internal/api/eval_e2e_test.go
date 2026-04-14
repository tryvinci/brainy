package api

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"

	"brainy/internal/memory"
	"brainy/internal/store/postgres"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"net/http/httptest"
)

func TestEvalHarnessAgainstHTTPServer(t *testing.T) {
	store := startEmbeddedStoreForAPI(t)
	defer store.Close()

	service := memory.NewService(store)
	server := httptest.NewServer(NewRouter(service))
	defer server.Close()

	repoRoot := repoRoot(t)
	command := exec.Command("python3", "evals/run_eval.py", "--base-url", server.URL)
	command.Dir = repoRoot
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("eval harness failed: %v\n%s", err, string(output))
	}
}

func startEmbeddedStoreForAPI(t *testing.T) *postgres.Store {
	t.Helper()

	root := t.TempDir()
	port := uint32(54529)
	db := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Port(port).
			Username("brainy").
			Password("brainy").
			Database("brainy").
			Version(embeddedpostgres.V17).
			RuntimePath("file://" + filepath.Join(root, "runtime")).
			DataPath(filepath.Join(root, "data")).
			BinariesPath(filepath.Join(root, "binaries")),
	)

	if err := db.Start(); err != nil {
		t.Skipf("embedded postgres unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Stop()
	})

	store, err := postgres.New(context.Background(), "postgres://brainy:brainy@localhost:54529/brainy?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.EnsureSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	return store
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}
