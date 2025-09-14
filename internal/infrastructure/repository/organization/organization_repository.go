package organization

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	orgDomain "brokle/internal/core/domain/organization"
	"brokle/pkg/ulid"
)

// organizationRepository implements orgDomain.OrganizationRepository using GORM
type organizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository creates a new organization repository instance
func NewOrganizationRepository(db *gorm.DB) orgDomain.OrganizationRepository {
	return &organizationRepository{
		db: db,
	}
}

// Create creates a new organization
func (r *organizationRepository) Create(ctx context.Context, org *orgDomain.Organization) error {
	return r.db.WithContext(ctx).Create(org).Error
}

// GetByID retrieves an organization by ID
func (r *organizationRepository) GetByID(ctx context.Context, id ulid.ULID) (*orgDomain.Organization, error) {
	var org orgDomain.Organization
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&org).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("get organization by ID %s: %w", id, orgDomain.ErrNotFound)
		}
		return nil, err
	}
	return &org, nil
}

// GetBySlug retrieves an organization by slug
func (r *organizationRepository) GetBySlug(ctx context.Context, slug string) (*orgDomain.Organization, error) {
	var org orgDomain.Organization
	err := r.db.WithContext(ctx).Where("slug = ? AND deleted_at IS NULL", slug).First(&org).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("get organization by slug %s: %w", slug, orgDomain.ErrNotFound)
		}
		return nil, err
	}
	return &org, nil
}

// Update updates an organization
func (r *organizationRepository) Update(ctx context.Context, org *orgDomain.Organization) error {
	return r.db.WithContext(ctx).Save(org).Error
}

// Delete soft deletes an organization
func (r *organizationRepository) Delete(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).Model(&orgDomain.Organization{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

// List retrieves organizations with pagination
func (r *organizationRepository) List(ctx context.Context, limit, offset int) ([]*orgDomain.Organization, error) {
	var orgs []*orgDomain.Organization
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&orgs).Error
	return orgs, err
}

// GetOrganizationsByUserID retrieves organizations for a user
func (r *organizationRepository) GetOrganizationsByUserID(ctx context.Context, userID ulid.ULID) ([]*orgDomain.Organization, error) {
	var orgs []*orgDomain.Organization
	err := r.db.WithContext(ctx).
		Table("organizations").
		Select("organizations.*").
		Joins("JOIN organization_members ON organizations.id = organization_members.organization_id").
		Where("organization_members.user_id = ? AND organizations.deleted_at IS NULL", userID).
		Order("organizations.created_at DESC").
		Find(&orgs).Error
	return orgs, err
}
