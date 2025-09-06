package migration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/infrastructure/database"
	sharedDB "github.com/brokle-ai/brokle-platform/shared/go/common/database"
	sharedLogger "github.com/brokle-ai/brokle-platform/shared/go/pkg/logger"
)

// Manager coordinates migrations across multiple databases
type Manager struct {
	config           *config.Config
	logger           *logrus.Logger
	postgresRunner   *sharedDB.MigrationRunner
	clickhouseRunner *sharedDB.ClickHouseMigrationRunner
	postgresDB       *database.PostgresDB
	clickhouseDB     *database.ClickHouseDB
}

// NewManager creates a new migration manager with all databases
func NewManager(cfg *config.Config) (*Manager, error) {
	return NewManagerWithDatabases(cfg, []DatabaseType{PostgresDB, ClickHouseDB})
}

// NewManagerWithDatabases creates a new migration manager with only specified databases
func NewManagerWithDatabases(cfg *config.Config, databases []DatabaseType) (*Manager, error) {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	manager := &Manager{
		config: cfg,
		logger: logger,
	}

	// Helper function to check if database is requested
	needsDatabase := func(dbType DatabaseType) bool {
		for _, db := range databases {
			if db == dbType {
				return true
			}
		}
		return false
	}

	// Conditionally initialize PostgreSQL
	if needsDatabase(PostgresDB) {
		postgresDB, err := database.NewPostgresDB(cfg, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize postgres database: %w", err)
		}
		manager.postgresDB = postgresDB

		// Initialize PostgreSQL migration runner
		if err := manager.initPostgresRunner(); err != nil {
			return nil, fmt.Errorf("failed to initialize postgres runner: %w", err)
		}
		logger.Info("PostgreSQL migration manager initialized")
	}

	// Conditionally initialize ClickHouse
	if needsDatabase(ClickHouseDB) {
		clickhouseDB, err := database.NewClickHouseDB(cfg, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize clickhouse database: %w", err)
		}
		manager.clickhouseDB = clickhouseDB

		// Initialize ClickHouse migration runner
		if err := manager.initClickHouseRunner(); err != nil {
			return nil, fmt.Errorf("failed to initialize clickhouse runner: %w", err)
		}
		logger.Info("ClickHouse migration manager initialized")
	}

	logger.WithField("databases", databases).Info("Migration manager initialized successfully")
	return manager, nil
}

// initPostgresRunner initializes the PostgreSQL migration runner
func (m *Manager) initPostgresRunner() error {
	if m.postgresDB == nil {
		return fmt.Errorf("postgres database not initialized")
	}

	// Get migrations path
	migrationsPath := m.getMigrationsPath(PostgresDB)
	
	// Create migration runner using shared library
	runner, err := sharedDB.NewMigrationRunner(
		&sharedDB.Database{DB: m.postgresDB.DB},
		"brokle-monolith",
		migrationsPath,
	)
	if err != nil {
		return fmt.Errorf("failed to create postgres migration runner: %w", err)
	}

	m.postgresRunner = runner
	m.logger.WithField("migrations_path", migrationsPath).Info("PostgreSQL migration runner initialized")
	return nil
}

// initClickHouseRunner initializes the ClickHouse migration runner
func (m *Manager) initClickHouseRunner() error {
	if m.clickhouseDB == nil {
		return fmt.Errorf("clickhouse database not initialized")
	}

	// Get migrations path
	migrationsPath := m.getMigrationsPath(ClickHouseDB)
	
	// Create ClickHouse migration configuration
	clickhouseConfig := &sharedDB.ClickHouseMigrationConfig{
		Host:             m.config.ClickHouse.Host,
		Port:             fmt.Sprintf("%d", m.config.ClickHouse.Port),
		Database:         m.config.ClickHouse.Database,
		Username:         m.config.ClickHouse.User,
		Password:         m.config.ClickHouse.Password,
		MigrationsTable:  m.config.ClickHouse.MigrationsTable,
		MigrationsEngine: m.config.ClickHouse.MigrationsEngine,
		SourcePath:       migrationsPath,
	}

	// Create shared logger for ClickHouse runner
	sharedLog := sharedLogger.New("info", "json")
	
	// Create ClickHouse migration runner using shared library
	runner, err := sharedDB.NewClickHouseMigrationRunner(
		clickhouseConfig,
		sharedLog,
	)
	if err != nil {
		return fmt.Errorf("failed to create clickhouse migration runner: %w", err)
	}

	m.clickhouseRunner = runner
	m.logger.WithField("migrations_path", migrationsPath).Info("ClickHouse migration runner initialized")
	return nil
}

