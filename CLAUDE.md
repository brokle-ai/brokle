# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Platform Overview

**Brokle** is an open-source AI control plane providing observability, routing, and governance for AI applications. The platform combines AI gateway functionality with comprehensive telemetry tracking and cost optimization.

## Architecture: Scalable Monolith with Independent Scaling

This project represents a **scalable monolith** architecture with **separate server and worker binaries** for independent scaling, migrated from microservices for better development velocity and open source adoption.

### High-Level Architecture
- **Separate Binaries**: HTTP server + Background workers (independent scaling)
- **Shared Codebase**: Single codebase with modular DI container
- **Multi-Database Strategy**: PostgreSQL (transactional) + ClickHouse (analytics) + Redis (cache/queues)
- **Domain-Driven Design** with clean separation of concerns
- **Enterprise Edition** toggle via build tags (`go build -tags="enterprise"`)
- **Transport Layer Pattern** for HTTP/WebSocket handling

### Deployment Architecture

**Server Process** (`cmd/server/main.go`):
- Set `APP_MODE=server` in environment
- HTTP API endpoints and WebSocket connections
- Runs database migrations (server owns migrations)
- Requires JWT_SECRET for authentication
- Port: 8080
- Scales independently (3-5 instances typical)

**Worker Process** (`cmd/worker/main.go`):
- Set `APP_MODE=worker` in environment
- Telemetry stream processing from Redis
- Gateway analytics aggregation
- Batch job processing
- No JWT_SECRET needed (doesn't handle auth)
- Scales independently (10-50+ instances at scale)

**Shared Infrastructure**:
- Same database connections (PostgreSQL, ClickHouse, Redis)
- Same service layer (reused via DI container)
- Different resource profiles (APIs: low latency, Workers: high throughput)

### Application Structure
```
brokle/
├── cmd/                    # Application entry points
│   ├── server/main.go     # HTTP server (API endpoints)
│   ├── worker/main.go     # Background workers (telemetry processing)
│   └── migrate/main.go    # Database migration runner & seeder
├── internal/              # Private application code
│   ├── app/               # Application bootstrap & DI container
│   ├── config/            # Configuration management
│   ├── core/              # Domain-driven design architecture
│   │   ├── domain/        # Domain layer (entities, interfaces) - 10 domains
│   │   │   ├── auth/      # Authentication domain
│   │   │   ├── billing/   # Billing domain
│   │   │   ├── common/    # Common patterns (transaction management)
│   │   │   ├── config/    # Configuration domain (planned)
│   │   │   ├── gateway/   # AI Gateway domain
│   │   │   ├── observability/ # Observability domain
│   │   │   ├── organization/  # Organization domain
│   │   │   ├── routing/   # Routing domain (planned)
│   │   │   └── user/      # User domain
│   │   └── services/      # Service implementations (matches domains)
│   │       ├── auth/      # Auth services
│   │       ├── billing/   # Billing services
│   │       ├── gateway/   # Gateway services
│   │       ├── observability/ # Observability services
│   │       ├── organization/  # Organization services
│   │       ├── registration/  # Registration orchestration
│   │       └── user/      # User services
│   ├── infrastructure/    # External integrations
│   │   ├── database/      # DB connections
│   │   │   ├── clickhouse/repository/ # ClickHouse-specific repos
│   │   │   └── postgres/repository/   # Postgres-specific repos
│   │   ├── repository/    # Main repository layer
│   │   └── streams/       # Redis streams for telemetry
│   ├── transport/http/    # HTTP transport layer
│   │   ├── handlers/      # HTTP handlers by domain
│   │   ├── middleware/    # HTTP middleware
│   │   └── server.go      # HTTP server setup
│   ├── middleware/        # Shared middleware
│   │   └── enterprise.go  # Enterprise feature gating
│   ├── workers/           # Background workers
│   │   ├── analytics_worker.go
│   │   ├── notification_worker.go
│   │   └── telemetry_stream_consumer.go
│   └── ee/                # Enterprise Edition features
│       ├── license/       # License validation service
│       ├── sso/           # Single Sign-On
│       ├── rbac/          # Role-Based Access Control
│       ├── compliance/    # Compliance features
│       └── analytics/     # Enterprise analytics
├── pkg/                   # Public shared packages
│   ├── errors/            # Error handling
│   ├── response/          # HTTP response utilities
│   ├── utils/             # Common utilities
│   └── websocket/         # WebSocket utilities
├── migrations/            # Database migration files (27 PostgreSQL, 4 ClickHouse)
├── seeds/                 # YAML-based seeding data
│   ├── dev.yaml          # Development seed data
│   ├── demo.yaml         # Demo seed data
│   └── test.yaml         # Test seed data
├── web/                   # Next.js 15.5.2 frontend
└── docs/                  # Public OSS documentation
```


### Domain-Driven Architecture

#### Current Domains (10 total)
| Domain | Status | Location | Purpose |
|--------|--------|----------|---------|
| auth | ✅ Active | `internal/core/domain/auth` | Authentication, sessions, API keys |
| billing | ✅ Active | `internal/core/domain/billing` | Usage tracking, billing |
| common | ✅ Active | `internal/core/domain/common` | Transaction patterns, shared utilities |
| config | 🔄 Planned | `internal/core/domain/config` | Configuration management |
| gateway | ✅ Active | `internal/core/domain/gateway` | AI provider routing |
| observability | ✅ Active | `internal/core/domain/observability` | Traces, observations, quality scores |
| organization | ✅ Active | `internal/core/domain/organization` | Multi-tenant org management |
| routing | 🔄 Planned | `internal/core/domain/routing` | Advanced routing logic |
| user | ✅ Active | `internal/core/domain/user` | User management |

#### Layer Organization
- **Domain Layer** (`internal/core/domain/`): Entities, repository interfaces, service interfaces
- **Service Layer** (`internal/core/services/`): Business logic implementations matching domains
- **Infrastructure Layer** (`internal/infrastructure/`): Database repos (3-tier: main → DB-specific → implementations), external clients, Redis streams
- **Transport Layer** (`internal/transport/http/`): HTTP handlers, middleware, WebSocket
- **Application Layer** (`internal/app/`): DI container, service registry, graceful shutdown

## Development Commands

### Essential Commands
```bash
# First time setup (installs deps, starts DBs, runs migrations, seeds data)
make setup

# Start full development stack (Server + Worker)
make dev              # Starts both server and worker with hot reload

# Start individual components
make dev-server       # HTTP server only (with hot reload)
make dev-worker       # Workers only (with hot reload)
make dev-frontend     # Next.js frontend only

# Run all tests
make test

# Build for development (faster builds)
make build-dev-server    # Build server for development
make build-dev-worker    # Build worker for development

# Build production binaries
make build-server-oss         # HTTP server (OSS)
make build-worker-oss         # Workers (OSS)
make build-server-enterprise  # HTTP server (Enterprise)
make build-worker-enterprise  # Workers (Enterprise)
make build-all               # Build all variants
```

### Production Deployment

```bash
# Server (3-5 instances for API traffic)
./bin/brokle-server
# Runs migrations on startup (if DB_AUTO_MIGRATE=true)
# Serves HTTP on :8080

# Worker (10-50+ instances for background processing)
./bin/brokle-worker
# Processes telemetry streams from Redis
# No HTTP server, no migrations

# Docker Compose Example
services:
  server:
    image: brokle-server:latest
    replicas: 3
    ports:
      - "8080:8080"
    environment:
      - DB_AUTO_MIGRATE=true  # First server instance runs migrations

  worker:
    image: brokle-worker:latest
    replicas: 10
    environment:
      - DB_AUTO_MIGRATE=false  # Workers never run migrations
```

### Database Operations

#### Quick Commands (Make)
```bash
# Run all database migrations
make migrate-up

# Rollback one migration
make migrate-down

# Check migration status
make migrate-status

# Create new migration
make create-migration DB=postgres NAME=add_users_table
make create-migration DB=clickhouse NAME=add_metrics_table

# Reset all databases (WARNING: destroys data)
make migrate-reset

# Seed with development data
make seed-dev

# Database shell access
make shell-db          # PostgreSQL
make shell-redis       # Redis CLI
make shell-clickhouse  # ClickHouse client
```

#### Advanced Migration CLI
The platform includes a comprehensive migration CLI with granular database control:

```bash
# Run migrations for specific databases
go run cmd/migrate/main.go -db postgres up      # PostgreSQL only
go run cmd/migrate/main.go -db clickhouse up    # ClickHouse only  
go run cmd/migrate/main.go up                    # All databases (default)

# Check detailed migration status
go run cmd/migrate/main.go -db postgres status
go run cmd/migrate/main.go -db clickhouse status
go run cmd/migrate/main.go status               # Both databases with health check

# Rollback migrations (requires confirmation)
go run cmd/migrate/main.go -db postgres down
go run cmd/migrate/main.go -db clickhouse down
go run cmd/migrate/main.go down                 # Both databases

# Create new migrations
go run cmd/migrate/main.go -db postgres -name create_users_table create
go run cmd/migrate/main.go -db clickhouse -name create_metrics_table create

# Advanced operations (DESTRUCTIVE - requires 'yes' confirmation)
go run cmd/migrate/main.go -db postgres drop    # Drop all PostgreSQL tables
go run cmd/migrate/main.go -db clickhouse drop  # Drop all ClickHouse tables
go run cmd/migrate/main.go drop                 # Drop all tables

# Granular step control
go run cmd/migrate/main.go -db postgres -steps 2 up      # Run 2 migrations forward
go run cmd/migrate/main.go -db postgres -steps -1 down   # Rollback 1 migration

# Force operations (DANGEROUS - use only when dirty)
go run cmd/migrate/main.go -db postgres -version 0 force  # Force clean state
go run cmd/migrate/main.go -db clickhouse -version 5 force # Force to version 5

# Information and debugging
go run cmd/migrate/main.go info                  # Detailed migration information
go run cmd/migrate/main.go -dry-run up          # Preview migrations without executing
```

#### Migration Safety Features
- Confirmation prompts for destructive operations
- Dry run mode with `-dry-run` flag
- Granular database control: `-db postgres|clickhouse|all`
- Health monitoring with dirty state detection

#### Database Schema
- **PostgreSQL**: User auth, organizations, projects, API keys, gateway config, billing
- **ClickHouse**: Traces, observations, quality_scores, request_logs (with TTL retention)
- **Seeding**: YAML files in `/seeds/` (dev.yaml, demo.yaml, test.yaml)

### Testing & Quality
```bash
# Run all tests
make test

# Run unit tests only (excludes integration)
make test-unit

# Run integration tests with real databases
make test-integration

# Run with coverage report
make test-coverage

# Run load tests
make test-load

# End-to-end tests
make test-e2e

# Lint all code
make lint

# Lint Go code only
make lint-go

# Lint frontend code only
make lint-frontend

# Security scanning
make security-scan

# Format all code
make fmt
make fmt-frontend
```

### Build Variants
```bash
# OSS build (default)
make build-oss

# Enterprise build with all features
make build-enterprise

# Development builds (faster, with debug info)
make build-dev-oss
make build-dev-enterprise
```

## Environment Configuration

Copy `.env.example` to `.env` and configure:

### Application Mode
- `APP_MODE` - Deployment mode (`server` or `worker`)
  - **server**: API server mode (requires JWT_SECRET, runs migrations, handles HTTP)
  - **worker**: Background worker mode (no JWT needed, processes telemetry streams)
  - Default: `server`

### Core Services
- `PORT` - HTTP server port (default: 8080, server mode only)
- `DATABASE_URL` - PostgreSQL connection
- `REDIS_URL` - Redis connection
- `CLICKHOUSE_URL` - ClickHouse analytics database

### AI Providers
- `OPENAI_API_KEY` - OpenAI integration
- `ANTHROPIC_API_KEY` - Anthropic integration
- `GOOGLE_AI_API_KEY` - Google AI integration
- `COHERE_API_KEY` - Cohere integration

### Business Integrations
- `STRIPE_SECRET_KEY` - Payment processing
- `SENDGRID_API_KEY` - Email notifications
- `TWILIO_AUTH_TOKEN` - SMS notifications

### Blob Storage (S3/MinIO)
**Note**: LLM input/output data is stored directly in ClickHouse with ZSTD compression for optimal cost and performance (78% cheaper than S3 auto-offload). Blob storage is available for future features: scheduled exports, media files, raw event storage, etc.

- `BLOB_STORAGE_PROVIDER` - Storage provider (default: "minio")
- `BLOB_STORAGE_BUCKET_NAME` - Bucket name (default: "brokle")
- `BLOB_STORAGE_REGION` - AWS region (default: "us-east-1")
- `BLOB_STORAGE_ENDPOINT` - Endpoint URL (default: "http://localhost:9100" for MinIO)
- `BLOB_STORAGE_ACCESS_KEY_ID` - Access key ID (MinIO default: "minioadmin")
- `BLOB_STORAGE_SECRET_ACCESS_KEY` - Secret access key (MinIO default: "minioadmin")
- `BLOB_STORAGE_USE_PATH_STYLE` - Use path-style URLs (default: true for MinIO)
- `BLOB_STORAGE_THRESHOLD` - **Deprecated** (no longer used for auto-offload)

## API Architecture

### Dual Route Architecture
The platform implements a clean separation between SDK and Dashboard routes:

#### SDK Routes (`/v1/*`) - API Key Authentication
**Authentication**: Industry-standard API keys (`bk_{40_char_random}`)
**Rate Limiting**: API key-based rate limiting
**Target Users**: SDK integration, programmatic access

- `POST /v1/chat/completions` - OpenAI-compatible chat completions
- `POST /v1/completions` - OpenAI-compatible text completions
- `POST /v1/embeddings` - OpenAI-compatible embeddings
- `GET /v1/models` - Available AI models
- `GET /v1/models/:model` - Specific model details

**AI Routing**:
- `POST /v1/route` - AI routing decisions

**Telemetry Ingestion (OpenTelemetry Protocol)**:
- `POST /v1/otlp/traces` - OTLP traces ingestion (primary endpoint)
- `POST /v1/traces` - Alternative OTLP traces endpoint
- Supports: Protobuf (binary) and JSON payloads
- Compression: Automatic gzip decompression via Content-Encoding header
- Handler: `OTLPHandler` in `internal/transport/http/handlers/observability/otlp.go`
- Converter: `OTLPConverterService` with intelligent root span detection
- Processing: Events consumed via Redis streams by `TelemetryStreamConsumer` worker

**Cache Management**:
- `GET /v1/cache/status` - Cache health status
- `POST /v1/cache/invalidate` - Cache invalidation

**SDK Authentication**:
- `POST /v1/auth/validate-key` - API key validation (public endpoint)

#### Dashboard Routes (`/api/v1/*`) - JWT Authentication
**Authentication**: Bearer JWT tokens with session management
**Rate Limiting**: IP-based and user-based rate limiting
**Target Users**: Web dashboard, administrative access

- `/api/v1/auth/*` - Authentication & session management
- `/api/v1/users/*` - User profile management
- `/api/v1/organizations/*` - Organization management with RBAC
- `/api/v1/projects/*` - Project management (supports organization_id filtering)
- `/api/v1/projects/:projectId/api-keys/*` - API key management
- `/api/v1/analytics/*` - Metrics & reporting (read-only dashboard views)
- `/api/v1/logs/*` - Request logs and export
- `/api/v1/billing/*` - Usage & billing management
- `/api/v1/rbac/*` - Role and permission management
- `/api/v1/admin/*` - Administrative token management

### Real-time APIs
- `/ws` - WebSocket connections for real-time updates
- `/api/v1/streaming/*` - Server-sent events

## SDK Architecture & Authentication

### Industry-Standard API Key System
The platform implements industry-standard API keys for SDK authentication (following GitHub/Stripe/OpenAI patterns):

#### API Key Format
```
bk_{40_char_random_secret}
```
- **Prefix**: `bk_` (Brokle identifier)
- **Secret**: 40 characters of cryptographically secure random data (alphanumeric)
- **Example**: `bk_AbCdEfGhIjKlMnOpQrStUvWxYz0123456789AbCd`

#### Key Features
- **Industry Standard**: Pure random format matching GitHub, Stripe, OpenAI
- **Secure Storage**: SHA-256 hashing (deterministic, enables O(1) lookup)
- **O(1) Validation**: Direct hash lookup with unique database index
- **Project Association**: Project ID stored in database, not embedded in key
- **Environment Support**: Environment tags sent in request body (not headers)
- **Security Best Practice**: No sensitive data embedded in key

#### SDK Authentication Flow
1. Extract API key from `X-API-Key` or `Authorization` header
2. Validate format: `bk_{40_char_random}`
3. SHA-256 hash lookup in database (O(1) performance)
4. Store authentication context with project ID

### Middleware Architecture

**Two middleware locations:**
1. **HTTP Middleware** (`internal/transport/http/middleware/`): auth.go, sdk_auth.go, rate_limit.go, csrf.go, scope_middleware.go
2. **Enterprise Middleware** (`internal/middleware/`): enterprise.go (feature gating)

**SDKAuthMiddleware** (`internal/transport/http/middleware/sdk_auth.go`):
- Validates API keys for SDK routes
- Stores authentication context in request
- Context keys: `SDKAuthContextKey`, `APIKeyIDKey`, `ProjectIDKey`, `EnvironmentKey`

#### Rate Limiting Strategy
```go
// API key-based rate limiting for SDK routes
router.Use(rateLimitMiddleware.RateLimitByAPIKey())

// IP-based rate limiting for dashboard routes
router.Use(rateLimitMiddleware.RateLimitByIP())

// User-based rate limiting after JWT authentication
protectedRoutes.Use(rateLimitMiddleware.RateLimitByUser())
```

### Server Route Setup
Routes are configured in `internal/transport/http/server.go`:
- **SDK routes** (`/v1`): API key auth + API key rate limiting
- **Dashboard routes** (`/api/v1`): JWT auth + IP/user rate limiting

## Data Architecture

### Primary Database (PostgreSQL)
Single database with domain-separated tables:
- `users`, `auth_sessions` - Authentication & user management
- `organizations`, `organization_members` - Multi-tenant structure
- `projects` - Project management (environments now handled as tags)
- `api_keys` - Project-scoped API key management
- `gateway_*` - AI provider configurations
- `billing_usage` - Usage tracking and billing

### Analytics Database (ClickHouse)
Time-series data optimized for analytical queries with TTL-based retention:

#### Observability Tables
- **observations** - LLM call observations with request/response data
  - `attributes` (String) - All OTEL + Brokle attributes with namespace prefixes
  - `metadata` (String) - OTEL metadata (resourceAttributes + instrumentation scope)
  - `version` (Nullable(String)) - Application version for A/B testing
  - ZSTD compression for input/output fields (78% cost reduction vs S3)
  - **OTEL-native**: Brokle data stored in attributes with `brokle.*` namespace

- **traces** - Distributed tracing data (30 day TTL)
  - `attributes` (String) - All OTEL + Brokle attributes with namespace prefixes
  - `metadata` (String) - OTEL metadata (resourceAttributes + scope)
  - `version` (Nullable(String)) - Application version for experiment tracking
  - Hierarchical trace organization with parent/child relationships
  - **OTEL-native**: Follows OpenTelemetry standard

- **quality_scores** - Model performance metrics
  - Links to traces and observations for evaluation context

- **request_logs** - API request logging (60 day TTL)

#### Performance Metrics
- AI routing decision metrics (365 day TTL)
- Performance and latency tracking
- Cost optimization data
- ML model performance scores

### Cache Layer (Redis)  
- JWT token and session storage
- Rate limiting counters
- Background job queues (analytics, notifications)
- Semantic cache for AI responses
- Real-time event pub/sub for WebSocket

## Development Patterns

### API Key Management
**Utilities** (`internal/core/domain/auth/apikey_utils.go`):
- `GenerateAPIKey()` - Creates new `bk_{40_char}` key
- `ValidateAPIKeyFormat()` - Validates key format
- `CreateKeyPreview()` - Creates secure preview (`bk_AbCd...XyZa`)
- SHA-256 hashing for secure storage

### Authentication Patterns

**SDK Routes**: Use `middleware.GetSDKAuthContext()` to access API key context
**Dashboard Routes**: Use `middleware.GetUserID()` for JWT-authenticated requests
**Context Access**: Project ID, environment, and organization available via middleware helpers

### Clean Architecture
Follow the repository → service → handler pattern as described in the Domain-Driven Architecture section above.

### Error Handling

**📖 See Error Handling Guides in `docs/development/`**:
- `ERROR_HANDLING_GUIDE.md` - Complete patterns
- `DOMAIN_ALIAS_PATTERNS.md` - Import patterns
- `ERROR_HANDLING_QUICK_REFERENCE.md` - Quick reference

**Key Points**:
- Repository → Service → Handler error flow
- Use AppError constructors in services
- Centralized `response.Error()` in handlers
- No logging in core services (use decorators)

### Logging
Use structured logging with correlation IDs:

```go
import "log/slog"

logger := slog.With(
    "request_id", middleware.GetRequestID(ctx),
    "user_id", auth.GetUserID(ctx),
)

logger.Info("user created successfully", "user_id", user.ID)
```

### Configuration
Uses Viper for configuration management. Loads from environment variables, `.env` file, and defaults.
See `.env.example` for all configuration options.

### Enterprise Edition Pattern
The codebase uses build tags for enterprise features:

```bash
# OSS build (default)
go build ./cmd/server

# Enterprise build
go build -tags="enterprise" ./cmd/server
```

Enterprise features are in `internal/ee/` with stub implementations for OSS builds:
- `internal/ee/sso/` - Single Sign-On
- `internal/ee/rbac/` - Role-Based Access Control  
- `internal/ee/compliance/` - Compliance features
- `internal/ee/analytics/` - Enterprise analytics

## Enterprise Edition

**Build Tags**: Use `-tags="enterprise"` for enterprise builds

**Features** (`internal/ee/`):
- `license/` - License validation
- `sso/` - Single Sign-On (SAML 2.0, OIDC/OAuth2)
- `rbac/` - Role-Based Access Control (scope-based, added Oct 2024)
- `compliance/` - SOC 2, HIPAA, GDPR compliance
- `analytics/` - Enterprise analytics

**Documentation**: See `/docs/ENTERPRISE.md` and `/docs/enterprise/` for detailed guides (SSO, RBAC, compliance, analytics)

## Testing Strategy

The Brokle platform follows a **pragmatic testing philosophy** that prioritizes high-value business logic tests over low-value granular tests.

### Core Testing Principle

**Test Business Logic, Not Framework Behavior**

We focus on:
- ✅ Complex business logic and calculations
- ✅ Batch operations and orchestration workflows
- ✅ Error handling patterns and retry mechanisms
- ✅ Analytics, aggregations, and metrics
- ✅ Multi-step operations with dependencies

We avoid testing:
- ❌ Simple CRUD operations without business logic
- ❌ Field validation (already in domain layer)
- ❌ Trivial constructors and getters
- ❌ Framework behavior (ULID, time.Now(), errors.Is)
- ❌ Static constant definitions

### Test Coverage Guidelines

**Target Approach:**
- Service Layer: Focus on business logic with comprehensive test coverage
- Domain Layer: Test only complex calculations and business rules
- Handler Layer: Critical workflows only (integration tests)

### Running Tests

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run with coverage
make test-coverage

# Run specific package
go test ./internal/core/services/observability -v

# Run with race detection
go test -race ./...
```

### Quick Reference

- **Detailed Guide**: See `docs/TESTING.md` for complete examples and patterns
- **AI Prompt**: See `prompts/testing.txt` for AI-assisted test generation
- **Reference Code**: See `internal/core/services/observability/*_test.go` for real examples

### Test Quality Checklist

Before committing tests:
- ✅ Table-driven test pattern with comprehensive scenarios
- ✅ Mocks implement full repository interfaces
- ✅ Tests business logic, not framework behavior
- ✅ Verifies mock expectations with `AssertExpectations()`
- ✅ Maintains ~1:1 test-to-code ratio
- ❌ No tests for simple CRUD, validation, or constructors

## Frontend Architecture

**Stack**: Next.js 15.5.2 with App Router, React 19.2.0, Tailwind CSS 4.1.15, runs on port `:3000`

### Feature-Based Architecture
The frontend uses a **feature-based structure** where each domain is self-contained:

```
web/src/
├── app/                   # Next.js App Router (routing only)
│   ├── (auth)/           # Auth route group
│   ├── (dashboard)/      # Dashboard routes
│   └── (errors)/         # Error pages
├── features/             # Domain features (self-contained)
│   ├── authentication/   # Auth domain (12 components, 4 hooks, store, API)
│   ├── organizations/    # Org management (7 components, 2 hooks, API)
│   ├── projects/         # Project dashboard (4 components, hooks, store, API)
│   ├── analytics/        # Analytics & metrics
│   ├── billing/          # Usage & billing
│   ├── gateway/          # AI gateway config
│   └── settings/         # User settings (7 components)
├── components/           # Shared components only
│   ├── ui/              # shadcn/ui primitives
│   ├── layout/          # App shell (header, sidebar, footer)
│   ├── guards/          # Auth guards
│   └── shared/          # Generic reusable components
├── lib/                 # Core infrastructure
│   ├── api/core/        # BrokleAPIClient (HTTP client)
│   ├── auth/            # JWT utilities
│   └── utils/           # Pure utilities
├── hooks/               # Global hooks (use-mobile, etc.)
├── stores/              # Global stores (ui-store.ts)
├── context/             # Cross-feature context (workspace-context)
├── types/               # Shared types
└── __tests__/           # Test infrastructure (MSW, utilities)
```

**Feature Structure**: Each feature has `components/`, `hooks/`, `api/`, `stores/` (optional), `types/`, `__tests__/`, and `index.ts` (public exports)

**Import Pattern**: Always use `@/features/[feature]` (never import internal paths)

### Key Technologies
- Next.js 15.5.2 (App Router, Turbopack), React 19.2.0, TypeScript 5.9.3 (strict mode)
- Tailwind CSS 4.1.15, shadcn/ui components
- State: Zustand (client) + React Query (server state)
- Forms: React Hook Form + Zod validation
- Testing: Vitest + React Testing Library + MSW (30% coverage target)
- Package manager: pnpm

### Frontend Commands
```bash
cd web && pnpm install     # Install dependencies
pnpm dev                   # Start dev server (Turbopack)
pnpm build                 # Build for production
pnpm lint                  # Lint code
pnpm test                  # Run tests
pnpm test:coverage         # Run tests with coverage
make dev-frontend          # Or use Makefile
```

**Documentation**: See `web/ARCHITECTURE.md` for detailed architecture guide

## Health & Monitoring

- **Metrics**: `/metrics` (request/response, business, infrastructure)
- **Health Checks**: `/health`, `/health/db`, `/health/cache`, `/health/providers`
- **Tracing**: Distributed tracing with correlation IDs

## Key Architectural Patterns

### Dependency Injection Container
The app uses a centralized DI container in `internal/app/app.go`:
- Initializes databases → repositories → services → handlers
- Graceful shutdown handling
- Health check aggregation

### Industrial Error Handling Pattern
See error handling documentation in `docs/development/` for complete patterns.

### Enterprise Feature Toggle
Enterprise features use interface-based design with build tags:
```go
// internal/ee/sso/build.go (OSS)
func New() SSOProvider {
    return &stubSSOProvider{}
}

// internal/ee/sso/build_enterprise.go (Enterprise)
func New() SSOProvider {
    return &enterpriseSSOProvider{}
}
```

### Background Workers
- `analytics_worker.go` - Metrics aggregation
- `notification_worker.go` - Email/SMS notifications
- `telemetry_stream_consumer.go` - Redis streams telemetry processing

### Multi-Database Strategy
- **PostgreSQL**: Transactional data (GORM), user data, configurations
- **ClickHouse**: Time-series analytics (raw SQL), request logs, metrics
- **Redis**: Cache, pub/sub, job queues, sessions, telemetry streams

### Authentication Features
- **API Keys**: `bk_{40_char}` format with SHA-256 hashing
- **JWT**: Dashboard authentication with session management
- **OAuth**: GitHub and Google providers
- **RBAC**: Scope-based role permissions

## Troubleshooting

- **Port conflicts**: `lsof -ti:8080` and kill
- **Migration dirty state**: `go run cmd/migrate/main.go -db <db> drop` then re-migrate
- **Enterprise build**: Ensure `-tags="enterprise"` flag
- **API key test**: `curl -X POST http://localhost:8080/v1/auth/validate-key -H "Content-Type: application/json" -d '{"api_key": "bk_..."}'`