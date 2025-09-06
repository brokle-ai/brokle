package organization

import (
	"context"

	"brokle/pkg/ulid"
)

// OrganizationService defines the organization management service interface.
type OrganizationService interface {
	// Organization CRUD operations
	CreateOrganization(ctx context.Context, userID ulid.ULID, req *CreateOrganizationRequest) (*Organization, error)
	GetOrganization(ctx context.Context, orgID ulid.ULID) (*Organization, error)
	GetOrganizationBySlug(ctx context.Context, slug string) (*Organization, error)
	UpdateOrganization(ctx context.Context, orgID ulid.ULID, req *UpdateOrganizationRequest) error
	DeleteOrganization(ctx context.Context, orgID ulid.ULID) error
	ListOrganizations(ctx context.Context, filters *OrganizationFilters) ([]*Organization, error)
	
	// User organization context
	GetUserOrganizations(ctx context.Context, userID ulid.ULID) ([]*Organization, error)
	GetUserDefaultOrganization(ctx context.Context, userID ulid.ULID) (*Organization, error)
	SetUserDefaultOrganization(ctx context.Context, userID, orgID ulid.ULID) error
}

// MemberService defines the organization member management service interface.
type MemberService interface {
	// Member management
	AddMember(ctx context.Context, orgID, userID, roleID ulid.ULID, addedByID ulid.ULID) error
	RemoveMember(ctx context.Context, orgID, userID ulid.ULID, removedByID ulid.ULID) error
	UpdateMemberRole(ctx context.Context, orgID, userID, roleID ulid.ULID, updatedByID ulid.ULID) error
	GetMember(ctx context.Context, orgID, userID ulid.ULID) (*Member, error)
	GetMembers(ctx context.Context, orgID ulid.ULID) ([]*Member, error)
	
	// Member validation
	IsMember(ctx context.Context, userID, orgID ulid.ULID) (bool, error)
	CanUserAccessOrganization(ctx context.Context, userID, orgID ulid.ULID) (bool, error)
	GetUserRole(ctx context.Context, userID, orgID ulid.ULID) (ulid.ULID, error)
	
	// Member statistics
	GetMemberCount(ctx context.Context, orgID ulid.ULID) (int, error)
	GetMembersByRole(ctx context.Context, orgID, roleID ulid.ULID) ([]*Member, error)
}

// ProjectService defines the project management service interface.
type ProjectService interface {
	// Project CRUD operations
	CreateProject(ctx context.Context, orgID ulid.ULID, req *CreateProjectRequest) (*Project, error)
	GetProject(ctx context.Context, projectID ulid.ULID) (*Project, error)
	GetProjectBySlug(ctx context.Context, orgID ulid.ULID, slug string) (*Project, error)
	UpdateProject(ctx context.Context, projectID ulid.ULID, req *UpdateProjectRequest) error
	DeleteProject(ctx context.Context, projectID ulid.ULID) error
	
	// Organization projects
	GetProjectsByOrganization(ctx context.Context, orgID ulid.ULID) ([]*Project, error)
	GetProjectCount(ctx context.Context, orgID ulid.ULID) (int, error)
	
	// Access validation
	CanUserAccessProject(ctx context.Context, userID, projectID ulid.ULID) (bool, error)
	ValidateProjectAccess(ctx context.Context, userID, projectID ulid.ULID) error
}

