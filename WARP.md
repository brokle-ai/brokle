# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

**Brokle** is The Open-Source AI Control Plane - See Everything. Control Everything. Provides observability, routing, and governance services for AI in production. Built as a **scalable monolith** with an **open-core business model**, it offers both OSS and Enterprise editions through Go build tags.

### Core Mission
See Everything. Control Everything. The unified open-source control plane for AI teams to monitor, route, and govern production AI applications with complete visibility and control.

### Key Features
- **AI Gateway & Routing**: OpenAI-compatible proxy with smart multi-provider routing across 250+ LLM providers
- **Advanced Observability**: 40+ AI-specific metrics with real-time monitoring and distributed tracing
- **Semantic Caching**: Vector-based response caching for up to 95% cost reduction potential
- **Multi-tenant Architecture**: Organization → Project → Environment isolation with RBAC
- **Enterprise Edition**: SSO (SAML/OIDC), compliance controls, predictive analytics, custom dashboards

## Essential Development Commands

### Initial Setup
```bash
# Complete development environment setup
# Installs deps, starts DBs, runs migrations, seeds data
make setup

# Quick development start (Go API + Next.js dashboard)
make dev

# Start services individually
make dev-backend    # Go API on :8080 with hot reload (air)
make dev-frontend   # Next.js on :3000 with Turbopack
```

### Database Operations
```bash
# Core migration commands
make migrate-up       # Run all pending migrations
make migrate-down     # Rollback one migration
make migrate-status   # Detailed health check across all databases
make migrate-reset    # DESTRUCTIVE - Reset all databases with confirmation

# Advanced migration CLI (granular control)
go run cmd/migrate/main.go -db postgres up         # PostgreSQL only
go run cmd/migrate/main.go -db clickhouse up       # ClickHouse only
go run cmd/migrate/main.go status                  # Health check with dirty state detection
go run cmd/migrate/main.go -steps 2 up             # Run 2 migrations forward
go run cmd/migrate/main.go -dry-run up              # Preview without executing

# Create new migrations
make create-migration DB=postgres NAME=add_users_table
make create-migration DB=clickhouse NAME=add_metrics_table

# Data seeding
make seed-dev         # Development data
make seed-prod        # Production data

# Database shell access
make shell-db         # PostgreSQL psql
make shell-redis      # Redis CLI
make shell-clickhouse # ClickHouse client
```

### Building & Testing
```bash
# OSS build (default)
make build-oss
make build-dev-oss    # Development build (faster, debug info)

# Enterprise build with all features
make build-enterprise
make build-dev-enterprise

# Testing
make test             # All tests
make test-coverage    # Tests with HTML coverage report
make test-unit        # Unit tests only (excludes integration)
make test-integration # Integration tests with real databases
make test-e2e         # End-to-end tests
make test-load        # Load testing

# Run specific Go tests
go test ./internal/core/domain/user/...
go test -run TestUserService_CreateUser ./internal/services/
go test -v ./internal/transport/http/handlers/auth/
```

### Code Quality
```bash
# Linting and formatting
make lint             # All linters (Go + frontend)
make lint-go          # golangci-lint with custom config
make lint-frontend    # Next.js ESLint + Prettier

make fmt              # Format Go code (gofmt + goimports)
make fmt-frontend     # Format frontend (Prettier)

make security-scan    # gosec security scanning
```

### Development Tools
```bash
# Install required development tools
make install-tools    # air, swag, golangci-lint, goimports, gosec

# Hot reload with air
make watch            # Watch Go files and restart server
make hot-reload       # Alias for watch

# API documentation
make docs-generate    # Generate Swagger docs
make docs-serve       # Serve docs locally on :8000
```

## Architecture Overview

### Scalable Monolith Design
Migrated from microservices to monolith for better development velocity and open source adoption:

