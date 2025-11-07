# Code Templates for Brokle Backend Development

## Repository Template

```go
package auth

import (
    "context"
    "fmt"

    "gorm.io/gorm"

    authDomain "brokle/internal/core/domain/auth"
    "brokle/pkg/ulid"
)

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) authDomain.UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *authDomain.User) error {
    if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
        return fmt.Errorf("create user %s: %w", user.Email, err)
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
        return nil, fmt.Errorf("database query failed for user ID %s: %w", id, err)
    }
    return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*authDomain.User, error) {
    var user authDomain.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, fmt.Errorf("get user by email %s: %w", email, authDomain.ErrNotFound)
        }
        return nil, fmt.Errorf("database query failed for email %s: %w", email, err)
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

func (r *userRepository) List(ctx context.Context, filter authDomain.UserFilter) ([]*authDomain.User, error) {
    var users []*authDomain.User
    query := r.db.WithContext(ctx)

    // Apply filters
    if filter.OrganizationID != nil {
        query = query.Where("organization_id = ?", filter.OrganizationID)
    }
    if filter.Status != "" {
        query = query.Where("status = ?", filter.Status)
    }
    if filter.Search != "" {
        query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
    }

    // Pagination
    offset := (filter.Page - 1) * filter.Limit
    query = query.Offset(offset).Limit(filter.Limit)

    // Sorting
    query = query.Order(filter.SortBy + " " + filter.SortDir)

    if err := query.Find(&users).Error; err != nil {
        return nil, fmt.Errorf("list users: %w", err)
    }

    return users, nil
}
```

## Service Template

```go
package auth

import (
    "context"
    "errors"

    authDomain "brokle/internal/core/domain/auth"
    "brokle/pkg/apperrors"
    "brokle/pkg/ulid"
)

type userService struct {
    userRepo authDomain.UserRepository
}

func NewUserService(userRepo authDomain.UserRepository) authDomain.UserService {
    return &userService{userRepo: userRepo}
}

func (s *userService) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
    // Input validation
    if err := req.Validate(); err != nil {
        return nil, appErrors.NewValidationError("Invalid user data", err)
    }

    // Check for existing user
    existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
    if err != nil && !errors.Is(err, authDomain.ErrNotFound) {
        return nil, appErrors.NewInternalError("Failed to check existing user", err)
    }
    if existingUser != nil {
        return nil, appErrors.NewConflictError("User already exists with this email")
    }

    // Create user entity
    user := &authDomain.User{
        ID:     ulid.New(),
        Email:  req.Email,
        Name:   req.Name,
        Status: authDomain.UserStatusActive,
        Role:   authDomain.RoleUser,
    }

    // Persist user
    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, appErrors.NewInternalError("Failed to create user", err)
    }

    return &CreateUserResponse{User: user}, nil
}

func (s *userService) GetUser(ctx context.Context, id ulid.ULID) (*GetUserResponse, error) {
    user, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, authDomain.ErrNotFound) {
            return nil, appErrors.NewNotFoundError("User")  // resource string
        }
        return nil, appErrors.NewInternalError("Failed to retrieve user", err)
    }

    return &GetUserResponse{User: user}, nil
}

func (s *userService) UpdateUser(ctx context.Context, req *UpdateUserRequest) (*UpdateUserResponse, error) {
    // Get existing user
    user, err := s.userRepo.GetByID(ctx, req.ID)
    if err != nil {
        if errors.Is(err, authDomain.ErrNotFound) {
            return nil, appErrors.NewNotFoundError("User")
        }
        return nil, appErrors.NewInternalError("Failed to retrieve user", err)
    }

    // Update fields (only if provided)
    if req.Name != nil {
        user.Name = *req.Name
    }
    if req.Email != nil {
        user.Email = *req.Email
    }
    if req.Status != nil {
        user.Status = authDomain.UserStatus(*req.Status)
    }

    // Persist changes
    if err := s.userRepo.Update(ctx, user); err != nil {
        return nil, appErrors.NewInternalError("Failed to update user", err)
    }

    return &UpdateUserResponse{User: user}, nil
}

func (s *userService) DeleteUser(ctx context.Context, id ulid.ULID) error {
    // Verify user exists
    _, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, authDomain.ErrNotFound) {
            return appErrors.NewNotFoundError("User")
        }
        return appErrors.NewInternalError("Failed to retrieve user", err)
    }

    // Delete user
    if err := s.userRepo.Delete(ctx, id); err != nil {
        return appErrors.NewInternalError("Failed to delete user", err)
    }

    return nil
}

func (s *userService) ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
    // Build filter
    filter := authDomain.UserFilter{
        Page:    req.Page,
        Limit:   req.Limit,
        Search:  req.Search,
        Status:  req.Status,
        SortBy:  req.SortBy,
        SortDir: req.SortDir,
    }

    users, err := s.userRepo.List(ctx, filter)
    if err != nil {
        return nil, appErrors.NewInternalError("Failed to list users", err)
    }

    // Get total count for pagination
    total, err := s.userRepo.Count(ctx, filter)
    if err != nil {
        return nil, appErrors.NewInternalError("Failed to count users", err)
    }

    pagination := &Pagination{
        Page:      req.Page,
        PageSize:  req.Limit,
        Total:     total,
        TotalPage: (int(total) + req.Limit - 1) / req.Limit,
        HasNext:   req.Page*req.Limit < int(total),
        HasPrev:   req.Page > 1,
    }

    return &ListUsersResponse{
        Users:      users,
        Pagination: pagination,
    }, nil
}
```

