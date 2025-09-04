# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Platform Overview

**Brokle** - Complete AI Infrastructure Platform that provides everything AI teams need to build, deploy, and scale production AI applications.

### Core Mission
Build the unified platform for AI teams to build, deploy, and scale production AI applications. Position as "The Stripe for AI Infrastructure" - handling all complexity so developers can focus on building great AI applications.

### Key Features
- **Intelligent Gateway**: OpenAI-compatible proxy with smart routing across 250+ LLM providers
- **Advanced Observability**: 40+ AI-specific metrics with real-time monitoring
- **Semantic Caching**: Vector-based caching for 95% cost reduction potential  
- **Cost Optimization**: 30-50% reduction in LLM costs through intelligent routing
- **Multi-tenant Architecture**: Organization → Project → Environment isolation

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
```bash
# Run database migrations
make migrate-up

# Rollback one migration
make migrate-down

# Check migration status
make migrate-status

# Seed with development data
make seed-dev

# Create new migration
make create-migration DB=postgres NAME=add_users_table
make create-migration DB=clickhouse NAME=add_metrics_table

# Reset all databases (WARNING: destroys data)
make migrate-reset

# Database shell access
make shell-db          # PostgreSQL
make shell-redis       # Redis CLI
make shell-clickhouse  # ClickHouse client
```

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

### OpenAI-Compatible Gateway
The monolith provides OpenAI-compatible endpoints:
- `POST /v1/chat/completions` - Chat completions
- `POST /v1/completions` - Text completions  
- `POST /v1/embeddings` - Text embeddings
- `GET /v1/models` - Available models

### Management APIs
- `/api/v1/auth/*` - Authentication & user management
- `/api/v1/organizations/*` - Organization management
- `/api/v1/projects/*` - Project management
- `/api/v1/analytics/*` - Metrics & reporting
- `/api/v1/billing/*` - Usage & billing

### Real-time APIs
- `/ws` - WebSocket connections for real-time updates
- `/api/v1/streaming/*` - Server-sent events

## Data Architecture

### Primary Database (PostgreSQL)
Single database with domain-separated tables:
- `users`, `auth_sessions` - Authentication & user management
- `organizations`, `organization_members` - Multi-tenant structure  
- `projects`, `environments` - Project and environment management
- `api_keys` - API key management and scoping
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
Use structured errors throughout:

```go
import "github.com/pkg/errors"

// Service layer
if err != nil {
    return errors.Wrap(err, "failed to create user")
}

// Handler layer
if err != nil {
    http.Error(w, "Internal server error", http.StatusInternalServerError)
    log.Error("create user failed", "error", err)
    return
}
```

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

### Unit Tests
Test business logic in isolation:

```go
func TestUserService_CreateUser(t *testing.T) {
    mockRepo := &MockUserRepository{}
    service := NewUserService(mockRepo)
    
    err := service.CreateUser(ctx, req)
    assert.NoError(t, err)
    assert.Called(t, mockRepo.Create)
}
```

### Integration Tests
Test with real database:

```go
func TestUserHandler_CreateUser(t *testing.T) {
    db := setupTestDB(t)
    handler := setupHandler(db)
    
    resp := httptest.NewRecorder()
    handler.CreateUser(resp, req)
    
    assert.Equal(t, http.StatusCreated, resp.Code)
}
```

### E2E Tests
Test complete user flows:

```bash
make test-e2e
```

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
- **Database migrations**: Run `make migrate-status` to check state
- **Enterprise build errors**: Ensure proper build tags usage
- **WebSocket connection issues**: Check CORS and proxy settings