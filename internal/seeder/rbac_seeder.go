package seeder

import (
	"context"
	"fmt"
	"log"
	"strings"

	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/organization"
	"brokle/pkg/ulid"
)

// RBACSeeder handles seeding of RBAC (roles, permissions, memberships) data
type RBACSeeder struct {
	roleRepo       auth.RoleRepository
	permissionRepo auth.PermissionRepository
	rolePermRepo   auth.RolePermissionRepository
	memberRepo     organization.MemberRepository
}

// NewRBACSeeder creates a new RBACSeeder instance
func NewRBACSeeder(roleRepo auth.RoleRepository, permissionRepo auth.PermissionRepository, rolePermRepo auth.RolePermissionRepository, memberRepo organization.MemberRepository) *RBACSeeder {
	return &RBACSeeder{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		rolePermRepo:   rolePermRepo,
		memberRepo:     memberRepo,
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
		resource := ""
		action := ""
		if len(parts) == 2 {
			resource = parts[0]
			action = parts[1]
		}

		// Create permission entity
		permission := &auth.Permission{
			ID:          ulid.New(),
			Name:        permSeed.Name,
			Resource:    resource,
			Action:      action,
			DisplayName: permSeed.DisplayName,
			Description: permSeed.Description,
			Category:    permSeed.Category,
		}

		// Set defaults
		if permission.Category == "" {
			permission.Category = "general"
		}

		// Create permission in database
		if err := rs.permissionRepo.Create(ctx, permission); err != nil {
			return fmt.Errorf("failed to create permission %s: %w", permSeed.Name, err)
		}

		// Store permission ID for later reference
		entityMaps.Permissions[permSeed.Name] = permission.ID

		if verbose {
			log.Printf("   ✅ Created permission: %s (%s)", permission.DisplayName, permission.Name)
		}
	}

	if verbose {
		log.Printf("✅ Permissions seeded successfully")
	}
	return nil
}

// SeedRoles seeds roles and their permission associations from the provided seed data
func (rs *RBACSeeder) SeedRoles(ctx context.Context, roleSeeds []RoleSeed, entityMaps *EntityMaps, verbose bool) error {
	if verbose {
		log.Printf("👑 Seeding %d roles...", len(roleSeeds))
	}

	for _, roleSeed := range roleSeeds {
		// Generate role key for mapping
		var roleKey string
		if roleSeed.IsSystemRole || roleSeed.OrganizationSlug == "" {
			roleKey = fmt.Sprintf("system:%s", roleSeed.Name)
		} else {
			roleKey = fmt.Sprintf("%s:%s", roleSeed.OrganizationSlug, roleSeed.Name)
		}

		// Check if role already exists
		var existing *auth.Role
		var err error
		if roleSeed.IsSystemRole {
			existing, err = rs.roleRepo.GetGlobalSystemRole(ctx, roleSeed.Name)
		} else {
			if orgID, ok := entityMaps.Organizations[roleSeed.OrganizationSlug]; ok {
				existing, err = rs.roleRepo.GetByName(ctx, &orgID, roleSeed.Name)
			}
		}

		if err == nil && existing != nil {
			if verbose {
				log.Printf("   Role %s already exists, skipping", roleKey)
			}
			entityMaps.Roles[roleKey] = existing.ID
			continue
		}

		// Get organization ID for non-system roles
		var orgID *ulid.ULID
		if !roleSeed.IsSystemRole && roleSeed.OrganizationSlug != "" {
			if orgULID, ok := entityMaps.Organizations[roleSeed.OrganizationSlug]; ok {
				orgID = &orgULID
			} else {
				return fmt.Errorf("organization %s not found for role %s", roleSeed.OrganizationSlug, roleSeed.Name)
			}
		}

		// Create role entity
		role := &auth.Role{
			ID:             ulid.New(),
			Name:           roleSeed.Name,
			DisplayName:    roleSeed.DisplayName,
			Description:    roleSeed.Description,
			IsSystemRole:   roleSeed.IsSystemRole,
			OrganizationID: orgID,
		}

		// Create role in database
		if err := rs.roleRepo.Create(ctx, role); err != nil {
			return fmt.Errorf("failed to create role %s: %w", roleKey, err)
		}

		// Store role ID for later reference
		entityMaps.Roles[roleKey] = role.ID

		// Associate permissions with role
		if len(roleSeed.Permissions) > 0 {
			for _, permName := range roleSeed.Permissions {
				permID, ok := entityMaps.Permissions[permName]
				if !ok {
					return fmt.Errorf("permission %s not found for role %s", permName, roleKey)
				}

				// Create role-permission association
				if err := rs.rolePermRepo.AssignPermissions(ctx, role.ID, []ulid.ULID{permID}); err != nil {
					return fmt.Errorf("failed to add permission %s to role %s: %w", permName, roleKey, err)
				}
			}
		}

		if verbose {
			log.Printf("   ✅ Created role: %s (%s) with %d permissions", role.DisplayName, roleKey, len(roleSeed.Permissions))
		}
	}

	if verbose {
		log.Printf("✅ Roles seeded successfully")
	}
	return nil
}

// SeedMemberships seeds organization memberships from the provided seed data
func (rs *RBACSeeder) SeedMemberships(ctx context.Context, membershipSeeds []MembershipSeed, entityMaps *EntityMaps, verbose bool) error {
	if verbose {
		log.Printf("👥 Seeding %d memberships...", len(membershipSeeds))
	}

	for _, membershipSeed := range membershipSeeds {
		// Get user ID
		userID, ok := entityMaps.Users[membershipSeed.UserEmail]
		if !ok {
			return fmt.Errorf("user %s not found for membership", membershipSeed.UserEmail)
		}

		// Get organization ID
		orgID, ok := entityMaps.Organizations[membershipSeed.OrganizationSlug]
		if !ok {
			return fmt.Errorf("organization %s not found for membership", membershipSeed.OrganizationSlug)
		}

		// Get role ID - try both system and organization-specific roles
		var roleID ulid.ULID
		var found bool
		
		// First try organization-specific role
		orgRoleKey := fmt.Sprintf("%s:%s", membershipSeed.OrganizationSlug, membershipSeed.RoleName)
		if roleID, found = entityMaps.Roles[orgRoleKey]; !found {
			// Try system role
			systemRoleKey := fmt.Sprintf("system:%s", membershipSeed.RoleName)
			if roleID, found = entityMaps.Roles[systemRoleKey]; !found {
				return fmt.Errorf("role %s not found for membership (tried both %s and %s)", membershipSeed.RoleName, orgRoleKey, systemRoleKey)
			}
		}

		// Check if membership already exists
		existing, err := rs.memberRepo.GetByUserAndOrg(ctx, userID, orgID)
		if err == nil && existing != nil {
			if verbose {
				log.Printf("   Membership for %s in %s already exists, skipping", membershipSeed.UserEmail, membershipSeed.OrganizationSlug)
			}
			continue
		}

		// Create membership entity
		member := &organization.Member{
			ID:             ulid.New(),
			OrganizationID: orgID,
			UserID:         userID,
			RoleID:         roleID,
		}

		// Create membership in database
		if err := rs.memberRepo.Create(ctx, member); err != nil {
			return fmt.Errorf("failed to create membership for %s in %s: %w", membershipSeed.UserEmail, membershipSeed.OrganizationSlug, err)
		}

		if verbose {
			log.Printf("   ✅ Created membership: %s in %s as %s", membershipSeed.UserEmail, membershipSeed.OrganizationSlug, membershipSeed.RoleName)
		}
	}

	if verbose {
		log.Printf("✅ Memberships seeded successfully")
	}
	return nil
}