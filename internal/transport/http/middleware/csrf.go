package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/pkg/response"
)

// CSRFMiddleware handles CSRF token validation using double-submit cookie pattern
type CSRFMiddleware struct {
	logger *logrus.Logger
}

// NewCSRFMiddleware creates a new CSRF validation middleware
func NewCSRFMiddleware(logger *logrus.Logger) *CSRFMiddleware {
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
			m.logger.WithFields(logrus.Fields{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
			}).Warn("CSRF validation failed: cookie missing")
			response.ErrorWithStatus(c, http.StatusForbidden, "CSRF_TOKEN_MISSING", "CSRF token missing in cookie", "")
			c.Abort()
			return
		}

		// Get CSRF token from request header
		headerToken := c.GetHeader("X-CSRF-Token")
		if headerToken == "" {
			m.logger.WithFields(logrus.Fields{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
			}).Warn("CSRF validation failed: header missing")
			response.ErrorWithStatus(c, http.StatusForbidden, "CSRF_HEADER_MISSING", "CSRF header missing", "")
			c.Abort()
			return
		}

		// Validate tokens match (double-submit cookie pattern)
		if cookieToken != headerToken {
			m.logger.WithFields(logrus.Fields{
				"method":         c.Request.Method,
				"path":           c.Request.URL.Path,
				"cookie_preview": cookieToken[:min(len(cookieToken), 10)] + "...",
				"header_preview": headerToken[:min(len(headerToken), 10)] + "...",
			}).Warn("CSRF validation failed: token mismatch")
			response.ErrorWithStatus(c, http.StatusForbidden, "CSRF_TOKEN_INVALID", "CSRF token mismatch", "")
			c.Abort()
			return
		}

		// CSRF validation passed
		c.Next()
	}
}

// Note: min() helper function available in auth.go (shared in package)
