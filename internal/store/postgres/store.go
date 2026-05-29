package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"brainy/internal/memory"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool     *pgxpool.Pool
	jobLease time.Duration
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	return &Store{pool: pool, jobLease: 30 * time.Second}, nil
}

func NewWithPool(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, jobLease: 30 * time.Second}
}

func NewWithPoolAndLease(pool *pgxpool.Pool, jobLease time.Duration) *Store {
	return &Store{pool: pool, jobLease: jobLease}
}

func (s *Store) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func decodeExplain(data []byte) map[string]any {
	if len(data) == 0 {
		return nil
	}

	var explain map[string]any
	_ = json.Unmarshal(data, &explain)
	return explain
}

func (s *Store) UpsertMemory(ctx context.Context, record memory.MemoryRecord) (memory.StoreUpsertResult, error) {
	explain, err := json.Marshal(record.Explain)
	if err != nil {
		return memory.StoreUpsertResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return memory.StoreUpsertResult{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	insertRow := tx.QueryRow(ctx, `
INSERT INTO memory_records (
    memory_id, tenant_id, subject_id, kind, content, source_text, source_type, dedupe_key,
    status, confidence, extraction_version, explain, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $13, $14
)
ON CONFLICT (tenant_id, subject_id, dedupe_key) DO NOTHING
RETURNING memory_id, tenant_id, subject_id, kind, content, source_text, source_type, dedupe_key, status, confidence, extraction_version, explain, created_at, updated_at
`, record.MemoryID, record.TenantID, record.SubjectID, record.Kind, record.Content, record.SourceText, record.SourceType, record.DedupeKey, record.Status, record.Confidence, record.ExtractionVersion, explain, record.CreatedAt, record.UpdatedAt)

	var inserted memory.MemoryRecord
	var insertedExplain []byte
	err = insertRow.Scan(
		&inserted.MemoryID,
		&inserted.TenantID,
		&inserted.SubjectID,
		&inserted.Kind,
		&inserted.Content,
		&inserted.SourceText,
		&inserted.SourceType,
		&inserted.DedupeKey,
		&inserted.Status,
		&inserted.Confidence,
		&inserted.ExtractionVersion,
		&insertedExplain,
		&inserted.CreatedAt,
		&inserted.UpdatedAt,
	)
	if err == nil {
		inserted.Explain = decodeExplain(insertedExplain)
		if err := tx.Commit(ctx); err != nil {
			return memory.StoreUpsertResult{}, err
		}
		return memory.StoreUpsertResult{Record: inserted, State: "created"}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return memory.StoreUpsertResult{}, err
	}

	var existing memory.MemoryRecord
	var rawExplain []byte
	row := tx.QueryRow(ctx, `
SELECT memory_id, tenant_id, subject_id, kind, content, source_text, source_type, dedupe_key, status, confidence, extraction_version, explain, created_at, updated_at
FROM memory_records
WHERE tenant_id = $1 AND subject_id = $2 AND dedupe_key = $3
FOR UPDATE
`, record.TenantID, record.SubjectID, record.DedupeKey)
	err = row.Scan(
		&existing.MemoryID,
		&existing.TenantID,
		&existing.SubjectID,
		&existing.Kind,
		&existing.Content,
		&existing.SourceText,
		&existing.SourceType,
		&existing.DedupeKey,
		&existing.Status,
		&existing.Confidence,
		&existing.ExtractionVersion,
		&rawExplain,
		&existing.CreatedAt,
		&existing.UpdatedAt,
	)
	if err != nil {
		return memory.StoreUpsertResult{}, err
	}
	existing.Explain = decodeExplain(rawExplain)

	if existing.Content == record.Content && existing.Status == memory.StatusActive {
		if err := tx.Commit(ctx); err != nil {
			return memory.StoreUpsertResult{}, err
		}
		return memory.StoreUpsertResult{Record: existing, State: "deduped"}, nil
	}

	record.MemoryID = existing.MemoryID
	record.CreatedAt = existing.CreatedAt
	_, err = tx.Exec(ctx, `
UPDATE memory_records
SET content = $4,
    source_text = $5,
    source_type = $6,
    status = $7,
    confidence = $8,
    extraction_version = $9,
    explain = $10,
    updated_at = $11
WHERE memory_id = $1 AND tenant_id = $2 AND subject_id = $3
`, record.MemoryID, record.TenantID, record.SubjectID, record.Content, record.SourceText, record.SourceType, record.Status, record.Confidence, record.ExtractionVersion, explain, record.UpdatedAt)
	if err != nil {
		return memory.StoreUpsertResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return memory.StoreUpsertResult{}, err
	}
	return memory.StoreUpsertResult{Record: record, State: "updated"}, nil
}

func (s *Store) ListActiveMemories(ctx context.Context, tenantID, subjectID string) ([]memory.MemoryRecord, error) {
	rows, err := s.pool.Query(ctx, `
SELECT memory_id, tenant_id, subject_id, kind, content, source_text, source_type, dedupe_key, status, confidence, extraction_version, explain, created_at, updated_at
FROM memory_records
WHERE tenant_id = $1 AND subject_id = $2 AND status = $3
ORDER BY updated_at DESC, memory_id ASC
`, tenantID, subjectID, memory.StatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []memory.MemoryRecord
	for rows.Next() {
		var record memory.MemoryRecord
		var rawExplain []byte
		if err := rows.Scan(
			&record.MemoryID,
			&record.TenantID,
			&record.SubjectID,
			&record.Kind,
			&record.Content,
			&record.SourceText,
			&record.SourceType,
			&record.DedupeKey,
			&record.Status,
			&record.Confidence,
			&record.ExtractionVersion,
			&rawExplain,
			&record.CreatedAt,
			&record.UpdatedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(rawExplain, &record.Explain)
		out = append(out, record)
	}
	return out, rows.Err()
}

func (s *Store) SearchActiveMemories(ctx context.Context, tenantID, subjectID string, patterns []string, limit int) ([]memory.MemoryRecord, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, `
SELECT memory_id, tenant_id, subject_id, kind, content, source_text, source_type, dedupe_key, status, confidence, extraction_version, explain, created_at, updated_at
FROM memory_records
WHERE tenant_id = $1 AND subject_id = $2 AND status = $3
  AND content ILIKE ANY($4)
ORDER BY updated_at DESC
LIMIT $5
`, tenantID, subjectID, memory.StatusActive, patterns, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []memory.MemoryRecord
	for rows.Next() {
		var record memory.MemoryRecord
		var rawExplain []byte
		if err := rows.Scan(
			&record.MemoryID,
			&record.TenantID,
			&record.SubjectID,
			&record.Kind,
			&record.Content,
			&record.SourceText,
			&record.SourceType,
			&record.DedupeKey,
			&record.Status,
			&record.Confidence,
			&record.ExtractionVersion,
			&rawExplain,
			&record.CreatedAt,
			&record.UpdatedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(rawExplain, &record.Explain)
		out = append(out, record)
	}
	return out, rows.Err()
}

func (s *Store) SuppressMemory(ctx context.Context, tenantID, subjectID, memoryID string) error {
	commandTag, err := s.pool.Exec(ctx, `
UPDATE memory_records
SET status = $4, updated_at = $5
WHERE tenant_id = $1 AND subject_id = $2 AND memory_id = $3
`, tenantID, subjectID, memoryID, memory.StatusSuppressed, time.Now().UTC())
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return memory.ErrMemoryNotFound
	}
	return nil
}

func (s *Store) CorrectMemory(ctx context.Context, tenantID, subjectID, memoryID, content, sourceText string) (memory.MemoryRecord, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return memory.MemoryRecord{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var record memory.MemoryRecord
	var rawExplain []byte
	row := tx.QueryRow(ctx, `
SELECT memory_id, tenant_id, subject_id, kind, content, source_text, source_type, dedupe_key, status, confidence, extraction_version, explain, created_at, updated_at
FROM memory_records
WHERE tenant_id = $1 AND subject_id = $2 AND memory_id = $3
FOR UPDATE
`, tenantID, subjectID, memoryID)
	if err := row.Scan(
		&record.MemoryID,
		&record.TenantID,
		&record.SubjectID,
		&record.Kind,
		&record.Content,
		&record.SourceText,
		&record.SourceType,
		&record.DedupeKey,
		&record.Status,
		&record.Confidence,
		&record.ExtractionVersion,
		&rawExplain,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return memory.MemoryRecord{}, err
	}
	record.Explain = decodeExplain(rawExplain)

	record.Content = content
	record.SourceText = sourceText
	record.DedupeKey = memory.DedupeKey(tenantID, subjectID, record.Kind, content)
	record.Status = memory.StatusActive
	record.UpdatedAt = time.Now().UTC()

	var duplicateMemoryID string
	err = tx.QueryRow(ctx, `
SELECT memory_id
FROM memory_records
WHERE tenant_id = $1 AND subject_id = $2 AND dedupe_key = $3 AND memory_id <> $4
LIMIT 1
`, tenantID, subjectID, record.DedupeKey, memoryID).Scan(&duplicateMemoryID)
	if err == nil {
		return memory.MemoryRecord{}, memory.ErrMemoryConflict
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return memory.MemoryRecord{}, err
	}

	explain, err := json.Marshal(record.Explain)
	if err != nil {
		return memory.MemoryRecord{}, err
	}

	if _, err := tx.Exec(ctx, `
UPDATE memory_records
SET content = $4,
    source_text = $5,
    dedupe_key = $6,
    status = $7,
    explain = $8,
    updated_at = $9
WHERE memory_id = $1 AND tenant_id = $2 AND subject_id = $3
`, record.MemoryID, record.TenantID, record.SubjectID, record.Content, record.SourceText, record.DedupeKey, record.Status, explain, record.UpdatedAt); err != nil {
		return memory.MemoryRecord{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return memory.MemoryRecord{}, err
	}
	return record, nil
}

func (s *Store) EnqueueIngestJob(ctx context.Context, ingestID, jobID, idempotencyKey string, req memory.IngestRequest) (memory.EnqueueResult, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return memory.EnqueueResult{}, err
	}

	now := time.Now().UTC()
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return memory.EnqueueResult{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if idempotencyKey != "" {
		var existingIngestID, existingJobID string
		err := tx.QueryRow(ctx, `
SELECT ingest_id, (
    SELECT job_id FROM extraction_jobs WHERE ingest_id = raw_ingests.ingest_id LIMIT 1
)
FROM raw_ingests
WHERE idempotency_key = $1
`, idempotencyKey).Scan(&existingIngestID, &existingJobID)
		if err == nil {
			_ = tx.Rollback(ctx)
			return memory.EnqueueResult{
				IngestID:  existingIngestID,
				JobID:     existingJobID,
				Accepted:  true,
				Duplicate: true,
			}, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return memory.EnqueueResult{}, err
		}
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO raw_ingests (
    ingest_id, tenant_id, subject_id, source_type, payload, status, idempotency_key, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
`, ingestID, req.TenantID, req.SubjectID, req.SourceType, payload, memory.RawIngestStatusPending, idempotencyKey, now, now); err != nil {
		return memory.EnqueueResult{}, err
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO extraction_jobs (
    job_id, ingest_id, status, attempts, max_attempts, lease_expires_at, created_at, updated_at
) VALUES (
    $1, $2, $3, 0, 3, NULL, $4, $5
)
`, jobID, ingestID, memory.JobStatusPending, now, now); err != nil {
		return memory.EnqueueResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return memory.EnqueueResult{}, err
	}
	return memory.EnqueueResult{
		IngestID:  ingestID,
		JobID:     jobID,
		Accepted:  true,
		Duplicate: false,
	}, nil
}

func (s *Store) ClaimNextExtractionJob(ctx context.Context) (memory.ExtractionJob, bool, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return memory.ExtractionJob{}, false, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	now := time.Now().UTC()
	row := tx.QueryRow(ctx, `
SELECT j.job_id, j.ingest_id, j.attempts, j.max_attempts, j.created_at, i.payload
FROM extraction_jobs j
JOIN raw_ingests i ON i.ingest_id = j.ingest_id
WHERE (j.status = $1 OR (j.status = $2 AND j.lease_expires_at <= $3))
  AND j.attempts < j.max_attempts
ORDER BY j.created_at ASC
FOR UPDATE OF j SKIP LOCKED
LIMIT 1
`, memory.JobStatusPending, memory.JobStatusInProgress, now)

	var job memory.ExtractionJob
	var payload []byte
	if err := row.Scan(&job.JobID, &job.IngestID, &job.Attempts, &job.MaxAttempts, &job.CreatedAt, &payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return memory.ExtractionJob{}, false, nil
		}
		return memory.ExtractionJob{}, false, err
	}

	if err := json.Unmarshal(payload, &job.Request); err != nil {
		return memory.ExtractionJob{}, false, err
	}

	backoff := s.jobLease * time.Duration(1<<min(job.Attempts, 6))
	if _, err := tx.Exec(ctx, `
UPDATE extraction_jobs
SET status = $2, attempts = attempts + 1, updated_at = $3, lease_expires_at = $4
WHERE job_id = $1
`, job.JobID, memory.JobStatusInProgress, now, now.Add(backoff)); err != nil {
		return memory.ExtractionJob{}, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return memory.ExtractionJob{}, false, err
	}
	job.Attempts++
	return job, true, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *Store) CompleteExtractionJob(ctx context.Context, jobID, ingestID string) error {
	now := time.Now().UTC()
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, `
UPDATE extraction_jobs
SET status = $2, updated_at = $3, processed_at = $3, lease_expires_at = NULL
WHERE job_id = $1
`, jobID, memory.JobStatusCompleted, now); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
UPDATE raw_ingests
SET status = $2, updated_at = $3, processed_at = $3
WHERE ingest_id = $1
`, ingestID, memory.RawIngestStatusProcessed, now); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) FailExtractionJob(ctx context.Context, jobID, ingestID, reason string) error {
	now := time.Now().UTC()
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var attempts, maxAttempts int
	err = tx.QueryRow(ctx, `
SELECT attempts, max_attempts FROM extraction_jobs WHERE job_id = $1
`, jobID).Scan(&attempts, &maxAttempts)
	if err != nil {
		return err
	}

	if attempts >= maxAttempts {
		deadLetterID := fmt.Sprintf("dl_%d", now.UnixNano())
		if _, err := tx.Exec(ctx, `
INSERT INTO dead_letter_jobs (
    dead_letter_id, job_id, ingest_id, failure_reason, attempts, created_at, dead_lettered_at
) VALUES ($1, $2, $3, $4, $5, $6, $6)
`, deadLetterID, jobID, ingestID, reason, attempts, now); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
UPDATE extraction_jobs
SET status = $2, failure_reason = $3, updated_at = $4, lease_expires_at = NULL
WHERE job_id = $1
`, jobID, memory.JobStatusFailed, reason, now); err != nil {
			return err
		}
	} else {
		if _, err := tx.Exec(ctx, `
UPDATE extraction_jobs
SET status = $2, failure_reason = $3, updated_at = $4, lease_expires_at = NULL
WHERE job_id = $1
`, jobID, memory.JobStatusPending, reason, now); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, `
UPDATE raw_ingests
SET status = $2, updated_at = $3
WHERE ingest_id = $1
`, ingestID, memory.RawIngestStatusFailed, now); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
