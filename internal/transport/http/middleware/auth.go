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
	jwtService        auth.JWTService
	blacklistedTokens auth.BlacklistedTokenService
	roleService       auth.RoleService
	logger            *logrus.Logger
}

// NewAuthMiddleware creates a new stateless authentication middleware
func NewAuthMiddleware(
	jwtService auth.JWTService,
	blacklistedTokens auth.BlacklistedTokenService,
	roleService auth.RoleService,
	logger *logrus.Logger,
) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:        jwtService,
		blacklistedTokens: blacklistedTokens,
		roleService:       roleService,
		logger:            logger,
	}
}

// Context keys for storing authentication data in Gin context
const (
	AuthContextKey = "auth_context"
	UserIDKey      = "user_id"
	TokenClaimsKey = "token_claims"
)

// RequireAuth middleware ensures valid JWT token with stateless authentication
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

		// GDPR/SOC2 Compliance: Check user-wide timestamp blacklisting
		// This ensures ALL tokens issued before user revocation are blocked
		isUserBlacklisted, err := m.blacklistedTokens.IsUserBlacklistedAfterTimestamp(
			c.Request.Context(), claims.UserID, claims.IssuedAt)
		if err != nil {
			m.logger.WithError(err).WithFields(logrus.Fields{
				"user_id": claims.UserID,
				"iat":     claims.IssuedAt,
			}).Error("Failed to check user timestamp blacklist status")
			response.InternalServerError(c, "Authentication verification failed")
			c.Abort()
			return
		}

		if isUserBlacklisted {
			m.logger.WithFields(logrus.Fields{
				"user_id": claims.UserID,
				"jti":     claims.JWTID,
				"iat":     claims.IssuedAt,
			}).Warn("User token revoked - all sessions were revoked")
			response.Unauthorized(c, "All user sessions have been revoked")
			c.Abort()
			return
		}

		// Store clean authentication data in Gin context
		authContext := claims.GetUserContext()
		c.Set(AuthContextKey, authContext)
		c.Set(UserIDKey, claims.UserID)
		c.Set(TokenClaimsKey, claims)

		// Log successful authentication
		m.logger.WithFields(logrus.Fields{
			"user_id": claims.UserID,
			"jti":     claims.JWTID,
		}).Debug("Authentication successful")

		c.Next()
	})
}

