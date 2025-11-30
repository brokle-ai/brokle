# Development Patterns

This guide covers common development patterns used throughout the Brokle codebase.

## Table of Contents
- [API Key Management](#api-key-management)
- [Authentication Middleware](#authentication-middleware)
- [Clean Architecture](#clean-architecture)
- [Error Handling](#error-handling)
- [Logging](#logging)
- [Configuration Management](#configuration-management)
- [Enterprise Edition Pattern](#enterprise-edition-pattern)

## API Key Management

### Industry-Standard API Key System

The platform uses industry-standard API keys (following GitHub/Stripe/OpenAI patterns) for SDK authentication.

**API Key Format:**
```
bk_{40_char_random_secret}
```
- **Prefix**: `bk_` (Brokle identifier)
- **Secret**: 40 characters of cryptographically secure random data
- **Example**: `bk_AbCdEfGhIjKlMnOpQrStUvWxYz0123456789AbCd`

### API Key Utilities

Location: `internal/core/domain/auth/apikey_utils.go`

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

### Key Features
- **Industry Standard**: Pure random format matching GitHub, Stripe, OpenAI
- **Secure Storage**: SHA-256 hashing (deterministic, enables O(1) lookup)
- **O(1) Validation**: Direct hash lookup with unique database index
- **Project Association**: Project ID stored in database, not embedded in key
- **Security Best Practice**: No sensitive data embedded in key

## Authentication Middleware

### SDK Authentication (API Keys)

For SDK routes requiring API key authentication:

```go
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

**Context Keys:**
- `SDKAuthContextKey` - Full authentication context
- `APIKeyIDKey` - API key identifier
- `ProjectIDKey` - Project ID (stored as pointer)
- `EnvironmentKey` - Environment tag from header

### Dashboard Authentication (JWT)

For dashboard routes requiring JWT authentication:

```go
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

### Middleware Architecture

Location: `internal/transport/http/middleware/`

**SDKAuthMiddleware** (`sdk_auth.go`):
```go
type SDKAuthMiddleware struct {
    apiKeyService auth.APIKeyService
    logger        *slog.Logger
}

// RequireSDKAuth validates API keys for SDK routes
func (m *SDKAuthMiddleware) RequireSDKAuth() gin.HandlerFunc {
    // Extracts and validates API key
    // Stores authentication context in Gin context
    // Handles both X-API-Key and Authorization headers
}
```

**Rate Limiting Strategy:**
```go
// API key-based rate limiting for SDK routes
router.Use(rateLimitMiddleware.RateLimitByAPIKey())

// IP-based rate limiting for dashboard routes
router.Use(rateLimitMiddleware.RateLimitByIP())

// User-based rate limiting after JWT authentication
protectedRoutes.Use(rateLimitMiddleware.RateLimitByUser())
```

## Clean Architecture

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

### Layer Responsibilities

**Repository Layer** (`internal/infrastructure/repository/`)
- Database operations
- Domain error mapping
- Data persistence

**Service Layer** (`internal/services/`)
- Business logic
- AppError constructors
- Pure domain operations

**Handler Layer** (`internal/transport/http/handlers/`)
- HTTP request/response handling
- Input validation
- Centralized error responses

## Error Handling

**📖 CRITICAL**: Use the comprehensive [Development Error Handling Guides](./):

- **[ERROR_HANDLING_GUIDE.md](ERROR_HANDLING_GUIDE.md)** - Complete industrial patterns
- **[DOMAIN_ALIAS_PATTERNS.md](DOMAIN_ALIAS_PATTERNS.md)** - Professional import patterns
- **[ERROR_HANDLING_QUICK_REFERENCE.md](ERROR_HANDLING_QUICK_REFERENCE.md)** - Developer cheat sheet

### Clean Architecture Error Flow

```
Repository (Domain Errors) → Service (AppErrors) → Handler (HTTP Response)
```

### Core Principles

- **Repository Layer**: Domain errors with proper wrapping
- **Service Layer**: AppError constructors (NewUnauthorizedError, NewNotFoundError, etc.)
- **Handler Layer**: Centralized `response.Error(c, err)` handling
- **Decorator Pattern**: Cross-cutting concerns (audit, logging) handled separately
- **Zero Logging**: Core services focus on pure business logic

### Example Implementation

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

### Key Requirements

- ❌ **No fmt.Errorf/errors.New** in services - use AppError constructors
- ❌ **No logging** in core services - use decorator pattern
- ✅ **Domain error mapping** at repository layer
- ✅ **Structured AppErrors** at service layer
- ✅ **Clean separation** of business logic and cross-cutting concerns

## Logging

Use structured logging with correlation IDs:

```go
import "log/slog"

logger := slog.With(
    "request_id", middleware.GetRequestID(ctx),
    "user_id", auth.GetUserID(ctx),
)

logger.Info("user created successfully", "user_id", user.ID)
```

### Logging Guidelines

**DO:**
- Use structured logging with `slog`
- Include correlation IDs (request_id, user_id, etc.)
- Log at appropriate levels (Info, Warn, Error)
- Log business events and errors

**DON'T:**
- Add logging to core service business logic (use decorator pattern)
- Log sensitive data (passwords, API keys, tokens)
- Use fmt.Println or log.Println
- Over-log (avoid debug logs in production)

## Configuration Management

The application uses Viper for configuration:

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

### Configuration Best Practices

1. **Environment Variables**: Use for deployment-specific values
2. **Defaults**: Provide sensible defaults for development
3. **Validation**: Validate required configuration on startup
4. **Secrets**: Never commit secrets to version control

## Enterprise Edition Pattern

The codebase uses build tags for enterprise features:

```bash
# OSS build (default)
go build ./cmd/server

# Enterprise build
go build -tags="enterprise" ./cmd/server
```

### Directory Structure

Enterprise features are in `internal/ee/` with stub implementations for OSS builds:
- `internal/ee/sso/` - Single Sign-On
- `internal/ee/rbac/` - Role-Based Access Control
- `internal/ee/compliance/` - Compliance features
- `internal/ee/analytics/` - Enterprise analytics

### Implementation Pattern

```go
// internal/ee/sso/build.go (OSS)
func New() SSOProvider {
    return &stubSSOProvider{}
}

// internal/ee/sso/build_enterprise.go (Enterprise)
// +build enterprise

func New() SSOProvider {
    return &enterpriseSSOProvider{}
}
```

### Usage in Code

```go
// Service initialization (works for both OSS and Enterprise)
ssoProvider := sso.New()

// Enterprise features automatically available when built with -tags="enterprise"
if err := ssoProvider.ConfigureSAML(config); err != nil {
    // Handle error
}
```

**📖 See Also:**
- [ENTERPRISE.md](../ENTERPRISE.md) - Complete enterprise documentation
- [CODING_STANDARDS.md](../CODING_STANDARDS.md) - Code style and standards
- [DEVELOPMENT.md](../DEVELOPMENT.md) - Development workflows
