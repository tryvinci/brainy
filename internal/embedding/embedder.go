package embedding

import "context"

// Embedder produces dense vectors for hybrid retrieval.
// CI / default path uses LocalEmbedder (deterministic hash bag).
// Production may use ProviderEmbedder (OpenAI-compatible /v1/embeddings).
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	Name() string
}

// LocalEmbedder wraps the deterministic hash embedder.
type LocalEmbedder struct{}

func NewLocalEmbedder() LocalEmbedder {
	return LocalEmbedder{}
}

func (LocalEmbedder) Name() string { return "local-hash-v1" }

func (LocalEmbedder) Identity() Identity {
	return Identity{
		Name:       "local-hash-v1",
		Provider:   "local-hash",
		Model:      "local-hash-v1",
		Dimensions: Dim,
		Version:    "local-hash-v1@128",
	}
}

func (LocalEmbedder) Stats() Stats { return Stats{} }

func (LocalEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	return Embed(text), nil
}

// Default returns the CI-safe local embedder.
func Default() Embedder {
	return NewLocalEmbedder()
}
