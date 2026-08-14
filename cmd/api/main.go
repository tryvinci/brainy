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
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	store, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to create store", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	if err := store.ApplyMigrations(ctx); err != nil {
		// Do not crash-loop the API on migration lock/timeout (staging hit this
		// when DROP/HNSW held the advisory lock past boot). Retry in background.
		logger.Error("failed to apply migrations at boot; retrying in background", "error", err)
		go func() {
			for attempt := 1; attempt <= 20; attempt++ {
				migCtx, migCancel := context.WithTimeout(context.Background(), 3*time.Minute)
				err := store.ApplyMigrations(migCtx)
				migCancel()
				if err == nil {
					logger.Info("background migrations applied", "attempt", attempt)
					indexCtx, indexCancel := context.WithTimeout(context.Background(), 30*time.Minute)
					store.EnsureEmbeddingVecIndex(indexCtx)
					indexCancel()
					return
				}
				logger.Error("background migration retry failed", "attempt", attempt, "error", err)
				time.Sleep(15 * time.Second)
			}
		}()
	} else {
		// Backfill + HNSW for hosted 768-d embeddings (additive column).
		go func() {
			indexCtx, indexCancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer indexCancel()
			store.EnsureEmbeddingVecIndex(indexCtx)
		}()
	}
	// Optional: build FTS GIN off the request path. Disabled by default on
	// starter plans — concurrent GIN builds can OOM small instances.
	if os.Getenv("BRAINY_ENSURE_FTS_INDEX") == "1" {
		go func() {
			indexCtx, indexCancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer indexCancel()
			store.EnsureContentFTSIndex(indexCtx)
		}()
	}

	metrics := observability.NewMetrics()
	service := memory.NewService(store).
		WithEmbedder(config.BuildEmbedder(cfg, logger)).
		WithEntityRanking(cfg.EntityRanking).
		WithIDFRanking(cfg.IDFRanking).
		WithHybridReader(memory.HybridReaderConfig{
			BaseURL: cfg.ProviderBaseURL,
			APIKey:  cfg.ProviderAPIKey,
			Model:   cfg.ProviderModel,
			Timeout: cfg.ProviderTimeout,
		})
	keyRing := auth.ParseKeyRing(cfg.APIKeys)
	router := api.NewRouter(service, metrics)
	router = api.APIKeyMiddleware(keyRing, cfg.RequireAPIKey)(router)
	router = observability.TraceIDMiddleware(router)
	router = observability.LoggingMiddleware(logger)(router)
	// Outermost so the body cap applies before auth's tenant_id scan reads it.
	router = api.MaxBytesMiddleware(cfg.MaxBodyBytes)(router)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		ReadTimeout:       cfg.HTTPReadTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}

	logger.Info("brainy api listening", "addr", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
