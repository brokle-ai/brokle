package auth

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// roleRepository implements clean auth.RoleRepository using GORM with scope_type + scope_id design
type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new clean role repository instance
func NewRoleRepository(db *gorm.DB) auth.RoleRepository {
	return &roleRepository{
		db: db,
	}
}

// Core CRUD operations

func (r *roleRepository) Create(ctx context.Context, role *auth.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) GetByID(ctx context.Context, id ulid.ULID) (*auth.Role, error) {
	var role auth.Role
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) Update(ctx context.Context, role *auth.Role) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", role.ID).
		Updates(role).Error
}

func (r *roleRepository) Delete(ctx context.Context, id ulid.ULID) error {
	// Soft delete
	return r.db.WithContext(ctx).
		Model(&auth.Role{}).
		Where("id = ?", id).
		Update("deleted_at", "NOW()").Error
}

// Clean scoped queries

func (r *roleRepository) GetByScope(ctx context.Context, scopeType string, scopeID *ulid.ULID) ([]*auth.Role, error) {
	query := r.db.WithContext(ctx).Where("scope_type = ? AND deleted_at IS NULL", scopeType)
	
	if scopeType == auth.ScopeSystem {
		query = query.Where("scope_id IS NULL")
	} else {
		query = query.Where("scope_id = ?", scopeID)
	}
	
	var roles []*auth.Role
	err := query.Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetByScopedName(ctx context.Context, scopeType string, scopeID *ulid.ULID, name string) (*auth.Role, error) {
	query := r.db.WithContext(ctx).
		Where("scope_type = ? AND name = ? AND deleted_at IS NULL", scopeType, name)
	
	if scopeType == auth.ScopeSystem {
		query = query.Where("scope_id IS NULL")
	} else {
		query = query.Where("scope_id = ?", scopeID)
	}
	
	var role auth.Role
	err := query.First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetSystemRoles(ctx context.Context) ([]*auth.Role, error) {
	return r.GetByScope(ctx, auth.ScopeSystem, nil)
}

func (r *roleRepository) GetOrganizationRoles(ctx context.Context, orgID ulid.ULID) ([]*auth.Role, error) {
	return r.GetByScope(ctx, auth.ScopeOrganization, &orgID)
}

func (r *roleRepository) GetProjectRoles(ctx context.Context, projectID ulid.ULID) ([]*auth.Role, error) {
	return r.GetByScope(ctx, auth.ScopeProject, &projectID)
}

// User role management (clean)

func (r *roleRepository) AssignUserRole(ctx context.Context, userID, roleID ulid.ULID) error {
	userRole := &auth.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return r.db.WithContext(ctx).Create(userRole).Error
}

func (r *roleRepository) RevokeUserRole(ctx context.Context, userID, roleID ulid.ULID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&auth.UserRole{}).Error
}

func (r *roleRepository) GetUserRoles(ctx context.Context, userID ulid.ULID) ([]*auth.Role, error) {
	var roles []*auth.Role
	err := r.db.WithContext(ctx).
		Table("roles").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.deleted_at IS NULL", userID).
		Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetUserRolesByScope(ctx context.Context, userID ulid.ULID, scopeType string) ([]*auth.Role, error) {
	var roles []*auth.Role
	err := r.db.WithContext(ctx).
		Table("roles").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.scope_type = ? AND roles.deleted_at IS NULL", userID, scopeType).
		Find(&roles).Error
	return roles, err
}

// Clean permission queries (effective permissions across all scopes)

func (r *roleRepository) GetUserEffectivePermissions(ctx context.Context, userID ulid.ULID) ([]string, error) {
	var permissions []string
	err := r.db.WithContext(ctx).
		Table("permissions").
		Select("DISTINCT permissions.name").
		Joins("JOIN role_permissions rp ON permissions.id = rp.permission_id").
		Joins("JOIN roles r ON rp.role_id = r.id").
		Joins("JOIN user_roles ur ON r.id = ur.role_id").
		Where("ur.user_id = ? AND permissions.deleted_at IS NULL AND r.deleted_at IS NULL", userID).
		Pluck("name", &permissions).Error
	return permissions, err
}

func (r *roleRepository) HasUserPermission(ctx context.Context, userID ulid.ULID, permission string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("permissions").
		Joins("JOIN role_permissions rp ON permissions.id = rp.permission_id").
		Joins("JOIN roles r ON rp.role_id = r.id").
		Joins("JOIN user_roles ur ON r.id = ur.role_id").
		Where("ur.user_id = ? AND permissions.name = ? AND permissions.deleted_at IS NULL AND r.deleted_at IS NULL", userID, permission).
		Count(&count).Error
	return count > 0, err
}

func (r *roleRepository) CheckUserPermissions(ctx context.Context, userID ulid.ULID, permissions []string) (map[string]bool, error) {
	result := make(map[string]bool)
	
	// Initialize all to false
	for _, permission := range permissions {
		result[permission] = false
	}
	
	var userPermissions []string
	err := r.db.WithContext(ctx).
		Table("permissions").
		Select("DISTINCT permissions.name").
		Joins("JOIN role_permissions rp ON permissions.id = rp.permission_id").
		Joins("JOIN roles r ON rp.role_id = r.id").
		Joins("JOIN user_roles ur ON r.id = ur.role_id").
		Where("ur.user_id = ? AND permissions.name IN ? AND permissions.deleted_at IS NULL AND r.deleted_at IS NULL", userID, permissions).
		Pluck("name", &userPermissions).Error
	
	if err != nil {
		return result, err
	}
	
	// Set found permissions to true
	for _, permission := range userPermissions {
		result[permission] = true
	}
	
	return result, nil
}

// Permission management

func (r *roleRepository) GetRolePermissions(ctx context.Context, roleID ulid.ULID) ([]*auth.Permission, error) {
	var permissions []*auth.Permission
	err := r.db.WithContext(ctx).
		Table("permissions").
		Joins("JOIN role_permissions rp ON permissions.id = rp.permission_id").
		Where("rp.role_id = ? AND permissions.deleted_at IS NULL", roleID).
		Find(&permissions).Error
	return permissions, err
}

func (r *roleRepository) AssignRolePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error {
	var rolePermissions []auth.RolePermission
	for _, permissionID := range permissionIDs {
		rolePermissions = append(rolePermissions, auth.RolePermission{
			RoleID:       roleID,
			PermissionID: permissionID,
		})
	}
	return r.db.WithContext(ctx).Create(&rolePermissions).Error
}

func (r *roleRepository) RevokeRolePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error {
	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id IN ?", roleID, permissionIDs).
		Delete(&auth.RolePermission{}).Error
}

func (r *roleRepository) UpdateRolePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Remove all existing permissions
		if err := tx.Where("role_id = ?", roleID).Delete(&auth.RolePermission{}).Error; err != nil {
			return err
		}
		
		// Add new permissions
		if len(permissionIDs) > 0 {
			var rolePermissions []auth.RolePermission
			for _, permissionID := range permissionIDs {
				rolePermissions = append(rolePermissions, auth.RolePermission{
					RoleID:       roleID,
					PermissionID: permissionID,
				})
			}
			return tx.Create(&rolePermissions).Error
		}
		
		return nil
	})
}

