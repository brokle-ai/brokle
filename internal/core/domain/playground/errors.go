package playground

import "errors"

// Domain errors for playground session management.
var (
	// ErrSessionNotFound is returned when a playground session does not
	// exist in the project (or has been hard-deleted). Repository wraps
	// it via fmt.Errorf("...: %w", ErrSessionNotFound); the service
	// layer translates to a 404 AppError.
	ErrSessionNotFound = errors.New("playground session not found")
)
