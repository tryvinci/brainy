package postgres

import (
	"context"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"brainy/internal/memory"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
)

func TestStoreUpsertListAndSuppressWithEmbeddedPostgres(t *testing.T) {
	t.Setenv("LANG", "C")
	t.Setenv("LC_ALL", "C")
	ctx := context.Background()
	port := randomPort(0)
	root := t.TempDir()
	dataPath := filepath.Join(root, "data")
	binariesPath := filepath.Join(root, "binaries")

	postgres := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Port(port).
			Username("brainy").
			Password("brainy").
			Database("brainy").
			Version(embeddedpostgres.V17).
			RuntimePath(filepath.Join(root, "runtime")).
			DataPath(dataPath).
			BinariesPath(binariesPath),
	)

	if err := postgres.Start(); err != nil {
		t.Fatalf("embedded postgres unavailable: %v", err)
	}
	defer func() {
		_ = postgres.Stop()
	}()

	store, err := New(ctx, dsn(port))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.ApplyMigrations(ctx); err != nil {
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

func TestStoreUpsertIsConcurrencySafeForDuplicateIngests(t *testing.T) {
	t.Setenv("LANG", "C")
	t.Setenv("LC_ALL", "C")
	ctx := context.Background()
	root := t.TempDir()
	port := randomPort(101)
	postgres := embeddedpostgres.NewDatabase(
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

	if err := postgres.Start(); err != nil {
		t.Fatalf("embedded postgres unavailable: %v", err)
	}
	defer func() {
		_ = postgres.Stop()
	}()

	store, err := New(ctx, dsn(port))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.ApplyMigrations(ctx); err != nil {
		t.Fatal(err)
	}

	const attempts = 8
	errs := make(chan error, attempts)
	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			now := time.Now().UTC()
			_, err := store.UpsertMemory(ctx, memory.MemoryRecord{
				MemoryID:          "mem_concurrent_" + strconv.Itoa(index),
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
			})
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("unexpected concurrent upsert error: %v", err)
		}
	}

	records, err := store.ListActiveMemories(ctx, "t1", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 active memory after concurrent upserts, got %d", len(records))
	}
}

func dsn(port uint32) string {
	return "postgres://brainy:brainy@localhost:" + fmtUint(port) + "/brainy?sslmode=disable"
}

func fmtUint(value uint32) string {
	return strconv.FormatUint(uint64(value), 10)
}

func randomPort(offset uint32) uint32 {
	return 45000 + uint32(time.Now().UTC().UnixNano()%5000) + offset
}

func TestApplyMigrationsSupportsFreshAndUpgradeFlows(t *testing.T) {
	t.Setenv("LANG", "C")
	t.Setenv("LC_ALL", "C")
	ctx := context.Background()
	root := t.TempDir()
	port := randomPort(202)
	postgres := embeddedpostgres.NewDatabase(
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
	if err := postgres.Start(); err != nil {
		t.Fatalf("embedded postgres unavailable: %v", err)
	}
	defer func() {
		_ = postgres.Stop()
	}()

	store, err := New(ctx, dsn(port))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.ApplyMigrations(ctx); err != nil {
		t.Fatalf("fresh migrations failed: %v", err)
	}
	if err := store.ApplyMigrations(ctx); err != nil {
		t.Fatalf("idempotent migration rerun failed: %v", err)
	}

	if _, err := store.pool.Exec(ctx, `
DROP TABLE IF EXISTS extraction_jobs;
DROP TABLE IF EXISTS raw_ingests;
DELETE FROM schema_migrations WHERE version = 2;
`); err != nil {
		t.Fatalf("failed to simulate v1 schema state: %v", err)
	}

	if err := store.ApplyMigrations(ctx); err != nil {
		t.Fatalf("upgrade migrations failed: %v", err)
	}

	var rawIngestsExists bool
	if err := store.pool.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_name = 'raw_ingests'
)
`).Scan(&rawIngestsExists); err != nil {
		t.Fatal(err)
	}
	if !rawIngestsExists {
		t.Fatalf("expected raw_ingests table after upgrade migration")
	}

	var appliedVersionTwo bool
	if err := store.pool.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
    FROM schema_migrations
    WHERE version = 2
)
`).Scan(&appliedVersionTwo); err != nil {
		t.Fatal(err)
	}
	if !appliedVersionTwo {
		t.Fatalf("expected migration version 2 to be recorded after upgrade")
	}
}
