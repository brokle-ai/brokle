package auth

import (
	"context"
	"time"

	"brokle/pkg/ulid"
)

// AuthService defines the core authentication service interface.
type AuthService interface {
	// Authentication
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	Register(ctx context.Context, req *RegisterRequest) (*LoginResponse, error)
	Logout(ctx context.Context, jti string, userID ulid.ULID) error
	RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*LoginResponse, error)
	
	// Password management
	ChangePassword(ctx context.Context, userID ulid.ULID, currentPassword, newPassword string) error
	ResetPassword(ctx context.Context, email string) error
	ConfirmPasswordReset(ctx context.Context, token, newPassword string) error
	
	// Email verification
	SendEmailVerification(ctx context.Context, userID ulid.ULID) error
	VerifyEmail(ctx context.Context, token string) error
	
	// Session management
	GetUserSessions(ctx context.Context, userID ulid.ULID) ([]*UserSession, error)
	RevokeSession(ctx context.Context, userID, sessionID ulid.ULID) error
	RevokeAllSessions(ctx context.Context, userID ulid.ULID) error
	
	// Token revocation (immediate)
	RevokeAccessToken(ctx context.Context, jti string, userID ulid.ULID, reason string) error
	RevokeUserAccessTokens(ctx context.Context, userID ulid.ULID, reason string) error
	IsTokenRevoked(ctx context.Context, jti string) (bool, error)
	
	// Authentication context
	GetAuthContext(ctx context.Context, token string) (*AuthContext, error)
	ValidateAuthToken(ctx context.Context, token string) (*AuthContext, error)
}

// SessionService defines the session management service interface.
type SessionService interface {
	// Session management
	GetSession(ctx context.Context, sessionID ulid.ULID) (*UserSession, error)
	RevokeSession(ctx context.Context, sessionID ulid.ULID) error
	
	// User session management
	GetUserSessions(ctx context.Context, userID ulid.ULID) ([]*UserSession, error)
	RevokeUserSessions(ctx context.Context, userID ulid.ULID) error
	
	// Session cleanup and maintenance
	CleanupExpiredSessions(ctx context.Context) error
	GetActiveSessions(ctx context.Context, userID ulid.ULID) ([]*UserSession, error)
}

// APIKeyService defines the API key management service interface.
type APIKeyService interface {
	// API key management
	CreateAPIKey(ctx context.Context, userID ulid.ULID, req *CreateAPIKeyRequest) (*CreateAPIKeyResponse, error)
	GetAPIKey(ctx context.Context, keyID ulid.ULID) (*APIKey, error)
	GetAPIKeys(ctx context.Context, filters *APIKeyFilters) ([]*APIKey, error)
	UpdateAPIKey(ctx context.Context, keyID ulid.ULID, req *UpdateAPIKeyRequest) error
	RevokeAPIKey(ctx context.Context, keyID ulid.ULID) error
	
	// API key validation and usage
	ValidateAPIKey(ctx context.Context, keyHash string) (*APIKey, error)
	UpdateLastUsed(ctx context.Context, keyID ulid.ULID) error
	CheckRateLimit(ctx context.Context, keyID ulid.ULID) (bool, error)
	
	// API key context and permissions
	GetAPIKeyContext(ctx context.Context, keyID ulid.ULID) (*AuthContext, error)
	CanAPIKeyAccessResource(ctx context.Context, keyID ulid.ULID, resource string) (bool, error)
	
	// API key scoping
	GetAPIKeysByUser(ctx context.Context, userID ulid.ULID) ([]*APIKey, error)
	GetAPIKeysByOrganization(ctx context.Context, orgID ulid.ULID) ([]*APIKey, error)
	GetAPIKeysByProject(ctx context.Context, projectID ulid.ULID) ([]*APIKey, error)
	GetAPIKeysByEnvironment(ctx context.Context, envID ulid.ULID) ([]*APIKey, error)
}

