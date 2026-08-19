package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment string
	HTTPAddr    string
	DatabaseURL string
	// MaxBodyBytes caps the size of request bodies read by the API (ingest,
	// recall, events, correct, supersede). 0 disables the limit.
	MaxBodyBytes int64
	// HTTP server timeouts (0 disables the corresponding timeout).
	HTTPReadHeaderTimeout time.Duration
	HTTPReadTimeout       time.Duration
	HTTPWriteTimeout      time.Duration
	HTTPIdleTimeout       time.Duration
	WorkerMode            string
	WorkerPollInterval    time.Duration
	WorkerConcurrency     int
	APIKeys               string
	RequireAPIKey         bool
	// Provider extract (async worker). Empty BaseURL/Model => deterministic only.
	ProviderBaseURL  string
	ProviderAPIKey   string
	ProviderModel    string
	ProviderTimeout  time.Duration
	ExtractionStrict bool
	// Provider embeddings (OpenAI-compatible). Empty Model => local hash only.
	EmbeddingBaseURL    string
	EmbeddingAPIKey     string
	EmbeddingModel      string
	EmbeddingTimeout    time.Duration
	EmbeddingDimensions int
	EmbeddingStrict     bool
	// RequireANN fails boot when a hosted 768-d embedder is configured but
	// pgvector / embedding_vec_768 is missing. Disable with BRAINY_REQUIRE_ANN=false.
	RequireANN bool
	// EntityRanking enables experimental entity-graph retrieval reranking.
	EntityRanking bool
	// IDFRanking enables experimental IDF-weighted lexical coverage.
	IDFRanking bool
}

func Load() Config {
	env := getenv("BRAINY_ENV", "local")
	requireKey := getenv("BRAINY_REQUIRE_API_KEY", "") == "true"
	if !requireKey && env == "production" {
		requireKey = true
	}
	providerBase := getenv("BRAINY_PROVIDER_BASE_URL", os.Getenv("LLM_BASE_URL"))
	providerKey := getenv("BRAINY_PROVIDER_API_KEY", os.Getenv("LLM_API_KEY"))
	providerTimeout := getenvDuration("BRAINY_PROVIDER_TIMEOUT", 45*time.Second)
	// Write deadline must exceed the provider ceiling so slow hybrid-reader
	// recall is not cut off mid-answer; the env override still wins.
	writeTimeout := getenvDuration("BRAINY_HTTP_WRITE_TIMEOUT", providerTimeout+60*time.Second)
	cfg := Config{
		Environment:           env,
		HTTPAddr:              getenv("BRAINY_HTTP_ADDR", ":8080"),
		DatabaseURL:           getenv("BRAINY_DATABASE_URL", "postgres://brainy:brainy@localhost:5432/brainy?sslmode=disable"),
		MaxBodyBytes:          getenvInt64("BRAINY_MAX_BODY_BYTES", 5<<20),
		HTTPReadHeaderTimeout: getenvDuration("BRAINY_HTTP_READ_HEADER_TIMEOUT", 10*time.Second),
		HTTPReadTimeout:       getenvDuration("BRAINY_HTTP_READ_TIMEOUT", 30*time.Second),
		HTTPWriteTimeout:      writeTimeout,
		HTTPIdleTimeout:       getenvDuration("BRAINY_HTTP_IDLE_TIMEOUT", 120*time.Second),
		WorkerMode:            getenv("BRAINY_WORKER_MODE", "once"),
		WorkerPollInterval:    getenvDuration("BRAINY_WORKER_POLL_INTERVAL", 2*time.Second),
		WorkerConcurrency:     getenvInt("BRAINY_WORKER_CONCURRENCY", 1),
		APIKeys:               os.Getenv("BRAINY_API_KEYS"),
		RequireAPIKey:         requireKey,
		ProviderBaseURL:       providerBase,
		ProviderAPIKey:        providerKey,
		ProviderModel:         getenv("BRAINY_PROVIDER_MODEL", os.Getenv("LLM_MODEL")),
		ProviderTimeout:       providerTimeout,
		ExtractionStrict:      getenvBool("BRAINY_EXTRACTION_STRICT", false),
		EmbeddingBaseURL:      getenv("BRAINY_EMBEDDING_BASE_URL", providerBase),
		EmbeddingAPIKey:       getenv("BRAINY_EMBEDDING_API_KEY", providerKey),
		EmbeddingModel:        getenv("BRAINY_EMBEDDING_MODEL", os.Getenv("EMBEDDING_MODEL")),
		EmbeddingTimeout:      getenvDuration("BRAINY_EMBEDDING_TIMEOUT", 30*time.Second),
		EmbeddingDimensions:   getenvEmbeddingDimensions(),
		EmbeddingStrict:       getenvBool("BRAINY_EMBEDDING_STRICT", false),
		EntityRanking:         entityRankingDefault(),
		IDFRanking:            getenv("BRAINY_IDF_RANKING", "") == "true",
	}
	cfg.RequireANN = requireANN(cfg.EmbeddingModel, cfg.EmbeddingDimensions)
	return cfg
}

func getenvInt64(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n < 0 {
		return fallback
	}
	return n
}

// entityRankingDefault: entity/graph reranking stays OFF unless explicitly
// enabled. Same-pin A/B with real dense embeddings (CF Workers AI bge-base,
// 2026-07-23) still regresses LOCOMO smoke (11/30 vs 13/30 entity-off) — see
// docs/benchmarks/entity-linking-ab.md. Extraction/persistence remain always on;
// flip BRAINY_ENTITY_RANKING=true only after a staging re-tune shows lift.
func entityRankingDefault() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("BRAINY_ENTITY_RANKING"))) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
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

func getenvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	n := 0
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return fallback
		}
		n = n*10 + int(ch-'0')
	}
	if n <= 0 {
		return fallback
	}
	if n > 32 {
		return 32
	}
	return n
}

func getenvBool(key string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func getenvEmbeddingDimensions() int {
	value := strings.TrimSpace(os.Getenv("BRAINY_EMBEDDING_DIMENSIONS"))
	if value == "" {
		return 0
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func requireANN(embeddingModel string, dims int) bool {
	if strings.TrimSpace(os.Getenv("BRAINY_REQUIRE_ANN")) != "" {
		return getenvBool("BRAINY_REQUIRE_ANN", false)
	}
	if strings.TrimSpace(embeddingModel) == "" {
		return false
	}
	if dims > 0 && dims != 768 {
		return false
	}
	return true
}
