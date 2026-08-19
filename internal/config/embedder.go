package config

import (
	"fmt"
	"log/slog"
	"strings"

	"brainy/internal/embedding"
)

// RequireStrictProviders fails boot when a strict flag is set but the
// corresponding provider is not configured. Product default (flags off)
// still soft-degrades.
func RequireStrictProviders(cfg Config) error {
	if cfg.EmbeddingStrict {
		if strings.TrimSpace(cfg.EmbeddingBaseURL) == "" || strings.TrimSpace(cfg.EmbeddingModel) == "" {
			return fmt.Errorf("BRAINY_EMBEDDING_STRICT requires BRAINY_EMBEDDING_BASE_URL and BRAINY_EMBEDDING_MODEL")
		}
	}
	if cfg.ExtractionStrict {
		if strings.TrimSpace(cfg.ProviderBaseURL) == "" || strings.TrimSpace(cfg.ProviderModel) == "" {
			return fmt.Errorf("BRAINY_EXTRACTION_STRICT requires BRAINY_PROVIDER_BASE_URL and BRAINY_PROVIDER_MODEL")
		}
	}
	return nil
}

// BuildEmbedder returns a provider embedder when EmbeddingModel is set,
// otherwise the deterministic local hash embedder (CI default).
func BuildEmbedder(cfg Config, logger *slog.Logger) embedding.Embedder {
	providerCfg := embedding.ProviderConfig{
		BaseURL:    cfg.EmbeddingBaseURL,
		APIKey:     cfg.EmbeddingAPIKey,
		Model:      cfg.EmbeddingModel,
		Timeout:    cfg.EmbeddingTimeout,
		Dimensions: cfg.EmbeddingDimensions,
		Strict:     cfg.EmbeddingStrict,
	}
	if !providerCfg.Configured() {
		if logger != nil {
			logger.Info("using local hash embedder")
		}
		return embedding.Default()
	}
	if logger != nil {
		logger.Info("using provider embedder", "base_url", providerCfg.BaseURL, "model", providerCfg.Model, "dimensions", providerCfg.Dimensions, "strict", providerCfg.Strict)
	}
	return embedding.NewProviderEmbedder(providerCfg, nil)
}
