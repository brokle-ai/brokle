package auth

import (
	"context"
	"net/http"
	"strings"

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
	config        *config.Config
	logger        *logrus.Logger
	authService   auth.AuthService
	apiKeyService auth.APIKeyService
	userService   user.UserService
}

// NewHandler creates a new auth handler
func NewHandler(config *config.Config, logger *logrus.Logger, authService auth.AuthService, apiKeyService auth.APIKeyService, userService user.UserService) *Handler {
	return &Handler{
		config:        config,
		logger:        logger,
		authService:   authService,
		apiKeyService: apiKeyService,
		userService:   userService,
	}
}

// LoginRequest represents the login request payload
// @Description User login credentials
type LoginRequest struct {
	Email      string                 `json:"email" binding:"required,email" example:"user@example.com" description:"User email address"`
	Password   string                 `json:"password" binding:"required" example:"password123" description:"User password (minimum 8 characters)"`
	DeviceInfo map[string]interface{} `json:"device_info,omitempty" description:"Device information for session tracking"`
}

// Login handles user login
// @Summary User login
// @Description Authenticate user with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} response.SuccessResponse "Login successful"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Invalid credentials"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid login request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Create auth login request
	authReq := &auth.LoginRequest{
		Email:      req.Email,
		Password:   req.Password,
		DeviceInfo: req.DeviceInfo,
	}

	// Attempt login
	loginResp, err := h.authService.Login(c.Request.Context(), authReq)
	if err != nil {
		h.logger.WithError(err).WithField("email", req.Email).Error("Login failed")
		response.Error(c, err)
		return
	}

	h.logger.Info("User logged in successfully")
	response.Success(c, loginResp)
}

// RegisterRequest represents the registration request payload
// @Description User registration information
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email" example:"user@example.com" description:"User email address"`
	FirstName string `json:"first_name" binding:"required,min=1,max=100" example:"John" description:"User first name"`
	LastName  string `json:"last_name" binding:"required,min=1,max=100" example:"Doe" description:"User last name"`
	Password  string `json:"password" binding:"required,min=8" example:"password123" description:"User password (minimum 8 characters)"`
	Timezone  string `json:"timezone,omitempty" example:"UTC" description:"User timezone (optional)"`
	Language  string `json:"language,omitempty" example:"en" description:"User language preference (optional)"`
}

// Signup handles user registration
// @Summary User registration
// @Description Register a new user account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration information"
// @Success 200 {object} response.SuccessResponse "Registration successful"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 409 {object} response.ErrorResponse "Email already exists"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/signup [post]
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

	// Register user and auto-login (returns tokens)
	loginResp, err := h.authService.Register(c.Request.Context(), authReq)
	if err != nil {
		h.logger.WithError(err).WithField("email", req.Email).Error("Registration failed")
		response.Error(c, err)
		return
	}

	h.logger.WithField("email", req.Email).Info("User registered and logged in successfully")
	response.Success(c, loginResp)
}

// RefreshTokenRequest represents the refresh token request payload
// @Description Refresh token credentials
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." description:"Valid refresh token"`
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Get a new access token using a valid refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} response.SuccessResponse "Token refresh successful"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Invalid refresh token"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/refresh [post]
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
		response.Error(c, err)
		return
	}

	h.logger.Info("Token refreshed successfully")
	response.Success(c, loginResp)
}

// ForgotPasswordRequest represents the forgot password request payload
// @Description Email for password reset
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com" description:"Email address for password reset"`
}

// ForgotPassword handles forgot password
// @Summary Request password reset
// @Description Initiate password reset process by sending reset email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "Email for password reset"
// @Success 200 {object} response.MessageResponse "Reset email sent if account exists"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/forgot-password [post]
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
// @Description Reset password with token
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required" example:"reset_token_123" description:"Password reset token from email"`
	NewPassword string `json:"new_password" binding:"required,min=8" example:"newpassword123" description:"New password (minimum 8 characters)"`
}

