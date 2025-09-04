package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/user"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// Handler handles authentication endpoints
type Handler struct {
	config      *config.Config
	logger      *logrus.Logger
	authService auth.AuthService
	userService user.Service
}

// NewHandler creates a new auth handler
func NewHandler(config *config.Config, logger *logrus.Logger, authService auth.AuthService, userService user.Service) *Handler {
	return &Handler{
		config:      config,
		logger:      logger,
		authService: authService,
		userService: userService,
	}
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login handles user login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid login request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Create auth login request
	authReq := &auth.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	// Attempt login
	loginResp, err := h.authService.Login(c.Request.Context(), authReq)
	if err != nil {
		h.logger.WithError(err).WithField("email", req.Email).Error("Login failed")
		response.Unauthorized(c, "Invalid credentials")
		return
	}

	h.logger.WithField("user_id", loginResp.User.ID).Info("User logged in successfully")
	response.Success(c, loginResp)
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	FirstName string `json:"first_name" binding:"required,min=1,max=100"`
	LastName  string `json:"last_name" binding:"required,min=1,max=100"`
	Password  string `json:"password" binding:"required,min=8"`
	Timezone  string `json:"timezone,omitempty"`
	Language  string `json:"language,omitempty"`
}

// Signup handles user registration  
func (h *Handler) Signup(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid registration request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Create auth register request
	authReq := &auth.RegisterRequest{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  req.Password,
		Timezone:  req.Timezone,
		Language:  req.Language,
	}

	// Register user
	createUserReq := &user.CreateUserRequest{
		Email:     authReq.Email,
		FirstName: authReq.FirstName,
		LastName:  authReq.LastName,
		Password:  authReq.Password,
		Timezone:  authReq.Timezone,
		Language:  authReq.Language,
	}
	
	newUser, err := h.userService.Register(c.Request.Context(), createUserReq)
	if err != nil {
		h.logger.WithError(err).WithField("email", req.Email).Error("Registration failed")
		response.Conflict(c, "Registration failed")
		return
	}

	h.logger.WithField("user_id", newUser.ID).Info("User registered successfully")
	response.Success(c, gin.H{
		"message": "Registration successful",
		"user_id": newUser.ID,
	})
}

// RefreshTokenRequest represents the refresh token request payload
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid refresh token request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Create auth refresh request
	authReq := &auth.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}

	// Refresh token
	loginResp, err := h.authService.RefreshToken(c.Request.Context(), authReq)
	if err != nil {
		h.logger.WithError(err).Error("Token refresh failed")
		response.Unauthorized(c, "Invalid refresh token")
		return
	}

	h.logger.WithField("user_id", loginResp.User.ID).Info("Token refreshed successfully")
	response.Success(c, loginResp)
}

// ForgotPasswordRequest represents the forgot password request payload
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ForgotPassword handles forgot password
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid forgot password request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Initiate password reset
	err := h.authService.ResetPassword(c.Request.Context(), req.Email)
	if err != nil {
		h.logger.WithError(err).WithField("email", req.Email).Error("Password reset initiation failed")
		// Don't reveal if email exists or not
	}

	h.logger.WithField("email", req.Email).Info("Password reset initiated")
	response.Success(c, gin.H{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// ResetPasswordRequest represents the reset password request payload
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ResetPassword handles reset password
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid reset password request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Confirm password reset
	err := h.authService.ConfirmPasswordReset(c.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		h.logger.WithError(err).Error("Password reset failed")
		response.BadRequest(c, "Password reset failed", err.Error())
		return
	}

	h.logger.Info("Password reset completed successfully")
	response.Success(c, gin.H{
		"message": "Password reset successfully",
	})
}

// Logout handles user logout
func (h *Handler) Logout(c *gin.Context) {
	// Get session ID from context (set by auth middleware)
	sessionIDValue, exists := c.Get("session_id")
	if !exists {
		h.logger.Error("Session ID not found in context")
		response.Unauthorized(c, "Invalid session")
		return
	}

	sessionID, ok := sessionIDValue.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid session ID type in context")
		response.InternalServerError(c, "Internal error")
		return
	}

	// Logout user
	err := h.authService.Logout(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("Logout failed")
		response.ErrorWithStatus(c, http.StatusInternalServerError, "logout_failed", "Logout failed", err.Error())
		return
	}

	h.logger.WithField("session_id", sessionID).Info("User logged out successfully")
	response.Success(c, gin.H{
		"message": "Logged out successfully",
	})
}

// GetProfile returns current user profile
func (h *Handler) GetProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		response.ErrorWithStatus(c, http.StatusUnauthorized, "unauthorized", "Unauthorized", "")
		return
	}

	userID, ok := userIDValue.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		response.ErrorWithStatus(c, http.StatusInternalServerError, "internal_error", "Internal error", "")
		return
	}

	// Get current user
	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user profile")
		response.ErrorWithStatus(c, http.StatusNotFound, "user_not_found", "User not found", err.Error())
		return
	}

	response.Success(c, user)
}

// UpdateProfileRequest represents the update profile request payload
type UpdateProfileRequest struct {
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,url"`
	Phone     *string `json:"phone,omitempty" validate:"omitempty,max=50"`
	Timezone  *string `json:"timezone,omitempty"`
	Language  *string `json:"language,omitempty" validate:"omitempty,len=2"`
}

// UpdateProfile updates user profile
func (h *Handler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid update profile request")
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid_request", "Invalid request payload", err.Error())
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.ErrorWithStatus(c, http.StatusUnauthorized, "unauthorized", "Unauthorized", "")
		return
	}

	userID, ok := userIDValue.(ulid.ULID)
	if !ok {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "internal_error", "Internal error", "")
		return
	}

	// Create user update request
	updateReq := &user.UpdateUserRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		AvatarURL: req.AvatarURL,
		Phone:     req.Phone,
		Timezone:  req.Timezone,
		Language:  req.Language,
	}

	// Update profile
	_, err := h.userService.UpdateUser(c.Request.Context(), userID, updateReq)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Profile update failed")
		response.ErrorWithStatus(c, http.StatusInternalServerError, "profile_update_failed", "Profile update failed", err.Error())
		return
	}

	h.logger.WithField("user_id", userID).Info("Profile updated successfully")
	response.Success(c, gin.H{
		"message": "Profile updated successfully",
	})
}

// ChangePasswordRequest represents the change password request payload
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// ChangePassword changes user password
func (h *Handler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid change password request")
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid_request", "Invalid request payload", err.Error())
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.ErrorWithStatus(c, http.StatusUnauthorized, "unauthorized", "Unauthorized", "")
		return
	}

	userID, ok := userIDValue.(ulid.ULID)
	if !ok {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "internal_error", "Internal error", "")
		return
	}

	// Change password
	err := h.userService.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Password change failed")
		response.ErrorWithStatus(c, http.StatusBadRequest, "password_change_failed", "Password change failed", err.Error())
		return
	}

	h.logger.WithField("user_id", userID).Info("Password changed successfully")
	response.Success(c, gin.H{
		"message": "Password changed successfully",
	})
}

// ValidateToken validates JWT tokens (for middleware)
func (h *Handler) ValidateToken(token string) (*auth.AuthContext, error) {
	return h.authService.ValidateAuthToken(context.Background(), token)
}

// ValidateAPIKey validates API keys (for middleware)
func (h *Handler) ValidateAPIKey(apiKey string) (*auth.AuthContext, error) {
	// TODO: Implement API key validation via auth service
	return nil, errors.New("API key validation not implemented")
}