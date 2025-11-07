---
name: brokle-domain-architecture
description: Use this skill when working with Brokle's domain-driven architecture, including creating new domains, modifying domain entities, designing cross-domain interactions, refactoring domain boundaries, or implementing complex domain logic. This is a specialized architectural skill.
---

# Brokle Domain Architecture Skill

Expert guidance for Brokle's Domain-Driven Design (DDD) architecture.

## 8 Active Domains (2 Planned)

| Domain | Location | Purpose | Status |
|--------|----------|---------|--------|
| auth | `internal/core/domain/auth` | Authentication, sessions, API keys | ✅ Active |
| billing | `internal/core/domain/billing` | Usage tracking, billing | ✅ Active |
| common | `internal/core/domain/common` | Transaction patterns, shared utilities | ✅ Active |
| config | `internal/core/domain/config` | Configuration management | 🔄 Planned (empty) |
| gateway | `internal/core/domain/gateway` | AI provider routing | ✅ Active |
| observability | `internal/core/domain/observability` | Traces, observations, quality scores | ✅ Active |
| organization | `internal/core/domain/organization` | Multi-tenant org management | ✅ Active |
| routing | `internal/core/domain/routing` | Advanced routing logic | 🔄 Planned (empty) |
| user | `internal/core/domain/user` | User management | ✅ Active |

## Domain Layer Structure

```go
internal/core/domain/{domain}/
├── entities.go          # Domain entities
├── repository.go        # Repository interfaces
├── service.go           # Service interfaces
├── errors.go            # Domain-specific errors
├── types.go             # Domain types and enums
└── validators.go        # Domain validation logic
```

## Entity Pattern

```go
// internal/core/domain/auth/entities.go
package auth

import (
    "time"
    "brokle/pkg/ulid"
)

type User struct {
    ID        ulid.ULID
    Email     string
    Name      string
    Status    UserStatus
    Role      Role
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Domain enums
type UserStatus string
const (
    UserStatusActive   UserStatus = "active"
    UserStatusInactive UserStatus = "inactive"
    UserStatusSuspended UserStatus = "suspended"
)

type Role string
const (
    RoleOwner  Role = "owner"
    RoleAdmin  Role = "admin"
    RoleUser   Role = "user"
    RoleViewer Role = "viewer"
)
```

## Domain Errors

```go
// internal/core/domain/auth/errors.go
package auth

import "errors"

var (
    ErrNotFound           = errors.New("user not found")
    ErrAlreadyExists      = errors.New("user already exists")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrSessionExpired     = errors.New("session expired")
)
```

## Repository Interfaces

```go
// internal/core/domain/auth/repository.go
package auth

import (
    "context"
    "brokle/pkg/ulid"
)

type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id ulid.ULID) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id ulid.ULID) error
    List(ctx context.Context, filter UserFilter) ([]*User, error)
}
```

## Service Interfaces

```go
// internal/core/domain/auth/service.go
package auth

import "context"

type AuthService interface {
    Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
    Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
    Logout(ctx context.Context, token string) error
    ValidateSession(ctx context.Context, token string) (*User, error)
}
```

## Multi-Tenant Scoping Patterns

**NOT all entities have `organization_id`** - scoping depends on entity type:

### 1. Organization-Scoped Entities (Direct `organization_id`)
```go
// organization/organization.go:54-57
type Project struct {
    ID             ulid.ULID
    OrganizationID ulid.ULID `json:"organization_id" gorm:"type:char(26);not null"`
    Name           string
}

// organization/organization.go:37-39
type Member struct {
    OrganizationID ulid.ULID `json:"organization_id" gorm:"type:char(26);not null;primaryKey"`
    UserID         ulid.ULID
}
```

### 2. Project-Scoped Entities (Organization via Project)
```go
// auth/auth.go:94 - APIKey is project-scoped
type APIKey struct {
    ID        ulid.ULID
    ProjectID ulid.ULID `json:"project_id" gorm:"type:char(26);not null;index"`
    // Organization derived via Project join
}
```

### 3. Scoped Entities (Flexible Scoping)
```go
// auth/auth.go:113-114 - Role uses flexible scope_type pattern
type Role struct {
    ID        ulid.ULID
    ScopeType string     `json:"scope_type" gorm:"size:20;not null"`  // "organization", "project", "global"
    ScopeID   *ulid.ULID `json:"scope_id,omitempty" gorm:"type:char(26);index"`
}
```

