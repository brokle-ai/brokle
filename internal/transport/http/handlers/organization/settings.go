package organization

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/core/domain/organization"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// SettingsHandler handles organization settings endpoints
type SettingsHandler struct {
	config          *config.Config
	logger          *logrus.Logger
	settingsService organization.OrganizationSettingsService
}

// NewSettingsHandler creates a new organization settings handler
func NewSettingsHandler(
	config *config.Config,
	logger *logrus.Logger,
	settingsService organization.OrganizationSettingsService,
) *SettingsHandler {
	return &SettingsHandler{
		config:          config,
		logger:          logger,
		settingsService: settingsService,
	}
}

// Request/Response Models

// OrganizationSetting represents an organization setting
type OrganizationSetting struct {
	ID             string      `json:"id" example:"set_1234567890" description:"Unique setting identifier"`
	OrganizationID string      `json:"organization_id" example:"org_1234567890" description:"Organization ID"`
	Key            string      `json:"key" example:"theme_color" description:"Setting key"`
	Value          interface{} `json:"value" swaggertype:"object" description:"Setting value (can be any JSON type)"`
	CreatedAt      string      `json:"created_at" example:"2024-01-01T00:00:00Z" description:"Creation timestamp"`
	UpdatedAt      string      `json:"updated_at" example:"2024-01-01T00:00:00Z" description:"Last update timestamp"`
}

// CreateSettingRequest represents the request to create an organization setting
type CreateSettingRequest struct {
	Key   string      `json:"key" binding:"required,min=1,max=255" example:"theme_color" description:"Setting key (1-255 characters)"`
	Value interface{} `json:"value" binding:"required" swaggertype:"object" description:"Setting value (any JSON type)"`
}

// UpdateSettingRequest represents the request to update an organization setting
type UpdateSettingRequest struct {
	Value interface{} `json:"value" binding:"required" swaggertype:"object" description:"New setting value (any JSON type)"`
}

// BulkSettingsRequest represents the request for bulk settings operations
type BulkSettingsRequest struct {
	Settings map[string]interface{} `json:"settings" binding:"required" swaggertype:"object" description:"Key-value pairs of settings"`
}

// SettingsListResponse represents the response when listing settings
type SettingsListResponse struct {
	Settings map[string]interface{} `json:"settings" swaggertype:"object" description:"Key-value pairs of all settings"`
}

