package organization

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"brokle/internal/core/domain/organization"
	"brokle/pkg/ulid"
)

// environmentRepository implements organization.EnvironmentRepository using GORM
type environmentRepository struct {
	db *gorm.DB
}

// NewEnvironmentRepository creates a new environment repository instance
func NewEnvironmentRepository(db *gorm.DB) organization.EnvironmentRepository {
	return &environmentRepository{
		db: db,
	}
}

// Create creates a new environment
func (r *environmentRepository) Create(ctx context.Context, env *organization.Environment) error {
	return r.db.WithContext(ctx).Create(env).Error
}

// GetByID retrieves an environment by ID
func (r *environmentRepository) GetByID(ctx context.Context, id ulid.ULID) (*organization.Environment, error) {
	var env organization.Environment
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&env).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("environment not found")
		}
		return nil, err
	}
	return &env, nil
}

// GetBySlug retrieves an environment by project and slug
func (r *environmentRepository) GetBySlug(ctx context.Context, projectID ulid.ULID, slug string) (*organization.Environment, error) {
	var env organization.Environment
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND slug = ? AND deleted_at IS NULL", projectID, slug).
		First(&env).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("environment not found")
		}
		return nil, err
	}
	return &env, nil
}

// Update updates an environment
func (r *environmentRepository) Update(ctx context.Context, env *organization.Environment) error {
	return r.db.WithContext(ctx).Save(env).Error
}

// Delete soft deletes an environment
func (r *environmentRepository) Delete(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).Model(&organization.Environment{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

// GetByProjectID retrieves all environments in a project
func (r *environmentRepository) GetByProjectID(ctx context.Context, projectID ulid.ULID) ([]*organization.Environment, error) {
	var environments []*organization.Environment
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		Order("created_at ASC").
		Find(&environments).Error
	return environments, err
}

// CountByProject counts environments in a project
func (r *environmentRepository) CountByProject(ctx context.Context, projectID ulid.ULID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&organization.Environment{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		Count(&count).Error
	return count, err
}

// GetEnvironmentCount returns the count of environments in a project
func (r *environmentRepository) GetEnvironmentCount(ctx context.Context, projectID ulid.ULID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&organization.Environment{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		Count(&count).Error
	return int(count), err
}

// CanUserAccessEnvironment checks if a user has access to an environment
func (r *environmentRepository) CanUserAccessEnvironment(ctx context.Context, userID, envID ulid.ULID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("environments").
		Joins("JOIN projects ON environments.project_id = projects.id").
		Joins("JOIN organization_members ON projects.organization_id = organization_members.organization_id").
		Where("environments.id = ? AND organization_members.user_id = ? AND environments.deleted_at IS NULL", envID, userID).
		Count(&count).Error
	return count > 0, err
}

// GetEnvironmentOrganization returns the organization ID that owns an environment
func (r *environmentRepository) GetEnvironmentOrganization(ctx context.Context, envID ulid.ULID) (ulid.ULID, error) {
	var orgID ulid.ULID
	err := r.db.WithContext(ctx).
		Table("environments").
		Select("projects.organization_id").
		Joins("JOIN projects ON environments.project_id = projects.id").
		Where("environments.id = ? AND environments.deleted_at IS NULL", envID).
		Pluck("projects.organization_id", &orgID).Error
	return orgID, err
}