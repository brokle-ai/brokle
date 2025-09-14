# 📝 Coding Standards

This document outlines the coding standards and best practices for the Brokle platform. Following these standards ensures consistency, maintainability, and quality across the codebase.

## General Principles

### Code Quality Principles

1. **Readability First** - Code is read more often than written
2. **Simplicity** - Choose the simplest solution that works
3. **Consistency** - Follow established patterns throughout the codebase
4. **Testability** - Write code that is easy to test
5. **Performance** - Optimize for performance without sacrificing readability

### SOLID Principles

- **Single Responsibility** - Each function/class should have one reason to change
- **Open/Closed** - Open for extension, closed for modification
- **Liskov Substitution** - Subtypes must be substitutable for their base types
- **Interface Segregation** - Many client-specific interfaces are better than one general-purpose interface
- **Dependency Inversion** - Depend on abstractions, not concretions

## Go Coding Standards

### Project Layout

Follow the standard Go project layout:

```
brokle/
├── cmd/                    # Main applications
├── internal/               # Private application and library code
├── pkg/                    # Library code that's ok to use by external apps
├── web/                    # Web application specific components
├── configs/                # Configuration file templates or defaults
├── deployments/            # System and container orchestration deployment configurations
├── tests/                  # Additional external test apps and test data
└── docs/                   # Design and user documents
```

### Naming Conventions

#### Package Names
```go
// Good - short, concise, lowercase
package user
package auth
package billing

// Bad - mixed case, underscores, too generic
package userManagement
package user_service
package utils
```

#### Interface Names
```go
// Good - describe what it does
type UserRepository interface{}
type EmailSender interface{}
type MetricsCollector interface{}

// Bad - generic or unclear
type UserInterface interface{}
type Manager interface{}
type Handler interface{}
```

#### Function Names
```go
// Good - clear, action-oriented
func CreateUser(ctx context.Context, user *User) error
func ValidateEmail(email string) bool
func CalculateUsageCost(usage *Usage) float64

// Bad - unclear or too generic
func Process(data interface{}) error
func DoIt() error
func Handle(req *Request) error
```

#### Variable Names
```go
// Good - descriptive but concise
func ProcessPayment(userID ulid.ULID, amount float64) error {
    user, err := userRepo.GetByID(ctx, userID)
    if err != nil {
        return err
    }
    
    payment := &Payment{
        UserID: userID,
        Amount: amount,
    }
    
    return paymentService.Process(ctx, payment)
}

// Bad - too short or too verbose
func ProcessPayment(u ulid.ULID, a float64) error {
    theUserFromDatabase, errorFromGettingUser := userRepo.GetByID(ctx, u)
    // ...
}
```

### Domain-Driven Design Patterns

#### Domain Structure

```go
// Domain entity - pure business object
type User struct {
    ID        ulid.ULID
    Email     string
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Domain repository interface - in domain layer
type Repository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id ulid.ULID) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id ulid.ULID) error
}

// Domain service interface - in domain layer  
type Service interface {
    Register(ctx context.Context, email, name, password string) (*User, error)
    Authenticate(ctx context.Context, email, password string) (*User, error)
    UpdateProfile(ctx context.Context, userID ulid.ULID, updates map[string]interface{}) error
}
```

#### Repository Implementation

