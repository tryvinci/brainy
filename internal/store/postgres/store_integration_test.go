package postgres

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"brainy/internal/memory"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
)

func TestStoreUpsertListAndSuppressWithEmbeddedPostgres(t *testing.T) {
	ctx := context.Background()
	port := uint32(54329)
	root := t.TempDir()
	runtimePath := "file://" + filepath.Join(root, "runtime")
	dataPath := filepath.Join(root, "data")
	binariesPath := filepath.Join(root, "binaries")

	postgres := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Port(port).
			Username("brainy").
			Password("brainy").
			Database("brainy").
			Version(embeddedpostgres.V17).
			RuntimePath(runtimePath).
			DataPath(dataPath).
			BinariesPath(binariesPath),
	)

	if err := postgres.Start(); err != nil {
		t.Skipf("embedded postgres unavailable: %v", err)
	}
	defer func() {
		_ = postgres.Stop()
	}()

	store, err := New(ctx, "postgres://brainy:brainy@localhost:54329/brainy?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	record := memory.MemoryRecord{
		MemoryID:          "mem_1",
		TenantID:          "t1",
		SubjectID:         "u1",
		Kind:              memory.KindPreference,
		Content:           "Prefers concise answers",
		SourceText:        "I prefer concise answers",
		SourceType:        "conversation",
		DedupeKey:         memory.DedupeKey("t1", "u1", memory.KindPreference, "Prefers concise answers"),
		Status:            memory.StatusActive,
		Confidence:        0.9,
		ExtractionVersion: "deterministic-v1",
		Explain:           map[string]any{"rule": "test"},
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	upserted, err := store.UpsertMemory(ctx, record)
	if err != nil {
		t.Fatal(err)
	}
	if upserted.State != "created" {
		t.Fatalf("expected created state, got %s", upserted.State)
	}

	memories, err := store.ListActiveMemories(ctx, "t1", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if len(memories) != 1 {
		t.Fatalf("expected 1 memory, got %d", len(memories))
	}

	if err := store.SuppressMemory(ctx, "t1", "u1", "mem_1"); err != nil {
		t.Fatal(err)
	}

	memoriesAfterSuppress, err := store.ListActiveMemories(ctx, "t1", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if len(memoriesAfterSuppress) != 0 {
		t.Fatalf("expected 0 memories after suppress, got %d", len(memoriesAfterSuppress))
	}
}
