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
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"embedding": []float64{0.1, 0.2, 0.3}},
			},
		})
	}))
	defer server.Close()

	embedder := NewProviderEmbedder(ProviderConfig{
		BaseURL: server.URL,
		APIKey:  "test",
		Model:   "test-embed",
	}, server.Client())

	vec, err := embedder.Embed(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if len(vec) != 3 || vec[0] != 0.1 || vec[2] != 0.3 {
		t.Fatalf("unexpected vector %#v", vec)
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
}