- **Single Go Binary**: HTTP server on `:8080` with WebSocket support
- **Multi-Database Strategy**: PostgreSQL (transactional) + ClickHouse (analytics) + Redis (cache/queues)
- **Domain-Driven Design**: Clean separation with repository → service → handler pattern
- **Enterprise Toggle**: Build tags control feature availability
- **Next.js 15 Frontend**: App Router with Server Components on `:3000`

### Directory Structure
```
brokle/
├── cmd/                    # Application entry points
│   ├── server/main.go     # Main HTTP server (primary entry)
│   ├── migrate/main.go    # Comprehensive migration CLI
│   └── seed/main.go       # Database seeding tool
├── internal/              # Private application code (DDD structure)
│   ├── app/               # Application bootstrap & modular DI container
│   ├── config/            # Viper configuration with enterprise toggles
│   ├── core/domain/       # Domain entities and business interfaces
│   ├── transport/http/    # HTTP handlers, middleware, WebSocket
│   ├── services/          # Business logic implementations
│   ├── workers/           # Background job processors
│   └── ee/                # Enterprise Edition features (build tag gated)
├── web/                   # Next.js 15 frontend
│   ├── src/app/          # App Router with route groups
│   ├── src/components/   # React components (shadcn/ui + custom)
│   ├── src/hooks/        # Custom React hooks
│   ├── src/lib/          # API clients and utilities
│   └── src/store/        # Zustand state management
├── migrations/            # Multi-database migrations
│   ├── postgres/         # Comprehensive schema (users, orgs, projects)
│   └── clickhouse/       # 5 table structure (metrics, events, traces, logs, routing)
└── docs/                 # Comprehensive documentation
    ├── ARCHITECTURE.md   # Detailed system architecture
    ├── DEVELOPMENT.md    # 880+ line development guide
    ├── ENTERPRISE.md     # 260+ page enterprise documentation
    └── enterprise/       # Specialized guides (SSO, RBAC, Compliance)
```

### Domain-Driven Architecture
Following DDD principles with clean boundaries:

**Core Domains**:
- **User Domain**: User management, profiles, authentication
- **Organization Domain**: Multi-tenant structure with role-based access
- **Auth Domain**: JWT sessions, API keys, blacklisted tokens
- **Routing Domain**: AI provider management and intelligent routing
- **Billing Domain**: Usage tracking, invoicing, subscription management
- **Observability Domain**: Metrics, events, tracing, analytics (ClickHouse)

**Architecture Layers**:
- **Domain Layer**: Entities, repository interfaces, service contracts
- **Infrastructure Layer**: Database implementations, external API clients
- **Transport Layer**: HTTP/WebSocket handlers, middleware
- **Application Layer**: Dependency injection, bootstrap, graceful shutdown

### Database Strategy
**PostgreSQL** (Transactional):
- Users, organizations, projects, environments
- API keys, auth sessions, configurations
- Single comprehensive schema with proper foreign keys

**ClickHouse** (Analytics):
- `metrics` - Real-time platform metrics (90-day TTL)
- `events` - Business/system events (180-day TTL)
- `traces` - Distributed tracing (30-day TTL)
- `request_logs` - API logging (60-day TTL)
- `ai_routing_metrics` - Provider routing decisions (365-day TTL)

**Redis** (Cache & Queues):
- JWT sessions and token blacklisting
- Rate limiting and quota management
- Background job queues (analytics, notifications)
- Semantic cache for AI responses
- Real-time WebSocket pub/sub

### Enterprise Architecture
**Open-Core Model**: 70% cost advantage vs competitors (Brokle Pro $29 vs Portkey $99+)

**Build Tag System**:
```bash
# OSS build (default)
go build ./cmd/server

# Enterprise build with full features
go build -tags="enterprise" ./cmd/server
```

**Enterprise Features**:
- **SSO Integration**: SAML 2.0, OIDC/OAuth2 with auto-provisioning
- **Advanced RBAC**: Custom roles, granular permissions, hierarchical scopes
- **Compliance**: SOC2/HIPAA/GDPR controls, audit trails, data retention
- **Predictive Analytics**: ML-powered insights, cost forecasting, anomaly detection
- **Custom Dashboards**: Drag-drop builder, executive reporting, embedded analytics

