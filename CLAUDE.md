# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Quick Commands

| Command | Description |
|---------|-------------|
| `make setup` | First time setup (deps, DBs, migrations, seeds, codegen) |
| `make install-tools` | Install dev tools (swag, air, golangci-lint) |
| `make generate` | Run go generate (swagger docs, etc.) |
| `make dev` | Start full stack (server + worker with hot reload) |
| `make dev-server` | HTTP server only |
| `make dev-worker` | Workers only |
| `make dev-frontend` | Next.js frontend only |
| `make test` | Run all tests (business logic) |
| `make test-all` | Run all tests including cmd/ (requires docs) |
| `make lint` | Lint all code |
| `make migrate-up` | Run database migrations |
| `make seed` | Seed system data |

## Development Tools

**Automated Installation:**
- `make install-tools` installs all required development tools
- Called automatically by `make setup`
- Tools are version-pinned via go.mod (swag) or latest (air)

**Code Generation (go generate):**
- Run `make generate` or `go generate ./...` after changing API annotations
- Swagger docs use `//go:generate` directive in `cmd/server/main.go`
- First-time setup runs this automatically via `make setup`

**Manual Installation (if needed):**
```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.6  # API documentation
go install github.com/air-verse/air@latest           # Hot reload (dev only)
```

**Tool Versions:**
- **swag**: v1.16.6 (tracked in go.mod via tools.go)
- **air**: latest (optional, dev-only hot reload)
- **golangci-lint**: v2.6.2 (Go 1.25 compatible)

**Troubleshooting:**
- `swag: command not found` → Should auto-install, or run `make install-tools`
- `air: command not found` → Run `make install-tools`
- Ensure `$GOPATH/bin` is in your `$PATH` (usually `~/go/bin`)

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
- Edit generated Swagger files in `docs/` (edit Go annotations in `cmd/server/main.go` and handlers instead)

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
make test                # All tests (requires docs: run make generate first)
make test-unit           # Unit tests only
make test-integration    # Integration tests
make test-coverage       # With coverage report
go test ./internal/core/services/observability -v  # Specific package
```

**Note**: Tests run on all packages (`./...`). Since `cmd/server` imports `brokle/docs`, run `make generate` after setup or API changes. If you see "cannot find package brokle/docs", run `make generate`.

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

### Transaction Management

**Pattern**: Context-Based Transactions (idiomatic Go)

**Usage in Services:**
```go
func (s *service) CreateWithTransaction(ctx context.Context, req *Request) error {
    return s.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
        // All repository calls within this function use the transaction
        if err := s.repo1.Create(ctx, entity); err != nil {
            return err // Auto-rollback on error
        }
        return s.repo2.Update(ctx, other) // Auto-commit on success
    })
}
```

**How It Works:**
1. `Transactor.WithinTransaction` begins a database transaction
2. Transaction is injected into the context
3. Repositories extract the transaction from context using `getDB(ctx)` helper
4. Automatic commit on `nil` return, rollback on error or panic

**Repository Pattern:**
```go
// Each repository has a getDB helper that extracts transaction from context
func (r *repository) getDB(ctx context.Context) *gorm.DB {
    return shared.GetDB(ctx, r.db)
}

// All repository methods use getDB(ctx) instead of r.db
func (r *repository) Create(ctx context.Context, entity *Entity) error {
    return r.getDB(ctx).WithContext(ctx).Create(entity).Error
}
```

**Testing:**
```go
// Simple mock transactor for tests
transactor := NewMockTransactor()
service := NewService(transactor, repo1, repo2, logger)
```

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

### Swagger/OpenAPI Documentation

**Generated at Build Time**: Swagger docs (`docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`) are **NOT tracked in Git**. They are auto-generated during development and build processes.

**How it works:**
- **Development**: `make dev-server` auto-generates docs before starting server
- **Production**: Dockerfile generates docs during Docker build
- **CI/CD**: GitHub Actions generates docs before running tests and builds
- **Manual**: Run `make docs-generate` to regenerate

**Why build-time generation?**
- Clean diffs (no 83K line generated file changes in PRs)
- No merge conflicts from generated code
- Always up-to-date docs (can't forget to regenerate)
- Follows OSS best practices (similar to Grafana's approach)

**To update API documentation:**
1. Edit Go annotations in `cmd/server/main.go` (API metadata)
2. Edit handler function comments with `@Summary`, `@Description`, `@Param`, etc.
3. Run `make docs-generate` or let build process handle it
4. Access Swagger UI at `http://localhost:8080/swagger/index.html`

**Example handler annotation:**
```go
// CreateProject creates a new project
// @Summary Create new project
// @Description Create a new project within an organization
// @Tags projects
// @Accept json
// @Produce json
// @Param request body CreateProjectRequest true "Project details"
// @Success 201 {object} ProjectResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/projects [post]
// @Security CookieAuth
func (h *Handler) CreateProject(ctx *gin.Context) { ... }
```

**IMPORTANT**: Never manually edit files in `docs/` directory - they are generated and gitignored.

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
