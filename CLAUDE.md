# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Quick Commands

| Command | Description |
|---------|-------------|
| `make setup` | First time setup (deps, DBs, migrations, seeds) |
| `make dev` | Start full stack (server + worker with hot reload) |
| `make dev-server` | HTTP server only |
| `make dev-worker` | Workers only |
| `make dev-frontend` | Next.js frontend only |
| `make test` | Run all tests |
| `make lint` | Lint all code |
| `make migrate-up` | Run database migrations |
| `make seed` | Seed system data |

## IMPORTANT: Development Rules

**MUST:**
- ALWAYS use CLI to create migrations: `go run cmd/migrate/main.go -db <postgres|clickhouse> -name <name> create`
- NEVER create migration files manually in `migrations/` directory
- Follow Repository → Service → Handler pattern for all features
- Use `AppError` constructors in services, `response.Error()` in handlers
- Use feature imports: `@/features/[feature]` (never import internal paths)

**NEVER:**
- Create documentation files unless explicitly requested
- Add tests for simple CRUD, validation, or constructors
- Log sensitive data (passwords, tokens, API keys, PII)

---

## Platform Overview

**Brokle** is an open-source AI observability platform providing tracing, evaluation, cost analytics, and prompt management for AI applications. Scalable monolith architecture with separate server and worker binaries.

## Architecture

### Deployment
- **Server** (`cmd/server`): HTTP API on port 8080, requires `JWT_SECRET`, runs migrations
- **Worker** (`cmd/worker`): Background processing, no JWT needed, processes telemetry streams
- **Databases**: PostgreSQL (transactional) + ClickHouse (analytics) + Redis (cache/queues)

### Key Directories

```
brokle/
├── cmd/                    # Entry points (server, worker, migrate)
├── internal/
│   ├── app/               # DI container, bootstrap
│   ├── core/
│   │   ├── domain/        # Entities, interfaces (6 domains)
│   │   └── services/      # Business logic
│   ├── infrastructure/
│   │   ├── repository/    # Database implementations
│   │   └── streams/       # Redis streams
│   ├── transport/http/
│   │   ├── handlers/      # HTTP handlers by domain
│   │   └── middleware/    # Auth, rate limiting
│   ├── workers/           # Background workers
│   └── ee/                # Enterprise features
├── pkg/                   # Shared packages (errors, response, utils)
├── migrations/            # PostgreSQL + ClickHouse migrations
├── seeds/                 # YAML seed data
└── web/                   # Next.js frontend
```

### Domains

| Domain | Location | Purpose |
|--------|----------|---------|
| auth | `internal/core/domain/auth` | Authentication, sessions, API keys |
| billing | `internal/core/domain/billing` | Usage tracking, billing |
| observability | `internal/core/domain/observability` | Traces, spans, quality scores |
| organization | `internal/core/domain/organization` | Multi-tenant org management |
| user | `internal/core/domain/user` | User management |
| analytics | `internal/core/domain/analytics` | Provider pricing, cost analytics |

**Full architecture**: See `docs/ARCHITECTURE.md`

## Development Commands

### Build Variants

```bash
make build-server-oss         # OSS server
make build-worker-oss         # OSS worker
make build-server-enterprise  # Enterprise (with -tags="enterprise")
make build-worker-enterprise  # Enterprise worker
```

### Testing

```bash
make test                # All tests
make test-unit           # Unit tests only
make test-integration    # Integration tests
make test-coverage       # With coverage report
go test ./internal/core/services/observability -v  # Specific package
```

### Database Operations

```bash
make migrate-up          # Run migrations
make migrate-down        # Rollback one
make migrate-status      # Check status
make create-migration DB=postgres NAME=add_users_table
make shell-db            # PostgreSQL shell
make shell-clickhouse    # ClickHouse shell
```

**Advanced CLI**: See `docs/development/MIGRATION_CLI.md`

## Environment Configuration

Copy `.env.example` to `.env`. Key variables:

| Variable | Description |
|----------|-------------|
| `APP_MODE` | `server` or `worker` |
| `PORT` | HTTP port (default: 8080) |
| `DATABASE_URL` | PostgreSQL connection |
| `REDIS_URL` | Redis connection |
| `CLICKHOUSE_URL` | ClickHouse connection |
| `JWT_SECRET` | Required for server mode |
| `AI_KEY_ENCRYPTION_KEY` | Encryption key for AI credentials (base64, 32 bytes) |

**Note**: AI API keys (OpenAI, Anthropic, etc.) are NOT configured via environment.
Configure via dashboard: Settings > AI Providers (per-project credentials).

## API Architecture

### Dual Route System

**SDK Routes** (`/v1/*`) - API Key Auth (`bk_{40_char}`):
- `POST /v1/traces` - OTLP telemetry ingestion
- `POST /v1/evaluations` - Quality evaluations

**Dashboard Routes** (`/api/v1/*`) - JWT Auth:
- `/api/v1/auth/*` - Authentication
- `/api/v1/organizations/*` - Org management
- `/api/v1/projects/*` - Project management
- `/api/v1/analytics/*` - Metrics & reporting

**Full API documentation**: See `docs/API.md`

### Authentication Context

```go
// SDK routes
ctx := middleware.GetSDKAuthContext(ctx)  // Returns project ID, API key ID

// Dashboard routes
userID := middleware.GetUserID(ctx)
```

### API Key Format

```
bk_{40_char_random_secret}
```
- SHA-256 hashed for storage (O(1) lookup)
- Validated via `X-API-Key` or `Authorization` header