// getMigrationsPath returns the migrations path for a specific database type
func (m *Manager) getMigrationsPath(dbType DatabaseType) string {
	basePath := "migrations"
	
	switch dbType {
	case PostgresDB:
		if m.config.Database.MigrationsPath != "" {
			return m.config.Database.MigrationsPath
		}
		return filepath.Join(basePath, "postgres")
	case ClickHouseDB:
		if m.config.ClickHouse.MigrationsPath != "" {
			return m.config.ClickHouse.MigrationsPath
		}
		return filepath.Join(basePath, "clickhouse")
	default:
		return basePath
	}
}

// MigratePostgresUp runs PostgreSQL migrations up
func (m *Manager) MigratePostgresUp(ctx context.Context, steps int, dryRun bool) error {
	if m.postgresRunner == nil {
		return fmt.Errorf("PostgreSQL not initialized - run with -db postgres or -db all")
	}

	if dryRun {
		m.logger.Info("DRY RUN: Would run PostgreSQL migrations up")
		return nil
	}

	m.logger.WithField("steps", steps).Info("Running PostgreSQL migrations up")
	
	if steps == 0 {
		return m.postgresRunner.Up()
	}
	return m.postgresRunner.Steps(steps)
}

// MigratePostgresDown runs PostgreSQL migrations down
func (m *Manager) MigratePostgresDown(ctx context.Context, steps int, dryRun bool) error {
	if m.postgresRunner == nil {
		return fmt.Errorf("PostgreSQL not initialized - run with -db postgres or -db all")
	}

	if dryRun {
		m.logger.Info("DRY RUN: Would run PostgreSQL migrations down")
		return nil
	}

	m.logger.WithField("steps", steps).Info("Running PostgreSQL migrations down")
	
	if steps == 0 {
		return m.postgresRunner.Down()
	}
	return m.postgresRunner.Steps(-steps)
}

// MigrateClickHouseUp runs ClickHouse migrations up
func (m *Manager) MigrateClickHouseUp(ctx context.Context, steps int, dryRun bool) error {
	if m.clickhouseRunner == nil {
		return fmt.Errorf("ClickHouse not initialized - run with -db clickhouse or -db all")
	}

	if dryRun {
		m.logger.Info("DRY RUN: Would run ClickHouse migrations up")
		return nil
	}

	m.logger.WithField("steps", steps).Info("Running ClickHouse migrations up")
	
	if steps == 0 {
		return m.clickhouseRunner.Up()
	}
	return m.clickhouseRunner.Steps(steps)
}

// MigrateClickHouseDown runs ClickHouse migrations down
func (m *Manager) MigrateClickHouseDown(ctx context.Context, steps int, dryRun bool) error {
	if m.clickhouseRunner == nil {
		return fmt.Errorf("ClickHouse not initialized - run with -db clickhouse or -db all")
	}

	if dryRun {
		m.logger.Info("DRY RUN: Would run ClickHouse migrations down")
		return nil
	}

	m.logger.WithField("steps", steps).Info("Running ClickHouse migrations down")
	
	if steps == 0 {
		return m.clickhouseRunner.Down()
	}
	return m.clickhouseRunner.Steps(-steps)
}

