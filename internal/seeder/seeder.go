package seeder

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

	"brokle/internal/config"
	"brokle/internal/core/domain/analytics"
	"brokle/internal/core/domain/auth"
	"brokle/internal/infrastructure/database"
	analyticsRepo "brokle/internal/infrastructure/repository/analytics"
	authRepo "brokle/internal/infrastructure/repository/auth"
	"brokle/pkg/logging"
	"brokle/pkg/ulid"
)

// Seeder handles all database seeding operations
type Seeder struct {
	db     *gorm.DB
	cfg    *config.Config
	logger *slog.Logger

	// Repositories
	roleRepo          auth.RoleRepository
	permissionRepo    auth.PermissionRepository
	rolePermRepo      auth.RolePermissionRepository
	providerModelRepo analytics.ProviderModelRepository
}

// New creates a new Seeder with the required dependencies
func New(cfg *config.Config) (*Seeder, error) {
	// Create logger for seeding - use Info level so progress and verbose output are visible
	logger := logging.NewLoggerWithFormat(slog.LevelInfo, cfg.Logging.Format)

	// Initialize PostgreSQL database
	postgresDB, err := database.NewPostgresDB(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	s := &Seeder{
		db:     postgresDB.DB,
		cfg:    cfg,
		logger: logger,
	}

	// Initialize repositories
	s.roleRepo = authRepo.NewRoleRepository(s.db)
	s.permissionRepo = authRepo.NewPermissionRepository(s.db)
	s.rolePermRepo = authRepo.NewRolePermissionRepository(s.db)
	s.providerModelRepo = analyticsRepo.NewProviderModelRepository(s.db)

	return s, nil
}

// Close closes the database connection
func (s *Seeder) Close() error {
	if sqlDB, err := s.db.DB(); err == nil {
		return sqlDB.Close()
	}
	return nil
}

// ============================================================================
// Public API
// ============================================================================

// SeedAll seeds all system data (permissions, roles, pricing)
func (s *Seeder) SeedAll(ctx context.Context, opts *Options) error {
	if opts.DryRun {
		fmt.Println("DRY RUN: Would seed PostgreSQL with permissions, roles, and pricing")
		return nil
	}

	s.logger.Info("Starting PostgreSQL seeding...")

	// Reset data if requested
	if opts.Reset {
		s.logger.Info("Resetting existing data...")
		if err := s.Reset(ctx, opts.Verbose); err != nil {
			return fmt.Errorf("failed to reset data: %w", err)
		}
	}

	// Load and seed RBAC data
	permissions, err := s.loadPermissions()
	if err != nil {
		return fmt.Errorf("failed to load permissions: %w", err)
	}

	roles, err := s.loadRoles()
	if err != nil {
		return fmt.Errorf("failed to load roles: %w", err)
	}

	entityMaps := NewEntityMaps()

	s.logger.Debug("Starting seeding process", "permissions", len(permissions), "roles", len(roles))

	// 1. Seed permissions (no dependencies)
	if err := s.seedPermissions(ctx, permissions, entityMaps, opts.Verbose); err != nil {
		return fmt.Errorf("failed to seed permissions: %w", err)
	}

	// 2. Seed roles (depends on permissions)
	if err := s.seedRoles(ctx, roles, entityMaps, opts.Verbose); err != nil {
		return fmt.Errorf("failed to seed roles: %w", err)
	}

	// 3. Seed pricing (independent)
	if err := s.seedPricingFromFile(ctx, opts.Verbose); err != nil {
		return fmt.Errorf("failed to seed pricing: %w", err)
	}

	s.logger.Info("PostgreSQL seeding completed successfully")
	return nil
}

// SeedRBAC seeds only permissions and roles
func (s *Seeder) SeedRBAC(ctx context.Context, opts *Options) error {
	if opts.DryRun {
		fmt.Println("DRY RUN: Would seed RBAC (permissions and roles)")
		return nil
	}

	s.logger.Info("Starting RBAC seeding...")

	// Reset RBAC data if requested
	if opts.Reset {
		s.logger.Info("Resetting existing RBAC data...")
		if err := s.resetRBAC(ctx, opts.Verbose); err != nil {
			return fmt.Errorf("failed to reset RBAC data: %w", err)
		}
	}

	// Load seed data
	permissions, err := s.loadPermissions()
	if err != nil {
		return fmt.Errorf("failed to load permissions: %w", err)
	}

	roles, err := s.loadRoles()
	if err != nil {
		return fmt.Errorf("failed to load roles: %w", err)
	}

	entityMaps := NewEntityMaps()

	// Seed permissions
	if err := s.seedPermissions(ctx, permissions, entityMaps, opts.Verbose); err != nil {
		return fmt.Errorf("failed to seed permissions: %w", err)
	}

	// Seed roles
	if err := s.seedRoles(ctx, roles, entityMaps, opts.Verbose); err != nil {
		return fmt.Errorf("failed to seed roles: %w", err)
	}

	// Print statistics
	stats, err := s.GetRBACStatistics(ctx)
	if err == nil {
		s.logger.Debug("RBAC Statistics", "permissions", stats.TotalPermissions, "roles", stats.TotalRoles)
	}

	s.logger.Info("RBAC seeding completed successfully")
	return nil
}

// SeedPricing seeds only provider pricing data
func (s *Seeder) SeedPricing(ctx context.Context, opts *Options) error {
	if opts.DryRun {
		fmt.Println("DRY RUN: Would seed provider pricing")
		return nil
	}

	s.logger.Info("Starting provider pricing seeding...")

	// Reset pricing data if requested
	if opts.Reset {
		s.logger.Info("Resetting existing pricing data...")
		if err := s.resetPricing(ctx, opts.Verbose); err != nil {
			return fmt.Errorf("failed to reset pricing data: %w", err)
		}
	}

	// Load and seed pricing data
	if err := s.seedPricingFromFile(ctx, opts.Verbose); err != nil {
		return fmt.Errorf("failed to seed pricing: %w", err)
	}

	// Print statistics
	stats, err := s.GetPricingStatistics(ctx)
	if err == nil {
		s.logger.Debug("Pricing Statistics", "models", stats.TotalModels, "prices", stats.TotalPrices)
	}

	s.logger.Info("Provider pricing seeding completed successfully")
	return nil
}

// Reset removes all seeded data
func (s *Seeder) Reset(ctx context.Context, verbose bool) error {
	s.logger.Info("Starting data reset...")

	// Reset RBAC data
	if err := s.resetRBAC(ctx, verbose); err != nil {
		s.logger.Warn(" Could not reset RBAC data", "error", err)
	}

	// Reset pricing data
	if err := s.resetPricing(ctx, verbose); err != nil {
		s.logger.Warn(" Could not reset pricing data", "error", err)
	}

	s.logger.Info("Data reset completed")
	return nil
}

// PrintSeedPlan prints a detailed plan of what will be seeded
func (s *Seeder) PrintSeedPlan(data *SeedData) {
	fmt.Println("\nSEED PLAN:")
	fmt.Println("=====================================")

	fmt.Printf("\nPermissions: %d\n", len(data.Permissions))
	for _, perm := range data.Permissions {
		fmt.Printf("  - %s\n", perm.Name)
	}

	fmt.Printf("\nRoles: %d\n", len(data.Roles))
	for _, role := range data.Roles {
		fmt.Printf("  - %s (%s scope) - %d permissions\n", role.Name, role.ScopeType, len(role.Permissions))
	}

	fmt.Println("=====================================")
}

// LoadSeedData loads permissions and roles from YAML files (for dry-run preview)
func (s *Seeder) LoadSeedData() (*SeedData, error) {
	permissions, err := s.loadPermissions()
	if err != nil {
		return nil, fmt.Errorf("failed to load permissions: %w", err)
	}

	roles, err := s.loadRoles()
	if err != nil {
		return nil, fmt.Errorf("failed to load roles: %w", err)
	}

	return &SeedData{
		Permissions: permissions,
		Roles:       roles,
	}, nil
}

// ============================================================================
// File Loading
// ============================================================================

// loadPermissions loads permissions from seeds/permissions.yaml
func (s *Seeder) loadPermissions() ([]PermissionSeed, error) {
	seedFile := findSeedFile("seeds/permissions.yaml")
	if seedFile == "" {
		return nil, errors.New("permissions file not found: seeds/permissions.yaml")
	}

	data, err := os.ReadFile(seedFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", seedFile, err)
	}

	var permissionsFile PermissionsFile
	if err := yaml.Unmarshal(data, &permissionsFile); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", seedFile, err)
	}

	// Validate
	permissionNames := make(map[string]bool)
	for _, permission := range permissionsFile.Permissions {
		if permission.Name == "" {
			return nil, errors.New("permission missing required field: name")
		}
		if permissionNames[permission.Name] {
			return nil, fmt.Errorf("duplicate permission name: %s", permission.Name)
		}
		permissionNames[permission.Name] = true
	}

	return permissionsFile.Permissions, nil
}

