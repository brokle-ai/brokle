package auth

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// permissionRepository implements auth.PermissionRepository using GORM
type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new permission repository instance
func NewPermissionRepository(db *gorm.DB) auth.PermissionRepository {
	return &permissionRepository{
		db: db,
	}
}

// Create creates a new permission
func (r *permissionRepository) Create(ctx context.Context, permission *auth.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

// GetByID retrieves a permission by ID
func (r *permissionRepository) GetByID(ctx context.Context, id ulid.ULID) (*auth.Permission, error) {
	var permission auth.Permission
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("permission not found")
		}
		return nil, err
	}
	return &permission, nil
}

// GetByName retrieves a permission by name
func (r *permissionRepository) GetByName(ctx context.Context, name string) (*auth.Permission, error) {
	var permission auth.Permission
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("permission not found")
		}
		return nil, err
	}
	return &permission, nil
}

// GetAll retrieves all permissions
func (r *permissionRepository) GetAll(ctx context.Context) ([]*auth.Permission, error) {
	var permissions []*auth.Permission
	err := r.db.WithContext(ctx).Order("category ASC, name ASC").Find(&permissions).Error
	return permissions, err
}

// GetAllPermissions retrieves all permissions (interface method)
func (r *permissionRepository) GetAllPermissions(ctx context.Context) ([]*auth.Permission, error) {
	return r.GetAll(ctx)
}

// GetByCategory retrieves permissions by category
func (r *permissionRepository) GetByCategory(ctx context.Context, category string) ([]*auth.Permission, error) {
	var permissions []*auth.Permission
	err := r.db.WithContext(ctx).
		Where("category = ?", category).
		Order("name ASC").
		Find(&permissions).Error
	return permissions, err
}

// GetByNames retrieves permissions by names
func (r *permissionRepository) GetByNames(ctx context.Context, names []string) ([]*auth.Permission, error) {
	var permissions []*auth.Permission
	err := r.db.WithContext(ctx).
		Where("name IN ?", names).
		Order("name ASC").
		Find(&permissions).Error
	return permissions, err
}

// Update updates a permission
func (r *permissionRepository) Update(ctx context.Context, permission *auth.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

// Delete soft deletes a permission
func (r *permissionRepository) Delete(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).Model(&auth.Permission{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

// GetPermissionsByRoleID retrieves permissions for a specific role
func (r *permissionRepository) GetPermissionsByRoleID(ctx context.Context, roleID ulid.ULID) ([]*auth.Permission, error) {
	var permissions []*auth.Permission
	err := r.db.WithContext(ctx).
		Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Order("permissions.category ASC, permissions.name ASC").
		Find(&permissions).Error
	return permissions, err
}

// GetUserPermissions retrieves permissions for a user in an organization
func (r *permissionRepository) GetUserPermissions(ctx context.Context, userID, orgID ulid.ULID) ([]string, error) {
	var permissionNames []string
	err := r.db.WithContext(ctx).
		Table("permissions").
		Select("permissions.name").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN roles ON role_permissions.role_id = roles.id").
		Joins("JOIN organization_members ON roles.id = organization_members.role_id").
		Where("organization_members.user_id = ? AND organization_members.organization_id = ?", userID, orgID).
		Pluck("permissions.name", &permissionNames).Error
	return permissionNames, err
}

// GetUserPermissionsByAPIKey retrieves permissions for a user through API key
func (r *permissionRepository) GetUserPermissionsByAPIKey(ctx context.Context, apiKeyID ulid.ULID) ([]string, error) {
	var permissionNames []string
	err := r.db.WithContext(ctx).
		Table("permissions").
		Select("permissions.name").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN roles ON role_permissions.role_id = roles.id").
		Joins("JOIN organization_members ON roles.id = organization_members.role_id").
		Joins("JOIN api_keys ON organization_members.user_id = api_keys.user_id AND organization_members.organization_id = api_keys.organization_id").
		Where("api_keys.id = ?", apiKeyID).
		Pluck("permissions.name", &permissionNames).Error
	return permissionNames, err
}