// Package main provides database migration tool for both PostgreSQL and ClickHouse.
//
// Usage:
//   go run cmd/migrate/main.go up           # Run all pending migrations
//   go run cmd/migrate/main.go down         # Rollback one migration
//   go run cmd/migrate/main.go postgres up  # Run PostgreSQL migrations only
//   go run cmd/migrate/main.go clickhouse up # Run ClickHouse migrations only
//   go run cmd/migrate/main.go status       # Show migration status
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"brokle/internal/config"
	"brokle/internal/migration"
)

func main() {
	var (
		database  = flag.String("db", "all", "Database to migrate: all, postgres, clickhouse")
		direction = flag.String("dir", "up", "Migration direction: up, down")
		steps     = flag.Int("steps", 0, "Number of migration steps (0 = all)")
		dryRun    = flag.Bool("dry-run", false, "Show what would be migrated without executing")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize migration manager
	manager, err := migration.NewManager(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize migration manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()

	// Handle different commands
	args := flag.Args()
	if len(args) > 0 {
		switch args[0] {
		case "status":
			if err := showStatus(ctx, manager, *database); err != nil {
				log.Fatalf("Failed to show status: %v", err)
			}
			return
		case "create":
			if len(args) < 2 {
				log.Fatal("Migration name is required for create command")
			}
			if err := createMigration(manager, *database, args[1]); err != nil {
				log.Fatalf("Failed to create migration: %v", err)
			}
			return
		}
	}

	// Run migrations
	if err := runMigrations(ctx, manager, *database, *direction, *steps, *dryRun); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migrations completed successfully")
}

func runMigrations(ctx context.Context, manager *migration.Manager, database, direction string, steps int, dryRun bool) error {
	switch database {
	case "postgres":
		if direction == "up" {
			return manager.MigratePostgresUp(ctx, steps, dryRun)
		}
		return manager.MigratePostgresDown(ctx, steps, dryRun)
	case "clickhouse":
		if direction == "up" {
			return manager.MigrateClickHouseUp(ctx, steps, dryRun)
		}
		return manager.MigrateClickHouseDown(ctx, steps, dryRun)
	case "all":
		if direction == "up" {
			if err := manager.MigratePostgresUp(ctx, steps, dryRun); err != nil {
				return fmt.Errorf("postgres migration failed: %w", err)
			}
			return manager.MigrateClickHouseUp(ctx, steps, dryRun)
		}
		if err := manager.MigrateClickHouseDown(ctx, steps, dryRun); err != nil {
			return fmt.Errorf("clickhouse migration failed: %w", err)
		}
		return manager.MigratePostgresDown(ctx, steps, dryRun)
	default:
		return fmt.Errorf("unknown database: %s", database)
	}
}

func showStatus(ctx context.Context, manager *migration.Manager, database string) error {
	switch database {
	case "postgres":
		return manager.ShowPostgresStatus(ctx)
	case "clickhouse":
		return manager.ShowClickHouseStatus(ctx)
	case "all":
		fmt.Println("PostgreSQL Migration Status:")
		if err := manager.ShowPostgresStatus(ctx); err != nil {
			return err
		}
		fmt.Println("\nClickHouse Migration Status:")
		return manager.ShowClickHouseStatus(ctx)
	default:
		return fmt.Errorf("unknown database: %s", database)
	}
}

func createMigration(manager *migration.Manager, database, name string) error {
	switch database {
	case "postgres":
		return manager.CreatePostgresMigration(name)
	case "clickhouse":
		return manager.CreateClickHouseMigration(name)
	default:
		return fmt.Errorf("unknown database: %s", database)
	}
}