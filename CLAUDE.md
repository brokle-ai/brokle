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
└── docs/                  # Documentation
```

### Domain-Driven Architecture
The codebase follows DDD patterns with:
- **Domain Layer** (`internal/core/domain/`): Entities, repositories, services
- **Infrastructure Layer** (`internal/infrastructure/`): Database, external APIs
- **Transport Layer** (`internal/transport/`): HTTP handlers, WebSocket
- **Application Layer** (`internal/app/`): Dependency injection, bootstrap

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
- **ClickHouse**: 5 separate table migrations for better granularity:
  - `metrics` - Real-time platform metrics (90 day TTL)
  - `events` - Business/system events (180 day TTL)  
  - `traces` - Distributed tracing (30 day TTL)
  - `request_logs` - API request logging (60 day TTL)
  - `ai_routing_metrics` - AI provider routing decisions (365 day TTL)

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
```go
// 1. Extract API key from X-API-Key or Authorization header
apiKey := c.GetHeader("X-API-Key")
// Fallback: Authorization: Bearer {api_key}

// 2. Validate API key format (bk_{40_chars})
if err := auth.ValidateAPIKeyFormat(apiKey); err != nil {
    return errors.New("invalid API key format")
}

// 3. Validate against database with SHA-256 hash lookup (O(1))
validateResp, err := apiKeyService.ValidateAPIKey(ctx, apiKey)
// This does: SHA-256(apiKey) -> database lookup -> return project_id + auth context

// 4. Store authentication context (project_id comes from database)
c.Set("project_id", &validateResp.ProjectID)
c.Set("api_key_id", validateResp.AuthContext.APIKeyID)
```

### Middleware Architecture

#### SDKAuthMiddleware
Located in `internal/transport/http/middleware/sdk_auth.go`:

```go
type SDKAuthMiddleware struct {
    apiKeyService auth.APIKeyService
    logger        *logrus.Logger
}

// RequireSDKAuth validates API keys for SDK routes
func (m *SDKAuthMiddleware) RequireSDKAuth() gin.HandlerFunc {
    // Extracts and validates API key
    // Stores authentication context in Gin context
    // Handles both X-API-Key and Authorization headers
}
```

**Context Keys**:
- `SDKAuthContextKey` - Full authentication context
- `APIKeyIDKey` - API key identifier
- `ProjectIDKey` - Project ID (stored as pointer)
- `EnvironmentKey` - Environment tag from header

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
The HTTP server implements clean route separation in `internal/transport/http/server.go`:

```go
// SDK routes (/v1) - API Key authentication
sdk := engine.Group("/v1")
sdk.Use(sdkAuthMiddleware.RequireSDKAuth())
sdk.Use(rateLimitMiddleware.RateLimitByAPIKey())
setupSDKRoutes(sdk)

// Dashboard routes (/api/v1) - JWT authentication
dashboard := engine.Group("/api/v1")
dashboard.Use(rateLimitMiddleware.RateLimitByIP()) // Global IP limiting
protected := dashboard.Group("")
protected.Use(authMiddleware.RequireAuth())       // JWT auth
protected.Use(rateLimitMiddleware.RateLimitByUser()) // User limiting
setupDashboardRoutes(protected)
```

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
Time-series data optimized for analytical queries:
- Request/response logs with full tracing
- AI routing decision metrics
- Performance and latency metrics  
- Cost tracking and optimization data
- Quality scores and ML model performance

### Cache Layer (Redis)  
- JWT token and session storage
- Rate limiting counters
- Background job queues (analytics, notifications)
- Semantic cache for AI responses
- Real-time event pub/sub for WebSocket

## Development Patterns

### API Key Management Utilities
The platform includes industry-standard API key utilities in `internal/core/domain/auth/apikey_utils.go`:

```go
// Generate new industry-standard API key (pure random)
fullKey, err := auth.GenerateAPIKey()
// Returns: "bk_AbCdEfGhIjKlMnOpQrStUvWxYz0123456789AbCd" (43 chars total)

// Validate API key format (bk_{40_chars})
err := auth.ValidateAPIKeyFormat(fullKey)
// Returns error if format is invalid

// Create preview for display (security best practice - GitHub pattern)
preview := auth.CreateKeyPreview(fullKey)
// Returns: "bk_AbCd...AbCd" (shows bk_ + first 4 + ... + last 4 chars)

// Example: Creating and storing an API key
fullKey, err := auth.GenerateAPIKey()
keyHash := sha256.Sum256([]byte(fullKey))
keyHashHex := hex.EncodeToString(keyHash[:])
keyPreview := auth.CreateKeyPreview(fullKey)

// Store in database: keyHashHex, keyPreview, project_id
// Return to user: fullKey (ONLY ONCE), keyPreview