// ShowPostgresStatus displays PostgreSQL migration status
func (m *Manager) ShowPostgresStatus(ctx context.Context) error {
	if m.postgresRunner == nil {
		fmt.Println("PostgreSQL: ❌ NOT INITIALIZED")
		fmt.Println("  Run with -db postgres or -db all to initialize PostgreSQL")
		return nil
	}

	version, dirty, err := m.postgresRunner.Version()
	if err != nil {
		return fmt.Errorf("failed to get postgres version: %w", err)
	}

	status := "clean"
	statusIcon := "✅"
	if dirty {
		status = "dirty"
		statusIcon = "⚠️"
	}

	migrationsPath := m.getMigrationsPath(PostgresDB)
	
	fmt.Printf("PostgreSQL Migration Status:\n")
	fmt.Printf("  %s Current Version: %d (%s)\n", statusIcon, version, status)
	fmt.Printf("  📁 Migrations Path: %s\n", migrationsPath)
	
	if info, err := m.postgresRunner.GetMigrationInfo(); err == nil {
		if count, ok := info["migration_count"].(int); ok {
			fmt.Printf("  📊 Total Migrations: %d\n", count)
		}
	}
	
	return nil
}

// ShowClickHouseStatus displays ClickHouse migration status
func (m *Manager) ShowClickHouseStatus(ctx context.Context) error {
	if m.clickhouseRunner == nil {
		fmt.Println("ClickHouse: ❌ NOT INITIALIZED")
		fmt.Println("  Run with -db clickhouse or -db all to initialize ClickHouse")
		return nil
	}

	version, dirty, err := m.clickhouseRunner.Version()
	if err != nil {
		return fmt.Errorf("failed to get clickhouse version: %w", err)
	}

	status := "clean"
	statusIcon := "✅"
	if dirty {
		status = "dirty"
		statusIcon = "⚠️"
	}

	migrationsPath := m.getMigrationsPath(ClickHouseDB)
	
	fmt.Printf("ClickHouse Migration Status:\n")
	fmt.Printf("  %s Current Version: %d (%s)\n", statusIcon, version, status)
	fmt.Printf("  📁 Migrations Path: %s\n", migrationsPath)
	
	if info, err := m.clickhouseRunner.GetMigrationInfo(); err == nil {
		if sourceType, ok := info["type"].(string); ok && sourceType != "" {
			fmt.Printf("  📊 Database Type: %s\n", sourceType)
		}
	}
	
	return nil
}

// GetMigrationInfo returns detailed migration information for both databases
func (m *Manager) GetMigrationInfo() (*MigrationInfo, error) {
	info := &MigrationInfo{}
	
	// Get PostgreSQL info
	if m.postgresRunner == nil {
		info.Postgres.Status = "not_initialized"
		info.Postgres.Error = "PostgreSQL not initialized - run with -db postgres or -db all"
		info.Postgres.Database = PostgresDB
		info.Postgres.MigrationsPath = m.getMigrationsPath(PostgresDB)
	} else {
		pgVersion, pgDirty, err := m.postgresRunner.Version()
		if err != nil {
			info.Postgres.Status = "error"
			info.Postgres.Error = err.Error()
		} else {
			info.Postgres.Database = PostgresDB
			info.Postgres.CurrentVersion = pgVersion
			info.Postgres.IsDirty = pgDirty
			info.Postgres.MigrationsPath = m.getMigrationsPath(PostgresDB)
			if pgDirty {
				info.Postgres.Status = "dirty"
			} else {
				info.Postgres.Status = "healthy"
			}
		}
	}
	
	// Get ClickHouse info
	if m.clickhouseRunner == nil {
		info.ClickHouse.Status = "not_initialized"
		info.ClickHouse.Error = "ClickHouse not initialized - run with -db clickhouse or -db all"
		info.ClickHouse.Database = ClickHouseDB
		info.ClickHouse.MigrationsPath = m.getMigrationsPath(ClickHouseDB)
	} else {
		chVersion, chDirty, err := m.clickhouseRunner.Version()
		if err != nil {
			info.ClickHouse.Status = "error"
			info.ClickHouse.Error = err.Error()
		} else {
			info.ClickHouse.Database = ClickHouseDB
			info.ClickHouse.CurrentVersion = chVersion
			info.ClickHouse.IsDirty = chDirty
			info.ClickHouse.MigrationsPath = m.getMigrationsPath(ClickHouseDB)
			if chDirty {
				info.ClickHouse.Status = "dirty"
			} else {
				info.ClickHouse.Status = "healthy"
			}
		}
	}
	
	// Determine overall status
	if info.Postgres.Status == "error" || info.ClickHouse.Status == "error" {
		info.Overall = "error"
	} else if info.Postgres.Status == "dirty" || info.ClickHouse.Status == "dirty" {
		info.Overall = "dirty"
	} else if info.Postgres.Status == "not_initialized" && info.ClickHouse.Status == "not_initialized" {
		info.Overall = "not_initialized"
	} else if info.Postgres.Status == "not_initialized" || info.ClickHouse.Status == "not_initialized" {
		info.Overall = "partial"
	} else {
		info.Overall = "healthy"
	}
	
	return info, nil
}