// GetAllSettings handles GET /organizations/:orgId/settings
// @Summary Get all organization settings
// @Description Get all settings for an organization as key-value pairs
// @Tags Organization Settings
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Success 200 {object} response.SuccessResponse{data=SettingsListResponse} "Organization settings retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid organization ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 404 {object} response.ErrorResponse "Organization not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/organizations/{orgId}/settings [get]
func (h *SettingsHandler) GetAllSettings(c *gin.Context) {
	orgIDStr := c.Param("orgId")
	orgID, err := ulid.Parse(orgIDStr)
	if err != nil {
		h.logger.WithError(err).Error("Invalid organization ID")
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid_id", "Invalid organization ID", "")
		return
	}

	// Get user ID from context for access validation
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		response.ErrorWithStatus(c, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	_, ok := userIDValue.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		response.ErrorWithStatus(c, http.StatusInternalServerError, "internal_error", "Internal error", "")
		return
	}

	// TODO: Add access control validation here
	// For now, just proceed if user is authenticated

	settings, err := h.settingsService.GetAllSettings(c.Request.Context(), orgID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get organization settings")
		response.ErrorWithStatus(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve settings", "")
		return
	}

	response.Success(c, SettingsListResponse{
		Settings: settings,
	})
}

// GetSetting handles GET /organizations/:orgId/settings/:key
// @Summary Get specific organization setting
// @Description Get a specific setting by key for an organization
// @Tags Organization Settings
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Param key path string true "Setting key" example("theme_color")
// @Success 200 {object} response.SuccessResponse{data=OrganizationSetting} "Setting retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 404 {object} response.ErrorResponse "Setting not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/organizations/{orgId}/settings/{key} [get]
func (h *SettingsHandler) GetSetting(c *gin.Context) {
	orgIDStr := c.Param("orgId")
	orgID, err := ulid.Parse(orgIDStr)
	if err != nil {
		h.logger.WithError(err).Error("Invalid organization ID")
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid_id", "Invalid organization ID", "")
		return
	}

	key := c.Param("key")
	if key == "" {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid_key", "Setting key is required", "")
		return
	}

	setting, err := h.settingsService.GetSetting(c.Request.Context(), orgID, key)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get organization setting")
		if err.Error() == "organization setting not found" {
			response.ErrorWithStatus(c, http.StatusNotFound, "setting_not_found", "Setting not found", "")
			return
		}
		response.ErrorWithStatus(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve setting", "")
		return
	}

	value, _ := setting.GetValue()
	response.Success(c, OrganizationSetting{
		ID:             setting.ID.String(),
		OrganizationID: setting.OrganizationID.String(),
		Key:            setting.Key,
		Value:          value,
		CreatedAt:      setting.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      setting.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// CreateSetting handles POST /organizations/:orgId/settings
// @Summary Create organization setting
// @Description Create a new setting for an organization
// @Tags Organization Settings
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Param request body CreateSettingRequest true "Setting details"
// @Success 201 {object} response.SuccessResponse{data=OrganizationSetting} "Setting created successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid input or validation errors"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 409 {object} response.ErrorResponse "Conflict - setting key already exists"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/organizations/{orgId}/settings [post]
func (h *SettingsHandler) CreateSetting(c *gin.Context) {
	orgIDStr := c.Param("orgId")
	orgID, err := ulid.Parse(orgIDStr)
	if err != nil {
		h.logger.WithError(err).Error("Invalid organization ID")
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid_id", "Invalid organization ID", "")
		return
	}

	var req CreateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		response.ErrorWithStatus(c, http.StatusBadRequest, "validation_error", "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		response.ErrorWithStatus(c, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	userID, ok := userIDValue.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		response.ErrorWithStatus(c, http.StatusInternalServerError, "internal_error", "Internal error", "")
		return
	}

	domainReq := &organization.CreateOrganizationSettingRequest{
		Key:   req.Key,
		Value: req.Value,
	}

	setting, err := h.settingsService.CreateSetting(c.Request.Context(), orgID, userID, domainReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create organization setting")
		if err.Error() == "setting with this key already exists" {
			response.ErrorWithStatus(c, http.StatusConflict, "setting_exists", "Setting with this key already exists", "")
			return
		}
		response.ErrorWithStatus(c, http.StatusInternalServerError, "internal_error", "Failed to create setting", err.Error())
		return
	}

	value, _ := setting.GetValue()
	response.SuccessWithStatus(c, http.StatusCreated, OrganizationSetting{
		ID:             setting.ID.String(),
		OrganizationID: setting.OrganizationID.String(),
		Key:            setting.Key,
		Value:          value,
		CreatedAt:      setting.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      setting.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// UpdateSetting handles PUT /organizations/:orgId/settings/:key
// @Summary Update organization setting
// @Description Update an existing setting for an organization
// @Tags Organization Settings
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Param key path string true "Setting key" example("theme_color")
// @Param request body UpdateSettingRequest true "Updated setting value"
// @Success 200 {object} response.SuccessResponse "Setting updated successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid input or validation errors"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 404 {object} response.ErrorResponse "Setting not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/organizations/{orgId}/settings/{key} [put]
func (h *SettingsHandler) UpdateSetting(c *gin.Context) {
	orgIDStr := c.Param("orgId")
	orgID, err := ulid.Parse(orgIDStr)
	if err != nil {
		h.logger.WithError(err).Error("Invalid organization ID")
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid_id", "Invalid organization ID", "")
		return
	}

	key := c.Param("key")
	if key == "" {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid_key", "Setting key is required", "")
		return
	}

	var req UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		response.ErrorWithStatus(c, http.StatusBadRequest, "validation_error", "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		response.ErrorWithStatus(c, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	userID, ok := userIDValue.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		response.ErrorWithStatus(c, http.StatusInternalServerError, "internal_error", "Internal error", "")
		return
	}

	domainReq := &organization.UpdateOrganizationSettingRequest{
		Value: req.Value,
	}

	err = h.settingsService.UpdateSetting(c.Request.Context(), orgID, key, userID, domainReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update organization setting")
		if err.Error() == "setting not found" {
			response.ErrorWithStatus(c, http.StatusNotFound, "setting_not_found", "Setting not found", "")
			return
		}
		response.ErrorWithStatus(c, http.StatusInternalServerError, "internal_error", "Failed to update setting", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Setting updated successfully"})
}

// DeleteSetting handles DELETE /organizations/:orgId/settings/:key
// @Summary Delete organization setting
// @Description Delete a setting for an organization
// @Tags Organization Settings
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Param key path string true "Setting key" example("theme_color")
// @Success 204 "Setting deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Bad request - invalid parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 404 {object} response.ErrorResponse "Setting not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/organizations/{orgId}/settings/{key} [delete]
func (h *SettingsHandler) DeleteSetting(c *gin.Context) {
	orgIDStr := c.Param("orgId")
	orgID, err := ulid.Parse(orgIDStr)
	if err != nil {
		h.logger.WithError(err).Error("Invalid organization ID")
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid_id", "Invalid organization ID", "")
		return
	}

	key := c.Param("key")
	if key == "" {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid_key", "Setting key is required", "")
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		response.ErrorWithStatus(c, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	userID, ok := userIDValue.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		response.ErrorWithStatus(c, http.StatusInternalServerError, "internal_error", "Internal error", "")
		return
	}

	err = h.settingsService.DeleteSetting(c.Request.Context(), orgID, key, userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete organization setting")
		if err.Error() == "setting not found" {
			response.ErrorWithStatus(c, http.StatusNotFound, "setting_not_found", "Setting not found", "")
			return
		}
		response.ErrorWithStatus(c, http.StatusInternalServerError, "internal_error", "Failed to delete setting", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// BulkCreateSettings handles POST /organizations/:orgId/settings/bulk
// @Summary Bulk create organization settings
// @Description Create multiple settings for an organization in a single request
// @Tags Organization Settings
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Param request body BulkSettingsRequest true "Settings to create"
// @Success 501 {object} response.ErrorResponse "Not implemented"
// @Security BearerAuth
// @Router /api/v1/organizations/{orgId}/settings/bulk [post]
func (h *SettingsHandler) BulkCreateSettings(c *gin.Context) {
	response.ErrorWithStatus(c, http.StatusNotImplemented, "not_implemented", "Bulk operations are not implemented", "")
}

// ExportSettings handles GET /organizations/:orgId/settings/export
// @Summary Export organization settings
// @Description Export all settings for an organization
// @Tags Organization Settings
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Success 501 {object} response.ErrorResponse "Not implemented"
// @Security BearerAuth
// @Router /api/v1/organizations/{orgId}/settings/export [get]
func (h *SettingsHandler) ExportSettings(c *gin.Context) {
	response.ErrorWithStatus(c, http.StatusNotImplemented, "not_implemented", "Export/import operations are not implemented", "")
}

// ImportSettings handles POST /organizations/:orgId/settings/import
// @Summary Import organization settings
// @Description Import settings for an organization, creating or updating as needed
// @Tags Organization Settings
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Param request body BulkSettingsRequest true "Settings to import"
// @Success 501 {object} response.ErrorResponse "Not implemented"
// @Security BearerAuth
// @Router /api/v1/organizations/{orgId}/settings/import [post]
func (h *SettingsHandler) ImportSettings(c *gin.Context) {
	response.ErrorWithStatus(c, http.StatusNotImplemented, "not_implemented", "Export/import operations are not implemented", "")
}

// ResetToDefaults handles POST /organizations/:orgId/settings/reset
// @Summary Reset organization settings to defaults
// @Description Reset all organization settings to default values (removes all current settings)
// @Tags Organization Settings
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID" example("org_1234567890")
// @Success 501 {object} response.ErrorResponse "Not implemented"
// @Security BearerAuth
// @Router /api/v1/organizations/{orgId}/settings/reset [post]
func (h *SettingsHandler) ResetToDefaults(c *gin.Context) {
	response.ErrorWithStatus(c, http.StatusNotImplemented, "not_implemented", "Reset operations are not implemented", "")
}
