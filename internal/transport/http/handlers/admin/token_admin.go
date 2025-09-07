package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/auth"
	"brokle/internal/transport/http/middleware"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// TokenAdminHandler handles administrative token management endpoints
type TokenAdminHandler struct {
	authService       auth.AuthService
	blacklistedTokens auth.BlacklistedTokenService
	logger            *logrus.Logger
}

// NewTokenAdminHandler creates a new token admin handler
func NewTokenAdminHandler(
	authService auth.AuthService,
	blacklistedTokens auth.BlacklistedTokenService,
	logger *logrus.Logger,
) *TokenAdminHandler {
	return &TokenAdminHandler{
		authService:       authService,
		blacklistedTokens: blacklistedTokens,
		logger:           logger,
	}
}

// RevokeTokenRequest represents the request to revoke a specific token
// @Description Request to revoke a specific access token
type RevokeTokenRequest struct {
	JTI    string `json:"jti" binding:"required" example:"01K4FHGHT3XX9WFM293QPZ5G9V" description:"JWT ID of the token to revoke"`
	Reason string `json:"reason" binding:"required,max=100" example:"security_incident" description:"Reason for token revocation"`
}

// RevokeUserTokensRequest represents the request to revoke all tokens for a user
// @Description Request to revoke all access tokens for a user
type RevokeUserTokensRequest struct {
	Reason string `json:"reason" binding:"required,max=100" example:"account_compromise" description:"Reason for bulk token revocation"`
}

// TokenStatsResponse represents token statistics
// @Description Token management statistics
type TokenStatsResponse struct {
	TotalBlacklisted    int64                    `json:"total_blacklisted" example:"1234" description:"Total number of blacklisted tokens"`
	BlacklistedToday    int64                    `json:"blacklisted_today" example:"45" description:"Tokens blacklisted today"`
	BlacklistedByReason map[string][]TokenByReason `json:"blacklisted_by_reason" description:"Breakdown of tokens by revocation reason"`
}

// TokenByReason represents tokens grouped by reason
type TokenByReason struct {
	Reason string                      `json:"reason" example:"security_incident"`
	Count  int                        `json:"count" example:"12"`
	Tokens []BlacklistedTokenResponse `json:"tokens,omitempty"`
}

// BlacklistedTokenResponse represents a blacklisted token
// @Description Blacklisted token information
type BlacklistedTokenResponse struct {
	JTI       string    `json:"jti" example:"01K4FHGHT3XX9WFM293QPZ5G9V" description:"JWT ID"`
	UserID    string    `json:"user_id" example:"01K4FHGHT3XX9WFM293QPZ5G9V" description:"User ID who owned the token"`
	Reason    string    `json:"reason" example:"security_incident" description:"Reason for revocation"`
	RevokedAt string    `json:"revoked_at" example:"2025-01-15T10:30:00Z" description:"When the token was revoked"`
	ExpiresAt string    `json:"expires_at" example:"2025-01-15T11:30:00Z" description:"When the token would have naturally expired"`
}

// RevokeToken revokes a specific access token immediately
// @Summary Revoke specific token
// @Description Immediately revoke a specific access token by JTI
// @Tags Admin, Token Management
// @Accept json
// @Produce json
// @Param request body RevokeTokenRequest true "Token revocation request"
// @Success 200 {object} response.SuccessResponse "Token revoked successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Insufficient permissions"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/admin/tokens/revoke [post]
func (h *TokenAdminHandler) RevokeToken(c *gin.Context) {
	var req RevokeTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid revoke token request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Get admin user from auth context
	authContext, exists := middleware.GetAuthContext(c)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	// Parse admin user ID for audit logging
	adminUserID, err := ulid.Parse(authContext.UserID.String())
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse admin user ID")
		response.InternalServerError(c, "Authentication error")
		return
	}

	// Check if token is already revoked
	isRevoked, err := h.authService.IsTokenRevoked(c.Request.Context(), req.JTI)
	if err != nil {
		h.logger.WithError(err).WithField("jti", req.JTI).Error("Failed to check token revocation status")
		response.InternalServerError(c, "Failed to check token status")
		return
	}

	if isRevoked {
		response.BadRequest(c, "Token is already revoked", "")
		return
	}

	// Get the blacklisted token to find the user ID (if it exists)
	// Since we're revoking by JTI, we need to find the user ID somehow
	// For now, we'll use the admin's user ID as a placeholder and update the implementation
	// In a production system, you might want to store JTI->UserID mapping
	err = h.authService.RevokeAccessToken(c.Request.Context(), req.JTI, adminUserID, req.Reason)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"jti":    req.JTI,
			"reason": req.Reason,
			"admin":  adminUserID,
		}).Error("Failed to revoke token")
		response.InternalServerError(c, "Failed to revoke token")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"jti":         req.JTI,
		"reason":      req.Reason,
		"admin_user":  adminUserID,
		"request_id":  c.GetString("request_id"),
	}).Info("Token revoked by admin")

	response.Success(c, gin.H{
		"message": "Token revoked successfully",
		"jti":     req.JTI,
		"reason":  req.Reason,
	})
}