// EnvironmentService defines the environment management service interface.
type EnvironmentService interface {
	// Environment CRUD operations
	CreateEnvironment(ctx context.Context, projectID ulid.ULID, req *CreateEnvironmentRequest) (*Environment, error)
	GetEnvironment(ctx context.Context, envID ulid.ULID) (*Environment, error)
	GetEnvironmentBySlug(ctx context.Context, projectID ulid.ULID, slug string) (*Environment, error)
	UpdateEnvironment(ctx context.Context, envID ulid.ULID, req *UpdateEnvironmentRequest) error
	DeleteEnvironment(ctx context.Context, envID ulid.ULID) error
	
	// Project environments
	GetEnvironmentsByProject(ctx context.Context, projectID ulid.ULID) ([]*Environment, error)
	GetEnvironmentCount(ctx context.Context, projectID ulid.ULID) (int, error)
	
	// Default environments
	CreateDefaultEnvironments(ctx context.Context, projectID ulid.ULID) error
	
	// Access validation
	CanUserAccessEnvironment(ctx context.Context, userID, envID ulid.ULID) (bool, error)
	ValidateEnvironmentAccess(ctx context.Context, userID, envID ulid.ULID) error
	GetEnvironmentOrganization(ctx context.Context, envID ulid.ULID) (ulid.ULID, error)
}

// InvitationService defines the user invitation service interface.
type InvitationService interface {
	// Invitation management
	InviteUser(ctx context.Context, orgID ulid.ULID, inviterID ulid.ULID, req *InviteUserRequest) (*Invitation, error)
	AcceptInvitation(ctx context.Context, token string, userID ulid.ULID) error
	DeclineInvitation(ctx context.Context, token string) error
	RevokeInvitation(ctx context.Context, invitationID ulid.ULID, revokedByID ulid.ULID) error
	ResendInvitation(ctx context.Context, invitationID ulid.ULID, resentByID ulid.ULID) error
	
	// Invitation queries
	GetInvitation(ctx context.Context, invitationID ulid.ULID) (*Invitation, error)
	GetInvitationByToken(ctx context.Context, token string) (*Invitation, error)
	GetPendingInvitations(ctx context.Context, orgID ulid.ULID) ([]*Invitation, error)
	GetUserInvitations(ctx context.Context, email string) ([]*Invitation, error)
	
	// Invitation validation
	ValidateInvitationToken(ctx context.Context, token string) (*Invitation, error)
	IsEmailAlreadyInvited(ctx context.Context, email string, orgID ulid.ULID) (bool, error)
	
	// Cleanup
	CleanupExpiredInvitations(ctx context.Context) error
}

// OrganizationSettingsService defines the organization settings management service interface.
type OrganizationSettingsService interface {
	// Settings CRUD operations
	CreateSetting(ctx context.Context, orgID ulid.ULID, userID ulid.ULID, req *CreateOrganizationSettingRequest) (*OrganizationSettings, error)
	GetSetting(ctx context.Context, orgID ulid.ULID, key string) (*OrganizationSettings, error)
	GetAllSettings(ctx context.Context, orgID ulid.ULID) (map[string]interface{}, error)
	UpdateSetting(ctx context.Context, orgID ulid.ULID, key string, userID ulid.ULID, req *UpdateOrganizationSettingRequest) error
	DeleteSetting(ctx context.Context, orgID ulid.ULID, key string, userID ulid.ULID) error
	
	// Bulk operations
	UpsertSetting(ctx context.Context, orgID ulid.ULID, key string, value interface{}, userID ulid.ULID) (*OrganizationSettings, error)
	CreateMultipleSettings(ctx context.Context, orgID ulid.ULID, userID ulid.ULID, settings map[string]interface{}) error
	GetSettingsByKeys(ctx context.Context, orgID ulid.ULID, keys []string) (map[string]interface{}, error)
	DeleteMultipleSettings(ctx context.Context, orgID ulid.ULID, keys []string, userID ulid.ULID) error
	
	// Access validation
	ValidateSettingsAccess(ctx context.Context, userID, orgID ulid.ULID, operation string) error
	CanUserManageSettings(ctx context.Context, userID, orgID ulid.ULID) (bool, error)
	
	// Settings management
	ResetToDefaults(ctx context.Context, orgID ulid.ULID, userID ulid.ULID) error
	ExportSettings(ctx context.Context, orgID ulid.ULID, userID ulid.ULID) (map[string]interface{}, error)
	ImportSettings(ctx context.Context, orgID ulid.ULID, userID ulid.ULID, settings map[string]interface{}) error
}

