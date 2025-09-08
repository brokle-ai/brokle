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

// RoleRepository defines the clean interface for role data access.
type RoleRepository interface {
	// Core CRUD operations
	Create(ctx context.Context, role *Role) error
	GetByID(ctx context.Context, id ulid.ULID) (*Role, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id ulid.ULID) error
	
	// Clean scoped queries
	GetByScope(ctx context.Context, scopeType string, scopeID *ulid.ULID) ([]*Role, error)
	GetByScopedName(ctx context.Context, scopeType string, scopeID *ulid.ULID, name string) (*Role, error)
	GetSystemRoles(ctx context.Context) ([]*Role, error)
	GetOrganizationRoles(ctx context.Context, orgID ulid.ULID) ([]*Role, error)
	GetProjectRoles(ctx context.Context, projectID ulid.ULID) ([]*Role, error)
	
	// User role management (clean)
	AssignUserRole(ctx context.Context, userID, roleID ulid.ULID) error
	RevokeUserRole(ctx context.Context, userID, roleID ulid.ULID) error
	GetUserRoles(ctx context.Context, userID ulid.ULID) ([]*Role, error)
	GetUserRolesByScope(ctx context.Context, userID ulid.ULID, scopeType string) ([]*Role, error)
	
	// Clean permission queries (effective permissions across all scopes)
	GetUserEffectivePermissions(ctx context.Context, userID ulid.ULID) ([]string, error)
	HasUserPermission(ctx context.Context, userID ulid.ULID, permission string) (bool, error)
	CheckUserPermissions(ctx context.Context, userID ulid.ULID, permissions []string) (map[string]bool, error)
	
	// Permission management
	GetRolePermissions(ctx context.Context, roleID ulid.ULID) ([]*Permission, error)
	AssignRolePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error
	RevokeRolePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error
	UpdateRolePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error
	
	// Statistics and validation
	GetRoleStatistics(ctx context.Context) (*RoleStatistics, error)
	CanDeleteRole(ctx context.Context, roleID ulid.ULID) (bool, error)
	
	// Bulk operations
	BulkAssignPermissions(ctx context.Context, assignments []RolePermissionAssignment) error
	BulkRevokePermissions(ctx context.Context, revocations []RolePermissionRevocation) error
}

// RolePermissionAssignment represents a bulk role permission assignment
type RolePermissionAssignment struct {
	RoleID       ulid.ULID `json:"role_id"`
	PermissionID ulid.ULID `json:"permission_id"`
}

// RolePermissionRevocation represents a bulk role permission revocation
type RolePermissionRevocation struct {
	RoleID       ulid.ULID `json:"role_id"`
	PermissionID ulid.ULID `json:"permission_id"`
}

// PermissionRepository defines the interface for permission data access.
type PermissionRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, permission *Permission) error
	GetByID(ctx context.Context, id ulid.ULID) (*Permission, error)
	GetByName(ctx context.Context, name string) (*Permission, error)                                    // Legacy name lookup
	GetByResourceAction(ctx context.Context, resource, action string) (*Permission, error)             // New resource:action lookup
	Update(ctx context.Context, permission *Permission) error
	Delete(ctx context.Context, id ulid.ULID) error
	
	// Permission queries
	GetAllPermissions(ctx context.Context) ([]*Permission, error)
	GetByCategory(ctx context.Context, category string) ([]*Permission, error)
	GetByResource(ctx context.Context, resource string) ([]*Permission, error)                         // Get all permissions for resource
	GetByNames(ctx context.Context, names []string) ([]*Permission, error)                             // Legacy bulk lookup
	GetByResourceActions(ctx context.Context, resourceActions []string) ([]*Permission, error)        // New bulk resource:action lookup
	ListPermissions(ctx context.Context, limit, offset int) ([]*Permission, int, error)               // Paginated list with total count
	SearchPermissions(ctx context.Context, query string, limit, offset int) ([]*Permission, int, error)
	
	// Resource and action queries
	GetAvailableResources(ctx context.Context) ([]string, error)                                       // Get all distinct resources
	GetActionsForResource(ctx context.Context, resource string) ([]string, error)                     // Get all actions for resource
	GetPermissionCategories(ctx context.Context) ([]string, error)                                    // Get all distinct categories
	
	// Role permissions
	GetPermissionsByRoleID(ctx context.Context, roleID ulid.ULID) ([]*Permission, error)
	GetRolePermissionMap(ctx context.Context, roleID ulid.ULID) (map[string]bool, error)              // resource:action -> true
	
	// User permissions (through roles)
	GetUserPermissions(ctx context.Context, userID, orgID ulid.ULID) ([]*Permission, error)           // Full permission objects
	GetUserPermissionStrings(ctx context.Context, userID, orgID ulid.ULID) ([]string, error)         // Just resource:action strings
	GetUserPermissionsByAPIKey(ctx context.Context, apiKeyID ulid.ULID) ([]string, error)
	GetUserEffectivePermissions(ctx context.Context, userID, orgID ulid.ULID) (map[string]bool, error) // resource:action -> true
	
	// Permission validation
	ValidateResourceAction(ctx context.Context, resource, action string) error
	PermissionExists(ctx context.Context, resource, action string) (bool, error)
	BulkPermissionExists(ctx context.Context, resourceActions []string) (map[string]bool, error)
	
	// Bulk operations
	BulkCreate(ctx context.Context, permissions []*Permission) error
	BulkUpdate(ctx context.Context, permissions []*Permission) error
	BulkDelete(ctx context.Context, permissionIDs []ulid.ULID) error
}

