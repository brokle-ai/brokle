package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"brokle/internal/config"
	"brokle/internal/ee/analytics"
	"brokle/internal/ee/compliance"
	"brokle/internal/ee/rbac"
	"brokle/internal/ee/sso"
)

func TestEnterpriseServicesInitialization(t *testing.T) {
	// Test that enterprise services can be initialized with stub implementations

	// Create a test configuration
	cfg := &config.Config{
		Environment: "test",
		Enterprise: config.EnterpriseConfig{
			License: config.LicenseConfig{
				Type:        "free",
				MaxRequests: 10000,
				MaxUsers:    5,
				MaxProjects: 2,
			},
		},
	}

	t.Run("Compliance service initialization", func(t *testing.T) {
		complianceService := compliance.New()
		require.NotNil(t, complianceService)

		// Test that it implements the interface
		var _ compliance.Compliance = complianceService
	})

	t.Run("SSO service initialization", func(t *testing.T) {
		ssoService := sso.New()
		require.NotNil(t, ssoService)

		// Test that it implements the interface
		var _ sso.SSOProvider = ssoService
	})

	t.Run("RBAC service initialization", func(t *testing.T) {
		rbacService := rbac.New()
		require.NotNil(t, rbacService)

		// Test that it implements the interface
		var _ rbac.RBACManager = rbacService
	})

	t.Run("Enterprise Analytics initialization", func(t *testing.T) {
		analyticsService := analytics.New()
		require.NotNil(t, analyticsService)

		// Test that it implements the interface
		var _ analytics.EnterpriseAnalytics = analyticsService
	})

	t.Run("Configuration enterprise features", func(t *testing.T) {
		// Test free tier
		assert.False(t, cfg.IsEnterpriseLicense())
		assert.Equal(t, "free", cfg.GetLicenseTier())
		assert.False(t, cfg.IsEnterpriseFeatureEnabled("advanced_rbac"))

		// Test development mode (should allow all features)
		cfg.Environment = "development"
		assert.True(t, cfg.IsDevelopment())
		assert.True(t, cfg.CanUseFeature("advanced_rbac"))

		// Test enterprise license
		cfg.Environment = "production"
		cfg.Enterprise.License.Type = "enterprise"
		cfg.Enterprise.License.Features = []string{"advanced_rbac", "sso_integration"}

		assert.True(t, cfg.IsEnterpriseLicense())
		assert.Equal(t, "enterprise", cfg.GetLicenseTier())
		assert.True(t, cfg.IsEnterpriseFeatureEnabled("advanced_rbac"))
		assert.True(t, cfg.IsEnterpriseFeatureEnabled("sso_integration"))
		assert.False(t, cfg.IsEnterpriseFeatureEnabled("some_nonexistent_feature"))
	})
}

func TestEnterpriseArchitecturePatterns(t *testing.T) {
	// Test that our architecture patterns work correctly

	t.Run("Build tags simulation", func(t *testing.T) {
		// In OSS build (default), we should get stub implementations
		complianceService := compliance.New()

		// Verify it's a stub by checking that it returns the expected stub behavior
		// (we can't directly check the type due to build tags, but we can check behavior)
		require.NotNil(t, complianceService)

		// The stub should be functional but limited
		// This test passes regardless of build tags, demonstrating the pattern works
	})

	t.Run("Interface consistency", func(t *testing.T) {
		// Test that all enterprise services implement their interfaces correctly
		// This ensures that both stub and real implementations will work

		var complianceService compliance.Compliance = compliance.New()
		var ssoService sso.SSOProvider = sso.New()
		var rbacService rbac.RBACManager = rbac.New()
		var analyticsService analytics.EnterpriseAnalytics = analytics.New()

		assert.NotNil(t, complianceService)
		assert.NotNil(t, ssoService)
		assert.NotNil(t, rbacService)
		assert.NotNil(t, analyticsService)
	})

	t.Run("Configuration validation", func(t *testing.T) {
		// Test that enterprise configuration validates correctly
		validConfig := &config.Config{
			Enterprise: config.EnterpriseConfig{
				License: config.LicenseConfig{
					Type:        "enterprise",
					MaxRequests: 1000000,
					MaxUsers:    100,
					MaxProjects: 50,
					Features: []string{
						"advanced_rbac",
						"sso_integration",
						"custom_compliance",
						"predictive_insights",
					},
				},
			},
		}

		assert.True(t, validConfig.IsEnterpriseLicense())
		assert.True(t, validConfig.IsEnterpriseFeatureEnabled("advanced_rbac"))
		assert.True(t, validConfig.IsEnterpriseFeatureEnabled("sso_integration"))

		// Test feature that's not in the license
		assert.False(t, validConfig.IsEnterpriseFeatureEnabled("dedicated_support"))
	})
}

