package auth

import (
	"context"
	"time"

	"brokle/pkg/pagination"
	"brokle/pkg/ulid"
)

// AuthService defines the core authentication service interface.
type AuthService interface {
	// Authentication
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	GenerateTokensForUser(ctx context.Context, userID ulid.ULID) (*LoginResponse, error) // Generate tokens without password validation
	Logout(ctx context.Context, jti string, userID ulid.ULID) error
	RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*LoginResponse, error)

	// OAuth session management (for two-step OAuth signup)
	CreateOAuthSession(ctx context.Context, session interface{}) (string, error)
	GetOAuthSession(ctx context.Context, sessionID string) (interface{}, error)
	DeleteOAuthSession(ctx context.Context, sessionID string) error

	// OAuth login token sessions (for existing user OAuth login)
	CreateLoginTokenSession(ctx context.Context, accessToken, refreshToken string, expiresIn int64, userID ulid.ULID) (string, error)
	GetLoginTokenSession(ctx context.Context, sessionID string) (map[string]interface{}, error)

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
	ValidateAPIKey(ctx context.Context, fullKey string) (*ValidateAPIKeyResponse, error)
	CheckRateLimit(ctx context.Context, keyID ulid.ULID) (bool, error)

	// API key context and permissions
	GetAPIKeyContext(ctx context.Context, keyID ulid.ULID) (*AuthContext, error)
	CanAPIKeyAccessResource(ctx context.Context, keyID ulid.ULID, resource string) (bool, error)

	// API key scoping
	GetAPIKeysByUser(ctx context.Context, userID ulid.ULID) ([]*APIKey, error)
	GetAPIKeysByOrganization(ctx context.Context, orgID ulid.ULID) ([]*APIKey, error)
	GetAPIKeysByProject(ctx context.Context, projectID ulid.ULID) ([]*APIKey, error)

	// Pagination support
	CountAPIKeys(ctx context.Context, filters *APIKeyFilters) (int64, error)
}

// RoleService defines both system template and custom scoped role management service interface.
type RoleService interface {
	// System template role management
	CreateRole(ctx context.Context, req *CreateRoleRequest) (*Role, error)
	GetRoleByID(ctx context.Context, roleID ulid.ULID) (*Role, error)
	GetRoleByNameAndScope(ctx context.Context, name, scopeType string) (*Role, error)
	UpdateRole(ctx context.Context, roleID ulid.ULID, req *UpdateRoleRequest) (*Role, error)
	DeleteRole(ctx context.Context, roleID ulid.ULID) error

	// System template role queries
	GetRolesByScopeType(ctx context.Context, scopeType string) ([]*Role, error)
	GetAllRoles(ctx context.Context) ([]*Role, error)
	GetSystemRoles(ctx context.Context) ([]*Role, error)

	// Custom scoped role management
	CreateCustomRole(ctx context.Context, scopeType string, scopeID ulid.ULID, req *CreateRoleRequest) (*Role, error)
	GetCustomRolesByOrganization(ctx context.Context, organizationID ulid.ULID) ([]*Role, error)
	UpdateCustomRole(ctx context.Context, roleID ulid.ULID, req *UpdateRoleRequest) (*Role, error)
	DeleteCustomRole(ctx context.Context, roleID ulid.ULID) error

	// Permission management for roles
	GetRolePermissions(ctx context.Context, roleID ulid.ULID) ([]*Permission, error)
	AssignRolePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID, grantedBy *ulid.ULID) error
	RevokeRolePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error

	// Statistics
	GetRoleStatistics(ctx context.Context) (*RoleStatistics, error)
}

// OrganizationMemberService defines the organization membership management service interface.
type OrganizationMemberService interface {
	// Membership management
	AddMember(ctx context.Context, userID, orgID, roleID ulid.ULID, invitedBy *ulid.ULID) (*OrganizationMember, error)
	RemoveMember(ctx context.Context, userID, orgID ulid.ULID) error
	UpdateMemberRole(ctx context.Context, userID, orgID, roleID ulid.ULID) error

	// Membership queries
	GetMember(ctx context.Context, userID, orgID ulid.ULID) (*OrganizationMember, error)
	GetUserMemberships(ctx context.Context, userID ulid.ULID) ([]*OrganizationMember, error)
	GetOrganizationMembers(ctx context.Context, orgID ulid.ULID) ([]*OrganizationMember, error)
	GetMembersByRole(ctx context.Context, roleID ulid.ULID) ([]*OrganizationMember, error)
	IsMember(ctx context.Context, userID, orgID ulid.ULID) (bool, error)

	// Permission checking via membership
	GetUserEffectivePermissions(ctx context.Context, userID ulid.ULID) ([]string, error)
	GetUserPermissionsInOrganization(ctx context.Context, userID, orgID ulid.ULID) ([]string, error)
	CheckUserPermission(ctx context.Context, userID ulid.ULID, permission string) (bool, error)
	CheckUserPermissions(ctx context.Context, userID ulid.ULID, permissions []string) (map[string]bool, error)

	// Status management
	ActivateMember(ctx context.Context, userID, orgID ulid.ULID) error
	SuspendMember(ctx context.Context, userID, orgID ulid.ULID) error
	GetActiveMembers(ctx context.Context, orgID ulid.ULID) ([]*OrganizationMember, error)

	// Statistics
	GetMemberCount(ctx context.Context, orgID ulid.ULID) (int, error)
	GetMembersByRoleCount(ctx context.Context, orgID ulid.ULID) (map[string]int, error)
}

