package auth

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"brokle/internal/core/domain/shared"
	"brokle/pkg/uid"
)

// ----- Roles --------------------------------------------------------

// Role represents both system template roles and custom scoped roles.
// A system template role has ScopeType == ScopeSystem and ScopeID == nil;
// a custom role has ScopeType ∈ {organization, project} and ScopeID
// pointing at the owning aggregate.
type Role struct {
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	ScopeID         *uuid.UUID       `json:"scope_id,omitempty"`
	Description     *string          `json:"description,omitempty"`
	Name            string           `json:"name"`
	ScopeType       string           `json:"scope_type"`
	Permissions     []Permission     `json:"permissions,omitempty"`
	RolePermissions []RolePermission `json:"role_permissions,omitempty"`
	ID              uuid.UUID        `json:"id"`
}

// Scope discriminators for roles.
const (
	ScopeSystem       = "system"       // System template roles
	ScopeOrganization = "organization" // Organization-specific roles
	ScopeProject      = "project"      // Project-specific roles
)

// IsSystemRole reports whether the role is a system template (no scope binding).
func (r *Role) IsSystemRole() bool { return r.ScopeType == ScopeSystem && r.ScopeID == nil }

// IsCustomRole reports whether the role is bound to a specific aggregate.
func (r *Role) IsCustomRole() bool { return r.ScopeType != ScopeSystem && r.ScopeID != nil }

// IsOrganizationRole reports whether the role is org-scoped.
func (r *Role) IsOrganizationRole() bool { return r.ScopeType == ScopeOrganization }

// IsProjectRole reports whether the role is project-scoped.
func (r *Role) IsProjectRole() bool { return r.ScopeType == ScopeProject }

// GetScopeDisplay returns a human-readable scope label for UI surfaces.
func (r *Role) GetScopeDisplay() string {
	switch r.ScopeType {
	case ScopeSystem:
		return "System"
	case ScopeOrganization:
		if r.ScopeID == nil {
			return "Organization Template"
		}
		return "Organization Custom"
	case ScopeProject:
		return "Project"
	default:
		return "Unknown"
	}
}

// AddPermission constructs a RolePermission grant linking this role to
// a permission. Caller persists the result via the repository.
func (r *Role) AddPermission(permissionID uuid.UUID, grantedBy *uuid.UUID) *RolePermission {
	return &RolePermission{
		RoleID:       r.ID,
		PermissionID: permissionID,
		GrantedAt:    time.Now(),
		GrantedBy:    grantedBy,
	}
}