// RoleService defines the role and permission management service interface.
type RoleService interface {
	// Role management
	CreateRole(ctx context.Context, orgID *ulid.ULID, req *CreateRoleRequest) (*Role, error)
	GetRole(ctx context.Context, roleID ulid.ULID) (*Role, error)
	GetRoleByName(ctx context.Context, orgID *ulid.ULID, name string) (*Role, error)
	UpdateRole(ctx context.Context, roleID ulid.ULID, req *UpdateRoleRequest) error
	DeleteRole(ctx context.Context, roleID ulid.ULID) error
	
	// Role queries
	ListRoles(ctx context.Context, orgID *ulid.ULID, limit, offset int) (*RoleListResponse, error) // Updated for pagination
	GetSystemRoles(ctx context.Context) ([]*Role, error)                                           // Global system roles only
	GetGlobalSystemRole(ctx context.Context, name string) (*Role, error)                          // Get specific global system role
	GetOrganizationRoles(ctx context.Context, orgID ulid.ULID) ([]*Role, error)                  // Org-specific roles only
	GetAvailableRoles(ctx context.Context, orgID ulid.ULID) ([]*Role, error)                     // System + org roles for assignment
	SearchRoles(ctx context.Context, orgID *ulid.ULID, query string, limit, offset int) (*RoleListResponse, error)
	
	// User role management
	GetUserRole(ctx context.Context, userID, orgID ulid.ULID) (*Role, error)
	AssignUserRole(ctx context.Context, userID, orgID, roleID ulid.ULID) error
	RevokeUserRole(ctx context.Context, userID, orgID ulid.ULID) error
	
	// Permission management
	GetRolePermissions(ctx context.Context, roleID ulid.ULID) ([]*Permission, error)
	AssignPermissionsToRole(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error
	RemovePermissionsFromRole(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error
	UpdateRolePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error // Replace all permissions
	
	// Permission checking (main authorization methods)
	HasPermission(ctx context.Context, userID, orgID ulid.ULID, resourceAction string) (bool, error)        // resource:action format
	HasResourceAction(ctx context.Context, userID, orgID ulid.ULID, resource, action string) (bool, error) // Separate resource/action
	HasPermissions(ctx context.Context, userID, orgID ulid.ULID, resourceActions []string) (bool, error)   // All must be true
	HasAnyPermission(ctx context.Context, userID, orgID ulid.ULID, resourceActions []string) (bool, error) // Any must be true
	CheckPermissions(ctx context.Context, userID, orgID ulid.ULID, resourceActions []string) (*CheckPermissionsResponse, error)
	
	// User permissions (through roles via organization membership)
	GetUserPermissions(ctx context.Context, userID, orgID ulid.ULID) (*UserPermissionsResponse, error) // Full response with role info
	GetUserPermissionStrings(ctx context.Context, userID, orgID ulid.ULID) ([]string, error)         // Just resource:action strings
	
	// Role validation and statistics
	ValidateRole(ctx context.Context, roleID ulid.ULID) error
	CanDeleteRole(ctx context.Context, roleID ulid.ULID) (bool, error)
	GetRoleStatistics(ctx context.Context, orgID ulid.ULID) (*RoleStatistics, error)
	
	// Legacy permission support (for backward compatibility)
	HasLegacyPermission(ctx context.Context, userID, orgID ulid.ULID, permission string) (bool, error) // Old dot notation
}

// PermissionService defines the permission management service interface.
type PermissionService interface {
	// Permission management
	CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*Permission, error)
	GetPermission(ctx context.Context, permissionID ulid.ULID) (*Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*Permission, error)                                    // Legacy name lookup
	GetPermissionByResourceAction(ctx context.Context, resource, action string) (*Permission, error)             // New resource:action lookup
	UpdatePermission(ctx context.Context, permissionID ulid.ULID, req *UpdatePermissionRequest) error
	DeletePermission(ctx context.Context, permissionID ulid.ULID) error
	
	// Permission queries
	ListPermissions(ctx context.Context, limit, offset int) (*PermissionListResponse, error)                     // Paginated list
	GetAllPermissions(ctx context.Context) ([]*Permission, error)
	GetPermissionsByCategory(ctx context.Context, category string) ([]*Permission, error)
	GetPermissionsByResource(ctx context.Context, resource string) ([]*Permission, error)                       // All permissions for resource
	GetPermissionsByNames(ctx context.Context, names []string) ([]*Permission, error)                           // Legacy bulk lookup
	GetPermissionsByResourceActions(ctx context.Context, resourceActions []string) ([]*Permission, error)      // New bulk resource:action lookup
	SearchPermissions(ctx context.Context, query string, limit, offset int) (*PermissionListResponse, error)
	
	// Resource and action queries  
	GetAvailableResources(ctx context.Context) ([]string, error)                                                 // Get all distinct resources
	GetActionsForResource(ctx context.Context, resource string) ([]string, error)                               // Get all actions for resource
	GetPermissionCategories(ctx context.Context) ([]string, error)                                              // Get all distinct categories
	
	// Permission validation
	ValidatePermissionName(ctx context.Context, name string) error                                               // Legacy name validation
	ValidateResourceAction(ctx context.Context, resource, action string) error                                  // New resource:action validation
	PermissionExists(ctx context.Context, resource, action string) (bool, error)
	BulkPermissionExists(ctx context.Context, resourceActions []string) (map[string]bool, error)
	
	// Utility methods
	ParseResourceAction(resourceAction string) (resource, action string, err error)
	FormatResourceAction(resource, action string) string
	IsValidResourceActionFormat(resourceAction string) bool
}

// JWTService defines the JWT token management service interface.
type JWTService interface {
	// Token generation
	GenerateAccessToken(ctx context.Context, userID ulid.ULID, claims map[string]interface{}) (string, error)
	GenerateAccessTokenWithJTI(ctx context.Context, userID ulid.ULID, claims map[string]interface{}) (string, string, error)
	GenerateRefreshToken(ctx context.Context, userID ulid.ULID) (string, error)
	GenerateAPIKeyToken(ctx context.Context, keyID ulid.ULID, scopes []string) (string, error)
	
	// Token validation
	ValidateAccessToken(ctx context.Context, token string) (*JWTClaims, error)
	ValidateRefreshToken(ctx context.Context, token string) (*JWTClaims, error)
	ValidateAPIKeyToken(ctx context.Context, token string) (*JWTClaims, error)
	
	// Token utilities
	GetTokenExpiry(ctx context.Context, token string) (time.Time, error)
	IsTokenExpired(ctx context.Context, token string) (bool, error)
}

// BlacklistedTokenService defines the token blacklisting service interface.
type BlacklistedTokenService interface {
	// Token blacklisting
	BlacklistToken(ctx context.Context, jti string, userID ulid.ULID, expiresAt time.Time, reason string) error
	IsTokenBlacklisted(ctx context.Context, jti string) (bool, error)
	GetBlacklistedToken(ctx context.Context, jti string) (*BlacklistedToken, error)
	
	// User-wide timestamp blacklisting (GDPR/SOC2 compliance)
	CreateUserTimestampBlacklist(ctx context.Context, userID ulid.ULID, reason string) error
	IsUserBlacklistedAfterTimestamp(ctx context.Context, userID ulid.ULID, tokenIssuedAt int64) (bool, error)
	GetUserBlacklistTimestamp(ctx context.Context, userID ulid.ULID) (*int64, error)
	
	// Bulk operations
	BlacklistUserTokens(ctx context.Context, userID ulid.ULID, reason string) error
	GetUserBlacklistedTokens(ctx context.Context, userID ulid.ULID, limit, offset int) ([]*BlacklistedToken, error)
	
	// Maintenance
	CleanupExpiredTokens(ctx context.Context) error
	CleanupOldTokens(ctx context.Context, olderThan time.Time) error
	
	// Statistics
	GetBlacklistedTokensCount(ctx context.Context) (int64, error)
	GetTokensByReason(ctx context.Context, reason string) ([]*BlacklistedToken, error)
}

// AuditLogService defines the audit logging service interface.
type AuditLogService interface {
	// Audit logging
	LogUserAction(ctx context.Context, userID *ulid.ULID, action, resource, resourceID string, metadata map[string]interface{}, ipAddress, userAgent string) error
	LogSystemAction(ctx context.Context, action, resource, resourceID string, metadata map[string]interface{}) error
	LogSecurityEvent(ctx context.Context, userID *ulid.ULID, event, description string, metadata map[string]interface{}, ipAddress, userAgent string) error
	
	// Audit log queries
	GetUserAuditLogs(ctx context.Context, userID ulid.ULID, limit, offset int) ([]*AuditLog, error)
	GetOrganizationAuditLogs(ctx context.Context, orgID ulid.ULID, limit, offset int) ([]*AuditLog, error)
	GetResourceAuditLogs(ctx context.Context, resource, resourceID string, limit, offset int) ([]*AuditLog, error)
	SearchAuditLogs(ctx context.Context, filters *AuditLogFilters) ([]*AuditLog, int, error)
	
	// Audit log maintenance
	CleanupOldAuditLogs(ctx context.Context, olderThan time.Time) error
	GetAuditLogStats(ctx context.Context) (*AuditLogStats, error)
}

// TokenClaims represents JWT token claims.
type TokenClaims struct {
	UserID         ulid.ULID  `json:"user_id"`
	Email          string     `json:"email"`
	OrganizationID *ulid.ULID `json:"organization_id,omitempty"`
	Scopes         []string   `json:"scopes,omitempty"`
	TokenType      string     `json:"token_type"` // access, refresh, api_key
	IssuedAt       int64      `json:"iat"`
	ExpiresAt      int64      `json:"exp"`
	NotBefore      int64      `json:"nbf"`
	Issuer         string     `json:"iss"`
	Subject        string     `json:"sub"`
}

// Request/Response DTOs
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
	Password  string `json:"password" validate:"required,min=8"`
	Timezone  string `json:"timezone,omitempty"`
	Language  string `json:"language,omitempty"`
}

