package enterprise

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/ee/analytics"
	"brokle/internal/ee/compliance"
	"brokle/internal/ee/rbac"
	"brokle/internal/ee/sso"
	"brokle/internal/middleware"
	"brokle/internal/services"
)

// Handler handles enterprise-specific endpoints
type Handler struct {
	config              *config.Config
	logger              *logrus.Logger
	licenseService      *services.LicenseService
	complianceService   compliance.Compliance
	ssoService          sso.SSOProvider
	rbacService         rbac.RBACManager
	enterpriseAnalytics analytics.EnterpriseAnalytics
}

// NewHandler creates a new enterprise handler
func NewHandler(
	cfg *config.Config,
	logger *logrus.Logger,
	licenseService *services.LicenseService,
	complianceService compliance.Compliance,
	ssoService sso.SSOProvider,
	rbacService rbac.RBACManager,
	enterpriseAnalytics analytics.EnterpriseAnalytics,
) *Handler {
	return &Handler{
		config:              cfg,
		logger:              logger,
		licenseService:      licenseService,
		complianceService:   complianceService,
		ssoService:          ssoService,
		rbacService:         rbacService,
		enterpriseAnalytics: enterpriseAnalytics,
	}
}

// RegisterRoutes registers enterprise routes with middleware
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// License management endpoints
	license := r.Group("/license")
	{
		license.GET("/status", h.GetLicenseStatus)
		license.GET("/features", h.GetAvailableFeatures)
	}

	// Enterprise SSO endpoints (requires enterprise license)
	sso := r.Group("/sso")
	sso.Use(middleware.RequireEnterpriseLicense(h.licenseService, h.logger))
	{
		sso.GET("/providers", h.GetSSOProviders)
		sso.POST("/configure", h.ConfigureSSO)
		sso.GET("/login", h.GetSSOLoginURL)
		sso.POST("/callback", h.HandleSSOCallback)
	}

	// Enterprise RBAC endpoints (requires business+ license)
	roles := r.Group("/rbac")
	roles.Use(middleware.EnterpriseFeature("advanced_rbac", h.licenseService, h.logger))
	{
		roles.GET("/roles", h.ListRoles)
		roles.POST("/roles", h.CreateRole)
		roles.PUT("/roles/:id", h.UpdateRole)
		roles.DELETE("/roles/:id", h.DeleteRole)
		roles.POST("/roles/:id/assign/:user_id", h.AssignRole)
		roles.DELETE("/roles/:id/unassign/:user_id", h.UnassignRole)
		roles.GET("/users/:id/permissions", h.GetUserPermissions)
	}

	// Enterprise compliance endpoints (requires business+ license)
	compliance := r.Group("/compliance")
	compliance.Use(middleware.EnterpriseFeature("custom_compliance", h.licenseService, h.logger))
	{
		compliance.POST("/validate", h.ValidateCompliance)
		compliance.GET("/reports/audit", h.GenerateAuditReport)
		compliance.POST("/anonymize", h.AnonymizePII)
		compliance.GET("/status/soc2", h.CheckSOC2Compliance)
		compliance.GET("/status/hipaa", h.CheckHIPAACompliance)
		compliance.GET("/status/gdpr", h.CheckGDPRCompliance)
	}

	// Enterprise analytics endpoints (requires business+ license)
	enterpriseAnalytics := r.Group("/analytics")
	enterpriseAnalytics.Use(middleware.EnterpriseFeature("predictive_insights", h.licenseService, h.logger))
	{
		enterpriseAnalytics.GET("/insights", h.GetPredictiveInsights)
		enterpriseAnalytics.GET("/dashboards", h.ListCustomDashboards)
		enterpriseAnalytics.POST("/dashboards", h.CreateCustomDashboard)
		enterpriseAnalytics.PUT("/dashboards/:id", h.UpdateCustomDashboard)
		enterpriseAnalytics.DELETE("/dashboards/:id", h.DeleteCustomDashboard)
		enterpriseAnalytics.POST("/reports/generate", h.GenerateAdvancedReport)
		enterpriseAnalytics.POST("/data/export", h.ExportData)
		enterpriseAnalytics.POST("/ml/predict", h.RunMLModel)
	}
}

