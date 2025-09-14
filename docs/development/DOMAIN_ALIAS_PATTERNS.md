# Professional Domain Alias Patterns

This guide documents the professional domain alias patterns implemented across the Brokle platform for clean imports, avoiding naming conflicts, and improving code maintainability.

## Overview

Domain alias patterns provide a standardized way to import domain packages while avoiding naming conflicts and improving code readability. This is particularly important in a large codebase with multiple domains.

## Standard Domain Aliases

### Core Domain Aliases

| Domain Package | Alias | Usage Context |
|---------------|--------|---------------|
| `brokle/internal/core/domain/auth` | `authDomain` | Authentication, authorization, user management |
| `brokle/internal/core/domain/organization` | `orgDomain` | Organizations, projects, environments, teams |
| `brokle/internal/core/domain/user` | `userDomain` | User profiles, preferences, onboarding |
| `brokle/internal/core/domain/billing` | `billingDomain` | Subscriptions, payments, usage tracking |
| `brokle/internal/core/domain/analytics` | `analyticsDomain` | Metrics, reporting, insights |

### Import Pattern Standard

```go
// ✅ Correct Professional Pattern
import (
    "context"
    "fmt"
    "time"

    "gorm.io/gorm"

    authDomain "brokle/internal/core/domain/auth"
    orgDomain "brokle/internal/core/domain/organization"
    userDomain "brokle/internal/core/domain/user"
    "brokle/pkg/ulid"
)
```

## Implementation Examples

### Authentication Domain Usage

```go
package auth

import (
    authDomain "brokle/internal/core/domain/auth"
)

// Repository interface implementation
func NewUserRepository(db *gorm.DB) authDomain.UserRepository {
    return &userRepository{db: db}
}

// Entity usage
func (r *userRepository) Create(ctx context.Context, user *authDomain.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

// Error handling with domain errors
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*authDomain.User, error) {
    var user authDomain.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, fmt.Errorf("get user by email %s: %w", email, authDomain.ErrNotFound)
        }
        return nil, fmt.Errorf("database query failed: %w", err)
    }
    return &user, nil
}

// Service interfaces and enums
func (s *authService) ValidateRole(role authDomain.Role) bool {
    return role != authDomain.RoleUnknown
}

// Status and type checking
func (s *sessionService) IsSessionActive(session *authDomain.UserSession) bool {
    return session.Status == authDomain.SessionStatusActive
}
```

### Organization Domain Usage

```go
package organization

import (
    orgDomain "brokle/internal/core/domain/organization"
    authDomain "brokle/internal/core/domain/auth"
    userDomain "brokle/internal/core/domain/user"
)

// Multi-domain repository
func NewMemberRepository(db *gorm.DB) orgDomain.MemberRepository {
    return &memberRepository{db: db}
}

// Cross-domain relationships
func (r *memberRepository) GetMemberWithUser(ctx context.Context, orgID, userID ulid.ULID) (*orgDomain.Member, *userDomain.User, error) {
    var member orgDomain.Member
    var user userDomain.User
    
    err := r.db.WithContext(ctx).
        Select("members.*, users.*").
        Joins("JOIN users ON members.user_id = users.id").
        Where("members.organization_id = ? AND members.user_id = ?", orgID, userID).
        First(&member).Error
        
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil, fmt.Errorf("get member %s in org %s: %w", userID, orgID, orgDomain.ErrMemberNotFound)
        }
        return nil, nil, fmt.Errorf("database query failed: %w", err)
    }
    
    return &member, &user, nil
}

// Enum and status usage
func (s *invitationService) SendInvitation(ctx context.Context, invitation *orgDomain.Invitation) error {
    invitation.Status = orgDomain.InvitationStatusPending
    invitation.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
    
    return s.invitationRepo.Create(ctx, invitation)
}
```

### Multi-Domain Service Example

```go
package services

import (
    authDomain "brokle/internal/core/domain/auth"
    orgDomain "brokle/internal/core/domain/organization"
    userDomain "brokle/internal/core/domain/user"
    billingDomain "brokle/internal/core/domain/billing"
)

type OrganizationService struct {
    orgRepo     orgDomain.OrganizationRepository
    userRepo    userDomain.UserRepository
    memberRepo  orgDomain.MemberRepository
    billingRepo billingDomain.SubscriptionRepository
    authService authDomain.AuthService
}

func (s *OrganizationService) CreateOrganization(ctx context.Context, req *CreateOrgRequest) (*CreateOrgResponse, error) {
    // Create organization
    org := &orgDomain.Organization{
        ID:    ulid.New(),
        Name:  req.Name,
        Slug:  req.Slug,
        Plan:  billingDomain.PlanFree,
    }
    
    // Add owner as first member
    member := &orgDomain.Member{
        ID:             ulid.New(),
        OrganizationID: org.ID,
        UserID:         req.OwnerID,
        Role:           authDomain.RoleOwner,
        Status:         orgDomain.MemberStatusActive,
    }
    
    // Create billing subscription
    subscription := &billingDomain.Subscription{
        ID:             ulid.New(),
        OrganizationID: org.ID,
        Plan:           billingDomain.PlanFree,
        Status:         billingDomain.SubscriptionStatusActive,
    }
    
    // Transaction across multiple domains
    return s.createOrgWithTransaction(ctx, org, member, subscription)
}
```

## Benefits of Domain Aliases

### 1. Conflict Resolution

```go
// ❌ Without aliases - naming conflicts
import (
    "brokle/internal/core/domain/auth"
    "brokle/internal/core/domain/organization"
)

// Ambiguous - which User?
func process(user auth.User, org organization.User) // Conflict!

// ✅ With aliases - clear distinction
import (
    authDomain "brokle/internal/core/domain/auth"
    orgDomain "brokle/internal/core/domain/organization"
)

func process(user authDomain.User, member orgDomain.Member) // Clear!
```

