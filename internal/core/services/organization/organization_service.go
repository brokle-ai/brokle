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
	orgRepo     organization.OrganizationRepository
	memberRepo  organization.MemberRepository
	projectRepo organization.ProjectRepository
	envRepo     organization.EnvironmentRepository
	inviteRepo  organization.InvitationRepository
	userRepo    user.Repository
	roleService auth.RoleService
	auditRepo   auth.AuditLogRepository
}

// NewOrganizationService creates a new organization service instance
func NewOrganizationService(
	orgRepo organization.OrganizationRepository,
	memberRepo organization.MemberRepository,
	projectRepo organization.ProjectRepository,
	envRepo organization.EnvironmentRepository,
	inviteRepo organization.InvitationRepository,
	userRepo user.Repository,
	roleService auth.RoleService,
	auditRepo auth.AuditLogRepository,
) organization.OrganizationService {
	return &organizationService{
		orgRepo:     orgRepo,
		memberRepo:  memberRepo,
		projectRepo: projectRepo,
		envRepo:     envRepo,
		inviteRepo:  inviteRepo,
		userRepo:    userRepo,
		roleService: roleService,
		auditRepo:   auditRepo,
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

	// Get owner role
	ownerRole, err := s.roleService.GetRoleByName(ctx, nil, "owner")
	if err != nil {
		return nil, fmt.Errorf("failed to get owner role: %w", err)
	}

	// Add creator as owner
	member := organization.NewMember(org.ID, userID, ownerRole.ID)
	err = s.memberRepo.Create(ctx, member)
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

	// Create default environments
	err = s.CreateDefaultEnvironments(ctx, org.ID)
	if err != nil {
		// Log but don't fail organization creation
		fmt.Printf("Failed to create default environments: %v\n", err)
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
	// Verify user is member of organization
	isMember, err := s.memberRepo.IsMember(ctx, userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return errors.New("user is not a member of this organization")
	}

	return s.userRepo.SetDefaultOrganization(ctx, userID, orgID)
}

// CreateDefaultEnvironments creates default environments for a new organization
func (s *organizationService) CreateDefaultEnvironments(ctx context.Context, orgID ulid.ULID) error {
	// Create a default project first
	defaultProject := organization.NewProject(orgID, "Default", "default", "Default project for getting started")
	err := s.projectRepo.Create(ctx, defaultProject)
	if err != nil {
		return fmt.Errorf("failed to create default project: %w", err)
	}

	// Create default environments: development, staging, production
	environments := []struct {
		name, slug string
	}{
		{"Development", "dev"},
		{"Staging", "staging"},
		{"Production", "prod"},
	}

	for _, env := range environments {
		environment := organization.NewEnvironment(defaultProject.ID, env.name, env.slug)
		err = s.envRepo.Create(ctx, environment)
		if err != nil {
			return fmt.Errorf("failed to create %s environment: %w", env.name, err)
		}
	}

	return nil
}

// AddMember adds a user to an organization with specified role
func (s *organizationService) AddMember(ctx context.Context, orgID, userID, roleID ulid.ULID) error {
	// Verify organization exists
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify role exists
	role, err := s.roleService.GetRole(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Check if user is already a member
	isMember, err := s.memberRepo.IsMember(ctx, userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if isMember {
		return errors.New("user is already a member of this organization")
	}

	// Create member
	member := organization.NewMember(orgID, userID, roleID)
	err = s.memberRepo.Create(ctx, member)
	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, &orgID, "member.added", "organization", orgID.String(),
		fmt.Sprintf(`{"user_email": "%s", "role": "%s"}`, user.Email, role.Name), "", ""))

	return nil
}

// RemoveMember removes a user from an organization
func (s *organizationService) RemoveMember(ctx context.Context, orgID, userID ulid.ULID) error {
	// Verify membership exists
	member, err := s.memberRepo.GetByUserAndOrganization(ctx, userID, orgID)
	if err != nil {
		return fmt.Errorf("member not found: %w", err)
	}

	// Check if this is the only owner
	ownerRole, err := s.roleService.GetRoleByName(ctx, &orgID, "owner")
	if err != nil {
		return fmt.Errorf("failed to get owner role: %w", err)
	}

	if member.RoleID == ownerRole.ID {
		ownerCount, err := s.memberRepo.CountByOrganizationAndRole(ctx, orgID, ownerRole.ID)
		if err != nil {
			return fmt.Errorf("failed to count owners: %w", err)
		}
		if ownerCount <= 1 {
			return errors.New("cannot remove the last owner of the organization")
		}
	}

	// Remove member
	err = s.memberRepo.Delete(ctx, member.ID)
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// If this was their default organization, clear it
	user, err := s.userRepo.GetByID(ctx, userID)
	if err == nil && user.DefaultOrganizationID != nil && *user.DefaultOrganizationID == orgID {
		err = s.userRepo.SetDefaultOrganization(ctx, userID, ulid.ULID{})
		if err != nil {
			fmt.Printf("Failed to clear default organization: %v\n", err)
		}
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, &orgID, "member.removed", "organization", orgID.String(),
		fmt.Sprintf(`{"user_id": "%s"}`, userID.String()), "", ""))

	return nil
}

// UpdateMemberRole updates a member's role in an organization
func (s *organizationService) UpdateMemberRole(ctx context.Context, orgID, userID, newRoleID ulid.ULID) error {
	// Verify membership exists
	member, err := s.memberRepo.GetByUserAndOrganization(ctx, userID, orgID)
	if err != nil {
		return fmt.Errorf("member not found: %w", err)
	}

	// Verify new role exists
	newRole, err := s.roleService.GetRole(ctx, newRoleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Check if demoting the last owner
	ownerRole, err := s.roleService.GetRoleByName(ctx, &orgID, "owner")
	if err != nil {
		return fmt.Errorf("failed to get owner role: %w", err)
	}

	if member.RoleID == ownerRole.ID && newRoleID != ownerRole.ID {
		ownerCount, err := s.memberRepo.CountByOrganizationAndRole(ctx, orgID, ownerRole.ID)
		if err != nil {
			return fmt.Errorf("failed to count owners: %w", err)
		}
		if ownerCount <= 1 {
			return errors.New("cannot demote the last owner of the organization")
		}
	}

	// Update role
	member.RoleID = newRoleID
	member.UpdatedAt = time.Now()
	err = s.memberRepo.Update(ctx, member)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, &orgID, "member.role_updated", "organization", orgID.String(),
		fmt.Sprintf(`{"user_id": "%s", "new_role": "%s"}`, userID.String(), newRole.Name), "", ""))

	return nil
}

