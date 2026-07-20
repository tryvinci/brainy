package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Environment        string
	HTTPAddr           string
	DatabaseURL        string
	WorkerMode         string
	WorkerPollInterval time.Duration
	APIKeys            string
	RequireAPIKey      bool
	// Provider extract (async worker). Empty BaseURL/Model => deterministic only.
	ProviderBaseURL string
	ProviderAPIKey  string
	ProviderModel   string
	ProviderTimeout time.Duration
	// Provider embeddings (OpenAI-compatible). Empty Model => local hash only.
	EmbeddingBaseURL string
	EmbeddingAPIKey  string
	EmbeddingModel   string
	EmbeddingTimeout time.Duration
	// EntityRanking enables experimental entity-graph retrieval reranking.
	EntityRanking bool
}

func Load() Config {
	env := getenv("BRAINY_ENV", "local")
	requireKey := getenv("BRAINY_REQUIRE_API_KEY", "") == "true"
	if !requireKey && env == "production" {
		requireKey = true
	}
	providerBase := getenv("BRAINY_PROVIDER_BASE_URL", os.Getenv("LLM_BASE_URL"))
	providerKey := getenv("BRAINY_PROVIDER_API_KEY", os.Getenv("LLM_API_KEY"))
	return Config{
		Environment:        env,
		HTTPAddr:           getenv("BRAINY_HTTP_ADDR", ":8080"),
		DatabaseURL:        getenv("BRAINY_DATABASE_URL", "postgres://brainy:brainy@localhost:5432/brainy?sslmode=disable"),
		WorkerMode:         getenv("BRAINY_WORKER_MODE", "once"),
		WorkerPollInterval: getenvDuration("BRAINY_WORKER_POLL_INTERVAL", 2*time.Second),
		APIKeys:            os.Getenv("BRAINY_API_KEYS"),
		RequireAPIKey:      requireKey,
		ProviderBaseURL:    providerBase,
		ProviderAPIKey:     providerKey,
		ProviderModel:      getenv("BRAINY_PROVIDER_MODEL", os.Getenv("LLM_MODEL")),
		ProviderTimeout:    getenvDuration("BRAINY_PROVIDER_TIMEOUT", 45*time.Second),
		EmbeddingBaseURL:   getenv("BRAINY_EMBEDDING_BASE_URL", providerBase),
		EmbeddingAPIKey:    getenv("BRAINY_EMBEDDING_API_KEY", providerKey),
		EmbeddingModel:     getenv("BRAINY_EMBEDDING_MODEL", os.Getenv("EMBEDDING_MODEL")),
		EmbeddingTimeout:   getenvDuration("BRAINY_EMBEDDING_TIMEOUT", 30*time.Second),
		EntityRanking:      entityRankingDefault(),
	}
}

// entityRankingDefault: entity/graph reranking only helps with dense semantic
// embeddings (proven by same-pin A/B — see docs/benchmarks/entity-linking-ab.md).
// So default it ON only when a provider embedding model is configured; allow an
// explicit BRAINY_ENTITY_RANKING override either way.
func entityRankingDefault() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("BRAINY_ENTITY_RANKING"))) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	}
	embModel := getenv("BRAINY_EMBEDDING_MODEL", os.Getenv("EMBEDDING_MODEL"))
	return strings.TrimSpace(embModel) != ""
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
