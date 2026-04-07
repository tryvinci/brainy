package main

import (
	"log"
	"os"

	"brainy/internal/config"
)

func main() {
	cfg := config.Load()
	log.Printf("brainy worker booted in %s mode with database target %s", cfg.Environment, cfg.DatabaseURL)
	_, _ = os.Stdout.WriteString("worker idle\n")
}
