package config

import (
	"os"
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
}

func Load() Config {
	env := getenv("BRAINY_ENV", "development")
	requireKey := getenv("BRAINY_REQUIRE_API_KEY", "") == "true"
	if !requireKey && env == "production" {
		requireKey = true
	}
	return Config{
		Environment:        env,
		HTTPAddr:           getenv("BRAINY_HTTP_ADDR", ":8080"),
		DatabaseURL:        getenv("BRAINY_DATABASE_URL", "postgres://brainy:brainy@localhost:5432/brainy?sslmode=disable"),
		WorkerMode:         getenv("BRAINY_WORKER_MODE", "once"),
		WorkerPollInterval: getenvDuration("BRAINY_WORKER_POLL_INTERVAL", 2*time.Second),
		APIKeys:            os.Getenv("BRAINY_API_KEYS"),
		RequireAPIKey:      requireKey,
		ProviderBaseURL:    getenv("BRAINY_PROVIDER_BASE_URL", os.Getenv("LLM_BASE_URL")),
		ProviderAPIKey:     getenv("BRAINY_PROVIDER_API_KEY", os.Getenv("LLM_API_KEY")),
		ProviderModel:      getenv("BRAINY_PROVIDER_MODEL", os.Getenv("LLM_MODEL")),
		ProviderTimeout:    getenvDuration("BRAINY_PROVIDER_TIMEOUT", 45*time.Second),
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
