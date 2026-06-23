package api

import (
	"context"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"brainy/internal/jobs"
	"brainy/internal/memory"
	"brainy/internal/observability"
	"brainy/internal/pack"
	"brainy/internal/store/postgres"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"net/http/httptest"
)

func TestEvalHarnessAgainstHTTPServer(t *testing.T) {
	runEvalHarness(t, "evals/run_eval.py", "fixtures/parity")
}

func TestVerticalEvalHarnessAgainstHTTPServer(t *testing.T) {
	runEvalHarness(t, "evals/run_vertical_eval.py", "fixtures/vertical/marketing")
}

func runEvalHarness(t *testing.T, script, fixtureDir string) {
	t.Helper()
	t.Setenv("LANG", "C")
	t.Setenv("LC_ALL", "C")
	store := startEmbeddedStoreForAPI(t)
	defer store.Close()

	root := repoRoot(t)
	reg, err := pack.LoadRegistryFromDir(filepath.Join(root, "packs"))
	if err != nil {
		t.Fatalf("load packs: %v", err)
	}
	service := memory.NewServiceWithPacks(store, reg)
	server := httptest.NewServer(NewRouter(service, observability.NewMetrics()))
	defer server.Close()

	command := exec.Command("python3", script, "--base-url", server.URL)
	if fixtureDir != "" && script == "evals/run_eval.py" {
		command.Args = append(command.Args, "--fixture-dir", fixtureDir)
	}
	command.Dir = root
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("eval harness failed: %v\n%s", err, string(output))
	}
}

func TestAsyncIngestBecomesSearchableAfterProcessorRuns(t *testing.T) {
	t.Setenv("LANG", "C")
	t.Setenv("LC_ALL", "C")
	store := startEmbeddedStoreForAPI(t)
	defer store.Close()

	service := memory.NewService(store)
	server := httptest.NewServer(NewRouter(service, observability.NewMetrics()))
	defer server.Close()

	asyncBody := `{"tenant_id":"t1","subject_id":"u1","source_type":"conversation","messages":[{"role":"user","content":"I prefer concise answers."}]}`
	response, err := exec.Command(
		"python3",
		"-c",
		"import json, sys, urllib.request; req=urllib.request.Request(sys.argv[1] + '/ingest/async', data=sys.argv[2].encode(), headers={'Content-Type':'application/json'}, method='POST'); print(urllib.request.urlopen(req).read().decode())",
		server.URL,
		asyncBody,
	).CombinedOutput()
	if err != nil {
		t.Fatalf("async ingest failed: %v\n%s", err, string(response))
	}

	searchBefore, err := service.Search(context.Background(), "t1", "u1", "", "How should I respond?")
	if err != nil {
		t.Fatalf("search before processing failed: %v", err)
	}
	if len(searchBefore.Results) != 0 {
		t.Fatalf("expected 0 results before worker processing, got %d", len(searchBefore.Results))
	}

	processor := jobs.NewProcessor(store, observability.NewMetrics())
	processed, err := processor.ProcessNext(context.Background())
	if err != nil {
		t.Fatalf("processor failed: %v", err)
	}
	if !processed {
		t.Fatalf("expected processor to claim one queued job")
	}

	searchAfter, err := service.Search(context.Background(), "t1", "u1", "", "How should I respond?")
	if err != nil {
		t.Fatalf("search after processing failed: %v", err)
	}
	if len(searchAfter.Results) != 1 {
		t.Fatalf("expected 1 result after worker processing, got %d", len(searchAfter.Results))
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
	if err := store.ApplyMigrations(context.Background()); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestCorrectionStickiness(t *testing.T) {
	t.Setenv("LANG", "C")
	t.Setenv("LC_ALL", "C")
	store := startEmbeddedStoreForAPI(t)
	defer store.Close()

	service := memory.NewService(store)
	ctx := context.Background()

	_, err := service.Ingest(ctx, memory.IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []memory.Message{{Role: "user", Content: "I prefer Python."}},
	})
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}

	searchBefore, err := service.Search(ctx, "t1", "u1", "", "How should I respond?")
	if err != nil {
		t.Fatalf("search before correction failed: %v", err)
	}
	if len(searchBefore.Results) == 0 {
		t.Fatal("expected at least one result before correction")
	}
	originalScore := searchBefore.Results[0].Score
	memoryID := searchBefore.Results[0].MemoryID

	_, err = service.Correct(ctx, "t1", "u1", memoryID, memory.CorrectionRequest{
		Content:    "Prefers Ruby",
		SourceText: "Actually I prefer Ruby.",
	})
	if err != nil {
		t.Fatalf("correct failed: %v", err)
	}

	searchAfter, err := service.Search(ctx, "t1", "u1", "", "How should I respond?")
	if err != nil {
		t.Fatalf("search after correction failed: %v", err)
	}
	if len(searchAfter.Results) == 0 {
		t.Fatal("expected at least one result after correction")
	}
	if searchAfter.Results[0].Content != "Prefers Ruby" {
		t.Fatalf("expected corrected content 'Prefers Ruby', got %q", searchAfter.Results[0].Content)
	}
	if searchAfter.Results[0].Score <= originalScore {
		t.Fatalf("expected corrected score > original score, got %f <= %f", searchAfter.Results[0].Score, originalScore)
	}
	if !searchAfter.Results[0].Explain["corrected"].(bool) {
		t.Fatalf("expected corrected flag in explain")
	}

	_, err = service.Ingest(ctx, memory.IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []memory.Message{{Role: "user", Content: "I prefer Python."}},
	})
	if err != nil {
		t.Fatalf("re-ingest failed: %v", err)
	}

	searchFinal, err := service.Search(ctx, "t1", "u1", "", "How should I respond?")
	if err != nil {
		t.Fatalf("search after re-ingest failed: %v", err)
	}
	if len(searchFinal.Results) == 0 || searchFinal.Results[0].Content != "Prefers Ruby" {
		t.Fatalf("correction did not stick after re-ingesting original content")
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}