// HealthCheck returns health status for monitoring endpoints
func (m *Manager) HealthCheck() map[string]interface{} {
	health := make(map[string]interface{})
	
	var pgErr, chErr error
	var pgDirty, chDirty bool
	var pgVersion, chVersion uint
	
	// Check PostgreSQL
	if m.postgresRunner == nil {
		health["postgres"] = map[string]interface{}{
			"status": "not_initialized",
			"error":  "PostgreSQL not initialized - run with -db postgres or -db all",
		}
		pgErr = fmt.Errorf("not initialized")
	} else {
		pgVersion, pgDirty, pgErr = m.postgresRunner.Version()
		health["postgres"] = map[string]interface{}{
			"status":          m.getHealthStatus(pgErr, pgDirty),
			"current_version": pgVersion,
			"dirty":          pgDirty,
		}
		if pgErr != nil {
			health["postgres"].(map[string]interface{})["error"] = pgErr.Error()
		}
	}
	
	// Check ClickHouse
	if m.clickhouseRunner == nil {
		health["clickhouse"] = map[string]interface{}{
			"status": "not_initialized",
			"error":  "ClickHouse not initialized - run with -db clickhouse or -db all",
		}
		chErr = fmt.Errorf("not initialized")
	} else {
		chVersion, chDirty, chErr = m.clickhouseRunner.Version()
		health["clickhouse"] = map[string]interface{}{
			"status":          m.getHealthStatus(chErr, chDirty),
			"current_version": chVersion,
			"dirty":          chDirty,
		}
		if chErr != nil {
			health["clickhouse"].(map[string]interface{})["error"] = chErr.Error()
		}
	}
	
	// Overall status
	overallHealthy := pgErr == nil && chErr == nil && !pgDirty && !chDirty
	if overallHealthy {
		health["overall_status"] = "healthy"
	} else {
		health["overall_status"] = "unhealthy"
	}
	
	return health
}

// getHealthStatus converts error and dirty state to health status string
func (m *Manager) getHealthStatus(err error, dirty bool) string {
	if err != nil {
		return "error"
	}
	if dirty {
		return "dirty"
	}
	return "healthy"
}

// GetStatus returns migration status for the manager (required by interface)
func (m *Manager) GetStatus() MigrationStatus {
	// Return overall status - in practice, this might return the most critical status
	pgVersion, pgDirty, pgErr := m.postgresRunner.Version()
	
	status := MigrationStatus{
		Database:        PostgresDB, // Primary database
		CurrentVersion:  pgVersion,
		IsDirty:        pgDirty,
		MigrationsPath: m.getMigrationsPath(PostgresDB),
	}
	
	if pgErr != nil {
		status.Status = "error"
		status.Error = pgErr.Error()
	} else if pgDirty {
		status.Status = "dirty"
	} else {
		status.Status = "healthy"
	}
	
	return status
}

