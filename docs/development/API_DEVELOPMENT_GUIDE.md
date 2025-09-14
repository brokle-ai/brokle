# API Development Guide

This guide ensures consistent, professional API development across the Brokle platform. Follow these patterns for maintainable, scalable, and user-friendly APIs.

## Table of Contents

- [Overview](#overview)
- [API Architecture](#api-architecture)
- [Handler Development Patterns](#handler-development-patterns)
- [Request/Response Standards](#requestresponse-standards)
- [Error Handling](#error-handling)
- [Validation Patterns](#validation-patterns)
- [Pagination Standards](#pagination-standards)
- [Authentication & Authorization](#authentication--authorization)
- [Documentation Requirements](#documentation-requirements)
- [Testing Standards](#testing-standards)
- [Examples](#examples)

## Overview

The Brokle platform follows **RESTful API design principles** with **clean architecture patterns**. Every API endpoint must adhere to these standards for consistency, maintainability, and excellent developer experience.

### Core Principles

1. **Consistent Structure**: All endpoints follow the same patterns
2. **Industrial Error Handling**: Structured error responses with proper HTTP status codes
3. **Comprehensive Validation**: Input validation with meaningful error messages
4. **Proper Authentication**: Secure endpoints with appropriate authorization
5. **OpenAPI Documentation**: Auto-generated, comprehensive API docs

## API Architecture

### URL Structure Standards

```
/api/v1/{domain}/{resource}[/{id}][/{sub-resource}]

Examples:
GET    /api/v1/auth/users                    # List users
POST   /api/v1/auth/users                    # Create user
GET    /api/v1/auth/users/{id}               # Get user
PUT    /api/v1/auth/users/{id}               # Update user
DELETE /api/v1/auth/users/{id}               # Delete user
GET    /api/v1/auth/users/{id}/sessions      # Get user sessions
```

### Domain Organization

| Domain | Prefix | Resources |
|--------|--------|-----------|
| Authentication | `/api/v1/auth/` | users, sessions, tokens, roles, permissions |
| Organizations | `/api/v1/organizations/` | organizations, projects, environments, members |
| Analytics | `/api/v1/analytics/` | metrics, reports, dashboards |
| Billing | `/api/v1/billing/` | subscriptions, usage, invoices |
| Gateway | `/api/v1/gateway/` | routes, providers, configurations |

## Handler Development Patterns

### Standard Handler Structure

```go
package http

import (
    "strconv"
    
    "github.com/gin-gonic/gin"
    
    authDomain "brokle/internal/core/domain/auth"
    "brokle/pkg/response"
    "brokle/pkg/ulid"
    "brokle/pkg/apperrors"
)

type UserHandler struct {
    userService authDomain.UserService
    authService authDomain.AuthService
}

func NewUserHandler(
    userService authDomain.UserService,
    authService authDomain.AuthService,
) *UserHandler {
    return &UserHandler{
        userService: userService,
        authService: authService,
    }
}

// RegisterRoutes sets up all user-related routes
func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
    users := r.Group("/users")
    {
        users.POST("", h.CreateUser)
        users.GET("", h.ListUsers)
        users.GET("/:id", h.GetUser)
        users.PUT("/:id", h.UpdateUser)
        users.DELETE("/:id", h.DeleteUser)
        
        // Sub-resources
        users.GET("/:id/sessions", h.GetUserSessions)
        users.POST("/:id/sessions", h.CreateUserSession)
    }
}
```

### Handler Method Template

```go
// @Summary      Get user by ID
// @Description  Retrieve a user by their unique identifier
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  GetUserResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /api/v1/auth/users/{id} [get]
// @Security     ApiKeyAuth
func (h *UserHandler) GetUser(c *gin.Context) {
    // 1. Extract and validate path parameters
    userID := c.Param("id")
    id, err := ulid.Parse(userID)
    if err != nil {
        response.Error(c, appErrors.NewValidationError("Invalid user ID format", map[string]string{
            "id": "must be a valid ULID",
        }))
        return
    }

    // 2. Extract and validate query parameters (if any)
    includeProfile := c.Query("include_profile") == "true"
    
    // 3. Call service layer
    resp, err := h.userService.GetUser(c.Request.Context(), &GetUserRequest{
        ID:             id,
        IncludeProfile: includeProfile,
    })
    if err != nil {
        response.Error(c, err) // Automatic HTTP status mapping
        return
    }

    // 4. Return success response
    response.Success(c, resp)
}
```

## Request/Response Standards

### Request Structures

```go
// Use clear, descriptive request structures
type CreateUserRequest struct {
    Email     string `json:"email" binding:"required,email" example:"user@example.com"`
    Name      string `json:"name" binding:"required,min=2,max=100" example:"John Doe"`
    Password  string `json:"password" binding:"required,min=8" example:"secure123"`
    Role      string `json:"role,omitempty" binding:"omitempty,oneof=admin user viewer" example:"user"`
}

type UpdateUserRequest struct {
    Name   *string `json:"name,omitempty" binding:"omitempty,min=2,max=100" example:"Jane Doe"`
    Email  *string `json:"email,omitempty" binding:"omitempty,email" example:"jane@example.com"`
    Status *string `json:"status,omitempty" binding:"omitempty,oneof=active inactive" example:"active"`
}

type ListUsersRequest struct {
    Page     int    `form:"page" binding:"omitempty,min=1" example:"1"`
    Limit    int    `form:"limit" binding:"omitempty,min=1,max=100" example:"20"`
    Search   string `form:"search" binding:"omitempty,max=100" example:"john"`
    Status   string `form:"status" binding:"omitempty,oneof=active inactive all" example:"active"`
    SortBy   string `form:"sort_by" binding:"omitempty,oneof=created_at updated_at name email" example:"created_at"`
    SortDir  string `form:"sort_dir" binding:"omitempty,oneof=asc desc" example:"desc"`
}
```

### Response Structures

```go
// Use consistent response wrappers
type CreateUserResponse struct {
    User *User `json:"user"`
}

type GetUserResponse struct {
    User    *User         `json:"user"`
    Profile *UserProfile  `json:"profile,omitempty"`
}

type ListUsersResponse struct {
    Users      []*User     `json:"users"`
    Pagination *Pagination `json:"pagination"`
}

type UpdateUserResponse struct {
    User *User `json:"user"`
}

// Standard pagination structure
type Pagination struct {
    Page       int `json:"page"`
    Limit      int `json:"limit"`
    Total      int `json:"total"`
    TotalPages int `json:"total_pages"`
    HasNext    bool `json:"has_next"`
    HasPrev    bool `json:"has_prev"`
}

// Standardized entity representation
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

## Error Handling

### Standard Error Response Format

All errors use the standardized `response.Error()` function for consistent format:

```json
{
  "error": {
    "type": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": {
      "email": "must be a valid email address",
      "password": "must be at least 8 characters"
    },
    "code": "INVALID_INPUT",
    "request_id": "req_01H4XJZQX3EXAMPLE"
  }
}
```

### Error Handling Patterns

```go
// ✅ Input validation errors
if err := c.ShouldBindJSON(&req); err != nil {
    response.Error(c, appErrors.NewValidationError("Invalid request body", err))
    return
}

// ✅ Parameter validation errors
id, err := ulid.Parse(c.Param("id"))
if err != nil {
    response.Error(c, appErrors.NewValidationError("Invalid ID format", map[string]string{
        "id": "must be a valid ULID",
    }))
    return
}

// ✅ Service layer error handling (automatic HTTP mapping)
resp, err := h.service.Method(c.Request.Context(), req)
if err != nil {
    response.Error(c, err) // Maps AppErrors to appropriate HTTP status
    return
}

// ✅ Custom business logic errors
if !h.authService.HasPermission(userID, "users:read") {
    response.Error(c, appErrors.NewForbiddenError("Insufficient permissions"))
    return
}
```

### HTTP Status Code Mapping

| AppError Type | HTTP Status | Description |
|---------------|-------------|-------------|
| `ValidationError` | 400 Bad Request | Invalid input data |
| `UnauthorizedError` | 401 Unauthorized | Authentication required |
| `ForbiddenError` | 403 Forbidden | Insufficient permissions |
| `NotFoundError` | 404 Not Found | Resource not found |
| `ConflictError` | 409 Conflict | Resource already exists |
| `RateLimitError` | 429 Too Many Requests | Rate limit exceeded |
| `InternalError` | 500 Internal Server Error | Unexpected server error |

## Validation Patterns

### Input Validation Standards

```go
// Use Gin's binding validation with custom tags
type CreateProjectRequest struct {
    Name         string `json:"name" binding:"required,min=2,max=100,project_name" example:"My Project"`
    Description  string `json:"description" binding:"omitempty,max=500" example:"Project description"`
    OrganizationID string `json:"organization_id" binding:"required,ulid" example:"01H4XJZQX3EXAMPLE"`
    Settings     *ProjectSettings `json:"settings,omitempty"`
}

// Custom validation functions
func validateProjectName(fl validator.FieldLevel) bool {
    name := fl.Field().String()
    // Project name validation logic
    return regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`).MatchString(name)
}

// Register custom validators
func RegisterValidators() {
    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
        v.RegisterValidation("project_name", validateProjectName)
        v.RegisterValidation("ulid", validateULID)
    }
}
```

### Query Parameter Validation

```go
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
    
    // Continue with service call...
}
```

## Pagination Standards

All list endpoints must implement **consistent pagination** following platform standards.

### 📖 Comprehensive Pagination Guide

For detailed pagination implementation patterns, see **[PAGINATION_GUIDE.md](./PAGINATION_GUIDE.md)** which covers:

- Offset-based and cursor-based pagination
- Request/response format standards  
- Database implementation patterns
- Performance optimization strategies
- Advanced filtering and search patterns

### Quick Pagination Pattern

```go
// Standard paginated request
type ListUsersRequest struct {
    Page    int    `form:"page" binding:"omitempty,min=1" example:"1"`
    Limit   int    `form:"limit" binding:"omitempty,min=1,max=100" example:"20"`
    SortBy  string `form:"sort_by" example:"created_at"`
    SortDir string `form:"sort_dir" binding:"omitempty,oneof=asc desc" example:"desc"`
    Search  string `form:"search" example:"john"`
}

// Standard paginated response
type ListUsersResponse struct {
    Users      []*User     `json:"users"`
    Pagination *Pagination `json:"pagination"`
}

type Pagination struct {
    Page      int   `json:"page" example:"2"`
    PageSize  int   `json:"page_size" example:"20"`
    Total     int64 `json:"total" example:"156"`
    TotalPage int   `json:"total_page" example:"8"`
    HasNext   bool  `json:"has_next" example:"true"`
    HasPrev   bool  `json:"has_prev" example:"true"`
}
```

### Handler Implementation

```go
func (h *UserHandler) ListUsers(c *gin.Context) {
    var req ListUsersRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        response.Error(c, appErrors.NewValidationError("Invalid query parameters", err))
        return
    }
    
    // Apply defaults
    if req.Page == 0 { req.Page = 1 }
    if req.Limit == 0 { req.Limit = 20 }
    if req.SortBy == "" { req.SortBy = "created_at" }
    if req.SortDir == "" { req.SortDir = "desc" }
    
    resp, err := h.userService.ListUsers(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, err)
        return
    }
    
    response.Success(c, resp)
}
```

## Authentication & Authorization

### Authentication Middleware

```go
// All protected endpoints must use authentication middleware
func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
    // Apply authentication middleware
    authenticated := r.Group("")
    authenticated.Use(middleware.Authentication())
    
    users := authenticated.Group("/users")
    {
        users.POST("", middleware.RequirePermission("users:create"), h.CreateUser)
        users.GET("", middleware.RequirePermission("users:read"), h.ListUsers)
        users.GET("/:id", middleware.RequirePermission("users:read"), h.GetUser)
        users.PUT("/:id", middleware.RequirePermission("users:write"), h.UpdateUser)
        users.DELETE("/:id", middleware.RequirePermission("users:delete"), h.DeleteUser)
    }
}
```

### Permission Checking in Handlers

```go
func (h *UserHandler) UpdateUser(c *gin.Context) {
    userID := c.Param("id")
    currentUser := middleware.GetCurrentUser(c)
    
    // Check if user can modify this resource
    if currentUser.ID != userID && !currentUser.HasPermission("users:write:all") {
        response.Error(c, appErrors.NewForbiddenError("Can only update your own profile"))
        return
    }
    
    // Continue with update logic...
}
```

## Documentation Requirements

### OpenAPI/Swagger Annotations

Every handler method must include complete Swagger documentation:

```go
// @Summary      Create a new user
// @Description  Create a new user account with the provided information
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request  body      CreateUserRequest  true  "User creation data"
// @Success      201      {object}  CreateUserResponse
// @Failure      400      {object}  response.ErrorResponse  "Validation error"
// @Failure      409      {object}  response.ErrorResponse  "User already exists"
// @Failure      500      {object}  response.ErrorResponse  "Internal server error"
// @Router       /api/v1/auth/users [post]
// @Security     ApiKeyAuth
```

### Required Documentation Elements

- **@Summary**: Brief description (1 line)
- **@Description**: Detailed description
- **@Tags**: Group endpoints logically
- **@Accept/@Produce**: Content types
- **@Param**: All parameters with types and validation
- **@Success**: All success responses
- **@Failure**: All possible error responses
- **@Router**: Exact route and method
- **@Security**: Authentication requirements

## Testing Standards

### Handler Testing Pattern

```go
func TestUserHandler_GetUser(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    mockService := &MockUserService{}
    handler := NewUserHandler(mockService, nil)
    
    router := gin.New()
    handler.RegisterRoutes(router.Group("/api/v1/auth"))
    
    tests := []struct {
        name           string
        userID         string
        mockResponse   *GetUserResponse
        mockError      error
        expectedStatus int
        expectedBody   string
    }{
        {
            name:   "successful user retrieval",
            userID: "01H4XJZQX3EXAMPLE",
            mockResponse: &GetUserResponse{
                User: &User{
                    ID:    "01H4XJZQX3EXAMPLE",
                    Email: "test@example.com",
                    Name:  "Test User",
                },
            },
            mockError:      nil,
            expectedStatus: http.StatusOK,
        },
        {
            name:           "invalid user ID format",
            userID:         "invalid-id",
            mockResponse:   nil,
            mockError:      nil,
            expectedStatus: http.StatusBadRequest,
        },
        {
            name:           "user not found",
            userID:         "01H4XJZQX3NOTFOUND",
            mockResponse:   nil,
            mockError:      appErrors.NewNotFoundError("User not found"),
            expectedStatus: http.StatusNotFound,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mock expectations
            if tt.mockResponse != nil || tt.mockError != nil {
                mockService.On("GetUser", mock.Anything, mock.Anything).
                    Return(tt.mockResponse, tt.mockError)
            }
            
            // Create request
            req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/users/"+tt.userID, nil)
            w := httptest.NewRecorder()
            
            // Execute request
            router.ServeHTTP(w, req)
            
            // Assert results
            assert.Equal(t, tt.expectedStatus, w.Code)
            
            if tt.mockResponse != nil {
                var response GetUserResponse
                err := json.Unmarshal(w.Body.Bytes(), &response)
                assert.NoError(t, err)
                assert.Equal(t, tt.mockResponse.User.Email, response.User.Email)
            }
            
            // Clean up mocks
            mockService.AssertExpectations(t)
        })
    }
}
```

## Complete API Example

### User Management API

```go
package http

import (
    "strconv"
    
    "github.com/gin-gonic/gin"
    
    authDomain "brokle/internal/core/domain/auth"
    "brokle/pkg/response"
    "brokle/pkg/ulid"
    "brokle/pkg/apperrors"
    "brokle/internal/transport/http/middleware"
)

type UserHandler struct {
    userService authDomain.UserService
    authService authDomain.AuthService
}

func NewUserHandler(
    userService authDomain.UserService,
    authService authDomain.AuthService,
) *UserHandler {
    return &UserHandler{
        userService: userService,
        authService: authService,
    }
}

func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
    authenticated := r.Group("")
    authenticated.Use(middleware.Authentication())
    
    users := authenticated.Group("/users")
    {
        users.POST("", middleware.RequirePermission("users:create"), h.CreateUser)
        users.GET("", middleware.RequirePermission("users:read"), h.ListUsers)
        users.GET("/:id", middleware.RequirePermission("users:read"), h.GetUser)
        users.PUT("/:id", middleware.RequirePermission("users:write"), h.UpdateUser)
        users.DELETE("/:id", middleware.RequirePermission("users:delete"), h.DeleteUser)
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
// @Param        sort_by   query     string  false  "Sort field"        Enums(created_at, updated_at, name, email)
// @Param        sort_dir  query     string  false  "Sort direction"    Enums(asc, desc)
// @Success      200       {object}  ListUsersResponse
// @Failure      400       {object}  response.ErrorResponse
// @Failure      500       {object}  response.ErrorResponse
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
// @Param        id   path      string  true   "User ID"
// @Param        include_profile  query  boolean  false  "Include user profile data"
// @Success      200  {object}  GetUserResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /api/v1/auth/users/{id} [get]
// @Security     ApiKeyAuth
func (h *UserHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")
    id, err := ulid.Parse(userID)
    if err != nil {
        response.Error(c, appErrors.NewValidationError("Invalid user ID format", map[string]string{
            "id": "must be a valid ULID",
        }))
        return
    }

    includeProfile := c.Query("include_profile") == "true"
    
    resp, err := h.userService.GetUser(c.Request.Context(), &GetUserRequest{
        ID:             id,
        IncludeProfile: includeProfile,
    })
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
// @Param        request  body      UpdateUserRequest  true  "User update data"
// @Success      200      {object}  UpdateUserResponse
// @Failure      400      {object}  response.ErrorResponse
// @Failure      403      {object}  response.ErrorResponse
// @Failure      404      {object}  response.ErrorResponse
// @Failure      500      {object}  response.ErrorResponse
// @Router       /api/v1/auth/users/{id} [put]
// @Security     ApiKeyAuth
func (h *UserHandler) UpdateUser(c *gin.Context) {
    userID := c.Param("id")
    id, err := ulid.Parse(userID)
    if err != nil {
        response.Error(c, appErrors.NewValidationError("Invalid user ID format", map[string]string{
            "id": "must be a valid ULID",
        }))
        return
    }

    var req UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, appErrors.NewValidationError("Invalid request body", err))
        return
    }

    // Authorization check
    currentUser := middleware.GetCurrentUser(c)
    if currentUser.ID != id.String() && !currentUser.HasPermission("users:write:all") {
        response.Error(c, appErrors.NewForbiddenError("Can only update your own profile"))
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
// @Description  Delete a user account
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "User ID"
// @Success      204 "No content"
// @Failure      400 {object}  response.ErrorResponse
// @Failure      403 {object}  response.ErrorResponse
// @Failure      404 {object}  response.ErrorResponse
// @Failure      500 {object}  response.ErrorResponse
// @Router       /api/v1/auth/users/{id} [delete]
// @Security     ApiKeyAuth
func (h *UserHandler) DeleteUser(c *gin.Context) {
    userID := c.Param("id")
    id, err := ulid.Parse(userID)
    if err != nil {
        response.Error(c, appErrors.NewValidationError("Invalid user ID format", map[string]string{
            "id": "must be a valid ULID",
        }))
        return
    }

    err = h.userService.DeleteUser(c.Request.Context(), id)
    if err != nil {
        response.Error(c, err)
        return
    }

    response.NoContent(c)
}
```

## Best Practices Checklist

### Development Checklist
- [ ] Follow URL structure standards (`/api/v1/{domain}/{resource}`)
- [ ] Use professional domain aliases in imports
- [ ] Implement comprehensive input validation
- [ ] Use structured error handling with `response.Error()`
- [ ] Include complete Swagger documentation
- [ ] Apply appropriate authentication and authorization
- [ ] Write comprehensive tests
- [ ] Follow consistent request/response patterns

### Code Review Checklist
- [ ] All endpoints follow the same structural patterns
- [ ] Error handling uses AppError constructors consistently
- [ ] Input validation covers all edge cases
- [ ] Response structures are consistent and well-documented
- [ ] Authentication and authorization are properly implemented
- [ ] Swagger documentation is complete and accurate
- [ ] Tests cover success and error scenarios
- [ ] Performance considerations are addressed

### Security Checklist
- [ ] All sensitive endpoints require authentication
- [ ] Authorization checks are implemented where needed
- [ ] Input validation prevents injection attacks
- [ ] Rate limiting is applied where appropriate
- [ ] Sensitive data is not logged or exposed
- [ ] CORS settings are properly configured

## Conclusion

Following these API development patterns ensures:

- **Consistency**: All endpoints follow the same professional standards
- **Maintainability**: Clear structure and error handling make code easy to maintain
- **Developer Experience**: Comprehensive documentation and consistent patterns
- **Security**: Proper authentication, authorization, and validation
- **Scalability**: Clean architecture supports growth and changes
- **Quality**: Testing and validation ensure reliable APIs

This guide, combined with our [Error Handling Guide](./ERROR_HANDLING_GUIDE.md) and [Domain Alias Patterns](./DOMAIN_ALIAS_PATTERNS.md), provides everything needed to build professional, maintainable APIs for the Brokle platform.