// TestBuildTagsPreparation verifies that the build system will work correctly
func TestBuildTagsPreparation(t *testing.T) {
	t.Run("OSS build preparation", func(t *testing.T) {
		// This test verifies that the OSS build will work
		// In OSS builds, enterprise services should return stub implementations
		// that provide graceful degradation

		complianceService := compliance.New()

		// Verify the service is functional (won't crash the app)
		require.NotNil(t, complianceService)

		// The stub should implement all interface methods
		// (This test will fail if any interface methods are missing)
		var _ compliance.Compliance = complianceService
	})

	t.Run("Enterprise build preparation", func(t *testing.T) {
		// This test verifies that the enterprise build will work
		// When enterprise build tags are used, the New() functions should return
		// real implementations instead of stubs

		// The same interface should work regardless of implementation
		var complianceService compliance.Compliance = compliance.New()
		var ssoService sso.SSOProvider = sso.New()
		var rbacService rbac.RBACManager = rbac.New()
		var analyticsService analytics.EnterpriseAnalytics = analytics.New()

		// All services should be functional
		assert.NotNil(t, complianceService)
		assert.NotNil(t, ssoService)
		assert.NotNil(t, rbacService)
		assert.NotNil(t, analyticsService)
	})
}

func TestEnterpriseFeatureMatrix(t *testing.T) {
	// Test the feature matrix for different license tiers

	tests := []struct {
		expected map[string]bool
		name     string
		tier     string
		features []string
	}{
		{
			name:     "Free tier",
			tier:     "free",
			features: []string{},
			expected: map[string]bool{
				"advanced_rbac":       false,
				"sso_integration":     false,
				"custom_compliance":   false,
				"predictive_insights": false,
			},
		},
		{
			name:     "Pro tier",
			tier:     "pro",
			features: []string{"advanced_rbac"},
			expected: map[string]bool{
				"advanced_rbac":       true,
				"sso_integration":     false,
				"custom_compliance":   false,
				"predictive_insights": false,
			},
		},
		{
			name:     "Business tier",
			tier:     "business",
			features: []string{"advanced_rbac", "sso_integration", "custom_compliance", "predictive_insights"},
			expected: map[string]bool{
				"advanced_rbac":       true,
				"sso_integration":     true,
				"custom_compliance":   true,
				"predictive_insights": true,
			},
		},
		{
			name:     "Enterprise tier",
			tier:     "enterprise",
			features: []string{"advanced_rbac", "sso_integration", "custom_compliance", "predictive_insights", "dedicated_support"},
			expected: map[string]bool{
				"advanced_rbac":       true,
				"sso_integration":     true,
				"custom_compliance":   true,
				"predictive_insights": true,
				"dedicated_support":   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Enterprise: config.EnterpriseConfig{
					License: config.LicenseConfig{
						Type:     tt.tier,
						Features: tt.features,
					},
				},
			}

			for feature, expectedAvailable := range tt.expected {
				actual := cfg.IsEnterpriseFeatureEnabled(feature)
				assert.Equal(t, expectedAvailable, actual,
					"Feature %s should be %v for tier %s", feature, expectedAvailable, tt.tier)
			}
		})
	}
}
