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
	{
		version: 4,
		name:    "harden_job_queue",
		sql: `
ALTER TABLE raw_ingests
ADD COLUMN IF NOT EXISTS idempotency_key TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS raw_ingests_idempotency_key_unique
ON raw_ingests (idempotency_key) WHERE idempotency_key IS NOT NULL;

ALTER TABLE extraction_jobs
ADD COLUMN IF NOT EXISTS max_attempts INTEGER NOT NULL DEFAULT 3;

CREATE TABLE IF NOT EXISTS dead_letter_jobs (
    dead_letter_id TEXT PRIMARY KEY,
    job_id TEXT NOT NULL,
    ingest_id TEXT NOT NULL,
    failure_reason TEXT NOT NULL,
    attempts INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    dead_lettered_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS dead_letter_jobs_lookup
ON dead_letter_jobs (job_id, ingest_id);
`,
	},
	{
		version: 5,
		name:    "add_trigram_search_index",
		sql: `
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS memory_records_content_trgm
ON memory_records USING gin(content gin_trgm_ops);
CREATE INDEX IF NOT EXISTS memory_records_source_text_trgm
ON memory_records USING gin(source_text gin_trgm_ops);
`,
	},
	{
		version: 6,
		name:    "add_correction_lineage",
		sql: `
ALTER TABLE memory_records
ADD COLUMN IF NOT EXISTS corrected_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS correction_history (
    history_id TEXT PRIMARY KEY,
    memory_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    previous_content TEXT NOT NULL,
    corrected_content TEXT NOT NULL,
    source_text TEXT NOT NULL,
    corrected_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS correction_history_memory_lookup
ON correction_history (memory_id, corrected_at DESC);
CREATE INDEX IF NOT EXISTS correction_history_subject_lookup
ON correction_history (tenant_id, subject_id, corrected_at DESC);
`,
	},
	{
		version: 7,
		name:    "add_vertical_pack_fields",
		sql: `
ALTER TABLE memory_records
ADD COLUMN IF NOT EXISTS vertical TEXT NOT NULL DEFAULT 'core',
ADD COLUMN IF NOT EXISTS primitive TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS label TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS scope TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
ADD COLUMN IF NOT EXISTS lifecycle_state TEXT NOT NULL DEFAULT 'active';

CREATE INDEX IF NOT EXISTS memory_records_vertical_lookup
ON memory_records (tenant_id, subject_id, vertical, lifecycle_state, status);
`,
	},
	{
		version: 8,
		name:    "add_memory_embeddings",
		sql: `
ALTER TABLE memory_records
ADD COLUMN IF NOT EXISTS embedding REAL[];

CREATE TABLE IF NOT EXISTS memory_embeddings (
    memory_id TEXT PRIMARY KEY REFERENCES memory_records(memory_id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    embedding REAL[] NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS memory_embeddings_subject_lookup
ON memory_embeddings (tenant_id, subject_id);
`,
	},
	{
		version: 9,
		name:    "enable_pgvector_optional",
		sql: `
DO $$
BEGIN
    CREATE EXTENSION IF NOT EXISTS vector;
EXCEPTION
    WHEN OTHERS THEN
        NULL;
END $$;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector') THEN
        ALTER TABLE memory_embeddings
        ADD COLUMN IF NOT EXISTS embedding_vec vector(128);

        UPDATE memory_embeddings
        SET embedding_vec = embedding::vector(128)
        WHERE embedding_vec IS NULL AND embedding IS NOT NULL;

        CREATE INDEX IF NOT EXISTS memory_embeddings_vec_hnsw
        ON memory_embeddings USING hnsw (embedding_vec vector_cosine_ops);
    END IF;
EXCEPTION
    WHEN OTHERS THEN
        NULL;
END $$;
`,
	},
	{
		version: 10,
		name:    "add_observed_at",
		sql: `
ALTER TABLE memory_records
ADD COLUMN IF NOT EXISTS observed_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS memory_records_observed_at_lookup
ON memory_records (tenant_id, subject_id, observed_at DESC NULLS LAST);
`,
	},
	{
		version: 11,
		name:    "add_supersession_lineage",
		sql: `
ALTER TABLE memory_records
ADD COLUMN IF NOT EXISTS supersedes_id TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS superseded_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS memory_records_supersedes_lookup
ON memory_records (tenant_id, subject_id, supersedes_id)
WHERE supersedes_id <> '';

CREATE INDEX IF NOT EXISTS memory_records_superseded_at_lookup
ON memory_records (tenant_id, subject_id, superseded_at DESC NULLS LAST);
`,
	},
	{
		version: 12,
		name:    "add_memory_entity_hub",
		sql: `
CREATE TABLE IF NOT EXISTS memory_entity_links (
    tenant_id TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    entity_key TEXT NOT NULL,
    linked_memory_ids TEXT[] NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, subject_id, entity_key)
);
CREATE INDEX IF NOT EXISTS memory_entity_links_subject_lookup
ON memory_entity_links (tenant_id, subject_id);
`,
	},
	{
		version: 13,
		name:    "add_memory_atoms_index",
		sql: `
CREATE TABLE IF NOT EXISTS memory_atoms (
    tenant_id TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    predicate TEXT NOT NULL,
    value_norm TEXT NOT NULL,
    memory_id TEXT NOT NULL,
    observed_at TIMESTAMPTZ,
    valid_to TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, subject_id, predicate, value_norm, memory_id)
);
CREATE INDEX IF NOT EXISTS memory_atoms_predicate_scan
ON memory_atoms (tenant_id, subject_id, predicate, value_norm);
CREATE INDEX IF NOT EXISTS memory_atoms_memory_lookup
ON memory_atoms (memory_id);
`,
	},
	{
		// Column + trigger only. GIN index is created best-effort outside the
		// migration transaction (see EnsureContentFTSIndex) — CREATE INDEX on
		// large staging tables exceeded the 10s boot timeout and crash-looped
		// the worker.
		version: 14,
		name:    "add_content_fts",
		sql: `
ALTER TABLE memory_records
ADD COLUMN IF NOT EXISTS content_tsv tsvector;

CREATE OR REPLACE FUNCTION memory_records_content_tsv_trigger() RETURNS trigger AS $$
BEGIN
  NEW.content_tsv := to_tsvector('english', coalesce(NEW.content, ''));
  RETURN NEW;
END
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS memory_records_content_tsv_update ON memory_records;
CREATE TRIGGER memory_records_content_tsv_update
BEFORE INSERT OR UPDATE OF content ON memory_records
FOR EACH ROW EXECUTE PROCEDURE memory_records_content_tsv_trigger();
`,
	},
	{
		version: 15,
		name:    "atoms_bitemporal_fields",
		sql: `
ALTER TABLE memory_atoms ADD COLUMN IF NOT EXISTS valid_from TIMESTAMPTZ;
ALTER TABLE memory_atoms ADD COLUMN IF NOT EXISTS recorded_at TIMESTAMPTZ;
ALTER TABLE memory_atoms ADD COLUMN IF NOT EXISTS retired_at TIMESTAMPTZ;
UPDATE memory_atoms SET recorded_at = COALESCE(recorded_at, updated_at) WHERE recorded_at IS NULL;
UPDATE memory_atoms SET valid_from = COALESCE(valid_from, observed_at, updated_at) WHERE valid_from IS NULL;
CREATE INDEX IF NOT EXISTS memory_atoms_current_scan
ON memory_atoms (tenant_id, subject_id, predicate)
WHERE retired_at IS NULL;
`,
	},
	{
		version: 16,
		name:    "memory_events_and_evidence_shadow",
		sql: `
CREATE TABLE IF NOT EXISTS memory_evidence (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    namespace TEXT NOT NULL DEFAULT 'default',
    subject_id TEXT,
    source_type TEXT NOT NULL,
    source_ref TEXT,
    session_id TEXT,
    content TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    occurred_at TIMESTAMPTZ,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    metadata JSONB NOT NULL DEFAULT '{}',
    suppression_status TEXT NOT NULL DEFAULT 'active',
    memory_id TEXT
);
CREATE INDEX IF NOT EXISTS memory_evidence_subject_lookup
ON memory_evidence (tenant_id, subject_id, occurred_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS memory_evidence_hash_uniq
ON memory_evidence (tenant_id, content_hash);

CREATE TABLE IF NOT EXISTS memory_events (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    namespace TEXT NOT NULL DEFAULT 'default',
    subject_id TEXT,
    event_type TEXT NOT NULL,
    title TEXT,
    description TEXT,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    time_precision TEXT,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0.5,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    pack_id TEXT,
    evidence_id TEXT,
    memory_id TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'
);
CREATE INDEX IF NOT EXISTS memory_events_subject_time
ON memory_events (tenant_id, subject_id, starts_at DESC);

CREATE TABLE IF NOT EXISTS memory_event_participants (
    event_id TEXT NOT NULL REFERENCES memory_events(id) ON DELETE CASCADE,
    entity_key TEXT NOT NULL,
    participant_role TEXT NOT NULL DEFAULT 'participant',
    metadata JSONB NOT NULL DEFAULT '{}',
    PRIMARY KEY (event_id, entity_key, participant_role)
);

CREATE TABLE IF NOT EXISTS memory_current_state (
    tenant_id TEXT NOT NULL,
    namespace TEXT NOT NULL DEFAULT 'default',
    subject_id TEXT NOT NULL,
    predicate TEXT NOT NULL,
    winning_memory_id TEXT NOT NULL,
    resolved_value TEXT NOT NULL,
    resolution_policy TEXT NOT NULL,
    resolved_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, namespace, subject_id, predicate)
);
`,
	},
	{
		version: 17,
		name:    "evidence_plane_v2_subject_safe_dedupe",
		sql: `
ALTER TABLE memory_evidence ADD COLUMN IF NOT EXISTS actor_role TEXT;
DROP INDEX IF EXISTS memory_evidence_hash_uniq;
CREATE UNIQUE INDEX IF NOT EXISTS memory_evidence_subject_hash_uniq
ON memory_evidence (tenant_id, namespace, subject_id, content_hash, source_ref);
CREATE INDEX IF NOT EXISTS memory_evidence_content_lookup
ON memory_evidence (tenant_id, subject_id, recorded_at DESC);
`,
	},
	{
		// Additive column only — do NOT drop embedding_vec (vector(128)+HNSW)
		// during boot: DROP/rewrite timed out on staging under the 120s
		// migration lock. Hosted ANN uses embedding_vec_768; hash/128 keeps
		// the legacy column. Backfill + HNSW: EnsureEmbeddingVecIndex.
		version: 18,
		name:    "pgvector_embedding_vec_768",
		sql: `
-- Hosted pin: bge-base-en-v1.5 (768-d) via embedding_vec_768.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector') THEN
        ALTER TABLE memory_embeddings
            ADD COLUMN IF NOT EXISTS embedding_vec_768 vector(768);
    END IF;
EXCEPTION
    WHEN OTHERS THEN
        NULL;
END $$;
`,
	},

	{
		// Staging may have recorded mig 18 via EXCEPTION no-op or a DROP rewrite.
		// Always ensure the additive hosted ANN column exists (IF NOT EXISTS).
		version: 19,
		name:    "pgvector_embedding_vec_768_additive",
		sql: `
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector') THEN
        ALTER TABLE memory_embeddings
            ADD COLUMN IF NOT EXISTS embedding_vec_768 vector(768);
    END IF;
EXCEPTION
    WHEN OTHERS THEN
        NULL;
END $$;
`,
	},

	{
		version: 20,
		name:    "add_memory_relations",
		sql: `
CREATE TABLE IF NOT EXISTS memory_relations (
    tenant_id TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    src_entity TEXT NOT NULL,
    relation TEXT NOT NULL,
    dst_entity TEXT NOT NULL,
    memory_id TEXT NOT NULL,
    observed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, subject_id, src_entity, relation, dst_entity, memory_id)
);
CREATE INDEX IF NOT EXISTS memory_relations_src
ON memory_relations (tenant_id, subject_id, src_entity, relation);
CREATE INDEX IF NOT EXISTS memory_relations_dst
ON memory_relations (tenant_id, subject_id, dst_entity);
`,
	},

	{
		// Lease ownership fencing: a claimed job is bound to a per-claim token
		// (lease_owner) so a worker whose lease was reclaimed cannot commit or
		// fail the job under a new owner. last_heartbeat_at lets long work renew
		// its lease instead of being reclaimed mid-extraction.
		version: 21,
		name:    "fenced_job_leases",
		sql: `
ALTER TABLE extraction_jobs
ADD COLUMN IF NOT EXISTS lease_owner TEXT;
ALTER TABLE extraction_jobs
ADD COLUMN IF NOT EXISTS last_heartbeat_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS extraction_jobs_lease_owner_lookup
ON extraction_jobs (lease_owner) WHERE lease_owner IS NOT NULL;
`,
	},
	{
		version: 22,
		name:    "canonical_entities_and_relation_ids",
		sql: `
CREATE TABLE IF NOT EXISTS memory_entities (
    tenant_id TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    canonical_label TEXT NOT NULL,
    aliases TEXT[] NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, subject_id, entity_id)
);
CREATE INDEX IF NOT EXISTS memory_entities_label
ON memory_entities (tenant_id, subject_id, canonical_label);
CREATE INDEX IF NOT EXISTS memory_entities_aliases
ON memory_entities USING gin (aliases);

ALTER TABLE memory_relations
ADD COLUMN IF NOT EXISTS src_entity_id TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS dst_entity_id TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS valid_from TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS valid_to TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS evidence_span TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS memory_relations_src_id
ON memory_relations (tenant_id, subject_id, src_entity_id, relation)
WHERE src_entity_id <> '';
`,
	},
}

