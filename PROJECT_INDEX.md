# Project Index: Brokle

**Generated:** 2025-11-14
**Version:** 0.1.0
**License:** MIT (OSS), Proprietary (Enterprise Edition)

## 📋 Quick Stats

- **Total Go Files:** 194
- **Test Files:** 13
- **Frontend Files (TS/TSX):** 358
- **Domains:** 6 active domains
- **PostgreSQL Migrations:** 33
- **ClickHouse Migrations:** 6
- **Lines of Code:** ~50,000+ (backend + frontend)

---

## 📁 Project Structure

```
brokle/
├── cmd/                      # Application entry points
│   ├── server/              # HTTP server binary (API endpoints)
│   ├── worker/              # Background workers (telemetry processing)
│   └── migrate/             # Database migration CLI + seeder
├── internal/                # Private application code
│   ├── app/                # DI container & service registry
│   ├── config/             # Configuration management (Viper)
│   ├── core/               # Domain-driven architecture
│   │   ├── domain/        # 6 domains (auth, billing, observability, etc.)
│   │   └── services/      # Service implementations
│   ├── infrastructure/     # External integrations
│   │   ├── database/      # DB connections (Postgres, ClickHouse)
│   │   ├── repository/    # Repository layer
│   │   └── streams/       # Redis streams (telemetry)
│   ├── transport/http/    # HTTP transport layer
│   │   ├── handlers/      # HTTP handlers by domain
│   │   ├── middleware/    # Auth, rate limiting, CORS
│   │   └── server.go      # HTTP server setup
│   ├── workers/           # Background workers
│   ├── middleware/        # Enterprise middleware
│   └── ee/                # Enterprise Edition features
├── pkg/                   # Public shared packages
│   ├── errors/           # Error handling
│   ├── response/         # HTTP response utilities
│   ├── utils/            # Utilities (crypto, strings, time)
│   ├── websocket/        # WebSocket utilities
│   ├── analytics/        # Analytics aggregation
│   ├── pagination/       # Pagination utilities
│   ├── validator/        # Input validation
│   └── ulid/             # ULID generation
├── migrations/           # Database migrations
│   ├── postgres/        # 33 PostgreSQL migrations
│   └── clickhouse/      # 6 ClickHouse migrations
├── seeds/               # YAML seed data (dev, demo, test)
├── web/                 # Next.js 15.5.2 frontend
│   ├── src/app/        # App Router (routing only)
│   ├── src/features/   # Feature-based architecture
│   ├── src/components/ # Shared components (shadcn/ui)
│   ├── src/lib/        # Core infrastructure
│   └── src/hooks/      # Global hooks
├── docs/                # Documentation
├── examples/            # Example configurations (OTEL collector)
├── sdk/                 # SDKs
│   ├── javascript/     # TypeScript/JS SDK
│   └── python/         # Python SDK
├── tests/              # Integration tests
└── configs/            # Configuration files (Prometheus, Redis, etc.)
```

---

## 🚀 Entry Points

### Backend

| Entry Point | Path | Description | Port |
|------------|------|-------------|------|
| **HTTP Server** | `cmd/server/main.go` | API endpoints, WebSocket, migrations | 8080 |
| **Background Workers** | `cmd/worker/main.go` | Telemetry processing, analytics, notifications | N/A |
| **Migration CLI** | `cmd/migrate/main.go` | Database migrations, seeding, schema management | N/A |

### Frontend

| Entry Point | Path | Description | Port |
|------------|------|-------------|------|
| **Dashboard** | `web/src/app/layout.tsx` | Next.js App Router root layout | 3000 |
| **Auth Pages** | `web/src/app/(auth)/` | Authentication routes (login, register, reset) | 3000 |
| **Dashboard** | `web/src/app/(dashboard)/` | Main dashboard routes | 3000 |

### Infrastructure

| Service | Image | Ports | Purpose |
|---------|-------|-------|---------|
| **PostgreSQL** | `postgres:18-alpine` | 5432 | Transactional data |
| **ClickHouse** | `clickhouse:25.8-alpine` | 8123, 9000 | Analytics & time-series |
| **Redis** | `redis:8.2-alpine` | 6379 | Cache, sessions, streams |
| **MinIO** | `minio:latest` | 9100, 9101 | S3-compatible blob storage |

