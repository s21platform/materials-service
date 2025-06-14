package main

import (
	"log"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/s21platform/materials-service/internal/config"
)

func main() {
	cfg := config.MustLoad()

	log.Printf("Configuration loaded:")
	log.Printf("Service Port: %s", cfg.Service.Port)
	log.Printf("Metrics Host: %s", cfg.Metrics.Host)
	log.Printf("Metrics Port: %d", cfg.Metrics.Port)
	log.Printf("Platform Env: %s", cfg.Platform.Env)
}
