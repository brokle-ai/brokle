package config

import (
	"fmt"
	"time"
)

// LicenseWrapper provides enhanced license management with validation and defaults
type LicenseWrapper struct {
	config *Config
}

// NewLicenseWrapper creates a new license wrapper
func NewLicenseWrapper(config *Config) *LicenseWrapper {
	return &LicenseWrapper{config: config}
}

// GetEffectiveLicense returns the effective license configuration with all defaults applied
func (lw *LicenseWrapper) GetEffectiveLicense() *LicenseConfig {
	license := &lw.config.Enterprise.License

	// Apply tier-based defaults if not explicitly set
	if license.Type == "" {
		license.Type = "free"
	}

	// Ensure license has proper limits based on tier
	lw.applyTierDefaults(license)

	// Ensure license has valid expiration
	if license.ValidUntil.IsZero() {
		license.ValidUntil = lw.getDefaultValidUntil(license.Type)
	}

	return license
}

// ValidateLicense performs comprehensive license validation
func (lw *LicenseWrapper) ValidateLicense() error {
	license := lw.GetEffectiveLicense()

	// Validate tier
	validTiers := []string{"free", "pro", "business", "enterprise"}
	if !lw.isValidTier(license.Type, validTiers) {
		return fmt.Errorf("invalid license tier: %s. Valid tiers: %v", license.Type, validTiers)
	}

	// Validate limits are reasonable
	if err := lw.validateLimits(license); err != nil {
		return fmt.Errorf("invalid license limits: %w", err)
	}

	// Validate expiration
	if license.ValidUntil.Before(time.Now()) && license.Type != "free" {
		return fmt.Errorf("license expired on %s", license.ValidUntil.Format("2006-01-02"))
	}

	// Validate features exist for the tier
	if err := lw.validateFeaturesForTier(license); err != nil {
		return fmt.Errorf("invalid features for tier: %w", err)
	}

	return nil
}

// GetTierLimits returns the standard limits for a given tier
func (lw *LicenseWrapper) GetTierLimits(tier string) (*LicenseConfig, error) {
	limits := &LicenseConfig{
		Type:        tier,
		ValidUntil:  lw.getDefaultValidUntil(tier),
		OfflineMode: false,
	}

	switch tier {
	case "free":
		limits.MaxRequests = 10000   // 10K requests/month
		limits.MaxUsers = 5          // 5 users
		limits.MaxProjects = 2       // 2 projects
		limits.Features = []string{} // No enterprise features

	case "pro":
		limits.MaxRequests = 100000 // 100K requests/month
		limits.MaxUsers = 10        // 10 users
		limits.MaxProjects = 10     // 10 projects
		limits.Features = []string{
			"advanced_rbac",
		}

	case "business":
		limits.MaxRequests = 1000000 // 1M requests/month
		limits.MaxUsers = 50         // 50 users
		limits.MaxProjects = 100     // 100 projects
		limits.Features = []string{
			"advanced_rbac",
			"sso_integration",
			"custom_compliance",
			"predictive_insights",
			"custom_dashboards",
		}

	case "enterprise":
		limits.MaxRequests = 10000000 // 10M requests/month (effectively unlimited)
		limits.MaxUsers = 1000        // 1000 users
		limits.MaxProjects = 1000     // 1000 projects
		limits.Features = []string{
			"advanced_rbac",
			"sso_integration",
			"custom_compliance",
			"predictive_insights",
			"custom_dashboards",
			"on_premise_deployment",
			"dedicated_support",
			"advanced_integrations",
			"cross_org_analytics",
		}

	default:
		return nil, fmt.Errorf("unknown tier: %s", tier)
	}

	return limits, nil
}

// IsFeatureAvailableInTier checks if a feature is available in the given tier
func (lw *LicenseWrapper) IsFeatureAvailableInTier(feature, tier string) bool {
	tierLimits, err := lw.GetTierLimits(tier)
	if err != nil {
		return false
	}

	for _, f := range tierLimits.Features {
		if f == feature {
			return true
		}
	}
	return false
}

// GetRecommendedTierForFeature returns the minimum tier that supports the feature
func (lw *LicenseWrapper) GetRecommendedTierForFeature(feature string) string {
	tiers := []string{"free", "pro", "business", "enterprise"}

	for _, tier := range tiers {
		if lw.IsFeatureAvailableInTier(feature, tier) {
			return tier
		}
	}

	return "enterprise" // Default to enterprise for unknown features
}