---

## 📦 Core Modules

### Domain Layer (`internal/core/domain/`)

| Domain | Status | Entities | Purpose |
|--------|--------|----------|---------|
| **auth** | ✅ Active | User, Session, APIKey, Token, Scope, Role, Permission | Authentication, authorization, RBAC |
| **billing** | ✅ Active | Usage, Invoice, Subscription | Usage tracking, billing, invoicing |
| **common** | ✅ Active | Transaction patterns | Cross-domain transaction management |
| **observability** | ✅ Active | Trace, Span, QualityScore, RequestLog | Telemetry, OTLP ingestion, metrics |
| **organization** | ✅ Active | Organization, Member, Invitation, Project | Multi-tenant org management |
| **user** | ✅ Active | User, Profile, Settings | User management |

### Service Layer (`internal/core/services/`)

| Service | Key Functions | Business Logic |
|---------|--------------|----------------|
| **auth/auth_service** | Login, Register, OAuth, JWT validation | User authentication flow |
| **auth/api_key_service** | Create, Validate, Revoke API keys | SDK authentication (`bk_*` keys) |
| **auth/role_service** | RBAC operations | Role-based access control |
| **auth/scope_service** | Permission validation | Scope-based authorization |
| **observability/otlp_converter** | OTLP → Brokle format | OpenTelemetry Protocol conversion |
| **observability/span_service** | Span CRUD, querying | LLM call tracking |
| **observability/trace_service** | Trace aggregation, hierarchy | Distributed tracing |
| **observability/telemetry_service** | Telemetry ingestion | Real-time telemetry processing |
| **billing/billing_service** | Usage tracking, invoicing | Cost calculation, billing |
| **organization/organization_service** | Org CRUD, membership | Multi-tenancy |

