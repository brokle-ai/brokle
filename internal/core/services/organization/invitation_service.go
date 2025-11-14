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

// invitationService implements the orgDomain.InvitationService interface
type invitationService struct {
	inviteRepo  orgDomain.InvitationRepository
	orgRepo     orgDomain.OrganizationRepository
	memberRepo  orgDomain.MemberRepository
	userRepo    userDomain.Repository
	roleService authDomain.RoleService
}

// NewInvitationService creates a new invitation service instance
func NewInvitationService(
	inviteRepo orgDomain.InvitationRepository,
	orgRepo orgDomain.OrganizationRepository,
	memberRepo orgDomain.MemberRepository,
	userRepo userDomain.Repository,
	roleService authDomain.RoleService,
) orgDomain.InvitationService {
	return &invitationService{
		inviteRepo:  inviteRepo,
		orgRepo:     orgRepo,
		memberRepo:  memberRepo,
		userRepo:    userRepo,
		roleService: roleService,
	}
}

// InviteUser creates an invitation for a user to join an organization
func (s *invitationService) InviteUser(ctx context.Context, orgID ulid.ULID, inviterID ulid.ULID, req *orgDomain.InviteUserRequest) (*orgDomain.Invitation, error) {
	// Verify organization exists
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Organization not found")
	}

	// Verify role exists
	_, err = s.roleService.GetRoleByID(ctx, req.RoleID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Role not found")
	}

	// Check for existing pending invitation
	if req.Email != "" {
		existing, _ := s.inviteRepo.GetPendingByEmail(ctx, orgID, req.Email)
		if existing != nil {
			return nil, appErrors.NewConflictError("Invitation already exists for this email")
		}
	}

	// Create invitation
	token := ulid.New().String()
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	invitation := orgDomain.NewInvitation(orgID, req.RoleID, inviterID, req.Email, token, expiresAt)
	err = s.inviteRepo.Create(ctx, invitation)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to create invitation", err)
	}

	return invitation, nil
}

// AcceptInvitation accepts an invitation and adds the user to the organization
func (s *invitationService) AcceptInvitation(ctx context.Context, token string, userID ulid.ULID) error {
	// Get invitation by token
	invitation, err := s.inviteRepo.GetByToken(ctx, token)
	if err != nil {
		return appErrors.NewNotFoundError("Invitation not found")
	}

	if invitation.Status != orgDomain.InvitationStatusPending {
		return appErrors.NewValidationError("status", "Invitation is not pending")
	}

	if invitation.ExpiresAt.Before(time.Now()) {
		return appErrors.NewValidationError("expiry", "Invitation has expired")
	}

	// Check if user is already a member
	isMember, err := s.memberRepo.IsMember(ctx, userID, invitation.OrganizationID)
	if err != nil {
		return appErrors.NewInternalError("Failed to check membership", err)
	}
	if isMember {
		return appErrors.NewConflictError("User is already a member of this organization")
	}

	// Add user as member
	member := orgDomain.NewMember(invitation.OrganizationID, userID, invitation.RoleID)
	err = s.memberRepo.Create(ctx, member)
	if err != nil {
		return appErrors.NewInternalError("Failed to add member", err)
	}

	// Mark invitation as accepted
	invitation.Status = orgDomain.InvitationStatusAccepted
	invitation.AcceptedAt = &time.Time{}
	*invitation.AcceptedAt = time.Now()
	invitation.UpdatedAt = time.Now()
	err = s.inviteRepo.Update(ctx, invitation)
	if err != nil {
		return appErrors.NewInternalError("Failed to update invitation", err)
	}

	// Set as default organization if user doesn't have one
	user, _ := s.userRepo.GetByID(ctx, userID)
	if user != nil && user.DefaultOrganizationID == nil {
		err = s.userRepo.SetDefaultOrganization(ctx, userID, invitation.OrganizationID)
		if err != nil {
			fmt.Printf("Failed to set default organization: %v\n", err)
		}
	}

	return nil
}

// DeclineInvitation declines an invitation
func (s *invitationService) DeclineInvitation(ctx context.Context, token string) error {
	// Get invitation by token
	invitation, err := s.inviteRepo.GetByToken(ctx, token)
	if err != nil {
		return appErrors.NewNotFoundError("Invitation not found")
	}

	if invitation.Status != orgDomain.InvitationStatusPending {
		return appErrors.NewValidationError("status", "Invitation is not pending")
	}

	// Mark invitation as declined
	invitation.Status = orgDomain.InvitationStatusRevoked
	invitation.UpdatedAt = time.Now()
	err = s.inviteRepo.Update(ctx, invitation)
	if err != nil {
		return appErrors.NewInternalError("Failed to update invitation", err)
	}

	return nil
}