## Handler Template

```go
package http

import (
    "github.com/gin-gonic/gin"

    authDomain "brokle/internal/core/domain/auth"
    "brokle/internal/transport/http/middleware"
    "brokle/pkg/apperrors"
    "brokle/pkg/response"
    "brokle/pkg/ulid"
)

type UserHandler struct {
    userService authDomain.UserService
}

func NewUserHandler(userService authDomain.UserService) *UserHandler {
    return &UserHandler{userService: userService}
}

// RegisterRoutes sets up all user-related routes
// Note: Called from server setup where middleware is already applied to parent group (server.go:221-223)
func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
    users := r.Group("/users")
    {
        users.POST("", h.CreateUser)
        users.GET("", h.ListUsers)
        users.GET("/:id", h.GetUser)
        users.PUT("/:id", h.UpdateUser)
        users.DELETE("/:id", h.DeleteUser)
    }
}

// @Summary      Create a new user
// @Description  Create a new user account with the provided information
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request  body      CreateUserRequest  true  "User creation data"
// @Success      201      {object}  CreateUserResponse
// @Failure      400      {object}  response.ErrorResponse
// @Failure      409      {object}  response.ErrorResponse
// @Failure      500      {object}  response.ErrorResponse
// @Router       /api/v1/auth/users [post]
// @Security     ApiKeyAuth
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, appErrors.NewValidationError("Invalid request body", err))
        return
    }

    resp, err := h.userService.CreateUser(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, err)
        return
    }

    response.Created(c, resp)
}

// @Summary      List users
// @Description  Retrieve a paginated list of users with optional filtering
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "Page number"       minimum(1)
// @Param        limit     query     int     false  "Items per page"    minimum(1) maximum(100)
// @Param        search    query     string  false  "Search term"
// @Param        status    query     string  false  "User status"       Enums(active, inactive, all)
// @Success      200       {object}  ListUsersResponse
// @Failure      400       {object}  response.ErrorResponse
// @Router       /api/v1/auth/users [get]
// @Security     ApiKeyAuth
func (h *UserHandler) ListUsers(c *gin.Context) {
    var req ListUsersRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        response.Error(c, appErrors.NewValidationError("Invalid query parameters", err))
        return
    }

    // Set defaults
    if req.Page == 0 {
        req.Page = 1
    }
    if req.Limit == 0 {
        req.Limit = 20
    }
    if req.SortBy == "" {
        req.SortBy = "created_at"
    }
    if req.SortDir == "" {
        req.SortDir = "desc"
    }

    resp, err := h.userService.ListUsers(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, err)
        return
    }

    response.Success(c, resp)
}

// @Summary      Get user by ID
// @Description  Retrieve a user by their unique identifier
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  GetUserResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /api/v1/auth/users/{id} [get]
// @Security     ApiKeyAuth
func (h *UserHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")
    id, err := ulid.Parse(userID)
    if err != nil {
        response.Error(c, appErrors.NewValidationError("Invalid user ID format", "id must be a valid ULID"))
        return
    }

    resp, err := h.userService.GetUser(c.Request.Context(), id)
    if err != nil {
        response.Error(c, err)
        return
    }

    response.Success(c, resp)
}

// @Summary      Update user
// @Description  Update user information
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id       path      string             true  "User ID"
// @Param        request  body      UpdateUserRequest  true  "Update data"
// @Success      200      {object}  UpdateUserResponse
// @Failure      400      {object}  response.ErrorResponse
// @Failure      404      {object}  response.ErrorResponse
// @Router       /api/v1/auth/users/{id} [put]
// @Security     ApiKeyAuth
func (h *UserHandler) UpdateUser(c *gin.Context) {
    userID := c.Param("id")
    id, err := ulid.Parse(userID)
    if err != nil {
        response.Error(c, appErrors.NewValidationError("Invalid user ID format", "id must be a valid ULID"))
        return
    }

    var req UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, appErrors.NewValidationError("Invalid request body", err))
        return
    }

    req.ID = id
    resp, err := h.userService.UpdateUser(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, err)
        return
    }

    response.Success(c, resp)
}

// @Summary      Delete user
// @Description  Delete a user by ID
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      204  "No Content"
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /api/v1/auth/users/{id} [delete]
// @Security     ApiKeyAuth
func (h *UserHandler) DeleteUser(c *gin.Context) {
    userID := c.Param("id")
    id, err := ulid.Parse(userID)
    if err != nil {
        response.Error(c, appErrors.NewValidationError("Invalid user ID format", "id must be a valid ULID"))
        return
    }

    if err := h.userService.DeleteUser(c.Request.Context(), id); err != nil {
        response.Error(c, err)
        return
    }

    response.NoContent(c)
}
```

