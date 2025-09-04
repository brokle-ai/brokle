package auth

import (
	"context"

	"gorm.io/gorm"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// rolePermissionRepository implements auth.RolePermissionRepository using GORM
type rolePermissionRepository struct {
	db *gorm.DB
}

// NewRolePermissionRepository creates a new role permission repository instance
func NewRolePermissionRepository(db *gorm.DB) auth.RolePermissionRepository {
	return &rolePermissionRepository{
		db: db,
	}
}

// Create creates a new role permission association
func (r *rolePermissionRepository) Create(ctx context.Context, rolePermission *auth.RolePermission) error {
	return r.db.WithContext(ctx).Create(rolePermission).Error
}

// GetByRoleID retrieves all role permissions for a role
func (r *rolePermissionRepository) GetByRoleID(ctx context.Context, roleID ulid.ULID) ([]*auth.RolePermission, error) {
	var rolePermissions []*auth.RolePermission
	err := r.db.WithContext(ctx).
		Where("role_id = ?", roleID).
		Find(&rolePermissions).Error
	return rolePermissions, err
}

// GetByPermissionID retrieves all role permissions for a permission
func (r *rolePermissionRepository) GetByPermissionID(ctx context.Context, permissionID ulid.ULID) ([]*auth.RolePermission, error) {
	var rolePermissions []*auth.RolePermission
	err := r.db.WithContext(ctx).
		Where("permission_id = ?", permissionID).
		Find(&rolePermissions).Error
	return rolePermissions, err
}

// Delete removes a role permission association
func (r *rolePermissionRepository) Delete(ctx context.Context, roleID, permissionID ulid.ULID) error {
	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&auth.RolePermission{}).Error
}

// DeleteByRoleID removes all permissions for a role
func (r *rolePermissionRepository) DeleteByRoleID(ctx context.Context, roleID ulid.ULID) error {
	return r.db.WithContext(ctx).
		Where("role_id = ?", roleID).
		Delete(&auth.RolePermission{}).Error
}

// DeleteByPermissionID removes all roles for a permission
func (r *rolePermissionRepository) DeleteByPermissionID(ctx context.Context, permissionID ulid.ULID) error {
	return r.db.WithContext(ctx).
		Where("permission_id = ?", permissionID).
		Delete(&auth.RolePermission{}).Error
}

// Exists checks if a role permission association exists
func (r *rolePermissionRepository) Exists(ctx context.Context, roleID, permissionID ulid.ULID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&auth.RolePermission{}).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Count(&count).Error
	return count > 0, err
}

// AssignPermissions assigns permissions to a role (bulk operation)
func (r *rolePermissionRepository) AssignPermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error {
	// First, remove existing permissions for this role
	err := r.DeleteByRoleID(ctx, roleID)
	if err != nil {
		return err
	}

	// Add new permissions
	for _, permissionID := range permissionIDs {
		rolePermission := &auth.RolePermission{
			RoleID:       roleID,
			PermissionID: permissionID,
		}
		if err := r.Create(ctx, rolePermission); err != nil {
			return err
		}
	}
	return nil
}

// RevokePermissions removes specific permissions from a role
func (r *rolePermissionRepository) RevokePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error {
	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id IN ?", roleID, permissionIDs).
		Delete(&auth.RolePermission{}).Error
}

// RevokeAllPermissions removes all permissions from a role
func (r *rolePermissionRepository) RevokeAllPermissions(ctx context.Context, roleID ulid.ULID) error {
	return r.DeleteByRoleID(ctx, roleID)
}

// HasPermission checks if a role has a specific permission
func (r *rolePermissionRepository) HasPermission(ctx context.Context, roleID, permissionID ulid.ULID) (bool, error) {
	return r.Exists(ctx, roleID, permissionID)
}