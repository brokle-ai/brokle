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

	// API Key errors
	ErrAPIKeyNotFound      = errors.New("api key not found")
	ErrAPIKeyAlreadyExists = errors.New("api key already exists")

	// RBAC role errors
	ErrRoleAlreadyExists = errors.New("role already exists")
)
