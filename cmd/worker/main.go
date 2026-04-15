package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"brainy/internal/config"
	"brainy/internal/jobs"
	"brainy/internal/store/postgres"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	if err := store.ApplyMigrations(ctx); err != nil {
		log.Fatal(err)
	}

	processor := jobs.NewProcessor(store)
	switch cfg.WorkerMode {
	case "loop":
		runLoop(processor, cfg.WorkerPollInterval)
	default:
		processed, err := processor.ProcessNext(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		if !processed {
			log.Printf("brainy worker booted in %s mode with database target %s and found no pending jobs", cfg.Environment, cfg.DatabaseURL)
			return
		}
		log.Printf("brainy worker processed one pending extraction job")
	}
}

func runLoop(processor *jobs.Processor, interval time.Duration) {
	log.Printf("brainy worker entering loop mode with poll interval %s", interval)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		processed, err := processor.ProcessNext(ctx)
		if err != nil {
			log.Fatal(err)
		}
		if processed {
			continue
		}

		select {
		case <-ctx.Done():
			log.Printf("brainy worker shutting down")
			return
		case <-ticker.C:
		}
	}
}