// loadRoles loads roles from seeds/roles.yaml
func (s *Seeder) loadRoles() ([]RoleSeed, error) {
	seedFile := findSeedFile("seeds/roles.yaml")
	if seedFile == "" {
		return nil, errors.New("roles file not found: seeds/roles.yaml")
	}

	data, err := os.ReadFile(seedFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", seedFile, err)
	}

	var rolesFile RolesFile
	if err := yaml.Unmarshal(data, &rolesFile); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", seedFile, err)
	}

	// Validate
	roleNames := make(map[string]bool)
	for _, role := range rolesFile.Roles {
		if role.Name == "" || role.ScopeType == "" {
			return nil, errors.New("role missing required fields (name, scope_type)")
		}
		if roleNames[role.Name] {
			return nil, fmt.Errorf("duplicate role name: %s", role.Name)
		}
		roleNames[role.Name] = true
	}

	return rolesFile.Roles, nil
}

// loadPricing loads pricing from seeds/pricing.yaml
func (s *Seeder) loadPricing() (*ProviderPricingSeedData, error) {
	seedFile := findSeedFile("seeds/pricing.yaml")
	if seedFile == "" {
		return nil, errors.New("pricing file not found: seeds/pricing.yaml")
	}

	data, err := os.ReadFile(seedFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", seedFile, err)
	}

	var pricingData ProviderPricingSeedData
	if err := yaml.Unmarshal(data, &pricingData); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", seedFile, err)
	}

	// Validate
	if len(pricingData.ProviderModels) == 0 {
		return nil, errors.New("no provider models defined")
	}

	modelNames := make(map[string]bool)
	for i, model := range pricingData.ProviderModels {
		if model.ModelName == "" {
			return nil, fmt.Errorf("model %d missing required field: model_name", i)
		}
		if model.MatchPattern == "" {
			return nil, fmt.Errorf("model %s missing required field: match_pattern", model.ModelName)
		}
		if model.StartDate == "" {
			return nil, fmt.Errorf("model %s missing required field: start_date", model.ModelName)
		}
		if modelNames[model.ModelName] {
			return nil, fmt.Errorf("duplicate model_name: %s", model.ModelName)
		}
		modelNames[model.ModelName] = true

		if len(model.Prices) == 0 {
			return nil, fmt.Errorf("model %s has no prices defined", model.ModelName)
		}

		usageTypes := make(map[string]bool)
		for _, price := range model.Prices {
			if price.UsageType == "" {
				return nil, fmt.Errorf("model %s has price with empty usage_type", model.ModelName)
			}
			if usageTypes[price.UsageType] {
				return nil, fmt.Errorf("model %s has duplicate usage_type: %s", model.ModelName, price.UsageType)
			}
			usageTypes[price.UsageType] = true
			if price.Price < 0 {
				return nil, fmt.Errorf("model %s has negative price for %s", model.ModelName, price.UsageType)
			}
		}
	}

	return &pricingData, nil
}

