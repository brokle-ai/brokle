package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	"brokle/pkg/pagination"
)

// UserSessionRepository defines the interface for user session data access.
type UserSessionRepository interface {
	Create(ctx context.Context, session *UserSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*UserSession, error)
	GetByJTI(ctx context.Context, jti string) (*UserSession, error)
	GetByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*UserSession, error)
	Update(ctx context.Context, session *UserSession) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*UserSession, error)
	GetActiveSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]*UserSession, error)

	RevokeSession(ctx context.Context, id uuid.UUID) error
	RevokeUserSessions(ctx context.Context, userID uuid.UUID) error
	CleanupExpiredSessions(ctx context.Context) error
}

// BlacklistedTokenRepository defines the interface for blacklisted token data access.
type BlacklistedTokenRepository interface {
	Create(ctx context.Context, blacklistedToken *BlacklistedToken) error
	GetByJTI(ctx context.Context, jti string) (*BlacklistedToken, error)
	IsTokenBlacklisted(ctx context.Context, jti string) (bool, error)

	// User-wide timestamp blacklisting (GDPR/SOC2 compliance)
	CreateUserTimestampBlacklist(ctx context.Context, userID uuid.UUID, blacklistTimestamp int64, reason string) error
	IsUserBlacklistedAfterTimestamp(ctx context.Context, userID uuid.UUID, tokenIssuedAt int64) (bool, error)
	GetUserBlacklistTimestamp(ctx context.Context, userID uuid.UUID) (*int64, error)

	// Cleanup
	CleanupExpiredTokens(ctx context.Context) error
	CleanupTokensOlderThan(ctx context.Context, olderThan time.Time) error

	// Bulk
	BlacklistUserTokens(ctx context.Context, userID uuid.UUID, reason string) error
	GetBlacklistedTokensByUser(ctx context.Context, filters *BlacklistedTokenFilter) ([]*BlacklistedToken, error)

	// Statistics
	GetBlacklistedTokensCount(ctx context.Context) (int64, error)
	GetBlacklistedTokensByReason(ctx context.Context, reason string) ([]*BlacklistedToken, error)
}

// BlacklistedTokenFilter represents filters for blacklisted token queries.
type BlacklistedTokenFilter struct {
	UserID *uuid.UUID
	Reason *string
	pagination.Params
}

// APIKeyRepository defines the interface for API key data access.
type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *APIKey) error
	GetByID(ctx context.Context, id uuid.UUID) (*APIKey, error)
	GetByKeyHash(ctx context.Context, keyHash string) (*APIKey, error)
	Update(ctx context.Context, apiKey *APIKey) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*APIKey, error)
	GetByOrganizationID(ctx context.Context, orgID uuid.UUID) ([]*APIKey, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*APIKey, error)

	UpdateLastUsed(ctx context.Context, id uuid.UUID) error

	GetByFilters(ctx context.Context, filters *APIKeyFilters) ([]*APIKey, error)
	CountByFilters(ctx context.Context, filters *APIKeyFilters) (int64, error)
}

// RoleRepository defines the interface for both system template and custom scoped roles.
type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
	GetByNameAndScope(ctx context.Context, name, scopeType string) (*Role, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetByScopeType(ctx context.Context, scopeType string) ([]*Role, error)
	ListRoles(ctx context.Context) ([]*Role, error)
	GetSystemRoles(ctx context.Context) ([]*Role, error)

	GetByNameScopeAndID(ctx context.Context, name, scopeType string, scopeID *uuid.UUID) (*Role, error)
	GetCustomRolesByOrganization(ctx context.Context, organizationID uuid.UUID) ([]*Role, error)

	// Permission management for roles
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*Permission, error)
	AssignRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID, grantedBy *uuid.UUID) error
	RevokeRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
	UpdateRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID, grantedBy *uuid.UUID) error

	// Statistics
	GetRoleStatistics(ctx context.Context) (*RoleStatistics, error)
}

// PermissionRepository defines the interface for normalized permission data access.
type PermissionRepository interface {
	Create(ctx context.Context, permission *Permission) error
	GetByID(ctx context.Context, id uuid.UUID) (*Permission, error)
	GetByName(ctx context.Context, name string) (*Permission, error)
	GetByResourceAction(ctx context.Context, resource, action string) (*Permission, error)
	Update(ctx context.Context, permission *Permission) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetAllPermissions(ctx context.Context) ([]*Permission, error)
	GetByResource(ctx context.Context, resource string) ([]*Permission, error)

	// Resource and action queries
	GetAvailableResources(ctx context.Context) ([]string, error)
	GetActionsForResource(ctx context.Context, resource string) ([]string, error)

	// Permission validation
	PermissionExists(ctx context.Context, resource, action string) (bool, error)
}