// EnsureContentFTSIndex builds the GIN index outside the migration txn.
// Safe to call repeatedly; ignores timeout/lock errors so boot is never blocked.
func (s *Store) EnsureContentFTSIndex(ctx context.Context) {
	// Drop leftover invalid indexes from prior timed-out CREATE INDEX attempts.
	_, _ = s.pool.Exec(ctx, `
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_class c
    JOIN pg_index i ON i.indexrelid = c.oid
    WHERE c.relname = 'memory_records_content_tsv_gin' AND NOT i.indisvalid
  ) THEN
    EXECUTE 'DROP INDEX IF EXISTS memory_records_content_tsv_gin';
  END IF;
END $$;
`)
	// CONCURRENTLY avoids long write locks on memory_records during build.
	_, _ = s.pool.Exec(ctx, `
CREATE INDEX CONCURRENTLY IF NOT EXISTS memory_records_content_tsv_gin
ON memory_records USING GIN (content_tsv);
`)
}

// EnsureEmbeddingVecIndex backfills hosted 768-d vectors into
// embedding_vec_768 and builds HNSW outside the boot migration txn.
//
// CREATE INDEX CONCURRENTLY must not inherit short boot/deploy timeouts —
// cancelling mid-build leaves an INVALID index that ANN cannot use.
func (s *Store) EnsureEmbeddingVecIndex(ctx context.Context) {
	var hasVector bool
	if err := s.pool.QueryRow(ctx, `
SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector')
`).Scan(&hasVector); err != nil || !hasVector {
		return
	}
	var hasCol bool
	if err := s.pool.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1 FROM information_schema.columns
  WHERE table_name = 'memory_embeddings' AND column_name = 'embedding_vec_768'
)
`).Scan(&hasCol); err != nil || !hasCol {
		return
	}

	// Batched backfill — full-table UPDATE on staging OOMs/starves the
	// basic Postgres plan for many minutes and blocks concurrent DDL.
	// Only copies already-hosted 768-d float[] → vector(768). Hash/128
	// rows need a separate hosted re-embed of memory_records.content.
	for {
		tag, err := s.pool.Exec(ctx, `
