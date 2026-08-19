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
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["max_tokens"] != float64(providerMaxCompletionTokens) {
			t.Fatalf("expected max_tokens=%d, got %#v", providerMaxCompletionTokens, body["max_tokens"])
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
	if providerFact.Explain["primitive"] == PrimitiveEpisode {
		t.Fatal("dated provider facts must not be tagged as provenance episodes")
	}
}

func TestProviderExtractorParsesTypedSlotsV3(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content": `{"memories":[{"kind":"fact","content":"Jordan is a nurse","source_text":"Jordan: I work as a nurse","confidence":0.91,"subject":"Jordan","predicate":"occupation","value":"nurse","assertion_kind":"explicit"}]}`,
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
		Messages:   []Message{{Role: "user", Content: "Jordan: I work as a nurse"}},
	})
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	var providerFact *ExtractedMemory
	for i := range memories {
		if memories[i].Explain["rule"] == "provider_extract" {
			providerFact = &memories[i]
			break
		}
	}
	if providerFact == nil {
		t.Fatalf("expected provider_extract memory")
	}
	if providerFact.Explain["predicate"] != PredicateOccupation {
		t.Fatalf("predicate=%v", providerFact.Explain["predicate"])
	}
	if providerFact.Explain["value_norm"] != "nurse" {
		t.Fatalf("value_norm=%v", providerFact.Explain["value_norm"])
	}
	if providerFact.Explain["subject"] != "Jordan" {
		t.Fatalf("subject=%v", providerFact.Explain["subject"])
	}
	if providerExtractionVersion != "provider-v4-ops" {
		t.Fatalf("version=%s", providerExtractionVersion)
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
		t.Fatalf("expected provider fact + baseline dated event, got %d (%#v)", len(memories), memories)
	}
	var sawProvider, sawBaseline bool
	for _, m := range memories {
		if m.Explain["rule"] == "provider_extract" {
			sawProvider = true
		}
		if m.Explain["rule"] == "dated_event" || m.Explain["rule"] == "conversation_episode" {
			sawBaseline = true
		}
	}
	if !sawProvider || !sawBaseline {
		t.Fatalf("expected merge of provider+baseline, got %#v", memories)
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

func TestParseProviderMemoriesCoercesUnknownKind(t *testing.T) {
	got, err := parseProviderMemories(`{"memories":[{"kind":"note","content":"x","source_text":"x"}]}`)
	if err != nil {
		t.Fatalf("unknown kind should coerce to fact, got %v", err)
	}
	if len(got) != 1 || got[0].Kind != KindFact {
		t.Fatalf("expected fact, got %#v", got)
	}
}

func TestParseProviderMemoriesPromotesPredicateKind(t *testing.T) {
	got, err := parseProviderMemories(`{"memories":[{"kind":"plan","content":"Caroline plans a trip","source_text":"I am planning a trip"}]}`)
	if err != nil {
		t.Fatalf("predicate-as-kind should coerce, got %v", err)
	}
	if len(got) != 1 || got[0].Kind != KindFact {
		t.Fatalf("expected fact, got %#v", got)
	}
	if got[0].Explain["predicate"] != PredicatePlan {
		t.Fatalf("expected plan predicate, got %#v", got[0].Explain)
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
		Explain:    map[string]any{"rule": "provider_extract"},
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
	if record.Primitive == PrimitiveEpisode {
		t.Fatal("dated provider facts are recall-primary, not provenance episodes")
	}
	if record.ExtractionVersion != providerExtractionVersion {
		t.Fatalf("expected provider version, got %s", record.ExtractionVersion)
	}
}

func TestEnrichRelativeEventTimeAppendsAbsoluteDate(t *testing.T) {
	at := mustParseTime(t, "2023-05-08T18:00:00Z")
	got := EnrichRelativeEventTime("Caroline: I went to the LGBTQ support group yesterday", at)
	if !strings.Contains(got, "7 May 2023") {
		t.Fatalf("expected yesterday resolved to prior day, got %q", got)
	}
	same := EnrichRelativeEventTime("Caroline went on 7 May 2023", at)
	if same != "Caroline went on 7 May 2023" {
		t.Fatalf("should not double-annotate absolute dates: %q", same)
	}
}

func TestParseFlexibleTimeLocomoSessionStamp(t *testing.T) {
	ts := parseFlexibleTime("1:56 pm on 8 May, 2023")
	if ts == nil {
		t.Fatal("expected LOCOMO session stamp to parse")
	}
	if ts.Format("2006-01-02") != "2023-05-08" {
		t.Fatalf("got %s", ts.Format("2006-01-02"))
	}
	record, err := BuildMemoryRecord("mem_locomo", mustParseTime(t, "2026-07-19T00:00:00Z"), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Metadata:   map[string]any{"observed_at": "1:56 pm on 8 May, 2023"},
	}, ExtractedMemory{
		Kind:       KindFact,
		Content:    "Caroline: I went to a LGBTQ support group yesterday",
		SourceText: "Caroline: I went to a LGBTQ support group yesterday",
		Confidence: 0.7,
		Explain:    map[string]any{"rule": "conversation_episode", "primitive": PrimitiveEpisode},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if record.ObservedAt == nil || record.ObservedAt.Format("2006-01-02") != "2023-05-08" {
		t.Fatalf("expected observed_at, got %#v", record.ObservedAt)
	}
	if !strings.Contains(record.Content, "7 May 2023") {
		t.Fatalf("expected yesterday resolved against session stamp, got %q", record.Content)
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

func TestProviderExtractorReadsJSONFromReasoningWhenContentNull(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content":       nil,
					"reasoning":     `analysis...\n{"memories":[{"kind":"fact","content":"Caroline has a cat named Whiskers","source_text":"I have a cat named Whiskers","confidence":0.9}]}`,
					"finish_reason": "stop",
				}},
			},
		})
	}))
	defer server.Close()

	extractor := NewProviderExtractor(ProviderConfig{
		BaseURL: server.URL,
		Model:   "test-model",
		Strict:  true,
	}, server.Client())
	memories, err := extractor.Extract(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []Message{{Role: "user", Content: "Caroline: I have a cat named Whiskers."}},
	})
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	found := false
	for _, mem := range memories {
		if strings.Contains(mem.Content, "Whiskers") && mem.Explain["rule"] == "provider_extract" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected provider memory from reasoning JSON, got %#v", memories)
	}
}