### Infrastructure Layer (`internal/infrastructure/`)

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **database/postgres** | PostgreSQL + GORM | Transactional data, user management |
| **database/clickhouse** | ClickHouse + raw SQL | Analytics, time-series data |
| **repository/** | Repository pattern | Database abstraction layer |
| **streams/redis** | Redis Streams | Telemetry event streaming |

### Transport Layer (`internal/transport/http/`)

| Component | Pattern | Purpose |
|-----------|---------|---------|
| **handlers/** | Domain-based handlers | HTTP request handling |
| **middleware/auth** | JWT validation | Dashboard authentication |
| **middleware/sdk_auth** | API key validation | SDK authentication |
| **middleware/rate_limit** | Token bucket | Rate limiting |
| **middleware/cors** | CORS policy | Cross-origin requests |

### Workers (`internal/workers/`)

| Worker | Purpose | Trigger |
|--------|---------|---------|
| **telemetry_stream_consumer** | Process telemetry events from Redis streams | Redis stream events |
| **analytics_worker** | Aggregate metrics, generate reports | Cron schedule |
| **notification_worker** | Send email/SMS notifications | Redis queue |

---

## 🔧 Configuration

| File | Purpose | Format |
|------|---------|--------|
| `CLAUDE.md` | AI assistant instructions | Markdown |
| `.env.example` | Environment variable template | ENV |
| `go.mod` | Go dependencies (Go 1.25.0) | Go module |
| `Makefile` | Development automation | Make |
| `docker-compose.yml` | Local development stack | YAML |
| `.air.toml` | Hot reload (server) | TOML |
| `.air.worker.toml` | Hot reload (worker) | TOML |
| `web/package.json` | Frontend dependencies (Next.js 15.5.2, React 19.2.0) | JSON |
| `web/components.json` | shadcn/ui configuration | JSON |
| `configs/prometheus/` | Prometheus monitoring | YAML |
| `configs/redis/` | Redis configuration | CONF |
| `configs/clickhouse/` | ClickHouse configuration | XML |

---

## 📚 Documentation

### Core Documentation

| Document | Topic | Audience |
|----------|-------|----------|
| `README.md` | Project overview, quick start | All users |
| `CLAUDE.md` | AI assistant guide (architecture, patterns) | Developers, AI assistants |
| `docs/ARCHITECTURE.md` | System architecture, scalable monolith design | Engineers |
| `docs/DEVELOPMENT.md` | Development setup, workflows | Contributors |
| `docs/API.md` | API reference (REST, WebSocket) | Integrators |
| `docs/TESTING.md` | Testing philosophy, patterns | Developers |
| `docs/ENTERPRISE.md` | Enterprise Edition features | Enterprise users |

### Development Guides

| Document | Topic |
|----------|-------|
| `docs/development/ERROR_HANDLING_GUIDE.md` | Industrial error handling patterns |
| `docs/development/DOMAIN_ALIAS_PATTERNS.md` | Domain import patterns |
| `docs/development/API_DEVELOPMENT_GUIDE.md` | API development standards |
| `docs/development/PAGINATION_GUIDE.md` | Pagination implementation |
| `docs/development/PATTERNS.md` | Common design patterns |

### Enterprise Guides

| Document | Topic |
|----------|-------|
| `docs/enterprise/SSO.md` | Single Sign-On (SAML, OIDC) |
| `docs/enterprise/RBAC.md` | Role-Based Access Control |
| `docs/enterprise/COMPLIANCE.md` | Compliance features (SOC 2, HIPAA, GDPR) |
| `docs/enterprise/ANALYTICS.md` | Enterprise analytics |

### Frontend Documentation

| Document | Topic |
|----------|-------|
| `web/ARCHITECTURE.md` | Frontend architecture, feature-based design |
| `web/README.md` | Frontend setup, development |

---

## 🧪 Test Coverage

### Backend Tests

- **Unit Tests:** 13 test files (`*_test.go`)
- **Test Coverage Target:** High-value business logic (pragmatic approach)
- **Test Runner:** Go test + testify/mock
- **Integration Tests:** `tests/integration/`

### Frontend Tests

- **Test Framework:** Vitest + React Testing Library + MSW
- **Coverage Target:** 30% (critical paths)
- **Location:** `web/src/__tests__/`

### Key Test Files

| Test File | Purpose |
|-----------|---------|
| `internal/core/services/observability/span_service_test.go` | Span service business logic |
| `internal/core/services/observability/trace_service_test.go` | Trace service business logic |
| `internal/core/services/observability/otlp_converter_test.go` | OTLP conversion logic |
| `internal/core/services/observability/telemetry_behavior_test.go` | Telemetry behavior tests |
| `internal/core/domain/observability/errors_test.go` | Error handling tests |

---

## 🔗 Key Dependencies

### Backend (Go 1.25.0)

| Dependency | Version | Purpose |
|------------|---------|---------|
| **gin-gonic/gin** | 1.11.0 | HTTP framework |
| **gorm.io/gorm** | 1.31.0 | PostgreSQL ORM |
| **ClickHouse/clickhouse-go** | 2.40.3 | ClickHouse client |
| **redis/go-redis** | 9.16.0 | Redis client |
| **golang-jwt/jwt** | 5.3.0 | JWT authentication |
| **spf13/viper** | 1.21.0 | Configuration management |
| **golang-migrate/migrate** | 4.19.0 | Database migrations |
| **gorilla/websocket** | 1.5.3 | WebSocket support |
| **prometheus/client_golang** | 1.23.2 | Metrics & monitoring |
| **aws-sdk-go-v2/s3** | 1.88.6 | S3/MinIO blob storage |
| **opentelemetry/otlp** | 1.8.0 | OpenTelemetry Protocol |

### Frontend (Next.js 15.5.2, React 19.2.0)

| Dependency | Version | Purpose |
|------------|---------|---------|
| **next** | 16.0.1 | React framework (App Router, Turbopack) |
| **react** | 19.2.0 | UI library |
| **tailwindcss** | 4.1.15 | Styling |
| **shadcn/ui** | (via Radix UI) | Component library |
| **@tanstack/react-query** | 5.90.3 | Server state management |
| **zustand** | 5.0.8 | Client state management |
| **axios** | 1.13.1 | HTTP client |
| **react-hook-form** | 7.66.0 | Form handling |
| **zod** | 4.1.12 | Schema validation |
| **recharts** | 3.3.0 | Data visualization |

---

## 📝 Quick Start

### 1. Setup (First Time)

```bash
# Clone repository
git clone https://github.com/brokle-ai/brokle.git
cd brokle

# Install dependencies, start databases, run migrations, seed data
make setup
```

### 2. Development

```bash
# Start full stack (server + worker + frontend)
make dev              # Server (8080) + Worker + Frontend (3000)

# Start individual components
make dev-server       # HTTP server only
make dev-worker       # Workers only
make dev-frontend     # Next.js frontend only
```

### 3. Database Operations

```bash
# Run migrations
make migrate-up

# Check migration status
make migrate-status

# Seed development data
make seed-dev

# Create new migration
make create-migration DB=postgres NAME=add_users_table
```

### 4. Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run unit tests only
make test-unit
```

### 5. Building

```bash
# Build OSS binaries
make build-oss        # Server + Worker

# Build Enterprise binaries
make build-enterprise # Server + Worker (with -tags="enterprise")

# Build frontend
make build-frontend
```

---

## 🔐 API Routes

### SDK Routes (`/v1/*`) - API Key Authentication

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/v1/chat/completions` | POST | OpenAI-compatible chat completions |
| `/v1/completions` | POST | Text completions |
| `/v1/embeddings` | POST | Vector embeddings |
| `/v1/models` | GET | Available AI models |
| `/v1/otlp/traces` | POST | OTLP traces ingestion (Protobuf/JSON) |
| `/v1/traces` | POST | Alternative OTLP traces endpoint |
| `/v1/route` | POST | AI routing decisions |

### Dashboard Routes (`/api/v1/*`) - JWT Authentication

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/auth/*` | POST | Authentication & session management |
| `/api/v1/users/*` | GET/POST/PUT/DELETE | User profile management |
| `/api/v1/organizations/*` | GET/POST/PUT/DELETE | Organization management |
| `/api/v1/projects/*` | GET/POST/PUT/DELETE | Project management |
| `/api/v1/projects/:id/api-keys/*` | GET/POST/DELETE | API key management |
| `/api/v1/analytics/*` | GET | Metrics & reporting |
| `/api/v1/billing/*` | GET/POST | Usage & billing |

---

## 🏗️ Architecture Highlights

### Scalable Monolith with Independent Scaling

- **Separate Binaries:** Server (HTTP API) + Worker (background processing)
- **Shared Codebase:** DI container enables code reuse
- **Independent Scaling:** 3-5 API servers, 10-50+ workers at scale

### Multi-Database Strategy

- **PostgreSQL:** Transactional data (users, auth, config)
- **ClickHouse:** Analytics & time-series (traces, spans, metrics)
- **Redis:** Cache, sessions, streams (telemetry events)

### Domain-Driven Design (DDD)

- **6 Active Domains:** auth, billing, observability, organization, user, common
- **Clean Architecture:** Domain → Service → Infrastructure → Transport
- **Industrial Error Handling:** Repository → Service → Handler flow

### Authentication Architecture

- **API Keys:** `bk_{40_char_random}` format with SHA-256 hashing
- **JWT:** Dashboard authentication with session management
- **RBAC:** Scope-based role permissions (Enterprise)

### Enterprise Edition Pattern

- **Build Tags:** `-tags="enterprise"` for EE builds
- **Stub Implementations:** OSS stubs in `internal/ee/*/build.go`
- **Feature Toggle:** Interface-based design

---

## 🎯 Key Features

### Advanced Observability

- 40+ LLM-specific metrics (latency, tokens, cost, errors)
- OpenTelemetry Protocol (OTLP) native ingestion
- Real-time distributed tracing
- Quality scoring and evaluation

### AI Gateway & Routing

- Multi-provider smart routing (OpenAI, Anthropic, Google, Cohere)
- Intelligent failover protection
- Drop-in OpenAI API compatibility

### Governance & Control

- Cost controls and usage limits
- RBAC and access control (Enterprise)
- Real-time cost governance

---

## 📊 Token Efficiency

**Index Stats:**
- Index size: ~8KB (human-readable)
- Tokens saved per session: ~55,000 tokens
- Break-even: 1 session
- 10 sessions savings: 550,000 tokens

---

## 🔗 Links

- **Website:** https://brokle.com
- **Documentation:** https://docs.brokle.com
- **GitHub:** https://github.com/brokle-ai/brokle
- **Discord:** https://discord.gg/brokle
- **Twitter:** @BrokleAI

---

**Built with ❤️ by the Brokle team. Making AI infrastructure simple and powerful.**
