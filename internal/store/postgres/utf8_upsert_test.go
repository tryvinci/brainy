package postgres

import (
	"context"
	"os"
	"testing"
	"time"
	"unicode/utf8"

	"brainy/internal/memory"
)

func TestUpsertMemorySanitizesInvalidUTF8(t *testing.T) {
	url := os.Getenv("BRAINY_DATABASE_URL")
	if url == "" {
		t.Skip("BRAINY_DATABASE_URL not set")
	}
	store, err := New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(store.Close)

	now := time.Now().UTC()
	suffix := now.Format("150405.000000")
	bad := "hello\x80world • bullets"
	rec := memory.MemoryRecord{
		MemoryID:          "mem_utf8_test_" + suffix,
		TenantID:          "t_utf8_test",
		SubjectID:         "s_utf8_test",
		Kind:              "fact",
		Content:           bad,
		SourceText:        bad,
		SourceType:        "test",
		DedupeKey:         "utf8-test-" + suffix,
		Status:            memory.StatusActive,
		Confidence:        0.9,
		ExtractionVersion: "test",
		Vertical:          memory.VerticalCore,
		LifecycleState:    memory.LifecycleActive,
		CreatedAt:         now,
		UpdatedAt:         now,
		Explain:           map[string]any{"note": bad},
		Metadata:          map[string]any{"x": bad},
	}
	got, err := store.UpsertMemory(context.Background(), rec)
	if err != nil {
		t.Fatalf("UpsertMemory: %v", err)
	}
	if !utf8.ValidString(got.Record.Content) {
		t.Fatalf("stored content still invalid utf8: %q", got.Record.Content)
	}
	if got.Record.Content != "helloworld • bullets" {
		t.Fatalf("content=%q", got.Record.Content)
	}
}
