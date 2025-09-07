package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// AuthMiddleware handles JWT token authentication and authorization
type AuthMiddleware struct {
	jwtService            auth.JWTService
	blacklistedTokens     auth.BlacklistedTokenService
	sessionService        auth.SessionService
	logger                *logrus.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(
	jwtService auth.JWTService,
	blacklistedTokens auth.BlacklistedTokenService,
	sessionService auth.SessionService,
	logger *logrus.Logger,
) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:        jwtService,
		blacklistedTokens: blacklistedTokens,
		sessionService:    sessionService,
		logger:           logger,
	}
}

// NewAuthMiddlewareSimple creates a new authentication middleware without session service
func NewAuthMiddlewareSimple(
	jwtService auth.JWTService,
	blacklistedTokens auth.BlacklistedTokenService,
	logger *logrus.Logger,
) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:        jwtService,
		blacklistedTokens: blacklistedTokens,
		sessionService:    nil, // Optional session service
		logger:           logger,
	}
}

// AuthContext keys for storing authentication data in Gin context
const (
	AuthContextKey     = "auth_context"
	UserIDKey          = "user_id"
	OrganizationIDKey  = "organization_id"
	PermissionsKey     = "permissions"
	SessionIDKey       = "session_id"
	TokenClaimsKey     = "token_claims"
)

// RequireAuth middleware ensures valid JWT token and active session
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Extract token from Authorization header
		token, err := m.extractToken(c)
		if err != nil {
			m.logger.WithError(err).Warn("Failed to extract authentication token")
			response.Unauthorized(c, "Authentication token required")
			c.Abort()
			return
		}

		// Validate JWT token structure and signature
		claims, err := m.jwtService.ValidateAccessToken(c.Request.Context(), token)
		if err != nil {
			m.logger.WithError(err).WithField("token_prefix", token[:min(len(token), 10)]).
				Warn("Invalid JWT token")
			response.Unauthorized(c, "Invalid authentication token")
			c.Abort()
			return
		}

		// Check if token is blacklisted (immediate revocation check)
		isBlacklisted, err := m.blacklistedTokens.IsTokenBlacklisted(c.Request.Context(), claims.JWTID)
		if err != nil {
			m.logger.WithError(err).WithField("jti", claims.JWTID).
				Error("Failed to check token blacklist status")
			response.InternalServerError(c, "Authentication verification failed")
			c.Abort()
			return
		}

		if isBlacklisted {
			m.logger.WithField("jti", claims.JWTID).WithField("user_id", claims.UserID).
				Warn("Blacklisted token attempted access")
			response.Unauthorized(c, "Authentication token has been revoked")
			c.Abort()
			return
		}

		// Validate session if JTI is present and session service is available (optional but recommended)
		var sessionID *string
		if claims.JWTID != "" && m.sessionService != nil {
			session, err := m.sessionService.GetSessionByToken(c.Request.Context(), token)
			if err != nil {
				m.logger.WithError(err).WithField("jti", claims.JWTID).
					Debug("Session validation failed - token may be valid but session inactive")
				// Don't block access for session validation failure - token is still valid
			} else {
				sessionIDStr := session.ID.String()
				sessionID = &sessionIDStr
				
				// Mark session as used (update last_used_at)
				m.sessionService.MarkSessionAsUsed(c.Request.Context(), session.ID)
			}
		}

		// Convert sessionID string to ULID if present
		var sessionULID *ulid.ULID
		if sessionID != nil {
			if parsedID, err := ulid.Parse(*sessionID); err == nil {
				sessionULID = &parsedID
			}
		}

		// Create authentication context
		authContext := &auth.AuthContext{
			UserID:         claims.UserID,
			OrganizationID: claims.OrganizationID,
			Permissions:    claims.Permissions,
			SessionID:      sessionULID,
		}

		// Store authentication data in Gin context
		c.Set(AuthContextKey, authContext)
		c.Set(UserIDKey, claims.UserID)
		c.Set(TokenClaimsKey, claims)
		
		if claims.OrganizationID != nil {
			c.Set(OrganizationIDKey, *claims.OrganizationID)
		}
		
		if len(claims.Permissions) > 0 {
			c.Set(PermissionsKey, claims.Permissions)
		}
		
		if sessionID != nil {
			c.Set(SessionIDKey, *sessionID)
		}

		// Log successful authentication
		m.logger.WithFields(logrus.Fields{
			"user_id":         claims.UserID,
			"organization_id": claims.OrganizationID,
			"jti":            claims.JWTID,
			"session_id":     sessionID,
		}).Debug("Authentication successful")

		c.Next()
	})
}

