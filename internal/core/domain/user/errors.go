package user

import "errors"

// Domain errors for user operations
var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	
	// ErrUserAlreadyExists is returned when trying to create a user that already exists
	ErrUserAlreadyExists = errors.New("user already exists")
	
	// ErrInvalidEmail is returned when the email format is invalid
	ErrInvalidEmail = errors.New("invalid email format")
	
	// ErrInvalidName is returned when the name is invalid
	ErrInvalidName = errors.New("invalid name")
	
	// ErrInvalidPassword is returned when the password is invalid
	ErrInvalidPassword = errors.New("invalid password")
	
	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
	
	// ErrWeakPassword is returned when the password doesn't meet strength requirements
	ErrWeakPassword = errors.New("password is too weak")
	
	// ErrUserInactive is returned when trying to operate on an inactive user
	ErrUserInactive = errors.New("user is inactive")
	
	// ErrUnauthorized is returned when user lacks permission for operation
	ErrUnauthorized = errors.New("unauthorized operation")
)