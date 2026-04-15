package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"brainy/internal/api"
	"brainy/internal/config"
	"brainy/internal/memory"
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

	service := memory.NewService(store)
	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: api.NewRouter(service),
	}

	log.Printf("brainy api listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
