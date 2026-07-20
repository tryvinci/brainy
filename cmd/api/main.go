package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"brainy/internal/api"
	"brainy/internal/auth"
	"brainy/internal/config"
	"brainy/internal/memory"
	"brainy/internal/observability"
	"brainy/internal/store/postgres"
)

func main() {
	cfg := config.Load()
	var logger *slog.Logger = observability.NewLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to create store", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	if err := store.ApplyMigrations(ctx); err != nil {
		logger.Error("failed to apply migrations", "error", err)
		os.Exit(1)
	}

	metrics := observability.NewMetrics()
	service := memory.NewService(store).
		WithEmbedder(config.BuildEmbedder(cfg, logger)).
		WithEntityRanking(cfg.EntityRanking)
	keyRing := auth.ParseKeyRing(cfg.APIKeys)
	router := api.NewRouter(service, metrics)
	router = api.APIKeyMiddleware(keyRing, cfg.RequireAPIKey)(router)
	router = observability.TraceIDMiddleware(router)
	router = observability.LoggingMiddleware(logger)(router)

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	logger.Info("brainy api listening", "addr", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