// PermissionService defines the normalized permission management service interface.
type PermissionService interface {
	// Permission management
	CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*Permission, error)
	GetPermission(ctx context.Context, permissionID ulid.ULID) (*Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*Permission, error)
	GetPermissionByResourceAction(ctx context.Context, resource, action string) (*Permission, error)
	UpdatePermission(ctx context.Context, permissionID ulid.ULID, req *UpdatePermissionRequest) error
	DeletePermission(ctx context.Context, permissionID ulid.ULID) error

	// Permission queries
	ListPermissions(ctx context.Context, limit, offset int) (*PermissionListResponse, error)
	GetAllPermissions(ctx context.Context) ([]*Permission, error)
	GetPermissionsByResource(ctx context.Context, resource string) ([]*Permission, error)
	GetPermissionsByNames(ctx context.Context, names []string) ([]*Permission, error)
	GetPermissionsByResourceActions(ctx context.Context, resourceActions []string) ([]*Permission, error)
	SearchPermissions(ctx context.Context, query string, limit, offset int) (*PermissionListResponse, error)

	// Resource and action queries
	GetAvailableResources(ctx context.Context) ([]string, error)
	GetActionsForResource(ctx context.Context, resource string) ([]string, error)

	// Permission validation
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
	GetUserBlacklistedTokens(ctx context.Context, filters *BlacklistedTokenFilter) ([]*BlacklistedToken, error)

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
	OrganizationID *ulid.ULID `json:"organization_id,omitempty"`
	Email          string     `json:"email"`
	TokenType      string     `json:"token_type"`
	Issuer         string     `json:"iss"`
	Subject        string     `json:"sub"`
	Scopes         []string   `json:"scopes,omitempty"`
	IssuedAt       int64      `json:"iat"`
	ExpiresAt      int64      `json:"exp"`
	NotBefore      int64      `json:"nbf"`
	UserID         ulid.ULID  `json:"user_id"`
}

// Request/Response DTOs
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
	Name     *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// Filter types
type APIKeyFilters struct {
	// Domain filters
	UserID         *ulid.ULID `json:"user_id,omitempty"`
	OrganizationID *ulid.ULID `json:"organization_id,omitempty"`
	ProjectID      *ulid.ULID `json:"project_id,omitempty"`
	IsActive       *bool      `json:"is_active,omitempty"`
	IsExpired      *bool      `json:"is_expired,omitempty"`

	// Pagination (embedded for DRY)
	pagination.Params `json:",inline"`
}

// Statistics types - AuditLogStats is defined in repository.go

// ScopeService defines the scope-based authorization service interface.
type ScopeService interface {
	// Scope resolution (context-aware)
	GetUserScopes(ctx context.Context, userID ulid.ULID, orgID *ulid.ULID, projectID *ulid.ULID) (*ScopeResolution, error)
	GetUserScopesInOrganization(ctx context.Context, userID, orgID ulid.ULID) (*ScopeResolution, error)
	GetUserScopesInProject(ctx context.Context, userID, orgID, projectID ulid.ULID) (*ScopeResolution, error)

	// Scope checking (boolean checks)
	HasScope(ctx context.Context, userID ulid.ULID, scope string, orgID *ulid.ULID, projectID *ulid.ULID) (bool, error)
	HasAnyScope(ctx context.Context, userID ulid.ULID, scopes []string, orgID *ulid.ULID, projectID *ulid.ULID) (bool, error)
	HasAllScopes(ctx context.Context, userID ulid.ULID, scopes []string, orgID *ulid.ULID, projectID *ulid.ULID) (bool, error)

	// Scope validation
	ValidateScope(ctx context.Context, scope string) error
	GetScopeLevel(ctx context.Context, scope string) (ScopeLevel, error)

	// Scope listing for UI
	GetAvailableScopes(ctx context.Context, level ScopeLevel) ([]string, error)
	GetScopesByCategory(ctx context.Context) ([]ScopeCategory, error)
}

// AuthServices aggregates all authentication-related services (normalized version).
type AuthServices interface {
	Auth() AuthService
	Sessions() SessionService
	APIKeys() APIKeyService
	Roles() RoleService
	OrganizationMembers() OrganizationMemberService
	Permissions() PermissionService
	JWT() JWTService
	BlacklistedTokens() BlacklistedTokenService
	AuditLogs() AuditLogService
	Scopes() ScopeService
}
