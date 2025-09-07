// Package main provides the main entry point for the Brokle API server.
//
// This is the single monolith server that handles:
// - HTTP API endpoints
// - WebSocket real-time connections
// - Background job processing
// - Multi-database operations (PostgreSQL + ClickHouse)
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "brokle/docs/swagger" // swagger docs
	"brokle/internal/app"
	"brokle/internal/config"
)

// @title Brokle AI Infrastructure Platform API
// @version 1.0.0
// @description Complete AI infrastructure platform providing gateway, observability, caching, and optimization services.
// @termsOfService https://brokle.ai/terms
//
// @contact.name Brokle Platform Team
// @contact.url https://brokle.ai/support
// @contact.email support@brokle.ai
//
// @license.name MIT License
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @schemes http https
//
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API key for authentication. Format: X-API-Key: bk_live_...
//
// Custom type definitions for Swagger
// @x-extension-openapi {"definitions": {"ULID": {"type": "string", "description": "ULID (Universally Unique Lexicographically Sortable Identifier)", "example": "01ARZ3NDEKTSV4RRFFQ69G5FAV", "pattern": "^[0-9A-Z]{26}$"}}}
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT token for authentication. Format: Authorization: Bearer <token>
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize application with all dependencies
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer application.Shutdown(context.Background())

	// Start the application (HTTP server + WebSocket + background workers)
	if err := application.Run(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server stopped")
}
