// Package main provides the main entry point for the Brokle worker process.
//
// This is the background worker that handles:
// - Telemetry stream processing from Redis
// - Gateway analytics aggregation
// - Batch job processing
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"brokle/internal/app"
	"brokle/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Workers do NOT run migrations (server owns this)

	// Initialize worker application (workers only, no HTTP)
	worker, err := app.NewWorker(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize worker: %v", err)
	}
	defer worker.Shutdown(context.Background())

	// Start workers
	if err := worker.Start(); err != nil {
		log.Fatalf("Failed to start workers: %v", err)
	}

	log.Println("Workers started successfully")

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down workers...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := worker.Shutdown(ctx); err != nil {
		log.Printf("Workers forced to shutdown: %v", err)
	}

	fmt.Println("Workers stopped")
}
