package user

import "errors"

// Domain errors for user operations - simple sentinel errors
var (
	// ErrNotFound is returned when a user is not found
	ErrNotFound = errors.New("not found")
	
	// ErrAlreadyExists is returned when trying to create a user that already exists
	ErrAlreadyExists = errors.New("already exists")
	
	// ErrInactive is returned when trying to operate on an inactive user
	ErrInactive = errors.New("inactive")
	
	// ErrInvalidEmail is returned when the email format is invalid
	ErrInvalidEmail = errors.New("invalid email format")
	
	// ErrWeakPassword is returned when the password doesn't meet strength requirements
	ErrWeakPassword = errors.New("password too weak")
)