package auth

import (
	"context"
	"time"

	"brokle/pkg/ulid"
)

// UserSessionRepository defines the interface for user session data access.
type UserSessionRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, session *UserSession) error
	GetByID(ctx context.Context, id ulid.ULID) (*UserSession, error)
	GetByJTI(ctx context.Context, jti string) (*UserSession, error)              // Get session by JWT ID (current access token)
	GetByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*UserSession, error) // Get session by refresh token hash
	Update(ctx context.Context, session *UserSession) error
	Delete(ctx context.Context, id ulid.ULID) error
	
	// User sessions
	GetByUserID(ctx context.Context, userID ulid.ULID) ([]*UserSession, error)
	GetActiveSessionsByUserID(ctx context.Context, userID ulid.ULID) ([]*UserSession, error)
	
	// Session management
	DeactivateSession(ctx context.Context, id ulid.ULID) error
	DeactivateUserSessions(ctx context.Context, userID ulid.ULID) error
	RevokeSession(ctx context.Context, id ulid.ULID) error
	RevokeUserSessions(ctx context.Context, userID ulid.ULID) error
	CleanupExpiredSessions(ctx context.Context) error
	CleanupRevokedSessions(ctx context.Context) error
	MarkAsUsed(ctx context.Context, id ulid.ULID) error
	
	// Device-specific queries
	GetByDeviceInfo(ctx context.Context, userID ulid.ULID, deviceInfo interface{}) ([]*UserSession, error)
	GetActiveSessionsCount(ctx context.Context, userID ulid.ULID) (int, error)
}

// BlacklistedTokenRepository defines the interface for blacklisted token data access.
type BlacklistedTokenRepository interface {
	// Basic operations
	Create(ctx context.Context, blacklistedToken *BlacklistedToken) error
	GetByJTI(ctx context.Context, jti string) (*BlacklistedToken, error)
	IsTokenBlacklisted(ctx context.Context, jti string) (bool, error)
	
	// User-wide timestamp blacklisting (GDPR/SOC2 compliance)
	CreateUserTimestampBlacklist(ctx context.Context, userID ulid.ULID, blacklistTimestamp int64, reason string) error
	IsUserBlacklistedAfterTimestamp(ctx context.Context, userID ulid.ULID, tokenIssuedAt int64) (bool, error)
	GetUserBlacklistTimestamp(ctx context.Context, userID ulid.ULID) (*int64, error)
	
	// Cleanup operations
	CleanupExpiredTokens(ctx context.Context) error
	CleanupTokensOlderThan(ctx context.Context, olderThan time.Time) error
	
	// Bulk operations
	BlacklistUserTokens(ctx context.Context, userID ulid.ULID, reason string) error
	GetBlacklistedTokensByUser(ctx context.Context, userID ulid.ULID, limit, offset int) ([]*BlacklistedToken, error)
	
	// Statistics
	GetBlacklistedTokensCount(ctx context.Context) (int64, error)
	GetBlacklistedTokensByReason(ctx context.Context, reason string) ([]*BlacklistedToken, error)
}

// APIKeyRepository defines the interface for API key data access.
type APIKeyRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, apiKey *APIKey) error
	GetByID(ctx context.Context, id ulid.ULID) (*APIKey, error)
	GetByKeyHash(ctx context.Context, keyHash string) (*APIKey, error)
	Update(ctx context.Context, apiKey *APIKey) error
	Delete(ctx context.Context, id ulid.ULID) error
	
	// User and organization scoped
	GetByUserID(ctx context.Context, userID ulid.ULID) ([]*APIKey, error)
	GetByOrganizationID(ctx context.Context, orgID ulid.ULID) ([]*APIKey, error)
	GetByProjectID(ctx context.Context, projectID ulid.ULID) ([]*APIKey, error)
	GetByEnvironmentID(ctx context.Context, envID ulid.ULID) ([]*APIKey, error)
	
	// API key management
	DeactivateAPIKey(ctx context.Context, id ulid.ULID) error
	MarkAsUsed(ctx context.Context, id ulid.ULID) error
	CleanupExpiredAPIKeys(ctx context.Context) error
	
	// Statistics
	GetAPIKeyCount(ctx context.Context, userID ulid.ULID) (int, error)
	GetActiveAPIKeyCount(ctx context.Context, userID ulid.ULID) (int, error)
}

// RoleRepository defines the interface for role data access.
type RoleRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, role *Role) error
	GetByID(ctx context.Context, id ulid.ULID) (*Role, error)
	GetByName(ctx context.Context, orgID *ulid.ULID, name string) (*Role, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id ulid.ULID) error
	
	// System vs custom roles
	GetSystemRoles(ctx context.Context) ([]*Role, error)
	GetOrganizationRoles(ctx context.Context, orgID ulid.ULID) ([]*Role, error)
	GetAllRoles(ctx context.Context, orgID ulid.ULID) ([]*Role, error) // System + org roles
	
	// Permission management
	GetRolePermissions(ctx context.Context, roleID ulid.ULID) ([]*Permission, error)
	AssignPermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error
	RevokePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error
	RevokeAllPermissions(ctx context.Context, roleID ulid.ULID) error
	
	// Role validation
	IsSystemRole(ctx context.Context, roleID ulid.ULID) (bool, error)
	CanDeleteRole(ctx context.Context, roleID ulid.ULID) (bool, error)
}