// License Management Handlers

func (h *Handler) GetLicenseStatus(c *gin.Context) {
	status, err := h.licenseService.ValidateLicense(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to validate license")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate license"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"license": status,
		"tier":    h.config.GetLicenseTier(),
	})
}

func (h *Handler) GetAvailableFeatures(c *gin.Context) {
	tier := h.config.GetLicenseTier()

	allFeatures := map[string]interface{}{
		"advanced_rbac": map[string]interface{}{
			"name":          "Advanced Role-Based Access Control",
			"description":   "Custom roles and fine-grained permissions",
			"required_tier": "business",
			"available":     h.config.IsEnterpriseFeatureEnabled("advanced_rbac"),
		},
		"sso_integration": map[string]interface{}{
			"name":          "Single Sign-On Integration",
			"description":   "SAML, OIDC, and OAuth2 integration",
			"required_tier": "business",
			"available":     h.config.IsEnterpriseFeatureEnabled("sso_integration"),
		},
		"custom_compliance": map[string]interface{}{
			"name":          "Custom Compliance Controls",
			"description":   "SOC2, HIPAA, GDPR compliance features",
			"required_tier": "business",
			"available":     h.config.IsEnterpriseFeatureEnabled("custom_compliance"),
		},
		"predictive_insights": map[string]interface{}{
			"name":          "Predictive Analytics",
			"description":   "ML-powered insights and forecasting",
			"required_tier": "business",
			"available":     h.config.IsEnterpriseFeatureEnabled("predictive_insights"),
		},
		"custom_dashboards": map[string]interface{}{
			"name":          "Custom Dashboard Builder",
			"description":   "Create custom analytics dashboards",
			"required_tier": "business",
			"available":     h.config.IsEnterpriseFeatureEnabled("custom_dashboards"),
		},
		"on_premise_deployment": map[string]interface{}{
			"name":          "On-Premise Deployment",
			"description":   "Deploy in your own infrastructure",
			"required_tier": "enterprise",
			"available":     h.config.IsEnterpriseFeatureEnabled("on_premise_deployment"),
		},
		"dedicated_support": map[string]interface{}{
			"name":          "Dedicated Support",
			"description":   "Priority support and dedicated success manager",
			"required_tier": "enterprise",
			"available":     h.config.IsEnterpriseFeatureEnabled("dedicated_support"),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"current_tier": tier,
		"features":     allFeatures,
		"upgrade_url":  "https://brokle.com/pricing",
	})
}

// SSO Handlers