// Example: Validating an API key
hash := sha256.Sum256([]byte(incomingKey))
hashHex := hex.EncodeToString(hash[:])
apiKey, err := repo.GetByKeyHash(ctx, hashHex) // O(1) with unique index
```

### Authentication Middleware Patterns

#### SDK Authentication (API Keys)
```go
// For SDK routes requiring API key authentication
func (h *Handler) SDKEndpoint(c *gin.Context) {
    // Get authentication context from middleware
    authCtx, exists := middleware.GetSDKAuthContext(c)
    if !exists {
        response.Unauthorized(c, "Authentication required")
        return
    }

    // Get project ID (stored as pointer)
    projectID, exists := middleware.GetProjectID(c)
    if !exists {
        response.InternalServerError(c, "Project context missing")
        return
    }

    // Optional environment tag
    environment, _ := middleware.GetEnvironment(c)

    // Use authentication context
    log.Printf("Project: %s, Environment: %s", projectID.String(), environment)
}
```

#### Dashboard Authentication (JWT)
```go
// For dashboard routes requiring JWT authentication
func (h *Handler) DashboardEndpoint(c *gin.Context) {
    // Get user context from JWT middleware
    userID, exists := middleware.GetUserID(c)
    if !exists {
        response.Unauthorized(c, "Authentication required")
        return
    }

    // Get organization context (if required)
    orgID, exists := middleware.GetOrganizationID(c)
    // Handle organization-scoped operations
}
```

### Clean Architecture
Follow repository → service → handler pattern:

```go
// Repository layer
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id string) (*models.User, error)
}

// Service layer  
type UserService struct {
    repo UserRepository
}

func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) error {
    // Business logic here
    user := &models.User{...}
    return s.repo.Create(ctx, user)
}

// Handler layer
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // HTTP handling
    err := h.service.CreateUser(r.Context(), req)
    // Response handling
}
```

### Error Handling

**📖 CRITICAL**: Use the comprehensive **[Development Error Handling Guides](docs/development/)** for all error handling implementations:

- **[ERROR_HANDLING_GUIDE.md](docs/development/ERROR_HANDLING_GUIDE.md)** - Complete industrial patterns across all layers
- **[DOMAIN_ALIAS_PATTERNS.md](docs/development/DOMAIN_ALIAS_PATTERNS.md)** - Professional import patterns  
- **[ERROR_HANDLING_QUICK_REFERENCE.md](docs/development/ERROR_HANDLING_QUICK_REFERENCE.md)** - Developer cheat sheet

The platform follows **industrial Go best practices** with structured error handling:

**Clean Architecture Error Flow:**
```
Repository (Domain Errors) → Service (AppErrors) → Handler (HTTP Response)
```

**Core Principles:**
- **Repository Layer**: Domain errors with proper wrapping
- **Service Layer**: AppError constructors (NewUnauthorizedError, NewNotFoundError, etc.)
- **Handler Layer**: Centralized `response.Error(c, err)` handling
- **Decorator Pattern**: Cross-cutting concerns (audit, logging) handled separately
- **Zero Logging**: Core services focus on pure business logic

**Example Implementation:**
```go
// Repository layer - Domain errors
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, fmt.Errorf("get user by email %s: %w", email, user.ErrNotFound)
}

// Service layer - AppError constructors  
if errors.Is(err, user.ErrNotFound) {
    return nil, appErrors.NewUnauthorizedError("Invalid email or password")
}

// Handler layer - Centralized error handling
resp, err := h.authService.Login(ctx, req)
if err != nil {
    response.Error(c, err) // Automatic HTTP status mapping
    return
}
response.Success(c, resp)
```

**Key Requirements:**
- ❌ **No fmt.Errorf/errors.New** in services - use AppError constructors
- ❌ **No logging** in core services - use decorator pattern  
- ✅ **Domain error mapping** at repository layer
- ✅ **Structured AppErrors** at service layer
- ✅ **Clean separation** of business logic and cross-cutting concerns

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
The application uses Viper for configuration management:

```go
// Configuration is loaded from:
// 1. Environment variables
// 2. .env file
// 3. Default values

