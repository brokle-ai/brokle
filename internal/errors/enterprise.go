package errors

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// EnterpriseErrorCode represents different types of enterprise errors
type EnterpriseErrorCode string

const (
	ErrorCodeFeatureNotAvailable EnterpriseErrorCode = "FEATURE_NOT_AVAILABLE"
	ErrorCodeLicenseRequired     EnterpriseErrorCode = "LICENSE_REQUIRED"
	ErrorCodeUsageLimitExceeded  EnterpriseErrorCode = "USAGE_LIMIT_EXCEEDED"
	ErrorCodeLicenseExpired      EnterpriseErrorCode = "LICENSE_EXPIRED"
	ErrorCodeLicenseInvalid      EnterpriseErrorCode = "LICENSE_INVALID"
	ErrorCodeTierUpgradeRequired EnterpriseErrorCode = "TIER_UPGRADE_REQUIRED"
)

// EnterpriseError represents a professional enterprise error response
type EnterpriseError struct {
	Code           EnterpriseErrorCode `json:"code"`
	Message        string              `json:"message"`
	Feature        string              `json:"feature,omitempty"`
	CurrentTier    string              `json:"current_tier,omitempty"`
	RequiredTier   string              `json:"required_tier,omitempty"`
	LimitType      string              `json:"limit_type,omitempty"`
	RemainingQuota int64               `json:"remaining_quota,omitempty"`
	ResetDate      *time.Time          `json:"reset_date,omitempty"`
	Actions        []ActionSuggestion  `json:"actions"`
	Support        SupportInfo         `json:"support"`
	Metadata       map[string]string   `json:"metadata,omitempty"`
}

// ActionSuggestion represents a suggested action for the user
type ActionSuggestion struct {
	Type        string `json:"type"` // "upgrade", "contact_sales", "trial", "documentation"
	Label       string `json:"label"`
	URL         string `json:"url"`
	UTMSource   string `json:"utm_source"`
	UTMCampaign string `json:"utm_campaign"`
	UTMContent  string `json:"utm_content"`
	Primary     bool   `json:"primary"`
}

// SupportInfo provides support contact information
type SupportInfo struct {
	Email        string `json:"email"`
	Phone        string `json:"phone,omitempty"`
	ChatURL      string `json:"chat_url,omitempty"`
	DocsURL      string `json:"docs_url"`
	CommunityURL string `json:"community_url,omitempty"`
}

// HTTPStatus returns the appropriate HTTP status code for the enterprise error
func (ee *EnterpriseError) HTTPStatus() int {
	switch ee.Code {
	case ErrorCodeFeatureNotAvailable, ErrorCodeLicenseRequired, ErrorCodeTierUpgradeRequired:
		return http.StatusPaymentRequired // 402
	case ErrorCodeUsageLimitExceeded:
		return http.StatusTooManyRequests // 429
	case ErrorCodeLicenseExpired, ErrorCodeLicenseInvalid:
		return http.StatusUnauthorized // 401
	default:
		return http.StatusForbidden // 403
	}
}

// NewFeatureNotAvailableError creates a feature not available error
func NewFeatureNotAvailableError(feature, currentTier, requiredTier string) *EnterpriseError {
	featureName := formatFeatureName(feature)

	return &EnterpriseError{
		Code:         ErrorCodeFeatureNotAvailable,
		Message:      fmt.Sprintf("%s requires %s tier or higher. You're currently on %s tier.", featureName, strings.Title(requiredTier), strings.Title(currentTier)),
		Feature:      feature,
		CurrentTier:  currentTier,
		RequiredTier: requiredTier,
		Actions:      buildUpgradeActions(currentTier, requiredTier, feature),
		Support:      buildSupportInfo(),
		Metadata: map[string]string{
			"feature_category": getFeatureCategory(feature),
			"pricing_tier":     requiredTier,
		},
	}
}