// GetOrganizationMembers returns all members of an organization
func (s *organizationService) GetOrganizationMembers(ctx context.Context, orgID ulid.ULID) ([]*organization.Member, error) {
	return s.memberRepo.GetByOrganizationID(ctx, orgID)
}

// GetMemberRole returns a user's role in an organization
func (s *organizationService) GetMemberRole(ctx context.Context, userID, orgID ulid.ULID) (*auth.Role, error) {
	member, err := s.memberRepo.GetByUserAndOrganization(ctx, userID, orgID)
	if err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	return s.roleService.GetRole(ctx, member.RoleID)
}

// IsMember checks if a user is a member of an organization
func (s *organizationService) IsMember(ctx context.Context, userID, orgID ulid.ULID) (bool, error) {
	return s.memberRepo.IsMember(ctx, userID, orgID)
}

// CreateProject creates a new project in an organization
func (s *organizationService) CreateProject(ctx context.Context, orgID ulid.ULID, req *organization.CreateProjectRequest) (*organization.Project, error) {
	// Verify organization exists
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Check if project with slug already exists in organization
	existing, _ := s.projectRepo.GetBySlug(ctx, orgID, req.Slug)
	if existing != nil {
		return nil, errors.New("project with this slug already exists in organization")
	}

	// Create project
	project := organization.NewProject(orgID, req.Name, req.Slug, req.Description)
	err = s.projectRepo.Create(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(nil, &orgID, "project.created", "project", project.ID.String(),
		fmt.Sprintf(`{"name": "%s", "slug": "%s"}`, project.Name, project.Slug), "", ""))

	return project, nil
}

// GetProject retrieves a project by ID
func (s *organizationService) GetProject(ctx context.Context, projectID ulid.ULID) (*organization.Project, error) {
	return s.projectRepo.GetByID(ctx, projectID)
}

// GetProjectBySlug retrieves a project by organization and slug
func (s *organizationService) GetProjectBySlug(ctx context.Context, orgID ulid.ULID, slug string) (*organization.Project, error) {
	return s.projectRepo.GetBySlug(ctx, orgID, slug)
}

// UpdateProject updates project details
func (s *organizationService) UpdateProject(ctx context.Context, projectID ulid.ULID, req *organization.UpdateProjectRequest) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}

	project.UpdatedAt = time.Now()

	err = s.projectRepo.Update(ctx, project)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(nil, &project.OrganizationID, "project.updated", "project", projectID.String(), "", "", ""))

	return nil
}

// DeleteProject soft deletes a project
func (s *organizationService) DeleteProject(ctx context.Context, projectID ulid.ULID) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	err = s.projectRepo.Delete(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(nil, &project.OrganizationID, "project.deleted", "project", projectID.String(),
		fmt.Sprintf(`{"name": "%s"}`, project.Name), "", ""))

	return nil
}