func (h *Handler) GetSSOProviders(c *gin.Context) {
	providers, err := h.ssoService.GetSupportedProviders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{
			"error":       err.Error(),
			"upgrade_url": "https://brokle.com/pricing",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

func (h *Handler) ConfigureSSO(c *gin.Context) {
	var req struct {
		Provider string `json:"provider" binding:"required"`
		Config   string `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.ssoService.ConfigureProvider(c.Request.Context(), req.Provider, req.Config)
	if err != nil {
		if strings.Contains(err.Error(), "Enterprise license") {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":       err.Error(),
				"upgrade_url": "https://brokle.com/pricing",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SSO provider configured successfully"})
}

func (h *Handler) GetSSOLoginURL(c *gin.Context) {
	url, err := h.ssoService.GetLoginURL(c.Request.Context())
	if err != nil {
		if strings.Contains(err.Error(), "Enterprise license") {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":       err.Error(),
				"upgrade_url": "https://brokle.com/pricing",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"login_url": url})
}

func (h *Handler) HandleSSOCallback(c *gin.Context) {
	assertion := c.PostForm("SAMLResponse")
	if assertion == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing SAML assertion"})
		return
	}

	user, err := h.ssoService.ValidateAssertion(c.Request.Context(), assertion)
	if err != nil {
		if strings.Contains(err.Error(), "Enterprise license") {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":       err.Error(),
				"upgrade_url": "https://brokle.com/pricing",
			})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "SSO authentication successful",
		"user":    user,
	})
}

// RBAC Handlers

func (h *Handler) ListRoles(c *gin.Context) {
	roles, err := h.rbacService.ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

func (h *Handler) CreateRole(c *gin.Context) {
	var role rbac.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.rbacService.CreateRole(c.Request.Context(), &role)
	if err != nil {
		if strings.Contains(err.Error(), "Enterprise license") {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":       err.Error(),
				"upgrade_url": "https://brokle.com/pricing",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Role created successfully", "role": role})
}

func (h *Handler) UpdateRole(c *gin.Context) {
	roleID := c.Param("id")

	var role rbac.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.rbacService.UpdateRole(c.Request.Context(), roleID, &role)
	if err != nil {
		if strings.Contains(err.Error(), "Enterprise license") {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":       err.Error(),
				"upgrade_url": "https://brokle.com/pricing",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role updated successfully"})
}

func (h *Handler) DeleteRole(c *gin.Context) {
	roleID := c.Param("id")

	err := h.rbacService.DeleteRole(c.Request.Context(), roleID)
	if err != nil {
		if strings.Contains(err.Error(), "Enterprise license") {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":       err.Error(),
				"upgrade_url": "https://brokle.com/pricing",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}

func (h *Handler) AssignRole(c *gin.Context) {
	roleID := c.Param("id")
	userID := c.Param("user_id")

	err := h.rbacService.AssignRoleToUser(c.Request.Context(), userID, roleID)
	if err != nil {
		if strings.Contains(err.Error(), "Enterprise license") {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":       err.Error(),
				"upgrade_url": "https://brokle.com/pricing",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role assigned successfully"})
}

func (h *Handler) UnassignRole(c *gin.Context) {
	roleID := c.Param("id")
	userID := c.Param("user_id")

	err := h.rbacService.RemoveRoleFromUser(c.Request.Context(), userID, roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role unassigned successfully"})
}

func (h *Handler) GetUserPermissions(c *gin.Context) {
	userID := c.Param("id")

	permissions, err := h.rbacService.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

// Compliance Handlers

func (h *Handler) ValidateCompliance(c *gin.Context) {
	var data interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.complianceService.ValidateCompliance(c.Request.Context(), data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Compliance validation passed"})
}

func (h *Handler) GenerateAuditReport(c *gin.Context) {
	report, err := h.complianceService.GenerateAuditReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=audit-report.json")
	c.Data(http.StatusOK, "application/json", report)
}

func (h *Handler) AnonymizePII(c *gin.Context) {
	var data interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	anonymized, err := h.complianceService.AnonymizePII(c.Request.Context(), data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"anonymized_data": anonymized})
}

func (h *Handler) CheckSOC2Compliance(c *gin.Context) {
	compliant, err := h.complianceService.CheckSOC2Compliance(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"soc2_compliant": compliant,
		"status": map[string]interface{}{
			"compliant":    compliant,
			"framework":    "SOC 2 Type II",
			"last_checked": "2024-01-01T00:00:00Z",
		},
	})
}

func (h *Handler) CheckHIPAACompliance(c *gin.Context) {
	compliant, err := h.complianceService.CheckHIPAACompliance(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hipaa_compliant": compliant,
		"status": map[string]interface{}{
			"compliant":    compliant,
			"framework":    "HIPAA",
			"last_checked": "2024-01-01T00:00:00Z",
		},
	})
}

func (h *Handler) CheckGDPRCompliance(c *gin.Context) {
	compliant, err := h.complianceService.CheckGDPRCompliance(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"gdpr_compliant": compliant,
		"status": map[string]interface{}{
			"compliant":    compliant,
			"framework":    "GDPR",
			"last_checked": "2024-01-01T00:00:00Z",
		},
	})
}

// Enterprise Analytics Handlers

func (h *Handler) GetPredictiveInsights(c *gin.Context) {
	timeRange := c.DefaultQuery("time_range", "30d")

	insights, err := h.enterpriseAnalytics.GeneratePredictiveInsights(c.Request.Context(), timeRange)
	if err != nil {
		if strings.Contains(err.Error(), "Enterprise license") {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":       err.Error(),
				"upgrade_url": "https://brokle.com/pricing",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"insights": insights})
}

func (h *Handler) ListCustomDashboards(c *gin.Context) {
	dashboards, err := h.enterpriseAnalytics.ListCustomDashboards(c.Request.Context())
	if err != nil {
		if strings.Contains(err.Error(), "Enterprise license") {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":       err.Error(),
				"upgrade_url": "https://brokle.com/pricing",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"dashboards": dashboards})
}

func (h *Handler) CreateCustomDashboard(c *gin.Context) {
	var dashboard analytics.Dashboard
	if err := c.ShouldBindJSON(&dashboard); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.enterpriseAnalytics.CreateCustomDashboard(c.Request.Context(), &dashboard)
	if err != nil {
		if strings.Contains(err.Error(), "Enterprise license") {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":       err.Error(),
				"upgrade_url": "https://brokle.com/pricing",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Dashboard created successfully", "dashboard": dashboard})
}

func (h *Handler) UpdateCustomDashboard(c *gin.Context) {
	dashboardID := c.Param("id")

	var dashboard analytics.Dashboard
	if err := c.ShouldBindJSON(&dashboard); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.enterpriseAnalytics.UpdateCustomDashboard(c.Request.Context(), dashboardID, &dashboard)
	if err != nil {
		if strings.Contains(err.Error(), "Enterprise license") {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":       err.Error(),
				"upgrade_url": "https://brokle.com/pricing",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dashboard updated successfully"})
}

func (h *Handler) DeleteCustomDashboard(c *gin.Context) {
	dashboardID := c.Param("id")

	err := h.enterpriseAnalytics.DeleteCustomDashboard(c.Request.Context(), dashboardID)
	if err != nil {
		if strings.Contains(err.Error(), "Enterprise license") {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":       err.Error(),
				"upgrade_url": "https://brokle.com/pricing",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dashboard deleted successfully"})
}

func (h *Handler) GenerateAdvancedReport(c *gin.Context) {
	var req analytics.ReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.enterpriseAnalytics.GenerateAdvancedReport(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "Enterprise license") {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":       err.Error(),
				"upgrade_url": "https://brokle.com/pricing",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"report": report})
}

func (h *Handler) ExportData(c *gin.Context) {
	format := c.DefaultQuery("format", "json")

	var query analytics.ExportQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := h.enterpriseAnalytics.ExportData(c.Request.Context(), format, &query)
	if err != nil {
		if strings.Contains(err.Error(), "Enterprise license") {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":       err.Error(),
				"upgrade_url": "https://brokle.com/pricing",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	contentType := "application/json"
	if format == "csv" {
		contentType = "text/csv"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename=export."+format)
	c.Data(http.StatusOK, contentType, data)
}

func (h *Handler) RunMLModel(c *gin.Context) {
	modelName := c.Query("model")
	if modelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Model name is required"})
		return
	}

	var data interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.enterpriseAnalytics.RunMLModel(c.Request.Context(), modelName, data)
	if err != nil {
		if strings.Contains(err.Error(), "Enterprise license") {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":       err.Error(),
				"upgrade_url": "https://brokle.com/pricing",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"model":  modelName,
		"result": result,
	})
}
