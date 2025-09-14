# 🚀 Brokle Database CLI

Production-ready database migration and seeding CLI for the Brokle AI Control Plane. Handles multi-database migrations and data seeding with comprehensive safety features and enterprise-grade reliability.

## Overview

The Brokle Database CLI manages database operations across multiple databases:

### 🔄 **Schema Migrations**
- **PostgreSQL** - Primary transactional data
- **ClickHouse** - Analytics and time-series data
- **Multi-database coordination** - Ensures consistency across all databases

### 🌱 **Data Seeding**
- **Development data** - Complete setup with users, organizations, RBAC
- **Demo data** - Minimal showcase data
- **Test data** - Basic fixtures for testing
- **Environment-aware** - Different data for dev, demo, test environments

## Quick Start

```bash
# Build the database CLI
go build -o migrate ./cmd/migrate/main.go

# Run all pending migrations
./migrate up

# Seed database with development data
./migrate seed

# Check migration status
./migrate status

# Create new migration
./migrate create -name "add_users_table"
```

## Commands

### 🔄 Schema Migration Commands

#### `migrate up`
Run all pending migrations for all databases.
```bash
./migrate up                    # All databases
./migrate -db postgres up       # PostgreSQL only
./migrate -db clickhouse up     # ClickHouse only
./migrate -dry-run up           # Preview changes
```

#### `migrate down`
Rollback migrations with safety confirmation.
```bash
./migrate down                  # Rollback all databases (with confirmation)
./migrate -db postgres down     # Rollback PostgreSQL only
./migrate -steps 2 down         # Rollback specific number of steps
```

#### `migrate status`
Show current migration status across all databases.
```bash
./migrate status                # All databases
./migrate -db postgres status   # PostgreSQL only
```

### Advanced Commands

#### `migrate steps -steps N`
Run a specific number of migration steps (positive for up, negative for down).
```bash
./migrate steps -steps 2        # Run 2 migrations forward
./migrate steps -steps -1       # Rollback 1 migration
./migrate -db postgres steps -steps 3  # PostgreSQL specific
```

#### `migrate goto -version N`
Migrate to a specific version (with safety confirmation).
```bash
./migrate goto -version 5       # Go to version 5 (with confirmation)
./migrate -db postgres goto -version 3  # PostgreSQL to version 3
```

#### `migrate create -name NAME`
Create new migration files for all databases.
```bash
./migrate create -name "add_users_table"
./migrate create -name "add_analytics_events"
```

#### `migrate info`
Show detailed migration information and health status.
```bash
./migrate info                  # Comprehensive system information
```

### 🌱 Data Seeding Commands

#### `migrate seed`
Populate database with environment-specific test data.
```bash
./migrate seed                  # Seed with development data (default)
./migrate seed -env demo        # Seed with demo data
./migrate seed -env test        # Seed with test data
./migrate seed -reset -verbose  # Reset existing data and seed with verbose output
./migrate seed -dry-run         # Preview seeding plan without executing
```

**Seeding includes:**
- **Users** - Sample users with proper authentication
- **Organizations** - Multi-tenant organization structure  
- **RBAC** - Roles, permissions, and memberships
- **Projects & Environments** - Complete project hierarchy
- **Onboarding Questions** - User onboarding workflow setup
- **⚠️ API Keys Skipped** - Create manually via web interface (JSON serialization issue)

### Dangerous Operations (Use with Caution)

#### `migrate force -version N`
Force set database version without running migrations.
```bash
./migrate force -version 0      # Mark as version 0 (DANGEROUS)
```

#### `migrate drop`
Drop all database tables (requires confirmation).
```bash
./migrate drop                  # Drop all tables (DANGEROUS)
```

## Flags

### Migration Flags
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-db` | Database to target: `all`, `postgres`, `clickhouse` | `all` | `-db postgres` |
| `-steps` | Number of migration steps (+ for up, - for down) | `0` (all) | `-steps 2` |
| `-version` | Target version for goto/force commands | `0` | `-version 5` |
| `-name` | Migration name for create command | Required | `-name "add_users"` |
| `-dry-run` | Preview changes without executing | `false` | `-dry-run` |

### Seeding Flags
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-env` | Seeding environment: `development`, `demo`, `test` | `development` | `-env demo` |
| `-reset` | Reset existing data before seeding (DANGEROUS) | `false` | `-reset` |
| `-verbose` | Show detailed seeding output | `false` | `-verbose` |
| `-dry-run` | Preview seeding plan without executing | `false` | `-dry-run` |

