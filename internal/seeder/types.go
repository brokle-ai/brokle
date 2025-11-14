package seeder

import "brokle/pkg/ulid"

// SeedData represents all the data to be seeded into the database
type SeedData struct {
	Organizations []OrganizationSeed `yaml:"organizations"`
	Users         []UserSeed         `yaml:"users"`
	RBAC          RBACSeeds          `yaml:"rbac"`
	Projects      []ProjectSeed      `yaml:"projects"`
}

// OrganizationSeed represents seed data for organizations
type OrganizationSeed struct {
	Name               string `yaml:"name"`
	BillingEmail       string `yaml:"billing_email"`
	Plan               string `yaml:"plan"`
	SubscriptionStatus string `yaml:"subscription_status"`
}

// UserSeed represents seed data for users
type UserSeed struct {
	Email         string `yaml:"email"`
	FirstName     string `yaml:"first_name"`
	LastName      string `yaml:"last_name"`
	Password      string `yaml:"password"`
	EmailVerified bool   `yaml:"email_verified"`
	IsActive      bool   `yaml:"is_active"`
	Timezone      string `yaml:"timezone"`
	Language      string `yaml:"language"`
}

// RBACSeeds represents all RBAC-related seed data
type RBACSeeds struct {
	Roles       []RoleSeed       `yaml:"roles"`
	Permissions []PermissionSeed `yaml:"permissions"`
	Memberships []MembershipSeed `yaml:"memberships"`
}

// RoleSeed represents clean seed data for roles with scope_type design
type RoleSeed struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	ScopeType   string   `yaml:"scope_type"` // 'organization' | 'project'
	Permissions []string `yaml:"permissions"`
}

// PermissionSeed represents seed data for permissions
type PermissionSeed struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// MembershipSeed represents clean seed data for user role assignments
type MembershipSeed struct {
	UserEmail        string `yaml:"user_email"`
	RoleName         string `yaml:"role_name"`
	ScopeType        string `yaml:"scope_type"`                  // 'system' | 'organization' | 'project'
	OrganizationName string `yaml:"organization_name,omitempty"` // Only for org/project scopes (changed from slug)
	ProjectName      string `yaml:"project_name,omitempty"`      // Only for project scope
}

// ProjectSeed represents seed data for projects
type ProjectSeed struct {
	Name             string `yaml:"name"`
	Description      string `yaml:"description"`
	OrganizationName string `yaml:"organization_name"` // Changed from slug to name
}

// Options represents the seeder configuration options
type Options struct {
	Environment string
	Reset       bool
	DryRun      bool
	Verbose     bool
}

// Internal maps for tracking created entities by their keys
type EntityMaps struct {
	Organizations map[string]ulid.ULID // name -> organization ID
	Users         map[string]ulid.ULID // email -> user ID
	Roles         map[string]ulid.ULID // org_name:role_name -> role ID
	Permissions   map[string]ulid.ULID // permission name -> permission ID
	Projects      map[string]ulid.ULID // org_name:project_name -> project ID
}

// NewEntityMaps creates a new EntityMaps instance with initialized maps
func NewEntityMaps() *EntityMaps {
	return &EntityMaps{
		Organizations: make(map[string]ulid.ULID),
		Users:         make(map[string]ulid.ULID),
		Roles:         make(map[string]ulid.ULID),
		Permissions:   make(map[string]ulid.ULID),
		Projects:      make(map[string]ulid.ULID),
	}
}