```go
// Infrastructure layer - implements domain repository interface
type userRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) user.Repository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *user.User) error {
    query := `
        INSERT INTO users (id, email, name, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
    `
    
    _, err := r.db.ExecContext(ctx, query, u.ID, u.Email, u.Name, u.CreatedAt, u.UpdatedAt)
    if err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }
    
    return nil
}
```

#### Service Implementation

```go
// Service layer - implements domain service interface
type userService struct {
    repo     user.Repository
    hasher   PasswordHasher
    emailSvc EmailService
}

func NewUserService(repo user.Repository, hasher PasswordHasher, emailSvc EmailService) user.Service {
    return &userService{
        repo:     repo,
        hasher:   hasher,
        emailSvc: emailSvc,
    }
}

func (s *userService) Register(ctx context.Context, email, name, password string) (*user.User, error) {
    // Validation
    if err := validateEmail(email); err != nil {
        return nil, fmt.Errorf("invalid email: %w", err)
    }
    
    if err := validatePassword(password); err != nil {
        return nil, fmt.Errorf("invalid password: %w", err)
    }
    
    // Check if user already exists
    existingUser, err := s.repo.GetByEmail(ctx, email)
    if err == nil && existingUser != nil {
        return nil, ErrUserAlreadyExists
    }
    
    // Hash password
    hashedPassword, err := s.hasher.Hash(password)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %w", err)
    }
    
    // Create user
    newUser := &user.User{
        ID:        ulid.New(),
        Email:     email,
        Name:      name,
        Password:  hashedPassword,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    if err := s.repo.Create(ctx, newUser); err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    // Send welcome email (async)
    go func() {
        if err := s.emailSvc.SendWelcomeEmail(email, name); err != nil {
            // Log error but don't fail the registration
            log.WithError(err).Error("Failed to send welcome email")
        }
    }()
    
    return newUser, nil
}
```

### Error Handling

The Brokle platform implements **industrial-grade error handling patterns** across all architectural layers. These patterns ensure clean architecture, proper error propagation, and maintainable code.

#### 📖 Comprehensive Error Handling Documentation

For complete error handling implementation patterns, see:

- **[Error Handling Guide](./development/ERROR_HANDLING_GUIDE.md)** - Complete industrial patterns across Repository → Service → Handler layers
- **[Domain Alias Patterns](./development/DOMAIN_ALIAS_PATTERNS.md)** - Professional import patterns for clean, conflict-free code
- **[Quick Reference](./development/ERROR_HANDLING_QUICK_REFERENCE.md)** - Developer cheat sheet for daily development

#### Key Patterns Overview

**Repository Layer**: Domain errors with context wrapping
```go
// Professional domain alias imports
import authDomain "brokle/internal/core/domain/auth"

// GORM error handling with domain error wrapping
if err == gorm.ErrRecordNotFound {
    return nil, fmt.Errorf("get user by ID %s: %w", id, authDomain.ErrNotFound)
}
```

**Service Layer**: AppError constructors for business logic
```go
// Convert domain errors to business errors
if errors.Is(err, userDomain.ErrNotFound) {
    return nil, appErrors.NewNotFoundError("User not found")
}
```

**Handler Layer**: Structured HTTP response handling
```go
// Automatic HTTP status mapping
resp, err := h.userService.GetUser(c, userID)
if err != nil {
    response.Error(c, err) // Maps AppErrors to HTTP status codes
    return
}
response.Success(c, resp)
```

#### Architecture Benefits

- **Clean Architecture**: Proper layer separation with structured error flow
- **Domain Separation**: Professional domain aliases prevent naming conflicts
- **Industrial Standards**: Following Go best practices for enterprise applications
- **Maintainable**: Consistent patterns across 18+ repository files and 14+ services
- **Production Ready**: Comprehensive error context and automated HTTP mapping

### HTTP Handler Patterns

#### Standard Response Format

```go
// pkg/response/response.go
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *Error      `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
    Timestamp time.Time `json:"timestamp"`
    RequestID string    `json:"request_id"`
}

func Success(w http.ResponseWriter, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    
    response := &Response{
        Success: true,
        Data:    data,
        Meta: &Meta{
            Timestamp: time.Now(),
            RequestID: GetRequestID(w),
        },
    }
    
    json.NewEncoder(w).Encode(response)
}

func Error(w http.ResponseWriter, err error, statusCode int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    
    response := &Response{
        Success: false,
        Error: &Error{
            Message: err.Error(),
        },
        Meta: &Meta{
            Timestamp: time.Now(),
            RequestID: GetRequestID(w),
        },
    }
    
    json.NewEncoder(w).Encode(response)
}
```

#### Handler Structure

```go
// internal/transport/http/handlers/user.go
type UserHandler struct {
    service   user.Service
    validator *validator.Validate
}

