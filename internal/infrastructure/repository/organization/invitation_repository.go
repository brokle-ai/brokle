package organization

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	orgDomain "brokle/internal/core/domain/organization"
	"brokle/pkg/ulid"
)

// invitationRepository implements orgDomain.InvitationRepository using GORM
type invitationRepository struct {
	db *gorm.DB
}

// NewInvitationRepository creates a new invitation repository instance
func NewInvitationRepository(db *gorm.DB) orgDomain.InvitationRepository {
	return &invitationRepository{
		db: db,
	}
}

// Create creates a new invitation
func (r *invitationRepository) Create(ctx context.Context, invitation *orgDomain.Invitation) error {
	return r.db.WithContext(ctx).Create(invitation).Error
}

// GetByID retrieves an invitation by ID
func (r *invitationRepository) GetByID(ctx context.Context, id ulid.ULID) (*orgDomain.Invitation, error) {
	var invitation orgDomain.Invitation
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&invitation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("get invitation by ID %s: %w", id, orgDomain.ErrInvitationNotFound)
		}
		return nil, err
	}
	return &invitation, nil
}

// Update updates an invitation
func (r *invitationRepository) Update(ctx context.Context, invitation *orgDomain.Invitation) error {
	return r.db.WithContext(ctx).Save(invitation).Error
}

// GetByOrganizationID retrieves all invitations for an organization
func (r *invitationRepository) GetByOrganizationID(ctx context.Context, orgID ulid.ULID) ([]*orgDomain.Invitation, error) {
	var invitations []*orgDomain.Invitation
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("created_at DESC").
		Find(&invitations).Error
	return invitations, err
}

// GetByOrganizationAndStatus retrieves invitations by organization and status
func (r *invitationRepository) GetByOrganizationAndStatus(ctx context.Context, orgID ulid.ULID, status orgDomain.InvitationStatus) ([]*orgDomain.Invitation, error) {
	var invitations []*orgDomain.Invitation
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND status = ?", orgID, status).
		Order("created_at DESC").
		Find(&invitations).Error
	return invitations, err
}

// GetByUserID retrieves all invitations for a user
func (r *invitationRepository) GetByUserID(ctx context.Context, userID ulid.ULID) ([]*orgDomain.Invitation, error) {
	var invitations []*orgDomain.Invitation
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&invitations).Error
	return invitations, err
}

// GetByUserAndStatus retrieves invitations by user and status
func (r *invitationRepository) GetByUserAndStatus(ctx context.Context, userID ulid.ULID, status orgDomain.InvitationStatus) ([]*orgDomain.Invitation, error) {
	var invitations []*orgDomain.Invitation
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, status).
		Order("created_at DESC").
		Find(&invitations).Error
	return invitations, err
}

// GetPendingByEmail retrieves pending invitations by email
func (r *invitationRepository) GetPendingByEmail(ctx context.Context, orgID ulid.ULID, email string) (*orgDomain.Invitation, error) {
	var invitation orgDomain.Invitation
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND email = ? AND status = ?", orgID, email, orgDomain.InvitationStatusPending).
		First(&invitation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("get pending invitation by org %s and email %s: %w", orgID, email, orgDomain.ErrInvitationNotFound)
		}
		return nil, err
	}
	return &invitation, nil
}

// GetExpiredInvitations retrieves expired invitations
func (r *invitationRepository) GetExpiredInvitations(ctx context.Context) ([]*orgDomain.Invitation, error) {
	var invitations []*orgDomain.Invitation
	err := r.db.WithContext(ctx).
		Where("status = ? AND expires_at < ?", orgDomain.InvitationStatusPending, time.Now()).
		Find(&invitations).Error
	return invitations, err
}

// Delete soft deletes an invitation
func (r *invitationRepository) Delete(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).Model(&orgDomain.Invitation{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

// GetByToken retrieves an invitation by token
func (r *invitationRepository) GetByToken(ctx context.Context, token string) (*orgDomain.Invitation, error) {
	var invitation orgDomain.Invitation
	err := r.db.WithContext(ctx).Where("token = ? AND deleted_at IS NULL", token).First(&invitation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("get invitation by token: %w", orgDomain.ErrInvitationNotFound)
		}
		return nil, err
	}
	return &invitation, nil
}

// GetByEmail retrieves all invitations for an email address
func (r *invitationRepository) GetByEmail(ctx context.Context, email string) ([]*orgDomain.Invitation, error) {
	var invitations []*orgDomain.Invitation
	err := r.db.WithContext(ctx).
		Where("email = ? AND deleted_at IS NULL", email).
		Order("created_at DESC").
		Find(&invitations).Error
	return invitations, err
}

// GetPendingInvitations retrieves pending invitations for an organization
func (r *invitationRepository) GetPendingInvitations(ctx context.Context, orgID ulid.ULID) ([]*orgDomain.Invitation, error) {
	var invitations []*orgDomain.Invitation
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND status = ? AND deleted_at IS NULL", orgID, orgDomain.InvitationStatusPending).
		Order("created_at DESC").
		Find(&invitations).Error
	return invitations, err
}

// MarkAccepted marks an invitation as accepted
func (r *invitationRepository) MarkAccepted(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).
		Model(&orgDomain.Invitation{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      orgDomain.InvitationStatusAccepted,
			"accepted_at": time.Now(),
			"updated_at":  time.Now(),
		}).Error
}

// CleanupExpiredInvitations removes expired invitations
func (r *invitationRepository) CleanupExpiredInvitations(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("status = ? AND expires_at < ?", orgDomain.InvitationStatusPending, time.Now()).
		Delete(&orgDomain.Invitation{}).Error
}

// IsEmailAlreadyInvited checks if an email already has a pending invitation for an organization
func (r *invitationRepository) IsEmailAlreadyInvited(ctx context.Context, email string, orgID ulid.ULID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&orgDomain.Invitation{}).
		Where("organization_id = ? AND email = ? AND status = ? AND deleted_at IS NULL", orgID, email, orgDomain.InvitationStatusPending).
		Count(&count).Error
	return count > 0, err
}

// MarkExpired marks an invitation as expired
func (r *invitationRepository) MarkExpired(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).
		Model(&orgDomain.Invitation{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     orgDomain.InvitationStatusExpired,
			"updated_at": time.Now(),
		}).Error
}

// RevokeInvitation revokes an invitation
func (r *invitationRepository) RevokeInvitation(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).
		Model(&orgDomain.Invitation{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     orgDomain.InvitationStatusRevoked,
			"updated_at": time.Now(),
		}).Error
}