# Industrial Go Error Handling Guide

This document outlines the standardized error handling patterns implemented across the Brokle platform following Go industrial best practices. Use this as a reference for consistent error handling throughout the codebase.

## 🎯 Core Principles

### 1. Clean Architecture Error Flow
```
Repository → Service → Handler
Domain Errors → AppErrors → HTTP Responses
```

### 2. Separation of Concerns
- **Core Services**: Pure business logic with structured errors
- **Decorators**: Cross-cutting concerns (audit, logging) handled separately
- **No Logging in Core**: Business logic services have zero logging dependencies

### 3. Structured Error Types
- Use domain-specific errors at repository level
- Transform to AppErrors at service level
- Handle HTTP responses at handler level

## 📋 Step-by-Step Implementation Checklist

### Phase 1: Domain Error Cleanup
- [ ] **Remove gorm leakage**: Replace raw `gorm.ErrRecordNotFound` with domain errors
- [ ] **Simplify domain errors**: Keep only essential errors (NotFound, AlreadyExists, Inactive, etc.)
- [ ] **Consistent error wrapping**: Use `fmt.Errorf("operation context: %w", domainError)`

### Phase 2: Service Layer Transformation
- [ ] **Remove all fmt.Errorf calls**: Convert to `appErrors.NewXxxError()` constructors
- [ ] **Remove all errors.New calls**: Convert to appropriate AppError types
- [ ] **Remove all fmt.Printf/log statements**: Core services should have zero logging
- [ ] **Fix variable shadowing**: Avoid package name conflicts (e.g., `user` variable vs `user` package)
- [ ] **Domain error mapping**: Properly map domain errors to AppErrors

### Phase 3: Dependency Cleanup
- [ ] **Remove audit dependencies**: Extract audit logging from core services
- [ ] **Remove logger dependencies**: Core services focus on business logic only
- [ ] **Update constructors**: Remove unnecessary parameters from service constructors

### Phase 4: Decorator Pattern Implementation
- [ ] **Create audit decorators**: Implement decorator pattern for cross-cutting concerns
- [ ] **Update DI container**: Wrap core services with decorators in providers
- [ ] **Clean imports**: Remove unused imports after cleanup

## 🔧 Implementation Patterns

### Repository Layer Pattern
```go
// BEFORE (gorm leakage)
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
    var u user.User
    err := r.db.Where("email = ?", email).First(&u).Error
    if err != nil {
        return nil, err // ❌ Raw gorm error
    }
    return &u, nil
}

// AFTER (domain errors)
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
    var u user.User
    err := r.db.Where("email = ?", email).First(&u).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("get user by email %s: %w", email, user.ErrNotFound)
        }
        return nil, fmt.Errorf("database error getting user by email %s: %w", email, err)
    }
    return &u, nil
}
```

### Domain Error Definition
```go
// internal/core/domain/user/errors.go
package user

import "errors"

var (
    ErrNotFound      = errors.New("not found")
    ErrAlreadyExists = errors.New("already exists")
    ErrInactive      = errors.New("inactive")
    ErrInvalidEmail  = errors.New("invalid email format")
    ErrWeakPassword  = errors.New("password too weak")
)
```

### Service Layer Pattern
```go
// BEFORE (inconsistent error handling)
func (s *authService) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
    user, err := s.userRepo.GetByEmailWithPassword(ctx, req.Email)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err) // ❌ Generic error
    }
    // ... audit logging mixed in ❌
    s.auditRepo.Create(ctx, auditLog)
    return response, nil
}

// AFTER (structured AppErrors)
func (s *authService) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
    foundUser, err := s.userRepo.GetByEmailWithPassword(ctx, req.Email)
    if err != nil {
        if errors.Is(err, user.ErrNotFound) {
            return nil, appErrors.NewUnauthorizedError("Invalid email or password")
        }
        return nil, appErrors.NewInternalError("Authentication service unavailable", err)
    }
    
    if !foundUser.IsActive {
        return nil, appErrors.NewForbiddenError("Account is inactive")
    }
    // ... pure business logic only ✅
    return response, nil
}
```

### Handler Layer Pattern
```go
// BEFORE (complex error switching)
func (h *AuthHandler) Login(c *gin.Context) {
    resp, err := h.authService.Login(ctx, req)
    if err != nil {
        // Complex error type switching ❌
        switch {
        case errors.Is(err, user.ErrNotFound):
            c.JSON(401, gin.H{"error": "unauthorized"})
        case errors.Is(err, user.ErrInactive):
            c.JSON(403, gin.H{"error": "forbidden"})
        default:
            c.JSON(500, gin.H{"error": "internal error"})
        }
        return
    }
    c.JSON(200, resp)
}

// AFTER (clean response handling)
func (h *AuthHandler) Login(c *gin.Context) {
    resp, err := h.authService.Login(ctx, req)
    if err != nil {
        response.Error(c, err) // ✅ Centralized error handling
        return
    }
    response.Success(c, resp)
}
```

