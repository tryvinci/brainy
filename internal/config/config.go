package config

import "os"

type Config struct {
	Environment string
	HTTPAddr    string
	DatabaseURL string
}

func Load() Config {
	return Config{
		Environment: getenv("BRAINY_ENV", "development"),
		HTTPAddr:    getenv("BRAINY_HTTP_ADDR", ":8080"),
		DatabaseURL: getenv("BRAINY_DATABASE_URL", "postgres://brainy:brainy@localhost:5432/brainy?sslmode=disable"),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
