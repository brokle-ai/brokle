package auth

import (
	"context"

	"gorm.io/gorm"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// roleRepository implements auth.RoleRepository using GORM for normalized template roles
type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new template role repository instance
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
		Where("id = ?", id).
		Preload("Permissions").
		First(&role).Error
	
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByNameAndScope(ctx context.Context, name, scopeType string) (*auth.Role, error) {
	var role auth.Role
	err := r.db.WithContext(ctx).
		Where("name = ? AND scope_type = ? AND scope_id IS NULL", name, scopeType).
		Preload("Permissions").
		First(&role).Error
	
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) Update(ctx context.Context, role *auth.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepository) Delete(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).Delete(&auth.Role{}, id).Error
}

// Template role queries

func (r *roleRepository) GetByScopeType(ctx context.Context, scopeType string) ([]*auth.Role, error) {
	var roles []*auth.Role
	err := r.db.WithContext(ctx).
		Where("scope_type = ?", scopeType).
		Preload("Permissions").
		Find(&roles).Error
	
	return roles, err
}

func (r *roleRepository) GetAllRoles(ctx context.Context) ([]*auth.Role, error) {
	var roles []*auth.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		Find(&roles).Error
	
	return roles, err
}

// Permission management for roles

func (r *roleRepository) GetRolePermissions(ctx context.Context, roleID ulid.ULID) ([]*auth.Permission, error) {
	var permissions []*auth.Permission
	err := r.db.WithContext(ctx).
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	
	return permissions, err
}

func (r *roleRepository) AssignRolePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID, grantedBy *ulid.ULID) error {
	var rolePermissions []auth.RolePermission
	for _, permissionID := range permissionIDs {
		rolePermissions = append(rolePermissions, auth.RolePermission{
			RoleID:       roleID,
			PermissionID: permissionID,
			GrantedBy:    grantedBy,
		})
	}
	return r.db.WithContext(ctx).Create(&rolePermissions).Error
}

func (r *roleRepository) RevokeRolePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error {
	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id IN ?", roleID, permissionIDs).
		Delete(&auth.RolePermission{}).Error
}

func (r *roleRepository) UpdateRolePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID, grantedBy *ulid.ULID) error {
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
					GrantedBy:    grantedBy,
				})
			}
			return tx.Create(&rolePermissions).Error
		}
		
		return nil
	})
}

// Statistics

func (r *roleRepository) GetRoleStatistics(ctx context.Context) (*auth.RoleStatistics, error) {
	var stats auth.RoleStatistics
	
	// Get total role count
	var totalCount int64
	if err := r.db.WithContext(ctx).Model(&auth.Role{}).Count(&totalCount).Error; err != nil {
		return nil, err
	}
	stats.TotalRoles = int(totalCount)
	
	// Get scope distribution
	var scopeCounts []struct {
		ScopeType string
		Count     int64
	}
	if err := r.db.WithContext(ctx).
		Model(&auth.Role{}).
		Select("scope_type, COUNT(*) as count").
		Group("scope_type").
		Find(&scopeCounts).Error; err != nil {
		return nil, err
	}
	
	stats.ScopeDistribution = make(map[string]int)
	for _, sc := range scopeCounts {
		stats.ScopeDistribution[sc.ScopeType] = int(sc.Count)
		
		switch sc.ScopeType {
		case auth.ScopeOrganization:
			stats.OrganizationRoles = int(sc.Count)
		case auth.ScopeProject:
			stats.ProjectRoles = int(sc.Count)
		}
	}
	
	// Get role usage distribution (how many members have each role)
	var roleUsage []struct {
		RoleName string
		Count    int64
	}
	if err := r.db.WithContext(ctx).
		Model(&auth.OrganizationMember{}).
		Select("r.name as role_name, COUNT(*) as count").
		Joins("JOIN roles r ON organization_members.role_id = r.id").
		Group("r.name").
		Find(&roleUsage).Error; err != nil {
		return nil, err
	}
	
	stats.RoleDistribution = make(map[string]int)
	for _, ru := range roleUsage {
		stats.RoleDistribution[ru.RoleName] = int(ru.Count)
	}
	
	return &stats, nil
}

// System template role methods

func (r *roleRepository) GetSystemRoles(ctx context.Context) ([]*auth.Role, error) {
	var roles []*auth.Role
	err := r.db.WithContext(ctx).
		Where("scope_type = ? AND scope_id IS NULL", auth.ScopeSystem).
		Preload("Permissions").
		Order("name ASC").
		Find(&roles).Error
	return roles, err
}

// Custom scoped role methods

func (r *roleRepository) GetCustomRolesByScopeID(ctx context.Context, scopeType string, scopeID ulid.ULID) ([]*auth.Role, error) {
	var roles []*auth.Role
	err := r.db.WithContext(ctx).
		Where("scope_type = ? AND scope_id = ?", scopeType, scopeID).
		Preload("Permissions").
		Order("name ASC").
		Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetByNameScopeAndID(ctx context.Context, name, scopeType string, scopeID *ulid.ULID) (*auth.Role, error) {
	var role auth.Role
	query := r.db.WithContext(ctx).
		Where("name = ? AND scope_type = ?", name, scopeType)
	
	if scopeID == nil {
		query = query.Where("scope_id IS NULL")
	} else {
		query = query.Where("scope_id = ?", *scopeID)
	}
	
	err := query.Preload("Permissions").First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetCustomRolesByOrganization(ctx context.Context, organizationID ulid.ULID) ([]*auth.Role, error) {
	var roles []*auth.Role
	err := r.db.WithContext(ctx).
		Where("scope_type = ? AND scope_id = ?", auth.ScopeOrganization, organizationID).
		Preload("Permissions").
		Order("name ASC").
		Find(&roles).Error
	return roles, err
}

// Bulk operations

func (r *roleRepository) BulkCreate(ctx context.Context, roles []*auth.Role) error {
	return r.db.WithContext(ctx).Create(&roles).Error
}