// NewRole constructs a system template role (no scope binding).
func NewRole(name, scopeType, description string) *Role {
	return &Role{
		ID:          uid.New(),
		Name:        name,
		ScopeType:   scopeType,
		ScopeID:     nil,
		Description: shared.NilIfEmpty(description),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// NewCustomRole constructs a role bound to a specific organization or project.
func NewCustomRole(name, scopeType, description string, scopeID uuid.UUID) *Role {
	return &Role{
		ID:          uid.New(),
		Name:        name,
		ScopeType:   scopeType,
		ScopeID:     &scopeID,
		Description: shared.NilIfEmpty(description),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// ----- Membership ---------------------------------------------------

// OrganizationMember represents user membership in an organization with a single role.
//
// Lifecycle is single-axis via the soft-delete `deleted_at` column on
// the `organization_members` row (set by RemoveMember, restored on
// re-invite). The legacy per-member `status` flag (active/suspended)
// was deleted on 2026-04-30 — it had no callers in any HTTP route or
// CLI flow, and production peers (GitHub/GitLab/Slack/Auth0/WorkOS/
// Clerk) all do suspension at the user level rather than per-org.
type OrganizationMember struct {
	JoinedAt       time.Time  `json:"joined_at"`
	InvitedBy      *uuid.UUID `json:"invited_by,omitempty"`
	Role           *Role      `json:"role,omitempty"`
	UserID         uuid.UUID  `json:"user_id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	RoleID         uuid.UUID  `json:"role_id"`
}

// NewOrganizationMember constructs an active membership record.
func NewOrganizationMember(userID, organizationID, roleID uuid.UUID, invitedBy *uuid.UUID) *OrganizationMember {
	return &OrganizationMember{
		UserID:         userID,
		OrganizationID: organizationID,
		RoleID:         roleID,
		JoinedAt:       time.Now(),
		InvitedBy:      invitedBy,
	}
}

// ProjectMember represents user membership in a project with a single role.
// Lifecycle is single-axis via the soft-delete `deleted_at` column on
// the parent `organization_members` row (see OrganizationMember). The
// project_members table itself has no lifecycle column — rows are
// hard-deleted by ProjectMemberRepository.DeleteAllInOrgForUser when
// the parent org membership is removed.
type ProjectMember struct {
	JoinedAt  time.Time `json:"joined_at"`
	Role      *Role     `json:"role,omitempty"`
	UserID    uuid.UUID `json:"user_id"`
	ProjectID uuid.UUID `json:"project_id"`
	RoleID    uuid.UUID `json:"role_id"`
}

// NewProjectMember constructs an active project-membership record.
func NewProjectMember(userID, projectID, roleID uuid.UUID) *ProjectMember {
	return &ProjectMember{
		UserID:    userID,
		ProjectID: projectID,
		RoleID:    roleID,
		JoinedAt:  time.Now(),
	}
}

// ----- Permissions --------------------------------------------------

// ScopeLevel defines where a scope applies in the hierarchy.
type ScopeLevel string

const (
	// ScopeLevelGlobal: platform-wide scopes (system admin only, future feature).
	ScopeLevelGlobal ScopeLevel = "global"

	// ScopeLevelOrganization: organization-wide scopes (apply to org + all projects).
	ScopeLevelOrganization ScopeLevel = "organization"

	// ScopeLevelProject: project-specific scopes.
	ScopeLevelProject ScopeLevel = "project"
)

// Permission represents a normalized permission using resource:action format.
type Permission struct {
	CreatedAt   time.Time  `json:"created_at"`
	Description *string    `json:"description,omitempty"`
	Category    *string    `json:"category,omitempty"`
	Name        string     `json:"name"`
	Resource    string     `json:"resource"`
	Action      string     `json:"action"`
	ScopeLevel  ScopeLevel `json:"scope_level"`
	Roles       []Role     `json:"roles,omitempty"`
	ID          uuid.UUID  `json:"id"`
}

// GetResourceAction returns the resource:action format string.
func (p *Permission) GetResourceAction() string {
	return fmt.Sprintf("%s:%s", p.Resource, p.Action)
}

// IsWildcardPermission reports whether this is a wildcard permission (*:* or resource:*).
func (p *Permission) IsWildcardPermission() bool {
	return p.Resource == "*" || p.Action == "*"
}

// MatchesResourceAction checks whether this permission matches the given
// resource:action — exact match, wildcard resource, wildcard action,
// or full wildcard.
func (p *Permission) MatchesResourceAction(resource, action string) bool {
	if p.Resource == resource && p.Action == action {
		return true
	}
	if p.Resource == "*" && p.Action == action {
		return true
	}
	if p.Resource == resource && p.Action == "*" {
		return true
	}
	if p.Resource == "*" && p.Action == "*" {
		return true
	}
	return false
}

// ParseResourceAction parses a resource:action string into its components.
func ParseResourceAction(resourceAction string) (resource, action string, err error) {
	parts := strings.Split(resourceAction, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid resource:action format: %s", resourceAction)
	}
	return parts[0], parts[1], nil
}

// ValidateResourceAction validates a resource:action string format.
func ValidateResourceAction(resourceAction string) error {
	_, _, err := ParseResourceAction(resourceAction)
	return err
}

// NewPermission constructs a permission with default ScopeLevel = organization.
func NewPermission(resource, action, description string) *Permission {
	name := fmt.Sprintf("%s:%s", resource, action)
	return &Permission{
		ID:          uid.New(),
		Name:        name,
		Resource:    resource,
		Action:      action,
		Description: shared.NilIfEmpty(description),
		ScopeLevel:  ScopeLevelOrganization,
		Category:    shared.NilIfEmpty(resource),
		CreatedAt:   time.Now(),
	}
}

// NewPermissionWithScope constructs a permission with explicit scope level and category.
func NewPermissionWithScope(resource, action, description string, scopeLevel ScopeLevel, category string) *Permission {
	name := fmt.Sprintf("%s:%s", resource, action)
	return &Permission{
		ID:          uid.New(),
		Name:        name,
		Resource:    resource,
		Action:      action,
		Description: shared.NilIfEmpty(description),
		ScopeLevel:  scopeLevel,
		Category:    shared.NilIfEmpty(category),
		CreatedAt:   time.Now(),
	}
}

// RolePermission represents the many-to-many relationship between
// roles and permissions, plus audit fields.
type RolePermission struct {
	GrantedAt    time.Time  `json:"granted_at"`
	GrantedBy    *uuid.UUID `json:"granted_by,omitempty"`
	Role         Role       `json:"role,omitempty"`
	Permission   Permission `json:"permission,omitempty"`
	RoleID       uuid.UUID  `json:"role_id"`
	PermissionID uuid.UUID  `json:"permission_id"`
}

// ----- Audit log ----------------------------------------------------

// AuditLog represents an audit log entry for compliance.
//
// Field types mirror the Postgres schema exactly so pgx can scan
// directly into this struct via RowToAddrOfStructByName. Nullable
// columns are pointers; Metadata is raw JSON preserved end-to-end
// without re-parsing.
type AuditLog struct {
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UserID         *uuid.UUID      `json:"user_id,omitempty" db:"user_id"`
	OrganizationID *uuid.UUID      `json:"organization_id,omitempty" db:"organization_id"`
	ResourceID     *string         `json:"resource_id,omitempty" db:"resource_id"`
	IPAddress      *string         `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent      *string         `json:"user_agent,omitempty" db:"user_agent"`
	Action         string          `json:"action" db:"action"`
	Resource       string          `json:"resource" db:"resource"`
	Metadata       json.RawMessage `json:"metadata" db:"metadata" swaggertype:"object"`
	ID             uuid.UUID       `json:"id" db:"id"`
}

// NewAuditLog constructs an audit log entry. Optional fields
// (resource_id, ip_address, user_agent) accept zero-value strings that
// the repository maps to SQL NULL; metadata is marshalled from a typed
// map so callers never hand-craft JSON strings with fmt.Sprintf.
func NewAuditLog(userID, orgID *uuid.UUID, action, resource, resourceID string, metadata map[string]any, ipAddress, userAgent string) *AuditLog {
	log := &AuditLog{
		ID:             uid.New(),
		UserID:         userID,
		OrganizationID: orgID,
		Action:         action,
		Resource:       resource,
		CreatedAt:      time.Now(),
	}
	if resourceID != "" {
		log.ResourceID = &resourceID
	}
	if ipAddress != "" {
		log.IPAddress = &ipAddress
	}
	if userAgent != "" {
		log.UserAgent = &userAgent
	}
	if len(metadata) > 0 {
		if raw, err := json.Marshal(metadata); err == nil {
			log.Metadata = raw
		}
	}
	return log
}

// ----- Wire DTOs for the RBAC handlers -------------------------------

// CreateRoleRequest is the body of POST /api/v1/roles.
type CreateRoleRequest struct {
	ScopeType     string      `json:"scope_type" validate:"required,oneof=system organization project environment"`
	Name          string      `json:"name" validate:"required,min=1,max=100"`
	Description   string      `json:"description,omitempty"`
	PermissionIDs []uuid.UUID `json:"permission_ids,omitempty"`
}

// UpdateRoleRequest is the body of PATCH /api/v1/roles/{id}.
type UpdateRoleRequest struct {
	Description   *string     `json:"description,omitempty"`
	PermissionIDs []uuid.UUID `json:"permission_ids,omitempty"`
}

// CreatePermissionRequest is the body of POST /api/v1/permissions.
type CreatePermissionRequest struct {
	Resource    string `json:"resource" validate:"required,min=1,max=50"`
	Action      string `json:"action" validate:"required,min=1,max=50"`
	Description string `json:"description,omitempty"`
}

// UpdatePermissionRequest is the body of PATCH /api/v1/permissions/{id}.
type UpdatePermissionRequest struct {
	Description *string `json:"description,omitempty"`
}

// AssignRoleRequest is the body of POST /api/v1/.../roles (for
// assigning a role to a user within a scope).
type AssignRoleRequest struct {
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

// UserPermissionsResponse represents a user's effective permissions
// across all scopes. Returned by GET /api/v1/me/permissions.
type UserPermissionsResponse struct {
	Roles           []*Role       `json:"roles"`
	Permissions     []*Permission `json:"permissions"`
	ResourceActions []string      `json:"resource_actions"`
	UserID          uuid.UUID     `json:"user_id"`
}

// CheckPermissionsRequest is the body of POST /api/v1/me/permissions/check.
type CheckPermissionsRequest struct {
	ResourceActions []string `json:"resource_actions" validate:"required,min=1"`
}

// CheckPermissionsResponse is the response body — keys are
// resource:action strings, values indicate whether the principal holds
// each permission.
type CheckPermissionsResponse struct {
	Results map[string]bool `json:"results"`
}

// RoleStatistics is the aggregate-counter snapshot surfaced by the
// platform overview's RBAC widget.
type RoleStatistics struct {
	LastUpdated       time.Time      `json:"last_updated"`
	ScopeDistribution map[string]int `json:"scope_distribution"`
	RoleDistribution  map[string]int `json:"role_distribution"`
	TotalRoles        int            `json:"total_roles"`
	SystemRoles       int            `json:"system_roles"`
	OrganizationRoles int            `json:"organization_roles"`
	ProjectRoles      int            `json:"project_roles"`
	PermissionCount   int            `json:"permission_count"`
}