// RevokeInvitation revokes a pending invitation
func (s *invitationService) RevokeInvitation(ctx context.Context, invitationID ulid.ULID, revokedByID ulid.ULID) error {
	// Get invitation
	invitation, err := s.inviteRepo.GetByID(ctx, invitationID)
	if err != nil {
		return appErrors.NewNotFoundError("Invitation not found")
	}

	if invitation.Status != orgDomain.InvitationStatusPending {
		return appErrors.NewValidationError("status", "Invitation is not pending")
	}

	// Mark invitation as revoked
	invitation.Status = orgDomain.InvitationStatusRevoked
	invitation.UpdatedAt = time.Now()
	err = s.inviteRepo.Update(ctx, invitation)
	if err != nil {
		return appErrors.NewInternalError("Failed to update invitation", err)
	}

	return nil
}

// ResendInvitation resends a pending invitation
func (s *invitationService) ResendInvitation(ctx context.Context, invitationID ulid.ULID, resentByID ulid.ULID) error {
	// Get invitation
	invitation, err := s.inviteRepo.GetByID(ctx, invitationID)
	if err != nil {
		return appErrors.NewNotFoundError("Invitation not found")
	}

	if invitation.Status != orgDomain.InvitationStatusPending {
		return appErrors.NewValidationError("status", "Invitation is not pending")
	}

	// Update expiration time
	invitation.ExpiresAt = time.Now().Add(7 * 24 * time.Hour) // 7 days from now
	invitation.UpdatedAt = time.Now()
	err = s.inviteRepo.Update(ctx, invitation)
	if err != nil {
		return appErrors.NewInternalError("Failed to update invitation", err)
	}

	return nil
}

// GetInvitation retrieves an invitation by ID
func (s *invitationService) GetInvitation(ctx context.Context, invitationID ulid.ULID) (*orgDomain.Invitation, error) {
	return s.inviteRepo.GetByID(ctx, invitationID)
}

// GetInvitationByToken retrieves an invitation by token
func (s *invitationService) GetInvitationByToken(ctx context.Context, token string) (*orgDomain.Invitation, error) {
	return s.inviteRepo.GetByToken(ctx, token)
}

// GetPendingInvitations retrieves all pending invitations for an organization
func (s *invitationService) GetPendingInvitations(ctx context.Context, orgID ulid.ULID) ([]*orgDomain.Invitation, error) {
	allInvitations, err := s.inviteRepo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	var pendingInvitations []*orgDomain.Invitation
	for _, invitation := range allInvitations {
		if invitation.Status == orgDomain.InvitationStatusPending {
			pendingInvitations = append(pendingInvitations, invitation)
		}
	}

	return pendingInvitations, nil
}

// GetUserInvitations retrieves all invitations for a user by email
func (s *invitationService) GetUserInvitations(ctx context.Context, email string) ([]*orgDomain.Invitation, error) {
	return s.inviteRepo.GetByEmail(ctx, email)
}

// ValidateInvitationToken validates an invitation token and returns the invitation
func (s *invitationService) ValidateInvitationToken(ctx context.Context, token string) (*orgDomain.Invitation, error) {
	invitation, err := s.inviteRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Invitation not found")
	}

	if invitation.Status != orgDomain.InvitationStatusPending {
		return nil, appErrors.NewValidationError("status", "Invitation is not pending")
	}

	if invitation.ExpiresAt.Before(time.Now()) {
		return nil, appErrors.NewValidationError("expiry", "Invitation has expired")
	}

	return invitation, nil
}

// IsEmailAlreadyInvited checks if an email already has a pending invitation for an organization
func (s *invitationService) IsEmailAlreadyInvited(ctx context.Context, email string, orgID ulid.ULID) (bool, error) {
	existing, err := s.inviteRepo.GetPendingByEmail(ctx, orgID, email)
	if err != nil {
		return false, err
	}
	return existing != nil, nil
}

// CleanupExpiredInvitations removes expired invitations
func (s *invitationService) CleanupExpiredInvitations(ctx context.Context) error {
	// This would typically be called by a background job
	// For now, return nil as the repository doesn't have a bulk delete method
	// This could be implemented by getting all invitations and filtering/deleting expired ones
	return nil
}