func NewUserHandler(service user.Service, validator *validator.Validate) *UserHandler {
    return &UserHandler{
        service:   service,
        validator: validator,
    }
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
    // Parse request
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.Error(w, errors.NewValidationError("Invalid request body"), http.StatusBadRequest)
        return
    }
    
    // Validate request
    if err := h.validator.Struct(req); err != nil {
        response.Error(w, errors.NewValidationError(err.Error()), http.StatusBadRequest)
        return
    }
    
    // Call service
    user, err := h.service.Register(r.Context(), req.Email, req.Name, req.Password)
    if err != nil {
        // Handle different error types
        var validationErr *errors.ValidationError
        if errors.As(err, &validationErr) {
            response.Error(w, err, http.StatusBadRequest)
            return
        }
        
        response.Error(w, errors.NewInternalError("Failed to register user"), http.StatusInternalServerError)
        return
    }
    
    // Success response
    response.Success(w, map[string]interface{}{
        "user": user,
        "message": "User registered successfully",
    })
}
```

### Database Patterns

#### Connection Management

```go
// internal/infrastructure/database/postgres/connection.go
type DB struct {
    *sql.DB
}

func NewPostgresConnection(cfg *config.DatabaseConfig) (*DB, error) {
    dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
        cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.SSLMode)
    
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    
    // Connection pool settings
    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    
    // Test connection
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    
    return &DB{db}, nil
}
```

#### Repository Query Patterns

```go
// Good - use prepared statements for repeated queries
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
    query := `
        SELECT id, email, name, password_hash, created_at, updated_at
        FROM users 
        WHERE email = $1 AND deleted_at IS NULL
    `
    
    var u user.User
    err := r.db.QueryRowContext(ctx, query, email).Scan(
        &u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, errors.NewNotFoundError("user")
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get user by email: %w", err)
    }
    
    return &u, nil
}

// Good - use transactions for multi-table operations
func (r *userRepository) CreateWithProfile(ctx context.Context, u *user.User, profile *user.Profile) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()
    
    // Create user
    userQuery := `INSERT INTO users (id, email, name) VALUES ($1, $2, $3)`
    if _, err := tx.ExecContext(ctx, userQuery, u.ID, u.Email, u.Name); err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }
    
    // Create profile
    profileQuery := `INSERT INTO user_profiles (user_id, bio, avatar_url) VALUES ($1, $2, $3)`
    if _, err := tx.ExecContext(ctx, profileQuery, profile.UserID, profile.Bio, profile.AvatarURL); err != nil {
        return fmt.Errorf("failed to create profile: %w", err)
    }
    
    return tx.Commit()
}
```

### Testing Standards

#### Unit Test Structure

```go
// internal/core/services/user/user_service_test.go
func TestUserService_Register(t *testing.T) {
    tests := []struct {
        name        string
        email       string
        username    string
        password    string
        mockSetup   func(*mocks.MockUserRepository)
        expectedErr error
    }{
        {
            name:     "successful registration",
            email:    "test@example.com",
            username: "testuser",
            password: "password123",
            mockSetup: func(repo *mocks.MockUserRepository) {
                repo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, errors.NewNotFoundError("user"))
                repo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
            },
            expectedErr: nil,
        },
        {
            name:     "user already exists",
            email:    "existing@example.com",
            username: "existinguser",
            password: "password123",
            mockSetup: func(repo *mocks.MockUserRepository) {
                existingUser := &user.User{Email: "existing@example.com"}
                repo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
            },
            expectedErr: user.ErrUserAlreadyExists,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockRepo := &mocks.MockUserRepository{}
            mockHasher := &mocks.MockPasswordHasher{}
            mockEmailSvc := &mocks.MockEmailService{}
            
            service := NewUserService(mockRepo, mockHasher, mockEmailSvc)
            
            // Setup mocks
            tt.mockSetup(mockRepo)
            mockHasher.On("Hash", tt.password).Return("hashed_password", nil)
            mockEmailSvc.On("SendWelcomeEmail", tt.email, tt.username).Return(nil)
            
            // Execute
            user, err := service.Register(context.Background(), tt.email, tt.username, tt.password)
            
            // Assert
            if tt.expectedErr != nil {
                assert.Error(t, err)
                assert.Equal(t, tt.expectedErr, err)
                assert.Nil(t, user)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, user)
                assert.Equal(t, tt.email, user.Email)
            }
            
            // Verify mocks
            mockRepo.AssertExpectations(t)
            mockHasher.AssertExpectations(t)
        })
    }
}
```

#### Integration Test Patterns

```go
// tests/integration/user_test.go
func TestUserIntegration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Setup dependencies
    repo := repository.NewUserRepository(db)
    hasher := bcrypt.NewPasswordHasher()
    emailSvc := &mocks.MockEmailService{}
    service := services.NewUserService(repo, hasher, emailSvc)
    
    t.Run("full user lifecycle", func(t *testing.T) {
        ctx := context.Background()
        
        // Register user
        user, err := service.Register(ctx, "test@example.com", "testuser", "password123")
        require.NoError(t, err)
        require.NotNil(t, user)
        
        // Authenticate user
        authUser, err := service.Authenticate(ctx, "test@example.com", "password123")
        require.NoError(t, err)
        require.Equal(t, user.ID, authUser.ID)
        
        // Update profile
        err = service.UpdateProfile(ctx, user.ID, map[string]interface{}{
            "name": "Updated Name",
        })
        require.NoError(t, err)
        
        // Verify update
        updatedUser, err := service.GetUser(ctx, user.ID)
        require.NoError(t, err)
        require.Equal(t, "Updated Name", updatedUser.Name)
    })
}
```

### Logging Standards

```go
// Use structured logging
import "github.com/sirupsen/logrus"