// Statistics and validation

func (r *roleRepository) GetRoleStatistics(ctx context.Context) (*auth.RoleStatistics, error) {
	stats := &auth.RoleStatistics{
		ScopeDistribution: make(map[string]int),
		RoleDistribution:  make(map[string]int),
	}
	
	// Count roles by scope type
	var scopeCounts []struct {
		ScopeType string
		Count     int
	}
	err := r.db.WithContext(ctx).
		Model(&auth.Role{}).
		Select("scope_type, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Group("scope_type").
		Scan(&scopeCounts).Error
	if err != nil {
		return nil, err
	}
	
	for _, sc := range scopeCounts {
		stats.ScopeDistribution[sc.ScopeType] = sc.Count
		stats.TotalRoles += sc.Count
		
		switch sc.ScopeType {
		case auth.ScopeSystem:
			stats.SystemRoles = sc.Count
		case auth.ScopeOrganization:
			stats.OrganizationRoles = sc.Count
		case auth.ScopeProject:
			stats.ProjectRoles = sc.Count
		}
	}
	
	// Count permission count
	var permissionCount int64
	err = r.db.WithContext(ctx).
		Model(&auth.Permission{}).
		Where("deleted_at IS NULL").
		Count(&permissionCount).Error
	if err != nil {
		return nil, err
	}
	stats.PermissionCount = int(permissionCount)
	
	// Count role assignments
	var roleCounts []struct {
		RoleName string
		Count    int
	}
	err = r.db.WithContext(ctx).
		Table("roles").
		Select("roles.name as role_name, COUNT(user_roles.user_id) as count").
		Joins("LEFT JOIN user_roles ON roles.id = user_roles.role_id").
		Where("roles.deleted_at IS NULL").
		Group("roles.name").
		Scan(&roleCounts).Error
	if err != nil {
		return nil, err
	}
	
	for _, rc := range roleCounts {
		stats.RoleDistribution[rc.RoleName] = rc.Count
	}
	
	return stats, nil
}

func (r *roleRepository) CanDeleteRole(ctx context.Context, roleID ulid.ULID) (bool, error) {
	var role auth.Role
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", roleID).
		First(&role).Error
	if err != nil {
		return false, err
	}
	
	// System roles cannot be deleted
	return !role.IsSystemRole(), nil
}

// Bulk operations

func (r *roleRepository) BulkAssignPermissions(ctx context.Context, assignments []auth.RolePermissionAssignment) error {
	var rolePermissions []auth.RolePermission
	for _, assignment := range assignments {
		rolePermissions = append(rolePermissions, auth.RolePermission{
			RoleID:       assignment.RoleID,
			PermissionID: assignment.PermissionID,
		})
	}
	return r.db.WithContext(ctx).Create(&rolePermissions).Error
}

func (r *roleRepository) BulkRevokePermissions(ctx context.Context, revocations []auth.RolePermissionRevocation) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, revocation := range revocations {
			if err := tx.Where("role_id = ? AND permission_id = ?", revocation.RoleID, revocation.PermissionID).
				Delete(&auth.RolePermission{}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}