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
}

func Load() Config {
	return Config{
		Environment:        getenv("BRAINY_ENV", "development"),
		HTTPAddr:           getenv("BRAINY_HTTP_ADDR", ":8080"),
		DatabaseURL:        getenv("BRAINY_DATABASE_URL", "postgres://brainy:brainy@localhost:5432/brainy?sslmode=disable"),
		WorkerMode:         getenv("BRAINY_WORKER_MODE", "once"),
		WorkerPollInterval: getenvDuration("BRAINY_WORKER_POLL_INTERVAL", 2*time.Second),
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