// NewUsageLimitExceededError creates a usage limit exceeded error
func NewUsageLimitExceededError(limitType, currentTier string, remaining int64, resetDate *time.Time) *EnterpriseError {
	return &EnterpriseError{
		Code:           ErrorCodeUsageLimitExceeded,
		Message:        buildUsageLimitMessage(limitType, currentTier),
		LimitType:      limitType,
		CurrentTier:    currentTier,
		RemainingQuota: remaining,
		ResetDate:      resetDate,
		Actions:        buildUsageActions(currentTier, limitType),
		Support:        buildSupportInfo(),
		Metadata: map[string]string{
			"limit_category": limitType,
			"tier":           currentTier,
		},
	}
}

// NewLicenseRequiredError creates a license required error
func NewLicenseRequiredError(currentTier string) *EnterpriseError {
	return &EnterpriseError{
		Code:        ErrorCodeLicenseRequired,
		Message:     "This endpoint requires a valid enterprise license. Please upgrade your plan or contact sales.",
		CurrentTier: currentTier,
		Actions:     buildEnterpriseActions(currentTier),
		Support:     buildSupportInfo(),
		Metadata: map[string]string{
			"access_type":  "enterprise_only",
			"current_tier": currentTier,
		},
	}
}

// Helper functions for building professional error responses

func buildUpgradeActions(currentTier, requiredTier, feature string) []ActionSuggestion {
	actions := []ActionSuggestion{}

	// Primary upgrade action
	upgradeURL := buildUpgradeURL(requiredTier, feature)
	actions = append(actions, ActionSuggestion{
		Type:        "upgrade",
		Label:       fmt.Sprintf("Upgrade to %s", strings.Title(requiredTier)),
		URL:         upgradeURL,
		UTMSource:   "api",
		UTMCampaign: "feature_upgrade",
		UTMContent:  feature,
		Primary:     true,
	})

	// Trial action for free tier users
	if currentTier == "free" {
		actions = append(actions, ActionSuggestion{
			Type:        "trial",
			Label:       "Start Free Trial",
			URL:         buildTrialURL(feature),
			UTMSource:   "api",
			UTMCampaign: "feature_trial",
			UTMContent:  feature,
			Primary:     false,
		})
	}

	// Contact sales for enterprise features
	if requiredTier == "enterprise" {
		actions = append(actions, ActionSuggestion{
			Type:        "contact_sales",
			Label:       "Contact Sales",
			URL:         buildSalesURL(feature),
			UTMSource:   "api",
			UTMCampaign: "enterprise_inquiry",
			UTMContent:  feature,
			Primary:     false,
		})
	}

	// Documentation
	actions = append(actions, ActionSuggestion{
		Type:        "documentation",
		Label:       "Learn More",
		URL:         buildFeatureDocsURL(feature),
		UTMSource:   "api",
		UTMCampaign: "feature_docs",
		UTMContent:  feature,
		Primary:     false,
	})

	return actions
}

func buildUsageActions(currentTier, limitType string) []ActionSuggestion {
	actions := []ActionSuggestion{
		{
			Type:        "upgrade",
			Label:       "Upgrade Plan",
			URL:         buildUsageUpgradeURL(limitType),
			UTMSource:   "api",
			UTMCampaign: "usage_upgrade",
			UTMContent:  limitType,
			Primary:     true,
		},
		{
			Type:        "documentation",
			Label:       "View Usage Guidelines",
			URL:         "https://docs.brokle.com/usage-limits",
			UTMSource:   "api",
			UTMCampaign: "usage_docs",
			UTMContent:  limitType,
			Primary:     false,
		},
	}

	return actions
}