### 4. Global Entities (No organization_id)
```go
// user/user.go:14-39 - User is global with optional org reference
type User struct {
    ID    ulid.ULID
    Email string
    DefaultOrganizationID *ulid.ULID `json:"default_organization_id,omitempty" gorm:"type:char(26)"`
    // NOT required - users can belong to multiple orgs via Member table
}

// organization/organization.go:16 - Organization IS the tenant
type Organization struct {
    ID   ulid.ULID
    Name string
    // No organization_id - it IS the organization
}
```

**Reference Files**:
- Organization-scoped: `internal/core/domain/organization/organization.go:54-73`
- Project-scoped: `internal/core/domain/auth/auth.go:94`
- Scoped (flexible): `internal/core/domain/auth/auth.go:113-114`
- Global: `internal/core/domain/user/user.go:14-39`

## Cross-Domain Relationships

```go
// Example: Organization domain referencing User domain
package organization

import (
    userDomain "brokle/internal/core/domain/user"
)

type Member struct {
    ID             ulid.ULID
    OrganizationID ulid.ULID
    UserID         ulid.ULID  // References user domain
    Role           string
    Status         MemberStatus
}

// Service can accept interfaces from other domains
type OrganizationService struct {
    orgRepo    OrganizationRepository
    userRepo   userDomain.UserRepository  // Cross-domain dependency
    memberRepo MemberRepository
}
```

## Creating a New Domain

### Step 1: Create Domain Structure

```bash
mkdir -p internal/core/domain/my-domain
touch internal/core/domain/my-domain/{entities,repository,service,errors,types}.go
```

### Step 2: Define Entities

```go
// entities.go
package mydomain

import (
    "time"
    "brokle/pkg/ulid"
)

type MyEntity struct {
    ID             ulid.ULID
    OrganizationID ulid.ULID  // Always include for multi-tenancy
    Name           string
    Status         MyStatus
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

### Step 3: Define Domain Errors

```go
// errors.go
package mydomain

import "errors"

var (
    ErrNotFound      = errors.New("entity not found")
    ErrAlreadyExists = errors.New("entity already exists")
    ErrInvalidInput  = errors.New("invalid input")
)
```

### Step 4: Define Repository Interface

```go
// repository.go
package mydomain

import (
    "context"
    "brokle/pkg/ulid"
)

type MyEntityRepository interface {
    Create(ctx context.Context, entity *MyEntity) error
    GetByID(ctx context.Context, id ulid.ULID) (*MyEntity, error)
    Update(ctx context.Context, entity *MyEntity) error
    Delete(ctx context.Context, id ulid.ULID) error
}
```

### Step 5: Define Service Interface

```go
// service.go
package mydomain

import "context"

type MyDomainService interface {
    CreateEntity(ctx context.Context, req *CreateEntityRequest) (*CreateEntityResponse, error)
    GetEntity(ctx context.Context, id ulid.ULID) (*GetEntityResponse, error)
}
```

### Step 6: Implement Service

In `internal/core/services/my-domain/`

### Step 7: Implement Repository

In `internal/infrastructure/repository/my-domain/`

### Step 8: Register in DI Container

In `internal/app/app.go`

## Domain Validation

```go
// validators.go
package auth

import (
    "errors"
    "regexp"
)

func (u *User) Validate() error {
    if u.Email == "" {
        return errors.New("email is required")
    }
    if !isValidEmail(u.Email) {
        return errors.New("invalid email format")
    }
    if u.Name == "" {
        return errors.New("name is required")
    }
    return nil
}

func isValidEmail(email string) bool {
    return regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(email)
}
```

## Key Principles

1. **Domain Purity**: Domain layer has no external dependencies
2. **Multi-Tenancy**: All entities scoped by organization
3. **Domain Errors**: Use domain-specific errors
4. **Validation**: Domain entities validate themselves
5. **Interfaces**: Define repository and service interfaces in domain
6. **Cross-Domain**: Use domain aliases for cross-domain references

## References

- Existing domains in `internal/core/domain/` for patterns
- `CLAUDE.md` - Architecture overview
- `docs/development/PATTERNS.md` - Domain patterns