// RequirePermission middleware ensures user has specific permission
func (m *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Get authentication context
		authContext, exists := c.Get(AuthContextKey)
		if !exists {
			m.logger.Warn("Permission check attempted without authentication context")
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		ctx, ok := authContext.(*auth.AuthContext)
		if !ok {
			m.logger.Error("Invalid authentication context type")
			response.InternalServerError(c, "Authentication context error")
			c.Abort()
			return
		}

		// Check if user has required permission
		hasPermission := false
		for _, perm := range ctx.Permissions {
			if perm == permission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			m.logger.WithFields(logrus.Fields{
				"user_id":            ctx.UserID,
				"required_permission": permission,
				"user_permissions":   ctx.Permissions,
			}).Warn("Insufficient permissions")
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequireAnyPermission middleware ensures user has at least one of the specified permissions
func (m *AuthMiddleware) RequireAnyPermission(permissions []string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Get authentication context
		authContext, exists := c.Get(AuthContextKey)
		if !exists {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		ctx, ok := authContext.(*auth.AuthContext)
		if !ok {
			response.InternalServerError(c, "Authentication context error")
			c.Abort()
			return
		}

		// Check if user has any of the required permissions
		hasPermission := false
		for _, userPerm := range ctx.Permissions {
			for _, requiredPerm := range permissions {
				if userPerm == requiredPerm {
					hasPermission = true
					break
				}
			}
			if hasPermission {
				break
			}
		}

		if !hasPermission {
			m.logger.WithFields(logrus.Fields{
				"user_id":             ctx.UserID,
				"required_permissions": permissions,
				"user_permissions":    ctx.Permissions,
			}).Warn("Insufficient permissions")
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	})
}

// OptionalAuth middleware extracts auth info if present but doesn't require it
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Try to extract token
		token, err := m.extractToken(c)
		if err != nil {
			// No token present, continue without auth context
			c.Next()
			return
		}

		// Validate token if present
		claims, err := m.jwtService.ValidateAccessToken(c.Request.Context(), token)
		if err != nil {
			// Invalid token, continue without auth context
			c.Next()
			return
		}

		// Check blacklist
		isBlacklisted, err := m.blacklistedTokens.IsTokenBlacklisted(c.Request.Context(), claims.JWTID)
		if err != nil || isBlacklisted {
			// Blacklisted or error, continue without auth context
			c.Next()
			return
		}

		// Create auth context for valid token
		authContext := &auth.AuthContext{
			UserID:         claims.UserID,
			OrganizationID: claims.OrganizationID,
			Permissions:    claims.Permissions,
		}

		c.Set(AuthContextKey, authContext)
		c.Set(UserIDKey, claims.UserID)
		c.Set(TokenClaimsKey, claims)

		c.Next()
	})
}

// extractToken extracts JWT token from Authorization header
func (m *AuthMiddleware) extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	// Check for Bearer token format
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", errors.New("invalid authorization header format")
	}

	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if token == "" {
		return "", errors.New("token missing in authorization header")
	}

	return token, nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper functions to get auth data from Gin context

// GetAuthContext retrieves authentication context from Gin context
func GetAuthContext(c *gin.Context) (*auth.AuthContext, bool) {
	authContext, exists := c.Get(AuthContextKey)
	if !exists {
		return nil, false
	}

	ctx, ok := authContext.(*auth.AuthContext)
	return ctx, ok
}

// GetUserID retrieves user ID from Gin context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return "", false
	}

	id, ok := userID.(string)
	return id, ok
}

// GetOrganizationID retrieves organization ID from Gin context
func GetOrganizationID(c *gin.Context) (string, bool) {
	orgID, exists := c.Get(OrganizationIDKey)
	if !exists {
		return "", false
	}

	id, ok := orgID.(string)
	return id, ok
}

// GetPermissions retrieves permissions from Gin context
func GetPermissions(c *gin.Context) ([]string, bool) {
	permissions, exists := c.Get(PermissionsKey)
	if !exists {
		return nil, false
	}

	perms, ok := permissions.([]string)
	return perms, ok
}

// RequireRole middleware enforces role-based access control
// This is a simplified implementation - in production you'd have proper role management
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get auth context from JWT middleware
		authContext, exists := GetAuthContext(c)
		if !exists {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		// For admin role, check if user has admin permissions
		// In a full implementation, this would check against user roles in the database
		if requiredRole == "admin" {
			// Check for admin-level permissions
			hasAdminPermission := false
			for _, perm := range authContext.Permissions {
				if perm == "admin:*" || perm == "token:revoke" || perm == "system:admin" {
					hasAdminPermission = true
					break
				}
			}

			if !hasAdminPermission {
				response.Forbidden(c, "Admin access required")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}