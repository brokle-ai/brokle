package user

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/core/domain/user"
	"brokle/pkg/response"
)

// Handler handles user endpoints
type Handler struct {
	config         *config.Config
	logger         *logrus.Logger
	userService    user.UserService
	profileService user.ProfileService
}

// NewHandler creates a new user handler
func NewHandler(config *config.Config, logger *logrus.Logger, userService user.UserService, profileService user.ProfileService) *Handler {
	return &Handler{
		config:         config,
		logger:         logger,
		userService:    userService,
		profileService: profileService,
	}
}

// GetProfile handles GET /users/me
// @Summary Get current user profile
// @Description Get the profile information of the currently authenticated user
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse "User profile retrieved successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/me [get]
func (h *Handler) GetProfile(c *gin.Context) {
	response.Success(c, gin.H{"message": "Get user profile - TODO"})
}

// UpdateProfileRequest represents the update profile request payload
// @Description User profile update information
type UpdateProfileRequest struct {
	FirstName *string `json:"first_name,omitempty" example:"John" description:"User first name"`
	LastName  *string `json:"last_name,omitempty" example:"Doe" description:"User last name"`
	AvatarURL *string `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg" description:"Profile avatar URL"`
	Phone     *string `json:"phone,omitempty" example:"+1234567890" description:"User phone number"`
	Timezone  *string `json:"timezone,omitempty" example:"UTC" description:"User timezone"`
	Language  *string `json:"language,omitempty" example:"en" description:"User language preference (ISO 639-1 code)"`
}

// UpdateProfile handles PUT /users/me
// @Summary Update current user profile
// @Description Update the profile information of the currently authenticated user
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateProfileRequest true "Profile update information"
// @Success 200 {object} response.MessageResponse "Profile updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/me [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	response.Success(c, gin.H{"message": "Update user profile - TODO"})
}

// GetPreferences handles GET /users/me/preferences
// @Summary Get user preferences
// @Description Get the preferences settings of the currently authenticated user
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse "User preferences retrieved successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/me/preferences [get]
func (h *Handler) GetPreferences(c *gin.Context) {
	response.Success(c, gin.H{"message": "Get user preferences - TODO"})
}

// UpdatePreferencesRequest represents the update preferences request payload
// @Description User preferences update information
type UpdatePreferencesRequest struct {
	Theme                string            `json:"theme,omitempty" example:"dark" description:"UI theme preference (light, dark, auto)"`
	Language             string            `json:"language,omitempty" example:"en" description:"Language preference (ISO 639-1 code)"`
	Timezone             string            `json:"timezone,omitempty" example:"UTC" description:"Timezone preference"`
	EmailNotifications   *bool             `json:"email_notifications,omitempty" example:"true" description:"Enable email notifications"`
	WebhookNotifications *bool             `json:"webhook_notifications,omitempty" example:"false" description:"Enable webhook notifications"`
	DashboardSettings    map[string]interface{} `json:"dashboard_settings,omitempty" description:"Custom dashboard configuration"`
}

// UpdatePreferences handles PUT /users/me/preferences
// @Summary Update user preferences
// @Description Update the preferences settings of the currently authenticated user
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdatePreferencesRequest true "Preferences update information"
// @Success 200 {object} response.MessageResponse "Preferences updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/me/preferences [put]
func (h *Handler) UpdatePreferences(c *gin.Context) {
	response.Success(c, gin.H{"message": "Update user preferences - TODO"})
}