## Safety Features

### 🛡️ Confirmation Prompts
Destructive operations require explicit confirmation:
- `migrate down` - Rollback confirmation
- `migrate drop` - Table drop confirmation  
- `migrate goto` - Version change confirmation
- `migrate force` - Force version confirmation
- `migrate seed -reset` - Data reset confirmation

### 🔍 Dry Run Mode
Preview operations without executing:
```bash
./migrate -dry-run up           # See what migrations would run
./migrate -dry-run down         # See what would be rolled back
./migrate seed -dry-run         # See what data would be seeded
```

### 📊 Status Monitoring
Check system health before migrations:
```bash
./migrate status                # Quick status overview
./migrate info                  # Detailed health information
```

### 🎯 Multi-Database Coordination
Ensures consistency across PostgreSQL and ClickHouse:
- Atomic operations where possible
- Comprehensive error reporting
- Rollback capabilities for failed migrations

## File Structure

### Migration Files

#### PostgreSQL Migrations
Located in `migrations/postgres/`
```
migrations/postgres/
├── 001_initial_schema.up.sql
├── 001_initial_schema.down.sql
├── 002_add_users.up.sql
└── 002_add_users.down.sql
```

#### ClickHouse Migrations  
Located in `migrations/clickhouse/`
```
migrations/clickhouse/
├── 001_initial_analytics.up.sql
├── 001_initial_analytics.down.sql
├── 002_add_events.up.sql
└── 002_add_events.down.sql
```

### Seed Data Files
Located in `seeds/` directory with YAML format:
```
seeds/
├── dev.yaml          # Development environment (full dataset)
├── demo.yaml         # Demo environment (showcase data)
└── test.yaml         # Test environment (minimal fixtures)
```

#### Seed Data Structure
Each YAML file contains:
```yaml
organizations:          # Multi-tenant organizations
users:                 # Sample users with authentication
rbac:                  # Roles, permissions, memberships
  permissions:         # System permissions (13 default)
  roles:               # User roles (7 default)
  memberships:         # User-organization-role assignments
projects:              # Projects and environments
api_keys:              # API keys (skipped due to JSON issue)
onboarding_questions:  # User onboarding workflow
```

## Usage Examples

### Development Workflow - Migrations
```bash
# 1. Create new migration
./migrate create -name "add_api_keys_table"

# 2. Edit the generated .up.sql and .down.sql files
# 3. Test with dry-run
./migrate -dry-run up

# 4. Run the migration
./migrate up

# 5. Verify status
./migrate status
```

### Development Workflow - Seeding
```bash
# 1. Set up fresh database
./migrate up

# 2. Preview what will be seeded
./migrate seed -dry-run

# 3. Seed with development data
./migrate seed

# 4. Or seed with reset for clean state
./migrate seed -reset -verbose
```

### Production Deployment
```bash
# 1. Check current status
./migrate status

# 2. Preview changes
./migrate -dry-run up

# 3. Run migrations
./migrate up

# 4. Verify completion
./migrate info
```

### Rollback Scenario
```bash
# 1. Check what would be rolled back
./migrate -dry-run down

# 2. Rollback specific number of steps
./migrate steps -steps -2

# 3. Verify rollback
./migrate status
```

### Emergency Recovery
```bash
# 1. Check database state
./migrate info

# 2. Force version if needed (DANGEROUS)
./migrate force -version 3

# 3. Re-run from correct state
./migrate up
```

### Seeding Different Environments
```bash
# Demo environment (minimal data)
./migrate seed -env demo

# Test environment (basic fixtures)
./migrate seed -env test

# Reset and seed development data
./migrate seed -env development -reset

# Preview seeding plan for any environment
./migrate seed -env demo -dry-run
```

## Configuration

The migration tool uses the main application configuration file:

```yaml
database:
  auto_migrate: false                    # Enable auto-migration on startup
  migrations_path: "migrations"          # Path to migration files
  migrations_table: "schema_migrations"  # Migration tracking table
  username: ""                          # Override database username

clickhouse:
  user: "brokle"                        # ClickHouse user for migrations
  migrations_engine: "MergeTree"        # ClickHouse table engine for migrations
```

