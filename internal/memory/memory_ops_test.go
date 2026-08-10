package memory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNormalizeMemoryEvent(t *testing.T) {
	if got := normalizeMemoryEvent("update"); got != MemoryEventUpdate {
		t.Fatalf("got %q", got)
	}
	if got := normalizeMemoryEvent("noop"); got != MemoryEventNone {
		t.Fatalf("got %q", got)
	}
}

func TestPrepareExtractedForPersistNoneAndUpdate(t *testing.T) {
	none := ExtractedMemory{Explain: map[string]any{"memory_event": MemoryEventNone}}
	if PrepareExtractedForPersist(&none) {
		t.Fatal("NONE should not persist")
	}
	upd := ExtractedMemory{
		Content: "Jordan is a doctor",
		Explain: map[string]any{
			"memory_event":     MemoryEventUpdate,
			"target_memory_id": "mem_prior",
		},
	}
	if !PrepareExtractedForPersist(&upd) {
		t.Fatal("UPDATE should persist")
	}
	if upd.Explain["supersedes_memory_id"] != "mem_prior" {
		t.Fatalf("expected supersedes_memory_id, got %#v", upd.Explain)
	}
}

func TestIngestUpdateSupersedesPrior(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	created, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-ops", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{{Role: "user", Content: "Jordan: I work as a nurse"}},
	})
	if err != nil || len(created.Memories) == 0 {
		t.Fatalf("seed ingest: %#v err=%v", created, err)
	}
	priorID := created.Memories[0].MemoryID

	// Force provider path with UPDATE event.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content": `{"memories":[{"event":"UPDATE","target_memory_id":"` + priorID + `","kind":"fact","content":"Jordan is a doctor","source_text":"Jordan: I am a doctor now","confidence":0.95,"subject":"Jordan","predicate":"occupation","value":"doctor","assertion_kind":"corrective"}]}`,
				}},
			},
		})
	}))
	defer server.Close()
	svc.WithExtractor(NewContextualExtractor(NewProviderExtractor(ProviderConfig{
		BaseURL: server.URL,
		APIKey:  "test",
		Model:   "test-model",
	}, server.Client()), store))

	out, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-ops", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{{Role: "user", Content: "Jordan: I am a doctor now"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Memories) == 0 {
		t.Fatal("expected replacement memory")
	}
	prior, err := store.GetMemory(context.Background(), "t-ops", "u1", priorID)
	if err != nil {
		t.Fatal(err)
	}
	if prior.LifecycleState != LifecycleSuperseded {
		t.Fatalf("expected prior superseded, got %#v", prior)
	}
	active, err := store.ListActiveMemories(context.Background(), "t-ops", "u1")
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, r := range active {
		joined += " " + strings.ToLower(r.Content)
		if r.MemoryID == priorID {
			t.Fatalf("superseded prior still active: %#v", r)
		}
	}
	if !strings.Contains(joined, "doctor") {
		t.Fatalf("expected doctor in active memories, got %q", joined)
	}
}

func TestIngestDeleteSuppressesTarget(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	created, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-del", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{{Role: "user", Content: "Sam keeps a pet snake named Noodle"}},
	})
	if err != nil || len(created.Memories) == 0 {
		t.Fatalf("seed: %#v err=%v", created, err)
	}
	target := created.Memories[0].MemoryID

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content": `{"memories":[{"event":"DELETE","target_memory_id":"` + target + `","kind":"fact","content":"Sam no longer has a pet snake","confidence":0.9}]}`,
				}},
			},
		})
	}))
	defer server.Close()
	svc.WithExtractor(NewProviderExtractor(ProviderConfig{
		BaseURL: server.URL, APIKey: "test", Model: "m",
	}, server.Client()))

	_, err = svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-del", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{{Role: "user", Content: "Sam: I rehomed Noodle, no more snake"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := store.GetMemory(context.Background(), "t-del", "u1", target)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusSuppressed && got.LifecycleState != LifecycleSuppressed {
		t.Fatalf("expected suppressed, got %#v", got)
	}
}

func TestParseProviderMemoriesEvents(t *testing.T) {
	memories, err := parseProviderMemories(`{"memories":[
	  {"event":"NONE","kind":"fact","content":"duplicate hello","confidence":0.5},
	  {"event":"ADD","kind":"fact","content":"Alex lives in Austin","confidence":0.9}
	]}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(memories) != 2 {
		t.Fatalf("got %d", len(memories))
	}
	if MemoryEventOf(memories[0]) != MemoryEventNone {
		t.Fatalf("event0=%s", MemoryEventOf(memories[0]))
	}
	if MemoryEventOf(memories[1]) != MemoryEventAdd {
		t.Fatalf("event1=%s", MemoryEventOf(memories[1]))
	}
}
