package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"brainy/internal/config"
	"brainy/internal/jobs"
	"brainy/internal/memory"
	"brainy/internal/observability"
	"brainy/internal/store/postgres"
)

func main() {
	cfg := config.Load()
	logger := observability.NewLogger()
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
	extractor := buildWorkerExtractor(cfg, logger)
	processor := jobs.NewProcessorWithExtractor(store, metrics, extractor).WithEmbedder(config.BuildEmbedder(cfg, logger))
	switch cfg.WorkerMode {
	case "loop":
		runLoop(processor, cfg.WorkerPollInterval, logger)
	default:
		processed, err := processor.ProcessNext(context.Background())
		if err != nil {
			logger.Error("processing failed", "error", err)
			os.Exit(1)
		}
		if !processed {
			logger.Info("brainy worker booted", "mode", cfg.Environment, "database", cfg.DatabaseURL, "pending_jobs", 0)
			return
		}
		logger.Info("brainy worker processed one pending extraction job")
	}
}

func buildWorkerExtractor(cfg config.Config, logger *slog.Logger) memory.Extractor {
	providerCfg := memory.ProviderConfig{
		BaseURL: cfg.ProviderBaseURL,
		APIKey:  cfg.ProviderAPIKey,
		Model:   cfg.ProviderModel,
		Timeout: cfg.ProviderTimeout,
	}
	if !providerCfg.Configured() {
		logger.Info("worker using deterministic extractor")
		return memory.NewDeterministicExtractor()
	}
	logger.Info("worker using provider extractor", "base_url", providerCfg.BaseURL, "model", providerCfg.Model)
	return memory.NewProviderExtractor(providerCfg, nil)
}

func runLoop(processor *jobs.Processor, interval time.Duration, logger *slog.Logger) {
	logger.Info("brainy worker entering loop mode", "poll_interval", interval.String())
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		processed, err := processor.ProcessNext(ctx)
		if err != nil {
			// Job already failed/requeued inside ProcessNext. Keep the loop
			// alive so a single provider flake cannot stall the queue.
			logger.Error("processing failed", "error", err)
			select {
			case <-ctx.Done():
				logger.Info("brainy worker shutting down")
				return
			case <-ticker.C:
			}
			continue
		}
		if processed {
			continue
		}

		select {
		case <-ctx.Done():
			logger.Info("brainy worker shutting down")
			return
		case <-ticker.C:
		}
	}
}