// findSeedFile finds the seed file in current dir or brokle subdir
func findSeedFile(seedFile string) string {
	if _, err := os.Stat(seedFile); err == nil {
		return seedFile
	}
	broklePath := filepath.Join("brokle", seedFile)
	if _, err := os.Stat(broklePath); err == nil {
		return broklePath
	}
	return ""
}

// ============================================================================
// Seeding Logic
// ============================================================================

// seedPermissions seeds permissions from the provided seed data
func (s *Seeder) seedPermissions(ctx context.Context, permissionSeeds []PermissionSeed, entityMaps *EntityMaps, verbose bool) error {
	if verbose {
		s.logger.Info("Seeding permissions", "count", len(permissionSeeds))
	}

	for _, permSeed := range permissionSeeds {
		// Parse resource and action from name (format: "resource:action")
		parts := strings.SplitN(permSeed.Name, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid permission name format: %s (expected resource:action)", permSeed.Name)
		}

		resource := parts[0]
		action := parts[1]

		// Check if permission already exists (idempotent)
		existing, err := s.permissionRepo.GetByResourceAction(ctx, resource, action)
		if err == nil && existing != nil {
			if verbose {
				s.logger.Info("Permission already exists, skipping", "name", permSeed.Name)
			}
			entityMaps.Permissions[permSeed.Name] = existing.ID
			continue
		}

		// Determine scope level from resource name
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

			if resource == "traces" || resource == "analytics" || resource == "costs" {
				category = "observability"
			} else {
				category = "gateway"
			}
		}

		// Create new permission with scope level
		permission := auth.NewPermissionWithScope(resource, action, permSeed.Description, scopeLevel, category)

		if err := s.permissionRepo.Create(ctx, permission); err != nil {
			return fmt.Errorf("failed to create permission %s: %w", permSeed.Name, err)
		}

		entityMaps.Permissions[permSeed.Name] = permission.ID

		if verbose {
			s.logger.Info("Created permission", "name", permSeed.Name)
		}
	}

	return nil
}