func TestProviderExtractorReadsArrayContentParts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"content": []map[string]any{
						{"type": "output_text", "text": `{"memories":[{"kind":"fact","content":"Jordan is a nurse","source_text":"I work as a nurse","confidence":0.9,"predicate":"occupation","value":"nurse"}]}`},
					},
				}},
			},
		})
	}))
	defer server.Close()

	extractor := NewProviderExtractor(ProviderConfig{
		BaseURL: server.URL,
		Model:   "test-model",
		Strict:  true,
	}, server.Client())
	memories, err := extractor.Extract(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []Message{{Role: "user", Content: "Jordan: I work as a nurse"}},
	})
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	found := false
	for _, mem := range memories {
		if strings.Contains(strings.ToLower(mem.Content), "nurse") && mem.Explain["rule"] == "provider_extract" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected provider memory from content parts, got %#v", memories)
	}
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

func TestProviderExtractorStrictDoesNotBaselineSubstitute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusGatewayTimeout)
	}))
	defer server.Close()

	extractor := NewProviderExtractor(ProviderConfig{
		BaseURL: server.URL,
		Model:   "test-model",
		Strict:  true,
	}, server.Client())
	_, err := extractor.Extract(context.Background(), IngestRequest{
		TenantID:   "t1",
		SubjectID:  "u1",
		SourceType: "conversation",
		Messages:   []Message{{Role: "user", Content: "I prefer concise answers."}},
	})
	if err == nil {
		t.Fatal("expected strict provider error")
	}
	if extractor.Stats().Fallbacks != 0 {
		t.Fatalf("strict must not fallback, got %+v", extractor.Stats())
	}
}

func TestEnrichLastSaturdayAndYearsAgo(t *testing.T) {
	// Session on Thursday 25 May 2023 → last Saturday = 20 May 2023
	at := mustParseTime(t, "2023-05-25T12:00:00Z")
	got := EnrichRelativeEventTime("I ran a charity race last Saturday", at)
	if !strings.Contains(got, "20 May 2023") {
		t.Fatalf("expected last Saturday resolved, got %q", got)
	}
	got = EnrichRelativeEventTime("A friend made it for my 18th birthday ten years ago", at)
	if !strings.Contains(got, "10 years ago") || !strings.Contains(got, "2013") {
		t.Fatalf("expected ten years ago absolute, got %q", got)
	}
}
