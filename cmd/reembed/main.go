package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"sync"
	"time"

	"brainy/internal/config"
	"brainy/internal/embedding"
	"brainy/internal/observability"
	"brainy/internal/store/postgres"
)

func main() {
	var (
		limit     = flag.Int("limit", 0, "max memories to re-embed (0 = all)")
		offset    = flag.Int("offset", 0, "skip this many memories")
		batchSize = flag.Int("batch", 200, "page size")
		workers   = flag.Int("workers", 0, "concurrent embedders (default BRAINY_WORKER_CONCURRENCY)")
	)
	flag.Parse()

	cfg := config.Load()
	logger := observability.NewLogger()
	if *workers <= 0 {
		*workers = cfg.WorkerConcurrency
		if *workers <= 0 {
			*workers = 4
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	defer cancel()

	store, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("store", "error", err)
		os.Exit(1)
	}
	defer store.Close()
	if err := store.ApplyMigrations(ctx); err != nil {
		logger.Error("migrate", "error", err)
		os.Exit(1)
	}

	embedder := config.BuildEmbedder(cfg, logger)
	ident := embedding.IdentityOf(embedder)
	logger.Info("reembed start", "embedder", ident.Name, "model", ident.Model, "dimensions", ident.Dimensions)

	sem := make(chan struct{}, *workers)
	var wg sync.WaitGroup
	var mu sync.Mutex
	done := 0
	failed := 0
	pageOff := *offset
	remaining := *limit

	for {
		n := *batchSize
		if remaining > 0 && remaining < n {
			n = remaining
		}
		targets, err := store.ListEmbeddingTargets(ctx, n, pageOff)
		if err != nil {
			logger.Error("list", "error", err)
			os.Exit(1)
		}
		if len(targets) == 0 {
			break
		}
		for _, target := range targets {
			target := target
			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				values, err := embedder.Embed(ctx, target.Content)
				if err != nil {
					mu.Lock()
					failed++
					mu.Unlock()
					logger.Error("embed", "memory_id", target.MemoryID, "error", err)
					return
				}
				rec := embedding.RecordFromEmbedder(embedder, target.MemoryID, target.TenantID, target.SubjectID, values)
				if err := store.WriteEmbedding(ctx, rec); err != nil {
					mu.Lock()
					failed++
					mu.Unlock()
					logger.Error("write", "memory_id", target.MemoryID, "error", err)
					return
				}
				mu.Lock()
				done++
				mu.Unlock()
			}()
		}
		wg.Wait()
		pageOff += len(targets)
		if remaining > 0 {
			remaining -= len(targets)
			if remaining <= 0 {
				break
			}
		}
		if len(targets) < n {
			break
		}
		logger.Info("reembed progress", "done", done, "failed", failed, "offset", pageOff)
	}
	logger.Info("reembed complete", "done", done, "failed", failed)
	payload, _ := json.Marshal(map[string]any{
		"embedder":           ident,
		"embedder_signature": ident.Signature(),
		"embedder_stats":     embedding.StatsOf(embedder),
		"source":             "cmd/reembed",
	})
	if err := store.UpsertProviderRuntime(ctx, "worker", payload); err != nil {
		logger.Error("runtime", "error", err)
	}
	if failed > 0 {
		os.Exit(1)
	}
}
