package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"brainy/internal/memory"
)

type apiStoreStub struct {
	memoryStore *memoryStoreAdapter
}

type memoryStoreAdapter struct {
	records map[string]memory.MemoryRecord
}

func newMemoryStoreAdapter() *memoryStoreAdapter {
	return &memoryStoreAdapter{records: map[string]memory.MemoryRecord{}}
}

func (s *memoryStoreAdapter) UpsertMemory(_ context.Context, record memory.MemoryRecord) (memory.StoreUpsertResult, error) {
	if existing, ok := s.records[record.DedupeKey]; ok {
		if existing.Content == record.Content && existing.Status == memory.StatusActive {
			return memory.StoreUpsertResult{Record: existing, State: "deduped"}, nil
		}
		record.MemoryID = existing.MemoryID
		record.CreatedAt = existing.CreatedAt
		s.records[record.DedupeKey] = record
		return memory.StoreUpsertResult{Record: record, State: "updated"}, nil
	}
	s.records[record.DedupeKey] = record
	return memory.StoreUpsertResult{Record: record, State: "created"}, nil
}

func (s *memoryStoreAdapter) ListActiveMemories(_ context.Context, tenantID, subjectID string) ([]memory.MemoryRecord, error) {
	var out []memory.MemoryRecord
	for _, record := range s.records {
		if record.TenantID == tenantID && record.SubjectID == subjectID && record.Status == memory.StatusActive {
			out = append(out, record)
		}
	}
	return out, nil
}

func (s *memoryStoreAdapter) SuppressMemory(_ context.Context, tenantID, subjectID, memoryID string) error {
	for key, record := range s.records {
		if record.TenantID == tenantID && record.SubjectID == subjectID && record.MemoryID == memoryID {
			record.Status = memory.StatusSuppressed
			s.records[key] = record
			return nil
		}
	}
	return nil
}

func TestRouterIngestAndSearch(t *testing.T) {
	service := memory.NewService(newMemoryStoreAdapter())
	handler := NewRouter(service)

	body, err := json.Marshal(memory.IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages: []memory.Message{
			{Role: "user", Content: "I prefer concise answers."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(body))
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected ingest status 200, got %d", recorder.Code)
	}

	searchRecorder := httptest.NewRecorder()
	searchRequest := httptest.NewRequest(http.MethodGet, "/memories/search?tenant_id=t1&subject_id=u1&q=How+should+I+respond", nil)
	handler.ServeHTTP(searchRecorder, searchRequest)
	if searchRecorder.Code != http.StatusOK {
		t.Fatalf("expected search status 200, got %d", searchRecorder.Code)
	}

	var response memory.SearchResponse
	if err := json.Unmarshal(searchRecorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode search response: %v", err)
	}
	if len(response.Results) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(response.Results))
	}
}