// UserRoleRepository defines the clean interface for user role assignments.
type UserRoleRepository interface {
	// Core CRUD operations
	Create(ctx context.Context, userRole *UserRole) error
	Delete(ctx context.Context, userID, roleID ulid.ULID) error
	GetByUser(ctx context.Context, userID ulid.ULID) ([]*UserRole, error)
	GetByRole(ctx context.Context, roleID ulid.ULID) ([]*UserRole, error)
	Exists(ctx context.Context, userID, roleID ulid.ULID) (bool, error)
	
	// Bulk operations
	BulkAssign(ctx context.Context, userRoles []*UserRole) error
	BulkRevoke(ctx context.Context, userID ulid.ULID, roleIDs []ulid.ULID) error
	
	// Statistics
	GetUserRoleCount(ctx context.Context, userID ulid.ULID) (int, error)
	GetRoleUserCount(ctx context.Context, roleID ulid.ULID) (int, error)
}

// RolePermissionRepository defines the clean interface for role-permission relationships.
type RolePermissionRepository interface {
	// Core operations
	Create(ctx context.Context, rolePermission *RolePermission) error
	Delete(ctx context.Context, roleID, permissionID ulid.ULID) error
	GetByRoleID(ctx context.Context, roleID ulid.ULID) ([]*RolePermission, error)
	HasPermission(ctx context.Context, roleID, permissionID ulid.ULID) (bool, error)
	
	// Batch operations
	AssignPermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error
	RevokePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error
	ReplaceAllPermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error
	
	// Bulk operations
	BulkAssign(ctx context.Context, assignments []RolePermissionAssignment) error
	BulkRevoke(ctx context.Context, revocations []RolePermissionRevocation) error
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
	
	// Statistics
	GetAuditLogStats(ctx context.Context) (*AuditLogStats, error)
	GetUserAuditLogStats(ctx context.Context, userID ulid.ULID) (*AuditLogStats, error)
	GetOrganizationAuditLogStats(ctx context.Context, orgID ulid.ULID) (*AuditLogStats, error)
	
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

// AuditLogStats represents audit log statistics
type AuditLogStats struct {
	TotalLogs      int64                `json:"total_logs"`
	LogsByAction   map[string]int64     `json:"logs_by_action"`
	LogsByResource map[string]int64     `json:"logs_by_resource"`
	LastLogTime    *time.Time           `json:"last_log_time,omitempty"`
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

// Repository aggregates all auth-related repositories (clean version).
type Repository interface {
	UserSessions() UserSessionRepository
	BlacklistedTokens() BlacklistedTokenRepository
	APIKeys() APIKeyRepository
	Roles() RoleRepository
	UserRoles() UserRoleRepository
	Permissions() PermissionRepository
	RolePermissions() RolePermissionRepository
	AuditLogs() AuditLogRepository
	PasswordResetTokens() PasswordResetTokenRepository
}