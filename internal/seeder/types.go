package seeder

import "brokle/pkg/ulid"

// SeedData represents all the data to be seeded into the database
type SeedData struct {
	Organizations       []OrganizationSeed   `yaml:"organizations"`
	Users               []UserSeed           `yaml:"users"`
	RBAC                RBACSeeds            `yaml:"rbac"`
	Projects            []ProjectSeed        `yaml:"projects"`
	OnboardingQuestions []OnboardingSeed     `yaml:"onboarding_questions"`
}

// OrganizationSeed represents seed data for organizations
type OrganizationSeed struct {
	Name               string `yaml:"name"`
	Slug               string `yaml:"slug"`
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
	ScopeType   string   `yaml:"scope_type"` // 'organization' | 'project' | 'environment'
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
	ScopeType        string `yaml:"scope_type"`                   // 'system' | 'organization' | 'project'
	OrganizationSlug string `yaml:"organization_slug,omitempty"`  // Only for org/project scopes
	ProjectName      string `yaml:"project_name,omitempty"`       // Only for project scope
}

// ProjectSeed represents seed data for projects
type ProjectSeed struct {
	Name             string            `yaml:"name"`
	Description      string            `yaml:"description"`
	OrganizationSlug string            `yaml:"organization_slug"`
	Environments     []EnvironmentSeed `yaml:"environments"`
}

// EnvironmentSeed represents seed data for environments
type EnvironmentSeed struct {
	Name string `yaml:"name"`
	Slug string `yaml:"slug"`
}

// OnboardingSeed represents seed data for onboarding questions
type OnboardingSeed struct {
	Step         int      `yaml:"step"`
	QuestionType string   `yaml:"question_type"`
	Title        string   `yaml:"title"`
	Description  string   `yaml:"description"`
	IsRequired   bool     `yaml:"is_required"`
	Options      []string `yaml:"options"`
	DisplayOrder int      `yaml:"display_order"`
	IsActive     bool     `yaml:"is_active"`
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
	Organizations map[string]ulid.ULID // slug -> organization ID
	Users         map[string]ulid.ULID // email -> user ID
	Roles         map[string]ulid.ULID // org_slug:role_name -> role ID
	Permissions   map[string]ulid.ULID // permission name -> permission ID
	Projects      map[string]ulid.ULID // org_slug:project_name -> project ID
	Environments  map[string]ulid.ULID // project_id:env_name -> environment ID
}

// NewEntityMaps creates a new EntityMaps instance with initialized maps
func NewEntityMaps() *EntityMaps {
	return &EntityMaps{
		Organizations: make(map[string]ulid.ULID),
		Users:         make(map[string]ulid.ULID),
		Roles:         make(map[string]ulid.ULID),
		Permissions:   make(map[string]ulid.ULID),
		Projects:      make(map[string]ulid.ULID),
		Environments:  make(map[string]ulid.ULID),
	}
}