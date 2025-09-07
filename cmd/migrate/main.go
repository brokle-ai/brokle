// Package main provides database migration tool for both PostgreSQL and ClickHouse.
//
// This is a production-ready migration CLI with comprehensive safety features,
// interactive confirmations for destructive operations, and full support for
// both PostgreSQL and ClickHouse databases.
//
// Usage Examples:
//   go run cmd/migrate/main.go up                    # Run all pending migrations
//   go run cmd/migrate/main.go down                  # Rollback migrations (with confirmation)
//   go run cmd/migrate/main.go -db postgres up       # Run PostgreSQL migrations only
//   go run cmd/migrate/main.go -db clickhouse up     # Run ClickHouse migrations only
//   go run cmd/migrate/main.go status                # Show migration status
//   go run cmd/migrate/main.go goto -version 5       # Migrate to specific version (with confirmation)
//   go run cmd/migrate/main.go force -version 3      # Force version (with confirmation)
//   go run cmd/migrate/main.go drop                  # Drop all tables (with confirmation)
//   go run cmd/migrate/main.go steps -steps 2        # Run 2 steps forward
//   go run cmd/migrate/main.go steps -steps -1       # Run 1 step backward
//   go run cmd/migrate/main.go info                  # Show detailed migration information
//   go run cmd/migrate/main.go create -name "add_users" -db postgres  # Create new migration
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"brokle/internal/config"
	"brokle/internal/migration"
	"brokle/internal/seeder"
)

