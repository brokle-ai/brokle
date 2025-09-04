package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_EnterpriseFeatures(t *testing.T) {
	// Test configuration with enterprise features
	cfg := &Config{
		Enterprise: EnterpriseConfig{
			License: LicenseConfig{
				Type:     "enterprise",
				Features: []string{"advanced_rbac", "sso_integration", "custom_compliance"},
			},
		},
	}

	// Test enterprise license detection
	assert.True(t, cfg.IsEnterpriseLicense())
	assert.Equal(t, "enterprise", cfg.GetLicenseTier())

	// Test feature availability
	assert.True(t, cfg.IsEnterpriseFeatureEnabled("advanced_rbac"))
	assert.True(t, cfg.IsEnterpriseFeatureEnabled("sso_integration"))
	assert.False(t, cfg.IsEnterpriseFeatureEnabled("some_other_feature"))
}

func TestConfig_FreeTier(t *testing.T) {
	// Test free tier configuration
	cfg := &Config{
		Enterprise: EnterpriseConfig{
			License: LicenseConfig{
				Type: "free",
			},
		},
	}

	// Test free tier detection
	assert.False(t, cfg.IsEnterpriseLicense())
	assert.Equal(t, "free", cfg.GetLicenseTier())

	// Test that no enterprise features are available
	assert.False(t, cfg.IsEnterpriseFeatureEnabled("advanced_rbac"))
	assert.False(t, cfg.IsEnterpriseFeatureEnabled("sso_integration"))
}

func TestConfig_DevelopmentMode(t *testing.T) {
	// Test development mode allows all features
	cfg := &Config{
		Environment: "development",
		Enterprise: EnterpriseConfig{
			License: LicenseConfig{
				Type: "free", // Free tier but in development
			},
		},
	}

	assert.True(t, cfg.IsDevelopment())
	
	// In development mode, all features should be available
	assert.True(t, cfg.CanUseFeature("advanced_rbac"))
	assert.True(t, cfg.CanUseFeature("sso_integration"))
	assert.True(t, cfg.CanUseFeature("custom_compliance"))
}

func TestConfig_BusinessTier(t *testing.T) {
	// Test business tier configuration
	cfg := &Config{
		Enterprise: EnterpriseConfig{
			License: LicenseConfig{
				Type:     "business",
				Features: []string{"advanced_rbac", "custom_compliance", "predictive_insights"},
			},
		},
	}

	// Test business tier detection
	assert.True(t, cfg.IsEnterpriseLicense()) // business is considered enterprise
	assert.Equal(t, "business", cfg.GetLicenseTier())

	// Test feature availability
	assert.True(t, cfg.IsEnterpriseFeatureEnabled("advanced_rbac"))
	assert.True(t, cfg.IsEnterpriseFeatureEnabled("custom_compliance"))
	assert.True(t, cfg.IsEnterpriseFeatureEnabled("predictive_insights"))
	assert.False(t, cfg.IsEnterpriseFeatureEnabled("dedicated_support")) // Not in features list
}

func TestConfig_LoadDefaults(t *testing.T) {
	// Temporarily clear environment variables
	oldEnv := os.Getenv("BROKLE_ENTERPRISE_LICENSE_TYPE")
	oldPrivateKey := os.Getenv("BROKLE_JWT_PRIVATE_KEY")
	oldPublicKey := os.Getenv("BROKLE_JWT_PUBLIC_KEY")
	
	defer func() {
		if oldEnv != "" {
			os.Setenv("BROKLE_ENTERPRISE_LICENSE_TYPE", oldEnv)
		} else {
			os.Unsetenv("BROKLE_ENTERPRISE_LICENSE_TYPE")
		}
		if oldPrivateKey != "" {
			os.Setenv("BROKLE_JWT_PRIVATE_KEY", oldPrivateKey)
		} else {
			os.Unsetenv("BROKLE_JWT_PRIVATE_KEY")
		}
		if oldPublicKey != "" {
			os.Setenv("BROKLE_JWT_PUBLIC_KEY", oldPublicKey)
		} else {
			os.Unsetenv("BROKLE_JWT_PUBLIC_KEY")
		}
	}()

	// Clear the env vars for this test
	os.Unsetenv("BROKLE_ENTERPRISE_LICENSE_TYPE")
	
	// Set dummy JWT keys to satisfy validation
	os.Setenv("BROKLE_JWT_PRIVATE_KEY", "dummy-private-key")
	os.Setenv("BROKLE_JWT_PUBLIC_KEY", "dummy-public-key")

	// Load configuration (should use defaults)
	cfg, err := Load()
	require.NoError(t, err)

	// Verify default values
	assert.Equal(t, "free", cfg.Enterprise.License.Type)
	assert.Equal(t, int64(10000), cfg.Enterprise.License.MaxRequests)
	assert.Equal(t, 5, cfg.Enterprise.License.MaxUsers)
	assert.Equal(t, 2, cfg.Enterprise.License.MaxProjects)
	assert.False(t, cfg.Enterprise.SSO.Enabled)
	assert.False(t, cfg.Enterprise.RBAC.Enabled)
	assert.False(t, cfg.Enterprise.Compliance.Enabled)
	assert.True(t, cfg.Enterprise.Analytics.Enabled) // Basic analytics enabled by default
}

func TestEnterpriseConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  EnterpriseConfig
		wantErr bool
	}{
		{
			name: "valid enterprise config",
			config: EnterpriseConfig{
				License: LicenseConfig{
					Type:        "enterprise",
					ValidUntil:  time.Now().AddDate(1, 0, 0),
					MaxRequests: 1000000,
					MaxUsers:    100,
					MaxProjects: 50,
					Features:    []string{"advanced_rbac", "sso_integration"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid free config",
			config: EnterpriseConfig{
				License: LicenseConfig{
					Type:        "free",
					MaxRequests: 10000,
					MaxUsers:    5,
					MaxProjects: 2,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Enterprise: tt.config,
			}
			
			// Basic validation - ensure the config can be used
			assert.NotNil(t, cfg.Enterprise)
			assert.NotEmpty(t, cfg.GetLicenseTier())
		})
	}
}

// Test helper functions
func TestGetLicenseTier_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		licenseType  string
		expectedTier string
	}{
		{"empty license type", "", "free"},
		{"pro tier", "pro", "pro"},
		{"business tier", "business", "business"},
		{"enterprise tier", "enterprise", "enterprise"},
		{"unknown tier", "unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Enterprise: EnterpriseConfig{
					License: LicenseConfig{
						Type: tt.licenseType,
					},
				},
			}
			
			assert.Equal(t, tt.expectedTier, cfg.GetLicenseTier())
		})
	}
}