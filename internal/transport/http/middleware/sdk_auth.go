package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// SDKAuthMiddleware handles API key authentication for SDK routes
type SDKAuthMiddleware struct {
	apiKeyService auth.APIKeyService
	logger        *logrus.Logger
}

// NewSDKAuthMiddleware creates a new SDK authentication middleware
func NewSDKAuthMiddleware(
	apiKeyService auth.APIKeyService,
	logger *logrus.Logger,
) *SDKAuthMiddleware {
	return &SDKAuthMiddleware{
		apiKeyService: apiKeyService,
		logger:        logger,
	}
}

// Context keys for SDK authentication
const (
	SDKAuthContextKey = "sdk_auth_context"
	APIKeyIDKey       = "api_key_id"
	ProjectIDKey      = "project_id"
	EnvironmentKey    = "environment"
)

// RequireSDKAuth middleware validates API keys with project scoping for SDK routes
func (m *SDKAuthMiddleware) RequireSDKAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Extract API key from X-API-Key header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Fallback to Authorization header with Bearer format
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if apiKey == "" {
			m.logger.Warn("SDK request missing API key")
			response.Unauthorized(c, "API key required")
			c.Abort()
			return
		}

		// Extract project ID from X-Project-ID header (required for SDK auth)
		projectIDHeader := c.GetHeader("X-Project-ID")
		if projectIDHeader == "" {
			m.logger.Warn("SDK request missing project ID")
			response.BadRequest(c, "X-Project-ID header required", "Project ID must be provided for SDK requests")
			c.Abort()
			return
		}

		projectID, err := ulid.Parse(projectIDHeader)
		if err != nil {
			m.logger.WithError(err).WithField("project_id", projectIDHeader).Warn("Invalid project ID format")
			response.BadRequest(c, "Invalid project ID format", err.Error())
			c.Abort()
			return
		}

		// Extract optional environment from X-Environment header
		environment := c.GetHeader("X-Environment")

		// Validate API key with project scoping
		validateReq := &auth.ValidateAPIKeyRequest{
			APIKey:      apiKey,
			ProjectID:   projectID,
			Environment: environment,
		}

		validateResp, err := m.apiKeyService.ValidateAPIKeyWithProjectScoping(c.Request.Context(), validateReq)
		if err != nil {
			m.logger.WithError(err).WithFields(logrus.Fields{
				"project_id": projectID,
				"has_env":    environment != "",
			}).Warn("API key validation failed")
			response.InternalServerError(c, "Authentication verification failed")
			c.Abort()
			return
		}

		if !validateResp.Valid {
			m.logger.WithFields(logrus.Fields{
				"error_code":    validateResp.ErrorCode,
				"error_message": validateResp.ErrorMessage,
				"project_id":    projectID,
			}).Warn("Invalid API key for SDK request")

			// Map error codes to appropriate HTTP responses
			switch validateResp.ErrorCode {
			case "invalid_environment":
				response.BadRequest(c, "Invalid environment", validateResp.ErrorMessage)
			case "project_mismatch":
				response.Forbidden(c, "API key does not belong to the specified project")
			case "unauthorized", "invalid_api_key":
				response.Unauthorized(c, "Invalid API key")
			default:
				response.Unauthorized(c, validateResp.ErrorMessage)
			}
			c.Abort()
			return
		}

		// Store SDK authentication context in Gin context
		c.Set(SDKAuthContextKey, validateResp.AuthContext)
		c.Set(APIKeyIDKey, validateResp.KeyID)
		c.Set(ProjectIDKey, validateResp.ProjectID)
		c.Set(EnvironmentKey, validateResp.Environment)

		// Log successful SDK authentication
		m.logger.WithFields(logrus.Fields{
			"api_key_id": validateResp.KeyID,
			"project_id": validateResp.ProjectID,
			"environment": validateResp.Environment,
		}).Debug("SDK authentication successful")

		c.Next()
	})
}

// Helper functions to get SDK auth data from Gin context

// GetSDKAuthContext retrieves SDK authentication context from Gin context
func GetSDKAuthContext(c *gin.Context) (*auth.AuthContext, bool) {
	authContext, exists := c.Get(SDKAuthContextKey)
	if !exists {
		return nil, false
	}

	ctx, ok := authContext.(*auth.AuthContext)
	return ctx, ok
}

// GetAPIKeyID retrieves API key ID from Gin context
func GetAPIKeyID(c *gin.Context) (*ulid.ULID, bool) {
	keyID, exists := c.Get(APIKeyIDKey)
	if !exists {
		return nil, false
	}

	id, ok := keyID.(*ulid.ULID)
	return id, ok
}

// GetProjectID retrieves project ID from Gin context
func GetProjectID(c *gin.Context) (*ulid.ULID, bool) {
	projectID, exists := c.Get(ProjectIDKey)
	if !exists {
		return nil, false
	}

	id, ok := projectID.(*ulid.ULID)
	return id, ok
}

// GetEnvironment retrieves environment from Gin context
func GetEnvironment(c *gin.Context) (string, bool) {
	environment, exists := c.Get(EnvironmentKey)
	if !exists {
		return "", false
	}

	env, ok := environment.(string)
	return env, ok
}