func main() {
	var (
		database    = flag.String("db", "all", "Database to migrate: all, postgres, clickhouse")
		steps       = flag.Int("steps", 0, "Number of migration steps (0 = all)")
		version     = flag.Int("version", 0, "Target version for goto/force commands")
		name        = flag.String("name", "", "Migration name for create command")
		dryRun      = flag.Bool("dry-run", false, "Show what would be migrated without executing")
		environment = flag.String("env", "development", "Environment for seeding: development, demo, test")
		reset       = flag.Bool("reset", false, "Reset existing data before seeding")
		verbose     = flag.Bool("verbose", false, "Verbose output for seeding")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Parse database selection
	databases, err := parseDatabaseSelection(*database)
	if err != nil {
		log.Fatalf("Invalid database selection: %v", err)
	}

	// Initialize migration manager with selected databases
	manager, err := migration.NewManagerWithDatabases(cfg, databases)
	if err != nil {
		log.Fatalf("Failed to initialize migration manager: %v", err)
	}
	defer manager.Shutdown()

	ctx := context.Background()

	// Handle different commands
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		return
	}

	command := args[0]
	switch command {
	case "up":
		if err := runMigrations(ctx, manager, *database, "up", *steps, *dryRun); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("✅ Migrations completed successfully")

	case "down":
		if !confirmDestructiveOperation("rollback migrations") {
			fmt.Println("Operation cancelled")
			return
		}
		if err := runMigrations(ctx, manager, *database, "down", *steps, *dryRun); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("✅ Rollback completed successfully")

	case "status":
		if err := showStatus(ctx, manager, *database); err != nil {
			log.Fatalf("Failed to show status: %v", err)
		}

	case "goto":
		if *version == 0 {
			log.Fatal("Version must be specified for goto command (use -version flag)")
		}
		if !confirmDestructiveOperation(fmt.Sprintf("migrate to version %d", *version)) {
			fmt.Println("Operation cancelled")
			return
		}
		if err := gotoVersion(manager, *database, uint(*version)); err != nil {
			log.Fatalf("Failed to migrate to version %d: %v", *version, err)
		}
		fmt.Printf("✅ Migrated to version %d successfully\n", *version)

	case "force":
		if *version == 0 {
			log.Fatal("Version must be specified for force command (use -version flag)")
		}
		if !confirmDestructiveOperation(fmt.Sprintf("FORCE migration to version %d (DANGEROUS)", *version)) {
			fmt.Println("Operation cancelled")
			return
		}
		if err := forceVersion(manager, *database, *version); err != nil {
			log.Fatalf("Failed to force migration to version %d: %v", *version, err)
		}
		fmt.Printf("⚠️  Forced migration to version %d successfully\n", *version)

	case "drop":
		if !confirmDestructiveOperation("DROP ALL TABLES (PERMANENT DATA LOSS)") {
			fmt.Println("Operation cancelled")
			return
		}
		if err := dropTables(manager, *database); err != nil {
			log.Fatalf("Failed to drop tables: %v", err)
		}
		fmt.Println("⚠️  Tables dropped successfully")

	case "steps":
		if *steps == 0 {
			log.Fatal("Steps must be specified for steps command (use -steps flag)")
		}
		if *steps < 0 && !confirmDestructiveOperation(fmt.Sprintf("rollback %d migration steps", -*steps)) {
			fmt.Println("Operation cancelled")
			return
		}
		if err := runSteps(manager, *database, *steps); err != nil {
			log.Fatalf("Failed to run %d migration steps: %v", *steps, err)
		}
		fmt.Printf("✅ Ran %d migration steps successfully\n", *steps)

	case "info":
		if err := showDetailedInfo(manager); err != nil {
			log.Fatalf("Failed to get migration info: %v", err)
		}

	case "create":
		if *name == "" {
			log.Fatal("Migration name is required for create command (use -name flag)")
		}
		if err := createMigration(manager, *database, *name); err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}

	case "seed":
		if err := runSeeding(ctx, cfg, *environment, *reset, *dryRun, *verbose); err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}
		fmt.Printf("✅ Seeding completed successfully with environment: %s\n", *environment)

	default:
		fmt.Printf("❌ Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

// confirmDestructiveOperation prompts user for confirmation on dangerous operations
func confirmDestructiveOperation(operation string) bool {
	fmt.Printf("⚠️  DANGER: About to %s.\n", operation)
	fmt.Printf("This action cannot be undone and may result in data loss.\n")
	fmt.Print("Type 'yes' to confirm (anything else will cancel): ")
	
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "yes"
}

func runMigrations(ctx context.Context, manager *migration.Manager, database, direction string, steps int, dryRun bool) error {
	if dryRun {
		fmt.Printf("🔍 DRY RUN: Would run %s migrations for %s", direction, database)
		if steps > 0 {
			fmt.Printf(" (%d steps)", steps)
		}
		fmt.Println()
	}

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
		// For down migrations, reverse order (ClickHouse first, then PostgreSQL)
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
		fmt.Println("🐘 PostgreSQL Migration Status:")
		if err := manager.ShowPostgresStatus(ctx); err != nil {
			fmt.Printf("❌ Error getting PostgreSQL status: %v\n", err)
		}
		fmt.Println()
		
		fmt.Println("🏠 ClickHouse Migration Status:")
		if err := manager.ShowClickHouseStatus(ctx); err != nil {
			fmt.Printf("❌ Error getting ClickHouse status: %v\n", err)
		}
		
		// Show overall health
		fmt.Println()
		info, err := manager.GetMigrationInfo()
		if err != nil {
			fmt.Printf("❌ Error getting overall status: %v\n", err)
		} else {
			switch info.Overall {
			case "healthy":
				fmt.Println("Overall Status: 🟢 HEALTHY")
			case "dirty":
				fmt.Println("Overall Status: 🟡 DIRTY (requires attention)")
			case "error":
				fmt.Println("Overall Status: 🔴 ERROR")
			default:
				fmt.Printf("Overall Status: ❓ %s\n", info.Overall)
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown database: %s", database)
	}
}

func gotoVersion(manager *migration.Manager, database string, version uint) error {
	switch database {
	case "postgres":
		return manager.GotoPostgres(version)
	case "clickhouse":
		return manager.GotoClickHouse(version)
	case "all":
		if err := manager.GotoPostgres(version); err != nil {
			return fmt.Errorf("postgres goto failed: %w", err)
		}
		return manager.GotoClickHouse(version)
	default:
		return fmt.Errorf("unknown database: %s", database)
	}
}

func forceVersion(manager *migration.Manager, database string, version int) error {
	switch database {
	case "postgres":
		return manager.ForcePostgres(version)
	case "clickhouse":
		return manager.ForceClickHouse(version)
	case "all":
		if err := manager.ForcePostgres(version); err != nil {
			return fmt.Errorf("postgres force failed: %w", err)
		}
		return manager.ForceClickHouse(version)
	default:
		return fmt.Errorf("unknown database: %s", database)
	}
}

func dropTables(manager *migration.Manager, database string) error {
	switch database {
	case "postgres":
		return manager.DropPostgres()
	case "clickhouse":
		return manager.DropClickHouse()
	case "all":
		if err := manager.DropClickHouse(); err != nil {
			return fmt.Errorf("clickhouse drop failed: %w", err)
		}
		return manager.DropPostgres()
	default:
		return fmt.Errorf("unknown database: %s", database)
	}
}

func runSteps(manager *migration.Manager, database string, steps int) error {
	switch database {
	case "postgres":
		return manager.StepsPostgres(steps)
	case "clickhouse":
		return manager.StepsClickHouse(steps)
	case "all":
		if err := manager.StepsPostgres(steps); err != nil {
			return fmt.Errorf("postgres steps failed: %w", err)
		}
		return manager.StepsClickHouse(steps)
	default:
		return fmt.Errorf("unknown database: %s", database)
	}
}

func showDetailedInfo(manager *migration.Manager) error {
	info, err := manager.GetMigrationInfo()
	if err != nil {
		return err
	}
	
	fmt.Println("📊 Detailed Migration Information")
	fmt.Println(strings.Repeat("=", 50))
	
	fmt.Println("\n🐘 PostgreSQL:")
	fmt.Printf("  Status: %s\n", getStatusIcon(info.Postgres.Status))
	fmt.Printf("  Current Version: %d\n", info.Postgres.CurrentVersion)
	fmt.Printf("  Dirty State: %v\n", info.Postgres.IsDirty)
	fmt.Printf("  Migrations Path: %s\n", info.Postgres.MigrationsPath)
	if info.Postgres.Error != "" {
		fmt.Printf("  Error: %s\n", info.Postgres.Error)
	}
	
	fmt.Println("\n🏠 ClickHouse:")
	fmt.Printf("  Status: %s\n", getStatusIcon(info.ClickHouse.Status))
	fmt.Printf("  Current Version: %d\n", info.ClickHouse.CurrentVersion)
	fmt.Printf("  Dirty State: %v\n", info.ClickHouse.IsDirty)
	fmt.Printf("  Migrations Path: %s\n", info.ClickHouse.MigrationsPath)
	if info.ClickHouse.Error != "" {
		fmt.Printf("  Error: %s\n", info.ClickHouse.Error)
	}
	
	fmt.Printf("\n🌐 Overall Status: %s\n", getStatusIcon(info.Overall))
	
	return nil
}

func getStatusIcon(status string) string {
	switch status {
	case "healthy":
		return "🟢 HEALTHY"
	case "dirty":
		return "🟡 DIRTY"
	case "error":
		return "🔴 ERROR"
	default:
		return "❓ " + strings.ToUpper(status)
	}
}

func createMigration(manager *migration.Manager, database, name string) error {
	switch database {
	case "postgres":
		return manager.CreatePostgresMigration(name)
	case "clickhouse":
		return manager.CreateClickHouseMigration(name)
	case "all":
		fmt.Println("Creating migrations for both databases...")
		if err := manager.CreatePostgresMigration(name); err != nil {
			return fmt.Errorf("failed to create postgres migration: %w", err)
		}
		return manager.CreateClickHouseMigration(name)
	default:
		return fmt.Errorf("unknown database: %s", database)
	}
}

// parseDatabaseSelection converts database flag string to DatabaseType slice
func parseDatabaseSelection(database string) ([]migration.DatabaseType, error) {
	switch database {
	case "postgres":
		return []migration.DatabaseType{migration.PostgresDB}, nil
	case "clickhouse":
		return []migration.DatabaseType{migration.ClickHouseDB}, nil
	case "all":
		return []migration.DatabaseType{migration.PostgresDB, migration.ClickHouseDB}, nil
	default:
		return nil, fmt.Errorf("unknown database: %s (valid options: postgres, clickhouse, all)", database)
	}
}

func printUsage() {
	fmt.Println("🚀 Brokle Migration Tool - Production Database Migration & Seeding CLI")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  migrate <command> [flags]")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  up                    Run all pending migrations")
	fmt.Println("  down                  Rollback migrations (with confirmation)")
	fmt.Println("  status                Show migration status for all databases")
	fmt.Println("  goto -version N       Migrate to specific version (with confirmation)")
	fmt.Println("  force -version N      Force version without migration (DANGEROUS)")
	fmt.Println("  drop                  Drop all tables (DANGEROUS)")
	fmt.Println("  steps -steps N        Run N migration steps (negative for rollback)")
	fmt.Println("  info                  Show detailed migration information")
	fmt.Println("  create -name NAME     Create new migration files")
	fmt.Println("  seed                  Seed database with test data")
	fmt.Println()
	fmt.Println("FLAGS:")
	fmt.Println("  -db string           Database to target: all, postgres, clickhouse (default: all)")
	fmt.Println("  -steps int           Number of migration steps")
	fmt.Println("  -version int         Target version for goto/force commands")
	fmt.Println("  -name string         Migration name for create command")
	fmt.Println("  -dry-run             Show what would happen without executing")
	fmt.Println("  -env string          Seeding environment: development, demo, test (default: development)")
	fmt.Println("  -reset               Reset existing data before seeding (DANGEROUS)")
	fmt.Println("  -verbose             Verbose output for seeding operations")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  migrate up                              # Run all pending migrations")
	fmt.Println("  migrate -db postgres up                 # Run PostgreSQL migrations only")
	fmt.Println("  migrate down                            # Rollback with confirmation")
	fmt.Println("  migrate goto -version 5                 # Go to version 5 with confirmation")
	fmt.Println("  migrate steps -steps 2                  # Run 2 migration steps")
	fmt.Println("  migrate steps -steps -1                 # Rollback 1 migration step")
	fmt.Println("  migrate create -name 'add_users'        # Create new migration")
	fmt.Println("  migrate status                          # Show migration status")
	fmt.Println("  migrate info                            # Show detailed information")
	fmt.Println("  migrate -dry-run up                     # Preview migrations")
	fmt.Println("  migrate seed                            # Seed with development data")
	fmt.Println("  migrate seed -env demo                  # Seed with demo data")
	fmt.Println("  migrate seed -env test                  # Seed with test data")
	fmt.Println("  migrate seed -reset -verbose            # Reset and seed with verbose output")
	fmt.Println("  migrate seed -dry-run                   # Preview seeding plan")
	fmt.Println()
	fmt.Println("SAFETY:")
	fmt.Println("  🛡️  Destructive operations require explicit 'yes' confirmation")
	fmt.Println("  🔍 Use -dry-run to preview changes safely")
	fmt.Println("  📊 Check 'status' and 'info' before running migrations")
	fmt.Println("  🌱 Use 'seed' to populate database with test data")
}

// runSeeding handles database seeding operations
func runSeeding(ctx context.Context, cfg *config.Config, environment string, reset, dryRun, verbose bool) error {
	// Initialize seeder manager
	manager, err := seeder.NewManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize seeder manager: %w", err)
	}
	defer manager.Close()

	// Configure seeder options
	options := &seeder.Options{
		Environment: environment,
		Reset:       reset,
		DryRun:      dryRun,
		Verbose:     verbose,
	}

	// Confirm reset operation if requested
	if reset && !dryRun {
		if !confirmDestructiveOperation(fmt.Sprintf("RESET ALL DATA and seed with %s environment", environment)) {
			return fmt.Errorf("seeding cancelled by user")
		}
	}

	// Show seed plan in dry-run mode
	if dryRun {
		fmt.Printf("🔍 DRY RUN: Seeding plan for environment '%s':\n", environment)
		
		// Load seed data to show plan
		dataLoader := seeder.NewDataLoader()
		seedData, err := dataLoader.LoadSeedData(environment)
		if err != nil {
			return fmt.Errorf("failed to load seed data for preview: %w", err)
		}
		
		manager.PrintSeedPlan(seedData)
		return nil
	}

	// Run actual seeding (PostgreSQL only for now)
	return manager.SeedPostgres(ctx, options)
}