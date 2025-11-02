package seeder

import (
	"context"
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"brokle/internal/config"
	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/organization"
	"brokle/internal/core/domain/user"
	"brokle/internal/infrastructure/database"
	authRepo "brokle/internal/infrastructure/repository/auth"
	orgRepo "brokle/internal/infrastructure/repository/organization"
	userRepo "brokle/internal/infrastructure/repository/user"
)

// Manager handles database seeding operations
type Manager struct {
	db  *gorm.DB
	cfg *config.Config

	// Repositories
	userRepo         user.Repository
	organizationRepo organization.OrganizationRepository
	memberRepo       organization.MemberRepository
	projectRepo      organization.ProjectRepository
	roleRepo         auth.RoleRepository
	orgMemberRepo    auth.OrganizationMemberRepository
	permissionRepo   auth.PermissionRepository
	rolePermRepo     auth.RolePermissionRepository

	// Component seeders
	userSeeder         *UserSeeder
	organizationSeeder *OrganizationSeeder
	rbacSeeder         *RBACSeeder
	projectSeeder      *ProjectSeeder
	onboardingSeeder   *OnboardingSeeder
}

// NewManager creates a new seeder manager with the required dependencies
func NewManager(cfg *config.Config) (*Manager, error) {
	// Create logger for database connection
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Less verbose for seeding

	// Initialize PostgreSQL database
	postgresDB, err := database.NewPostgresDB(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	manager := &Manager{
		db:  postgresDB.DB, // Use the GORM DB instance
		cfg: cfg,
	}

	// Initialize repositories
	manager.initializeRepositories()

	// Initialize component seeders
	manager.initializeComponentSeeders()

	return manager, nil
}

// initializeRepositories initializes all required repositories
func (m *Manager) initializeRepositories() {
	// User repositories
	m.userRepo = userRepo.NewUserRepository(m.db)

	// Organization repositories
	m.organizationRepo = orgRepo.NewOrganizationRepository(m.db)
	m.memberRepo = orgRepo.NewMemberRepository(m.db)
	m.projectRepo = orgRepo.NewProjectRepository(m.db)

	// Auth repositories
	m.roleRepo = authRepo.NewRoleRepository(m.db)
	m.orgMemberRepo = authRepo.NewOrganizationMemberRepository(m.db)
	m.permissionRepo = authRepo.NewPermissionRepository(m.db)
	m.rolePermRepo = authRepo.NewRolePermissionRepository(m.db)
}

// initializeComponentSeeders initializes all component seeders
func (m *Manager) initializeComponentSeeders() {
	m.userSeeder = NewUserSeeder(m.userRepo)
	m.organizationSeeder = NewOrganizationSeeder(m.organizationRepo)
	m.rbacSeeder = NewRBACSeeder(m.roleRepo, m.permissionRepo, m.rolePermRepo, m.orgMemberRepo)
	m.projectSeeder = NewProjectSeeder(m.projectRepo)
	m.onboardingSeeder = NewOnboardingSeeder(m.userRepo)
}

// SeedPostgres seeds PostgreSQL with the provided options
func (m *Manager) SeedPostgres(ctx context.Context, options *Options) error {
	if options.DryRun {
		fmt.Printf("🔍 DRY RUN: Would seed PostgreSQL with environment: %s\n", options.Environment)
		return nil
	}

	log.Printf("🌱 Starting PostgreSQL seeding with environment: %s", options.Environment)

	// Load seed data
	dataLoader := NewDataLoader()
	seedData, err := dataLoader.LoadSeedData(options.Environment)
	if err != nil {
		return fmt.Errorf("failed to load seed data: %w", err)
	}

	// Reset data if requested
	if options.Reset {
		log.Println("🧹 Resetting existing data...")
		if err := m.resetData(); err != nil {
			return fmt.Errorf("failed to reset data: %w", err)
		}
	}

	// Run seeding
	if err := m.seedData(ctx, seedData, options); err != nil {
		return fmt.Errorf("failed to seed data: %w", err)
	}

	log.Printf("✅ PostgreSQL seeding completed successfully")
	return nil
}

// SeedClickHouse is a placeholder for ClickHouse seeding (if needed in the future)
func (m *Manager) SeedClickHouse(ctx context.Context, options *Options) error {
	if options.DryRun {
		fmt.Printf("🔍 DRY RUN: Would seed ClickHouse with environment: %s\n", options.Environment)
		return nil
	}

	// ClickHouse seeding not implemented yet - it's primarily for analytics data
	log.Printf("ℹ️  ClickHouse seeding is not implemented (analytics data doesn't require seeding)")
	return nil
}

// seedData performs the actual seeding process
func (m *Manager) seedData(ctx context.Context, data *SeedData, options *Options) error {
	entityMaps := NewEntityMaps()

	log.Printf("🌱 Starting seeding process with %d organizations, %d users, %d permissions, %d roles", 
		len(data.Organizations), len(data.Users), len(data.RBAC.Permissions), len(data.RBAC.Roles))

	// Execute seeding in proper dependency order
	
	// 1. Permissions first (system-level, no dependencies)
	if err := m.rbacSeeder.SeedPermissions(ctx, data.RBAC.Permissions, entityMaps, options.Verbose); err != nil {
		return fmt.Errorf("failed to seed permissions: %w", err)
	}

	// 2. Organizations
	if err := m.organizationSeeder.SeedOrganizations(ctx, data.Organizations, entityMaps, options.Verbose); err != nil {
		return fmt.Errorf("failed to seed organizations: %w", err)
	}

	// 3. Users
	if err := m.userSeeder.SeedUsers(ctx, data.Users, entityMaps, options.Verbose); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	// 4. Roles (depends on organizations and permissions)
	if err := m.rbacSeeder.SeedRoles(ctx, data.RBAC.Roles, entityMaps, options.Verbose); err != nil {
		return fmt.Errorf("failed to seed roles: %w", err)
	}

	// 5. Memberships (depends on users, organizations, and roles)
	if err := m.rbacSeeder.SeedMemberships(ctx, data.RBAC.Memberships, entityMaps, options.Verbose); err != nil {
		return fmt.Errorf("failed to seed memberships: %w", err)
	}

	// 6. Projects and environments
	if err := m.projectSeeder.SeedProjects(ctx, data.Projects, entityMaps, options.Verbose); err != nil {
		return fmt.Errorf("failed to seed projects: %w", err)
	}

	// 7. API keys are not supported in seeding (removed due to PostgreSQL JSON issues)

	// 8. Onboarding questions
	if err := m.onboardingSeeder.SeedOnboardingQuestions(ctx, data.OnboardingQuestions, entityMaps, options.Verbose); err != nil {
		return fmt.Errorf("failed to seed onboarding questions: %w", err)
	}

	log.Println("✅ Seeding process completed successfully")
	return nil
}

// resetData clears all existing data from the database
func (m *Manager) resetData() error {
	log.Println("🧹 Starting data reset...")

	// Delete in reverse dependency order to avoid foreign key constraints
	tables := []string{
		"user_onboarding_responses",
		"onboarding_questions",
		"environments",
		"projects",
		"organization_members",
		"role_permissions",
		"roles",
		"permissions",
		"user_invitations",
		"password_reset_tokens",
		"email_verification_tokens",
		"user_profiles",
		"user_preferences",
		"sessions",
		"audit_logs",
		"users",
		"organizations",
	}

	for _, table := range tables {
		if err := m.db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)).Error; err != nil {
			// Log warning but continue - table might not exist or might be empty
			log.Printf("⚠️  Warning: Could not truncate table %s: %v", table, err)
		} else {
			log.Printf("🗑️  Truncated table: %s", table)
		}
	}

	log.Println("✅ Data reset completed")
	return nil
}