// seedRoles seeds template roles from the provided seed data
func (s *Seeder) seedRoles(ctx context.Context, roleSeeds []RoleSeed, entityMaps *EntityMaps, verbose bool) error {
	if verbose {
		s.logger.Info("Seeding template roles", "count", len(roleSeeds))
	}

	for _, roleSeed := range roleSeeds {
		// Check if template role already exists
		existing, err := s.roleRepo.GetByNameAndScope(ctx, roleSeed.Name, roleSeed.ScopeType)
		if err == nil && existing != nil {
			// Role exists - sync permissions to match YAML definition
			entityMaps.Roles[fmt.Sprintf("%s:%s", roleSeed.ScopeType, roleSeed.Name)] = existing.ID

			// Build permission ID list from YAML
			var permissionIDs []ulid.ULID
			for _, permName := range roleSeed.Permissions {
				permID, exists := entityMaps.Permissions[permName]
				if !exists {
					return fmt.Errorf("permission not found: %s for role %s", permName, roleSeed.Name)
				}
				permissionIDs = append(permissionIDs, permID)
			}

			// Sync permissions (replaces all existing with YAML definition)
			if err := s.roleRepo.UpdateRolePermissions(ctx, existing.ID, permissionIDs, nil); err != nil {
				return fmt.Errorf("failed to sync permissions for role %s: %w", roleSeed.Name, err)
			}

			if verbose {
				s.logger.Info("Synced role", "name", roleSeed.Name, "scope", roleSeed.ScopeType, "permissions", len(roleSeed.Permissions))
			}
			continue
		}

		// Create new template role
		role := auth.NewRole(roleSeed.Name, roleSeed.ScopeType, roleSeed.Description)

		if err := s.roleRepo.Create(ctx, role); err != nil {
			return fmt.Errorf("failed to create template role %s: %w", roleSeed.Name, err)
		}

		entityMaps.Roles[fmt.Sprintf("%s:%s", roleSeed.ScopeType, roleSeed.Name)] = role.ID

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

			if err := s.roleRepo.AssignRolePermissions(ctx, role.ID, permissionIDs, nil); err != nil {
				return fmt.Errorf("failed to assign permissions to role %s: %w", roleSeed.Name, err)
			}
		}

		if verbose {
			s.logger.Info("Created template role", "name", roleSeed.Name, "scope", roleSeed.ScopeType, "permissions", len(roleSeed.Permissions))
		}
	}

	return nil
}

