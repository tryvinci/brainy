package memory

import (
	"context"
	"testing"
	"time"
)

type memoryStoreStub struct {
	records map[string]MemoryRecord
}

func newMemoryStoreStub() *memoryStoreStub {
	return &memoryStoreStub{records: map[string]MemoryRecord{}}
}

func (s *memoryStoreStub) UpsertMemory(_ context.Context, record MemoryRecord) (StoreUpsertResult, error) {
	if existing, ok := s.records[record.DedupeKey]; ok {
		if existing.Content == record.Content && existing.Status == StatusActive {
			return StoreUpsertResult{Record: existing, State: "deduped"}, nil
		}
		record.MemoryID = existing.MemoryID
		record.CreatedAt = existing.CreatedAt
		s.records[record.DedupeKey] = record
		return StoreUpsertResult{Record: record, State: "updated"}, nil
	}
	s.records[record.DedupeKey] = record
	return StoreUpsertResult{Record: record, State: "created"}, nil
}

func (s *memoryStoreStub) ListActiveMemories(_ context.Context, tenantID, subjectID string) ([]MemoryRecord, error) {
	var out []MemoryRecord
	for _, record := range s.records {
		if record.TenantID == tenantID && record.SubjectID == subjectID && record.Status == StatusActive {
			out = append(out, record)
		}
	}
	return out, nil
}

func (s *memoryStoreStub) SuppressMemory(_ context.Context, tenantID, subjectID, memoryID string) error {
	for key, record := range s.records {
		if record.TenantID == tenantID && record.SubjectID == subjectID && record.MemoryID == memoryID {
			record.Status = StatusSuppressed
			record.UpdatedAt = time.Now().UTC()
			s.records[key] = record
			return nil
		}
	}
	return nil
}

func TestServiceIngestSearchAndSuppress(t *testing.T) {
	store := newMemoryStoreStub()
	service := NewService(store)

	result, err := service.Ingest(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "I prefer concise, direct answers."},
		},
	})
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if result.Created != 1 {
		t.Fatalf("expected 1 created memory, got %d", result.Created)
	}

	search, err := service.Search(context.Background(), "t1", "u1", "How should I respond?")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(search.Results) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(search.Results))
	}

	if err := service.Suppress(context.Background(), "t1", "u1", result.Memories[0].MemoryID); err != nil {
		t.Fatalf("suppress failed: %v", err)
	}

	searchAfterSuppress, err := service.Search(context.Background(), "t1", "u1", "How should I respond?")
	if err != nil {
		t.Fatalf("search after suppress failed: %v", err)
	}
	if len(searchAfterSuppress.Results) != 0 {
		t.Fatalf("expected 0 results after suppression, got %d", len(searchAfterSuppress.Results))
	}
}
