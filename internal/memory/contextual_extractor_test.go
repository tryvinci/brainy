package memory

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestContextualExtractorRecentMemoriesNewestFirst(t *testing.T) {
	now := time.Now().UTC()
	store := newMemoryStoreStub()
	// Insert oldest first so map iteration order is irrelevant; ListActiveMemories
	// on the stub does not sort, so we control order by overriding via a thin wrapper.
	ordered := &orderedRecentStore{
		memoryStoreStub: store,
		recent: []MemoryRecord{
			{TenantID: "t1", SubjectID: "s1", Status: StatusActive, LifecycleState: LifecycleActive, Content: "newest fact", UpdatedAt: now},
			{TenantID: "t1", SubjectID: "s1", Status: StatusActive, LifecycleState: LifecycleActive, Content: "middle fact", UpdatedAt: now.Add(-time.Hour)},
			{TenantID: "t1", SubjectID: "s1", Status: StatusActive, LifecycleState: LifecycleActive, Content: "oldest fact", UpdatedAt: now.Add(-2 * time.Hour)},
		},
	}
	cap := &captureExtractor{}
	ext := NewContextualExtractor(cap, ordered)
	ext.limit = 2

	_, err := ext.Extract(context.Background(), IngestRequest{
		TenantID:  "t1",
		SubjectID: "s1",
		Messages:  []Message{{Role: "user", Content: "I like hiking"}},
	})
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	ctxBlock, _ := cap.last.Metadata["extract_context"].(string)
	if ctxBlock == "" {
		t.Fatal("expected extract_context metadata")
	}
	newestIdx := strings.Index(ctxBlock, "newest fact")
	middleIdx := strings.Index(ctxBlock, "middle fact")
	oldestIdx := strings.Index(ctxBlock, "oldest fact")
	if newestIdx < 0 || middleIdx < 0 {
		t.Fatalf("expected newest and middle in context block:\n%s", ctxBlock)
	}
	if newestIdx > middleIdx {
		t.Fatalf("expected newest before middle in Recent memories:\n%s", ctxBlock)
	}
	if oldestIdx >= 0 {
		t.Fatalf("limit=2 should omit oldest fact:\n%s", ctxBlock)
	}
}

type orderedRecentStore struct {
	*memoryStoreStub
	recent []MemoryRecord
}

func (s *orderedRecentStore) ListActiveMemories(context.Context, string, string) ([]MemoryRecord, error) {
	return append([]MemoryRecord(nil), s.recent...), nil
}

type captureExtractor struct {
	last IngestRequest
}

func (c *captureExtractor) Extract(_ context.Context, req IngestRequest) ([]ExtractedMemory, error) {
	c.last = req
	return nil, nil
}
