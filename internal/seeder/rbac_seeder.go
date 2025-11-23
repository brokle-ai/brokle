package seeder

import (
	"context"
	"fmt"
	"log"
	"strings"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// RBACSeeder handles seeding of normalized RBAC (roles, permissions, organization members) data
type RBACSeeder struct {
	roleRepo       auth.RoleRepository
	permissionRepo auth.PermissionRepository
	rolePermRepo   auth.RolePermissionRepository
	orgMemberRepo  auth.OrganizationMemberRepository
}

// NewRBACSeeder creates a new normalized RBACSeeder instance
func NewRBACSeeder(
	roleRepo auth.RoleRepository,
	permissionRepo auth.PermissionRepository,
	rolePermRepo auth.RolePermissionRepository,
	orgMemberRepo auth.OrganizationMemberRepository,
) *RBACSeeder {
	return &RBACSeeder{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		rolePermRepo:   rolePermRepo,
		orgMemberRepo:  orgMemberRepo,
	}
}

// SeedPermissions seeds normalized permissions from the provided seed data
func (rs *RBACSeeder) SeedPermissions(ctx context.Context, permissionSeeds []PermissionSeed, entityMaps *EntityMaps, verbose bool) error {
	if verbose {
		log.Printf("🔑 Seeding %d permissions...", len(permissionSeeds))
	}

	for _, permSeed := range permissionSeeds {
		// Parse resource and action from name (format: "resource:action")
		parts := strings.SplitN(permSeed.Name, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid permission name format: %s (expected resource:action)", permSeed.Name)
		}

		resource := parts[0]
		action := parts[1]

		// Check if permission already exists
		existing, err := rs.permissionRepo.GetByResourceAction(ctx, resource, action)
		if err == nil && existing != nil {
			if verbose {
				log.Printf("   Permission %s already exists, skipping", permSeed.Name)
			}
			entityMaps.Permissions[permSeed.Name] = existing.ID
			continue
		}

		// Determine scope level from resource name
		// Project-level resources: traces, analytics, models, providers, costs, prompts
		// Everything else: organization-level
		scopeLevel := auth.ScopeLevelOrganization
		category := resource

		projectResources := map[string]bool{
			"traces":          true,
			"analytics":       true,
			"provider_models": true,
			"providers":       true,
			"costs":           true,
			"prompts":         true,
		}

		if projectResources[resource] {
			scopeLevel = auth.ScopeLevelProject

			// Categorize project resources
			if resource == "traces" || resource == "analytics" || resource == "costs" {
				category = "observability"
			} else {
				category = "gateway"
			}
		}

		// Create new normalized permission with scope level
		permission := auth.NewPermissionWithScope(resource, action, permSeed.Description, scopeLevel, category)

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

// SeedRoles seeds template roles from the provided seed data
func (rs *RBACSeeder) SeedRoles(ctx context.Context, roleSeeds []RoleSeed, entityMaps *EntityMaps, verbose bool) error {
	if verbose {
		log.Printf("👤 Seeding %d template roles...", len(roleSeeds))
	}

	for _, roleSeed := range roleSeeds {
		// Check if template role already exists
		existing, err := rs.roleRepo.GetByNameAndScope(ctx, roleSeed.Name, roleSeed.ScopeType)
		if err == nil && existing != nil {
			if verbose {
				log.Printf("   Template role %s (%s) already exists, skipping", roleSeed.Name, roleSeed.ScopeType)
			}
			entityMaps.Roles[rs.getRoleKey(roleSeed)] = existing.ID
			continue
		}

		// Create new template role (available to all organizations)
		role := auth.NewRole(roleSeed.Name, roleSeed.ScopeType, roleSeed.Description)

		if err := rs.roleRepo.Create(ctx, role); err != nil {
			return fmt.Errorf("failed to create template role %s: %w", roleSeed.Name, err)
		}

		entityMaps.Roles[rs.getRoleKey(roleSeed)] = role.ID

		// Assign permissions to role template
		if len(roleSeed.Permissions) > 0 {
			var permissionIDs []ulid.ULID
			for _, permName := range roleSeed.Permissions {
				permID, exists := entityMaps.Permissions[permName]
				if !exists {
					return fmt.Errorf("permission not found: %s for role %s", permName, roleSeed.Name)
				}
				permissionIDs = append(permissionIDs, permID)
			}

			if err := rs.roleRepo.AssignRolePermissions(ctx, role.ID, permissionIDs, nil); err != nil {
				return fmt.Errorf("failed to assign permissions to role %s: %w", roleSeed.Name, err)
			}
		}

		if verbose {
			log.Printf("   ✓ Created template role: %s (%s scope) with %d permissions",
				roleSeed.Name, roleSeed.ScopeType, len(roleSeed.Permissions))
		}
	}

	return nil
}

// SeedOrganizationMemberships seeds organization memberships with roles from the provided seed data
func (rs *RBACSeeder) SeedOrganizationMemberships(ctx context.Context, membershipSeeds []MembershipSeed, entityMaps *EntityMaps, verbose bool) error {
	if verbose {
		log.Printf("🤝 Seeding %d organization memberships...", len(membershipSeeds))
	}

	for _, membershipSeed := range membershipSeeds {
		// Only process organization memberships
		if membershipSeed.ScopeType != auth.ScopeOrganization {
			continue
		}

		// Find user
		userID, exists := entityMaps.Users[membershipSeed.UserEmail]
		if !exists {
			return fmt.Errorf("user not found: %s", membershipSeed.UserEmail)
		}

		// Find organization
		orgID, exists := entityMaps.Organizations[membershipSeed.OrganizationName]
		if !exists {
			return fmt.Errorf("organization not found: %s", membershipSeed.OrganizationName)
		}

		// Find template role
		roleKey := fmt.Sprintf("%s:%s", membershipSeed.ScopeType, membershipSeed.RoleName)
		roleID, exists := entityMaps.Roles[roleKey]
		if !exists {
			return fmt.Errorf("template role not found for key: %s", roleKey)
		}

		// Check if membership already exists
		existing, err := rs.orgMemberRepo.GetByUserAndOrganization(ctx, userID, orgID)
		if err == nil && existing != nil {
			if verbose {
				log.Printf("   Membership already exists for %s in %s, skipping",
					membershipSeed.UserEmail, membershipSeed.OrganizationName)
			}
			continue
		}

		// Create organization membership with single role
		member := auth.NewOrganizationMember(userID, orgID, roleID, nil)

		if err := rs.orgMemberRepo.Create(ctx, member); err != nil {
			return fmt.Errorf("failed to create membership for %s in %s: %w",
				membershipSeed.UserEmail, membershipSeed.OrganizationName, err)
		}

		if verbose {
			log.Printf("   ✓ Added %s to %s with role %s",
				membershipSeed.UserEmail, membershipSeed.OrganizationName, membershipSeed.RoleName)
		}
	}

	return nil
}

// SeedMemberships is the main method that delegates to specific membership types
func (rs *RBACSeeder) SeedMemberships(ctx context.Context, membershipSeeds []MembershipSeed, entityMaps *EntityMaps, verbose bool) error {
	// Seed organization memberships
	if err := rs.SeedOrganizationMemberships(ctx, membershipSeeds, entityMaps, verbose); err != nil {
		return fmt.Errorf("failed to seed organization memberships: %w", err)
	}

	// TODO: Add project memberships when those are implemented
	// if err := rs.SeedProjectMemberships(ctx, membershipSeeds, entityMaps, verbose); err != nil {
	// 	return fmt.Errorf("failed to seed project memberships: %w", err)
	// }

	return nil
}

// Helper methods

func (rs *RBACSeeder) getRoleKey(roleSeed RoleSeed) string {
	// Template roles are keyed by scope_type:name (no specific entity IDs)
	return fmt.Sprintf("%s:%s", roleSeed.ScopeType, roleSeed.Name)
}

// GetRoleStatistics provides statistics about seeded roles (for verification)
func (rs *RBACSeeder) GetRoleStatistics(ctx context.Context) (*RoleStatistics, error) {
	allRoles, err := rs.roleRepo.GetAllRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all roles: %w", err)
	}

	stats := &RoleStatistics{
		TotalRoles:        len(allRoles),
		ScopeDistribution: make(map[string]int),
		RoleDistribution:  make(map[string]int),
	}

	for _, role := range allRoles {
		stats.ScopeDistribution[role.ScopeType]++
		stats.RoleDistribution[role.Name]++
	}

	return stats, nil
}

// RoleStatistics represents statistics about seeded roles
type RoleStatistics struct {
	ScopeDistribution map[string]int `json:"scope_distribution"`
	RoleDistribution  map[string]int `json:"role_distribution"`
	TotalRoles        int            `json:"total_roles"`
}