// RequirePermission middleware ensures user has specific permission with dynamic resolution
func (m *AuthMiddleware) RequirePermission(resourceAction string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Get user ID from context
		userID, exists := GetUserID(c)
		if !exists {
			m.logger.Warn("Permission check attempted without authentication")
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		userIDParsed, err := ulid.Parse(userID)
		if err != nil {
			m.logger.WithError(err).Error("Invalid user ID format in context")
			response.InternalServerError(c, "Authentication error")
			c.Abort()
			return
		}

		// Resolve organization context from headers or URL
		orgID := ResolveOrganizationID(c)
		if orgID == nil {
			m.logger.WithField("user_id", userID).Warn("Organization context required but not found")
			response.BadRequest(c, "Organization context required", "Missing X-Org-ID header or organization in URL path")
			c.Abort()
			return
		}

		// Check permission dynamically using role service
		hasPermission, err := m.roleService.CheckPermissions(
			c.Request.Context(),
			userIDParsed,
			*orgID,
			[]string{resourceAction},
		)
		if err != nil {
			m.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":         userID,
				"organization_id": orgID,
				"permission":      resourceAction,
			}).Error("Failed to check user permissions")
			response.InternalServerError(c, "Permission verification failed")
			c.Abort()
			return
		}

		// Check if user has the required permission
		if !hasPermission.Results[resourceAction] {
			m.logger.WithFields(logrus.Fields{
				"user_id":         userID,
				"organization_id": orgID,
				"permission":      resourceAction,
			}).Warn("Insufficient permissions")
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequireAnyPermission middleware ensures user has at least one of the specified permissions
func (m *AuthMiddleware) RequireAnyPermission(resourceActions []string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Get user ID from context
		userID, exists := GetUserID(c)
		if !exists {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		userIDParsed, err := ulid.Parse(userID)
		if err != nil {
			m.logger.WithError(err).Error("Invalid user ID format in context")
			response.InternalServerError(c, "Authentication error")
			c.Abort()
			return
		}

		// Resolve organization context
		orgID := ResolveOrganizationID(c)
		if orgID == nil {
			response.BadRequest(c, "Organization context required", "Missing organization context")
			c.Abort()
			return
		}

		// Check if user has ANY of the permissions using role service
		hasPermission, err := m.roleService.CheckPermissions(
			c.Request.Context(),
			userIDParsed,
			*orgID,
			resourceActions,
		)
		if err != nil {
			m.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":         userID,
				"organization_id": orgID,
				"permissions":     resourceActions,
			}).Error("Failed to check user permissions")
			response.InternalServerError(c, "Permission verification failed")
			c.Abort()
			return
		}

		// Check if user has any of the required permissions
		hasAnyPermission := false
		for _, resourceAction := range resourceActions {
			if hasPermission.Results[resourceAction] {
				hasAnyPermission = true
				break
			}
		}

		if !hasAnyPermission {
			m.logger.WithFields(logrus.Fields{
				"user_id":         userID,
				"organization_id": orgID,
				"permissions":     resourceActions,
			}).Warn("Insufficient permissions - none of the required permissions found")
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequireAllPermissions middleware ensures user has ALL specified permissions
func (m *AuthMiddleware) RequireAllPermissions(resourceActions []string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Get user ID from context
		userID, exists := GetUserID(c)
		if !exists {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		userIDParsed, err := ulid.Parse(userID)
		if err != nil {
			m.logger.WithError(err).Error("Invalid user ID format in context")
			response.InternalServerError(c, "Authentication error")
			c.Abort()
			return
		}

		// Resolve context
		orgID := ResolveOrganizationID(c)
		if orgID == nil {
			response.BadRequest(c, "Organization context required", "Missing organization context")
			c.Abort()
			return
		}

		// Check all permissions individually
		for _, resourceAction := range resourceActions {
			hasPermission, err := m.roleService.CheckPermissions(
				c.Request.Context(),
				userIDParsed,
				*orgID,
				[]string{resourceAction},
			)
			if err != nil {
				m.logger.WithError(err).WithFields(logrus.Fields{
					"user_id":         userID,
					"organization_id": orgID,
					"permission":      resourceAction,
				}).Error("Failed to check user permissions")
				response.InternalServerError(c, "Permission verification failed")
				c.Abort()
				return
			}

			// Check if user has this specific permission
			if !hasPermission.Results[resourceAction] {
				m.logger.WithFields(logrus.Fields{
					"user_id":         userID,
					"organization_id": orgID,
					"permissions":     resourceActions,
					"failed_on":       resourceAction,
				}).Warn("Insufficient permissions - missing required permission")
				response.Forbidden(c, "Insufficient permissions")
				c.Abort()
				return
			}
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

		// Store clean auth context for valid token
		authContext := claims.GetUserContext()
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

	// Handle both ulid.ULID and string types for compatibility
	switch id := userID.(type) {
	case ulid.ULID:
		return id.String(), true
	case string:
		return id, true
	default:
		return "", false
	}
}

// GetUserIDULID retrieves user ID as ULID from Gin context
func GetUserIDULID(c *gin.Context) (ulid.ULID, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return ulid.ULID{}, false
	}

	id, ok := userID.(ulid.ULID)
	return id, ok
}

// Context Resolution Helper Functions

// ContextType represents the type of context to resolve
type ContextType string

const (
	ContextOrg ContextType = "org"
	ContextProject ContextType = "project"
	ContextEnv ContextType = "env"
)

// ContextResolver resolves context IDs from headers or URL parameters
type ContextResolver struct {
	OrganizationID *ulid.ULID
	ProjectID      *ulid.ULID
	EnvironmentID  *ulid.ULID
}

// ResolveContext resolves specified context types (variadic - any combination is optional)
func ResolveContext(c *gin.Context, contextTypes ...ContextType) *ContextResolver {
	resolver := &ContextResolver{}
	
	// Build a set for faster lookup
	typeSet := make(map[ContextType]bool)
	for _, ctxType := range contextTypes {
		typeSet[ctxType] = true
	}
	
	// If no types specified, resolve all (backward compatibility)
	if len(contextTypes) == 0 {
		typeSet[ContextOrg] = true
		typeSet[ContextProject] = true
		typeSet[ContextEnv] = true
	}
	
	// Resolve organization ID if requested
	if typeSet[ContextOrg] {
		// Try X-Org-ID header first
		if orgIDHeader := c.GetHeader("X-Org-ID"); orgIDHeader != "" {
			if orgID, err := ulid.Parse(orgIDHeader); err == nil {
				resolver.OrganizationID = &orgID
			}
		}
		// Try orgId URL parameter if header failed
		if resolver.OrganizationID == nil {
			if orgIDParam := c.Param("orgId"); orgIDParam != "" {
				if orgID, err := ulid.Parse(orgIDParam); err == nil {
					resolver.OrganizationID = &orgID
				}
			}
		}
	}
	
	// Resolve project ID if requested
	if typeSet[ContextProject] {
		// Try X-Project-ID header first
		if projectIDHeader := c.GetHeader("X-Project-ID"); projectIDHeader != "" {
			if projectID, err := ulid.Parse(projectIDHeader); err == nil {
				resolver.ProjectID = &projectID
			}
		}
		// Try projectId URL parameter if header failed
		if resolver.ProjectID == nil {
			if projectIDParam := c.Param("projectId"); projectIDParam != "" {
				if projectID, err := ulid.Parse(projectIDParam); err == nil {
					resolver.ProjectID = &projectID
				}
			}
		}
	}
	
	// Resolve environment ID if requested
	if typeSet[ContextEnv] {
		// Try X-Environment-ID header first
		if envIDHeader := c.GetHeader("X-Environment-ID"); envIDHeader != "" {
			if envID, err := ulid.Parse(envIDHeader); err == nil {
				resolver.EnvironmentID = &envID
			}
		}
		// Try envId URL parameter if header failed
		if resolver.EnvironmentID == nil {
			if envIDParam := c.Param("envId"); envIDParam != "" {
				if envID, err := ulid.Parse(envIDParam); err == nil {
					resolver.EnvironmentID = &envID
				}
			}
		}
	}
	
	return resolver
}

// Convenience functions for single context resolution
func ResolveOrganizationID(c *gin.Context) *ulid.ULID {
	return ResolveContext(c, ContextOrg).OrganizationID
}

func ResolveProjectID(c *gin.Context) *ulid.ULID {
	return ResolveContext(c, ContextProject).ProjectID
}

func ResolveEnvironmentID(c *gin.Context) *ulid.ULID {
	return ResolveContext(c, ContextEnv).EnvironmentID
}

