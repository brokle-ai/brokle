//go:build !enterprise

package config

import "time"

// EnterpriseConfig contains minimal enterprise configuration for OSS builds
type EnterpriseConfig struct {
	License    LicenseConfig    `mapstructure:"license"`
	SSO        SSOConfig        `mapstructure:"sso"`
	RBAC       RBACConfig       `mapstructure:"rbac"`
	Compliance ComplianceConfig `mapstructure:"compliance"`
	Analytics  AnalyticsConfig  `mapstructure:"analytics"`
	Support    SupportConfig    `mapstructure:"support"`
}

// LicenseConfig contains basic license configuration for OSS builds
type LicenseConfig struct {
	Type        string    `mapstructure:"type"`         // free, pro, business, enterprise
	Key         string    `mapstructure:"key"`          // License key
	ValidUntil  time.Time `mapstructure:"valid_until"`  // License expiration
	MaxRequests int       `mapstructure:"max_requests"` // Monthly request limit
	MaxUsers    int       `mapstructure:"max_users"`    // User limit
	MaxProjects int       `mapstructure:"max_projects"` // Project limit
	Features    []string  `mapstructure:"features"`     // Enabled features
	OfflineMode bool      `mapstructure:"offline_mode"` // Offline license validation
}

// SSOConfig contains minimal SSO configuration for OSS builds
type SSOConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// RBACConfig contains minimal RBAC configuration for OSS builds
type RBACConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// ComplianceConfig contains minimal compliance configuration for OSS builds
type ComplianceConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// AnalyticsConfig contains minimal analytics configuration for OSS builds
type AnalyticsConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// SupportConfig contains minimal support configuration for OSS builds
type SupportConfig struct {
	Level string `mapstructure:"level"` // standard only
}

// Validate methods for OSS builds (minimal validation)
func (ec *EnterpriseConfig) Validate() error { return nil }
func (lc *LicenseConfig) Validate() error    { return nil }
func (sc *SSOConfig) Validate() error        { return nil }
func (rc *RBACConfig) Validate() error       { return nil }
func (cc *ComplianceConfig) Validate() error { return nil }
func (ac *AnalyticsConfig) Validate() error  { return nil }
func (suc *SupportConfig) Validate() error   { return nil }
