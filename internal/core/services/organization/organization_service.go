package organization

import (
	"context"
	"errors"
	"fmt"
	"time"

	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/organization"
	"brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
)

// organizationService implements the organization.OrganizationService interface
type organizationService struct {
	orgRepo        organization.OrganizationRepository
	userRepo       user.Repository
	memberSvc      organization.MemberService
	projectSvc     organization.ProjectService
	environmentSvc organization.EnvironmentService
	roleService    auth.RoleService
	auditRepo      auth.AuditLogRepository
}

// NewOrganizationService creates a new organization service instance
func NewOrganizationService(
	orgRepo organization.OrganizationRepository,
	userRepo user.Repository,
	memberSvc organization.MemberService,
	projectSvc organization.ProjectService,
	environmentSvc organization.EnvironmentService,
	roleService auth.RoleService,
	auditRepo auth.AuditLogRepository,
) organization.OrganizationService {
	return &organizationService{
		orgRepo:        orgRepo,
		userRepo:       userRepo,
		memberSvc:      memberSvc,
		projectSvc:     projectSvc,
		environmentSvc: environmentSvc,
		roleService:    roleService,
		auditRepo:      auditRepo,
	}
}

// CreateOrganization creates a new organization with the user as owner
func (s *organizationService) CreateOrganization(ctx context.Context, userID ulid.ULID, req *organization.CreateOrganizationRequest) (*organization.Organization, error) {
	// Check if organization with slug already exists
	existing, _ := s.orgRepo.GetBySlug(ctx, req.Slug)
	if existing != nil {
		return nil, errors.New("organization with this slug already exists")
	}

	// Create organization
	org := organization.NewOrganization(req.Name, req.Slug)
	if req.BillingEmail != "" {
		org.BillingEmail = req.BillingEmail
	}

	err := s.orgRepo.Create(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Get owner role for this organization
	ownerRole, err := s.roleService.GetRoleByNameAndScope(ctx, "owner", auth.ScopeOrganization)
	if err != nil {
		return nil, fmt.Errorf("failed to get owner role: %w", err)
	}

	// Add creator as owner using member service
	err = s.memberSvc.AddMember(ctx, org.ID, userID, ownerRole.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to add user as organization owner: %w", err)
	}

	// Set as user's default organization if they don't have one
	user, _ := s.userRepo.GetByID(ctx, userID)
	if user != nil && user.DefaultOrganizationID == nil {
		err = s.userRepo.SetDefaultOrganization(ctx, userID, org.ID)
		if err != nil {
			// Log but don't fail
			fmt.Printf("Failed to set default organization: %v\n", err)
		}
	}

	// Create default project and environments
	err = s.createDefaultProjectAndEnvironments(ctx, org.ID)
	if err != nil {
		// Log but don't fail organization creation
		fmt.Printf("Failed to create default project and environments: %v\n", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, &org.ID, "organization.created", "organization", org.ID.String(), 
		fmt.Sprintf(`{"name": "%s", "slug": "%s"}`, org.Name, org.Slug), "", ""))

	return org, nil
}

// GetOrganization retrieves organization by ID
func (s *organizationService) GetOrganization(ctx context.Context, orgID ulid.ULID) (*organization.Organization, error) {
	return s.orgRepo.GetByID(ctx, orgID)
}

// GetOrganizationBySlug retrieves organization by slug
func (s *organizationService) GetOrganizationBySlug(ctx context.Context, slug string) (*organization.Organization, error) {
	return s.orgRepo.GetBySlug(ctx, slug)
}

// UpdateOrganization updates organization details
func (s *organizationService) UpdateOrganization(ctx context.Context, orgID ulid.ULID, req *organization.UpdateOrganizationRequest) error {
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		org.Name = *req.Name
	}
	if req.BillingEmail != nil {
		org.BillingEmail = *req.BillingEmail
	}
	if req.Plan != nil {
		org.Plan = *req.Plan
	}

	org.UpdatedAt = time.Now()

	err = s.orgRepo.Update(ctx, org)
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(nil, &orgID, "organization.updated", "organization", orgID.String(), "", "", ""))

	return nil
}

// DeleteOrganization soft deletes an organization
func (s *organizationService) DeleteOrganization(ctx context.Context, orgID ulid.ULID) error {
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	err = s.orgRepo.Delete(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(nil, &orgID, "organization.deleted", "organization", orgID.String(), 
		fmt.Sprintf(`{"name": "%s"}`, org.Name), "", ""))

	return nil
}

// ListOrganizations lists organizations with filters
func (s *organizationService) ListOrganizations(ctx context.Context, filters *organization.OrganizationFilters) ([]*organization.Organization, error) {
	return s.orgRepo.List(ctx, filters.Limit, filters.Offset)
}

// GetUserOrganizations returns organizations for a user
func (s *organizationService) GetUserOrganizations(ctx context.Context, userID ulid.ULID) ([]*organization.Organization, error) {
	return s.orgRepo.GetOrganizationsByUserID(ctx, userID)
}

// GetUserDefaultOrganization returns user's default organization
func (s *organizationService) GetUserDefaultOrganization(ctx context.Context, userID ulid.ULID) (*organization.Organization, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user.DefaultOrganizationID == nil {
		return nil, errors.New("user has no default organization")
	}

	return s.orgRepo.GetByID(ctx, *user.DefaultOrganizationID)
}

// SetUserDefaultOrganization sets user's default organization
func (s *organizationService) SetUserDefaultOrganization(ctx context.Context, userID, orgID ulid.ULID) error {
	// Verify user is member of organization using member service
	isMember, err := s.memberSvc.IsMember(ctx, userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return errors.New("user is not a member of this organization")
	}

	return s.userRepo.SetDefaultOrganization(ctx, userID, orgID)
}

// createDefaultProjectAndEnvironments creates default project and environments for a new organization
func (s *organizationService) createDefaultProjectAndEnvironments(ctx context.Context, orgID ulid.ULID) error {
	// Create default project using project service
	defaultProjectReq := &organization.CreateProjectRequest{
		Name:        "Default",
		Slug:        "default",
		Description: "Default project for getting started",
	}
	
	project, err := s.projectSvc.CreateProject(ctx, orgID, defaultProjectReq)
	if err != nil {
		return fmt.Errorf("failed to create default project: %w", err)
	}

	// Create default environments using environment service
	err = s.environmentSvc.CreateDefaultEnvironments(ctx, project.ID)
	if err != nil {
		return fmt.Errorf("failed to create default environments: %w", err)
	}

	return nil
}