// AutoMigrate runs migrations automatically on startup if configured
func (m *Manager) AutoMigrate(ctx context.Context) error {
	if !m.CanAutoMigrate() {
		return fmt.Errorf("auto-migration is disabled")
	}

	m.logger.Info("Starting auto-migration")
	
	// Run PostgreSQL migrations
	if err := m.MigratePostgresUp(ctx, 0, false); err != nil {
		return fmt.Errorf("postgres auto-migration failed: %w", err)
	}
	
	// Run ClickHouse migrations
	if err := m.MigrateClickHouseUp(ctx, 0, false); err != nil {
		return fmt.Errorf("clickhouse auto-migration failed: %w", err)
	}
	
	m.logger.Info("Auto-migration completed successfully")
	return nil
}

// CanAutoMigrate returns true if auto-migration is enabled
func (m *Manager) CanAutoMigrate() bool {
	return m.config.Database.AutoMigrate
}

// Advanced operations

// GotoPostgres migrates PostgreSQL to a specific version
func (m *Manager) GotoPostgres(version uint) error {
	if m.postgresRunner == nil {
		return fmt.Errorf("PostgreSQL not initialized - run with -db postgres or -db all")
	}
	return m.postgresRunner.Goto(version)
}

// GotoClickHouse migrates ClickHouse to a specific version
func (m *Manager) GotoClickHouse(version uint) error {
	if m.clickhouseRunner == nil {
		return fmt.Errorf("ClickHouse not initialized - run with -db clickhouse or -db all")
	}
	return m.clickhouseRunner.Goto(version)
}

// ForcePostgres forces PostgreSQL to a specific version
func (m *Manager) ForcePostgres(version int) error {
	if m.postgresRunner == nil {
		return fmt.Errorf("PostgreSQL not initialized - run with -db postgres or -db all")
	}
	return m.postgresRunner.Force(version)
}

// ForceClickHouse forces ClickHouse to a specific version
func (m *Manager) ForceClickHouse(version int) error {
	if m.clickhouseRunner == nil {
		return fmt.Errorf("ClickHouse not initialized - run with -db clickhouse or -db all")
	}
	return m.clickhouseRunner.Force(version)
}

// DropPostgres drops all PostgreSQL tables
func (m *Manager) DropPostgres() error {
	if m.postgresRunner == nil {
		return fmt.Errorf("PostgreSQL not initialized - run with -db postgres or -db all")
	}
	return m.postgresRunner.Drop()
}

// DropClickHouse drops all ClickHouse tables
func (m *Manager) DropClickHouse() error {
	if m.clickhouseRunner == nil {
		return fmt.Errorf("ClickHouse not initialized - run with -db clickhouse or -db all")
	}
	return m.clickhouseRunner.Drop()
}

// StepsPostgres runs n PostgreSQL migration steps
func (m *Manager) StepsPostgres(n int) error {
	if m.postgresRunner == nil {
		return fmt.Errorf("PostgreSQL not initialized - run with -db postgres or -db all")
	}
	return m.postgresRunner.Steps(n)
}

// StepsClickHouse runs n ClickHouse migration steps
func (m *Manager) StepsClickHouse(n int) error {
	if m.clickhouseRunner == nil {
		return fmt.Errorf("ClickHouse not initialized - run with -db clickhouse or -db all")
	}
	return m.clickhouseRunner.Steps(n)
}

