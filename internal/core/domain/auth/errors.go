package auth

import "errors"

// Domain errors for auth operations - simple sentinel errors
var (
	// Generic errors
	ErrNotFound = errors.New("not found")
	
	// Credential errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	
	// Token errors
	ErrTokenExpired = errors.New("token expired")
	ErrTokenInvalid = errors.New("token invalid")
	
	// Session errors
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
)