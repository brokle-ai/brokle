package organization

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"brokle/internal/core/domain/organization"
	"brokle/pkg/ulid"
)

// memberRepository implements organization.MemberRepository using GORM
type memberRepository struct {
	db *gorm.DB
}

// NewMemberRepository creates a new member repository instance
func NewMemberRepository(db *gorm.DB) organization.MemberRepository {
	return &memberRepository{
		db: db,
	}
}

// Create creates a new member
func (r *memberRepository) Create(ctx context.Context, member *organization.Member) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// GetByID retrieves a member by ID
func (r *memberRepository) GetByID(ctx context.Context, id ulid.ULID) (*organization.Member, error) {
	var member organization.Member
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("member not found")
		}
		return nil, err
	}
	return &member, nil
}

// GetByUserAndOrganization retrieves a member by user and organization
func (r *memberRepository) GetByUserAndOrganization(ctx context.Context, userID, orgID ulid.ULID) (*organization.Member, error) {
	var member organization.Member
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND organization_id = ? AND deleted_at IS NULL", userID, orgID).
		First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("member not found")
		}
		return nil, err
	}
	return &member, nil
}

// Update updates a member
func (r *memberRepository) Update(ctx context.Context, member *organization.Member) error {
	return r.db.WithContext(ctx).Save(member).Error
}

// Delete soft deletes a member
func (r *memberRepository) Delete(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).Model(&organization.Member{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

// GetByOrganizationID retrieves all members of an organization
func (r *memberRepository) GetByOrganizationID(ctx context.Context, orgID ulid.ULID) ([]*organization.Member, error) {
	var members []*organization.Member
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND deleted_at IS NULL", orgID).
		Order("created_at ASC").
		Find(&members).Error
	return members, err
}

// GetByUserID retrieves all memberships for a user
func (r *memberRepository) GetByUserID(ctx context.Context, userID ulid.ULID) ([]*organization.Member, error) {
	var members []*organization.Member
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at ASC").
		Find(&members).Error
	return members, err
}

// IsMember checks if a user is a member of an organization
func (r *memberRepository) IsMember(ctx context.Context, userID, orgID ulid.ULID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&organization.Member{}).
		Where("user_id = ? AND organization_id = ? AND deleted_at IS NULL", userID, orgID).
		Count(&count).Error
	return count > 0, err
}

// CountByOrganization counts members in an organization
func (r *memberRepository) CountByOrganization(ctx context.Context, orgID ulid.ULID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&organization.Member{}).
		Where("organization_id = ? AND deleted_at IS NULL", orgID).
		Count(&count).Error
	return count, err
}

// GetByUserAndOrg is an alias for GetByUserAndOrganization for interface compliance
func (r *memberRepository) GetByUserAndOrg(ctx context.Context, userID, orgID ulid.ULID) (*organization.Member, error) {
	return r.GetByUserAndOrganization(ctx, userID, orgID)
}

// CountByOrganizationAndRole counts members with a specific role in an organization
func (r *memberRepository) CountByOrganizationAndRole(ctx context.Context, orgID, roleID ulid.ULID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&organization.Member{}).
		Where("organization_id = ? AND role_id = ? AND deleted_at IS NULL", orgID, roleID).
		Count(&count).Error
	return int(count), err
}

// GetByOrganizationAndRole retrieves members with a specific role in an organization
func (r *memberRepository) GetByOrganizationAndRole(ctx context.Context, orgID, roleID ulid.ULID) ([]*organization.Member, error) {
	var members []*organization.Member
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND role_id = ? AND deleted_at IS NULL", orgID, roleID).
		Order("created_at ASC").
		Find(&members).Error
	return members, err
}

// GetMembersByOrganizationID is an alias for GetByOrganizationID for interface compliance
func (r *memberRepository) GetMembersByOrganizationID(ctx context.Context, orgID ulid.ULID) ([]*organization.Member, error) {
	return r.GetByOrganizationID(ctx, orgID)
}

// GetMembersByUserID is an alias for GetByUserID for interface compliance
func (r *memberRepository) GetMembersByUserID(ctx context.Context, userID ulid.ULID) ([]*organization.Member, error) {
	return r.GetByUserID(ctx, userID)
}

// GetMemberCount is an alias for CountByOrganization for interface compliance
func (r *memberRepository) GetMemberCount(ctx context.Context, orgID ulid.ULID) (int, error) {
	count, err := r.CountByOrganization(ctx, orgID)
	return int(count), err
}

// UpdateMemberRole updates the role of a member
func (r *memberRepository) UpdateMemberRole(ctx context.Context, orgID, userID, roleID ulid.ULID) error {
	return r.db.WithContext(ctx).
		Model(&organization.Member{}).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Update("role_id", roleID).Error
}

// GetMemberRole retrieves the role of a member
func (r *memberRepository) GetMemberRole(ctx context.Context, userID, orgID ulid.ULID) (ulid.ULID, error) {
	var member organization.Member
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		First(&member).Error
	if err != nil {
		return ulid.ULID{}, err
	}
	return member.RoleID, nil
}