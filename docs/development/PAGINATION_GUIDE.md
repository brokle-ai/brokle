# API Pagination Guide

Comprehensive guide for implementing consistent, efficient, and user-friendly pagination across all Brokle platform APIs.

## Table of Contents

- [Overview](#overview)
- [Pagination Standards](#pagination-standards)
- [Implementation Patterns](#implementation-patterns)
- [Request Parameters](#request-parameters)
- [Response Format](#response-format)
- [Database Implementation](#database-implementation)
- [Performance Optimization](#performance-optimization)
- [Advanced Patterns](#advanced-patterns)
- [Testing Pagination](#testing-pagination)
- [Examples](#examples)

## Overview

The Brokle platform uses **offset-based pagination** as the primary pattern, with **cursor-based pagination** for high-performance scenarios. All paginated endpoints follow consistent patterns for predictable developer experience.

### Why Consistent Pagination Matters

- **Developer Experience**: Predictable patterns across all APIs
- **Performance**: Efficient database queries and memory usage
- **Scalability**: Handles large datasets without performance degradation
- **User Interface**: Enables rich pagination controls in frontends

## Pagination Standards

### Default Pagination Rules

```go
// Standard pagination defaults
const (
    DefaultPage     = 1    // Start from page 1
    DefaultLimit    = 20   // 20 items per page
    MaxLimit        = 100  // Maximum items per page
    MinLimit        = 1    // Minimum items per page
)
```

### URL Parameter Standards

```
GET /api/v1/organizations?page=1&limit=20&sort_by=created_at&sort_dir=DESC
```

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `page` | integer | 1 | 1-∞ | Page number (1-based) |
| `limit` | integer | 20 | 1-100 | Items per page |
| `sort_by` | string | `created_at` | varies | Sort field |
| `sort_dir` | string | `DESC` | `ASC`, `DESC` | Sort direction |

## Implementation Patterns

### Request Structure

```go
// Standard paginated request structure
type PaginatedRequest struct {
    Page    int    `form:"page" binding:"omitempty,min=1" example:"1"`
    Limit   int    `form:"limit" binding:"omitempty,min=1,max=100" example:"20"`
    SortBy  string `form:"sort_by" binding:"omitempty" example:"created_at"`
    SortDir string `form:"sort_dir" binding:"omitempty,oneof=asc desc" example:"desc"`
}

// Domain-specific request extending base pagination
type ListUsersRequest struct {
    PaginatedRequest
    Search   string `form:"search" binding:"omitempty,max=100" example:"john"`
    Status   string `form:"status" binding:"omitempty,oneof=active inactive all" example:"active"`
    Role     string `form:"role" binding:"omitempty,oneof=admin user viewer" example:"user"`
    OrgID    string `form:"org_id" binding:"omitempty,ulid" example:"01H4XJZQX3EXAMPLE"`
}
```

### Response Structure (Meta-based)

**Actual API Response:**
```json
{
  "success": true,
  "data": [...],
  "meta": {
    "request_id": "01K54D9DX9KD4MMKXYNRK4CZS3",
    "timestamp": "2025-09-14T15:25:24Z",
    "version": "v1",
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 0,
      "total_page": 1,
      "has_next": false,
      "has_prev": false
    }
  }
}
```

**Go Structures:**
```go
// Standard API Response structure
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
    RequestID  string      `json:"request_id,omitempty"`
    Timestamp  string      `json:"timestamp,omitempty"`
    Version    string      `json:"version,omitempty"`
    Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
    Page      int   `json:"page"`
    PageSize  int   `json:"page_size"`
    Total     int64 `json:"total"`
    TotalPage int   `json:"total_page"`
    HasNext   bool  `json:"has_next"`
    HasPrev   bool  `json:"has_prev"`
}
```

### Handler Implementation

```go
// @Summary      List users with pagination
// @Description  Retrieve a paginated list of users with filtering and sorting
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "Page number"           minimum(1) default(1)
// @Param        limit     query     int     false  "Items per page"        minimum(1) maximum(100) default(20)
// @Param        sort_by   query     string  false  "Sort field"            Enums(created_at, updated_at, name, email) default(created_at)
// @Param        sort_dir  query     string  false  "Sort direction"        Enums(asc, desc) default(desc)
// @Param        search    query     string  false  "Search term"
// @Param        status    query     string  false  "Filter by status"      Enums(active, inactive, all) default(all)
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
    
    // Apply defaults
    req = h.applyPaginationDefaults(req)
    
    // Validate sort field
    if err := h.validateSortField(req.SortBy, []string{"created_at", "updated_at", "name", "email"}); err != nil {
        response.Error(c, err)
        return
    }
    
    resp, err := h.userService.ListUsers(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, err)
        return
    }
    
    response.Success(c, resp)
}

// Helper function to apply pagination defaults
func (h *UserHandler) applyPaginationDefaults(req ListUsersRequest) ListUsersRequest {
    if req.Page == 0 {
        req.Page = DefaultPage
    }
    if req.Limit == 0 {
        req.Limit = DefaultLimit
    }
    if req.SortBy == "" {
        req.SortBy = "created_at"
    }
    if req.SortDir == "" {
        req.SortDir = "desc"
    }
    return req
}

// Helper function to validate sort fields
func (h *UserHandler) validateSortField(sortBy string, allowedFields []string) error {
    for _, field := range allowedFields {
        if sortBy == field {
            return nil
        }
    }
    return appErrors.NewValidationError("Invalid sort field", map[string]string{
        "sort_by": fmt.Sprintf("must be one of: %s", strings.Join(allowedFields, ", ")),
    })
}
```

## Request Parameters

### Standard Parameters

```go
type PaginationParams struct {
    // Page number (1-based indexing)
    Page int `form:"page" binding:"omitempty,min=1" example:"1"`
    
    // Number of items per page
    Limit int `form:"limit" binding:"omitempty,min=1,max=100" example:"20"`
    
    // Field to sort by
    SortBy string `form:"sort_by" binding:"omitempty" example:"created_at"`
    
    // Sort direction (asc/desc)
    SortDir string `form:"sort_dir" binding:"omitempty,oneof=asc desc" example:"desc"`
}
```

### Filtering Parameters

```go
// Add domain-specific filters to pagination
type ListProjectsRequest struct {
    PaginationParams
    
    // Text search across multiple fields
    Search string `form:"search" binding:"omitempty,max=100" example:"my project"`
    
    // Status filtering
    Status string `form:"status" binding:"omitempty,oneof=active inactive archived all" example:"active"`
    
    // Organization filtering
    OrganizationID string `form:"org_id" binding:"omitempty,ulid" example:"01H4XJZQX3EXAMPLE"`
    
    // Date range filtering
    CreatedAfter  *time.Time `form:"created_after" time_format:"2006-01-02" example:"2023-01-01"`
    CreatedBefore *time.Time `form:"created_before" time_format:"2006-01-02" example:"2023-12-31"`
    
    // Tag filtering
    Tags []string `form:"tags" binding:"omitempty" example:"production,api"`
}
```

## Response Format

### Standard Pagination Response

```json
{
  "users": [
    {
      "id": "01H4XJZQX3EXAMPLE",
      "email": "user1@example.com",
      "name": "John Doe",
      "status": "active",
      "created_at": "2023-07-01T12:00:00Z"
    },
    {
      "id": "01H4XJZQX4EXAMPLE",
      "email": "user2@example.com", 
      "name": "Jane Smith",
      "status": "active",
      "created_at": "2023-07-01T11:30:00Z"
    }
  ],
  "pagination": {
    "page": 2,
    "limit": 20,
    "total": 156,
    "total_pages": 8,
    "has_next": true,
    "has_prev": true
  }
}
```

### Pagination Metadata Details

```go
type Pagination struct {
    // Current page number (1-based)
    Page int `json:"page" example:"2"`
    
    // Number of items per page
    Limit int `json:"limit" example:"20"`
    
    // Total number of items across all pages
    Total int `json:"total" example:"156"`
    
    // Total number of pages
    TotalPages int `json:"total_pages" example:"8"`
    
    // Whether there's a next page
    HasNext bool `json:"has_next" example:"true"`
    
    // Whether there's a previous page
    HasPrev bool `json:"has_prev" example:"true"`
    
    // Optional: Direct page navigation links
    FirstPage *int `json:"first_page,omitempty" example:"1"`
    LastPage  *int `json:"last_page,omitempty" example:"8"`
    NextPage  *int `json:"next_page,omitempty" example:"3"`
    PrevPage  *int `json:"prev_page,omitempty" example:"1"`
}
```

## Database Implementation

### Repository Layer Pattern

```go
// Repository interface for paginated queries
type UserRepository interface {
    List(ctx context.Context, req *ListUsersRequest) ([]*User, *PaginationResult, error)
    Count(ctx context.Context, filters *UserFilters) (int, error)
}

// Pagination result from repository layer
type PaginationResult struct {
    Total      int
    Page       int
    Limit      int
    TotalPages int
    HasNext    bool
    HasPrev    bool
}

// Repository implementation
func (r *userRepository) List(ctx context.Context, req *ListUsersRequest) ([]*User, *PaginationResult, error) {
    // Build base query
    query := r.db.WithContext(ctx).Model(&userDomain.User{})
    
    // Apply filters
    query = r.applyFilters(query, req)
    
    // Get total count (before applying limit/offset)
    var total int64
    if err := query.Count(&total).Error; err != nil {
        return nil, nil, fmt.Errorf("count users: %w", err)
    }
    
    // Apply sorting
    orderBy := fmt.Sprintf("%s %s", req.SortBy, strings.ToUpper(req.SortDir))
    query = query.Order(orderBy)
    
    // Apply pagination
    offset := (req.Page - 1) * req.Limit
    query = query.Offset(offset).Limit(req.Limit)
    
    // Execute query
    var users []*userDomain.User
    if err := query.Find(&users).Error; err != nil {
        return nil, nil, fmt.Errorf("find users: %w", err)
    }
    
    // Calculate pagination metadata
    totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))
    pagination := &PaginationResult{
        Total:      int(total),
        Page:       req.Page,
        Limit:      req.Limit,
        TotalPages: totalPages,
        HasNext:    req.Page < totalPages,
        HasPrev:    req.Page > 1,
    }
    
    return users, pagination, nil
}

// Apply filters helper
func (r *userRepository) applyFilters(query *gorm.DB, req *ListUsersRequest) *gorm.DB {
    // Text search
    if req.Search != "" {
        searchTerm := "%" + req.Search + "%"
        query = query.Where("name ILIKE ? OR email ILIKE ?", searchTerm, searchTerm)
    }
    
    // Status filtering
    if req.Status != "" && req.Status != "all" {
        query = query.Where("status = ?", req.Status)
    }
    
    // Organization filtering
    if req.OrganizationID != "" {
        query = query.Joins("JOIN organization_members ON users.id = organization_members.user_id").
            Where("organization_members.organization_id = ?", req.OrganizationID)
    }
    
    // Date range filtering
    if req.CreatedAfter != nil {
        query = query.Where("created_at >= ?", req.CreatedAfter)
    }
    if req.CreatedBefore != nil {
        query = query.Where("created_at <= ?", req.CreatedBefore)
    }
    
    return query
}
```

### Service Layer Implementation

```go
func (s *userService) ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
    // Validate request
    if err := s.validateListRequest(req); err != nil {
        return nil, err
    }
    
    // Get data from repository
    users, paginationResult, err := s.userRepo.List(ctx, req)
    if err != nil {
        if errors.Is(err, userDomain.ErrNotFound) {
            // Return empty result for not found
            return &ListUsersResponse{
                Users: []*User{},
                Pagination: &Pagination{
                    Page:       req.Page,
                    Limit:      req.Limit,
                    Total:      0,
                    TotalPages: 0,
                    HasNext:    false,
                    HasPrev:    false,
                },
            }, nil
        }
        return nil, appErrors.NewInternalError("Failed to retrieve users")
    }
    
    // Convert domain models to response models
    responseUsers := make([]*User, len(users))
    for i, user := range users {
        responseUsers[i] = s.userToResponse(user)
    }
    
    // Build pagination response
    pagination := &Pagination{
        Page:       paginationResult.Page,
        Limit:      paginationResult.Limit,
        Total:      paginationResult.Total,
        TotalPages: paginationResult.TotalPages,
        HasNext:    paginationResult.HasNext,
        HasPrev:    paginationResult.HasPrev,
    }
    
    return &ListUsersResponse{
        Users:      responseUsers,
        Pagination: pagination,
    }, nil
}
```

## Performance Optimization

### Database Query Optimization

```sql
-- Ensure proper indexes for pagination queries
CREATE INDEX CONCURRENTLY idx_users_pagination ON users(created_at DESC, id);
CREATE INDEX CONCURRENTLY idx_users_status_pagination ON users(status, created_at DESC, id);
CREATE INDEX CONCURRENTLY idx_users_search ON users USING gin(to_tsvector('english', name || ' ' || email));
```

### Efficient Counting

```go
// Use separate count query for better performance
func (r *userRepository) ListWithCount(ctx context.Context, req *ListUsersRequest) ([]*User, int, error) {
    // Build base query
    baseQuery := r.db.WithContext(ctx).Model(&userDomain.User{})
    baseQuery = r.applyFilters(baseQuery, req)
    
    // Execute count and data queries in parallel
    var (
        users []*userDomain.User
        total int64
        wg    sync.WaitGroup
        errCh = make(chan error, 2)
    )
    
    // Count query
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := baseQuery.Count(&total).Error; err != nil {
            errCh <- fmt.Errorf("count users: %w", err)
        }
    }()
    
    // Data query
    wg.Add(1)
    go func() {
        defer wg.Done()
        dataQuery := baseQuery.
            Order(fmt.Sprintf("%s %s", req.SortBy, strings.ToUpper(req.SortDir))).
            Offset((req.Page - 1) * req.Limit).
            Limit(req.Limit)
            
        if err := dataQuery.Find(&users).Error; err != nil {
            errCh <- fmt.Errorf("find users: %w", err)
        }
    }()
    
    wg.Wait()
    close(errCh)
    
    // Check for errors
    for err := range errCh {
        if err != nil {
            return nil, 0, err
        }
    }
    
    return users, int(total), nil
}
```

### Caching Strategies

```go
// Cache pagination results for expensive queries
func (s *userService) ListUsersWithCache(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
    // Generate cache key
    cacheKey := fmt.Sprintf("users:list:%s", s.generateCacheKey(req))
    
    // Try to get from cache
    var cachedResponse ListUsersResponse
    if err := s.cache.Get(ctx, cacheKey, &cachedResponse); err == nil {
        return &cachedResponse, nil
    }
    
    // Get from database
    response, err := s.ListUsers(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // Cache the result (5 minutes for list queries)
    _ = s.cache.Set(ctx, cacheKey, response, 5*time.Minute)
    
    return response, nil
}
```

## Advanced Patterns

### Cursor-Based Pagination (High Performance)

```go
// For high-performance scenarios with large datasets
type CursorPaginationRequest struct {
    Cursor string `form:"cursor" example:"01H4XJZQX3EXAMPLE"`
    Limit  int    `form:"limit" binding:"omitempty,min=1,max=100" example:"20"`
}

type CursorPaginationResponse struct {
    Data       []*User `json:"data"`
    NextCursor *string `json:"next_cursor,omitempty"`
    HasMore    bool    `json:"has_more"`
}

// Repository implementation for cursor pagination
func (r *userRepository) ListWithCursor(ctx context.Context, cursor string, limit int) ([]*User, string, bool, error) {
    query := r.db.WithContext(ctx).Model(&userDomain.User{}).Order("id ASC")
    
    // Apply cursor filter
    if cursor != "" {
        query = query.Where("id > ?", cursor)
    }
    
    // Get one extra record to check if there are more
    query = query.Limit(limit + 1)
    
    var users []*userDomain.User
    if err := query.Find(&users).Error; err != nil {
        return nil, "", false, err
    }
    
    // Check if there are more records
    hasMore := len(users) > limit
    if hasMore {
        users = users[:limit] // Remove the extra record
    }
    
    // Generate next cursor
    var nextCursor string
    if hasMore && len(users) > 0 {
        nextCursor = users[len(users)-1].ID.String()
    }
    
    return users, nextCursor, hasMore, nil
}
```

### Search with Pagination

```go
// Full-text search with pagination
type SearchUsersRequest struct {
    PaginatedRequest
    Query      string   `form:"q" binding:"required,min=2" example:"john doe"`
    SearchIn   []string `form:"search_in" example:"name,email"`
    Highlight  bool     `form:"highlight" example:"true"`
}

type SearchUsersResponse struct {
    Users       []*UserSearchResult `json:"users"`
    Pagination  *Pagination         `json:"pagination"`
    SearchMeta  *SearchMetadata     `json:"search_meta"`
}

type UserSearchResult struct {
    *User
    Score       float64            `json:"score,omitempty"`
    Highlights  map[string]string  `json:"highlights,omitempty"`
}

type SearchMetadata struct {
    Query       string  `json:"query"`
    SearchTime  string  `json:"search_time"`
    TotalHits   int     `json:"total_hits"`
}
```

### Aggregated Pagination

```go
// Pagination with aggregated data
type ListUsersWithStatsResponse struct {
    Users      []*User            `json:"users"`
    Pagination *Pagination        `json:"pagination"`
    Stats      *UserListStats     `json:"stats"`
}

type UserListStats struct {
    TotalActive   int `json:"total_active"`
    TotalInactive int `json:"total_inactive"`
    NewThisWeek   int `json:"new_this_week"`
    NewThisMonth  int `json:"new_this_month"`
}

// Service implementation with aggregated stats
func (s *userService) ListUsersWithStats(ctx context.Context, req *ListUsersRequest) (*ListUsersWithStatsResponse, error) {
    // Get paginated users
    usersResponse, err := s.ListUsers(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // Get aggregated stats
    stats, err := s.getUserStats(ctx, req)
    if err != nil {
        return nil, err
    }
    
    return &ListUsersWithStatsResponse{
        Users:      usersResponse.Users,
        Pagination: usersResponse.Pagination,
        Stats:      stats,
    }, nil
}
```

## Testing Pagination

### Unit Tests

```go
func TestUserHandler_ListUsers_Pagination(t *testing.T) {
    tests := []struct {
        name           string
        queryParams    string
        mockUsers      []*User
        mockTotal      int
        expectedPage   int
        expectedLimit  int
        expectedTotal  int
        expectedHasNext bool
        expectedHasPrev bool
    }{
        {
            name:           "first page with default limit",
            queryParams:    "",
            mockUsers:      createMockUsers(20),
            mockTotal:      100,
            expectedPage:   1,
            expectedLimit:  20,
            expectedTotal:  100,
            expectedHasNext: true,
            expectedHasPrev: false,
        },
        {
            name:           "middle page",
            queryParams:    "page=3&limit=10",
            mockUsers:      createMockUsers(10),
            mockTotal:      100,
            expectedPage:   3,
            expectedLimit:  10,
            expectedTotal:  100,
            expectedHasNext: true,
            expectedHasPrev: true,
        },
        {
            name:           "last page",
            queryParams:    "page=10&limit=10",
            mockUsers:      createMockUsers(10),
            mockTotal:      100,
            expectedPage:   10,
            expectedLimit:  10,
            expectedTotal:  100,
            expectedHasNext: false,
            expectedHasPrev: true,
        },
        {
            name:           "invalid page number defaults to 1",
            queryParams:    "page=0",
            mockUsers:      createMockUsers(20),
            mockTotal:      100,
            expectedPage:   1,
            expectedLimit:  20,
            expectedTotal:  100,
            expectedHasNext: true,
            expectedHasPrev: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockService := &MockUserService{}
            handler := NewUserHandler(mockService, nil)
            
            // Mock expectations
            mockService.On("ListUsers", mock.Anything, mock.MatchedBy(func(req *ListUsersRequest) bool {
                return req.Page == tt.expectedPage && req.Limit == tt.expectedLimit
            })).Return(&ListUsersResponse{
                Users: tt.mockUsers,
                Pagination: &Pagination{
                    Page:       tt.expectedPage,
                    Limit:      tt.expectedLimit,
                    Total:      tt.expectedTotal,
                    TotalPages: int(math.Ceil(float64(tt.expectedTotal) / float64(tt.expectedLimit))),
                    HasNext:    tt.expectedHasNext,
                    HasPrev:    tt.expectedHasPrev,
                },
            }, nil)
            
            // Execute request
            req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/users?"+tt.queryParams, nil)
            w := httptest.NewRecorder()
            
            router := gin.New()
            handler.RegisterRoutes(router.Group("/api/v1/auth"))
            router.ServeHTTP(w, req)
            
            // Assert response
            assert.Equal(t, http.StatusOK, w.Code)
            
            var response ListUsersResponse
            err := json.Unmarshal(w.Body.Bytes(), &response)
            assert.NoError(t, err)
            
            assert.Equal(t, tt.expectedPage, response.Pagination.Page)
            assert.Equal(t, tt.expectedLimit, response.Pagination.Limit)
            assert.Equal(t, tt.expectedTotal, response.Pagination.Total)
            assert.Equal(t, tt.expectedHasNext, response.Pagination.HasNext)
            assert.Equal(t, tt.expectedHasPrev, response.Pagination.HasPrev)
            
            mockService.AssertExpectations(t)
        })
    }
}
```

### Performance Tests

```go
func TestUserRepository_List_Performance(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance test in short mode")
    }
    
    // Setup test database with large dataset
    db := setupTestDB(t)
    repo := NewUserRepository(db)
    
    // Create test data
    createTestUsers(t, db, 10000) // 10k users
    
    tests := []struct {
        name     string
        page     int
        limit    int
        maxTime  time.Duration
    }{
        {"first page small limit", 1, 20, 100 * time.Millisecond},
        {"middle page small limit", 250, 20, 100 * time.Millisecond},
        {"last page small limit", 500, 20, 100 * time.Millisecond},
        {"first page large limit", 1, 100, 200 * time.Millisecond},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := &ListUsersRequest{
                PaginatedRequest: PaginatedRequest{
                    Page:  tt.page,
                    Limit: tt.limit,
                },
            }
            
            start := time.Now()
            users, pagination, err := repo.List(context.Background(), req)
            duration := time.Since(start)
            
            assert.NoError(t, err)
            assert.NotNil(t, users)
            assert.NotNil(t, pagination)
            assert.True(t, duration < tt.maxTime, 
                "Query took %v, expected less than %v", duration, tt.maxTime)
        })
    }
}
```

## Examples

### Complete Pagination Implementation

```go
// Complete example: Organization Projects API
package http

type ProjectHandler struct {
    projectService orgDomain.ProjectService
}

// @Summary      List organization projects
// @Description  Retrieve paginated list of projects for an organization
// @Tags         Projects
// @Param        org_id    path      string  true   "Organization ID"
// @Param        page      query     int     false  "Page number"           minimum(1)
// @Param        limit     query     int     false  "Items per page"        minimum(1) maximum(100)
// @Param        search    query     string  false  "Search projects"
// @Param        status    query     string  false  "Filter by status"      Enums(active,inactive,archived,all)
// @Param        sort_by   query     string  false  "Sort field"           Enums(created_at,updated_at,name)
// @Param        sort_dir  query     string  false  "Sort direction"       Enums(asc,desc)
// @Success      200       {object}  ListProjectsResponse
// @Router       /api/v1/organizations/{org_id}/projects [get]
func (h *ProjectHandler) ListProjects(c *gin.Context) {
    // Extract organization ID
    orgID := c.Param("org_id")
    orgULID, err := ulid.Parse(orgID)
    if err != nil {
        response.Error(c, appErrors.NewValidationError("Invalid organization ID", map[string]string{
            "org_id": "must be a valid ULID",
        }))
        return
    }
    
    // Parse query parameters
    var req ListProjectsRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        response.Error(c, appErrors.NewValidationError("Invalid query parameters", err))
        return
    }
    
    // Set organization ID and defaults
    req.OrganizationID = orgULID
    req = h.applyDefaults(req)
    
    // Validate sort field
    allowedSortFields := []string{"created_at", "updated_at", "name", "status"}
    if err := h.validateSortField(req.SortBy, allowedSortFields); err != nil {
        response.Error(c, err)
        return
    }
    
    // Call service
    resp, err := h.projectService.ListProjects(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, err)
        return
    }
    
    response.Success(c, resp)
}

func (h *ProjectHandler) applyDefaults(req ListProjectsRequest) ListProjectsRequest {
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
    if req.Status == "" {
        req.Status = "all"
    }
    return req
}
```

## Best Practices Summary

### Development Guidelines
- ✅ Always use 1-based page numbering
- ✅ Set reasonable default and maximum limits
- ✅ Include comprehensive pagination metadata
- ✅ Validate all pagination parameters
- ✅ Provide meaningful sort options
- ✅ Implement efficient database queries with proper indexes
- ✅ Test pagination edge cases (empty results, last page, etc.)

### Performance Guidelines  
- ✅ Use separate count queries for large datasets
- ✅ Implement cursor-based pagination for high-performance scenarios
- ✅ Add database indexes for common sort and filter combinations
- ✅ Consider caching for expensive pagination queries
- ✅ Monitor query performance and optimize as needed

### User Experience Guidelines
- ✅ Provide clear navigation information (has_next, has_prev)
- ✅ Include total counts for UI pagination controls
- ✅ Support flexible filtering and sorting options
- ✅ Return empty results gracefully (don't error on page beyond range)
- ✅ Maintain consistent response format across all paginated endpoints

This comprehensive pagination guide ensures consistent, performant, and user-friendly pagination across all Brokle platform APIs.