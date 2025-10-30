# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Platform Overview

**Brokle** - The Open-Source AI Control Plane. See Everything. Control Everything. Observability, routing, and governance for AI — open source and built for scale.

### Core Mission
See Everything. Control Everything. Build the unified open-source control plane for AI teams to monitor, route, and govern production AI applications with complete visibility and control.

### Key Features
- **See Everything**: 40+ AI-specific metrics with real-time observability and monitoring
- **Control Everything**: Intelligent routing across 250+ LLM providers with governance controls
- **Open-Source Control**: Complete visibility into your AI infrastructure with no vendor lock-in
- **Production Scale**: Multi-tenant architecture with enterprise-grade governance
- **Cost Intelligence**: 30-50% reduction in LLM costs through smart routing and optimization

## Architecture: Scalable Monolith

This project represents a **scalable monolith** architecture, migrated from microservices for better development velocity and open source adoption.

### High-Level Architecture
- **Single Go Binary** with HTTP server at `:8080`
- **Multi-Database Strategy**: PostgreSQL (transactional) + ClickHouse (analytics) + Redis (cache/queues)
- **Domain-Driven Design** with clean separation of concerns
- **Enterprise Edition** toggle via build tags (`go build -tags="enterprise"`)
- **Transport Layer Pattern** for HTTP/WebSocket handling

### Application Structure
```
brokle/
├── cmd/                    # Application entry points
│   ├── server/main.go     # Main HTTP server
│   ├── migrate/main.go    # Database migration runner
│   └── seed/main.go       # Database seeding tool
├── internal/              # Private application code
│   ├── app/               # Application bootstrap & DI container
│   ├── config/            # Configuration management
│   ├── core/domain/       # Domain-driven design structure
│   │   ├── user/          # User domain (entity, repo, service)
│   │   └── auth/          # Authentication domain
│   ├── infrastructure/    # External integrations
│   │   ├── database/      # DB connections (postgres, redis, clickhouse)
│   │   └── repository/    # Repository implementations
│   ├── transport/http/    # HTTP transport layer
│   │   ├── handlers/      # HTTP handlers by domain
│   │   ├── middleware/    # HTTP middleware
│   │   └── server.go      # HTTP server setup
│   ├── services/          # Service implementations
│   ├── workers/           # Background workers
│   └── ee/                # Enterprise Edition features
├── pkg/                   # Public shared packages
│   ├── errors/            # Error handling
│   ├── response/          # HTTP response utilities
│   ├── utils/             # Common utilities
│   └── websocket/         # WebSocket utilities
├── migrations/            # Database migration files
├── web/                   # Next.js frontend
└── docs/                  # Public OSS documentation only
```

### Documentation Organization

**IMPORTANT**: This project maintains a strict separation between public and internal documentation:

#### Public Documentation (`docs/`)
**Location**: `brokle/docs/` (committed to OSS repository)
**Purpose**: User-facing documentation for OSS contributors and users
**Contents**:
- API documentation and reference
- Architecture overview for contributors
- Development guides and coding standards
- Enterprise feature documentation
- Testing guides and best practices
- Deployment instructions

#### Internal Documentation (`internal-docs/`)
**Location**: Separate private repository
**Purpose**: Internal research, planning, and exploration
**Contents**:
- `research/` - Competitor analysis, technology evaluations, market research
- `migrations/` - Migration documentation (e.g., OpenTelemetry migration)
- `planning/` - Feature planning, roadmaps, priorities
- `decisions/` - Architecture Decision Records (ADRs), RFCs
- `data/` - CSV files, analysis scripts, data exports

**Documentation Workflow**:
1. **Research/Exploration** → Create in `internal-docs/`
2. **Implementation** → Implementation guides in `internal-docs/` during development
3. **Public Release** → Only user-facing docs move to `brokle/docs/`
4. **Never commit** internal docs to main OSS repository

**What Goes Where**:
```
✅ OSS Repo (brokle/docs/)          ❌ Internal Docs Only (internal-docs/)
- API reference                     - Competitor analysis
- Architecture guides               - Migration research
- Development workflows             - Technical explorations
- Testing standards                 - Feature planning
- Enterprise features               - Decision records
- Deployment guides                 - Data analysis
                                    - Internal roadmaps
```

### Domain-Driven Architecture
The codebase follows DDD patterns with clear separation of concerns:

#### Layer Organization
- **Domain Layer** (`internal/core/domain/`): Core business concepts
  - Entities: Domain models and business rules
  - Repository Interfaces: Data access contracts
  - Service Interfaces: Business operation contracts

