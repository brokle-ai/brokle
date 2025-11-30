# Enterprise Developer Guide

## Table of Contents
- [Architecture Overview](#architecture-overview)
- [Build System](#build-system)
- [Configuration Management](#configuration-management)
- [Feature Implementation](#feature-implementation)
- [Middleware System](#middleware-system)
- [License Service](#license-service)
- [Error Handling](#error-handling)
- [Testing Strategy](#testing-strategy)
- [Development Workflow](#development-workflow)
- [Debugging](#debugging)

## Architecture Overview

Brokle Enterprise uses a **build tag-based architecture** to cleanly separate OSS and Enterprise features while maintaining a single codebase. This approach provides:

- **Compile-time separation**: OSS and Enterprise builds are completely separate binaries
- **Clean interfaces**: Both versions implement the same interfaces for seamless compatibility
- **Zero runtime overhead**: No feature flags or runtime checks in production
- **Easy maintenance**: Single codebase with clear separation boundaries

### Core Architectural Principles

1. **Interface-first design**: All enterprise features define interfaces that both stub and real implementations satisfy
2. **Build tag separation**: `//go:build enterprise` and `//go:build !enterprise` tags control compilation
3. **Graceful degradation**: OSS version provides meaningful stub implementations
4. **Professional errors**: Enterprise-gated features return business-appropriate error messages

## Build System

### Build Commands

```bash
# OSS build (default)
make build                    # Builds OSS version
make build-backend           # Builds OSS backend only  
go build -o brokle ./cmd/server

# Enterprise build
make build-enterprise        # Builds Enterprise version
make build-backend-enterprise # Builds Enterprise backend only
go build -tags="enterprise" -o brokle-enterprise ./cmd/server

# Development builds (faster, with debug info)
make build-dev              # OSS development build
make build-dev-enterprise   # Enterprise development build

# Build both versions
make build-all              # Builds both OSS and Enterprise
```

### Makefile Targets

```makefile
# Enhanced build system with enterprise support
build: build-oss ## Build OSS version by default

build-oss: build-backend-oss build-frontend ## Build OSS backend and frontend
	@echo "✅ OSS build complete!"

build-enterprise: build-backend-enterprise build-frontend ## Build Enterprise backend and frontend
	@echo "✅ Enterprise build complete!"

build-backend-oss: ## Build Go API server (OSS version)
	@echo "🔨 Building Go API server (OSS)..."
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/brokle-oss cmd/server/main.go

build-backend-enterprise: ## Build Go API server (Enterprise version)
	@echo "🔨 Building Go API server (Enterprise)..."
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags="enterprise" -ldflags="-w -s" -o bin/brokle-enterprise cmd/server/main.go
```

### Docker Builds

```dockerfile
# Dockerfile (OSS version)
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o brokle cmd/server/main.go

FROM alpine:latest
COPY --from=builder /app/brokle .
CMD ["./brokle"]

# Dockerfile.enterprise
FROM golang:1.25-alpine AS builder
WORKDIR /app  
COPY . .
RUN CGO_ENABLED=0 go build -tags="enterprise" -ldflags="-w -s" -o brokle-enterprise cmd/server/main.go

FROM alpine:latest
COPY --from=builder /app/brokle-enterprise .
CMD ["./brokle-enterprise"]
```

## Configuration Management

### Build Tag Configuration Files

The configuration system uses build tags to include different config structs:

```go
// internal/config/ee.go
//go:build enterprise

package config

type EnterpriseConfig struct {
    License     LicenseConfig     `mapstructure:"license"`
    SSO         SSOConfig         `mapstructure:"sso"`
    RBAC        RBACConfig        `mapstructure:"rbac"`
    Compliance  ComplianceConfig  `mapstructure:"compliance"`
    Analytics   AnalyticsConfig   `mapstructure:"analytics"`
    Support     SupportConfig     `mapstructure:"support"`
}

type SSOConfig struct {
    Enabled     bool              `mapstructure:"enabled"`
    Provider    string            `mapstructure:"provider"`
    MetadataURL string            `mapstructure:"metadata_url"`
    EntityID    string            `mapstructure:"entity_id"`
    Certificate string            `mapstructure:"certificate"`
    Attributes  map[string]string `mapstructure:"attributes"`
}
// ... full enterprise config structs
```

```go
// internal/config/ee_stub.go
//go:build !enterprise

package config

// Minimal stub structs for OSS builds
type EnterpriseConfig struct {
    License     LicenseConfig     `mapstructure:"license"`
    SSO         SSOConfig         `mapstructure:"sso"`
    RBAC        RBACConfig        `mapstructure:"rbac"`
    Compliance  ComplianceConfig  `mapstructure:"compliance"`
    Analytics   AnalyticsConfig   `mapstructure:"analytics"`
    Support     SupportConfig     `mapstructure:"support"`
}

type SSOConfig struct {
    Enabled bool `mapstructure:"enabled"`
}
// ... minimal stub structs
```

### License Wrapper

The license wrapper provides enhanced license management:

```go
// internal/config/license.go
type LicenseWrapper struct {
    config *Config
}

func (lw *LicenseWrapper) GetEffectiveLicense() *LicenseConfig {
    // Apply tier-based defaults
    // Validate license configuration  
    // Return effective license with all defaults applied
}

func (lw *LicenseWrapper) ValidateLicense() error {
    // Comprehensive license validation
    // Check tier validity, limits, expiration, features
}

func (lw *LicenseWrapper) GetTierLimits(tier string) (*LicenseConfig, error) {
    // Return standard limits for free/pro/business/enterprise tiers
}
```

### Environment Variable Mapping

```bash
# License configuration
BROKLE_ENTERPRISE_LICENSE_KEY="license-jwt-token"
BROKLE_ENTERPRISE_LICENSE_TYPE="business"
BROKLE_ENTERPRISE_LICENSE_OFFLINE_MODE="false"

# SSO configuration
BROKLE_ENTERPRISE_SSO_ENABLED="true"
BROKLE_ENTERPRISE_SSO_PROVIDER="saml"
BROKLE_ENTERPRISE_SSO_METADATA_URL="https://idp.example.com/metadata"

# RBAC configuration  
BROKLE_ENTERPRISE_RBAC_ENABLED="true"

# Compliance configuration
BROKLE_ENTERPRISE_COMPLIANCE_ENABLED="true"
BROKLE_ENTERPRISE_COMPLIANCE_SOC2_COMPLIANCE="true"
BROKLE_ENTERPRISE_COMPLIANCE_PII_ANONYMIZATION="true"
```

## Feature Implementation

### Interface-Based Design Pattern

All enterprise features follow a consistent interface pattern:

```go
// internal/ee/sso/sso.go
type SSOProvider interface {
    Authenticate(ctx context.Context, token string) (*User, error)
    GetLoginURL(ctx context.Context) (string, error)
    ValidateAssertion(ctx context.Context, assertion string) (*User, error)
    GetSupportedProviders(ctx context.Context) ([]string, error)
    ConfigureProvider(ctx context.Context, provider, config string) error
}

// Stub implementation (OSS)
type StubSSO struct{}

func New() SSOProvider {
    return &StubSSO{}
}

func (s *StubSSO) Authenticate(ctx context.Context, token string) (*User, error) {
    return nil, errors.New("SSO authentication requires Enterprise license")
}
```

### Enterprise Service Directory Structure

```
internal/ee/
├── sso/
│   ├── sso.go              # Interface + stub implementation
│   ├── build.go            # Build constraints and factory (OSS)
│   └── build_enterprise.go # Real implementation (Enterprise)
├── rbac/
│   ├── rbac.go             # Interface + stub implementation
│   ├── build.go            # Build constraints and factory (OSS)
│   └── build_enterprise.go # Real implementation (Enterprise)
├── compliance/
│   ├── compliance.go       # Interface + stub implementation
│   ├── build.go            # Build constraints and factory (OSS)
│   └── build_enterprise.go # Real implementation (Enterprise)
└── analytics/
    ├── analytics.go        # Interface + stub implementation
    ├── build.go            # Build constraints and factory (OSS)
    └── build_enterprise.go # Real implementation (Enterprise)
```

### Adding New Enterprise Features

1. **Define the Interface**:
```go
// internal/ee/newfeature/newfeature.go
type NewFeature interface {
    DoSomething(ctx context.Context, param string) (*Result, error)
    Configure(ctx context.Context, config *Config) error
}

type StubNewFeature struct{}

func New() NewFeature {
    return &StubNewFeature{}  
}

func (s *StubNewFeature) DoSomething(ctx context.Context, param string) (*Result, error) {
    return nil, errors.New("New feature requires Enterprise license")
}
```

2. **Create Build Constraint Files**:
```go
// internal/ee/newfeature/build.go
//go:build !enterprise

package newfeature

// OSS version uses stub implementation from newfeature.go
```

```go
// internal/ee/newfeature/build_enterprise.go  
//go:build enterprise

package newfeature

type EnterpriseNewFeature struct {
    // Real implementation fields
}

func New() NewFeature {
    return &EnterpriseNewFeature{}
}

func (e *EnterpriseNewFeature) DoSomething(ctx context.Context, param string) (*Result, error) {
    // Real enterprise implementation
}
```

3. **Add Configuration Support**:
```go
// Add to internal/config/ee.go
type EnterpriseConfig struct {
    // ... existing fields
    NewFeature NewFeatureConfig `mapstructure:"newfeature"`
}

type NewFeatureConfig struct {
    Enabled bool `mapstructure:"enabled"`
    // ... feature-specific config
}
```

4. **Update Stub Config**:
```go
// Add to internal/config/ee_stub.go  
type NewFeatureConfig struct {
    Enabled bool `mapstructure:"enabled"`
}
```

## Middleware System

### Enterprise Feature Gating

```go
// internal/middleware/enterprise.go
func EnterpriseFeature(feature string, licenseService *services.LicenseService, logger *logrus.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        cfg := c.MustGet("config").(*config.Config)
        
        // Allow all features in development mode
        if cfg.IsDevelopment() {
            c.Header("X-Feature-Mode", "development")
            c.Next()
            return
        }

        // Check if feature is available in current license
        available, err := licenseService.CheckFeatureEntitlement(c.Request.Context(), feature)
        if err != nil {
            logger.WithError(err).WithField("feature", feature).Error("Failed to check feature entitlement")
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate feature access"})
            c.Abort()
            return
        }

        if !available {
            currentTier := cfg.GetLicenseTier()
            requiredTier := getRequiredTierForFeature(feature)
            
            enterpriseError := errors.NewFeatureNotAvailableError(feature, currentTier, requiredTier)
            c.JSON(enterpriseError.HTTPStatus(), gin.H{"error": enterpriseError})
            c.Abort()
            return
        }

        // Feature is available, add to context and continue
        c.Set("enterprise_feature", feature)
        c.Next()
    }
}
```

### Usage Limit Middleware

```go
func CheckUsageLimit(limitType string, licenseService *services.LicenseService, logger *logrus.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        withinLimit, remaining, err := licenseService.CheckUsageLimit(c.Request.Context(), limitType)
        if err != nil {
            logger.WithError(err).Error("Failed to check usage limit")
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate usage limits"})
            c.Abort()
            return
        }

        if !withinLimit {
            resetDate := time.Now().AddDate(0, 1, -time.Now().Day())
            enterpriseError := errors.NewUsageLimitExceededError(limitType, cfg.GetLicenseTier(), remaining, &resetDate)
            c.JSON(enterpriseError.HTTPStatus(), gin.H{"error": enterpriseError})
            c.Abort()
            return
        }

        c.Header("X-Usage-Remaining", fmt.Sprintf("%d", remaining))
        c.Next()
    }
}
```

### Usage in Routes

```go
// Apply enterprise feature middleware
router.Group("/api/v1/advanced").Use(
    middleware.EnterpriseFeature("advanced_rbac", licenseService, logger),
).POST("/roles", handlers.CreateAdvancedRole)

// Apply usage limit middleware
router.Group("/api/v1/ai").Use(
    middleware.CheckUsageLimit("requests", licenseService, logger),
).POST("/chat", handlers.ChatCompletion)

// Require enterprise license
router.Group("/api/v1/enterprise").Use(
    middleware.RequireEnterpriseLicense(licenseService, logger),
).GET("/compliance/audit", handlers.GenerateAuditReport)
```

## License Service

### License Validation Architecture

```go
// internal/services/license/
type LicenseService struct {
    config     *config.Config
    logger     *logrus.Logger
    redis      *redis.Client
    httpClient *http.Client
    publicKey  *rsa.PublicKey
}

type LicenseInfo struct {
    Key           string    `json:"key"`
    Type          string    `json:"type"`
    ValidUntil    time.Time `json:"valid_until"`
    MaxRequests   int64     `json:"max_requests"`
    MaxUsers      int       `json:"max_users"`
    MaxProjects   int       `json:"max_projects"`
    Features      []string  `json:"features"`
    IsValid       bool      `json:"is_valid"`
    LastValidated time.Time `json:"last_validated"`
}
```

### License Validation Flow

```go
func (ls *LicenseService) ValidateLicense(ctx context.Context) (*LicenseStatus, error) {
    // 1. Check cache first
    if cachedLicense, err := ls.getCachedLicense(ctx); err == nil {
        if time.Since(cachedLicense.LastValidated) < licenseCacheTTL {
            return ls.buildLicenseStatus(cachedLicense), nil
        }
    }

    // 2. Perform fresh validation
    licenseInfo, err := ls.performLicenseValidation(ctx)
    if err != nil {
        // 3. Fallback to cached license if within grace period
        if cachedLicense, cacheErr := ls.getCachedLicense(ctx); cacheErr == nil {
            if time.Since(cachedLicense.LastValidated) < offlineGracePeriod {
                return ls.buildLicenseStatus(cachedLicense), nil
            }
        }
        // 4. Return free tier if all validation fails
        return ls.getFreetierStatus(ctx), nil
    }

    // 5. Cache validated license
    ls.cacheLicense(ctx, licenseInfo)
    return ls.buildLicenseStatus(licenseInfo), nil
}
```

### Online vs Offline Validation

```go
func (ls *LicenseService) performLicenseValidation(ctx context.Context) (*LicenseInfo, error) {
    license := &ls.config.Enterprise.License

    // No license key = free tier
    if license.Key == "" {
        return ls.createFreeTierLicense(), nil
    }

    // Offline mode or development = local validation
    if license.OfflineMode || ls.config.IsDevelopment() {
        return ls.validateLicenseLocally(license)
    }

    // Online validation with license server
    return ls.validateLicenseOnline(ctx, license)
}
```

### JWT License Format

```go
// License JWT claims structure
type LicenseClaims struct {
    Type          string   `json:"type"`
    ValidUntil    int64    `json:"valid_until"`
    MaxRequests   int64    `json:"max_requests"`
    MaxUsers      int      `json:"max_users"`
    MaxProjects   int      `json:"max_projects"`
    Features      []string `json:"features"`
    Organization  string   `json:"organization,omitempty"`
    ContactEmail  string   `json:"contact_email,omitempty"`
    jwt.RegisteredClaims
}
```

## Error Handling

### Professional Error Response System

```go
// internal/errors/enterprise.go
type EnterpriseError struct {
    Code            EnterpriseErrorCode `json:"code"`
    Message         string              `json:"message"`
    Feature         string              `json:"feature,omitempty"`
    CurrentTier     string              `json:"current_tier,omitempty"`
    RequiredTier    string              `json:"required_tier,omitempty"`
    Actions         []ActionSuggestion  `json:"actions"`
    Support         SupportInfo         `json:"support"`
    Metadata        map[string]string   `json:"metadata,omitempty"`
}

type ActionSuggestion struct {
    Type        string `json:"type"`        // "upgrade", "contact_sales", "trial"
    Label       string `json:"label"`
    URL         string `json:"url"`
    UTMSource   string `json:"utm_source"`  // Conversion tracking
    UTMCampaign string `json:"utm_campaign"`
    UTMContent  string `json:"utm_content"`
    Primary     bool   `json:"primary"`
}
```

### Error Response Examples

```json
{
  "error": {
    "code": "FEATURE_NOT_AVAILABLE",
    "message": "Single Sign-On Integration requires Business tier or higher. You're currently on Free tier.",
    "feature": "sso_integration",
    "current_tier": "free",
    "required_tier": "business",
    "actions": [
      {
        "type": "upgrade",
        "label": "Upgrade to Business",
        "url": "https://brokle.com/pricing?tier=business&utm_source=api&utm_campaign=feature_upgrade&utm_content=sso_integration",
        "primary": true
      },
      {
        "type": "trial",
        "label": "Start Free Trial",
        "url": "https://brokle.com/trial?utm_source=api&utm_campaign=feature_trial&utm_content=sso_integration",
        "primary": false
      }
    ],
    "support": {
      "email": "support@brokle.com",
      "chat_url": "https://brokle.com/chat",
      "docs_url": "https://docs.brokle.com"
    }
  }
}
```

### HTTP Status Code Mapping

```go
func (ee *EnterpriseError) HTTPStatus() int {
    switch ee.Code {
    case ErrorCodeFeatureNotAvailable, ErrorCodeLicenseRequired:
        return http.StatusPaymentRequired // 402
    case ErrorCodeUsageLimitExceeded:
        return http.StatusTooManyRequests // 429
    case ErrorCodeLicenseExpired, ErrorCodeLicenseInvalid:
        return http.StatusUnauthorized // 401
    default:
        return http.StatusForbidden // 403
    }
}
```

## Testing Strategy

### Interface Compliance Testing

```go
// internal/test/interface_compliance_test.go
func TestEnterpriseInterfaceCompliance(t *testing.T) {
    ctx := context.Background()

    t.Run("Compliance interface compliance", func(t *testing.T) {
        service := compliance.New()
        require.NotNil(t, service)
        
        // Verify interface compliance at compile time
        var _ compliance.Compliance = service
        
        // Test all interface methods are callable
        assert.NotPanics(t, func() {
            err := service.ValidateCompliance(ctx, map[string]interface{}{"test": "data"})
            assert.NoError(t, err) // Stub should not error
        })
    })
}
```

### Build Tag Testing

```bash
# Test OSS build
go test -v ./internal/test/

# Test Enterprise build  
go test -tags="enterprise" -v ./internal/test/

# Test both builds work
make test-builds
```

### Stub Behavior Testing

```go
func TestStubBehaviorConsistency(t *testing.T) {
    ctx := context.Background()
    
    t.Run("Stub services provide safe defaults", func(t *testing.T) {
        // Compliance should be permissive in stub mode
        compliance := compliance.New()
        err := compliance.ValidateCompliance(ctx, map[string]interface{}{"test": "data"})
        assert.NoError(t, err, "Stub compliance should not block operations")
        
        // RBAC should allow basic permissions in stub mode for development
        rbac := rbac.New()
        hasPermission, err := rbac.CheckPermission(ctx, "user", "project", "read")
        assert.NoError(t, err)
        assert.True(t, hasPermission, "Stub RBAC allows basic permissions for development")
    })
}
```

### License Service Testing

```go
func TestLicenseService(t *testing.T) {
    // Test license validation logic
    // Test tier limits
    // Test feature entitlements  
    // Test usage tracking
    // Test cache behavior
}
```

## Development Workflow

### Setting Up Development Environment

1. **Clone and Setup**:
```bash
git clone https://github.com/brokle-ai/brokle-platform.git
cd brokle-platform/brokle
make setup
```

2. **Configure Enterprise License** (for testing):
```bash
# Set development license
export BROKLE_ENTERPRISE_LICENSE_TYPE="enterprise"
export BROKLE_ENTERPRISE_LICENSE_KEY="development-key"
export BROKLE_ENVIRONMENT="development"
```

3. **Build Both Versions**:
```bash
make build-all
```

4. **Run Tests**:
```bash
make test
make test-enterprise  # If available
```

### Development Best Practices

1. **Always Test Both Builds**:
   - Every feature should work in OSS (with appropriate stubs)
   - Every feature should be fully functional in Enterprise

2. **Interface-First Development**:
   - Define interfaces before implementation
   - Ensure stubs implement all interface methods
   - Use meaningful error messages in stubs

3. **Professional Error Messages**:
   - Include upgrade paths and pricing information
   - Add UTM tracking for conversion optimization
   - Provide clear support contact information

4. **License Validation**:
   - Always check feature entitlements before execution
   - Handle license validation failures gracefully
   - Provide clear upgrade guidance

### Adding Enterprise Routes

```go
// Define enterprise routes
func SetupEnterpriseRoutes(router *gin.Engine, services *Services) {
    api := router.Group("/api/v1")
    
    // SSO routes (requires sso_integration feature)
    sso := api.Group("/sso").Use(
        middleware.EnterpriseFeature("sso_integration", services.License, services.Logger),
    )
    sso.GET("/login", handlers.SSOLogin)
    sso.POST("/callback", handlers.SSOCallback)
    
    // RBAC routes (requires advanced_rbac feature)  
    rbac := api.Group("/rbac").Use(
        middleware.EnterpriseFeature("advanced_rbac", services.License, services.Logger),
    )
    rbac.POST("/roles", handlers.CreateCustomRole)
    rbac.GET("/roles", handlers.ListCustomRoles)
    
    // Compliance routes (requires enterprise license)
    compliance := api.Group("/compliance").Use(
        middleware.RequireEnterpriseLicense(services.License, services.Logger),
    )
    compliance.GET("/audit", handlers.GenerateAuditReport)
    compliance.POST("/anonymize", handlers.AnonymizePII)
}
```

## Debugging

### Common Issues and Solutions

1. **Build Tag Issues**:
```bash
# Error: undefined function/type
# Solution: Check build tags are correct
go build -tags="enterprise" -v ./cmd/server

# Check what files are being compiled
go list -tags="enterprise" -f '{{.GoFiles}}' ./internal/config
```

2. **License Validation Failures**:
```bash
# Enable debug logging
export BROKLE_LOGGING_LEVEL="debug"

# Check license service logs
grep "license" /var/log/brokle.log

# Test license validation endpoint
curl -v http://localhost:8080/api/v1/enterprise/license
```

3. **Feature Gating Issues**:
```bash
# Check feature entitlements
curl -H "Authorization: Bearer $TOKEN" \
     http://localhost:8080/api/v1/enterprise/features/advanced_rbac
```

### Debug Endpoints (Development Only)

```go
// Only available in development mode
if cfg.IsDevelopment() {
    debug := router.Group("/debug")
    debug.GET("/license", handlers.DebugLicense)
    debug.GET("/features", handlers.DebugFeatures)
    debug.GET("/build-info", handlers.DebugBuildInfo)
}
```

### Logging Best Practices

```go
// Structured logging for enterprise features
logger.WithFields(logrus.Fields{
    "feature":      "advanced_rbac",
    "user_id":      userID,
    "license_tier": cfg.GetLicenseTier(),
    "action":       "create_role",
}).Info("Enterprise feature accessed")

// Log license validation events
logger.WithFields(logrus.Fields{
    "license_type": license.Type,
    "valid_until":  license.ValidUntil,
    "features":     license.Features,
    "validation_result": "success",
}).Info("License validated")
```

---

This developer guide covers the essential aspects of working with Brokle's enterprise architecture. For specific feature implementation details, refer to the individual feature documentation in this directory.