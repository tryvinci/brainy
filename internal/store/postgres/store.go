package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"brainy/internal/memory"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func NewWithPool(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
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
		_ = json.Unmarshal(insertedExplain, &inserted.Explain)
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
	_ = json.Unmarshal(rawExplain, &existing.Explain)

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
		return errors.New("memory not found")
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
	_ = json.Unmarshal(rawExplain, &record.Explain)

	record.Content = content
	record.SourceText = sourceText
	record.DedupeKey = memory.DedupeKey(tenantID, subjectID, record.Kind, content)
	record.Status = memory.StatusActive
	record.UpdatedAt = time.Now().UTC()

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
