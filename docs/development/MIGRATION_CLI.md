# Database Migration CLI Reference

Complete reference for the Brokle database migration CLI.

## Quick Reference

```bash
# Common operations
make migrate-up          # Run all migrations
make migrate-down        # Rollback one migration
make migrate-status      # Check status
make seed                # Seed system data
```

## CLI Usage

The migration CLI supports granular database control:

```bash
go run cmd/migrate/main.go [flags] <command>
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-db` | Target database: `postgres`, `clickhouse`, `all` | `all` |
| `-name` | Migration name (for `create` command) | - |
| `-steps` | Number of migrations to run | all |
| `-version` | Force to specific version (dangerous) | - |
| `-dry-run` | Preview without executing | false |
| `-verbose` | Verbose output | false |
| `-reset` | Reset before seeding | false |

### Commands

| Command | Description |
|---------|-------------|
| `up` | Run pending migrations |
| `down` | Rollback migrations |
| `status` | Show migration status |
| `create` | Create new migration files |
| `drop` | Drop all tables (destructive) |
| `force` | Force version (use when dirty) |
| `info` | Detailed migration info |
| `seed` | Seed all system data |
| `seed-rbac` | Seed permissions and roles |
| `seed-pricing` | Seed provider pricing |

## Migration Operations

### Run Migrations

```bash
# All databases
go run cmd/migrate/main.go up

# Specific database
go run cmd/migrate/main.go -db postgres up
go run cmd/migrate/main.go -db clickhouse up

# Run specific number
go run cmd/migrate/main.go -db postgres -steps 2 up

# Preview only
go run cmd/migrate/main.go -dry-run up
```

### Rollback Migrations

```bash
# Rollback one migration
go run cmd/migrate/main.go -db postgres down

# Rollback specific number
go run cmd/migrate/main.go -db postgres -steps 3 down

# Rollback all (use with caution)
go run cmd/migrate/main.go down -steps 999
```

### Check Status

```bash
# Both databases with health check
go run cmd/migrate/main.go status

# Specific database
go run cmd/migrate/main.go -db postgres status
go run cmd/migrate/main.go -db clickhouse status
```

### Create New Migrations

**IMPORTANT**: Always use CLI to create migrations. Never create files manually.

```bash
# PostgreSQL migration
go run cmd/migrate/main.go -db postgres -name create_users_table create

# ClickHouse migration
go run cmd/migrate/main.go -db clickhouse -name create_metrics_table create
```

This generates properly named files with timestamps:
- `migrations/postgres/YYYYMMDDHHMMSS_create_users_table.up.sql`
- `migrations/postgres/YYYYMMDDHHMMSS_create_users_table.down.sql`

## Seeding System Data

System template data is managed via YAML seeds in `/seeds/`:

```bash
# Seed all (permissions, roles, pricing)
go run cmd/migrate/main.go seed

# With verbose output
go run cmd/migrate/main.go seed -verbose

# Reset and reseed
go run cmd/migrate/main.go seed -reset

# Seed specific data
go run cmd/migrate/main.go seed-rbac           # Permissions + roles
go run cmd/migrate/main.go seed-pricing        # Provider pricing
go run cmd/migrate/main.go seed-pricing -reset -verbose
```

### Seed Files

| File | Contents |
|------|----------|
| `seeds/permissions.yaml` | 63 system permissions |
| `seeds/roles.yaml` | 4 role templates (owner, admin, developer, viewer) |
| `seeds/pricing.yaml` | 20 AI models, 78 prices (OpenAI, Anthropic, Google) |

## Destructive Operations

These commands require confirmation:

```bash
# Drop all tables
go run cmd/migrate/main.go drop
go run cmd/migrate/main.go -db postgres drop
go run cmd/migrate/main.go -db clickhouse drop

# Force version (use when migration state is dirty)
go run cmd/migrate/main.go -db postgres -version 0 force
```

## Safety Features

- Confirmation prompts for destructive operations
- Dry run mode with `-dry-run` flag
- Granular database control via `-db` flag
- Health monitoring with dirty state detection

## Database Schema Overview

### PostgreSQL Tables
- `users`, `auth_sessions` - Authentication
- `organizations`, `organization_members` - Multi-tenant
- `projects`, `api_keys` - Project management
- `gateway_*` - AI provider configs
- `billing_usage` - Usage tracking
- `permissions`, `roles`, `role_permissions` - RBAC
- `provider_models`, `model_prices` - Pricing

### ClickHouse Tables
- `otel_traces` - OTLP traces and spans (365 day TTL)
- `quality_scores` - Model performance metrics
- `request_logs` - API request logs (60 day TTL)
