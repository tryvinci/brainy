package main

import (
	"context"
	"log"
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