// OrganizationMemberRepository defines the interface for organization membership management.
type OrganizationMemberRepository interface {
	Create(ctx context.Context, member *OrganizationMember) error
	GetByUserAndOrganization(ctx context.Context, userID, orgID uuid.UUID) (*OrganizationMember, error)
	Update(ctx context.Context, member *OrganizationMember) error
	Delete(ctx context.Context, userID, orgID uuid.UUID) error

	// Membership queries
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*OrganizationMember, error)
	GetByOrganizationID(ctx context.Context, orgID uuid.UUID) ([]*OrganizationMember, error)
	GetByRole(ctx context.Context, roleID uuid.UUID) ([]*OrganizationMember, error)
	Exists(ctx context.Context, userID, orgID uuid.UUID) (bool, error)

	// Permission queries
	GetUserEffectivePermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	CheckUserPermissions(ctx context.Context, userID uuid.UUID, permissions []string) (map[string]bool, error)
	GetUserPermissionsInOrganization(ctx context.Context, userID, orgID uuid.UUID) ([]string, error)

	// Active-member listing (filters out soft-deleted rows; suspension
	// is no longer modeled — see auth.go OrganizationMember docstring).
	GetActiveMembers(ctx context.Context, orgID uuid.UUID) ([]*OrganizationMember, error)

	// Role management
	UpdateMemberRole(ctx context.Context, userID, orgID, roleID uuid.UUID) error

	// Statistics
	GetMemberCount(ctx context.Context, orgID uuid.UUID) (int, error)
	GetMembersByRole(ctx context.Context, orgID uuid.UUID) (map[string]int, error)
}

// ProjectMemberRepository defines the interface for project membership.
// Project membership is an ADDITIVE role grant on top of the user's
// organization role — per-resource scopes UNION with the org role's
// project-tier projection. The resolver lives in
// ListUserEffectivePermissionsInScope and is consumed by the scope-
// aware permission middleware. See docs/adr/0001-rbac-additive-
// semantics.md for the migration history (round 11 OVERRIDE → round 24
// additive).
type ProjectMemberRepository interface {
	// Core CRUD
	//
	// Create executes the atomic UPSERT-WHERE defined by the
	// CreateProjectMember query and returns the affected-row count.
	// 1 = inserted (fresh add) or updated in-place (orphan repair).
	// 0 = row exists AND is visible (active org membership) — the
	//     UPSERT WHERE-clause refused to overwrite; caller should
	//     return Conflict.
	// See the SQL query comment block + CLAUDE.md 2026-04-30
	// (project-rbac) for the full rationale.
	Create(ctx context.Context, member *ProjectMember) (int64, error)
	GetByUserAndProject(ctx context.Context, userID, projectID uuid.UUID) (*ProjectMember, error)
	UpdateRole(ctx context.Context, userID, projectID, roleID uuid.UUID) error
	Delete(ctx context.Context, userID, projectID uuid.UUID) error

	// DeleteAllInOrgForUser hard-deletes every project_members row this
	// user holds within the given organization's projects. Called by
	// the org-removal application service to cascade-clean per-resource
	// grants atomically with the org_members soft-delete. The caller
	// owns the transaction scope (use-case-level atomicity per
	// Vernon's DDD); this method participates in whatever tx ctx
	// already carries.
	DeleteAllInOrgForUser(ctx context.Context, userID, orgID uuid.UUID) error

	// Membership queries
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]*ProjectMember, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*ProjectMember, error)
	IsMember(ctx context.Context, userID, projectID uuid.UUID) (bool, error)
	GetMemberCount(ctx context.Context, projectID uuid.UUID) (int, error)

	// Effective-permission resolution (additive semantics, round 24:
	// org role's project-tier projection UNIONs with the project_members
	// override grant; both branches require active org membership).
	ListUserEffectivePermissionsInScope(ctx context.Context, userID, orgID, projectID uuid.UUID) ([]string, error)
}

// RolePermissionRepository defines the interface for role-permission relationships.
type RolePermissionRepository interface {
	Create(ctx context.Context, rolePermission *RolePermission) error
	Delete(ctx context.Context, roleID, permissionID uuid.UUID) error
}

// AuditLogRepository defines the interface for audit log data access.
type AuditLogRepository interface {
	Create(ctx context.Context, auditLog *AuditLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*AuditLog, error)

	// Audit log queries
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*AuditLog, error)
	GetByOrganizationID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*AuditLog, error)
	GetByResource(ctx context.Context, resource, resourceID string, limit, offset int) ([]*AuditLog, error)
}

// AuditLogFilters represents filters for audit log queries.
type AuditLogFilters struct {
	UserID         *uuid.UUID `json:"user_id,omitempty"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	Action         *string    `json:"action,omitempty"`
	Resource       *string    `json:"resource,omitempty"`
	ResourceID     *string    `json:"resource_id,omitempty"`
	IPAddress      *string    `json:"ip_address,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`

	pagination.Params `json:",inline"`
}

// AuditLogStats represents audit log statistics.
type AuditLogStats struct {
	LogsByAction   map[string]int64 `json:"logs_by_action"`
	LogsByResource map[string]int64 `json:"logs_by_resource"`
	LastLogTime    *time.Time       `json:"last_log_time,omitempty"`
	TotalLogs      int64            `json:"total_logs"`
}

// PasswordResetTokenRepository defines the interface for password reset token data access.
type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *PasswordResetToken) error
	GetByID(ctx context.Context, id uuid.UUID) (*PasswordResetToken, error)
	GetByToken(ctx context.Context, token string) (*PasswordResetToken, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*PasswordResetToken, error)
	Update(ctx context.Context, token *PasswordResetToken) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Token lifecycle
	MarkAsUsed(ctx context.Context, id uuid.UUID) error
	IsValid(ctx context.Context, id uuid.UUID) (bool, error)

	// Cleanup
	CleanupExpiredTokens(ctx context.Context) error
	InvalidateAllUserTokens(ctx context.Context, userID uuid.UUID) error
}
