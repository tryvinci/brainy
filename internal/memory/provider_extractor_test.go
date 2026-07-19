package memory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestProviderExtractorParsesStructuredMemories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content": `{"memories":[{"kind":"fact","content":"Caroline went to LGBTQ support group","source_text":"I went to the LGBTQ support group on 7 May 2023","confidence":0.9,"when":"2023-05-07"}]}`,
				}},
			},
		})
	}))
	defer server.Close()

	extractor := NewProviderExtractor(ProviderConfig{
		BaseURL: server.URL,
		APIKey:  "test",
		Model:   "test-model",
	}, server.Client())

	memories, err := extractor.Extract(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Caroline: I went to the LGBTQ support group on 7 May 2023"},
		},
	})
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	if len(memories) < 1 {
		t.Fatalf("expected at least provider memory, got %d", len(memories))
	}
	var providerFact *ExtractedMemory
	for i := range memories {
		if memories[i].Explain["rule"] == "provider_extract" {
			providerFact = &memories[i]
			break
		}
	}
	if providerFact == nil {
		t.Fatalf("expected provider_extract memory, got %#v", memories)
	}
	if providerFact.When != "2023-05-07" {
		t.Fatalf("expected when slot, got %#v", providerFact)
	}
	if providerFact.Explain["primitive"] != PrimitiveEpisode {
		t.Fatalf("expected episode primitive for dated fact")
	}
}

func TestProviderExtractorMergesBaselineEpisodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content": `{"memories":[{"kind":"fact","content":"Caroline attended LGBTQ support group","source_text":"LGBTQ support group","confidence":0.9,"when":"2023-05-07"}]}`,
				}},
			},
		})
	}))
	defer server.Close()

	extractor := NewProviderExtractor(ProviderConfig{
		BaseURL: server.URL,
		Model:   "test-model",
	}, server.Client())

	utterance := "Caroline: I went to the LGBTQ support group on 7 May 2023"
	memories, err := extractor.Extract(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []Message{{Role: "user", Content: utterance}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(memories) < 2 {
		t.Fatalf("expected provider fact + baseline episode, got %d (%#v)", len(memories), memories)
	}
	var sawProvider, sawEpisode bool
	for _, m := range memories {
		if m.Explain["rule"] == "provider_extract" {
			sawProvider = true
		}
		if m.Explain["rule"] == "conversation_episode" {
			sawEpisode = true
		}
	}
	if !sawProvider || !sawEpisode {
		t.Fatalf("expected merge of provider+episode, got %#v", memories)
	}
}

func TestProviderExtractorFailureSoftDegradesToBaseline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusGatewayTimeout)
	}))
	defer server.Close()

	extractor := NewProviderExtractor(ProviderConfig{
		BaseURL: server.URL,
		Model:   "test-model",
	}, server.Client())

	memories, err := extractor.Extract(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []Message{{Role: "user", Content: "I prefer concise answers."}},
	})
	if err != nil {
		t.Fatalf("expected soft-degrade to baseline, got %v", err)
	}
	if len(memories) == 0 {
		t.Fatal("expected deterministic preference baseline")
	}
}

func TestProviderExtractorFailureWithEmptyBaselinePropagates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusGatewayTimeout)
	}))
	defer server.Close()

	extractor := NewProviderExtractor(ProviderConfig{
		BaseURL: server.URL,
		Model:   "test-model",
	}, server.Client())

	_, err := extractor.Extract(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "note",
		Messages:   []Message{{Role: "user", Content: "hello there friend"}},
	})
	if err == nil {
		t.Fatal("expected provider failure when baseline empty")
	}
}

func TestParseProviderMemoriesRejectsInvalidKind(t *testing.T) {
	_, err := parseProviderMemories(`{"memories":[{"kind":"note","content":"x","source_text":"x"}]}`)
	if err == nil {
		t.Fatal("expected invalid kind error")
	}
}

func TestBuildMemoryRecordSetsObservedAt(t *testing.T) {
	record, err := BuildMemoryRecord("mem_1", mustParseTime(t, "2026-07-19T00:00:00Z"), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Metadata: map[string]any{
			"observed_at": "2023-05-07T18:00:00Z",
		},
	}, ExtractedMemory{
		Kind:       KindFact,
		Content:    "Caroline went to LGBTQ support group",
		SourceText: "I went to the LGBTQ support group on 7 May 2023",
		Confidence: 0.9,
		Explain:    map[string]any{"rule": "provider_extract", "primitive": PrimitiveEpisode},
		When:       "2023-05-07",
	}, nil)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if record.ObservedAt == nil {
		t.Fatal("expected ObservedAt")
	}
	if record.ObservedAt.Format("2006-01-02") != "2023-05-07" {
		t.Fatalf("expected observed_at from metadata, got %s", record.ObservedAt)
	}
	if record.Primitive != PrimitiveEpisode {
		t.Fatalf("expected episode primitive, got %s", record.Primitive)
	}
	if record.ExtractionVersion != providerExtractionVersion {
		t.Fatalf("expected provider version, got %s", record.ExtractionVersion)
	}
}

func TestEnrichRelativeEventTimeAppendsAbsoluteDate(t *testing.T) {
	at := mustParseTime(t, "2023-05-08T18:00:00Z")
	got := EnrichRelativeEventTime("Caroline: I went to the LGBTQ support group yesterday", at)
	if !strings.Contains(got, "8 May 2023") {
		t.Fatalf("expected absolute date annotation, got %q", got)
	}
	same := EnrichRelativeEventTime("Caroline went on 8 May 2023", at)
	if same != "Caroline went on 8 May 2023" {
		t.Fatalf("should not double-annotate absolute dates: %q", same)
	}
}

func mustParseTime(t *testing.T, raw string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t.Fatal(err)
	}
	return ts.UTC()
}

func TestProviderExtractorEmptyCompletionFallsBackToBaseline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": ""}},
			},
		})
	}))
	defer server.Close()

	extractor := NewProviderExtractor(ProviderConfig{
		BaseURL: server.URL,
		Model:   "test-model",
	}, server.Client())

	memories, err := extractor.Extract(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []Message{{Role: "user", Content: "Caroline: I went to the LGBTQ support group on 7 May 2023"}},
	})
	if err != nil {
		t.Fatalf("empty completion should soft-degrade, got %v", err)
	}
	if len(memories) == 0 {
		t.Fatal("expected baseline episode memories")
	}
}
