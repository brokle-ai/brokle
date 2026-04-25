package user

import "errors"

// Domain errors for user operations.
var (
	// ErrNotFound is returned when a user is not found.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists is returned when trying to create a user that already exists.
	ErrAlreadyExists = errors.New("already exists")
)