### Frontend Architecture (Next.js 15)
**Modern Stack**:
- **Next.js 15**: App Router with Server Components and Turbopack
- **React 19**: Latest React features with concurrent rendering
- **TypeScript**: Strict type checking throughout
- **Tailwind CSS v4**: Custom design system
- **shadcn/ui**: Radix UI primitives with custom styling
- **Zustand**: Lightweight state management
- **TanStack Query**: Server state management with real-time updates
- **React Hook Form + Zod**: Type-safe form validation

**Route Structure**:
```
src/app/
├── (auth)/               # Auth route group (/auth/*)
├── (dashboard)/          # Dashboard routes
│   ├── [orgSlug]/       # Organization-scoped routes
│   └── settings/        # User settings
└── layout.tsx           # Root layout with providers
```

## Key Development Patterns

### Adding New Features
Follow the established DDD pattern:

1. **Define Domain**: Create entity and interfaces in `internal/core/domain/{domain}/`
2. **Repository**: Implement data access in `internal/infrastructure/repository/`
3. **Service**: Add business logic in `internal/services/`
4. **Handler**: Create HTTP endpoints in `internal/transport/http/handlers/`
5. **Routes**: Wire routes in router configuration
6. **Tests**: Write unit, integration, and E2E tests
7. **Migration**: Add database changes if needed

### Enterprise Feature Development
Enterprise features use interface-based design with stub implementations:

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

### Configuration Management
**Viper-based** configuration with environment variable overrides:
- **Primary Source**: Environment variables
- **Fallback**: `.env` file in project root
- **Structure**: Nested configuration objects
- **Enterprise**: Build-tag controlled feature flags

### Migration System Safety
The migration CLI includes comprehensive safety features:
- **Confirmation Prompts**: All destructive operations require explicit 'yes'
- **Dry Run Mode**: Preview changes with `-dry-run` flag
- **Granular Control**: Target specific databases with `-db` flag
- **Health Monitoring**: Detailed status reporting with dirty state detection
- **Rollback Support**: Safe rollback with proper down migrations

### Error Handling & Logging
**Structured Logging** with correlation IDs:
```go
import "github.com/sirupsen/logrus"

logger := logrus.WithFields(logrus.Fields{
    "user_id": userID,
    "request_id": requestID,
    "organization_id": orgID,
})
logger.Info("Processing user request")
```

**Error Patterns**:
```go
// Service layer - wrap errors with context
if err != nil {
    return errors.Wrap(err, "failed to create user")
}

// Handler layer - proper HTTP status and logging
if err != nil {
    http.Error(w, "Internal server error", http.StatusInternalServerError)
    log.Error("create user failed", "error", err)
    return
}
```

## Real-time & Background Processing

### WebSocket Architecture
- **Hub Pattern**: Central connection manager in `pkg/websocket/`
- **Room Subscriptions**: Users subscribe to relevant data streams
- **Event Types**: `metrics.updated`, `usage.threshold`, `routing.decision`, `system.alert`
- **Authentication**: JWT token validation on connection

### Background Workers
- **Goroutine Pool**: Controlled concurrency for job processing
- **Redis Queues**: Persistent job storage and distribution
- **Job Categories**: High/Medium/Low priority with different processing rates
- **Worker Types**: Analytics aggregation, email notifications, webhook delivery

## API Design Principles

### OpenAI Compatibility
- **Drop-in Replacement**: `/v1/chat/completions`, `/v1/embeddings`, `/v1/models`
- **Extended Responses**: Additional Brokle metadata (provider, cost, quality score)
- **Provider Abstraction**: Unified interface across 250+ LLM providers

### Management APIs
- **REST Design**: Standard HTTP methods with proper status codes
- **Multi-tenant**: Organization-scoped resources with RBAC
- **Pagination**: Cursor-based pagination for large datasets
- **Real-time**: WebSocket integration for live updates