Environment variables:
```bash
DB_AUTO_MIGRATE=false
DB_MIGRATIONS_PATH=migrations
DB_USERNAME=postgres
DB_MIGRATIONS_TABLE=schema_migrations
CLICKHOUSE_USER=brokle
CLICKHOUSE_MIGRATIONS_ENGINE=MergeTree
```

## Error Handling

### Common Issues

**Migration fails partway through:**
```bash
./migrate status    # Check current state
./migrate info      # Get detailed error information
```

**Database connection issues:**
```bash
# Check configuration
./migrate info

# Try single database
./migrate -db postgres status
```

**Version conflicts:**
```bash
# Check version state
./migrate status

# Force correct version if needed (DANGEROUS)
./migrate force -version N
```

### Recovery Strategies

1. **Always check status first:** `./migrate status`
2. **Use dry-run for validation:** `./migrate -dry-run up`
3. **Single database at a time:** `./migrate -db postgres up`
4. **Step-by-step approach:** `./migrate steps -steps 1`

## Auto-Migration

For Kubernetes/Docker deployments, enable auto-migration:

```yaml
# config.yaml
database:
  auto_migrate: true
```

Or via environment:
```bash
DB_AUTO_MIGRATE=true
```

Auto-migration runs during application startup with:
- 5-minute timeout
- Comprehensive error logging  
- Graceful failure handling
- Safe for production deployments

## Best Practices

### Development
- Always create both `.up.sql` and `.down.sql` files
- Test migrations locally before deploying
- Use descriptive migration names
- Keep migrations small and focused

### Staging
- Run migrations on staging data first
- Use `--dry-run` to validate changes
- Test rollback scenarios
- Verify application compatibility

### Production
- Schedule migrations during maintenance windows
- Use `./migrate info` to check system health
- Have rollback plan ready
- Monitor database performance post-migration

### Multi-Database Considerations
- PostgreSQL migrations should complete before ClickHouse
- Consider data consistency between databases
- Plan for partial failure scenarios
- Use transactions where appropriate

## Troubleshooting

### Check System Status
```bash
./migrate info                  # Comprehensive health check
./migrate status               # Quick status overview
```

### Verbose Logging
The migration tool provides detailed logging for troubleshooting. Check application logs for:
- Connection errors
- SQL execution failures
- Version conflicts
- Lock timeouts

### Common Solutions
- **Connection refused**: Check database connectivity and credentials
- **Permission denied**: Verify database user permissions
- **Lock timeout**: Another migration process may be running
- **Version conflict**: Check migration file consistency
- **Seeding fails**: Check YAML syntax and referenced entities exist
- **API keys skipped**: Known PostgreSQL JSON serialization issue

## Architecture

The database CLI is built on:
- **golang-migrate** library for PostgreSQL migrations
- **Custom ClickHouse implementation** for analytics migrations
- **Multi-database coordination** logic
- **YAML-based seeding system** for data population
- **Production safety features** and confirmations
- **Health monitoring system**

Key components:
- `internal/migration/manager.go` - Core migration coordinator
- `internal/migration/health.go` - Health monitoring system
- `internal/seeder/manager.go` - Data seeding coordinator
- `internal/seeder/` - Component seeders (users, orgs, RBAC, etc.)
- `cmd/migrate/main.go` - Unified CLI interface
- `migrations/` - Migration file storage
- `seeds/` - YAML seed data files

## Contributing

When adding new database capabilities:

### For Migrations:
1. Update both PostgreSQL and ClickHouse implementations
2. Add comprehensive error handling
3. Include safety confirmations for destructive operations
4. Test with both databases

### For Seeding:
1. Add new seeder components to `internal/seeder/`
2. Update YAML seed data files as needed
3. Follow entity dependency ordering
4. Add validation and error handling

**Always:**
- Update this documentation
- Test thoroughly in development environment
- Follow existing CLI patterns and safety features

## Security

### For Migrations:
- Migration files should not contain sensitive data
- Use environment variables for credentials
- Limit database permissions to minimum required
- Audit migration files before production deployment

### For Seeding:
- Seed data uses bcrypt for password hashing
- Avoid sensitive data in YAML seed files
- Use environment-specific configurations
- Test data should be safe for development sharing

**General Security:**
- Consider using encrypted connections for production
- Regularly audit database access and permissions
- Monitor migration and seeding operations in production

---

**🚀 Built for Brokle - The Open-Source AI Control Plane**

This unified database CLI handles both schema migrations and data seeding for the Brokle AI Control Plane. For more information about the platform, see the main project documentation.