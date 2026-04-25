package dashboard

import "errors"

// Domain errors for dashboard operations
var (
	// Dashboard errors
	ErrDashboardNotFound = errors.New("dashboard not found")

	// Template errors
	ErrTemplateNotFound = errors.New("template not found")
)