// RevokeUserTokens revokes all active tokens for a specific user
// @Summary Revoke all user tokens
// @Description Immediately revoke all active access tokens for a specific user
// @Tags Admin, Token Management
// @Accept json
// @Produce json
// @Param userID path string true "User ID" example("01K4FHGHT3XX9WFM293QPZ5G9V")
// @Param request body RevokeUserTokensRequest true "User token revocation request"
// @Success 200 {object} response.SuccessResponse "User tokens revoked successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Insufficient permissions"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/admin/users/{userID}/tokens/revoke [post]
func (h *TokenAdminHandler) RevokeUserTokens(c *gin.Context) {
	userIDStr := c.Param("userID")
	if userIDStr == "" {
		response.BadRequest(c, "User ID is required", "")
		return
	}

	userID, err := ulid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid user ID format", err.Error())
		return
	}

	var req RevokeUserTokensRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid revoke user tokens request")
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	// Get admin user from auth context for audit logging
	authContext, exists := middleware.GetAuthContext(c)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	// Revoke all user tokens
	err = h.authService.RevokeUserAccessTokens(c.Request.Context(), userID, req.Reason)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"reason":  req.Reason,
			"admin":   authContext.UserID,
		}).Error("Failed to revoke user tokens")
		response.InternalServerError(c, "Failed to revoke user tokens")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"reason":      req.Reason,
		"admin_user":  authContext.UserID,
		"request_id":  c.GetString("request_id"),
	}).Info("All user tokens revoked by admin")

	response.Success(c, gin.H{
		"message": "All user tokens revoked successfully",
		"user_id": userID,
		"reason":  req.Reason,
	})
}