// Good - structured logging with context
func (s *userService) Register(ctx context.Context, email, name, password string) (*user.User, error) {
    logger := logrus.WithFields(logrus.Fields{
        "operation": "user.register",
        "email":     email,
        "request_id": GetRequestID(ctx),
    })
    
    logger.Info("Starting user registration")
    
    // ... business logic
    
    if err != nil {
        logger.WithError(err).Error("Failed to register user")
        return nil, err
    }
    
    logger.WithField("user_id", user.ID).Info("User registered successfully")
    return user, nil
}
```

## TypeScript/React Standards

### Project Structure

```
web/src/
├── app/                   # Next.js App Router
├── components/           # Reusable components
│   ├── ui/              # Base components (shadcn/ui)
│   └── feature/         # Feature-specific components
├── hooks/               # Custom React hooks
├── lib/                 # Utilities and configurations
├── store/               # State management
└── types/               # TypeScript definitions
```

### TypeScript Best Practices

#### Type Definitions

```typescript
// types/api.ts - API response types
export interface User {
  id: string
  email: string
  name: string
  createdAt: string
  updatedAt: string
}

export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: {
    type: string
    message: string
    code: string
  }
  meta?: {
    timestamp: string
    requestId: string
  }
}

// Use generic types for reusable patterns
export interface PaginatedResponse<T> extends ApiResponse<T[]> {
  meta: {
    page: number
    limit: number
    total: number
    totalPages: number
  }
}
```

#### Component Patterns

```typescript
// components/user/UserProfile.tsx
interface UserProfileProps {
  user: User
  onUpdate: (updates: Partial<User>) => void
  className?: string
}

export function UserProfile({ user, onUpdate, className }: UserProfileProps) {
  const [isEditing, setIsEditing] = useState(false)
  
  const handleSave = async (updates: Partial<User>) => {
    try {
      await onUpdate(updates)
      setIsEditing(false)
    } catch (error) {
      console.error('Failed to update user:', error)
    }
  }
  
  return (
    <div className={cn("user-profile", className)}>
      {isEditing ? (
        <UserEditForm user={user} onSave={handleSave} onCancel={() => setIsEditing(false)} />
      ) : (
        <UserDisplay user={user} onEdit={() => setIsEditing(true)} />
      )}
    </div>
  )
}
```

#### Custom Hooks

```typescript
// hooks/useApi.ts
import { useState, useEffect } from 'react'