### Performance Targets
- **P95 Response Time**: <200ms for API endpoints
- **P99 Response Time**: <500ms for complex analytics queries
- **WebSocket Latency**: <100ms for real-time updates
- **Throughput**: 10,000+ requests/second with proper caching

## Important Development Notes

### Docker Development
```bash
# Complete stack with infrastructure
make docker-dev       # Uses docker-compose with override files

# Production deployment
make docker-prod      # Production-ready containers

# Resource management
make docker-clean     # Clean volumes and orphaned containers
make health           # Check all service health
```

### Service Dependencies
The platform requires specific service startup order:
1. **Databases**: PostgreSQL, ClickHouse, Redis must be healthy
2. **Migrations**: All databases must be migrated before API start
3. **Seeding**: Development data loaded for proper testing
4. **API Server**: Starts with dependency health checks
5. **Frontend**: Connects to API with proper proxy configuration

### Testing Strategy
- **Unit Tests**: Business logic isolation with interface mocks
- **Integration Tests**: Real database connections with cleanup
- **E2E Tests**: Complete user flows through HTTP API
- **Load Tests**: Performance validation for gateway routing

### Monitoring & Observability
- **Prometheus Metrics**: Exposed at `/metrics` endpoint
- **Distributed Tracing**: Request correlation across all operations
- **Health Checks**: `/health`, `/health/db`, `/health/cache` with detailed status
- **Structured Logs**: JSON format with correlation IDs

## Enterprise Considerations

### License Tiers
- **Free (OSS)**: 10K requests/month, 5 users, community support
- **Pro ($29/month)**: 100K requests, 10 users, RBAC, email support
- **Business ($99/month)**: 1M requests, SSO, compliance, predictive analytics
- **Enterprise (Custom)**: Unlimited scale, on-premise, dedicated support

### Compliance Features
- **SOC 2 Type II**: Security controls for availability, confidentiality, integrity
- **HIPAA**: Health information privacy and security controls
- **GDPR**: Data protection with right to deletion workflows
- **PII Anonymization**: Automatic detection and anonymization

### Enterprise Support
- **SLA Guarantees**: 4-hour response for Enterprise, 24/7 on-call
- **Dedicated Success Manager**: For Enterprise tier customers
- **Professional Services**: Migration, training, architecture review
- **On-Premise Deployment**: Air-gapped, Kubernetes, custom infrastructure

## Key Documentation References

For comprehensive technical details, refer to:
- **[docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)** - 880+ line development guide with detailed patterns
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Complete system architecture with performance targets
- **[docs/API.md](docs/API.md)** - Full API reference with WebSocket events
- **[docs/ENTERPRISE.md](docs/ENTERPRISE.md)** - 260+ page enterprise guide
- **[docs/enterprise/](docs/enterprise/)** - Specialized guides for SSO, RBAC, compliance
- **[CLAUDE.md](CLAUDE.md)** - Comprehensive technical reference (600+ lines)

The CLAUDE.md file contains extensive implementation details including migration architecture, enterprise patterns, frontend development workflows, and troubleshooting guides.

## Troubleshooting

### Common Issues
```bash
# Database connection problems
make migrate-status      # Check database health
go run cmd/migrate/main.go -db postgres drop  # Reset dirty state (DESTRUCTIVE)

# Port conflicts
lsof -ti:8080            # Check port usage
kill -9 $(lsof -ti:8080) # Kill conflicting processes

# Service health monitoring
make health              # Check all service status
make logs                # View aggregated logs
make logs-api            # API server specific logs
make logs-db             # Database service logs
```

### Performance Debugging
- **Go Profiling**: Built-in pprof endpoints for CPU/memory analysis
- **Database Query Analysis**: EXPLAIN plans for optimization
- **Redis Monitoring**: Built-in Redis CLI for cache analysis
- **Frontend Performance**: Next.js built-in performance monitoring

This platform represents a production-grade AI infrastructure solution with enterprise-ready features, comprehensive observability, and a clean, maintainable architecture designed for both open source adoption and commercial scaling.