// ListBlacklistedTokens returns a paginated list of blacklisted tokens
// @Summary List blacklisted tokens
// @Description Get a paginated list of blacklisted tokens with filtering options
// @Tags Admin, Token Management
// @Produce json
// @Param limit query int false "Number of tokens to return (default: 50, max: 200)" example(50)
// @Param offset query int false "Number of tokens to skip (default: 0)" example(0)
// @Param user_id query string false "Filter by user ID" example("01K4FHGHT3XX9WFM293QPZ5G9V")
// @Param reason query string false "Filter by revocation reason" example("security_incident")
// @Success 200 {object} response.SuccessResponse "Blacklisted tokens list"
// @Failure 400 {object} response.ErrorResponse "Invalid query parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Insufficient permissions"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/admin/tokens/blacklisted [get]
func (h *TokenAdminHandler) ListBlacklistedTokens(c *gin.Context) {
	// Parse query parameters
	limit := 50 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			if parsedLimit > 200 {
				parsedLimit = 200 // max limit
			}
			limit = parsedLimit
		}
	}

	offset := 0 // default
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	userIDStr := c.Query("user_id")
	reason := c.Query("reason")

	var tokens []*auth.BlacklistedToken
	var err error

	// Filter by user ID if provided
	if userIDStr != "" {
		userID, parseErr := ulid.Parse(userIDStr)
		if parseErr != nil {
			response.BadRequest(c, "Invalid user ID format", parseErr.Error())
			return
		}

		tokens, err = h.blacklistedTokens.GetUserBlacklistedTokens(c.Request.Context(), userID, limit, offset)
	} else if reason != "" {
		// Filter by reason
		allTokens, reasonErr := h.blacklistedTokens.GetTokensByReason(c.Request.Context(), reason)
		if reasonErr != nil {
			err = reasonErr
		} else {
			// Apply pagination manually for reason-filtered results
			start := offset
			end := offset + limit
			if start >= len(allTokens) {
				tokens = []*auth.BlacklistedToken{}
			} else {
				if end > len(allTokens) {
					end = len(allTokens)
				}
				tokens = allTokens[start:end]
			}
		}
	} else {
		// TODO: Implement GetAllBlacklistedTokens with pagination in service
		// For now, we'll return an error asking for filters
		response.BadRequest(c, "Please specify user_id or reason filter for token listing", "")
		return
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to retrieve blacklisted tokens")
		response.InternalServerError(c, "Failed to retrieve tokens")
		return
	}

	// Convert to response format
	var responseTokens []BlacklistedTokenResponse
	for _, token := range tokens {
		responseTokens = append(responseTokens, BlacklistedTokenResponse{
			JTI:       token.JTI,
			UserID:    token.UserID.String(),
			Reason:    token.Reason,
			RevokedAt: token.RevokedAt.Format("2006-01-02T15:04:05Z"),
			ExpiresAt: token.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	response.Success(c, gin.H{
		"tokens": responseTokens,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(responseTokens),
		},
		"filters": gin.H{
			"user_id": userIDStr,
			"reason":  reason,
		},
	})
}

// GetTokenStats returns statistics about token management
// @Summary Get token statistics
// @Description Get comprehensive statistics about token usage and revocation
// @Tags Admin, Token Management
// @Produce json
// @Success 200 {object} response.SuccessResponse "Token statistics"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Insufficient permissions"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/admin/tokens/stats [get]
func (h *TokenAdminHandler) GetTokenStats(c *gin.Context) {
	// Get total blacklisted tokens count
	totalCount, err := h.blacklistedTokens.GetBlacklistedTokensCount(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get blacklisted tokens count")
		response.InternalServerError(c, "Failed to retrieve token statistics")
		return
	}

	// TODO: Implement GetBlacklistedTokensToday method in service
	// For now, we'll set it to 0
	todayCount := int64(0)

	// Get common revocation reasons with sample tokens
	commonReasons := []string{"logout", "security_incident", "suspicious_activity", "admin_revocation", "password_change"}
	reasonBreakdown := make(map[string][]TokenByReason)

	for _, reason := range commonReasons {
		tokens, reasonErr := h.blacklistedTokens.GetTokensByReason(c.Request.Context(), reason)
		if reasonErr != nil {
			h.logger.WithError(reasonErr).WithField("reason", reason).Warn("Failed to get tokens by reason")
			continue
		}

		// Convert to response format (limit to first 5 for stats)
		var sampleTokens []BlacklistedTokenResponse
		limit := 5
		if len(tokens) < limit {
			limit = len(tokens)
		}
		
		for i := 0; i < limit; i++ {
			token := tokens[i]
			sampleTokens = append(sampleTokens, BlacklistedTokenResponse{
				JTI:       token.JTI,
				UserID:    token.UserID.String(),
				Reason:    token.Reason,
				RevokedAt: token.RevokedAt.Format("2006-01-02T15:04:05Z"),
				ExpiresAt: token.ExpiresAt.Format("2006-01-02T15:04:05Z"),
			})
		}

		if len(tokens) > 0 {
			reasonBreakdown[reason] = []TokenByReason{
				{
					Reason: reason,
					Count:  len(tokens),
					Tokens: sampleTokens,
				},
			}
		}
	}

	stats := TokenStatsResponse{
		TotalBlacklisted:    totalCount,
		BlacklistedToday:    todayCount,
		BlacklistedByReason: reasonBreakdown,
	}

	response.Success(c, stats)
}