package embedding

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProviderEmbedderParsesVectors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if _, ok := body["dimensions"]; ok {
			t.Fatal("did not expect dimensions for non-embedding-3 model")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"embedding": []float64{0.1, 0.2, 0.3}},
			},
		})
	}))
	defer server.Close()

	embedder := NewProviderEmbedder(ProviderConfig{
		BaseURL:    server.URL,
		APIKey:     "test",
		Model:      "test-embed",
		Dimensions: 768,
	}, server.Client())

	vec, err := embedder.Embed(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if len(vec) != 3 || vec[0] != 0.1 || vec[2] != 0.3 {
		t.Fatalf("unexpected vector %#v", vec)
	}
	if embedder.Stats().Calls != 1 || embedder.Stats().Fallbacks != 0 {
		t.Fatalf("unexpected stats %+v", embedder.Stats())
	}
}

func TestProviderEmbedderSendsDimensionsForEmbedding3(t *testing.T) {
	seen := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["dimensions"] != float64(768) {
			t.Fatalf("expected dimensions=768, got %#v", body["dimensions"])
		}
		seen = true
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"embedding": make([]float64, 768)},
			},
		})
	}))
	defer server.Close()

	embedder := NewProviderEmbedder(ProviderConfig{
		BaseURL:    server.URL,
		Model:      "text-embedding-3-large",
		Dimensions: 768,
	}, server.Client())
	vec, err := embedder.Embed(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if !seen || len(vec) != 768 {
		t.Fatalf("seen=%v len=%d", seen, len(vec))
	}
}

func TestProviderEmbedderSoftDegradesOnFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	embedder := NewProviderEmbedder(ProviderConfig{
		BaseURL: server.URL,
		Model:   "test-embed",
	}, server.Client())

	vec, err := embedder.Embed(context.Background(), "Brand voice is warm and concise")
	if err != nil {
		t.Fatalf("expected soft-degrade, got %v", err)
	}
	if len(vec) != Dim {
		t.Fatalf("expected local dim %d, got %d", Dim, len(vec))
	}
	local, _ := NewLocalEmbedder().Embed(context.Background(), "Brand voice is warm and concise")
	if CosineSimilarity(vec, local) != 1 {
		t.Fatal("expected fallback to match local embedder")
	}
	if embedder.Stats().Fallbacks != 1 || embedder.Stats().Failures != 1 {
		t.Fatalf("expected failure+fallback counters, got %+v", embedder.Stats())
	}
}

func TestProviderEmbedderStrictDoesNotHashFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	embedder := NewProviderEmbedder(ProviderConfig{
		BaseURL: server.URL,
		Model:   "test-embed",
		Strict:  true,
	}, server.Client())
	_, err := embedder.Embed(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected strict error")
	}
	if embedder.Stats().Fallbacks != 0 {
		t.Fatalf("strict must not fallback, got %+v", embedder.Stats())
	}
}

func TestSupportsDimensions(t *testing.T) {
	if !SupportsDimensions("text-embedding-3-large") {
		t.Fatal("expected large to support dimensions")
	}
	if SupportsDimensions("hosted-bge-base") {
		t.Fatal("BGE must not receive dimensions")
	}
}
