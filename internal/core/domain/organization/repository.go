package organization

import (
	"context"

	"brokle/pkg/ulid"
)

// OrganizationRepository defines the interface for organization data access.
type OrganizationRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, org *Organization) error
	GetByID(ctx context.Context, id ulid.ULID) (*Organization, error)
	GetBySlug(ctx context.Context, slug string) (*Organization, error)
	Update(ctx context.Context, org *Organization) error
	Delete(ctx context.Context, id ulid.ULID) error
	List(ctx context.Context, limit, offset int) ([]*Organization, error)
	
	// User context
	GetOrganizationsByUserID(ctx context.Context, userID ulid.ULID) ([]*Organization, error)
}

// MemberRepository defines the interface for organization member data access.
type MemberRepository interface {
	// Member management
	Create(ctx context.Context, member *Member) error
	GetByID(ctx context.Context, id ulid.ULID) (*Member, error)
	GetByUserAndOrg(ctx context.Context, userID, orgID ulid.ULID) (*Member, error)
	GetByUserAndOrganization(ctx context.Context, userID, orgID ulid.ULID) (*Member, error) // Alias for GetByUserAndOrg
	Update(ctx context.Context, member *Member) error
	Delete(ctx context.Context, id ulid.ULID) error
	DeleteByUserAndOrg(ctx context.Context, orgID, userID ulid.ULID) error
	
	// Organization members
	GetMembersByOrganizationID(ctx context.Context, orgID ulid.ULID) ([]*Member, error)
	GetByOrganizationID(ctx context.Context, orgID ulid.ULID) ([]*Member, error) // Alias for GetMembersByOrganizationID
	GetMembersByUserID(ctx context.Context, userID ulid.ULID) ([]*Member, error)
	
	// Role operations
	UpdateMemberRole(ctx context.Context, orgID, userID, roleID ulid.ULID) error
	GetMemberRole(ctx context.Context, userID, orgID ulid.ULID) (ulid.ULID, error)
	CountByOrganizationAndRole(ctx context.Context, orgID, roleID ulid.ULID) (int, error)
	
	// Membership validation
	IsMember(ctx context.Context, userID, orgID ulid.ULID) (bool, error)
	GetMemberCount(ctx context.Context, orgID ulid.ULID) (int, error)
}

// ProjectRepository defines the interface for project data access.
type ProjectRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id ulid.ULID) (*Project, error)
	GetBySlug(ctx context.Context, orgID ulid.ULID, slug string) (*Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id ulid.ULID) error
	
	// Organization scoped
	GetByOrganizationID(ctx context.Context, orgID ulid.ULID) ([]*Project, error)
	GetProjectCount(ctx context.Context, orgID ulid.ULID) (int, error)
	
	// Access validation
	CanUserAccessProject(ctx context.Context, userID, projectID ulid.ULID) (bool, error)
}

// EnvironmentRepository defines the interface for environment data access.
type EnvironmentRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, env *Environment) error
	GetByID(ctx context.Context, id ulid.ULID) (*Environment, error)
	GetBySlug(ctx context.Context, projectID ulid.ULID, slug string) (*Environment, error)
	Update(ctx context.Context, env *Environment) error
	Delete(ctx context.Context, id ulid.ULID) error
	
	// Project scoped
	GetByProjectID(ctx context.Context, projectID ulid.ULID) ([]*Environment, error)
	GetEnvironmentCount(ctx context.Context, projectID ulid.ULID) (int, error)
	
	// Access validation
	CanUserAccessEnvironment(ctx context.Context, userID, envID ulid.ULID) (bool, error)
	GetEnvironmentOrganization(ctx context.Context, envID ulid.ULID) (ulid.ULID, error)
}

// InvitationRepository defines the interface for user invitation data access.
type InvitationRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, invitation *Invitation) error
	GetByID(ctx context.Context, id ulid.ULID) (*Invitation, error)
	GetByToken(ctx context.Context, token string) (*Invitation, error)
	Update(ctx context.Context, invitation *Invitation) error
	Delete(ctx context.Context, id ulid.ULID) error
	
	// Organization invitations
	GetByOrganizationID(ctx context.Context, orgID ulid.ULID) ([]*Invitation, error)
	GetByEmail(ctx context.Context, email string) ([]*Invitation, error)
	GetPendingByEmail(ctx context.Context, orgID ulid.ULID, email string) (*Invitation, error)
	GetPendingInvitations(ctx context.Context, orgID ulid.ULID) ([]*Invitation, error)
	
	// Invitation management
	MarkAccepted(ctx context.Context, id ulid.ULID) error
	MarkExpired(ctx context.Context, id ulid.ULID) error
	RevokeInvitation(ctx context.Context, id ulid.ULID) error
	CleanupExpiredInvitations(ctx context.Context) error
	
	// Validation
	IsEmailAlreadyInvited(ctx context.Context, email string, orgID ulid.ULID) (bool, error)
}

// OrganizationFilters represents filters for organization queries.
type OrganizationFilters struct {
	Name   *string
	Plan   *string
	Status *string
	Limit  int
	Offset int
}

// MemberFilters represents filters for member queries.
type MemberFilters struct {
	OrganizationID *ulid.ULID
	UserID         *ulid.ULID
	RoleID         *ulid.ULID
	Limit          int
	Offset         int
}

// ProjectFilters represents filters for project queries.
type ProjectFilters struct {
	OrganizationID *ulid.ULID
	Name           *string
	Limit          int
	Offset         int
}

// InvitationFilters represents filters for invitation queries.
type InvitationFilters struct {
	OrganizationID *ulid.ULID
	Status         *string
	Email          *string
	Limit          int
	Offset         int
}

// OrganizationSettingsRepository defines the interface for organization settings data access.
type OrganizationSettingsRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, setting *OrganizationSettings) error
	GetByID(ctx context.Context, id ulid.ULID) (*OrganizationSettings, error)
	GetByKey(ctx context.Context, orgID ulid.ULID, key string) (*OrganizationSettings, error)
	Update(ctx context.Context, setting *OrganizationSettings) error
	Delete(ctx context.Context, id ulid.ULID) error
	
	// Organization scoped operations
	GetAllByOrganizationID(ctx context.Context, orgID ulid.ULID) ([]*OrganizationSettings, error)
	GetSettingsMap(ctx context.Context, orgID ulid.ULID) (map[string]interface{}, error)
	DeleteByKey(ctx context.Context, orgID ulid.ULID, key string) error
	UpsertSetting(ctx context.Context, orgID ulid.ULID, key string, value interface{}) (*OrganizationSettings, error)
	
	// Bulk operations
	CreateMultiple(ctx context.Context, settings []*OrganizationSettings) error
	GetByKeys(ctx context.Context, orgID ulid.ULID, keys []string) ([]*OrganizationSettings, error)
	DeleteMultiple(ctx context.Context, orgID ulid.ULID, keys []string) error
}

// Repository aggregates all organization-related repositories.
type Repository interface {
	Organizations() OrganizationRepository
	Members() MemberRepository
	Projects() ProjectRepository
	Environments() EnvironmentRepository
	Invitations() InvitationRepository
	Settings() OrganizationSettingsRepository
}