// GetUpgradePath returns the recommended upgrade path from current to target tier
func (lw *LicenseWrapper) GetUpgradePath(currentTier, targetTier string) []string {
	tierOrder := map[string]int{
		"free":       0,
		"pro":        1,
		"business":   2,
		"enterprise": 3,
	}

	currentOrder, currentExists := tierOrder[currentTier]
	targetOrder, targetExists := tierOrder[targetTier]

	if !currentExists || !targetExists || currentOrder >= targetOrder {
		return []string{targetTier}
	}

	path := []string{}
	for tier, order := range tierOrder {
		if order > currentOrder && order <= targetOrder {
			path = append(path, tier)
		}
	}

	return path
}

// GetLicenseStatus returns a human-readable status of the current license
func (lw *LicenseWrapper) GetLicenseStatus() string {
	license := lw.GetEffectiveLicense()

	if license.Type == "free" {
		return "Free tier - upgrade to unlock enterprise features"
	}

	if license.ValidUntil.Before(time.Now()) {
		return fmt.Sprintf("%s license expired on %s",
			license.Type, license.ValidUntil.Format("January 2, 2006"))
	}

	daysUntilExpiry := int(time.Until(license.ValidUntil).Hours() / 24)
	if daysUntilExpiry <= 30 && license.Type != "free" {
		return fmt.Sprintf("%s license expires in %d days", license.Type, daysUntilExpiry)
	}

	return license.Type + " license active"
}

// Private helper methods

func (lw *LicenseWrapper) applyTierDefaults(license *LicenseConfig) {
	if license.MaxRequests == 0 || license.MaxUsers == 0 || license.MaxProjects == 0 {
		defaults, err := lw.GetTierLimits(license.Type)
		if err == nil {
			if license.MaxRequests == 0 {
				license.MaxRequests = defaults.MaxRequests
			}
			if license.MaxUsers == 0 {
				license.MaxUsers = defaults.MaxUsers
			}
			if license.MaxProjects == 0 {
				license.MaxProjects = defaults.MaxProjects
			}
			if len(license.Features) == 0 {
				license.Features = defaults.Features
			}
		}
	}
}

func (lw *LicenseWrapper) getDefaultValidUntil(tier string) time.Time {
	switch tier {
	case "free":
		return time.Now().AddDate(10, 0, 0) // Free tier valid for 10 years
	default:
		return time.Now().AddDate(1, 0, 0) // Paid tiers default to 1 year
	}
}

func (lw *LicenseWrapper) isValidTier(tier string, validTiers []string) bool {
	for _, valid := range validTiers {
		if tier == valid {
			return true
		}
	}
	return false
}

func (lw *LicenseWrapper) validateLimits(license *LicenseConfig) error {
	if license.MaxRequests <= 0 {
		return fmt.Errorf("max_requests must be positive, got %d", license.MaxRequests)
	}
	if license.MaxUsers <= 0 {
		return fmt.Errorf("max_users must be positive, got %d", license.MaxUsers)
	}
	if license.MaxProjects <= 0 {
		return fmt.Errorf("max_projects must be positive, got %d", license.MaxProjects)
	}

	// Validate limits are reasonable for the tier
	expectedLimits, err := lw.GetTierLimits(license.Type)
	if err == nil {
		if license.MaxRequests > expectedLimits.MaxRequests*10 {
			return fmt.Errorf("max_requests %d exceeds reasonable limit for %s tier",
				license.MaxRequests, license.Type)
		}
	}

	return nil
}

func (lw *LicenseWrapper) validateFeaturesForTier(license *LicenseConfig) error {
	expectedLimits, err := lw.GetTierLimits(license.Type)
	if err != nil {
		return err
	}

	// Check if license has features that shouldn't be available in this tier
	for _, feature := range license.Features {
		found := false
		for _, expectedFeature := range expectedLimits.Features {
			if feature == expectedFeature {
				found = true
				break
			}
		}
		if !found && license.Type != "enterprise" {
			return fmt.Errorf("feature '%s' not available in %s tier", feature, license.Type)
		}
	}

	return nil
}
