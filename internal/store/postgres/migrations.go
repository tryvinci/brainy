package postgres

import (
	"context"
	"fmt"
	"time"
)

type migration struct {
	version int
	name    string
	sql     string
}

var migrations = []migration{
	{
		version: 1,
		name:    "create_memory_records",
		sql: `
CREATE TABLE IF NOT EXISTS memory_records (
    memory_id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    kind TEXT NOT NULL,
    content TEXT NOT NULL,
    source_text TEXT NOT NULL,
    source_type TEXT NOT NULL,
    dedupe_key TEXT NOT NULL,
    status TEXT NOT NULL,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
    extraction_version TEXT NOT NULL DEFAULT 'deterministic-v1',
    explain JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS memory_records_unique_dedupe
ON memory_records (tenant_id, subject_id, dedupe_key);
CREATE INDEX IF NOT EXISTS memory_records_lookup
ON memory_records (tenant_id, subject_id, status, updated_at DESC);
`,
	},
	{
		version: 2,
		name:    "create_async_ingest_tables",
		sql: `
CREATE TABLE IF NOT EXISTS raw_ingests (
    ingest_id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    source_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    processed_at TIMESTAMPTZ
);
CREATE TABLE IF NOT EXISTS extraction_jobs (
    job_id TEXT PRIMARY KEY,
    ingest_id TEXT NOT NULL REFERENCES raw_ingests(ingest_id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    failure_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    processed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS extraction_jobs_pending_lookup
ON extraction_jobs (status, created_at);
`,
	},
}

func (s *Store) ApplyMigrations(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL
);
`); err != nil {
		return err
	}

	for _, migration := range migrations {
		var applied bool
		if err := s.pool.QueryRow(ctx, `
SELECT EXISTS(
    SELECT 1
    FROM schema_migrations
    WHERE version = $1
)
`, migration.version).Scan(&applied); err != nil {
			return err
		}
		if applied {
			continue
		}

		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, migration.sql); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %d (%s): %w", migration.version, migration.name, err)
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO schema_migrations (version, name, applied_at)
VALUES ($1, $2, $3)
`, migration.version, migration.name, time.Now().UTC()); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %d (%s): %w", migration.version, migration.name, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}

	return nil
}