UPDATE memory_embeddings
SET embedding_vec_768 = embedding::vector(768)
WHERE memory_id IN (
  SELECT memory_id
  FROM memory_embeddings
  WHERE embedding_vec_768 IS NULL
    AND embedding IS NOT NULL
    AND cardinality(embedding) = 768
  LIMIT 500
)
`)
		if err != nil || tag.RowsAffected() == 0 {
			break
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
	}

	var valid bool
	_ = s.pool.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1 FROM pg_class c
  JOIN pg_index i ON i.indexrelid = c.oid
  WHERE c.relname = 'memory_embeddings_vec_768_hnsw' AND i.indisvalid
)
`).Scan(&valid)
	if valid {
		return
	}

	var building bool
	_ = s.pool.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1 FROM pg_stat_activity
  WHERE pid <> pg_backend_pid()
    AND query ILIKE '%CREATE INDEX%CONCURRENTLY%memory_embeddings_vec_768_hnsw%'
)
`).Scan(&building)
	if building {
		return
	}

	// Drop INVALID leftovers from cancelled concurrent builds, then rebuild
	// with an uncancellable context so deploy SIGTERM/timeouts cannot flap.
	_, _ = s.pool.Exec(ctx, `
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_class c
    JOIN pg_index i ON i.indexrelid = c.oid
    WHERE c.relname = 'memory_embeddings_vec_768_hnsw' AND NOT i.indisvalid
  ) THEN
    EXECUTE 'DROP INDEX IF EXISTS memory_embeddings_vec_768_hnsw';
  END IF;
END $$;
`)
	ddlCtx := context.WithoutCancel(ctx)
	_, _ = s.pool.Exec(ddlCtx, `
CREATE INDEX CONCURRENTLY IF NOT EXISTS memory_embeddings_vec_768_hnsw
ON memory_embeddings USING hnsw (embedding_vec_768 vector_cosine_ops);
`)
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