// ListOrganizationProjects lists all projects in an organization
func (s *organizationService) ListOrganizationProjects(ctx context.Context, orgID ulid.ULID) ([]*organization.Project, error) {
	return s.projectRepo.GetByOrganizationID(ctx, orgID)
}

// CreateEnvironment creates a new environment in a project
func (s *organizationService) CreateEnvironment(ctx context.Context, projectID ulid.ULID, req *organization.CreateEnvironmentRequest) (*organization.Environment, error) {
	// Verify project exists
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Check if environment with slug already exists in project
	existing, _ := s.envRepo.GetBySlug(ctx, projectID, req.Slug)
	if existing != nil {
		return nil, errors.New("environment with this slug already exists in project")
	}

	// Create environment
	environment := organization.NewEnvironment(projectID, req.Name, req.Slug)
	err = s.envRepo.Create(ctx, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(nil, &project.OrganizationID, "environment.created", "environment", environment.ID.String(),
		fmt.Sprintf(`{"name": "%s", "slug": "%s", "project": "%s"}`, environment.Name, environment.Slug, project.Name), "", ""))

	return environment, nil
}

// GetEnvironment retrieves an environment by ID
func (s *organizationService) GetEnvironment(ctx context.Context, envID ulid.ULID) (*organization.Environment, error) {
	return s.envRepo.GetByID(ctx, envID)
}

// GetEnvironmentBySlug retrieves an environment by project and slug
func (s *organizationService) GetEnvironmentBySlug(ctx context.Context, projectID ulid.ULID, slug string) (*organization.Environment, error) {
	return s.envRepo.GetBySlug(ctx, projectID, slug)
}

// UpdateEnvironment updates environment details
func (s *organizationService) UpdateEnvironment(ctx context.Context, envID ulid.ULID, req *organization.UpdateEnvironmentRequest) error {
	environment, err := s.envRepo.GetByID(ctx, envID)
	if err != nil {
		return fmt.Errorf("environment not found: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		environment.Name = *req.Name
	}

	environment.UpdatedAt = time.Now()

	err = s.envRepo.Update(ctx, environment)
	if err != nil {
		return fmt.Errorf("failed to update environment: %w", err)
	}

	// Get project for audit log
	project, _ := s.projectRepo.GetByID(ctx, environment.ProjectID)
	var orgID *ulid.ULID
	if project != nil {
		orgID = &project.OrganizationID
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(nil, orgID, "environment.updated", "environment", envID.String(), "", "", ""))

	return nil
}

// DeleteEnvironment soft deletes an environment
func (s *organizationService) DeleteEnvironment(ctx context.Context, envID ulid.ULID) error {
	environment, err := s.envRepo.GetByID(ctx, envID)
	if err != nil {
		return fmt.Errorf("environment not found: %w", err)
	}

	err = s.envRepo.Delete(ctx, envID)
	if err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}

	// Get project for audit log
	project, _ := s.projectRepo.GetByID(ctx, environment.ProjectID)
	var orgID *ulid.ULID
	if project != nil {
		orgID = &project.OrganizationID
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(nil, orgID, "environment.deleted", "environment", envID.String(),
		fmt.Sprintf(`{"name": "%s"}`, environment.Name), "", ""))

	return nil
}

// ListProjectEnvironments lists all environments in a project
func (s *organizationService) ListProjectEnvironments(ctx context.Context, projectID ulid.ULID) ([]*organization.Environment, error) {
	return s.envRepo.GetByProjectID(ctx, projectID)
}

// InviteUser creates an invitation for a user to join an organization
func (s *organizationService) InviteUser(ctx context.Context, orgID ulid.ULID, req *organization.InviteUserRequest) (*organization.Invitation, error) {
	// Verify organization exists
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Verify role exists
	role, err := s.roleService.GetRole(ctx, req.RoleID)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	// Note: We can't check if user is already a member since we only have email
	// This check would need to be done after user lookup by email

	// Check for existing pending invitation
	if req.Email != "" {
		existing, _ := s.inviteRepo.GetPendingByEmail(ctx, orgID, req.Email)
		if existing != nil {
			return nil, errors.New("invitation already exists for this email")
		}
	}

	// Create invitation
	token := ulid.New().String()
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	// TODO: Get inviter ID from auth context
	inviterID := ulid.New() // Placeholder - should come from authenticated user context
	invitation := organization.NewInvitation(orgID, req.RoleID, inviterID, req.Email, token, expiresAt)
	err = s.inviteRepo.Create(ctx, invitation)
	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(nil, &orgID, "invitation.created", "invitation", invitation.ID.String(),
		fmt.Sprintf(`{"email": "%s", "role": "%s"}`, req.Email, role.Name), "", ""))

	return invitation, nil
}

