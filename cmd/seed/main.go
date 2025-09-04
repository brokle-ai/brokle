// Package main provides database seeding tool for both PostgreSQL and ClickHouse.
//
// Usage:
//   go run cmd/seed/main.go --env=development   # Seed with development data
//   go run cmd/seed/main.go --env=production    # Seed with production data
//   go run cmd/seed/main.go --db=postgres       # Seed PostgreSQL only
//   go run cmd/seed/main.go --db=clickhouse     # Seed ClickHouse only
//   go run cmd/seed/main.go --reset             # Reset all data before seeding
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"brokle/internal/config"
	"brokle/internal/seeder"
)

func main() {
	var (
		environment = flag.String("env", "development", "Environment: development, production")
		database    = flag.String("db", "all", "Database to seed: all, postgres, clickhouse")
		reset       = flag.Bool("reset", false, "Reset all data before seeding")
		dryRun      = flag.Bool("dry-run", false, "Show what would be seeded without executing")
		verbose     = flag.Bool("verbose", false, "Verbose output")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize seeder manager
	manager, err := seeder.NewManager(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize seeder manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()

	// Configure seeder options
	options := &seeder.Options{
		Environment: *environment,
		Reset:       *reset,
		DryRun:      *dryRun,
		Verbose:     *verbose,
	}

	// Run seeding
	if err := runSeeding(ctx, manager, *database, options); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	fmt.Println("Seeding completed successfully")
}

func runSeeding(ctx context.Context, manager *seeder.Manager, database string, options *seeder.Options) error {
	switch database {
	case "postgres":
		return manager.SeedPostgres(ctx, options)
	case "clickhouse":
		return manager.SeedClickHouse(ctx, options)
	case "all":
		if err := manager.SeedPostgres(ctx, options); err != nil {
			return fmt.Errorf("postgres seeding failed: %w", err)
		}
		return manager.SeedClickHouse(ctx, options)
	default:
		return fmt.Errorf("unknown database: %s", database)
	}
}