package seeder

import (
	"context"
	"fmt"
	"log"
	"strings"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// RBACSeeder handles clean seeding of RBAC (roles, permissions, user roles) data
type RBACSeeder struct {
	roleRepo       auth.RoleRepository
	userRoleRepo   auth.UserRoleRepository
	permissionRepo auth.PermissionRepository
	rolePermRepo   auth.RolePermissionRepository
}

// NewRBACSeeder creates a new clean RBACSeeder instance
func NewRBACSeeder(
	roleRepo auth.RoleRepository,
	userRoleRepo auth.UserRoleRepository,
	permissionRepo auth.PermissionRepository,
	rolePermRepo auth.RolePermissionRepository,
) *RBACSeeder {
	return &RBACSeeder{
		roleRepo:       roleRepo,
		userRoleRepo:   userRoleRepo,
		permissionRepo: permissionRepo,
		rolePermRepo:   rolePermRepo,
	}
}

// SeedPermissions seeds permissions from the provided seed data
func (rs *RBACSeeder) SeedPermissions(ctx context.Context, permissionSeeds []PermissionSeed, entityMaps *EntityMaps, verbose bool) error {
	if verbose {
		log.Printf("🔑 Seeding %d permissions...", len(permissionSeeds))
	}

	for _, permSeed := range permissionSeeds {
		// Check if permission already exists
		existing, err := rs.permissionRepo.GetByName(ctx, permSeed.Name)
		if err == nil && existing != nil {
			if verbose {
				log.Printf("   Permission %s already exists, skipping", permSeed.Name)
			}
			entityMaps.Permissions[permSeed.Name] = existing.ID
			continue
		}

		// Parse resource and action from name (format: "resource:action")
		parts := strings.SplitN(permSeed.Name, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid permission name format: %s (expected resource:action)", permSeed.Name)
		}

		resource := parts[0]
		action := parts[1]

		// Create new permission
		permission := auth.NewPermission(
			resource,
			action,
			permSeed.DisplayName,
			permSeed.Description,
			permSeed.Category,
		)

		if err := rs.permissionRepo.Create(ctx, permission); err != nil {
			return fmt.Errorf("failed to create permission %s: %w", permSeed.Name, err)
		}

		entityMaps.Permissions[permSeed.Name] = permission.ID

		if verbose {
			log.Printf("   ✓ Created permission: %s", permSeed.Name)
		}
	}

	return nil
}

// SeedRoles seeds clean scoped roles from the provided seed data
func (rs *RBACSeeder) SeedRoles(ctx context.Context, roleSeeds []RoleSeed, entityMaps *EntityMaps, verbose bool) error {
	if verbose {
		log.Printf("👤 Seeding %d roles...", len(roleSeeds))
	}

	for _, roleSeed := range roleSeeds {
		// Resolve scope ID based on scope type
		// IMPORTANT: For organization and project roles, scope_id should be NULL
		// These are role TEMPLATES available to all entities of that type
		var scopeID *ulid.ULID = nil
		
		// Only system roles have scope_id = NULL explicitly
		// Organization and project roles also have scope_id = NULL (they are templates)
		// Future: Custom entity-specific roles would have scope_id set to specific entity

		// Check if role already exists in this scope
		existing, err := rs.roleRepo.GetByScopedName(ctx, roleSeed.ScopeType, scopeID, roleSeed.Name)
		if err == nil && existing != nil {
			if verbose {
				log.Printf("   Role %s (%s) already exists, skipping", roleSeed.Name, roleSeed.ScopeType)
			}
			entityMaps.Roles[rs.getRoleKey(roleSeed)] = existing.ID
			continue
		}

		// Create new role
		role := auth.NewRole(
			roleSeed.ScopeType,
			scopeID,
			roleSeed.Name,
			roleSeed.DisplayName,
			roleSeed.Description,
		)

		if err := rs.roleRepo.Create(ctx, role); err != nil {
			return fmt.Errorf("failed to create role %s: %w", roleSeed.Name, err)
		}

		entityMaps.Roles[rs.getRoleKey(roleSeed)] = role.ID

		// Assign permissions to role
		if len(roleSeed.Permissions) > 0 {
			var permissionIDs []ulid.ULID
			for _, permName := range roleSeed.Permissions {
				permID, exists := entityMaps.Permissions[permName]
				if !exists {
					return fmt.Errorf("permission not found: %s for role %s", permName, roleSeed.Name)
				}
				permissionIDs = append(permissionIDs, permID)
			}

			if err := rs.roleRepo.AssignRolePermissions(ctx, role.ID, permissionIDs); err != nil {
				return fmt.Errorf("failed to assign permissions to role %s: %w", roleSeed.Name, err)
			}
		}

		if verbose {
			log.Printf("   ✓ Created role: %s (%s scope)", roleSeed.Name, roleSeed.ScopeType)
		}
	}

	return nil
}

// SeedMemberships seeds clean user role assignments from the provided seed data
func (rs *RBACSeeder) SeedMemberships(ctx context.Context, membershipSeeds []MembershipSeed, entityMaps *EntityMaps, verbose bool) error {
	if verbose {
		log.Printf("🤝 Seeding %d memberships...", len(membershipSeeds))
	}

	for _, membershipSeed := range membershipSeeds {
		// Find user
		userID, exists := entityMaps.Users[membershipSeed.UserEmail]
		if !exists {
			return fmt.Errorf("user not found: %s", membershipSeed.UserEmail)
		}

		// Find role by constructing role key
		roleKey := rs.getMembershipRoleKey(membershipSeed)
		roleID, exists := entityMaps.Roles[roleKey]
		if !exists {
			return fmt.Errorf("role not found for key: %s", roleKey)
		}

		// Check if assignment already exists
		exists, err := rs.userRoleRepo.Exists(ctx, userID, roleID)
		if err != nil {
			return fmt.Errorf("failed to check user role assignment: %w", err)
		}
		if exists {
			if verbose {
				log.Printf("   User role assignment already exists for %s -> %s, skipping", membershipSeed.UserEmail, membershipSeed.RoleName)
			}
			continue
		}

		// Create user role assignment
		userRole := auth.NewUserRole(userID, roleID)
		if err := rs.userRoleRepo.Create(ctx, userRole); err != nil {
			return fmt.Errorf("failed to assign role %s to user %s: %w", membershipSeed.RoleName, membershipSeed.UserEmail, err)
		}

		if verbose {
			log.Printf("   ✓ Assigned role %s (%s) to user %s", membershipSeed.RoleName, membershipSeed.ScopeType, membershipSeed.UserEmail)
		}
	}

	return nil
}

// Helper methods

func (rs *RBACSeeder) getRoleKey(roleSeed RoleSeed) string {
	// Organization and project roles are templates, so key is just scope_type:name
	switch roleSeed.ScopeType {
	case auth.ScopeSystem:
		return fmt.Sprintf("system:%s", roleSeed.Name)
	case auth.ScopeOrganization:
		return fmt.Sprintf("organization:%s", roleSeed.Name)  // No org slug - it's a template
	case auth.ScopeProject:
		return fmt.Sprintf("project:%s", roleSeed.Name)       // No org/project - it's a template
	default:
		return fmt.Sprintf("%s:%s", roleSeed.ScopeType, roleSeed.Name)
	}
}

func (rs *RBACSeeder) getMembershipRoleKey(membershipSeed MembershipSeed) string {
	// When assigning roles, we reference the template roles
	switch membershipSeed.ScopeType {
	case auth.ScopeSystem:
		return fmt.Sprintf("system:%s", membershipSeed.RoleName)
	case auth.ScopeOrganization:
		return fmt.Sprintf("organization:%s", membershipSeed.RoleName)  // Reference template role
	case auth.ScopeProject:
		return fmt.Sprintf("project:%s", membershipSeed.RoleName)       // Reference template role
	default:
		return fmt.Sprintf("%s:%s", membershipSeed.ScopeType, membershipSeed.RoleName)
	}
}