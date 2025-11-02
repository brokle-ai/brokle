package organization

import (
	"context"
	"fmt"
	"time"

	authDomain "brokle/internal/core/domain/auth"
	orgDomain "brokle/internal/core/domain/organization"
	userDomain "brokle/internal/core/domain/user"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/ulid"
)

// organizationService implements the orgDomain.OrganizationService interface
type organizationService struct {
	orgRepo     orgDomain.OrganizationRepository
	userRepo    userDomain.Repository
	memberSvc   orgDomain.MemberService
	projectSvc  orgDomain.ProjectService
	roleService authDomain.RoleService
}

// NewOrganizationService creates a new organization service instance
func NewOrganizationService(
	orgRepo orgDomain.OrganizationRepository,
	userRepo userDomain.Repository,
	memberSvc orgDomain.MemberService,
	projectSvc orgDomain.ProjectService,
	roleService authDomain.RoleService,
) orgDomain.OrganizationService {
	return &organizationService{
		orgRepo:     orgRepo,
		userRepo:    userRepo,
		memberSvc:   memberSvc,
		projectSvc:  projectSvc,
		roleService: roleService,
	}
}

// CreateOrganization creates a new organization with the user as owner
func (s *organizationService) CreateOrganization(ctx context.Context, userID ulid.ULID, req *orgDomain.CreateOrganizationRequest) (*orgDomain.Organization, error) {
	// Create organization (no slug - use ULID only)
	org := orgDomain.NewOrganization(req.Name)
	if req.BillingEmail != "" {
		org.BillingEmail = req.BillingEmail
	}

	err := s.orgRepo.Create(ctx, org)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to create organization", err)
	}

	// Get owner role for this organization
	ownerRole, err := s.roleService.GetRoleByNameAndScope(ctx, "owner", authDomain.ScopeOrganization)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to get owner role", err)
	}

	// Add creator as owner using member service
	err = s.memberSvc.AddMember(ctx, org.ID, userID, ownerRole.ID, userID)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to add user as organization owner", err)
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

	return org, nil
}

// GetOrganization retrieves organization by ID
func (s *organizationService) GetOrganization(ctx context.Context, orgID ulid.ULID) (*orgDomain.Organization, error) {
	return s.orgRepo.GetByID(ctx, orgID)
}

// GetOrganizationBySlug retrieves organization by slug
func (s *organizationService) GetOrganizationBySlug(ctx context.Context, slug string) (*orgDomain.Organization, error) {
	return s.orgRepo.GetBySlug(ctx, slug)
}

// UpdateOrganization updates organization details
func (s *organizationService) UpdateOrganization(ctx context.Context, orgID ulid.ULID, req *orgDomain.UpdateOrganizationRequest) error {
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return appErrors.NewNotFoundError("Organization not found")
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
		return appErrors.NewInternalError("Failed to update organization", err)
	}

	return nil
}

// DeleteOrganization soft deletes an organization
func (s *organizationService) DeleteOrganization(ctx context.Context, orgID ulid.ULID) error {
	// Verify organization exists before deletion
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return appErrors.NewNotFoundError("Organization not found")
	}

	err = s.orgRepo.Delete(ctx, orgID)
	if err != nil {
		return appErrors.NewInternalError("Failed to delete organization", err)
	}

	return nil
}

// ListOrganizations lists organizations with filters
func (s *organizationService) ListOrganizations(ctx context.Context, filters *orgDomain.OrganizationFilters) ([]*orgDomain.Organization, error) {
	return s.orgRepo.List(ctx, filters.Limit, filters.Offset)
}

// GetUserOrganizations returns organizations for a user
func (s *organizationService) GetUserOrganizations(ctx context.Context, userID ulid.ULID) ([]*orgDomain.Organization, error) {
	return s.orgRepo.GetOrganizationsByUserID(ctx, userID)
}

// GetUserDefaultOrganization returns user's default organization
func (s *organizationService) GetUserDefaultOrganization(ctx context.Context, userID ulid.ULID) (*orgDomain.Organization, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("User not found")
	}

	if user.DefaultOrganizationID == nil {
		return nil, appErrors.NewNotFoundError("User has no default organization")
	}

	return s.orgRepo.GetByID(ctx, *user.DefaultOrganizationID)
}

// SetUserDefaultOrganization sets user's default organization
func (s *organizationService) SetUserDefaultOrganization(ctx context.Context, userID, orgID ulid.ULID) error {
	// Verify user is member of organization using member service
	isMember, err := s.memberSvc.IsMember(ctx, userID, orgID)
	if err != nil {
		return appErrors.NewInternalError("Failed to check membership", err)
	}
	if !isMember {
		return appErrors.NewForbiddenError("User is not a member of this organization")
	}

	return s.userRepo.SetDefaultOrganization(ctx, userID, orgID)
}