type UpdateProfileRequest struct {
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,url"`
	Phone     *string `json:"phone,omitempty" validate:"omitempty,max=50"`
	Timezone  *string `json:"timezone,omitempty"`
	Language  *string `json:"language,omitempty" validate:"omitempty,len=2"`
}

type CreateSessionRequest struct {
	IPAddress *string `json:"ip_address,omitempty"`
	UserAgent *string `json:"user_agent,omitempty"`
	Remember  bool    `json:"remember"` // Extend session duration
}


type UpdateAPIKeyRequest struct {
	Name         *string  `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Scopes       []string `json:"scopes,omitempty"`
	RateLimitRPM *int     `json:"rate_limit_rpm,omitempty" validate:"omitempty,min=1,max=10000"`
	IsActive     *bool    `json:"is_active,omitempty"`
}

// Filter types
type APIKeyFilters struct {
	// Pagination
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	
	// Filters
	UserID         *ulid.ULID `json:"user_id,omitempty"`
	OrganizationID *ulid.ULID `json:"organization_id,omitempty"`
	ProjectID      *ulid.ULID `json:"project_id,omitempty"`
	EnvironmentID  *ulid.ULID `json:"environment_id,omitempty"`
	IsActive       *bool      `json:"is_active,omitempty"`
	IsExpired      *bool      `json:"is_expired,omitempty"`
	
	// Sorting
	SortBy    string `json:"sort_by"`    // name, created_at, last_used_at
	SortOrder string `json:"sort_order"` // asc, desc
}

// Statistics types - AuditLogStats is defined in repository.go

// AuthServices aggregates all authentication-related services.
type AuthServices interface {
	Auth() AuthService
	Sessions() SessionService
	APIKeys() APIKeyService
	Roles() RoleService
	Permissions() PermissionService
	JWT() JWTService
	BlacklistedTokens() BlacklistedTokenService
	AuditLogs() AuditLogService
}