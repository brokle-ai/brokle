package migration

import (
	"context"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	PostgresDB  DatabaseType = "postgres"
	ClickHouseDB DatabaseType = "clickhouse"
)

// MigrationDirection represents the direction of migration
type MigrationDirection string

const (
	Up   MigrationDirection = "up"
	Down MigrationDirection = "down"
)

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Database       DatabaseType `json:"database"`
	CurrentVersion uint         `json:"current_version"`
	IsDirty        bool         `json:"is_dirty"`
	Status         string       `json:"status"` // "healthy", "dirty", "error"
	Error          string       `json:"error,omitempty"`
	MigrationsPath string       `json:"migrations_path"`
	TotalMigrations int         `json:"total_migrations"`
}

// MigrationInfo represents detailed information about migrations
type MigrationInfo struct {
	Postgres   MigrationStatus `json:"postgres"`
	ClickHouse MigrationStatus `json:"clickhouse"`
	Overall    string          `json:"overall_status"`
}

// DatabaseRunner defines the interface for database-specific migration runners
type DatabaseRunner interface {
	// Core migration operations
	Up() error
	Down() error
	Steps(n int) error
	Goto(version uint) error
	Force(version int) error
	Drop() error
	
	// Information and status
	Version() (uint, bool, error)
	GetMigrationInfo() (map[string]interface{}, error)
	
	// Lifecycle
	Close() (error, error)
}

// HealthChecker defines the interface for migration health checks
type HealthChecker interface {
	HealthCheck() map[string]interface{}
	GetStatus() MigrationStatus
}

// AutoMigrator defines the interface for automatic migrations
type AutoMigrator interface {
	AutoMigrate(ctx context.Context) error
	CanAutoMigrate() bool
}

// MigrationManager defines the complete interface for the migration system
type MigrationManager interface {
	// Multi-database operations
	MigratePostgresUp(ctx context.Context, steps int, dryRun bool) error
	MigratePostgresDown(ctx context.Context, steps int, dryRun bool) error
	MigrateClickHouseUp(ctx context.Context, steps int, dryRun bool) error
	MigrateClickHouseDown(ctx context.Context, steps int, dryRun bool) error
	
	// Status and information
	ShowPostgresStatus(ctx context.Context) error
	ShowClickHouseStatus(ctx context.Context) error
	GetMigrationInfo() (*MigrationInfo, error)
	HealthCheck() map[string]interface{}
	GetStatus() MigrationStatus
	
	// Migration creation
	CreatePostgresMigration(name string) error
	CreateClickHouseMigration(name string) error
	
	// Advanced operations
	GotoPostgres(version uint) error
	GotoClickHouse(version uint) error
	ForcePostgres(version int) error
	ForceClickHouse(version int) error
	DropPostgres() error
	DropClickHouse() error
	StepsPostgres(n int) error
	StepsClickHouse(n int) error
	
	// Auto-migration
	AutoMigrate(ctx context.Context) error
	CanAutoMigrate() bool
	
	// Lifecycle
	Shutdown() error
}