// CreatePostgresMigration creates a new PostgreSQL migration file
func (m *Manager) CreatePostgresMigration(name string) error {
	migrationsPath := m.getMigrationsPath(PostgresDB)
	
	// Create migrations directory if it doesn't exist
	if err := os.MkdirAll(migrationsPath, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}
	
	timestamp := time.Now().Format("20060102150405")
	
	// Create up migration file
	upFile := filepath.Join(migrationsPath, fmt.Sprintf("%s_%s.up.sql", timestamp, name))
	if err := os.WriteFile(upFile, []byte("-- Migration: "+name+"\n-- Created: "+time.Now().Format(time.RFC3339)+"\n\n"), 0644); err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}
	
	// Create down migration file
	downFile := filepath.Join(migrationsPath, fmt.Sprintf("%s_%s.down.sql", timestamp, name))
	if err := os.WriteFile(downFile, []byte("-- Rollback: "+name+"\n-- Created: "+time.Now().Format(time.RFC3339)+"\n\n"), 0644); err != nil {
		return fmt.Errorf("failed to create down migration file: %w", err)
	}
	
	m.logger.WithFields(logrus.Fields{
		"name":      name,
		"up_file":   upFile,
		"down_file": downFile,
	}).Info("PostgreSQL migration files created")
	
	fmt.Printf("PostgreSQL migration files created:\n")
	fmt.Printf("  Up:   %s\n", upFile)
	fmt.Printf("  Down: %s\n", downFile)
	
	return nil
}

// CreateClickHouseMigration creates a new ClickHouse migration file
func (m *Manager) CreateClickHouseMigration(name string) error {
	migrationsPath := m.getMigrationsPath(ClickHouseDB)
	
	// Create migrations directory if it doesn't exist
	if err := os.MkdirAll(migrationsPath, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}
	
	timestamp := time.Now().Format("20060102150405")
	
	// Create up migration file
	upFile := filepath.Join(migrationsPath, fmt.Sprintf("%s_%s.up.sql", timestamp, name))
	if err := os.WriteFile(upFile, []byte("-- ClickHouse Migration: "+name+"\n-- Created: "+time.Now().Format(time.RFC3339)+"\n\n"), 0644); err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}
	
	// Create down migration file
	downFile := filepath.Join(migrationsPath, fmt.Sprintf("%s_%s.down.sql", timestamp, name))
	if err := os.WriteFile(downFile, []byte("-- ClickHouse Rollback: "+name+"\n-- Created: "+time.Now().Format(time.RFC3339)+"\n\n"), 0644); err != nil {
		return fmt.Errorf("failed to create down migration file: %w", err)
	}
	
	m.logger.WithFields(logrus.Fields{
		"name":      name,
		"up_file":   upFile,
		"down_file": downFile,
	}).Info("ClickHouse migration files created")
	
	fmt.Printf("ClickHouse migration files created:\n")
	fmt.Printf("  Up:   %s\n", upFile)
	fmt.Printf("  Down: %s\n", downFile)
	
	return nil
}

// Shutdown gracefully shuts down the migration manager
func (m *Manager) Shutdown() error {
	m.logger.Info("Shutting down migration manager")
	
	var lastErr error
	
	// Close PostgreSQL runner
	if m.postgresRunner != nil {
		if _, err := m.postgresRunner.Close(); err != nil {
			m.logger.WithError(err).Error("Failed to close PostgreSQL migration runner")
			lastErr = err
		}
	}
	
	// Close ClickHouse runner
	if m.clickhouseRunner != nil {
		if _, err := m.clickhouseRunner.Close(); err != nil {
			m.logger.WithError(err).Error("Failed to close ClickHouse migration runner")
			lastErr = err
		}
	}
	
	// Close databases
	// Close PostgreSQL
	if m.postgresDB != nil {
		if err := m.postgresDB.Close(); err != nil {
			m.logger.WithError(err).Error("Failed to close PostgreSQL connection")
			lastErr = err
		}
	}
	
	
	// Close ClickHouse
	if m.clickhouseDB != nil {
		if err := m.clickhouseDB.Close(); err != nil {
			m.logger.WithError(err).Error("Failed to close ClickHouse connection")
			lastErr = err
		}
	}
	
	m.logger.Info("Migration manager shutdown completed")
	return lastErr
}