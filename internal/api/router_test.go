package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"brainy/internal/memory"
	"brainy/internal/observability"
)

type memoryStoreAdapter struct {
	records map[string]memory.MemoryRecord
	jobs    map[string]memory.ExtractionJob
}

func newMemoryStoreAdapter() *memoryStoreAdapter {
	return &memoryStoreAdapter{
		records: map[string]memory.MemoryRecord{},
		jobs:    map[string]memory.ExtractionJob{},
	}
}

func (s *memoryStoreAdapter) UpsertMemory(_ context.Context, record memory.MemoryRecord) (memory.StoreUpsertResult, error) {
	if existing, ok := s.records[record.DedupeKey]; ok {
		if existing.Status == memory.StatusSuppressed {
			return memory.StoreUpsertResult{Record: existing, State: "deduped"}, nil
		}
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

func (s *memoryStoreAdapter) SearchActiveMemories(_ context.Context, tenantID, subjectID string, patterns []string, limit int) ([]memory.MemoryRecord, error) {
	_ = patterns
	return s.ListActiveMemories(context.Background(), tenantID, subjectID)
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

func (s *memoryStoreAdapter) CorrectMemory(_ context.Context, tenantID, subjectID, memoryID, content, sourceText string) (memory.MemoryRecord, error) {
	for key, record := range s.records {
		if record.TenantID == tenantID && record.SubjectID == subjectID && record.MemoryID == memoryID {
			newDedupeKey := memory.DedupeKey(tenantID, subjectID, record.Kind, content)
			for otherKey, other := range s.records {
				if otherKey == key {
					continue
				}
				if other.TenantID == tenantID && other.SubjectID == subjectID && other.DedupeKey == newDedupeKey {
					return memory.MemoryRecord{}, memory.ErrMemoryConflict
				}
			}
			delete(s.records, key)
			record.Content = content
			record.SourceText = sourceText
			record.DedupeKey = newDedupeKey
			record.Status = memory.StatusActive
			s.records[record.DedupeKey] = record
			return record, nil
		}
	}
	return memory.MemoryRecord{}, errors.New("memory not found")
}

func (s *memoryStoreAdapter) EnqueueIngestJob(_ context.Context, ingestID, jobID, _ string, req memory.IngestRequest) (memory.EnqueueResult, error) {
	s.jobs[jobID] = memory.ExtractionJob{JobID: jobID, IngestID: ingestID, Request: req}
	return memory.EnqueueResult{IngestID: ingestID, JobID: jobID, Accepted: true}, nil
}

func (s *memoryStoreAdapter) ClaimNextExtractionJob(_ context.Context) (memory.ExtractionJob, bool, error) {
	for jobID, job := range s.jobs {
		delete(s.jobs, jobID)
		return job, true, nil
	}
	return memory.ExtractionJob{}, false, nil
}

func (s *memoryStoreAdapter) CompleteExtractionJob(_ context.Context, _, _ string) error {
	return nil
}

func (s *memoryStoreAdapter) FailExtractionJob(_ context.Context, _, _, _ string) error {
	return nil
}

func TestRouterIngestAndSearch(t *testing.T) {
	service := memory.NewService(newMemoryStoreAdapter())
	handler := NewRouter(service, observability.NewMetrics())

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

func TestRouterReturnsStructuredErrorPayload(t *testing.T) {
	service := memory.NewService(newMemoryStoreAdapter())
	handler := NewRouter(service, observability.NewMetrics())

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/memories/search", nil)
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}

	var payload map[string]map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode error payload: %v", err)
	}
	if payload["error"]["code"] != "bad_request" {
		t.Fatalf("expected bad_request error code, got %q", payload["error"]["code"])
	}
}

func TestRouterCorrectsMemory(t *testing.T) {
	service := memory.NewService(newMemoryStoreAdapter())
	handler := NewRouter(service, observability.NewMetrics())

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

	ingestRecorder := httptest.NewRecorder()
	ingestRequest := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(body))
	handler.ServeHTTP(ingestRecorder, ingestRequest)
	if ingestRecorder.Code != http.StatusOK {
		t.Fatalf("expected ingest status 200, got %d", ingestRecorder.Code)
	}

	var ingestResponse memory.IngestResult
	if err := json.Unmarshal(ingestRecorder.Body.Bytes(), &ingestResponse); err != nil {
		t.Fatal(err)
	}

	correctRecorder := httptest.NewRecorder()
	correctRequest := httptest.NewRequest(
		http.MethodPost,
		"/memories/"+ingestResponse.Memories[0].MemoryID+"/correct?tenant_id=t1&subject_id=u1",
		bytes.NewReader([]byte(`{"content":"Prefers detailed answers"}`)),
	)
	handler.ServeHTTP(correctRecorder, correctRequest)
	if correctRecorder.Code != http.StatusOK {
		t.Fatalf("expected correction status 200, got %d", correctRecorder.Code)
	}

	searchRecorder := httptest.NewRecorder()
	searchRequest := httptest.NewRequest(http.MethodGet, "/memories/search?tenant_id=t1&subject_id=u1&q=How+should+I+answer", nil)
	handler.ServeHTTP(searchRecorder, searchRequest)

	var searchResponse memory.SearchResponse
	if err := json.Unmarshal(searchRecorder.Body.Bytes(), &searchResponse); err != nil {
		t.Fatal(err)
	}
	if len(searchResponse.Results) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(searchResponse.Results))
	}
	if searchResponse.Results[0].Content != "Prefers detailed answers" {
		t.Fatalf("expected corrected content, got %q", searchResponse.Results[0].Content)
	}
}

