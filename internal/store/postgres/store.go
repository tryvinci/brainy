package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

func decodeMetadata(data []byte) map[string]any {
	if len(data) == 0 {
		return map[string]any{}
	}
	var metadata map[string]any
	if err := json.Unmarshal(data, &metadata); err != nil || metadata == nil {
		return map[string]any{}
	}
	return metadata
}

func metadataEqual(left, right map[string]any) bool {
	leftJSON, err := json.Marshal(left)
	if err != nil {
		return false
	}
	rightJSON, err := json.Marshal(right)
	if err != nil {
		return false
	}
	return string(leftJSON) == string(rightJSON)
}

const memoryRecordSelectCols = `
memory_id, tenant_id, subject_id, kind, content, source_text, source_type, dedupe_key,
status, confidence, extraction_version, explain, created_at, updated_at, corrected_at,
vertical, primitive, label, scope, metadata, lifecycle_state, observed_at,
supersedes_id, superseded_at`

func scanMemoryRow(row pgx.Row) (memory.MemoryRecord, error) {
	var record memory.MemoryRecord
	var rawExplain, rawMetadata []byte
	var correctedAt, observedAt, supersededAt *time.Time
	err := row.Scan(
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
		&correctedAt,
		&record.Vertical,
		&record.Primitive,
		&record.Label,
		&record.Scope,
		&rawMetadata,
		&record.LifecycleState,
		&observedAt,
		&record.SupersedesID,
		&supersededAt,
	)
	if err != nil {
		return memory.MemoryRecord{}, err
	}
	record.Explain = decodeExplain(rawExplain)
	record.Metadata = decodeMetadata(rawMetadata)
	record.CorrectedAt = correctedAt
	record.ObservedAt = observedAt
	record.SupersededAt = supersededAt
	if record.Vertical == "" {
		record.Vertical = memory.VerticalCore
	}
	if record.LifecycleState == "" {
		record.LifecycleState = memory.LifecycleActive
	}
	return record, nil
}

func (s *Store) UpsertMemory(ctx context.Context, record memory.MemoryRecord) (memory.StoreUpsertResult, error) {
	explain, err := json.Marshal(record.Explain)
	if err != nil {
		return memory.StoreUpsertResult{}, err
	}
	metadata, err := json.Marshal(record.Metadata)
	if err != nil {
		return memory.StoreUpsertResult{}, err
	}
	if record.Vertical == "" {
		record.Vertical = memory.VerticalCore
	}
	if record.LifecycleState == "" {
		record.LifecycleState = memory.LifecycleActive
	}
	if record.Metadata == nil {
		record.Metadata = map[string]any{}
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
    status, confidence, extraction_version, explain, created_at, updated_at,
    vertical, primitive, label, scope, metadata, lifecycle_state, observed_at,
    supersedes_id, superseded_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $13, $14,
    $15, $16, $17, $18, $19, $20, $21,
    $22, $23
)
ON CONFLICT (tenant_id, subject_id, dedupe_key) DO NOTHING
RETURNING `+memoryRecordSelectCols+`
`, record.MemoryID, record.TenantID, record.SubjectID, record.Kind, record.Content, record.SourceText, record.SourceType, record.DedupeKey, record.Status, record.Confidence, record.ExtractionVersion, explain, record.CreatedAt, record.UpdatedAt, record.Vertical, record.Primitive, record.Label, record.Scope, metadata, record.LifecycleState, record.ObservedAt, record.SupersedesID, record.SupersededAt)

	inserted, err := scanMemoryRow(insertRow)
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return memory.StoreUpsertResult{}, err
		}
		return memory.StoreUpsertResult{Record: inserted, State: "created"}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return memory.StoreUpsertResult{}, err
	}

	row := tx.QueryRow(ctx, `
SELECT `+memoryRecordSelectCols+`
FROM memory_records
WHERE tenant_id = $1 AND subject_id = $2 AND dedupe_key = $3
FOR UPDATE
`, record.TenantID, record.SubjectID, record.DedupeKey)
	existing, err := scanMemoryRow(row)
	if err != nil {
		return memory.StoreUpsertResult{}, err
	}

	if existing.Status == memory.StatusSuppressed {
		if err := tx.Commit(ctx); err != nil {
			return memory.StoreUpsertResult{}, err
		}
		return memory.StoreUpsertResult{Record: existing, State: "deduped"}, nil
	}

	if existing.Content == record.Content && existing.Status == memory.StatusActive {
		if metadataEqual(existing.Metadata, record.Metadata) &&
			existing.LifecycleState == record.LifecycleState &&
			existing.Label == record.Label &&
			existing.Scope == record.Scope {
			if err := tx.Commit(ctx); err != nil {
				return memory.StoreUpsertResult{}, err
			}
			return memory.StoreUpsertResult{Record: existing, State: "deduped"}, nil
		}
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
    updated_at = $11,
    vertical = $12,
    primitive = $13,
    label = $14,
    scope = $15,
    metadata = $16,
    lifecycle_state = $17,
    observed_at = $18,
    supersedes_id = $19,
    superseded_at = $20
WHERE memory_id = $1 AND tenant_id = $2 AND subject_id = $3
`, record.MemoryID, record.TenantID, record.SubjectID, record.Content, record.SourceText, record.SourceType, record.Status, record.Confidence, record.ExtractionVersion, explain, record.UpdatedAt, record.Vertical, record.Primitive, record.Label, record.Scope, metadata, record.LifecycleState, record.ObservedAt, record.SupersedesID, record.SupersededAt)
	if err != nil {
		return memory.StoreUpsertResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return memory.StoreUpsertResult{}, err
	}
	return memory.StoreUpsertResult{Record: record, State: "updated"}, nil
}

