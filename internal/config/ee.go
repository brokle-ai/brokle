//go:build enterprise

package config

import (
	"errors"
	"fmt"
	"time"
)

// EnterpriseConfig contains enterprise-only configuration
// This struct is always present but features are license-gated
type EnterpriseConfig struct {
	SSO        SSOConfig        `mapstructure:"sso"`
	RBAC       RBACConfig       `mapstructure:"rbac"`
	Support    SupportConfig    `mapstructure:"support"`
	Analytics  AnalyticsConfig  `mapstructure:"analytics"`
	License    LicenseConfig    `mapstructure:"license"`
	Compliance ComplianceConfig `mapstructure:"compliance"`
}

// LicenseConfig handles license validation and entitlements
type LicenseConfig struct {
	ValidUntil  time.Time `mapstructure:"valid_until"`
	Key         string    `mapstructure:"key"`
	Type        string    `mapstructure:"type"`
	Features    []string  `mapstructure:"features"`
	MaxRequests int       `mapstructure:"max_requests"`
	MaxUsers    int       `mapstructure:"max_users"`
	MaxProjects int       `mapstructure:"max_projects"`
	OfflineMode bool      `mapstructure:"offline_mode"`
}

// SSOConfig for enterprise authentication
type SSOConfig struct {
	Attributes  map[string]string `mapstructure:"attributes"`
	Provider    string            `mapstructure:"provider"`
	MetadataURL string            `mapstructure:"metadata_url"`
	EntityID    string            `mapstructure:"entity_id"`
	Certificate string            `mapstructure:"certificate"`
	Enabled     bool              `mapstructure:"enabled"`
}

// RBACConfig for advanced role-based access control
type RBACConfig struct {
	Permissions map[string][]string `mapstructure:"permissions"`
	Inheritance map[string][]string `mapstructure:"inheritance"`
	CustomRoles []CustomRole        `mapstructure:"custom_roles"`
	Enabled     bool                `mapstructure:"enabled"`
}

// CustomRole represents a custom RBAC role
type CustomRole struct {
	Name        string   `mapstructure:"name"`
	Permissions []string `mapstructure:"permissions"`
	Scopes      []string `mapstructure:"scopes"` // org, project, environment
}

// ComplianceConfig for enterprise compliance features
type ComplianceConfig struct {
	AuditRetention   time.Duration `mapstructure:"audit_retention"`
	DataRetention    time.Duration `mapstructure:"data_retention"`
	Enabled          bool          `mapstructure:"enabled"`
	PIIAnonymization bool          `mapstructure:"pii_anonymization"`
	SOC2Compliance   bool          `mapstructure:"soc2_compliance"`
	HIPAACompliance  bool          `mapstructure:"hipaa_compliance"`
	GDPRCompliance   bool          `mapstructure:"gdpr_compliance"`
}

// AnalyticsConfig for enterprise analytics features
type AnalyticsConfig struct {
	ExportFormats      []string `mapstructure:"export_formats"`
	Enabled            bool     `mapstructure:"enabled"`
	PredictiveInsights bool     `mapstructure:"predictive_insights"`
	CustomDashboards   bool     `mapstructure:"custom_dashboards"`
	MLModels           bool     `mapstructure:"ml_models"`
}

// SupportConfig for enterprise support features
type SupportConfig struct {
	Level            string `mapstructure:"level"` // standard, priority, dedicated
	SLA              string `mapstructure:"sla"`   // 99.9%, 99.95%, 99.99%
	DedicatedManager bool   `mapstructure:"dedicated_manager"`
	OnCallSupport    bool   `mapstructure:"on_call_support"`
}

// Validation methods for enterprise builds

// Validate validates enterprise configuration.
func (ec *EnterpriseConfig) Validate() error {
	if err := ec.License.Validate(); err != nil {
		return fmt.Errorf("license config: %w", err)
	}

	if err := ec.SSO.Validate(); err != nil {
		return fmt.Errorf("sso config: %w", err)
	}

	if err := ec.RBAC.Validate(); err != nil {
		return fmt.Errorf("rbac config: %w", err)
	}

	if err := ec.Compliance.Validate(); err != nil {
		return fmt.Errorf("compliance config: %w", err)
	}

	if err := ec.Analytics.Validate(); err != nil {
		return fmt.Errorf("analytics config: %w", err)
	}

	if err := ec.Support.Validate(); err != nil {
		return fmt.Errorf("support config: %w", err)
	}

	return nil
}

// Validate validates license configuration.
func (lc *LicenseConfig) Validate() error {
	validTypes := []string{"free", "pro", "business", "enterprise"}
	isValid := false
	for _, t := range validTypes {
		if lc.Type == t {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid license type: %s (must be one of %v)", lc.Type, validTypes)
	}

	if lc.MaxRequests <= 0 {
		return errors.New("max_requests must be positive")
	}

	if lc.MaxUsers <= 0 {
		return errors.New("max_users must be positive")
	}

	if lc.MaxProjects <= 0 {
		return errors.New("max_projects must be positive")
	}

	// Validate license expiration for non-free licenses
	if lc.Type != "free" && !lc.ValidUntil.IsZero() && lc.ValidUntil.Before(time.Now()) {
		return fmt.Errorf("license expired on %s", lc.ValidUntil.Format("2006-01-02"))
	}

	return nil
}

// Validate validates SSO configuration.
func (sc *SSOConfig) Validate() error {
	if sc.Enabled {
		validProviders := []string{"saml", "oidc", "oauth2"}
		isValid := false
		for _, provider := range validProviders {
			if sc.Provider == provider {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid sso provider: %s (must be one of %v)", sc.Provider, validProviders)
		}

		// Provider-specific validation
		switch sc.Provider {
		case "saml":
			if sc.EntityID == "" {
				return errors.New("entity_id is required for SAML")
			}
			if sc.MetadataURL == "" && sc.Certificate == "" {
				return errors.New("either metadata_url or certificate is required for SAML")
			}
		case "oidc", "oauth2":
			if sc.MetadataURL == "" {
				return errors.New("metadata_url is required for OIDC/OAuth2")
			}
		}
	}

	return nil
}

// Validate validates RBAC configuration.
func (rc *RBACConfig) Validate() error {
	// Validate custom roles if defined
	for _, role := range rc.CustomRoles {
		if role.Name == "" {
			return errors.New("custom role name cannot be empty")
		}
		if len(role.Permissions) == 0 {
			return fmt.Errorf("custom role '%s' must have at least one permission", role.Name)
		}
	}

	return nil
}

// Validate validates compliance configuration.
func (cc *ComplianceConfig) Validate() error {
	if cc.Enabled {
		if cc.AuditRetention <= 0 {
			return errors.New("audit_retention must be positive when compliance is enabled")
		}

		if cc.DataRetention <= 0 {
			return errors.New("data_retention must be positive when compliance is enabled")
		}
	}

	return nil
}

// Validate validates analytics configuration.
func (ac *AnalyticsConfig) Validate() error {
	// Validate export formats if specified
	validFormats := []string{"csv", "json", "parquet", "excel"}
	for _, format := range ac.ExportFormats {
		isValid := false
		for _, valid := range validFormats {
			if format == valid {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid export format: %s (must be one of %v)", format, validFormats)
		}
	}

	return nil
}

// Validate validates support configuration.
func (sc *SupportConfig) Validate() error {
	validLevels := []string{"standard", "priority", "dedicated"}
	isValid := false
	for _, level := range validLevels {
		if sc.Level == level {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid support level: %s (must be one of %v)", sc.Level, validLevels)
	}

	return nil
}
