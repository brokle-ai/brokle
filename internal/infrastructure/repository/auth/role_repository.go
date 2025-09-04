package auth

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// roleRepository implements auth.RoleRepository using GORM
type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new role repository instance
func NewRoleRepository(db *gorm.DB) auth.RoleRepository {
	return &roleRepository{
		db: db,
	}
}

// Create creates a new role
func (r *roleRepository) Create(ctx context.Context, role *auth.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// GetByID retrieves a role by ID
func (r *roleRepository) GetByID(ctx context.Context, id ulid.ULID) (*auth.Role, error) {
	var role auth.Role
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return &role, nil
}

// GetByName retrieves a role by name and organization
func (r *roleRepository) GetByName(ctx context.Context, orgID *ulid.ULID, name string) (*auth.Role, error) {
	var role auth.Role
	query := r.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", name)
	
	if orgID != nil {
		query = query.Where("organization_id = ?", *orgID)
	} else {
		query = query.Where("organization_id IS NULL")
	}
	
	err := query.First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return &role, nil
}

// Update updates a role
func (r *roleRepository) Update(ctx context.Context, role *auth.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// Delete soft deletes a role
func (r *roleRepository) Delete(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).Model(&auth.Role{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

// GetByOrganizationID retrieves roles for an organization
func (r *roleRepository) GetByOrganizationID(ctx context.Context, orgID ulid.ULID) ([]*auth.Role, error) {
	var roles []*auth.Role
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND deleted_at IS NULL", orgID).
		Order("name ASC").
		Find(&roles).Error
	return roles, err
}

// GetSystemRoles retrieves system-level roles
func (r *roleRepository) GetSystemRoles(ctx context.Context) ([]*auth.Role, error) {
	var roles []*auth.Role
	err := r.db.WithContext(ctx).
		Where("organization_id IS NULL AND deleted_at IS NULL").
		Order("name ASC").
		Find(&roles).Error
	return roles, err
}

// GetOrganizationRoles retrieves organization-specific roles
func (r *roleRepository) GetOrganizationRoles(ctx context.Context, orgID ulid.ULID) ([]*auth.Role, error) {
	var roles []*auth.Role
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND deleted_at IS NULL", orgID).
		Order("name ASC").
		Find(&roles).Error
	return roles, err
}

// AssignPermissions assigns permissions to a role
func (r *roleRepository) AssignPermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error {
	// First, remove existing permissions for this role
	err := r.db.WithContext(ctx).
		Where("role_id = ?", roleID).
		Delete(&auth.RolePermission{}).Error
	if err != nil {
		return err
	}

	// Add new permissions
	for _, permissionID := range permissionIDs {
		rolePermission := &auth.RolePermission{
			RoleID:       roleID,
			PermissionID: permissionID,
		}
		if err := r.db.WithContext(ctx).Create(rolePermission).Error; err != nil {
			return err
		}
	}
	return nil
}

// RevokePermissions removes specific permissions from a role
func (r *roleRepository) RevokePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error {
	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id IN ?", roleID, permissionIDs).
		Delete(&auth.RolePermission{}).Error
}

// RevokeAllPermissions removes all permissions from a role
func (r *roleRepository) RevokeAllPermissions(ctx context.Context, roleID ulid.ULID) error {
	return r.db.WithContext(ctx).
		Where("role_id = ?", roleID).
		Delete(&auth.RolePermission{}).Error
}

// GetRolePermissions retrieves all permissions for a role
func (r *roleRepository) GetRolePermissions(ctx context.Context, roleID ulid.ULID) ([]*auth.Permission, error) {
	var permissions []*auth.Permission
	err := r.db.WithContext(ctx).
		Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	return permissions, err
}

// IsSystemRole checks if a role is a system role
func (r *roleRepository) IsSystemRole(ctx context.Context, roleID ulid.ULID) (bool, error) {
	var role auth.Role
	err := r.db.WithContext(ctx).
		Select("organization_id").
		Where("id = ?", roleID).
		First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("role not found")
		}
		return false, err
	}
	return role.OrganizationID == nil, nil
}

// CanDeleteRole checks if a role can be deleted (not system role and not assigned to users)
func (r *roleRepository) CanDeleteRole(ctx context.Context, roleID ulid.ULID) (bool, error) {
	// Check if it's a system role
	isSystem, err := r.IsSystemRole(ctx, roleID)
	if err != nil {
		return false, err
	}
	if isSystem {
		return false, nil // Cannot delete system roles
	}

	// Check if role is assigned to any organization members
	var memberCount int64
	err = r.db.WithContext(ctx).
		Model(&auth.Role{}).
		Joins("JOIN organization_members ON roles.id = organization_members.role_id").
		Where("roles.id = ?", roleID).
		Count(&memberCount).Error
	if err != nil {
		return false, err
	}

	return memberCount == 0, nil
}

// GetAllRoles retrieves both system and organization roles
func (r *roleRepository) GetAllRoles(ctx context.Context, orgID ulid.ULID) ([]*auth.Role, error) {
	var roles []*auth.Role
	err := r.db.WithContext(ctx).
		Where("(organization_id IS NULL OR organization_id = ?) AND deleted_at IS NULL", orgID).
		Order("organization_id DESC, name ASC"). // System roles first
		Find(&roles).Error
	return roles, err
}