func TestRouterReturnsConflictForDuplicateCorrection(t *testing.T) {
	service := memory.NewService(newMemoryStoreAdapter())
	handler := NewRouter(service, observability.NewMetrics())

	for _, content := range []string{"I prefer concise answers.", "I prefer detailed answers."} {
		body, err := json.Marshal(memory.IngestRequest{
			TenantID:   "t1",
			SubjectID:  "u1",
			SourceType: "conversation",
			Messages: []memory.Message{
				{Role: "user", Content: content},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(body))
		handler.ServeHTTP(recorder, request)
	}

	searchRecorder := httptest.NewRecorder()
	searchRequest := httptest.NewRequest(http.MethodGet, "/memories/search?tenant_id=t1&subject_id=u1&q=How+should+I+answer", nil)
	handler.ServeHTTP(searchRecorder, searchRequest)
	var searchResponse memory.SearchResponse
	if err := json.Unmarshal(searchRecorder.Body.Bytes(), &searchResponse); err != nil {
		t.Fatal(err)
	}

	var conciseMemoryID string
	for _, result := range searchResponse.Results {
		if strings.Contains(strings.ToLower(result.Content), "concise") {
			conciseMemoryID = result.MemoryID
			break
		}
	}
	if conciseMemoryID == "" {
		t.Fatal("expected concise preference in search results")
	}

	correctRecorder := httptest.NewRecorder()
	correctRequest := httptest.NewRequest(
		http.MethodPost,
		"/memories/"+conciseMemoryID+"/correct?tenant_id=t1&subject_id=u1",
		bytes.NewReader([]byte(`{"content":"Prefers detailed answers"}`)),
	)
	handler.ServeHTTP(correctRecorder, correctRequest)
	if correctRecorder.Code != http.StatusConflict {
		t.Fatalf("expected conflict status 409, got %d", correctRecorder.Code)
	}
}

func TestRouterAsyncIngestReturnsAcceptedJob(t *testing.T) {
	service := memory.NewService(newMemoryStoreAdapter())
	handler := NewRouter(service, observability.NewMetrics())

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
	request := httptest.NewRequest(http.MethodPost, "/ingest/async", bytes.NewReader(body))
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected async ingest status 202, got %d", recorder.Code)
	}

	var response memory.AsyncIngestResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if !response.Accepted || response.JobID == "" || response.IngestID == "" {
		t.Fatalf("expected accepted async ingest with ids, got %+v", response)
	}
}
