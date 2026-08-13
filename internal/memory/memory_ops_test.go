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

func TestWriteMutationModeOf(t *testing.T) {
	if got := WriteMutationModeOf(IngestRequest{SourceType: "conversation"}); got != WriteModeAppendOnly {
		t.Fatalf("empty vertical: %s", got)
	}
	if got := WriteMutationModeOf(IngestRequest{Vertical: VerticalCore}); got != WriteModeAppendOnly {
		t.Fatalf("core: %s", got)
	}
	if got := WriteMutationModeOf(IngestRequest{Vertical: "support"}); got != WriteModeGoverned {
		t.Fatalf("support: %s", got)
	}
	if got := WriteMutationModeOf(IngestRequest{Vertical: "marketing"}); got != WriteModeGoverned {
		t.Fatalf("marketing: %s", got)
	}
}

func TestPrepareExtractedForPersistNoneAndUpdate(t *testing.T) {
	none := ExtractedMemory{Explain: map[string]any{"memory_event": MemoryEventNone}}
	if PrepareExtractedForPersist(&none, WriteModeGoverned) {
		t.Fatal("NONE should not persist")
	}
	upd := ExtractedMemory{
		Content: "Jordan is a doctor",
		Explain: map[string]any{
			"memory_event":     MemoryEventUpdate,
			"target_memory_id": "mem_prior",
		},
	}
	if !PrepareExtractedForPersist(&upd, WriteModeGoverned) {
		t.Fatal("UPDATE should persist")
	}
	if upd.Explain["supersedes_memory_id"] != "mem_prior" {
		t.Fatalf("expected supersedes_memory_id, got %#v", upd.Explain)
	}

	convUpd := ExtractedMemory{
		Content: "Alex lives in Austin",
		Explain: map[string]any{
			"memory_event":     MemoryEventUpdate,
			"target_memory_id": "mem_ny",
		},
	}
	if !PrepareExtractedForPersist(&convUpd, WriteModeAppendOnly) {
		t.Fatal("conversational UPDATE should persist as ADD")
	}
	if MemoryEventOf(convUpd) != MemoryEventAdd {
		t.Fatalf("expected ADD rewrite, got %s", MemoryEventOf(convUpd))
	}
	if _, ok := convUpd.Explain["supersedes_memory_id"]; ok {
		t.Fatalf("append-only UPDATE must not set supersedes_memory_id: %#v", convUpd.Explain)
	}

	convDel := ExtractedMemory{
		Content: "Sam rehomed the snake",
		Explain: map[string]any{
			"memory_event":     MemoryEventDelete,
			"target_memory_id": "mem_snake",
		},
	}
	if !PrepareExtractedForPersist(&convDel, WriteModeAppendOnly) {
		t.Fatal("conversational DELETE should persist as ADD")
	}
	if MemoryEventOf(convDel) != MemoryEventAdd {
		t.Fatalf("expected ADD rewrite, got %s", MemoryEventOf(convDel))
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
	if prior.LifecycleState == LifecycleSuperseded || prior.Status == StatusSuppressed {
		t.Fatalf("conversational UPDATE must retain prior history, got %#v", prior)
	}
	active, err := store.ListActiveMemories(context.Background(), "t-ops", "u1")
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	foundPrior := false
	for _, r := range active {
		joined += " " + strings.ToLower(r.Content)
		if r.MemoryID == priorID {
			foundPrior = true
		}
	}
	if !foundPrior {
		t.Fatalf("expected prior nurse fact still active, got %q", joined)
	}
	if !strings.Contains(joined, "doctor") {
		t.Fatalf("expected doctor in active memories, got %q", joined)
	}
}

func TestIngestUpdateSupersedesPriorGoverned(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	created, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-gov", SubjectID: "u1", SourceType: "crm", Vertical: "support",
		Messages: []Message{{Role: "user", Content: "Jordan: I work as a nurse"}},
	})
	if err != nil || len(created.Memories) == 0 {
		t.Fatalf("seed ingest: %#v err=%v", created, err)
	}
	priorID := created.Memories[0].MemoryID

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
		TenantID: "t-gov", SubjectID: "u1", SourceType: "crm", Vertical: "support",
		Messages: []Message{{Role: "user", Content: "Jordan: I am a doctor now"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Memories) == 0 {
		t.Fatal("expected replacement memory")
	}
	prior, err := store.GetMemory(context.Background(), "t-gov", "u1", priorID)
	if err != nil {
		t.Fatal(err)
	}
	if prior.LifecycleState != LifecycleSuperseded {
		t.Fatalf("expected governed prior superseded, got %#v", prior)
	}
	active, err := store.ListActiveMemories(context.Background(), "t-gov", "u1")
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
	if got.Status == StatusSuppressed || got.LifecycleState == LifecycleSuppressed {
		t.Fatalf("conversational DELETE must not suppress prior, got %#v", got)
	}
}