// seedPricingFromFile loads and seeds pricing data
func (s *Seeder) seedPricingFromFile(ctx context.Context, verbose bool) error {
	pricingData, err := s.loadPricing()
	if err != nil {
		// Pricing data is optional - log warning but don't fail
		s.logger.Warn(" Could not load pricing data", "error", err)
		return nil
	}

	return s.seedPricingData(ctx, pricingData, verbose)
}

// seedPricingData seeds provider models and prices
func (s *Seeder) seedPricingData(ctx context.Context, data *ProviderPricingSeedData, verbose bool) error {
	if verbose {
		s.logger.Info("Seeding provider models with pricing", "count", len(data.ProviderModels))
	}

	for _, modelSeed := range data.ProviderModels {
		if err := s.seedModel(ctx, modelSeed, verbose); err != nil {
			return fmt.Errorf("failed to seed model %s: %w", modelSeed.ModelName, err)
		}
	}

	if verbose {
		s.logger.Info("Provider pricing seeded successfully", "version", data.Version)
	}
	return nil
}

// seedModel seeds a single provider model with its prices
func (s *Seeder) seedModel(ctx context.Context, modelSeed ProviderModelSeed, verbose bool) error {
	// Parse start date
	startDate, err := time.Parse("2006-01-02", modelSeed.StartDate)
	if err != nil {
		return fmt.Errorf("invalid start_date format: %w", err)
	}

	// Check if model already exists (idempotent seeding)
	existing, _ := s.providerModelRepo.GetProviderModelByName(ctx, nil, modelSeed.ModelName)
	if existing != nil {
		if verbose {
			s.logger.Info("Model already exists, skipping", "name", modelSeed.ModelName, "id", existing.ID.String())
		}
		return nil
	}

	// Determine unit (default to TOKENS)
	unit := modelSeed.Unit
	if unit == "" {
		unit = "TOKENS"
	}

	// Create provider model with new ULID
	model := &analytics.ProviderModel{
		ID:           ulid.New(),
		ModelName:    modelSeed.ModelName,
		MatchPattern: modelSeed.MatchPattern,
		StartDate:    startDate,
		Unit:         unit,
	}

	// Set tokenizer fields if provided
	if modelSeed.TokenizerID != "" {
		model.TokenizerID = &modelSeed.TokenizerID
	}
	if len(modelSeed.TokenizerConfig) > 0 {
		model.TokenizerConfig = modelSeed.TokenizerConfig
	}

	// Create model in database
	if err := s.providerModelRepo.CreateProviderModel(ctx, model); err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	if verbose {
		s.logger.Info("Created model", "name", model.ModelName, "id", model.ID.String())
	}

	// Create prices for this model
	for _, priceSeed := range modelSeed.Prices {
		price := &analytics.ProviderPrice{
			ID:              ulid.New(),
			ProviderModelID: model.ID,
			UsageType:       priceSeed.UsageType,
			Price:           decimal.NewFromFloat(priceSeed.Price),
		}

		if err := s.providerModelRepo.CreateProviderPrice(ctx, price); err != nil {
			return fmt.Errorf("failed to create price %s: %w", priceSeed.UsageType, err)
		}

		if verbose {
			s.logger.Info("Created price", "usage_type", priceSeed.UsageType, "price", priceSeed.Price)
		}
	}

	return nil
}