- **Service Implementation Layer** (`internal/services/`):
  - Concrete implementations of domain service interfaces
  - Business logic orchestration
  - Example: `internal/services/observability/` implements interfaces from `internal/core/domain/observability/`

- **Infrastructure Layer** (`internal/infrastructure/`): External integrations
  - Database connections and configurations
  - Repository implementations (`internal/infrastructure/repository/`)
  - External API clients

- **Transport Layer** (`internal/transport/`): Request/Response handling
  - HTTP handlers and routing
  - WebSocket connections
  - Middleware components

- **Application Layer** (`internal/app/`): Application bootstrap
  - Dependency injection container
  - Service registry and wiring
  - Graceful shutdown coordination

## Development Commands

### Essential Commands
```bash
# First time setup (installs deps, starts DBs, runs migrations, seeds data)
make setup

# Start full development stack (Go API + Next.js frontend)
make dev

# Start backend only
make dev-backend

# Start frontend only  
make dev-frontend

# Run all tests
make test

# Build for development (faster builds)
make build-dev

# Build production (OSS and Enterprise)
make build-oss        # Default OSS build
make build-enterprise  # Enterprise build with all features
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
- **✅ Confirmation Prompts**: All destructive operations require explicit 'yes' confirmation
- **✅ Dry Run Mode**: Preview changes with `-dry-run` flag
- **✅ Granular Control**: Target specific databases with `-db postgres|clickhouse|all`
- **✅ Health Monitoring**: Comprehensive status reporting with dirty state detection
- **✅ Rollback Support**: Safe rollback with proper down migration support

#### Migration Architecture
- **PostgreSQL**: Single comprehensive schema migration (user management, auth, organizations)
- **ClickHouse**: Enhanced observability schema with recent updates (Oct 2024):

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

### Core Services
- `PORT` - HTTP server port (default: 8080)
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

**Unified Telemetry Batch System (SDK Observability)**:
- `POST /v1/ingest/batch` - High-performance batch processing for all telemetry events (traces, observations, quality scores)
- `GET /v1/telemetry/health` - Telemetry service health monitoring
- `GET /v1/telemetry/metrics` - Telemetry performance metrics
- `GET /v1/telemetry/performance` - Performance statistics
- `GET /v1/ingest/batch/:batch_id` - Batch status tracking
- `POST /v1/telemetry/validate` - Event validation

**OpenTelemetry Protocol (OTLP) Native Support** (Added Oct 2024):
- `POST /v1/otlp/traces` - OTLP traces ingestion endpoint
- `POST /v1/traces` - Alternative OTLP traces endpoint
- **Format Support**: Protobuf (binary) and JSON payloads
- **Compression**: Automatic gzip decompression via Content-Encoding header
- **Smart Processing**: OTLPConverterService with intelligent root span detection
- **Multi-Exporter Compatible**: Handles spans from various OTLP exporters

**Three Parallel Ingestion Systems**:
1. **Brokle Native Batch** (`/v1/ingest/batch`) - High-performance batch processing optimized for Brokle SDKs
   - ULID-based deduplication, mixed event types, Redis caching
   - Supports traces, observations, quality scores, and events in single batch
2. **OpenTelemetry Protocol** (OTLP endpoints above) - Industry-standard compatibility
3. **Redis Streams Backend** (`TelemetryStreamConsumer` worker) - Async high-throughput processing

All systems converge at `internal/services/observability/` for unified processing.

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
- `/api/v1/onboarding/*` - User onboarding flow
- `/api/v1/organizations/*` - Organization management with RBAC
- `/api/v1/projects/*` - Project management and API key management
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

#### Middleware Components

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

## 📚 COMPREHENSIVE ENTERPRISE DOCUMENTATION

**CRITICAL**: Complete enterprise documentation is available in `/docs/ENTERPRISE.md` and `/docs/enterprise/` directory.

### Enterprise Documentation Suite
- **[ENTERPRISE.md](docs/ENTERPRISE.md)** - Main enterprise overview (260+ pages total)
  - Open-core business model & strategy
  - Complete license tier breakdown (Free → Pro → Business → Enterprise)  
  - All enterprise features with detailed explanations
  - Architecture overview & build system
  - Getting started guides & configuration examples

### Specialized Enterprise Guides
- **[DEVELOPER_GUIDE.md](docs/enterprise/DEVELOPER_GUIDE.md)** - Technical implementation
  - Build tag architecture & compilation details
  - Configuration management systems
  - Middleware & feature gating implementation
  - License service architecture
  - Testing strategies & debugging guides

- **[SSO.md](docs/enterprise/SSO.md)** - Single Sign-On integration
  - SAML 2.0 & OIDC/OAuth2 setup guides
  - Azure AD, Okta, Google Workspace configurations
  - User provisioning & role mapping
  - Troubleshooting & debugging tools

- **[RBAC.md](docs/enterprise/RBAC.md)** - Role-Based Access Control
  - Built-in & custom role systems
  - Granular permissions & scope hierarchy
  - SSO integration & automatic role assignment
  - Best practices & organizational patterns

- **[COMPLIANCE.md](docs/enterprise/COMPLIANCE.md)** - Data governance & compliance
  - SOC 2, HIPAA, GDPR compliance frameworks
  - Data governance & retention policies
  - Audit trails & privacy controls
  - Certification support & best practices

- **[ANALYTICS.md](docs/enterprise/ANALYTICS.md)** - Advanced analytics & intelligence
  - ML-powered predictive insights & cost forecasting
  - Custom dashboard builder with 50+ visualization types
  - Business intelligence & executive reporting
  - Real-time analytics & optimization recommendations

### Enterprise Architecture Highlights
- **Open-Core Model**: Clean OSS/Enterprise separation with professional upgrade paths
- **70% Cost Advantage**: Brokle Pro ($29) vs Portkey ($99+) with superior features
- **Production Ready**: SOC 2/HIPAA/GDPR compliance, enterprise SSO, advanced RBAC
- **Complete Platform**: Gateway + Observability + Caching + Optimization + Future Model Hosting
- **Business Intelligence**: ML-powered insights, cost optimization, predictive analytics

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

**Target Metrics:**
- Service Layer: ~1:1 test-to-code ratio (focus on business logic)
- Domain Layer: Minimal (only complex calculations and business rules)
- Handler Layer: Critical workflows only (integration tests)

**Current Coverage:**
- Observability Services: 3,485 lines of tests (0.96:1 ratio) ✅
- Observability Domain: 594 lines of tests (business logic only) ✅
- All tests passing with healthy coverage

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
go test ./internal/services/observability -v

# Run with race detection
go test -race ./...
```

### Quick Reference

- **Detailed Guide**: See `docs/TESTING.md` for complete examples and patterns
- **AI Prompt**: See `prompts/testing.txt` for AI-assisted test generation
- **Reference Code**: See `internal/services/observability/*_test.go` for real examples

### Test Quality Checklist

Before committing tests:
- ✅ Table-driven test pattern with comprehensive scenarios
- ✅ Mocks implement full repository interfaces
- ✅ Tests business logic, not framework behavior
- ✅ Verifies mock expectations with `AssertExpectations()`
- ✅ Maintains ~1:1 test-to-code ratio
- ❌ No tests for simple CRUD, validation, or constructors

## Frontend Architecture

The frontend uses **Next.js 15** with App Router and runs on port `:3000`:

```
web/src/
├── app/                   # Next.js App Router
│   ├── (auth)/           # Auth route group (/auth/*)
│   ├── (dashboard)/      # Dashboard routes
│   │   ├── [orgSlug]/    # Organization-scoped routes
│   │   ├── settings/     # User settings
│   │   └── onboarding/   # User onboarding
│   └── layout.tsx        # Root layout
├── components/           # React components
│   ├── ui/              # shadcn/ui components
│   ├── analytics/       # Analytics dashboards
│   ├── auth/            # Authentication components
│   └── layout/          # Layout components
├── hooks/               # Custom React hooks
├── lib/                 # API clients and utilities
├── store/               # Zustand state management
└── types/               # TypeScript definitions
```

### Key Frontend Technologies
- **Next.js 15** with App Router and Turbopack
- **React 19** with React Server Components
- **TypeScript** with strict type checking
- **Tailwind CSS v4** for styling
- **shadcn/ui** component library
- **Zustand** for state management
- **TanStack Query** for API state
- **React Hook Form** with Zod validation
- **pnpm** for package management

### Frontend Development
```bash
# Navigate to frontend directory
cd web

# Install dependencies
pnpm install

# Start development server with Turbopack
pnpm run dev

# Build for production
pnpm run build

# Run linting
pnpm run lint

# Format code
pnpm run format
```

**Note**: The Makefile uses `pnpm` for all frontend operations. You can use either Make commands or direct `pnpm` commands for frontend development.

## Performance & Monitoring

### Metrics
The application exposes metrics at `/metrics`:
- Request/response metrics
- Business metrics (users, requests, costs)
- Infrastructure metrics (database, cache, queues)

### Tracing
Distributed tracing with correlation IDs across all operations.

### Health Checks
- `/health` - Application health
- `/health/db` - Database connectivity  
- `/health/cache` - Redis connectivity
- `/health/providers` - AI provider status

## Deployment

### Docker
```bash
# Development
docker compose up -d

# Production  
docker compose -f docker-compose.prod.yml up -d
```

### Production Checklist
- [ ] Environment variables configured
- [ ] Database migrations run
- [ ] SSL certificates configured
- [ ] Monitoring alerts configured
- [ ] Backup strategy implemented

## Contributing Guidelines

### Code Style
- Use `gofmt` and `golint`
- Follow Go naming conventions
- Write meaningful commit messages
- Add tests for new functionality

### Pull Request Process
1. Create feature branch from `main`
2. Implement feature with tests
3. Run full test suite: `make test`
4. Run linting: `make lint`
5. Create pull request with description

### Commit Message Format
```
feat(domain): add user authentication
fix(gateway): resolve provider routing issue
docs: update API documentation
```

### Documentation Guidelines

**CRITICAL**: Always create documentation in the correct location:

#### Creating New Documentation

1. **Determine the audience**:
   - OSS contributors/users → `brokle/docs/`
   - Internal team only → `internal-docs/` (separate repo)

2. **Choose the category** (for internal docs):
   - Research/exploration → `internal-docs/research/`
   - Migration documentation → `internal-docs/migrations/`
   - Feature planning → `internal-docs/planning/`
   - Technical decisions → `internal-docs/decisions/`
   - Data analysis → `internal-docs/data/`

3. **File naming conventions**:
   - Use descriptive, kebab-case names: `feature-implementation-guide.md`
   - Avoid generic names: ❌ `notes.md` → ✅ `otel-migration-notes.md`
   - Include context in filename when useful

4. **Never commit to main repo root**:
   - ❌ Root-level .md files (except README.md, CLAUDE.md, CONTRIBUTING.md, SECURITY.md)
   - ❌ Exploration/research docs in OSS repo
   - ❌ Internal planning docs in OSS repo
   - ✅ All internal docs go to `internal-docs/` repo

#### Internal Documentation Best Practices

**For internal-docs repository**:
- Use clear directory structure (research/, migrations/, planning/, decisions/, data/)
- Add README files to explain directory contents
- Link related documents together
- Archive outdated documents (don't delete immediately)
- Use date prefixes for time-sensitive docs: `2024-10-27-otel-migration.md`

**For OSS documentation**:
- Focus on user value and contributor guidance
- Keep examples practical and tested
- Update when features change
- Link to relevant code sections
- Use clear headings and structure

## Common Tasks

### Adding New Domain
1. Create domain directory in `internal/`
2. Implement repository interface
3. Add service layer with business logic
4. Create HTTP handlers
5. Add routes to main router
6. Write tests
7. Update documentation

### Adding New AI Provider
1. Implement provider interface in `internal/gateway/providers/`
2. Add provider configuration
3. Update routing logic
4. Add monitoring metrics
5. Write integration tests

### Database Changes
1. Create migration file in `migrations/`
2. Update models in relevant domain
3. Update repository methods
4. Run migration: `make migrate`
5. Update tests

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
Asynchronous processing in `internal/workers/`:
- `analytics_worker.go` - Metrics aggregation
- `notification_worker.go` - Email/SMS notifications

### Multi-Database Strategy
- **PostgreSQL**: Transactional data with GORM ORM, user data, configurations
- **ClickHouse**: Time-series analytics with raw SQL, request logs, metrics aggregation
- **Redis**: Caching layer, pub/sub messaging, background job queues, session storage

### Running Single Tests
```bash
# Run specific Go test
go test ./internal/core/domain/user/...
go test -run TestUserService_CreateUser ./internal/services/

# Run specific Go test with verbose output
go test -v ./internal/transport/http/handlers/auth/

# Run frontend tests (if implemented)
cd web && pnpm test
```

## Troubleshooting

### Common Issues
- **Port conflicts**: Check with `lsof -ti:8080` and kill processes
- **Database migrations**: Run `go run cmd/migrate/main.go status` for detailed health check
- **Migration dirty state**: Use `go run cmd/migrate/main.go -db <database> drop` then re-run migrations
- **Enterprise build errors**: Ensure proper build tags usage
- **WebSocket connection issues**: Check CORS and proxy settings

### SDK Authentication Issues
```bash
# Test API key validation (public endpoint)
curl -X POST http://localhost:8080/v1/auth/validate-key \
  -H "Content-Type: application/json" \
  -d '{"api_key": "bk_AbCdEfGhIjKlMnOpQrStUvWxYz0123456789AbCd"}'