func TestIngestDeleteSuppressesTargetGoverned(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	created, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-del-gov", SubjectID: "u1", SourceType: "crm", Vertical: "support",
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
		TenantID: "t-del-gov", SubjectID: "u1", SourceType: "crm", Vertical: "support",
		Messages: []Message{{Role: "user", Content: "Sam: I rehomed Noodle, no more snake"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := store.GetMemory(context.Background(), "t-del-gov", "u1", target)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusSuppressed && got.LifecycleState != LifecycleSuppressed {
		t.Fatalf("expected governed suppressed, got %#v", got)
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

func mergeOpsFixture() (baseline, provider []ExtractedMemory) {
	baseline = []ExtractedMemory{
		{
			Kind: KindFact, Content: "Alex lives in New York",
			SourceText: "Alex: I live in New York",
			Explain:    map[string]any{"rule": "deterministic", "subject": "Alex", "predicate": PredicateResidence},
		},
		{
			Kind: KindFact, Content: "Alex likes hiking on weekends",
			SourceText: "Alex: I like hiking",
			Explain:    map[string]any{"rule": "deterministic", "subject": "Alex", "predicate": PredicatePreference},
		},
		{
			Kind: KindFact, Content: "Sam keeps a pet snake named Noodle",
			SourceText: "Sam keeps a pet snake named Noodle",
			Explain:    map[string]any{"rule": "deterministic"},
		},
	}
	provider = []ExtractedMemory{
		{
			Kind: KindFact, Content: "Alex lives in Austin",
			SourceText: "Alex: I moved to Austin",
			Explain: map[string]any{
				"rule": "provider_extract", "memory_event": MemoryEventUpdate,
				"subject": "Alex", "predicate": PredicateResidence, "target_memory_id": "mem_ny",
			},
		},
		{
			Kind: KindFact, Content: "Sam rehomed the snake",
			SourceText: "Sam keeps a pet snake named Noodle",
			Explain: map[string]any{
				"rule": "provider_extract", "memory_event": MemoryEventDelete,
				"target_memory_id": "mem_snake",
			},
		},
		{
			Kind: KindFact, Content: "Alex prefers trail running",
			SourceText: "Alex: I like trail running now",
			Explain: map[string]any{
				"rule": "provider_extract", "memory_event": MemoryEventUpdate,
				"subject": "Alex", "predicate": PredicatePreference, "target_memory_id": "mem_hike",
			},
		},
	}
	return baseline, provider
}

func TestMergeProviderAppendOnlyKeepsBaselineOnNone(t *testing.T) {
	baseline := []ExtractedMemory{{
		Kind: KindFact, Content: "Alex lives in Austin Texas",
		SourceText: "Alex: I live in Austin",
		Explain:    map[string]any{"rule": "deterministic", "subject": "Alex", "predicate": PredicateResidence},
	}}
	provider := []ExtractedMemory{{
		Kind: KindFact, Content: "duplicate austin residence",
		SourceText: "Alex: I live in Austin",
		Explain: map[string]any{
			"rule": "provider_extract", "memory_event": MemoryEventNone,
			"subject": "Alex", "predicate": PredicateResidence,
		},
	}}
	merged := mergeProviderAndBaseline(baseline, provider, WriteModeAppendOnly)
	joined := ""
	for _, m := range merged {
		joined += " | " + m.Content
	}
	if !strings.Contains(joined, "Austin Texas") {
		t.Fatalf("NONE must not drop conversational baseline: %q", joined)
	}
}

func TestMergeProviderAppendOnlyRetainsHistory(t *testing.T) {
	baseline, provider := mergeOpsFixture()
	merged := mergeProviderAndBaseline(baseline, provider, WriteModeAppendOnly)
	joined := ""
	for _, m := range merged {
		joined += " | " + m.Content
	}
	if !strings.Contains(joined, "New York") {
		t.Fatalf("moved-to-Austin must retain NY history: %q", joined)
	}
	if !strings.Contains(joined, "Austin") {
		t.Fatalf("expected Austin update kept, got %q", joined)
	}
	if !strings.Contains(joined, "hiking") {
		t.Fatalf("expected hiking history kept, got %q", joined)
	}
	if !strings.Contains(joined, "Noodle") {
		t.Fatalf("expected snake history kept, got %q", joined)
	}
	if !strings.Contains(joined, "trail running") {
		t.Fatalf("expected UPDATE content kept, got %q", joined)
	}
}

func TestMergeProviderGovernedSuppressesBaseline(t *testing.T) {
	baseline, provider := mergeOpsFixture()
	merged := mergeProviderAndBaseline(baseline, provider, WriteModeGoverned)
	joined := ""
	for _, m := range merged {
		joined += " | " + m.Content
		if MemoryEventOf(m) == MemoryEventAdd && strings.Contains(strings.ToLower(m.Content), "hiking") {
			t.Fatalf("governed baseline preference should be suppressed by UPDATE: %#v", merged)
		}
		if strings.Contains(m.Content, "New York") {
			t.Fatalf("governed baseline residence should be suppressed by UPDATE: %#v", merged)
		}
		if strings.Contains(strings.ToLower(m.Content), "noodle") && MemoryEventOf(m) == MemoryEventAdd {
			t.Fatalf("governed baseline snake fact should be suppressed by DELETE: %#v", merged)
		}
	}
	if !strings.Contains(joined, "Austin") {
		t.Fatalf("expected UPDATE residence kept, got %q", joined)
	}
	if !strings.Contains(joined, "trail running") {
		t.Fatalf("expected UPDATE content kept, got %q", joined)
	}
}