// Close closes the database connections
func (m *Manager) Close() error {
	if sqlDB, err := m.db.DB(); err == nil {
		return sqlDB.Close()
	}
	return nil
}

// PrintSeedPlan prints a detailed plan of what will be seeded
func (m *Manager) PrintSeedPlan(data *SeedData) {
	fmt.Println("\n📋 SEED PLAN:")
	fmt.Println("=====================================")

	fmt.Printf("Organizations: %d\n", len(data.Organizations))
	for _, org := range data.Organizations {
		fmt.Printf("  - %s - Plan: %s\n", org.Name, org.Plan)
	}

	fmt.Printf("\nUsers: %d\n", len(data.Users))
	for _, user := range data.Users {
		fmt.Printf("  - %s %s (%s)\n", user.FirstName, user.LastName, user.Email)
	}

	fmt.Printf("\nPermissions: %d\n", len(data.RBAC.Permissions))
	for _, permission := range data.RBAC.Permissions {
		fmt.Printf("  - %s\n", permission.Name)
	}

	fmt.Printf("\nRoles: %d\n", len(data.RBAC.Roles))
	for _, role := range data.RBAC.Roles {
		fmt.Printf("  - %s (%s) - %d permissions\n", role.Name, role.ScopeType, len(role.Permissions))
	}

	fmt.Printf("\nProjects: %d\n", len(data.Projects))
	for _, project := range data.Projects {
		fmt.Printf("  - %s (Org: %s)\n", project.Name, project.OrganizationName)
	}

	fmt.Printf("\nMemberships: %d\n", len(data.RBAC.Memberships))
	for _, membership := range data.RBAC.Memberships {
		fmt.Printf("  - %s in %s as %s\n", membership.UserEmail, membership.OrganizationName, membership.RoleName)
	}

	fmt.Printf("\nOnboarding Questions: %d\n", len(data.OnboardingQuestions))
	for _, question := range data.OnboardingQuestions {
		required := "optional"
		if question.IsRequired {
			required = "required"
		}
		fmt.Printf("  - Step %d: %s (%s, %s)\n", question.Step, question.Title, question.QuestionType, required)
	}

	fmt.Println("=====================================")
}