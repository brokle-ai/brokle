package middleware

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"brokle/internal/config"
	license "brokle/internal/ee/licensing"
	"brokle/internal/errors"
)

// EnterpriseFeature middleware checks if an enterprise feature is enabled
func EnterpriseFeature(feature string, licenseService *license.LicenseService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := c.MustGet("config").(*config.Config)

		// Allow all features in development mode
		if cfg.IsDevelopment() {
			c.Header("X-Feature-Mode", "development")
			c.Next()
			return
		}

		// Check if feature is available in current license
		available, err := licenseService.CheckFeatureEntitlement(c.Request.Context(), feature)
		if err != nil {
			logger.Error("Failed to check feature entitlement", "error", err, "feature", feature)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to validate feature access",
				"feature": feature,
			})
			c.Abort()
			return
		}

		if !available {
			currentTier := cfg.GetLicenseTier()
			requiredTier := getRequiredTierForFeature(feature)

			// Log feature access attempt for analytics
			logger.Info("Enterprise feature access denied", "feature", feature, "current_tier", currentTier, "required_tier", requiredTier, "user_agent", c.Request.UserAgent(), "ip", c.ClientIP())

			enterpriseError := errors.NewFeatureNotAvailableError(feature, currentTier, requiredTier)
			c.JSON(enterpriseError.HTTPStatus(), gin.H{
				"error": enterpriseError,
			})
			c.Abort()
			return
		}

		// Feature is available, add to context and continue
		c.Header("X-Feature-Tier", cfg.GetLicenseTier())
		c.Set("enterprise_feature", feature)
		c.Next()
	}
}

// RequireEnterpriseLicense checks for valid enterprise license
func RequireEnterpriseLicense(licenseService *license.LicenseService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := c.MustGet("config").(*config.Config)

		// Allow in development mode
		if cfg.IsDevelopment() {
			c.Header("X-License-Mode", "development")
			c.Next()
			return
		}

		status, err := licenseService.ValidateLicense(c.Request.Context())
		if err != nil {
			logger.Error("Failed to validate license", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "License validation failed",
			})
			c.Abort()
			return
		}

		if !status.IsValid || !cfg.IsEnterpriseLicense() {
			logger.Info("Enterprise license required", "is_valid", status.IsValid, "tier", cfg.GetLicenseTier())

			enterpriseError := errors.NewLicenseRequiredError(cfg.GetLicenseTier())
			c.JSON(enterpriseError.HTTPStatus(), gin.H{
				"error": enterpriseError,
			})
			c.Abort()
			return
		}

		// Valid enterprise license, continue
		c.Set("license_status", status)
		c.Next()
	}
}

// CheckUsageLimit middleware validates usage against license limits
func CheckUsageLimit(limitType string, licenseService *license.LicenseService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := c.MustGet("config").(*config.Config)

		// Skip in development mode
		if cfg.IsDevelopment() {
			c.Next()
			return
		}

		withinLimit, remaining, err := licenseService.CheckUsageLimit(c.Request.Context(), limitType)
		if err != nil {
			logger.Error("Failed to check usage limit", "error", err, "limit_type", limitType)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to validate usage limits",
			})
			c.Abort()
			return
		}

		if !withinLimit {
			logger.Info("Usage limit exceeded", "limit_type", limitType, "remaining", remaining, "tier", cfg.GetLicenseTier())

			// Calculate next reset date (monthly reset)
			resetDate := time.Now().AddDate(0, 1, -time.Now().Day()).Add(-time.Hour * time.Duration(time.Now().Hour()))

			enterpriseError := errors.NewUsageLimitExceededError(limitType, cfg.GetLicenseTier(), remaining, &resetDate)
			c.JSON(enterpriseError.HTTPStatus(), gin.H{
				"error": enterpriseError,
			})
			c.Abort()
			return
		}

		// Within limits, add usage info to context
		c.Header("X-Usage-Remaining", strconv.FormatInt(remaining, 10))
		c.Next()

		// Increment usage after successful request (in background)
		go func() {
			if err := licenseService.UpdateUsage(c.Request.Context(), limitType, 1); err != nil {
				logger.Error("Failed to update usage", "error", err, "limit_type", limitType)
			}
		}()
	}
}

// EnterpriseFeatureWithFallback allows graceful degradation
func EnterpriseFeatureWithFallback(feature string, fallbackHandler gin.HandlerFunc, licenseService *license.LicenseService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := c.MustGet("config").(*config.Config)

		// Always allow in development
		if cfg.IsDevelopment() {
			c.Next()
			return
		}

		// Check feature availability
		available, err := licenseService.CheckFeatureEntitlement(c.Request.Context(), feature)
		if err != nil || !available {
			if err != nil {
				logger.Warn("Feature check failed, using fallback", "error", err, "feature", feature)
			} else {
				logger.Info("Feature not available, using fallback", "feature", feature)
			}

			// Use fallback handler
			c.Set("feature_fallback", true)
			fallbackHandler(c)
			return
		}

		// Feature available, use enterprise implementation
		c.Set("enterprise_feature", feature)
		c.Next()
	}
}

// Helper functions

func getRequiredTierForFeature(feature string) string {
	// Map features to required tiers
	tierMap := map[string]string{
		"advanced_rbac":         "business",
		"sso_integration":       "business",
		"custom_compliance":     "business",
		"predictive_insights":   "business",
		"custom_dashboards":     "business",
		"on_premise_deployment": "enterprise",
		"dedicated_support":     "enterprise",
		"advanced_integrations": "business",
		"cross_org_analytics":   "enterprise",
	}

	if tier, exists := tierMap[feature]; exists {
		return tier
	}
	return "business" // Default to business tier
}
