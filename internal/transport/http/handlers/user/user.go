package user

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/core/domain/user"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
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

// UserProfileResponse represents the complete user profile response
// @Description Complete user profile information including basic info and extended profile
type UserProfileResponse struct {
	ID                    ulid.ULID        `json:"id" example:"01K4FHGHT3XX9WFM293QPZ5G9V" description:"User unique identifier" swaggertype:"string"`
	Email                 string           `json:"email" example:"user@example.com" description:"User email address"`
	Name                  string           `json:"name" example:"John Doe" description:"User full name"`
	FirstName             string           `json:"first_name" example:"John" description:"User first name"`
	LastName              string           `json:"last_name" example:"Doe" description:"User last name"`
	AvatarURL             string           `json:"avatar_url" example:"https://example.com/avatar.jpg" description:"Profile avatar URL"`
	IsEmailVerified       bool             `json:"is_email_verified" example:"true" description:"Email verification status"`
	OnboardingCompleted   bool             `json:"onboarding_completed" example:"true" description:"Onboarding completion status"`
	IsActive              bool             `json:"is_active" example:"true" description:"Account active status"`
	CreatedAt             time.Time        `json:"created_at" example:"2025-01-01T00:00:00Z" description:"Account creation timestamp"`
	LastLoginAt           *time.Time       `json:"last_login_at,omitempty" example:"2025-01-02T10:30:00Z" description:"Last login timestamp"`
	DefaultOrganizationID *ulid.ULID       `json:"default_organization_id,omitempty" example:"01K4FHGHT3XX9WFM293QPZ5G9V" description:"Default organization ID" swaggertype:"string"`
	Profile               *UserProfileData `json:"profile,omitempty" description:"Extended profile information"`
	Completeness          int              `json:"completeness" example:"85" description:"Profile completeness percentage"`
}

// UserProfileData represents extended profile information
// @Description Extended user profile data including bio, location, and social links
type UserProfileData struct {
	Bio         *string `json:"bio,omitempty" example:"Software engineer passionate about AI" description:"User biography"`
	Location    *string `json:"location,omitempty" example:"San Francisco, CA" description:"User location"`
	Website     *string `json:"website,omitempty" example:"https://johndoe.com" description:"Personal website URL"`
	TwitterURL  *string `json:"twitter_url,omitempty" example:"https://twitter.com/johndoe" description:"Twitter profile URL"`
	LinkedInURL *string `json:"linkedin_url,omitempty" example:"https://linkedin.com/in/johndoe" description:"LinkedIn profile URL"`
	GitHubURL   *string `json:"github_url,omitempty" example:"https://github.com/johndoe" description:"GitHub profile URL"`
	Timezone    string  `json:"timezone" example:"UTC" description:"User timezone preference"`
	Language    string  `json:"language" example:"en" description:"User language preference"`
	Theme       string  `json:"theme" example:"dark" description:"UI theme preference"`
}

// GetProfile handles GET /users/me
// @Summary Get current user profile
// @Description Get the profile information of the currently authenticated user
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserProfileResponse "User profile retrieved successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/me [get]
func (h *Handler) GetProfile(c *gin.Context) {
	// Get user ID from middleware (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in request")
		response.Unauthorized(c, "Authentication required")
		return
	}

	userID, ok := userIDInterface.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type")
		response.InternalServerError(c, "Authentication error")
		return
	}

	// Get basic user information
	userData, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user")
		response.NotFound(c, "User not found")
		return
	}

	// Get extended profile information (this might not exist for all users)
	profileData, err := h.profileService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		// Profile might not exist yet, which is okay
		h.logger.WithError(err).WithField("user_id", userID).Debug("Profile not found, using defaults")
		profileData = nil
	}

	// Get profile completeness
	completeness, err := h.profileService.GetProfileCompleteness(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get profile completeness")
		// Continue with 0% completeness
	}

	// Build response
	profileResponse := &UserProfileResponse{
		ID:                    userData.ID,
		Email:                 userData.Email,
		Name:                  userData.GetFullName(),
		FirstName:             userData.FirstName,
		LastName:              userData.LastName,
		AvatarURL:             "", // Now stored in profile
		IsEmailVerified:       userData.IsEmailVerified,
		OnboardingCompleted:   userData.OnboardingCompleted,
		IsActive:              userData.IsActive,
		CreatedAt:             userData.CreatedAt,
		LastLoginAt:           userData.LastLoginAt,
		DefaultOrganizationID: userData.DefaultOrganizationID,
		Completeness:          0, // Default
	}

	// Add extended profile data if available
	if profileData != nil {
		profileResponse.Profile = &UserProfileData{
			Bio:         profileData.Bio,
			Location:    profileData.Location,
			Website:     profileData.Website,
			TwitterURL:  profileData.TwitterURL,
			LinkedInURL: profileData.LinkedInURL,
			GitHubURL:   profileData.GitHubURL,
			Timezone:    profileData.Timezone,
			Language:    profileData.Language,
			Theme:       profileData.Theme,
		}
	}

	// Add completeness percentage if available
	if completeness != nil {
		profileResponse.Completeness = completeness.OverallScore
	}

	h.logger.WithField("user_id", userID).Info("User profile retrieved successfully")
	response.Success(c, profileResponse)
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
// @Success 200 {object} UserProfileResponse "Profile updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/users/me [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	// Get user ID from middleware (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in request")
		response.Unauthorized(c, "Authentication required")
		return
	}

	userID, ok := userIDInterface.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type")
		response.InternalServerError(c, "Authentication error")
		return
	}

	// Parse request body
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid profile update request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Update basic user information (name) via user service
	if req.FirstName != nil || req.LastName != nil {
		userUpdateReq := &user.UpdateUserRequest{
			FirstName: req.FirstName,
			LastName:  req.LastName,
		}

		_, err := h.userService.UpdateUser(c.Request.Context(), userID, userUpdateReq)
		if err != nil {
			h.logger.WithError(err).WithField("user_id", userID).Error("Failed to update user")
			response.InternalServerError(c, "Failed to update user information")
			return
		}
	}

	// Update profile-specific information (timezone, language) via profile service
	if req.Timezone != nil || req.Language != nil {
		profileUpdateReq := &user.UpdateProfileRequest{
			Timezone: req.Timezone,
			Language: req.Language,
		}

		_, err := h.profileService.UpdateProfile(c.Request.Context(), userID, profileUpdateReq)
		if err != nil {
			// Profile might not exist yet - log but don't fail the entire operation
			h.logger.WithError(err).WithField("user_id", userID).Debug("Profile update failed, profile may not exist yet")
			// For now, we'll skip profile updates if the profile doesn't exist
			// In a future iteration, we could create the profile automatically
		}
	}

	// Return updated profile (call GetProfile internally to get consistent response)
	h.logger.WithField("user_id", userID).Info("Profile updated successfully")

	// Re-fetch and return updated profile
	h.GetProfile(c)
}

