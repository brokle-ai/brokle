package playground

import (
	"errors"
	"fmt"
)

// Domain errors for playground session management
var (
	// Session errors
	ErrSessionNotFound     = errors.New("playground session not found")
	ErrSessionAccessDenied = errors.New("access denied to playground session")

	// Validation errors
	ErrTooManyWindows = errors.New("too many comparison windows")
	ErrTooManyTags    = errors.New("too many tags")
	ErrNameRequired   = errors.New("name is required")
	ErrNameTooLong    = errors.New("session name too long")
)

// Error codes for structured API responses
const (
	ErrCodeSessionNotFound  = "SESSION_NOT_FOUND"
	ErrCodeAccessDenied     = "ACCESS_DENIED"
	ErrCodeValidationFailed = "VALIDATION_FAILED"
)

// Convenience functions for creating contextualized errors

// NewSessionNotFoundError creates a session not found error with context
func NewSessionNotFoundError(sessionID string) error {
	return fmt.Errorf("%w: id=%s", ErrSessionNotFound, sessionID)
}

// NewSessionAccessDeniedError creates an access denied error
func NewSessionAccessDeniedError(sessionID string, projectID string) error {
	return fmt.Errorf("%w: session=%s does not belong to project=%s", ErrSessionAccessDenied, sessionID, projectID)
}

// NewTooManyWindowsError creates a too many windows error
func NewTooManyWindowsError(count int) error {
	return fmt.Errorf("%w: got %d, max %d", ErrTooManyWindows, count, MaxWindowsCount)
}

// NewTooManyTagsError creates a too many tags error
func NewTooManyTagsError(count int) error {
	return fmt.Errorf("%w: got %d, max %d", ErrTooManyTags, count, MaxTagsCount)
}

// NewNameTooLongError creates a name too long error
func NewNameTooLongError(length int) error {
	return fmt.Errorf("%w: got %d chars, max %d", ErrNameTooLong, length, MaxNameLength)
}

// Error classification helpers

// IsNotFoundError checks if the error is a not-found error
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrSessionNotFound)
}

// IsAccessDeniedError checks if the error is an access denied error
func IsAccessDeniedError(err error) bool {
	return errors.Is(err, ErrSessionAccessDenied)
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	return errors.Is(err, ErrTooManyWindows) ||
		errors.Is(err, ErrTooManyTags) ||
		errors.Is(err, ErrNameRequired) ||
		errors.Is(err, ErrNameTooLong)
}
