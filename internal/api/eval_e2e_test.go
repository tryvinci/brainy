package api

import (
	"context"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"brainy/internal/memory"
	"brainy/internal/store/postgres"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"net/http/httptest"
)

func TestEvalHarnessAgainstHTTPServer(t *testing.T) {
	t.Setenv("LANG", "C")
	t.Setenv("LC_ALL", "C")
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
	port := 52000 + uint32(time.Now().UTC().UnixNano()%5000)
	db := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Port(port).
			Username("brainy").
			Password("brainy").
			Database("brainy").
			Version(embeddedpostgres.V17).
			RuntimePath(filepath.Join(root, "runtime")).
			DataPath(filepath.Join(root, "data")).
			BinariesPath(filepath.Join(root, "binaries")),
	)

	if err := db.Start(); err != nil {
		t.Fatalf("embedded postgres unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Stop()
	})

	store, err := postgres.New(context.Background(), "postgres://brainy:brainy@localhost:"+strconv.FormatUint(uint64(port), 10)+"/brainy?sslmode=disable")
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
