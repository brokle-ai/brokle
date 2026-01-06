package organization

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	authDomain "brokle/internal/core/domain/auth"
	billingDomain "brokle/internal/core/domain/billing"
	orgDomain "brokle/internal/core/domain/organization"
	userDomain "brokle/internal/core/domain/user"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/ulid"
)

// organizationService implements the orgDomain.OrganizationService interface
type organizationService struct {
	orgRepo           orgDomain.OrganizationRepository
	userRepo          userDomain.Repository
	memberSvc         orgDomain.MemberService
	projectSvc        orgDomain.ProjectService
	roleService       authDomain.RoleService
	billingRepo       billingDomain.OrganizationBillingRepository
	pricingConfigRepo billingDomain.PricingConfigRepository
	logger            *slog.Logger
}

func NewOrganizationService(
	orgRepo orgDomain.OrganizationRepository,
	userRepo userDomain.Repository,
	memberSvc orgDomain.MemberService,
	projectSvc orgDomain.ProjectService,
	roleService authDomain.RoleService,
	billingRepo billingDomain.OrganizationBillingRepository,
	pricingConfigRepo billingDomain.PricingConfigRepository,
	logger *slog.Logger,
) orgDomain.OrganizationService {
	return &organizationService{
		orgRepo:           orgRepo,
		userRepo:          userRepo,
		memberSvc:         memberSvc,
		projectSvc:        projectSvc,
		roleService:       roleService,
		billingRepo:       billingRepo,
		pricingConfigRepo: pricingConfigRepo,
		logger:            logger,
	}
}

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

	// Provision billing with Free plan
	if err := s.provisionBilling(ctx, org.ID); err != nil {
		s.logger.Error("failed to provision billing for organization",
			"error", err,
			"organization_id", org.ID,
		)
		// Don't fail org creation if billing provisioning fails - it can be retried
	}

	ownerRole, err := s.roleService.GetRoleByNameAndScope(ctx, "owner", authDomain.ScopeOrganization)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to get owner role", err)
	}

	err = s.memberSvc.AddMember(ctx, org.ID, userID, ownerRole.ID, userID)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to add user as organization owner", err)
	}

	// Set as user's default organization if they don't have one
	user, _ := s.userRepo.GetByID(ctx, userID)
	if user != nil && user.DefaultOrganizationID == nil {
		err = s.userRepo.SetDefaultOrganization(ctx, userID, org.ID)
		if err != nil {
			s.logger.Warn("failed to set default organization",
				"error", err,
				"user_id", userID,
				"organization_id", org.ID,
			)
		}
	}

	return org, nil
}

// provisionBilling creates a billing record for a new organization with the default pricing plan
func (s *organizationService) provisionBilling(ctx context.Context, orgID ulid.ULID) error {
	// Look up the default pricing plan (dynamically, not hardcoded)
	defaultPlan, err := s.pricingConfigRepo.GetDefault(ctx)
	if err != nil {
		return fmt.Errorf("get default pricing plan: %w", err)
	}

	now := time.Now()
	billingRecord := &billingDomain.OrganizationBilling{
		OrganizationID:        orgID,
		PricingConfigID:       defaultPlan.ID,
		BillingCycleStart:     now,
		BillingCycleAnchorDay: 1,
		// Free tier remaining (from default plan)
		FreeSpansRemaining:  defaultPlan.FreeSpans,
		FreeBytesRemaining:  int64(defaultPlan.FreeGB * 1024 * 1024 * 1024), // Convert GB to bytes
		FreeScoresRemaining: defaultPlan.FreeScores,
		CurrentPeriodSpans:  0,
		CurrentPeriodBytes:  0,
		CurrentPeriodScores: 0,
		CurrentPeriodCost:   0,
		LastSyncedAt:        now,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := s.billingRepo.Create(ctx, billingRecord); err != nil {
		return fmt.Errorf("create billing record: %w", err)
	}

	s.logger.Info("provisioned billing for organization",
		"organization_id", orgID,
		"pricing_plan", defaultPlan.Name,
	)

	return nil
}

func (s *organizationService) GetOrganization(ctx context.Context, orgID ulid.ULID) (*orgDomain.Organization, error) {
	return s.orgRepo.GetByID(ctx, orgID)
}

func (s *organizationService) GetOrganizationBySlug(ctx context.Context, slug string) (*orgDomain.Organization, error) {
	return s.orgRepo.GetBySlug(ctx, slug)
}

func (s *organizationService) UpdateOrganization(ctx context.Context, orgID ulid.ULID, req *orgDomain.UpdateOrganizationRequest) error {
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return appErrors.NewNotFoundError("Organization not found")
	}

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

func (s *organizationService) ListOrganizations(ctx context.Context, filters *orgDomain.OrganizationFilters) ([]*orgDomain.Organization, error) {
	return s.orgRepo.List(ctx, filters)
}

func (s *organizationService) GetUserOrganizations(ctx context.Context, userID ulid.ULID) ([]*orgDomain.Organization, error) {
	return s.orgRepo.GetOrganizationsByUserID(ctx, userID)
}

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

func (s *organizationService) GetUserOrganizationsWithProjects(
	ctx context.Context,
	userID ulid.ULID,
) ([]*orgDomain.OrganizationWithProjectsAndRole, error) {
	return s.orgRepo.GetUserOrganizationsWithProjectsBatch(ctx, userID)
}