// ============================================================================
// Reset Logic
// ============================================================================

// resetRBAC removes all existing RBAC data
func (s *Seeder) resetRBAC(ctx context.Context, verbose bool) error {
	if verbose {
		s.logger.Info("Resetting RBAC data...")
	}

	// Get all roles and delete them
	roles, err := s.roleRepo.GetAllRoles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list roles for reset: %w", err)
	}

	deletedRoles := 0
	for _, role := range roles {
		if err := s.roleRepo.Delete(ctx, role.ID); err != nil {
			s.logger.Warn(" Could not delete role", "name", role.Name, "error", err)
		} else {
			deletedRoles++
			if verbose {
				s.logger.Info("Deleted role", "name", role.Name)
			}
		}
	}

	// Get all permissions and delete them
	permissions, err := s.permissionRepo.GetAllPermissions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list permissions for reset: %w", err)
	}

	deletedPerms := 0
	for _, perm := range permissions {
		if err := s.permissionRepo.Delete(ctx, perm.ID); err != nil {
			s.logger.Warn(" Could not delete permission", "resource", perm.Resource, "action", perm.Action, "error", err)
		} else {
			deletedPerms++
			if verbose {
				s.logger.Info("Deleted permission", "resource", perm.Resource, "action", perm.Action)
			}
		}
	}

	if verbose {
		s.logger.Info("RBAC reset completed", "roles_deleted", deletedRoles, "permissions_deleted", deletedPerms)
	}
	return nil
}

// resetPricing removes all existing pricing data
func (s *Seeder) resetPricing(ctx context.Context, verbose bool) error {
	if verbose {
		s.logger.Info("Resetting provider pricing data...")
	}

	// Get all models and delete them (cascade deletes prices)
	models, err := s.providerModelRepo.ListProviderModels(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list models for reset: %w", err)
	}

	deletedCount := 0
	for _, model := range models {
		if err := s.providerModelRepo.DeleteProviderModel(ctx, model.ID); err != nil {
			s.logger.Warn(" Could not delete model", "name", model.ModelName, "error", err)
		} else {
			deletedCount++
			if verbose {
				s.logger.Info("Deleted model", "name", model.ModelName)
			}
		}
	}

	if verbose {
		s.logger.Info("Provider pricing reset completed", "models_deleted", deletedCount)
	}
	return nil
}

// ============================================================================
// Statistics
// ============================================================================

// GetRBACStatistics provides statistics about seeded RBAC data
func (s *Seeder) GetRBACStatistics(ctx context.Context) (*RBACStatistics, error) {
	allRoles, err := s.roleRepo.GetAllRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all roles: %w", err)
	}

	allPerms, err := s.permissionRepo.GetAllPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all permissions: %w", err)
	}

	stats := &RBACStatistics{
		TotalRoles:        len(allRoles),
		TotalPermissions:  len(allPerms),
		ScopeDistribution: make(map[string]int),
		RoleDistribution:  make(map[string]int),
	}

	for _, role := range allRoles {
		stats.ScopeDistribution[role.ScopeType]++
		stats.RoleDistribution[role.Name]++
	}

	return stats, nil
}

// GetPricingStatistics provides statistics about seeded pricing
func (s *Seeder) GetPricingStatistics(ctx context.Context) (*PricingStatistics, error) {
	models, err := s.providerModelRepo.ListProviderModels(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %w", err)
	}

	stats := &PricingStatistics{
		TotalModels:          len(models),
		ProviderDistribution: make(map[string]int),
	}

	for _, model := range models {
		provider := InferProvider(model.ModelName)
		stats.ProviderDistribution[provider]++

		prices, err := s.providerModelRepo.GetProviderPrices(ctx, model.ID, nil)
		if err == nil {
			stats.TotalPrices += len(prices)
		}
	}

	return stats, nil
}
