# Error Handling Quick Reference

Quick reference for the industrial Go error handling patterns used in the Brokle platform.

## 🚀 Quick Start Checklist

### Repository Layer
```go
// ✅ Required imports
import (
    authDomain "brokle/internal/core/domain/auth"
)

// ✅ GORM error handling
if err == gorm.ErrRecordNotFound {
    return nil, fmt.Errorf("get user by ID %s: %w", id, authDomain.ErrNotFound)
}
```

### Service Layer  
```go
// ✅ AppError constructors
if errors.Is(err, userDomain.ErrNotFound) {
    return nil, appErrors.NewNotFoundError("User not found")
}
```

### Handler Layer
```go
// ✅ Structured responses
resp, err := h.service.Method(c, req)
if err != nil {
    response.Error(c, err)
    return
}
response.Success(c, resp)
```

## 📋 Domain Aliases Reference

| Domain | Alias | Common Usage |
|--------|-------|-------------|
| Authentication | `authDomain` | `authDomain.User`, `authDomain.ErrNotFound` |
| Organization | `orgDomain` | `orgDomain.Organization`, `orgDomain.ErrNotFound` |
| User | `userDomain` | `userDomain.User`, `userDomain.ErrNotFound` |
| Billing | `billingDomain` | `billingDomain.Subscription` |
| Analytics | `analyticsDomain` | `analyticsDomain.Metric` |

## 🔧 Common Patterns

### Repository Constructor
```go
func NewUserRepository(db *gorm.DB) authDomain.UserRepository {
    return &userRepository{db: db}
}
```

### GORM Error Conversion
```go
if err == gorm.ErrRecordNotFound {
    return nil, fmt.Errorf("get user by email %s: %w", email, userDomain.ErrNotFound)
}
return nil, fmt.Errorf("database query failed for email %s: %w", email, err)
```

### Service Error Handling
```go
if errors.Is(err, userDomain.ErrNotFound) {
    return nil, appErrors.NewNotFoundError("User not found")
}
return nil, appErrors.NewInternalError("Failed to retrieve user")
```

## ❌ Common Mistakes

```go
// ❌ Don't use errors.New in repositories
return nil, errors.New("user not found")

// ❌ Don't use errors.Is with GORM errors  
if errors.Is(err, gorm.ErrRecordNotFound)

// ❌ Don't use direct domain imports
import "brokle/internal/core/domain/auth"

// ❌ Don't use fmt.Errorf in services
return nil, fmt.Errorf("user not found")
```

## 🧪 Testing Patterns

```go
// Repository test
assert.True(t, errors.Is(err, userDomain.ErrNotFound))

// Service test
assert.True(t, appErrors.IsNotFoundError(err))

// Mock setup
mockRepo.On("GetByID", mock.Anything, id).Return(nil, userDomain.ErrNotFound)
```

## 🚨 Error Types

### Domain Errors (Repository → Service)
```go
userDomain.ErrNotFound
authDomain.ErrNotFound
orgDomain.ErrNotFound
```

### AppErrors (Service → Handler)
```go
appErrors.NewNotFoundError("User not found")
appErrors.NewValidationError("Invalid email", "email")
appErrors.NewConflictError("User already exists")
appErrors.NewInternalError("Database connection failed")
```

### HTTP Status Mapping (Automatic)
```go
response.Error(c, err) // Maps AppErrors to HTTP status codes
```

## 🔍 Debugging Tips

1. **Check error chain**: `fmt.Printf("%+v\n", err)`
2. **Verify domain errors**: Ensure constants are defined
3. **Test error flow**: Write tests for error propagation
4. **Use error context**: Include relevant IDs and parameters

## 📚 Full Documentation

- [ERROR_HANDLING_GUIDE.md](./ERROR_HANDLING_GUIDE.md) - Complete implementation guide
- [DOMAIN_ALIAS_PATTERNS.md](./DOMAIN_ALIAS_PATTERNS.md) - Domain alias patterns and examples