## Request/Response Types

```go
package auth

import (
    "time"

    authDomain "brokle/internal/core/domain/auth"
    "brokle/pkg/ulid"
)

// Create User
type CreateUserRequest struct {
    Email    string `json:"email" binding:"required,email" example:"user@example.com"`
    Name     string `json:"name" binding:"required,min=2,max=100" example:"John Doe"`
    Password string `json:"password" binding:"required,min=8" example:"secure123"`
    Role     string `json:"role,omitempty" binding:"omitempty,oneof=admin user viewer" example:"user"`
}

func (r *CreateUserRequest) Validate() error {
    // Additional custom validation beyond struct tags
    return nil
}

type CreateUserResponse struct {
    User *User `json:"user"`
}

// Get User
type GetUserResponse struct {
    User *User `json:"user"`
}

// Update User
type UpdateUserRequest struct {
    ID     ulid.ULID `json:"-"`
    Name   *string   `json:"name,omitempty" binding:"omitempty,min=2,max=100" example:"Jane Doe"`
    Email  *string   `json:"email,omitempty" binding:"omitempty,email" example:"jane@example.com"`
    Status *string   `json:"status,omitempty" binding:"omitempty,oneof=active inactive" example:"active"`
}

type UpdateUserResponse struct {
    User *User `json:"user"`
}

// List Users
type ListUsersRequest struct {
    Page    int    `form:"page" binding:"omitempty,min=1" example:"1"`
    Limit   int    `form:"limit" binding:"omitempty,min=1,max=100" example:"20"`
    Search  string `form:"search" binding:"omitempty,max=100" example:"john"`
    Status  string `form:"status" binding:"omitempty,oneof=active inactive all" example:"active"`
    SortBy  string `form:"sort_by" binding:"omitempty,oneof=created_at updated_at name email" example:"created_at"`
    SortDir string `form:"sort_dir" binding:"omitempty,oneof=asc desc" example:"desc"`
}

type ListUsersResponse struct {
    Users      []*User     `json:"users"`
    Pagination *Pagination `json:"pagination"`
}

// Pagination
type Pagination struct {
    Page      int   `json:"page" example:"2"`
    PageSize  int   `json:"page_size" example:"20"`
    Total     int64 `json:"total" example:"156"`
    TotalPage int   `json:"total_page" example:"8"`
    HasNext   bool  `json:"has_next" example:"true"`
    HasPrev   bool  `json:"has_prev" example:"true"`
}

// User DTO
type User struct {
    ID        string    `json:"id" example:"01H4XJZQX3EXAMPLE"`
    Email     string    `json:"email" example:"user@example.com"`
    Name      string    `json:"name" example:"John Doe"`
    Status    string    `json:"status" example:"active"`
    Role      string    `json:"role" example:"user"`
    CreatedAt time.Time `json:"created_at" example:"2023-07-01T12:00:00Z"`
    UpdatedAt time.Time `json:"updated_at" example:"2023-07-01T12:00:00Z"`
}
```

## Mock Repository Template (for tests)

```go
package auth

import (
    "context"

    "github.com/stretchr/testify/mock"

    authDomain "brokle/internal/core/domain/auth"
    "brokle/pkg/ulid"
)

type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *authDomain.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id ulid.ULID) (*authDomain.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*authDomain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*authDomain.User, error) {
    args := m.Called(ctx, email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*authDomain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *authDomain.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id ulid.ULID) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, filter authDomain.UserFilter) ([]*authDomain.User, error) {
    args := m.Called(ctx, filter)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*authDomain.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context, filter authDomain.UserFilter) (int64, error) {
    args := m.Called(ctx, filter)
    return args.Get(0).(int64), args.Error(1)
}
```
