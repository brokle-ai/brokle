package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// ScopeMiddleware handles scope-based authorization
// This is the NEW middleware that replaces RequirePermission*
type ScopeMiddleware struct {
	scopeService auth.ScopeService
	logger       *logrus.Logger
}

// NewScopeMiddleware creates a new scope-based authorization middleware
func NewScopeMiddleware(
	scopeService auth.ScopeService,
	logger *logrus.Logger,
) *ScopeMiddleware {
	return &ScopeMiddleware{
		scopeService: scopeService,
		logger:       logger,
	}
}

// RequireScope middleware ensures user has a specific scope in the current context
//
// This middleware automatically resolves organization and project context from:
// - Headers: X-Org-ID, X-Project-ID
// - URL params: orgId, projectId
//
// Scope Resolution:
// - Organization-level scope (e.g., "members:invite") → requires org context
// - Project-level scope (e.g., "traces:delete") → requires org + project context
//
// Usage:
//
//	router.POST("/members", authMiddleware.RequireAuth(), scopeMiddleware.RequireScope("members:invite"), handler.InviteMember)
//	router.DELETE("/traces/:id", authMiddleware.RequireAuth(), scopeMiddleware.RequireScope("traces:delete"), handler.DeleteTrace)
func (m *ScopeMiddleware) RequireScope(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from auth context (set by RequireAuth middleware)
		userID, exists := GetUserIDULID(c)
		if !exists {
			m.logger.Warn("Scope check attempted without authentication")
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		// Resolve organization and project context from request
		resolver := ResolveContext(c, ContextOrg, ContextProject)

		// Check if user has the required scope
		hasScope, err := m.scopeService.HasScope(
			c.Request.Context(),
			userID,
			scope,
			resolver.OrganizationID,
			resolver.ProjectID,
		)

		if err != nil {
			m.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":    userID,
				"scope":      scope,
				"org_id":     resolver.OrganizationID,
				"project_id": resolver.ProjectID,
			}).Error("Failed to check user scope")
			response.InternalServerError(c, "Scope verification failed")
			c.Abort()
			return
		}

		if !hasScope {
			m.logger.WithFields(logrus.Fields{
				"user_id":    userID,
				"scope":      scope,
				"org_id":     resolver.OrganizationID,
				"project_id": resolver.ProjectID,
			}).Warn("Insufficient scopes")
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		// Store scope context in Gin context for handlers
		scopeContext := &ScopeContext{
			UserID:         userID,
			OrganizationID: resolver.OrganizationID,
			ProjectID:      resolver.ProjectID,
			Scopes:         []string{scope}, // At least this scope is guaranteed
		}
		c.Set(ScopeContextKey, scopeContext)

		m.logger.WithFields(logrus.Fields{
			"user_id":    userID,
			"scope":      scope,
			"org_id":     resolver.OrganizationID,
			"project_id": resolver.ProjectID,
		}).Debug("Scope check passed")

		c.Next()
	}
}

// RequireAnyScope middleware ensures user has at least one of the specified scopes
//
// Useful for endpoints that accept multiple permission levels, e.g.:
//
//	scopeMiddleware.RequireAnyScope([]string{"billing:manage", "billing:admin"})
func (m *ScopeMiddleware) RequireAnyScope(scopes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserIDULID(c)
		if !exists {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		resolver := ResolveContext(c, ContextOrg, ContextProject)

		hasAny, err := m.scopeService.HasAnyScope(
			c.Request.Context(),
			userID,
			scopes,
			resolver.OrganizationID,
			resolver.ProjectID,
		)

		if err != nil {
			m.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":    userID,
				"scopes":     scopes,
				"org_id":     resolver.OrganizationID,
				"project_id": resolver.ProjectID,
			}).Error("Failed to check user scopes")
			response.InternalServerError(c, "Scope verification failed")
			c.Abort()
			return
		}

		if !hasAny {
			m.logger.WithFields(logrus.Fields{
				"user_id":    userID,
				"scopes":     scopes,
				"org_id":     resolver.OrganizationID,
				"project_id": resolver.ProjectID,
			}).Warn("Insufficient scopes - none of the required scopes found")
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAllScopes middleware ensures user has ALL of the specified scopes
//
// Useful for endpoints that require multiple permissions, e.g.:
//
//	scopeMiddleware.RequireAllScopes([]string{"traces:read", "analytics:export"})
func (m *ScopeMiddleware) RequireAllScopes(scopes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserIDULID(c)
		if !exists {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		resolver := ResolveContext(c, ContextOrg, ContextProject)

		hasAll, err := m.scopeService.HasAllScopes(
			c.Request.Context(),
			userID,
			scopes,
			resolver.OrganizationID,
			resolver.ProjectID,
		)

		if err != nil {
			m.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":    userID,
				"scopes":     scopes,
				"org_id":     resolver.OrganizationID,
				"project_id": resolver.ProjectID,
			}).Error("Failed to check user scopes")
			response.InternalServerError(c, "Scope verification failed")
			c.Abort()
			return
		}

		if !hasAll {
			m.logger.WithFields(logrus.Fields{
				"user_id":    userID,
				"scopes":     scopes,
				"org_id":     resolver.OrganizationID,
				"project_id": resolver.ProjectID,
			}).Warn("Insufficient scopes - missing required scopes")
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ScopeContext holds resolved scope context for a request
type ScopeContext struct {
	UserID         ulid.ULID
	OrganizationID *ulid.ULID
	ProjectID      *ulid.ULID
	Scopes         []string // User's effective scopes in this context
}

// Context key for storing scope context
const ScopeContextKey = "scope_context"

// GetScopeContext retrieves scope context from Gin context (for handlers)
func GetScopeContext(c *gin.Context) (*ScopeContext, bool) {
	scopeCtx, exists := c.Get(ScopeContextKey)
	if !exists {
		return nil, false
	}

	ctx, ok := scopeCtx.(*ScopeContext)
	return ctx, ok
}
