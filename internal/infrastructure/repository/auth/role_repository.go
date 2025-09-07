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

// GetGlobalSystemRole retrieves a specific global system role by name
func (r *roleRepository) GetGlobalSystemRole(ctx context.Context, name string) (*auth.Role, error) {
	var role auth.Role
	err := r.db.WithContext(ctx).
		Where("name = ? AND organization_id IS NULL AND deleted_at IS NULL", name).
		First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return &role, nil
}

// GetAvailableRoles retrieves system + org roles for assignment
func (r *roleRepository) GetAvailableRoles(ctx context.Context, orgID ulid.ULID) ([]*auth.Role, error) {
	var roles []*auth.Role
	err := r.db.WithContext(ctx).
		Where("(organization_id IS NULL OR organization_id = ?) AND deleted_at IS NULL", orgID).
		Order("organization_id DESC, name ASC").
		Find(&roles).Error
	return roles, err
}

// ListRoles returns paginated list with total count
func (r *roleRepository) ListRoles(ctx context.Context, orgID *ulid.ULID, limit, offset int) ([]*auth.Role, int, error) {
	query := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	
	// Apply organization filter
	if orgID != nil {
		query = query.Where("(organization_id IS NULL OR organization_id = ?)", *orgID)
	}
	
	// Get total count
	var total int64
	if err := query.Model(&auth.Role{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get roles with pagination
	var roles []*auth.Role
	err := query.
		Order("organization_id DESC, name ASC").
		Limit(limit).
		Offset(offset).
		Find(&roles).Error
	
	return roles, int(total), err
}

// SearchRoles searches roles with pagination
func (r *roleRepository) SearchRoles(ctx context.Context, orgID *ulid.ULID, query string, limit, offset int) ([]*auth.Role, int, error) {
	dbQuery := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	
	// Apply organization filter
	if orgID != nil {
		dbQuery = dbQuery.Where("(organization_id IS NULL OR organization_id = ?)", *orgID)
	}
	
	// Apply search filter
	searchPattern := "%" + query + "%"
	dbQuery = dbQuery.Where("(name ILIKE ? OR description ILIKE ?)", searchPattern, searchPattern)
	
	// Get total count
	var total int64
	if err := dbQuery.Model(&auth.Role{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get roles with pagination
	var roles []*auth.Role
	err := dbQuery.
		Order("organization_id DESC, name ASC").
		Limit(limit).
		Offset(offset).
		Find(&roles).Error
	
	return roles, int(total), err
}

// GetRolesByPermission finds roles that have a specific permission
func (r *roleRepository) GetRolesByPermission(ctx context.Context, permissionID ulid.ULID) ([]*auth.Role, error) {
	var roles []*auth.Role
	err := r.db.WithContext(ctx).
		Table("roles").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Where("role_permissions.permission_id = ? AND roles.deleted_at IS NULL", permissionID).
		Find(&roles).Error
	return roles, err
}

// GetUserRole retrieves user's role in an organization
func (r *roleRepository) GetUserRole(ctx context.Context, userID, orgID ulid.ULID) (*auth.Role, error) {
	var role auth.Role
	err := r.db.WithContext(ctx).
		Table("roles").
		Joins("JOIN organization_members ON roles.id = organization_members.role_id").
		Where("organization_members.user_id = ? AND organization_members.organization_id = ? AND roles.deleted_at IS NULL", userID, orgID).
		First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user role not found")
		}
		return nil, err
	}
	return &role, nil
}

// AssignUserRole assigns a role to a user in an organization
func (r *roleRepository) AssignUserRole(ctx context.Context, userID, orgID, roleID ulid.ULID) error {
	// Update the organization_members table with the new role
	return r.db.WithContext(ctx).
		Table("organization_members").
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		Update("role_id", roleID).Error
}

// RevokeUserRole removes a user's role in an organization
func (r *roleRepository) RevokeUserRole(ctx context.Context, userID, orgID ulid.ULID) error {
	// Set role_id to NULL in organization_members
	return r.db.WithContext(ctx).
		Table("organization_members").
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		Update("role_id", nil).Error
}

// UpdateRolePermissions replaces all permissions for a role
func (r *roleRepository) UpdateRolePermissions(ctx context.Context, roleID ulid.ULID, permissionIDs []ulid.ULID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Remove all existing permissions
		if err := tx.Where("role_id = ?", roleID).Delete(&auth.RolePermission{}).Error; err != nil {
			return err
		}
		
		// Add new permissions
		for _, permissionID := range permissionIDs {
			rolePermission := &auth.RolePermission{
				RoleID:       roleID,
				PermissionID: permissionID,
			}
			if err := tx.Create(rolePermission).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// HasPermission checks if a user has a specific permission through their role
func (r *roleRepository) HasPermission(ctx context.Context, userID, orgID ulid.ULID, resourceAction string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("organization_members").
		Joins("JOIN roles ON organization_members.role_id = roles.id").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("organization_members.user_id = ? AND organization_members.organization_id = ? AND permissions.resource_action = ?", userID, orgID, resourceAction).
		Count(&count).Error
	return count > 0, err
}

// HasResourceAction checks if a user has specific resource:action permission
func (r *roleRepository) HasResourceAction(ctx context.Context, userID, orgID ulid.ULID, resource, action string) (bool, error) {
	resourceAction := resource + ":" + action
	return r.HasPermission(ctx, userID, orgID, resourceAction)
}

// CheckPermissions checks multiple permissions at once
func (r *roleRepository) CheckPermissions(ctx context.Context, userID, orgID ulid.ULID, resourceActions []string) (map[string]bool, error) {
	result := make(map[string]bool)
	
	if len(resourceActions) == 0 {
		return result, nil
	}
	
	// Query for all permissions the user has through their role
	var permissions []string
	err := r.db.WithContext(ctx).
		Table("organization_members").
		Joins("JOIN roles ON organization_members.role_id = roles.id").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("organization_members.user_id = ? AND organization_members.organization_id = ? AND permissions.resource_action IN ?", userID, orgID, resourceActions).
		Pluck("permissions.resource_action", &permissions).Error
	if err != nil {
		return nil, err
	}
	
	// Build result map
	permissionSet := make(map[string]bool)
	for _, perm := range permissions {
		permissionSet[perm] = true
	}
	
	for _, resourceAction := range resourceActions {
		result[resourceAction] = permissionSet[resourceAction]
	}
	
	return result, nil
}

// ValidateRole validates if a role exists and can be used
func (r *roleRepository) ValidateRole(ctx context.Context, roleID ulid.ULID) error {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&auth.Role{}).
		Where("id = ? AND deleted_at IS NULL", roleID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("role not found or deleted")
	}
	return nil
}

// GetRoleStatistics returns statistics about roles in an organization
func (r *roleRepository) GetRoleStatistics(ctx context.Context, orgID ulid.ULID) (*auth.RoleStatistics, error) {
	stats := &auth.RoleStatistics{
		OrganizationID: orgID,
		LastUpdated:    time.Now(),
	}
	
	// Total roles count (system + org)
	var totalRoles int64
	err := r.db.WithContext(ctx).
		Model(&auth.Role{}).
		Where("(organization_id IS NULL OR organization_id = ?) AND deleted_at IS NULL", orgID).
		Count(&totalRoles).Error
	if err != nil {
		return nil, err
	}
	stats.TotalRoles = int(totalRoles)
	
	// System roles count
	var systemRoles int64
	err = r.db.WithContext(ctx).
		Model(&auth.Role{}).
		Where("organization_id IS NULL AND deleted_at IS NULL").
		Count(&systemRoles).Error
	if err != nil {
		return nil, err
	}
	stats.SystemRoles = int(systemRoles)
	
	// Custom roles count (org-specific)
	var customRoles int64
	err = r.db.WithContext(ctx).
		Model(&auth.Role{}).
		Where("organization_id = ? AND deleted_at IS NULL", orgID).
		Count(&customRoles).Error
	if err != nil {
		return nil, err
	}
	stats.CustomRoles = int(customRoles)
	
	// Total members in organization
	var totalMembers int64
	err = r.db.WithContext(ctx).
		Table("organization_members").
		Where("organization_id = ?", orgID).
		Count(&totalMembers).Error
	if err != nil {
		return nil, err
	}
	stats.TotalMembers = int(totalMembers)
	
	// Role distribution
	type roleCount struct {
		RoleName string
		Count    int64
	}
	var roleCounts []roleCount
	err = r.db.WithContext(ctx).
		Table("organization_members").
		Select("roles.name as role_name, COUNT(organization_members.user_id) as count").
		Joins("LEFT JOIN roles ON organization_members.role_id = roles.id").
		Where("organization_members.organization_id = ?", orgID).
		Group("roles.name").
		Find(&roleCounts).Error
	if err != nil {
		return nil, err
	}
	
	stats.RoleDistribution = make(map[string]int)
	for _, rc := range roleCounts {
		roleName := rc.RoleName
		if roleName == "" {
			roleName = "No Role"
		}
		stats.RoleDistribution[roleName] = int(rc.Count)
	}
	
	// Permission count (total permissions across all roles)
	var permissionCount int64
	err = r.db.WithContext(ctx).
		Model(&auth.Permission{}).
		Count(&permissionCount).Error
	if err != nil {
		return nil, err
	}
	stats.PermissionCount = int(permissionCount)
	
	return stats, nil
}

// CreateDefaultRoles creates default roles for an organization
func (r *roleRepository) CreateDefaultRoles(ctx context.Context, orgID ulid.ULID) error {
	defaultRoles := []*auth.Role{
		{
			ID:             ulid.New(),
			Name:           "Organization Owner",
			Description:    "Full access to organization",
			OrganizationID: &orgID,
			IsSystemRole:   false,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             ulid.New(),
			Name:           "Organization Admin",
			Description:    "Administrative access to organization",
			OrganizationID: &orgID,
			IsSystemRole:   false,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             ulid.New(),
			Name:           "Developer",
			Description:    "Development access",
			OrganizationID: &orgID,
			IsSystemRole:   false,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             ulid.New(),
			Name:           "Viewer",
			Description:    "Read-only access",
			OrganizationID: &orgID,
			IsSystemRole:   false,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, role := range defaultRoles {
			if err := tx.Create(role).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BulkAssignPermissions assigns permissions to multiple roles in bulk
func (r *roleRepository) BulkAssignPermissions(ctx context.Context, assignments []auth.RolePermissionAssignment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, assignment := range assignments {
			rolePermission := &auth.RolePermission{
				RoleID:       assignment.RoleID,
				PermissionID: assignment.PermissionID,
			}
			if err := tx.Create(rolePermission).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BulkRevokePermissions revokes permissions from multiple roles in bulk
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