package postgres

import (
	"context"
	"fmt"
	"time"
)

const migrationLockID int64 = 884211

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
    lease_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    processed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS extraction_jobs_pending_lookup
ON extraction_jobs (status, created_at);
`,
	},
	{
		version: 3,
		name:    "add_job_lease_expiry",
		sql: `
ALTER TABLE extraction_jobs
ADD COLUMN IF NOT EXISTS lease_expires_at TIMESTAMPTZ;
`,
	},
}

func (s *Store) ApplyMigrations(ctx context.Context) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, migrationLockID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
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
		if err := tx.QueryRow(ctx, `
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

		if _, err := tx.Exec(ctx, migration.sql); err != nil {
			return fmt.Errorf("apply migration %d (%s): %w", migration.version, migration.name, err)
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO schema_migrations (version, name, applied_at)
VALUES ($1, $2, $3)
`, migration.version, migration.name, time.Now().UTC()); err != nil {
			return fmt.Errorf("record migration %d (%s): %w", migration.version, migration.name, err)
		}
	}

	return tx.Commit(ctx)
}