// AcceptInvitation accepts an invitation and adds the user to the organization
func (s *organizationService) AcceptInvitation(ctx context.Context, inviteID ulid.ULID, userID ulid.ULID) error {
	// Get invitation
	invitation, err := s.inviteRepo.GetByID(ctx, inviteID)
	if err != nil {
		return fmt.Errorf("invitation not found: %w", err)
	}

	if invitation.Status != organization.InvitationStatusPending {
		return errors.New("invitation is not pending")
	}

	if invitation.ExpiresAt.Before(time.Now()) {
		return errors.New("invitation has expired")
	}

	// Note: Email verification would happen at the UI/handler level
	// The invitation is email-based, not user-ID based

	// Check if user is already a member
	isMember, err := s.memberRepo.IsMember(ctx, userID, invitation.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if isMember {
		return errors.New("user is already a member of this organization")
	}

	// Add user as member
	member := organization.NewMember(invitation.OrganizationID, userID, invitation.RoleID)
	err = s.memberRepo.Create(ctx, member)
	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	// Mark invitation as accepted
	invitation.Status = organization.InvitationStatusAccepted
	invitation.AcceptedAt = &time.Time{}
	*invitation.AcceptedAt = time.Now()
	invitation.UpdatedAt = time.Now()
	err = s.inviteRepo.Update(ctx, invitation)
	if err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	// Set as default organization if user doesn't have one
	user, _ := s.userRepo.GetByID(ctx, userID)
	if user != nil && user.DefaultOrganizationID == nil {
		err = s.userRepo.SetDefaultOrganization(ctx, userID, invitation.OrganizationID)
		if err != nil {
			fmt.Printf("Failed to set default organization: %v\n", err)
		}
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, &invitation.OrganizationID, "invitation.accepted", "invitation", inviteID.String(),
		fmt.Sprintf(`{"user_id": "%s"}`, userID.String()), "", ""))

	return nil
}

// DeclineInvitation declines an invitation
func (s *organizationService) DeclineInvitation(ctx context.Context, inviteID ulid.ULID, userID ulid.ULID) error {
	// Get invitation
	invitation, err := s.inviteRepo.GetByID(ctx, inviteID)
	if err != nil {
		return fmt.Errorf("invitation not found: %w", err)
	}

	if invitation.Status != organization.InvitationStatusPending {
		return errors.New("invitation is not pending")
	}

	// Note: Email verification would happen at the UI/handler level
	// The invitation is email-based, not user-ID based

	// Mark invitation as declined
	invitation.Status = organization.InvitationStatusRevoked
	invitation.UpdatedAt = time.Now()
	err = s.inviteRepo.Update(ctx, invitation)
	if err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, &invitation.OrganizationID, "invitation.declined", "invitation", inviteID.String(),
		fmt.Sprintf(`{"user_id": "%s"}`, userID.String()), "", ""))

	return nil
}

// RevokeInvitation revokes a pending invitation
func (s *organizationService) RevokeInvitation(ctx context.Context, inviteID ulid.ULID) error {
	// Get invitation
	invitation, err := s.inviteRepo.GetByID(ctx, inviteID)
	if err != nil {
		return fmt.Errorf("invitation not found: %w", err)
	}

	if invitation.Status != organization.InvitationStatusPending {
		return errors.New("invitation is not pending")
	}

	// Mark invitation as revoked
	invitation.Status = organization.InvitationStatusRevoked
	invitation.UpdatedAt = time.Now()
	err = s.inviteRepo.Update(ctx, invitation)
	if err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(nil, &invitation.OrganizationID, "invitation.revoked", "invitation", inviteID.String(), "", "", ""))

	return nil
}

// GetInvitation retrieves an invitation by ID
func (s *organizationService) GetInvitation(ctx context.Context, inviteID ulid.ULID) (*organization.Invitation, error) {
	return s.inviteRepo.GetByID(ctx, inviteID)
}

// ListOrganizationInvitations lists all invitations for an organization
func (s *organizationService) ListOrganizationInvitations(ctx context.Context, orgID ulid.ULID, status *organization.InvitationStatus) ([]*organization.Invitation, error) {
	// For now, get all invitations and filter by status if needed
	// TODO: Add status filtering to repository if needed for performance
	invitations, err := s.inviteRepo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	
	if status != nil {
		filtered := make([]*organization.Invitation, 0)
		for _, invite := range invitations {
			if invite.Status == *status {
				filtered = append(filtered, invite)
			}
		}
		return filtered, nil
	}
	
	return invitations, nil
}

// ListUserInvitations lists all invitations for a user by their email
func (s *organizationService) ListUserInvitations(ctx context.Context, userID ulid.ULID, status *organization.InvitationStatus) ([]*organization.Invitation, error) {
	// TODO: This method should be modified to take email instead of userID
	// or we need to look up user email first, then get invitations by email
	// For now, return empty list as this method signature doesn't align with email-based invitations
	return []*organization.Invitation{}, nil
}