interface UseApiResult<T> {
  data: T | null
  loading: boolean
  error: string | null
  refetch: () => void
}

export function useApi<T>(endpoint: string): UseApiResult<T> {
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  
  const fetchData = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await apiClient.request<T>(endpoint)
      setData(response)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred')
    } finally {
      setLoading(false)
    }
  }
  
  useEffect(() => {
    fetchData()
  }, [endpoint])
  
  return { data, loading, error, refetch: fetchData }
}

// Usage
function UserList() {
  const { data: users, loading, error, refetch } = useApi<User[]>('/users')
  
  if (loading) return <Loading />
  if (error) return <Error message={error} onRetry={refetch} />
  
  return (
    <ul>
      {users?.map(user => (
        <li key={user.id}>{user.name}</li>
      ))}
    </ul>
  )
}
```

### State Management with Zustand

```typescript
// store/userStore.ts
import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'

interface UserState {
  users: User[]
  selectedUser: User | null
  loading: boolean
  error: string | null
}

interface UserActions {
  fetchUsers: () => Promise<void>
  selectUser: (user: User) => void
  updateUser: (id: string, updates: Partial<User>) => Promise<void>
  clearError: () => void
}

export const useUserStore = create<UserState & UserActions>()(
  devtools(
    persist(
      (set, get) => ({
        // State
        users: [],
        selectedUser: null,
        loading: false,
        error: null,
        
        // Actions
        fetchUsers: async () => {
          set({ loading: true, error: null })
          try {
            const response = await apiClient.request<User[]>('/users')
            set({ users: response, loading: false })
          } catch (error) {
            set({ 
              error: error instanceof Error ? error.message : 'Failed to fetch users',
              loading: false 
            })
          }
        },
        
        selectUser: (user) => set({ selectedUser: user }),
        
        updateUser: async (id, updates) => {
          try {
            const updatedUser = await apiClient.request<User>(`/users/${id}`, {
              method: 'PATCH',
              body: JSON.stringify(updates),
            })
            
            set((state) => ({
              users: state.users.map(user => 
                user.id === id ? updatedUser : user
              ),
              selectedUser: state.selectedUser?.id === id ? updatedUser : state.selectedUser,
            }))
          } catch (error) {
            set({ 
              error: error instanceof Error ? error.message : 'Failed to update user'
            })
          }
        },
        
        clearError: () => set({ error: null }),
      }),
      {
        name: 'user-storage',
        partialize: (state) => ({ selectedUser: state.selectedUser }),
      }
    ),
    { name: 'user-store' }
  )
)
```

## Code Review Guidelines

### What to Look For

#### Architecture & Design
- [ ] Follows domain-driven design principles
- [ ] Proper separation of concerns
- [ ] Interface-based design for testability
- [ ] Appropriate error handling

#### Code Quality
- [ ] Clear, descriptive naming
- [ ] Functions are single-purpose and reasonably sized
- [ ] Proper error handling and logging
- [ ] No code duplication

#### Performance
- [ ] Efficient database queries with proper indexing
- [ ] Appropriate caching strategies
- [ ] No N+1 query problems
- [ ] Reasonable resource usage

#### Security
- [ ] Input validation
- [ ] SQL injection prevention
- [ ] Authentication and authorization
- [ ] No sensitive data in logs

#### Testing
- [ ] Unit tests for business logic
- [ ] Integration tests for critical paths
- [ ] Proper test coverage
- [ ] Tests are readable and maintainable

### Review Process

1. **Automated Checks** - All CI checks must pass
2. **Self Review** - Author reviews their own code first
3. **Peer Review** - At least one peer review required
4. **Approval** - Code owner approval for significant changes

---

Following these coding standards ensures that the Brokle platform maintains high code quality, remains maintainable, and continues to scale effectively as the project grows.