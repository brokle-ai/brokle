package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"brokle/pkg/response"
)

// CSRFMiddleware handles CSRF token validation using double-submit cookie pattern
type CSRFMiddleware struct {
	logger *slog.Logger
}

// NewCSRFMiddleware creates a new CSRF validation middleware
func NewCSRFMiddleware(logger *slog.Logger) *CSRFMiddleware {
	return &CSRFMiddleware{
		logger: logger,
	}
}

// ValidateCSRF validates CSRF token using double-submit cookie pattern
// Applies to all non-idempotent methods (POST, PUT, PATCH, DELETE)
// Skips safe methods (GET, HEAD, OPTIONS)
func (m *CSRFMiddleware) ValidateCSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF validation for safe methods (idempotent operations)
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Get CSRF token from cookie
		cookieToken, err := c.Cookie("csrf_token")
		if err != nil || cookieToken == "" {
			m.logger.Warn("CSRF validation failed: cookie missing", "method", c.Request.Method, "path", c.Request.URL.Path)
			response.ErrorWithStatus(c, http.StatusForbidden, "CSRF_TOKEN_MISSING", "CSRF token missing in cookie", "")
			c.Abort()
			return
		}

		// Get CSRF token from request header
		headerToken := c.GetHeader("X-CSRF-Token")
		if headerToken == "" {
			m.logger.Warn("CSRF validation failed: header missing", "method", c.Request.Method, "path", c.Request.URL.Path)
			response.ErrorWithStatus(c, http.StatusForbidden, "CSRF_HEADER_MISSING", "CSRF header missing", "")
			c.Abort()
			return
		}

		// Validate tokens match (double-submit cookie pattern)
		if cookieToken != headerToken {
			m.logger.Warn("CSRF validation failed: token mismatch", "method", c.Request.Method, "path", c.Request.URL.Path)
			response.ErrorWithStatus(c, http.StatusForbidden, "CSRF_TOKEN_INVALID", "CSRF token mismatch", "")
			c.Abort()
			return
		}

		// CSRF validation passed
		c.Next()
	}
}

// Note: min() helper function available in auth.go (shared in package)