func (s *Store) ListActiveMemories(ctx context.Context, tenantID, subjectID string) ([]memory.MemoryRecord, error) {
	return s.ListMemories(ctx, tenantID, subjectID, false)
}

func (s *Store) ListMemories(ctx context.Context, tenantID, subjectID string, includeSuperseded bool) ([]memory.MemoryRecord, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if includeSuperseded {
		rows, err = s.pool.Query(ctx, `
SELECT `+memoryRecordSelectCols+`
FROM memory_records
WHERE tenant_id = $1 AND subject_id = $2 AND status = $3
  AND lifecycle_state NOT IN ($4, $5)
ORDER BY updated_at DESC, memory_id ASC
`, tenantID, subjectID, memory.StatusActive, memory.LifecycleArchived, memory.LifecycleSuppressed)
	} else {
		rows, err = s.pool.Query(ctx, `
SELECT `+memoryRecordSelectCols+`
FROM memory_records
WHERE tenant_id = $1 AND subject_id = $2 AND status = $3
  AND lifecycle_state NOT IN ($4, $5, $6)
ORDER BY updated_at DESC, memory_id ASC
`, tenantID, subjectID, memory.StatusActive, memory.LifecycleArchived, memory.LifecycleSuperseded, memory.LifecycleSuppressed)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []memory.MemoryRecord
	for rows.Next() {
		record, err := scanMemoryRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return out, rows.Err()
}

// ListMemoriesLimited is the hot-path subject corpus scan with an explicit cap
// (program Phase 1 / MEM-015). Prefer this over unbounded ListMemories.
func (s *Store) ListMemoriesLimited(ctx context.Context, tenantID, subjectID string, includeSuperseded bool, limit int) ([]memory.MemoryRecord, error) {
	if limit <= 0 {
		limit = 400
	}
	var (
		rows pgx.Rows
		err  error
	)
	if includeSuperseded {
		rows, err = s.pool.Query(ctx, `
SELECT `+memoryRecordSelectCols+`
FROM memory_records
WHERE tenant_id = $1 AND subject_id = $2 AND status = $3
  AND lifecycle_state NOT IN ($4, $5)
ORDER BY updated_at DESC, memory_id ASC
LIMIT $6
`, tenantID, subjectID, memory.StatusActive, memory.LifecycleArchived, memory.LifecycleSuppressed, limit)
	} else {
		rows, err = s.pool.Query(ctx, `
SELECT `+memoryRecordSelectCols+`
FROM memory_records
WHERE tenant_id = $1 AND subject_id = $2 AND status = $3
  AND lifecycle_state NOT IN ($4, $5, $6)
ORDER BY updated_at DESC, memory_id ASC
LIMIT $7
`, tenantID, subjectID, memory.StatusActive, memory.LifecycleArchived, memory.LifecycleSuperseded, memory.LifecycleSuppressed, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]memory.MemoryRecord, 0, limit)
	for rows.Next() {
		record, err := scanMemoryRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return out, rows.Err()
}

func (s *Store) SearchActiveMemories(ctx context.Context, tenantID, subjectID string, patterns []string, limit int) ([]memory.MemoryRecord, error) {
	return s.SearchMemories(ctx, tenantID, subjectID, patterns, limit, false)
}

func (s *Store) SearchMemories(ctx context.Context, tenantID, subjectID string, patterns []string, limit int, includeSuperseded bool) ([]memory.MemoryRecord, error) {
	if limit <= 0 {
		limit = 100
	}
	// Prefer FTS when a plain query string can be derived from patterns.
	query := ftsQueryFromPatterns(patterns)
	if query != "" {
		if results, err := s.searchMemoriesFTS(ctx, tenantID, subjectID, query, limit, includeSuperseded); err == nil && len(results) > 0 {
			return results, nil
		}
	}
	var (
		rows pgx.Rows
		err  error
	)
	if includeSuperseded {
		rows, err = s.pool.Query(ctx, `
SELECT `+memoryRecordSelectCols+`
FROM memory_records
WHERE tenant_id = $1 AND subject_id = $2 AND status = $3
  AND lifecycle_state NOT IN ($4, $5)
  AND content ILIKE ANY($6)
ORDER BY updated_at DESC
LIMIT $7
`, tenantID, subjectID, memory.StatusActive, memory.LifecycleArchived, memory.LifecycleSuppressed, patterns, limit)
	} else {
		rows, err = s.pool.Query(ctx, `
SELECT `+memoryRecordSelectCols+`
FROM memory_records
WHERE tenant_id = $1 AND subject_id = $2 AND status = $3
  AND lifecycle_state NOT IN ($4, $5, $6)
  AND content ILIKE ANY($7)
ORDER BY updated_at DESC
LIMIT $8
`, tenantID, subjectID, memory.StatusActive, memory.LifecycleArchived, memory.LifecycleSuperseded, memory.LifecycleSuppressed, patterns, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []memory.MemoryRecord
	for rows.Next() {
		record, err := scanMemoryRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return out, rows.Err()
}

func ftsQueryFromPatterns(patterns []string) string {
	terms := make([]string, 0, len(patterns))
	for _, p := range patterns {
		t := strings.Trim(p, "% ")
		t = strings.ReplaceAll(t, "'", "")
		if len(t) < 2 {
			continue
		}
		terms = append(terms, t)
	}
	if len(terms) == 0 {
		return ""
	}
	return strings.Join(terms, " ")
}

func (s *Store) searchMemoriesFTS(ctx context.Context, tenantID, subjectID, query string, limit int, includeSuperseded bool) ([]memory.MemoryRecord, error) {
	var rows pgx.Rows
	var err error
	if includeSuperseded {
		rows, err = s.pool.Query(ctx, `
SELECT `+memoryRecordSelectCols+`
FROM memory_records
WHERE tenant_id = $1 AND subject_id = $2 AND status = $3
  AND lifecycle_state NOT IN ($4, $5)
  AND content_tsv @@ plainto_tsquery('english', $6)
ORDER BY ts_rank_cd(content_tsv, plainto_tsquery('english', $6)) DESC, updated_at DESC
LIMIT $7
`, tenantID, subjectID, memory.StatusActive, memory.LifecycleArchived, memory.LifecycleSuppressed, query, limit)
	} else {
		rows, err = s.pool.Query(ctx, `
SELECT `+memoryRecordSelectCols+`
FROM memory_records
WHERE tenant_id = $1 AND subject_id = $2 AND status = $3
  AND lifecycle_state NOT IN ($4, $5, $6)
  AND content_tsv @@ plainto_tsquery('english', $7)
ORDER BY ts_rank_cd(content_tsv, plainto_tsquery('english', $7)) DESC, updated_at DESC
LIMIT $8
`, tenantID, subjectID, memory.StatusActive, memory.LifecycleArchived, memory.LifecycleSuperseded, memory.LifecycleSuppressed, query, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []memory.MemoryRecord
	for rows.Next() {
		record, err := scanMemoryRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return out, rows.Err()
}

func (s *Store) GetMemory(ctx context.Context, tenantID, subjectID, memoryID string) (memory.MemoryRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT `+memoryRecordSelectCols+`
FROM memory_records
WHERE tenant_id = $1 AND subject_id = $2 AND memory_id = $3
`, tenantID, subjectID, memoryID)
	record, err := scanMemoryRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return memory.MemoryRecord{}, memory.ErrMemoryNotFound
		}
		return memory.MemoryRecord{}, err
	}
	return record, nil
}

func (s *Store) MarkSuperseded(ctx context.Context, tenantID, subjectID, memoryID string) error {
	now := time.Now().UTC()
	commandTag, err := s.pool.Exec(ctx, `
UPDATE memory_records
SET lifecycle_state = $4,
    superseded_at = $5,
    updated_at = $5
WHERE tenant_id = $1 AND subject_id = $2 AND memory_id = $3
`, tenantID, subjectID, memoryID, memory.LifecycleSuperseded, now)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return memory.ErrMemoryNotFound
	}
	return nil
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
	row := tx.QueryRow(ctx, `
SELECT `+memoryRecordSelectCols+`
FROM memory_records
WHERE tenant_id = $1 AND subject_id = $2 AND memory_id = $3
FOR UPDATE
`, tenantID, subjectID, memoryID)
	record, err = scanMemoryRow(row)
	if err != nil {
		return memory.MemoryRecord{}, err
	}

	previousContent := record.Content
	now := time.Now().UTC()
	record.Content = content
	record.SourceText = sourceText
	record.DedupeKey = memory.DedupeKey(tenantID, subjectID, record.Kind, content)
	record.Status = memory.StatusActive
	record.UpdatedAt = now
	record.CorrectedAt = &now

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
    updated_at = $9,
    corrected_at = $10
WHERE memory_id = $1 AND tenant_id = $2 AND subject_id = $3
`, record.MemoryID, record.TenantID, record.SubjectID, record.Content, record.SourceText, record.DedupeKey, record.Status, explain, record.UpdatedAt, record.CorrectedAt); err != nil {
		return memory.MemoryRecord{}, err
	}

	historyID := fmt.Sprintf("hist_%d", now.UnixNano())
	if _, err := tx.Exec(ctx, `
INSERT INTO correction_history (
    history_id, memory_id, tenant_id, subject_id, previous_content, corrected_content, source_text, corrected_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, historyID, memoryID, tenantID, subjectID, previousContent, content, sourceText, now); err != nil {
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
