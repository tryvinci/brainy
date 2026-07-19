package config

import (
	"log/slog"

	"brainy/internal/embedding"
)

// BuildEmbedder returns a provider embedder when EmbeddingModel is set,
// otherwise the deterministic local hash embedder (CI default).
func BuildEmbedder(cfg Config, logger *slog.Logger) embedding.Embedder {
	providerCfg := embedding.ProviderConfig{
		BaseURL: cfg.EmbeddingBaseURL,
		APIKey:  cfg.EmbeddingAPIKey,
		Model:   cfg.EmbeddingModel,
		Timeout: cfg.EmbeddingTimeout,
	}
	if !providerCfg.Configured() {
		if logger != nil {
			logger.Info("using local hash embedder")
		}
		return embedding.Default()
	}
	if logger != nil {
		logger.Info("using provider embedder", "base_url", providerCfg.BaseURL, "model", providerCfg.Model)
	}
	return embedding.NewProviderEmbedder(providerCfg, nil)
}
