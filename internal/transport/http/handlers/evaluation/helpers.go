package evaluation

import (
	"github.com/gin-gonic/gin"

	"brokle/internal/transport/http/middleware"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/ulid"
)

func extractProjectID(c *gin.Context) (ulid.ULID, error) {
	// Try SDK auth context first
	if projectIDPtr, exists := middleware.GetProjectID(c); exists && projectIDPtr != nil {
		return *projectIDPtr, nil
	}

	// Fall back to URL path param for dashboard routes
	projectIDStr := c.Param("projectId")
	if projectIDStr == "" {
		return ulid.ULID{}, appErrors.NewValidationError("projectId", "project ID is required")
	}

	return ulid.Parse(projectIDStr)
}
