package postgres

import (
	"context"
	"errors"
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

func TestApplyMigrationsIsSafeUnderConcurrentStartup(t *testing.T) {
	t.Setenv("LANG", "C")
	t.Setenv("LC_ALL", "C")
	ctx := context.Background()
	root := t.TempDir()
	port := randomPort(404)
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

	runMigration := func() error {
		store, err := New(ctx, dsn(port))
		if err != nil {
			return err
		}
		defer store.Close()
		return store.ApplyMigrations(ctx)
	}

	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- runMigration()
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("unexpected concurrent migration error: %v", err)
		}
	}
}

func TestCorrectMemoryReturnsConflictOnDuplicateContent(t *testing.T) {
	t.Setenv("LANG", "C")
	t.Setenv("LC_ALL", "C")
	ctx := context.Background()
	root := t.TempDir()
	port := randomPort(505)
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

	now := time.Now().UTC()
	first := memory.MemoryRecord{
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
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	second := memory.MemoryRecord{
		MemoryID:          "mem_2",
		TenantID:          "t1",
		SubjectID:         "u1",
		Kind:              memory.KindPreference,
		Content:           "Prefers detailed answers",
		SourceText:        "I prefer detailed answers",
		SourceType:        "conversation",
		DedupeKey:         memory.DedupeKey("t1", "u1", memory.KindPreference, "Prefers detailed answers"),
		Status:            memory.StatusActive,
		Confidence:        0.9,
		ExtractionVersion: "deterministic-v1",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if _, err := store.UpsertMemory(ctx, first); err != nil {
		t.Fatal(err)
	}
	if _, err := store.UpsertMemory(ctx, second); err != nil {
		t.Fatal(err)
	}

	_, err = store.CorrectMemory(ctx, "t1", "u1", "mem_2", "Prefers concise answers", "Prefers concise answers")
	if !errors.Is(err, memory.ErrMemoryConflict) {
		t.Fatalf("expected memory conflict, got %v", err)
	}
}

func TestUpsertMemoryPreservesSuppressedStatus(t *testing.T) {
	t.Setenv("LANG", "C")
	t.Setenv("LC_ALL", "C")
	ctx := context.Background()
	root := t.TempDir()
	port := randomPort(605)
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

	now := time.Now().UTC()
	record := memory.MemoryRecord{
		MemoryID:          "mem_sup",
		TenantID:          "t1",
		SubjectID:         "u1",
		Kind:              memory.KindFact,
		Content:           "Never share the door code with vendors",
		SourceText:        "Never share the door code with vendors.",
		SourceType:        "conversation",
		DedupeKey:         memory.DedupeKey("t1", "u1", memory.KindFact, "Never share the door code with vendors"),
		Status:            memory.StatusActive,
		Confidence:        0.95,
		ExtractionVersion: "deterministic-v1",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if _, err := store.UpsertMemory(ctx, record); err != nil {
		t.Fatal(err)
	}
	if err := store.SuppressMemory(ctx, "t1", "u1", "mem_sup"); err != nil {
		t.Fatal(err)
	}

	reingest := record
	reingest.MemoryID = "mem_sup_new"
	reingest.UpdatedAt = now.Add(time.Minute)
	upserted, err := store.UpsertMemory(ctx, reingest)
	if err != nil {
		t.Fatal(err)
	}
	if upserted.State != "deduped" {
		t.Fatalf("expected deduped re-ingest of suppressed memory, got %q", upserted.State)
	}
	if upserted.Record.Status != memory.StatusSuppressed {
		t.Fatalf("expected suppressed status to persist, got %q", upserted.Record.Status)
	}

	results, err := store.SearchActiveMemories(ctx, "t1", "u1", []string{"%door%"}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected suppressed memory to stay out of search, got %d results", len(results))
	}
}

func TestClaimNextExtractionJobReclaimsExpiredInProgressJob(t *testing.T) {
	t.Setenv("LANG", "C")
	t.Setenv("LC_ALL", "C")
	ctx := context.Background()
	root := t.TempDir()
	port := randomPort(606)
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

	baseStore, err := New(ctx, dsn(port))
	if err != nil {
		t.Fatal(err)
	}
	defer baseStore.Close()

	if err := baseStore.ApplyMigrations(ctx); err != nil {
		t.Fatal(err)
	}

	store := &Store{pool: baseStore.pool, jobLease: 5 * time.Millisecond}
	req := memory.IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages: []memory.Message{
			{Role: "user", Content: "I prefer concise answers."},
		},
	}
	if _, err := store.EnqueueIngestJob(ctx, "ing_1", "job_1", "", req); err != nil {
		t.Fatal(err)
	}

	job, ok, err := store.ClaimNextExtractionJob(ctx)
	if err != nil || !ok {
		t.Fatalf("expected first claim to succeed, got ok=%v err=%v", ok, err)
	}
	if job.JobID != "job_1" {
		t.Fatalf("expected claimed job_1, got %s", job.JobID)
	}

	time.Sleep(10 * time.Millisecond)

	reclaimed, ok, err := store.ClaimNextExtractionJob(ctx)
	if err != nil || !ok {
		t.Fatalf("expected stale job reclaim to succeed, got ok=%v err=%v", ok, err)
	}
	if reclaimed.JobID != "job_1" {
		t.Fatalf("expected reclaimed job_1, got %s", reclaimed.JobID)
	}
}

func TestClaimNextExtractionJobSerializesSameSubject(t *testing.T) {
	t.Setenv("LANG", "C")
	t.Setenv("LC_ALL", "C")
	ctx := context.Background()
	root := t.TempDir()
	port := randomPort(607)
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

	baseStore, err := New(ctx, dsn(port))
	if err != nil {
		t.Fatal(err)
	}
	defer baseStore.Close()
	if err := baseStore.ApplyMigrations(ctx); err != nil {
		t.Fatal(err)
	}
	store := &Store{pool: baseStore.pool, jobLease: 30 * time.Second}

	reqA := memory.IngestRequest{
		TenantID: "t1", SubjectID: "alice", SourceType: "conversation",
		Messages: []memory.Message{{Role: "user", Content: "Alice turn 1"}},
	}
	reqB := memory.IngestRequest{
		TenantID: "t1", SubjectID: "bob", SourceType: "conversation",
		Messages: []memory.Message{{Role: "user", Content: "Bob turn 1"}},
	}
	for _, item := range []struct {
		ingest, job, idem string
		req               memory.IngestRequest
	}{
		{"ing_a1", "job_a1", "idem_a1", reqA},
		{"ing_a2", "job_a2", "idem_a2", reqA},
		{"ing_b1", "job_b1", "idem_b1", reqB},
	} {
		if _, err := store.EnqueueIngestJob(ctx, item.ingest, item.job, item.idem, item.req); err != nil {
			t.Fatal(err)
		}
		// Ensure created_at ordering is strict across inserts.
		time.Sleep(2 * time.Millisecond)
	}

	first, ok, err := store.ClaimNextExtractionJob(ctx)
	if err != nil || !ok {
		t.Fatalf("first claim: ok=%v err=%v", ok, err)
	}
	if first.JobID != "job_a1" {
		t.Fatalf("expected job_a1 first, got %s", first.JobID)
	}

	second, ok, err := store.ClaimNextExtractionJob(ctx)
	if err != nil || !ok {
		t.Fatalf("second claim should take other subject: ok=%v err=%v", ok, err)
	}
	if second.JobID != "job_b1" {
		t.Fatalf("expected job_b1 while alice job_a1 is live, got %s", second.JobID)
	}

	third, ok, err := store.ClaimNextExtractionJob(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("expected no claim while alice earlier job is live, got %s", third.JobID)
	}

	if err := store.CompleteExtractionJob(ctx, first.JobID, first.IngestID); err != nil {
		t.Fatal(err)
	}
	next, ok, err := store.ClaimNextExtractionJob(ctx)
	if err != nil || !ok {
		t.Fatalf("after complete: ok=%v err=%v", ok, err)
	}
	if next.JobID != "job_a2" {
		t.Fatalf("expected job_a2 after job_a1 completed, got %s", next.JobID)
	}
}

func TestClaimNextExtractionJobConcurrentSameSubject(t *testing.T) {
	t.Setenv("LANG", "C")
	t.Setenv("LC_ALL", "C")
	ctx := context.Background()
	root := t.TempDir()
	port := randomPort(608)
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

	baseStore, err := New(ctx, dsn(port))
	if err != nil {
		t.Fatal(err)
	}
	defer baseStore.Close()
	if err := baseStore.ApplyMigrations(ctx); err != nil {
		t.Fatal(err)
	}
	store := &Store{pool: baseStore.pool, jobLease: 30 * time.Second}

	req := memory.IngestRequest{
		TenantID: "t1", SubjectID: "same", SourceType: "conversation",
		Messages: []memory.Message{{Role: "user", Content: "turn"}},
	}
	for i := 1; i <= 4; i++ {
		n := strconv.Itoa(i)
		if _, err := store.EnqueueIngestJob(ctx, "ing_"+n, "job_"+n, "idem_"+n, req); err != nil {
			t.Fatal(err)
		}
		time.Sleep(2 * time.Millisecond)
	}

	type claimResult struct {
		id string
		ok bool
	}
	ch := make(chan claimResult, 4)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			job, ok, err := store.ClaimNextExtractionJob(ctx)
			if err != nil {
				t.Errorf("claim error: %v", err)
				ch <- claimResult{}
				return
			}
			if ok {
				ch <- claimResult{id: job.JobID, ok: true}
				return
			}
			ch <- claimResult{}
		}()
	}
	wg.Wait()
	close(ch)

	claimed := make([]string, 0, 4)
	for r := range ch {
		if r.ok {
			claimed = append(claimed, r.id)
		}
	}
	if len(claimed) != 1 {
		t.Fatalf("expected exactly one same-subject claim under concurrency=4, got %v", claimed)
	}
	if claimed[0] != "job_1" {
		t.Fatalf("expected earliest job_1, got %s", claimed[0])
	}
}

func TestMemoryRelationsUpsertAndList(t *testing.T) {
	t.Setenv("LANG", "C")
	t.Setenv("LC_ALL", "C")
	ctx := context.Background()
	port := randomPort(3)
	root := t.TempDir()
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
	defer func() { _ = postgres.Stop() }()

	store, err := New(ctx, dsn(port))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.ApplyMigrations(ctx); err != nil {
		t.Fatal(err)
	}

	rel := memory.MemoryRelation{
		TenantID:  "t1",
		SubjectID: "u1",
		SrcEntity: "Jordan",
		Relation:  memory.PredicateOrigin,
		DstEntity: "Portugal",
		MemoryID:  "mem_origin",
	}
	if err := store.UpsertMemoryRelation(ctx, rel); err != nil {
		t.Fatal(err)
	}
	got, err := store.ListRelationsFrom(ctx, "t1", "u1", "jordan", memory.PredicateOrigin, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].DstEntity != "portugal" || got[0].MemoryID != "mem_origin" {
		t.Fatalf("got %+v", got)
	}
}