type Config struct {
    Server struct {
        Port int `mapstructure:"port"`
        Host string `mapstructure:"host"`
    }
    Database struct {
        PostgresURL   string `mapstructure:"postgres_url"`
        ClickHouseURL string `mapstructure:"clickhouse_url"`
        RedisURL      string `mapstructure:"redis_url"`
    }
    // Enterprise features controlled by build tags
    Enterprise struct {
        Enabled bool `mapstructure:"enterprise_enabled"`
    }
}
```

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

### Working with SDK Observability
The platform uses a unified high-performance telemetry batch system for all SDK observability operations.

#### Unified Telemetry Batch Endpoint
The SDK sends all telemetry data (traces, observations, quality scores, events) through a single batch endpoint:

```go
// SDK telemetry ingestion (require API key authentication)
POST /v1/ingest/batch           // Unified batch processing for all event types
GET  /v1/telemetry/health          // Service health monitoring
GET  /v1/telemetry/metrics         // Performance metrics
GET  /v1/telemetry/performance     // Performance statistics
GET  /v1/ingest/batch/:batch_id // Batch status tracking
POST /v1/telemetry/validate        // Event validation
```

**Example Telemetry Batch Request**:
```json
{
  "environment": "production",
  "events": [
    {
      "event_id": "01ABCDEFGHIJKLMNOPQRSTUVWXYZ",
      "event_type": "trace",
      "payload": {"name": "my-trace", "user_id": "user_123"}
    },
    {
      "event_id": "01BCDEFGHIJKLMNOPQRSTUVWXYZ0",
      "event_type": "observation",
      "payload": {"trace_id": "...", "type": "llm"}
    },
    {
      "event_id": "01CDEFGHIJKLMNOPQRSTUVWXYZ01",
      "event_type": "quality_score",
      "payload": {"trace_id": "...", "score": 0.95}
    },
    {
      "event_id": "01DEFGHIJKLMNOPQRSTUVWXYZ012",
      "event_type": "event",
      "payload": {"event_name": "user_action", "metadata": {...}}
    }
  ],
  "deduplication": {
    "enabled": true,
    "ttl": 3600,
    "use_redis_cache": true
  }
}
```

**Key Features**:
- ✅ ULID-based deduplication (prevents duplicate events)
- ✅ High-throughput batch processing (1000+ events per request)
- ✅ Mixed event types in single request
- ✅ Async processing support
- ✅ Redis-backed caching for performance

#### Dashboard Analytics (Read-only)
```go
// Dashboard routes for viewing telemetry (require JWT authentication)
GET /api/v1/analytics/traces                      // List traces
GET /api/v1/analytics/traces/:id                  // Get specific trace
GET /api/v1/analytics/traces/:id/observations     // Get trace with observations
GET /api/v1/analytics/traces/:id/stats            // Trace statistics
GET /api/v1/analytics/observations                // List observations
GET /api/v1/analytics/observations/:id            // Get observation details
GET /api/v1/analytics/quality-scores              // List quality scores
GET /api/v1/analytics/quality-scores/:id          // Get quality score
GET /api/v1/analytics/traces/:id/quality-scores   // Quality scores by trace
GET /api/v1/analytics/observations/:id/quality-scores // Quality scores by observation
```

### Environment Tag System
The platform now uses environment tags instead of environment entities:

```go
// Environment passed in request body (optional)
// Example in telemetry batch request:
// {
//   "environment": "production",  // Optional environment tag
//   "events": [...]
// }

// Used in observability and routing decisions
telemetryData := TelemetryData{
    ProjectID:   projectID,
    Environment: environment, // Optional tag
    // ... other fields
}
```

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
**📖 See [INDUSTRIAL_ERROR_HANDLING_GUIDE.md](INDUSTRIAL_ERROR_HANDLING_GUIDE.md) for complete implementation guide.**

The platform implements clean architecture error handling:
- **Core Services**: Pure business logic with AppError constructors
- **Decorator Pattern**: Cross-cutting concerns (audit logging) handled separately
- **Zero Logging**: Business logic services have no logging dependencies
- **Structured Flow**: Repository → Service → Handler with proper error transformation

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

### Service Health Checks
```bash
# Check all services
make health

# Check specific services
curl http://localhost:8080/health
curl http://localhost:8080/health/db
curl http://localhost:8080/health/cache
```

### Docker Development
```bash
# Start with Docker Compose
make docker-dev

# Production Docker build
make docker-prod

# Clean Docker resources
make docker-clean
```

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

# Test SDK endpoint with API key
curl -X GET http://localhost:8080/v1/models \
  -H "X-API-Key: bk_AbCdEfGhIjKlMnOpQrStUvWxYz0123456789AbCd"

# Test SDK endpoint (environment in body for telemetry endpoints)
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "X-API-Key: your_api_key" \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-3.5-turbo", "messages": [{"role": "user", "content": "Hello"}]}'
```

### API Key Debugging
```go
// Validate API key format
if err := auth.ValidateAPIKeyFormat(apiKey); err != nil {
    log.Printf("Invalid format: %v", err)
    return
}

// Generate key preview for logging (security best practice - GitHub pattern)
preview := auth.CreateKeyPreview(apiKey)
log.Printf("Key preview: %s", preview) // "bk_rvOJ...yym0"

// Hash API key for database lookup
hash := sha256.Sum256([]byte(apiKey))
hashHex := hex.EncodeToString(hash[:])
log.Printf("Key hash: %s", hashHex[:16]+"...") // Partial hash for debugging

// Validate and retrieve API key from database
apiKey, err := repo.GetByKeyHash(ctx, hashHex)
if err != nil {
    log.Printf("Key not found or invalid: %v", err)
    return
}
log.Printf("Project ID: %s", apiKey.ProjectID.String()) // Retrieved from database
```