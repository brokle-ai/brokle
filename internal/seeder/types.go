package seeder

import (
	"encoding/json"

	"brokle/pkg/ulid"
)

// ============================================================================
// File Wrapper Types (for loading separate YAML files)
// ============================================================================

// PermissionsFile wraps the permissions.yaml structure
type PermissionsFile struct {
	Permissions []PermissionSeed `yaml:"permissions"`
}

// RolesFile wraps the roles.yaml structure
type RolesFile struct {
	Roles []RoleSeed `yaml:"roles"`
}

// ============================================================================
// Seed Data Types
// ============================================================================

// SeedData represents all the data to be seeded into the database
type SeedData struct {
	Permissions []PermissionSeed
	Roles       []RoleSeed
}

// PermissionSeed represents seed data for permissions
type PermissionSeed struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// RoleSeed represents seed data for roles with scope_type design
type RoleSeed struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	ScopeType   string   `yaml:"scope_type"` // 'organization' | 'project'
	Permissions []string `yaml:"permissions"`
}

// ============================================================================
// Seeder Options
// ============================================================================

// Options represents the seeder configuration options
type Options struct {
	Reset   bool
	DryRun  bool
	Verbose bool
}

// ============================================================================
// Entity Maps (for tracking created entities)
// ============================================================================

// EntityMaps holds internal maps for tracking created entities by their keys
type EntityMaps struct {
	Permissions map[string]ulid.ULID // permission name -> permission ID
	Roles       map[string]ulid.ULID // role_name -> role ID
}

// NewEntityMaps creates a new EntityMaps instance with initialized maps
func NewEntityMaps() *EntityMaps {
	return &EntityMaps{
		Permissions: make(map[string]ulid.ULID),
		Roles:       make(map[string]ulid.ULID),
	}
}

// ============================================================================
// Provider Pricing Seed Types
// ============================================================================

// ProviderPricingSeedData represents all provider pricing data to be seeded
type ProviderPricingSeedData struct {
	Version        string              `yaml:"version"`
	ProviderModels []ProviderModelSeed `yaml:"provider_models"`
}

// ProviderModelSeed represents seed data for an AI provider model
type ProviderModelSeed struct {
	ModelName       string                 `yaml:"model_name"`
	MatchPattern    string                 `yaml:"match_pattern"`
	StartDate       string                 `yaml:"start_date"` // Format: "2024-05-13"
	Unit            string                 `yaml:"unit"`       // Default: "TOKENS"
	TokenizerID     string                 `yaml:"tokenizer_id,omitempty"`
	TokenizerConfig map[string]interface{} `yaml:"tokenizer_config,omitempty"`
	Prices          []PriceSeed            `yaml:"prices"`
}

// PriceSeed represents seed data for a provider price
type PriceSeed struct {
	UsageType string  `yaml:"usage_type"` // "input", "output", "cache_read_input_tokens", etc.
	Price     float64 `yaml:"price"`      // Price per 1M tokens
}

// ============================================================================
// Statistics Types
// ============================================================================

// RBACStatistics represents statistics about seeded RBAC data
type RBACStatistics struct {
	TotalRoles        int            `json:"total_roles"`
	TotalPermissions  int            `json:"total_permissions"`
	ScopeDistribution map[string]int `json:"scope_distribution"`
	RoleDistribution  map[string]int `json:"role_distribution"`
}

// String returns a formatted string representation of RBAC statistics
func (s *RBACStatistics) String() string {
	data, _ := json.MarshalIndent(s, "", "  ")
	return string(data)
}

// PricingStatistics represents statistics about seeded pricing
type PricingStatistics struct {
	TotalModels          int            `json:"total_models"`
	TotalPrices          int            `json:"total_prices"`
	ProviderDistribution map[string]int `json:"provider_distribution"`
}

// String returns a formatted string representation of pricing statistics
func (s *PricingStatistics) String() string {
	data, _ := json.MarshalIndent(s, "", "  ")
	return string(data)
}

// InferProvider infers the provider name from model name
func InferProvider(modelName string) string {
	switch {
	case len(modelName) >= 3 && modelName[:3] == "gpt":
		return "openai"
	case len(modelName) >= 6 && modelName[:6] == "claude":
		return "anthropic"
	case len(modelName) >= 6 && modelName[:6] == "gemini":
		return "google"
	default:
		return "unknown"
	}
}