// PermissionRepository defines the interface for permission data access.
type PermissionRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, permission *Permission) error
	GetByID(ctx context.Context, id ulid.ULID) (*Permission, error)
	GetByName(ctx context.Context, name string) (*Permission, error)
	Update(ctx context.Context, permission *Permission) error
	Delete(ctx context.Context, id ulid.ULID) error
	
	// Permission queries
	GetAllPermissions(ctx context.Context) ([]*Permission, error)
	GetByCategory(ctx context.Context, category string) ([]*Permission, error)
	GetByNames(ctx context.Context, names []string) ([]*Permission, error)
	
	// Role permissions
	GetPermissionsByRoleID(ctx context.Context, roleID ulid.ULID) ([]*Permission, error)
	
	// User permissions (through roles)
	GetUserPermissions(ctx context.Context, userID, orgID ulid.ULID) ([]string, error)
	GetUserPermissionsByAPIKey(ctx context.Context, apiKeyID ulid.ULID) ([]string, error)
}

// RolePermissionRepository defines the interface for role-permission relationship data access.
type RolePermissionRepository interface {
	// Relationship management
	Create(ctx context.Context, rolePermission *RolePermission) error
	Delete(ctx context.Context, roleID, permissionID ulid.ULID) error
	
	// Batch operations
	AssignPermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error
	RevokePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error
	RevokeAllPermissions(ctx context.Context, roleID ulid.ULID) error
	
	// Queries
	GetByRoleID(ctx context.Context, roleID ulid.ULID) ([]*RolePermission, error)
	GetByPermissionID(ctx context.Context, permissionID ulid.ULID) ([]*RolePermission, error)
	HasPermission(ctx context.Context, roleID, permissionID ulid.ULID) (bool, error)
}

// AuditLogRepository defines the interface for audit log data access.
type AuditLogRepository interface {
	// Basic operations
	Create(ctx context.Context, auditLog *AuditLog) error
	GetByID(ctx context.Context, id ulid.ULID) (*AuditLog, error)
	
	// Audit log queries
	GetByUserID(ctx context.Context, userID ulid.ULID, limit, offset int) ([]*AuditLog, error)
	GetByOrganizationID(ctx context.Context, orgID ulid.ULID, limit, offset int) ([]*AuditLog, error)
	GetByResource(ctx context.Context, resource, resourceID string, limit, offset int) ([]*AuditLog, error)
	GetByAction(ctx context.Context, action string, limit, offset int) ([]*AuditLog, error)
	GetByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*AuditLog, error)
	
	// Advanced queries
	Search(ctx context.Context, filters *AuditLogFilters) ([]*AuditLog, int, error)
	
	// Cleanup
	CleanupOldLogs(ctx context.Context, olderThan time.Time) error
}

// AuditLogFilters represents filters for audit log queries.
type AuditLogFilters struct {
	// Pagination
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	
	// Filters
	UserID         *ulid.ULID `json:"user_id,omitempty"`
	OrganizationID *ulid.ULID `json:"organization_id,omitempty"`
	Action         *string    `json:"action,omitempty"`
	Resource       *string    `json:"resource,omitempty"`
	ResourceID     *string    `json:"resource_id,omitempty"`
	IPAddress      *string    `json:"ip_address,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	
	// Sorting
	SortBy    string `json:"sort_by"`    // created_at, action, resource
	SortOrder string `json:"sort_order"` // asc, desc
}

// PasswordResetTokenRepository defines the interface for password reset token data access.
type PasswordResetTokenRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, token *PasswordResetToken) error
	GetByID(ctx context.Context, id ulid.ULID) (*PasswordResetToken, error)
	GetByToken(ctx context.Context, token string) (*PasswordResetToken, error)
	GetByUserID(ctx context.Context, userID ulid.ULID) ([]*PasswordResetToken, error)
	Update(ctx context.Context, token *PasswordResetToken) error
	Delete(ctx context.Context, id ulid.ULID) error
	
	// Token management
	MarkAsUsed(ctx context.Context, id ulid.ULID) error
	IsUsed(ctx context.Context, id ulid.ULID) (bool, error)
	IsValid(ctx context.Context, id ulid.ULID) (bool, error)
	GetValidTokenByUserID(ctx context.Context, userID ulid.ULID) (*PasswordResetToken, error)
	
	// Cleanup operations
	CleanupExpiredTokens(ctx context.Context) error
	CleanupUsedTokens(ctx context.Context, olderThan time.Time) error
	InvalidateAllUserTokens(ctx context.Context, userID ulid.ULID) error
}

// Repository aggregates all auth-related repositories.
type Repository interface {
	UserSessions() UserSessionRepository
	BlacklistedTokens() BlacklistedTokenRepository // NEW: Added blacklisted tokens repository
	APIKeys() APIKeyRepository
	Roles() RoleRepository
	Permissions() PermissionRepository
	RolePermissions() RolePermissionRepository
	AuditLogs() AuditLogRepository
	PasswordResetTokens() PasswordResetTokenRepository
}