### 2. Code Readability

```go
// ✅ Self-documenting code
user := &authDomain.User{
    Email: req.Email,
    Role:  authDomain.RoleUser,
}

org := &orgDomain.Organization{
    Name:   req.Name,
    Status: orgDomain.OrganizationStatusActive,
}

invitation := &orgDomain.Invitation{
    Status: orgDomain.InvitationStatusPending,
}
```

### 3. Refactoring Safety

```go
// Domain aliases make refactoring safer
// If domain package structure changes, only import needs updating

// Before refactor
import authDomain "brokle/internal/core/domain/auth"

// After refactor  
import authDomain "brokle/internal/domains/authentication"

// All usage remains the same: authDomain.User, authDomain.ErrNotFound, etc.
```

## Anti-Patterns to Avoid

### ❌ Direct Domain Imports

```go
// Don't import domains directly
import (
    "brokle/internal/core/domain/auth"
    "brokle/internal/core/domain/organization"
)

// Creates ambiguity and potential conflicts
func handle(user auth.User) error {
    // Which domain does this belong to?
    org := organization.NewOrg()  // Unclear
}
```

### ❌ Inconsistent Alias Names

```go
// Don't use inconsistent aliases
import (
    a "brokle/internal/core/domain/auth"        // ❌ Too short
    authPkg "brokle/internal/core/domain/auth"  // ❌ Inconsistent
    auth_domain "brokle/internal/core/domain/auth" // ❌ Underscore
)
```

### ❌ Overly Long Aliases

```go
// Don't make aliases too verbose
import (
    authenticationDomain "brokle/internal/core/domain/auth" // ❌ Too long
)
```

## Repository-Specific Patterns

### Complete Repository File Template

```go
package auth

import (
    "context"
    "fmt"
    "time"

    "gorm.io/gorm"

    authDomain "brokle/internal/core/domain/auth"
    "brokle/pkg/ulid"
)

// Repository implementation with interface compliance
type userRepository struct {
    db *gorm.DB
}

// Constructor with domain interface return
func NewUserRepository(db *gorm.DB) authDomain.UserRepository {
    return &userRepository{db: db}
}

// CRUD methods with proper domain alias usage
func (r *userRepository) Create(ctx context.Context, user *authDomain.User) error {
    if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
        return fmt.Errorf("create user: %w", err)
    }
    return nil
}

func (r *userRepository) GetByID(ctx context.Context, id ulid.ULID) (*authDomain.User, error) {
    var user authDomain.User
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, fmt.Errorf("get user by ID %s: %w", id, authDomain.ErrNotFound)
        }
        return nil, fmt.Errorf("database query failed: %w", err)
    }
    return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *authDomain.User) error {
    if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
        return fmt.Errorf("update user %s: %w", user.ID, err)
    }
    return nil
}

func (r *userRepository) Delete(ctx context.Context, id ulid.ULID) error {
    if err := r.db.WithContext(ctx).Delete(&authDomain.User{}, "id = ?", id).Error; err != nil {
        return fmt.Errorf("delete user %s: %w", id, err)
    }
    return nil
}

// Complex queries with proper domain usage
func (r *userRepository) GetActiveUsersInOrganization(ctx context.Context, orgID ulid.ULID) ([]*authDomain.User, error) {
    var users []*authDomain.User
    err := r.db.WithContext(ctx).
        Table("users").
        Joins("JOIN organization_members ON users.id = organization_members.user_id").
        Where("organization_members.organization_id = ? AND users.status = ?", 
            orgID, authDomain.UserStatusActive).
        Find(&users).Error
    
    if err != nil {
        return nil, fmt.Errorf("get active users in organization %s: %w", orgID, err)
    }
    
    return users, nil
}
```

## Validation and Code Quality

### Automated Checks

You can add these validation rules to your linter configuration:

```yaml
# .golangci.yml
linters-settings:
  goimports:
    local-prefixes: brokle/internal/core/domain
  
  revive:
    rules:
      - name: import-alias-naming
        arguments:
          - "^[a-z][a-zA-Z0-9]*Domain$"
```

### Code Review Checklist

- [ ] All domain imports use professional aliases
- [ ] Alias names follow the `{domain}Domain` pattern
- [ ] No direct domain imports without aliases
- [ ] Consistent alias usage throughout the file
- [ ] Domain entities, errors, and enums use proper aliases

## Migration Guide

### Converting Existing Code

1. **Identify Domain Imports**:
```bash
grep -r "brokle/internal/core/domain" --include="*.go" ./internal/
```

2. **Add Professional Aliases**:
```go
// Before
import "brokle/internal/core/domain/auth"

// After  
import authDomain "brokle/internal/core/domain/auth"
```

3. **Update Usage with sed**:
```bash
sed -i 's/auth\./authDomain\./g' file.go
sed -i 's/organization\./orgDomain\./g' file.go
```

4. **Verify Compilation**:
```bash
go build ./internal/infrastructure/repository/...
```

## Best Practices Summary

1. **Consistency**: Always use the same alias for the same domain across all files
2. **Clarity**: Aliases should be descriptive but not overly verbose
3. **Standards**: Follow the `{domain}Domain` naming pattern
4. **Documentation**: Document any custom aliases in team conventions
5. **Automation**: Use tooling to enforce alias patterns
6. **Migration**: Convert existing code systematically with verification

This domain alias system ensures clean, maintainable, and collision-free imports across the entire Brokle platform while providing excellent developer experience and AI assistant comprehension.