package auth

import "errors"

// Domain errors for auth operations - simple sentinel errors
var (
	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
	
	// ErrTokenExpired is returned when a token has expired
	ErrTokenExpired = errors.New("token expired")
	
	// ErrTokenInvalid is returned when a token is invalid
	ErrTokenInvalid = errors.New("token invalid")
	
	// ErrSessionNotFound is returned when a session is not found
	ErrSessionNotFound = errors.New("session not found")
	
	// ErrSessionExpired is returned when a session has expired
	ErrSessionExpired = errors.New("session expired")
)