// ResetPassword handles reset password
// @Summary Reset password
// @Description Complete password reset using token from email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Reset password information"
// @Success 200 {object} response.MessageResponse "Password reset successful"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload or expired token"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/reset-password [post]
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
// @Summary User logout
// @Description Logout user and invalidate session
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.MessageResponse "Logout successful"
// @Failure 401 {object} response.ErrorResponse "Invalid session"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	// Get token claims from context (set by auth middleware)
	claimsValue, exists := c.Get("token_claims")
	if !exists {
		h.logger.Error("Token claims not found in context")
		response.Unauthorized(c, "Authentication required")
		return
	}

	claims, ok := claimsValue.(*auth.JWTClaims)
	if !ok {
		h.logger.Error("Invalid token claims type in context")
		response.InternalServerError(c, "Internal error")
		return
	}

	// Logout user by blacklisting current access token JTI
	err := h.authService.Logout(c.Request.Context(), claims.JWTID, claims.UserID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"jti":     claims.JWTID,
			"user_id": claims.UserID,
		}).Error("Logout failed")
		response.ErrorWithStatus(c, http.StatusInternalServerError, "logout_failed", "Logout failed", err.Error())
		return
	}

	h.logger.WithFields(logrus.Fields{
		"jti":     claims.JWTID,
		"user_id": claims.UserID,
	}).Info("User logged out successfully")
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
// @Description User profile update information
type UpdateProfileRequest struct {
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100" example:"John" description:"User first name"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100" example:"Doe" description:"User last name"`
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,url" example:"https://example.com/avatar.jpg" description:"Profile avatar URL"`
	Phone     *string `json:"phone,omitempty" validate:"omitempty,max=50" example:"+1234567890" description:"User phone number"`
	Timezone  *string `json:"timezone,omitempty" example:"UTC" description:"User timezone"`
	Language  *string `json:"language,omitempty" validate:"omitempty,len=2" example:"en" description:"User language preference (ISO 639-1 code)"`
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
// @Description Password change information
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"currentpass123" description:"Current password"`
	NewPassword     string `json:"new_password" binding:"required,min=8" example:"newpass123" description:"New password (minimum 8 characters)"`
}

// ChangePassword changes user password
// @Summary Change password
// @Description Change user password with current password verification
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "Password change information"
// @Success 200 {object} response.MessageResponse "Password changed successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request or wrong current password"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/change-password [post]
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
	ctx := context.Background()

	// Validate the API key using the APIKeyService
	key, err := h.apiKeyService.ValidateAPIKey(ctx, apiKey)
	if err != nil {
		h.logger.WithError(err).WithField("key_id", extractKeyID(apiKey)).
			Warn("API key validation failed")
		return nil, err
	}

	// Create AuthContext from the validated key
	authContext := key.AuthContext

	// Log successful validation (without the actual key)
	h.logger.WithFields(map[string]interface{}{
		"user_id":    key.AuthContext.UserID,
		"api_key_id": key.AuthContext.APIKeyID,
		"project_id": key.ProjectID,
	}).Debug("API key validation successful")

	return authContext, nil
}

// ValidateAPIKeyHandler validates self-contained API keys (industry standard)
// @Summary Validate API key
// @Description Validates a self-contained API key and extracts project information automatically
// @Tags Authentication
// @Accept json
// @Produce json
// @Param X-API-Key header string false "API key (format: bk_proj_{project_id}_{secret})"
// @Param Authorization header string false "Bearer token format: Bearer {api_key}"
// @Success 200 {object} response.SuccessResponse "API key validation successful"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 401 {object} response.ErrorResponse "Invalid, inactive, or expired API key"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /v1/auth/validate-key [post]
func (h *Handler) ValidateAPIKeyHandler(c *gin.Context) {
	// Extract API key from X-API-Key header or Authorization Bearer
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		// Fallback to Authorization header with Bearer format
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			apiKey = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if apiKey == "" {
		h.logger.Warn("API key validation request missing API key")
		response.BadRequest(c, "Missing API key", "API key must be provided via X-API-Key header or Authorization Bearer token")
		return
	}

	// Validate the self-contained API key
	resp, err := h.apiKeyService.ValidateAPIKey(c.Request.Context(), apiKey)
	if err != nil {
		h.logger.WithError(err).WithField("key_id", extractKeyID(apiKey)).
			Warn("API key validation failed")
		response.Error(c, err) // Properly propagate AppError status codes (401, etc.)
		return
	}

	// Log successful validation (without the actual key)
	h.logger.WithFields(logrus.Fields{
		"user_id":    resp.AuthContext.UserID,
		"api_key_id": resp.AuthContext.APIKeyID,
		"project_id": resp.ProjectID,
	}).Info("API key validation successful")

	response.Success(c, resp)
}

// extractKeyID safely extracts the key_id portion (bk_proj_{project_id}) for logging
func extractKeyID(apiKey string) string {
	parts := strings.Split(apiKey, "_")
	if len(parts) >= 3 {
		// Return bk_proj_{project_id} portion
		return strings.Join(parts[:3], "_")
	}
	// Fallback for invalid format
	if len(apiKey) < 8 {
		return apiKey
	}
	return apiKey[:8]
}

// ListSessionsRequest represents request for listing user sessions
type ListSessionsRequest struct {
	Page     int  `form:"page,default=1" example:"1" description:"Page number for pagination"`
	PageSize int  `form:"page_size,default=10" example:"10" description:"Number of sessions per page"`
	Active   bool `form:"active,default=false" example:"false" description:"Filter for active sessions only"`
}

// ListSessions lists all user sessions
// @Summary List user sessions
// @Description Get paginated list of user sessions with device info
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param active query bool false "Active sessions only" default(false)
// @Success 200 {object} response.SuccessResponse "Sessions retrieved successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/sessions [get]
func (h *Handler) ListSessions(c *gin.Context) {
	var req ListSessionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Error("Invalid list sessions request")
		response.BadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "")
		return
	}

	userID, ok := userIDValue.(ulid.ULID)
	if !ok {
		response.InternalServerError(c, "")
		return
	}

	// Get user sessions (using GetUserSessions method)
	sessions, err := h.authService.GetUserSessions(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to list sessions")
		response.InternalServerError(c, "Failed to retrieve sessions")
		return
	}

	// Filter active sessions if requested
	var filteredSessions []*auth.UserSession
	if req.Active {
		for _, session := range sessions {
			if session.IsValid() {
				filteredSessions = append(filteredSessions, session)
			}
		}
	} else {
		filteredSessions = sessions
	}

	// Manual pagination
	startIdx := (req.Page - 1) * req.PageSize
	endIdx := startIdx + req.PageSize
	total := int64(len(filteredSessions))

	if startIdx >= len(filteredSessions) {
		filteredSessions = []*auth.UserSession{}
	} else {
		if endIdx > len(filteredSessions) {
			endIdx = len(filteredSessions)
		}
		filteredSessions = filteredSessions[startIdx:endIdx]
	}

	// Create pagination metadata
	pagination := response.NewPagination(req.Page, req.PageSize, total)

	h.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"count":   len(filteredSessions),
		"total":   total,
	}).Info("Sessions listed successfully")

	response.SuccessWithPagination(c, filteredSessions, pagination)
}

// GetSessionRequest represents request for getting session by ID
type GetSessionRequest struct {
	SessionID ulid.ULID `uri:"session_id" binding:"required" example:"01FXYZ123456789ABCDEFGHIJK0" description:"Session ID" swaggertype:"string"`
}

// GetSession gets a specific user session by ID
// @Summary Get user session
// @Description Get details of a specific user session
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param session_id path string true "Session ID"
// @Success 200 {object} response.SuccessResponse "Session retrieved successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "Session not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/sessions/{session_id} [get]
func (h *Handler) GetSession(c *gin.Context) {
	var req GetSessionRequest
	if err := c.ShouldBindUri(&req); err != nil {
		h.logger.WithError(err).Error("Invalid get session request")
		response.BadRequest(c, "Invalid session ID", err.Error())
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "")
		return
	}

	userID, ok := userIDValue.(ulid.ULID)
	if !ok {
		response.InternalServerError(c, "")
		return
	}

	// Get all user sessions first, then filter by session ID
	sessions, err := h.authService.GetUserSessions(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":    userID,
			"session_id": req.SessionID,
		}).Error("Failed to get sessions")
		response.InternalServerError(c, "Failed to retrieve session")
		return
	}

	// Find the specific session
	var session *auth.UserSession
	for _, s := range sessions {
		if s.ID == req.SessionID {
			session = s
			break
		}
	}

	if session == nil {
		h.logger.WithFields(logrus.Fields{
			"user_id":    userID,
			"session_id": req.SessionID,
		}).Warn("Session not found")
		response.NotFound(c, "Session")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"session_id": req.SessionID,
	}).Info("Session retrieved successfully")

	response.Success(c, session)
}

// RevokeSessionRequest represents request for revoking a session
type RevokeSessionRequest struct {
	SessionID ulid.ULID `uri:"session_id" binding:"required" example:"01FXYZ123456789ABCDEFGHIJK0" description:"Session ID to revoke" swaggertype:"string"`
}

// RevokeSession revokes a specific user session
// @Summary Revoke user session
// @Description Revoke a specific user session (logout from specific device)
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param session_id path string true "Session ID"
// @Success 200 {object} response.MessageResponse "Session revoked successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "Session not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/sessions/{session_id}/revoke [post]
func (h *Handler) RevokeSession(c *gin.Context) {
	var req RevokeSessionRequest
	if err := c.ShouldBindUri(&req); err != nil {
		h.logger.WithError(err).Error("Invalid revoke session request")
		response.BadRequest(c, "Invalid session ID", err.Error())
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "")
		return
	}

	userID, ok := userIDValue.(ulid.ULID)
	if !ok {
		response.InternalServerError(c, "")
		return
	}

	// Revoke session
	err := h.authService.RevokeSession(c.Request.Context(), userID, req.SessionID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":    userID,
			"session_id": req.SessionID,
		}).Error("Failed to revoke session")
		response.NotFound(c, "Session")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"session_id": req.SessionID,
	}).Info("Session revoked successfully")

	response.Success(c, gin.H{
		"message": "Session revoked successfully",
	})
}

// RevokeAllSessionsRequest represents request for revoking all user sessions
type RevokeAllSessionsRequest struct {
	// Note: This struct is kept for future extensibility but currently has no fields
}

// RevokeAllSessions revokes all user sessions
// @Summary Revoke all user sessions
// @Description Revoke all user sessions (logout from all devices). This will invalidate ALL active sessions for the user.
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body RevokeAllSessionsRequest false "Request body (currently unused but kept for future extensibility)"
// @Success 200 {object} response.MessageResponse "All sessions revoked successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/sessions/revoke-all [post]
func (h *Handler) RevokeAllSessions(c *gin.Context) {
	var req RevokeAllSessionsRequest
	// Don't require body, use defaults
	c.ShouldBindJSON(&req)

	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "")
		return
	}

	userID, ok := userIDValue.(ulid.ULID)
	if !ok {
		response.InternalServerError(c, "")
		return
	}

	// Get current sessions count before revoking
	sessions, err := h.authService.GetUserSessions(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get sessions for count")
		response.InternalServerError(c, "Failed to revoke sessions")
		return
	}

	count := 0
	for _, session := range sessions {
		if session.IsValid() {
			count++
		}
	}

	// Revoke all sessions
	err = h.authService.RevokeAllSessions(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to revoke all sessions")
		response.InternalServerError(c, "Failed to revoke sessions")
		return
	}

	// GDPR/SOC2 Compliance: Create user-wide timestamp blacklist to immediately block ALL tokens
	// This ensures complete compliance - any token issued before this timestamp is immediately invalid
	err = h.authService.RevokeUserAccessTokens(c.Request.Context(), userID, "user_requested_revoke_all_sessions")
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to create user-wide token blacklist")
		// Log error but don't fail the request since sessions were already revoked
		// This maintains partial security even if timestamp blacklisting fails
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":       userID,
		"revoked_count": count,
	}).Info("All sessions and access tokens revoked successfully")

	response.Success(c, gin.H{
		"message": "All sessions revoked successfully",
		"count":   count,
	})
}