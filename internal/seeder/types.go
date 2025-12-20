package seeder

import (
	"encoding/json"

	"brokle/pkg/ulid"
)

type PermissionsFile struct {
	Permissions []PermissionSeed `yaml:"permissions"`
}

type RolesFile struct {
	Roles []RoleSeed `yaml:"roles"`
}

type SeedData struct {
	Permissions []PermissionSeed
	Roles       []RoleSeed
}

type PermissionSeed struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type RoleSeed struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	ScopeType   string   `yaml:"scope_type"` // 'organization' | 'project'
	Permissions []string `yaml:"permissions"`
}

type Options struct {
	Reset   bool
	DryRun  bool
	Verbose bool
}

type EntityMaps struct {
	Permissions map[string]ulid.ULID // permission name -> permission ID
	Roles       map[string]ulid.ULID // role_name -> role ID
}

func NewEntityMaps() *EntityMaps {
	return &EntityMaps{
		Permissions: make(map[string]ulid.ULID),
		Roles:       make(map[string]ulid.ULID),
	}
}

type ProviderPricingSeedData struct {
	Version        string              `yaml:"version"`
	ProviderModels []ProviderModelSeed `yaml:"provider_models"`
}

type ProviderModelSeed struct {
	ModelName       string                 `yaml:"model_name"`
	Provider        string                 `yaml:"provider"`                  // "openai", "anthropic", "gemini"
	DisplayName     string                 `yaml:"display_name,omitempty"`    // User-friendly name for UI
	MatchPattern    string                 `yaml:"match_pattern"`
	StartDate       string                 `yaml:"start_date"` // Format: "2024-05-13"
	Unit            string                 `yaml:"unit"`       // Default: "TOKENS"
	TokenizerID     string                 `yaml:"tokenizer_id,omitempty"`
	TokenizerConfig map[string]interface{} `yaml:"tokenizer_config,omitempty"`
	Prices          []PriceSeed            `yaml:"prices"`
}

type PriceSeed struct {
	UsageType string  `yaml:"usage_type"` // "input", "output", "cache_read_input_tokens", etc.
	Price     float64 `yaml:"price"`      // Price per 1M tokens
}

type RBACStatistics struct {
	TotalRoles        int            `json:"total_roles"`
	TotalPermissions  int            `json:"total_permissions"`
	ScopeDistribution map[string]int `json:"scope_distribution"`
	RoleDistribution  map[string]int `json:"role_distribution"`
}

func (s *RBACStatistics) String() string {
	data, _ := json.MarshalIndent(s, "", "  ")
	return string(data)
}

type PricingStatistics struct {
	TotalModels          int            `json:"total_models"`
	TotalPrices          int            `json:"total_prices"`
	ProviderDistribution map[string]int `json:"provider_distribution"`
}

func (s *PricingStatistics) String() string {
	data, _ := json.MarshalIndent(s, "", "  ")
	return string(data)
}

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
