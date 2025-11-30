# 🚀 Brokle Database CLI

Production-ready database migration and seeding CLI for the Brokle AI Control Plane. Handles multi-database migrations and system data seeding with comprehensive safety features.

## Overview

The Brokle Database CLI manages database operations across multiple databases:

### 🔄 **Schema Migrations**
- **PostgreSQL** - Primary transactional data
- **ClickHouse** - Analytics and time-series data
- **Multi-database coordination** - Ensures consistency across all databases

### 🌱 **System Data Seeding**
- **Permissions** - 63 system permissions
- **Roles** - 4 role templates (owner, admin, developer, viewer)
- **Pricing** - 20 AI models, 78 prices (OpenAI, Anthropic, Google)

## Quick Start

```bash
# Build the database CLI
go build -o migrate ./cmd/migrate/main.go

# Run all pending migrations
./migrate up

# Seed system data (permissions, roles, pricing)
./migrate seed

# Check migration status
./migrate status

# Create new migration
./migrate create -name "add_users_table" -db postgres
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

#### `migrate create -name NAME -db DB`
Create new migration files.
```bash
./migrate create -name "add_users_table" -db postgres
./migrate create -name "add_analytics_events" -db clickhouse
```

#### `migrate info`
Show detailed migration information and health status.
```bash
./migrate info                  # Comprehensive system information
```

### 🌱 Data Seeding Commands

#### `migrate seed`
Seed system template data (permissions, roles, pricing).
```bash
./migrate seed                  # Seed all system data
./migrate seed -verbose         # With detailed output
./migrate seed -reset           # Reset existing data and reseed
./migrate seed -dry-run         # Preview seeding plan
```

#### `migrate seed-rbac`
Seed only RBAC data (permissions and roles).
```bash
./migrate seed-rbac             # Seed permissions and roles
./migrate seed-rbac -verbose    # With detailed output
./migrate seed-rbac -reset      # Reset and reseed
```

#### `migrate seed-pricing`
Seed only provider pricing data.
```bash
./migrate seed-pricing          # Seed AI model pricing
./migrate seed-pricing -verbose # With detailed output
./migrate seed-pricing -reset   # Reset and reseed
```

**Seeding includes:**
- **Permissions** - 63 system permissions (resource:action format)
- **Roles** - 4 role templates with permission assignments
- **Provider Pricing** - 20 AI models with 78 price entries

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

## File Structure

### Migration Files

#### PostgreSQL Migrations
Located in `migrations/postgres/`
```
migrations/postgres/
├── 20240101000000_initial_schema.up.sql
├── 20240101000000_initial_schema.down.sql
├── 20240102000000_add_users.up.sql
└── 20240102000000_add_users.down.sql
```

#### ClickHouse Migrations
Located in `migrations/clickhouse/`
```
migrations/clickhouse/
├── 20240101000000_initial_analytics.up.sql
├── 20240101000000_initial_analytics.down.sql
├── 20240102000000_add_events.up.sql
└── 20240102000000_add_events.down.sql
```

### Seed Data Files
Located in `seeds/` directory with YAML format:
```
seeds/
├── permissions.yaml  # 63 system permissions
├── roles.yaml        # 4 role templates (owner, admin, developer, viewer)
└── pricing.yaml      # 20 AI models, 78 prices
```

#### Seed Data Structure

**permissions.yaml:**
```yaml
permissions:
  - name: "organizations:read"
    description: "View organization details"
  - name: "projects:write"
    description: "Create and update projects"
```

**roles.yaml:**
```yaml
roles:
  - name: "owner"
    scope_type: "organization"
    permissions:
      - "organizations:read"
      - "organizations:write"
      # ... 63 permissions
```

**pricing.yaml:**
```yaml
provider_models:
  - model_name: "gpt-4o"
    match_pattern: "^gpt-4o$"
    start_date: "2024-05-13"
    prices:
      - usage_type: "input"
        price: 2.50
      - usage_type: "output"
        price: 10.00
```

## Usage Examples

### Development Workflow - Migrations
```bash
# 1. Create new migration
./migrate create -name "add_api_keys_table" -db postgres

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

# 3. Seed system data
./migrate seed -verbose

# 4. Or reset and reseed for clean state
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

# 4. Seed system data (if needed)
./migrate seed

# 5. Verify completion
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

## Configuration

The migration tool uses environment variables:

```bash
DATABASE_URL=postgres://user:pass@localhost:5432/brokle
CLICKHOUSE_URL=clickhouse://localhost:9000/brokle
DB_AUTO_MIGRATE=false
DB_MIGRATIONS_PATH=migrations
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
./migrate info      # Check configuration
./migrate -db postgres status  # Try single database
```

**Version conflicts:**
```bash
./migrate status    # Check version state
./migrate force -version N  # Force correct version if needed (DANGEROUS)
```

## Best Practices

### Development
- Always create both `.up.sql` and `.down.sql` files
- Test migrations locally before deploying
- Use descriptive migration names
- Keep migrations small and focused

### Production
- Schedule migrations during maintenance windows
- Use `./migrate info` to check system health
- Have rollback plan ready
- Run `./migrate seed` after initial deployment

## Architecture

The database CLI is built on:
- **golang-migrate** library for PostgreSQL migrations
- **Custom ClickHouse implementation** for analytics migrations
- **YAML-based seeding system** for system template data
- **Production safety features** and confirmations

Key components:
- `internal/migration/manager.go` - Core migration coordinator
- `internal/seeder/seeder.go` - Unified seeder implementation
- `cmd/migrate/main.go` - CLI interface
- `migrations/` - Migration file storage
- `seeds/` - YAML seed data files

---

**🚀 Built for Brokle - The Open-Source AI Control Plane**