### Decorator Pattern for Cross-Cutting Concerns
```go
// audit_decorator.go
type auditDecorator struct {
    core      auth.AuthService
    auditRepo auth.AuditLogRepository
    logger    *logrus.Logger
}

func NewAuditDecorator(core auth.AuthService, auditRepo auth.AuditLogRepository, logger *logrus.Logger) auth.AuthService {
    return &auditDecorator{
        core:      core,
        auditRepo: auditRepo,
        logger:    logger,
    }
}

func (d *auditDecorator) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
    resp, err := d.core.Login(ctx, req) // ✅ Delegate to core
    
    // Handle audit logging based on result
    auditReason := "auth.login.success"
    if err != nil {
        auditReason = d.mapErrorToAuditReason(err)
    }
    
    auditLog := auth.NewAuditLog(nil, nil, auditReason, "user", req.Email, "", "", "")
    d.auditRepo.Create(ctx, auditLog)
    
    return resp, err
}
```

## 📊 AppError Constructor Mapping Guide

| Scenario | AppError Constructor | HTTP Status |
|----------|---------------------|-------------|
| Domain NotFound | `appErrors.NewNotFoundError("Resource not found")` | 404 |
| Invalid credentials | `appErrors.NewUnauthorizedError("Invalid credentials")` | 401 |
| Permission denied | `appErrors.NewForbiddenError("Access denied")` | 403 |
| Validation failed | `appErrors.NewValidationError("field", "validation message")` | 400 |
| Resource exists | `appErrors.NewConflictError("Resource already exists")` | 409 |
| External service down | `appErrors.NewInternalError("Service unavailable", err)` | 500 |
| Feature not ready | `appErrors.NewNotImplementedError("Feature not implemented")` | 501 |

## 🚫 Anti-Patterns to Avoid

### ❌ Don't Do This
```go
// Gorm leakage
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, errors.New("user not found")
}

// Generic errors
return fmt.Errorf("something went wrong: %w", err)

// Logging in core services
log.Error("Failed to create user", err)
fmt.Printf("Debug: %v\n", data)

// Variable shadowing
user, err := repo.GetUser(...)
if errors.Is(err, user.ErrNotFound) { // ❌ 'user' is variable, not package
}

// Mixed concerns
func (s *service) CreateUser(...) {
    // business logic
    // audit logging  ❌ Mixed concerns
    // more business logic
}
```

### ✅ Do This Instead
```go
// Proper domain error wrapping
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, fmt.Errorf("get user by ID %s: %w", id, user.ErrNotFound)
}

// Specific AppError constructors
return appErrors.NewNotFoundError("User not found")

// No logging in core services - pure business logic only
func (s *service) CreateUser(...) {
    // pure business logic only ✅
}

// Avoid variable shadowing
foundUser, err := repo.GetUser(...)
if errors.Is(err, user.ErrNotFound) { // ✅ Clear reference to package
}

// Decorator pattern for cross-cutting concerns
core := service.NewUserService(...)
decorated := service.NewAuditDecorator(core, auditRepo, logger)
```

## 🔄 Migration Workflow

### Step 1: Analysis
1. Identify all `fmt.Errorf` and `errors.New` calls in service
2. Find all `fmt.Printf` and logging statements  
3. Locate audit logging mixed with business logic
4. Check for variable shadowing issues

### Step 2: Repository Layer
1. Replace gorm error leakage with domain errors
2. Add proper error wrapping with context
3. Test repository layer independently

### Step 3: Service Layer  
1. Convert all error returns to AppError constructors
2. Remove all logging statements
3. Fix variable shadowing (use `foundUser`, `retrievedUser`, etc.)
4. Remove audit logging from core service

### Step 4: Decorator Implementation
1. Create audit decorator for the service
2. Move audit logging to decorator
3. Update DI container to wrap service with decorator

### Step 5: Verification
1. `go build` - ensure compilation
2. `go vet` - check for issues
3. Test error scenarios return appropriate HTTP status codes
4. Verify audit logging still works via decorator

## 📁 File Structure Example

```
internal/core/services/auth/
├── auth_service.go          # Core business logic (no logging)
├── audit_decorator.go       # Audit logging decorator  
├── session_service.go       # Session management
└── jwt_service.go           # JWT operations

internal/core/domain/auth/
├── errors.go                # Domain-specific errors
├── entities.go              # Domain entities
└── interfaces.go            # Service interfaces

internal/app/
└── providers.go             # DI container with decorator wiring
```

## 🎯 Success Criteria

When implementing this pattern, you should achieve:

- [ ] **Zero logging** in core business logic services
- [ ] **Consistent error types** using AppError constructors
- [ ] **Clean separation** between business logic and cross-cutting concerns
- [ ] **Proper error flow** from repository → service → handler
- [ ] **No variable shadowing** issues
- [ ] **Compilation success** with `go build` and `go vet`
- [ ] **Maintainable code** following single responsibility principle

## 🔧 Tools & Commands

```bash
# Find all error handling issues
grep -r "fmt\.Errorf\|errors\.New" internal/core/services/
grep -r "fmt\.Printf\|log\." internal/core/services/
grep -r "gorm\.ErrRecordNotFound" internal/

# Verify cleanup
go build ./...
go vet ./...
go test ./...
```

## 📚 Additional Resources

- [Go Error Handling Best Practices](https://blog.golang.org/error-handling-and-go)
- [Clean Architecture in Go](https://github.com/bxcodec/go-clean-arch)
- [Effective Error Handling](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)

---

**Next Steps**: Apply this pattern systematically across all services in the codebase. Start with critical services like user management, then proceed to other business domains.