## Development Patterns

### Clean Architecture Flow

```
Handler → Service → Repository → Database
   ↓         ↓           ↓
HTTP     Business    Data Access
```

### Error Handling

- Repository: Return raw errors
- Service: Wrap with `AppError` constructors (`NewNotFound`, `NewValidation`, etc.)
- Handler: Use `response.Error(ctx, err)`

**Full guide**: See `docs/development/ERROR_HANDLING_GUIDE.md`

### Required Configuration Validation

Services use **fail-fast validation** at startup. Missing required config causes immediate startup failure with clear error messages.

**To add a new required configuration:**

1. **Add `Validate()` method** to config struct in `internal/config/config.go`:
```go
func (nc *NewConfig) Validate() error {
    if os.Getenv("APP_MODE") == "worker" {
        return nil  // Skip for workers if not needed
    }
    if nc.RequiredKey == "" {
        return errors.New("NEW_REQUIRED_KEY is required")
    }
    return nil
}
```

2. **Call from `Config.Validate()`**:
```go
if err := c.NewConfig.Validate(); err != nil {
    return fmt.Errorf("new config validation failed: %w", err)
}
```

**Required configs (server mode):** `JWT_SECRET`, `AI_KEY_ENCRYPTION_KEY`, `DATABASE_URL`, `CLICKHOUSE_URL`, `REDIS_URL`

### OTEL-Native Observability

- **Single table**: `otel_traces` stores both traces and spans
- **Traces are virtual**: Derived from root spans (`WHERE parent_span_id IS NULL`)
- **Bulk processing**: Worker uses batch inserts

```go
// Query traces
traces, err := traceService.ListTraces(ctx, &TraceFilter{ProjectID: "proj123"})

// Get root span
rootSpan, err := traceService.GetRootSpan(ctx, traceID)
```

### Logging Standards

**Where to log:**

| Layer | Logging | Notes |
|-------|---------|-------|
| Services | ✅ Direct logging | Use `*slog.Logger` field |
| Handlers | ❌ No logging | Use `response.Error()` only |
| Repositories | ❌ No logging | Return errors to service |
| Workers | ✅ Direct logging | Background job status |

**Exception:** Auth service uses `audit_decorator.go` for security compliance.

**Log levels:**

| Level | When to Use | Example |
|-------|-------------|---------|
| `Error` | Operation failed, needs attention | DB connection failed, external API error |
| `Warn` | Unexpected but handled | Cache miss, retry succeeded, deprecated usage |
| `Info` | Significant business events | Created, deleted, payment processed |
| `Debug` | Troubleshooting (not in prod) | Request details, intermediate state |

**Required context** - Always include relevant IDs:

```go
import "log/slog"

// Good: structured with context
s.logger.Info("prompt created",
    "prompt_id", prompt.ID,
    "project_id", projectID,
    "user_id", userID,
)

// Good: error with context
s.logger.Error("failed to create prompt",
    "error", err,
    "project_id", projectID,
    "name", req.Name,
)

// Bad: no context
s.logger.Info("created")  // Created what?
```

**Never log:** passwords, tokens, API keys, full request bodies, PII without masking.

## Testing Strategy

**Philosophy**: Test business logic, not framework behavior.

**DO test:**
- Complex business logic and calculations
- Batch operations and orchestration
- Error handling patterns
- Analytics and aggregations

**DON'T test:**
- Simple CRUD without business logic
- Field validation (already in domain layer)
- Trivial constructors and getters

**Full guide**: See `docs/TESTING.md`

## Frontend Architecture

**Stack**: Next.js 15.5.2, React 19.2.0, TypeScript 5.9.3, Tailwind CSS 4.1.15

```bash
cd web && pnpm install     # Install deps
pnpm dev                   # Dev server on :3000
pnpm build                 # Production build
pnpm test                  # Run tests
```

### Feature-Based Structure

```
web/src/
├── app/                   # Next.js App Router (routing only)
├── features/              # Domain features (self-contained)
│   ├── authentication/    # Auth domain
│   ├── organizations/     # Org management
│   ├── projects/          # Project dashboard
│   └── ...
├── components/ui/         # shadcn/ui primitives
└── lib/api/core/          # BrokleAPIClient
```

**Full guide**: See `web/ARCHITECTURE.md`

## Enterprise Edition

Build with `-tags="enterprise"` for:
- SSO (SAML 2.0, OIDC/OAuth2)
- RBAC (scope-based permissions)
- Compliance (SOC 2, HIPAA, GDPR)
- Enterprise analytics

Features in `internal/ee/`. **Guide**: See `docs/ENTERPRISE.md`

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Port 8080 in use | `lsof -ti:8080 \| xargs kill` |
| Migration dirty state | `go run cmd/migrate/main.go -db <db> drop` then re-migrate |
| Enterprise features missing | Ensure `-tags="enterprise"` build flag |
| API key validation | `curl -X POST localhost:8080/v1/auth/validate-key -H "Content-Type: application/json" -d '{"api_key": "bk_..."}'` |

## Additional Documentation

- `docs/ARCHITECTURE.md` - System architecture
- `docs/API.md` - API reference
- `docs/DEVELOPMENT.md` - Development guide
- `docs/TESTING.md` - Testing philosophy
- `docs/ENTERPRISE.md` - Enterprise features
- `docs/development/ERROR_HANDLING_GUIDE.md` - Error patterns
- `docs/development/MIGRATION_CLI.md` - Migration CLI reference