func buildEnterpriseActions(currentTier string) []ActionSuggestion {
	return []ActionSuggestion{
		{
			Type:        "contact_sales",
			Label:       "Contact Sales",
			URL:         buildSalesURL("enterprise_access"),
			UTMSource:   "api",
			UTMCampaign: "enterprise_access",
			UTMContent:  "license_required",
			Primary:     true,
		},
		{
			Type:        "upgrade",
			Label:       "View Enterprise Plans",
			URL:         "https://brokle.com/pricing?tier=enterprise&utm_source=api&utm_campaign=enterprise_access&utm_content=license_required",
			UTMSource:   "api",
			UTMCampaign: "enterprise_access",
			UTMContent:  "license_required",
			Primary:     false,
		},
	}
}

func buildSupportInfo() SupportInfo {
	return SupportInfo{
		Email:        "support@brokle.comm",
		ChatURL:      "https://brokle.com/chat",
		DocsURL:      "https://docs.brokle.com",
		CommunityURL: "https://community.brokle.com",
	}
}

// URL builders with proper UTM tracking

func buildUpgradeURL(tier, feature string) string {
	return fmt.Sprintf("https://brokle.com/pricing?tier=%s&utm_source=api&utm_campaign=feature_upgrade&utm_content=%s", tier, feature)
}

func buildTrialURL(feature string) string {
	return fmt.Sprintf("https://brokle.com/trial?utm_source=api&utm_campaign=feature_trial&utm_content=%s", feature)
}

func buildSalesURL(feature string) string {
	return fmt.Sprintf("https://brokle.com/contact?utm_source=api&utm_campaign=enterprise_inquiry&utm_content=%s", feature)
}

func buildUsageUpgradeURL(limitType string) string {
	return fmt.Sprintf("https://brokle.com/pricing?utm_source=api&utm_campaign=usage_upgrade&utm_content=%s", limitType)
}

func buildFeatureDocsURL(feature string) string {
	return fmt.Sprintf("https://docs.brokle.com/features/%s?utm_source=api&utm_campaign=feature_docs&utm_content=%s", feature, feature)
}

// Feature categorization and formatting

func formatFeatureName(feature string) string {
	featureNames := map[string]string{
		"advanced_rbac":         "Advanced Role-Based Access Control",
		"sso_integration":       "Single Sign-On Integration",
		"custom_compliance":     "Custom Compliance Controls",
		"predictive_insights":   "Predictive Analytics & Insights",
		"custom_dashboards":     "Custom Dashboard Builder",
		"on_premise_deployment": "On-Premise Deployment",
		"dedicated_support":     "Dedicated Support & Success Manager",
		"advanced_integrations": "Advanced Enterprise Integrations",
		"cross_org_analytics":   "Cross-Organization Analytics",
	}

	if name, exists := featureNames[feature]; exists {
		return name
	}

	// Format unknown features nicely
	return strings.Title(strings.ReplaceAll(feature, "_", " "))
}

func getFeatureCategory(feature string) string {
	categories := map[string]string{
		"advanced_rbac":         "security",
		"sso_integration":       "authentication",
		"custom_compliance":     "compliance",
		"predictive_insights":   "analytics",
		"custom_dashboards":     "visualization",
		"on_premise_deployment": "deployment",
		"dedicated_support":     "support",
		"advanced_integrations": "integrations",
		"cross_org_analytics":   "analytics",
	}

	if category, exists := categories[feature]; exists {
		return category
	}
	return "feature"
}

func buildUsageLimitMessage(limitType, tier string) string {
	messages := map[string]string{
		"requests": fmt.Sprintf("Monthly API request limit reached for %s tier. Upgrade your plan to increase your quota and continue accessing the Brokle AI platform.", strings.Title(tier)),
		"users":    fmt.Sprintf("User limit reached for %s tier. Upgrade your plan to add more team members to your organization.", strings.Title(tier)),
		"projects": fmt.Sprintf("Project limit reached for %s tier. Upgrade your plan to create additional projects in your organization.", strings.Title(tier)),
	}

	if msg, exists := messages[limitType]; exists {
		return msg
	}
	return fmt.Sprintf("Usage limit exceeded for %s tier. Please upgrade your plan to continue using Brokle AI platform.", strings.Title(tier))
}
