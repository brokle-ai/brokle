// Package credentials provides the credentials management domain model.
//
// The credentials domain handles secure storage of user-provided LLM API keys
// with AES-256-GCM encryption at rest. Keys are scoped per-project and support
// multiple providers (OpenAI, Anthropic, etc.).
package credentials

import (
	"errors"
	"fmt"
)

// Domain errors for credential management
var (
	// Credential errors
	ErrCredentialNotFound     = errors.New("credential not found")
	ErrCredentialExists       = errors.New("credential already exists for this provider")
	ErrInvalidProvider        = errors.New("invalid LLM provider")
	ErrInvalidAPIKey          = errors.New("invalid API key format")
	ErrNoKeyConfigured        = errors.New("no API key configured for provider")

	// Encryption errors
	ErrEncryptionFailed       = errors.New("failed to encrypt API key")
	ErrDecryptionFailed       = errors.New("failed to decrypt API key")
	ErrEncryptionKeyMissing   = errors.New("encryption key not configured")

	// Validation errors
	ErrAPIKeyValidationFailed = errors.New("API key validation failed")
	ErrInvalidBaseURL         = errors.New("invalid base URL")
)

// Error codes for structured API responses
const (
	ErrCodeCredentialNotFound   = "CREDENTIAL_NOT_FOUND"
	ErrCodeCredentialExists     = "CREDENTIAL_EXISTS"
	ErrCodeInvalidProvider      = "INVALID_PROVIDER"
	ErrCodeInvalidAPIKey        = "INVALID_API_KEY"
	ErrCodeNoKeyConfigured      = "NO_KEY_CONFIGURED"
	ErrCodeEncryptionFailed     = "ENCRYPTION_FAILED"
	ErrCodeValidationFailed     = "VALIDATION_FAILED"
)

// Convenience functions for creating contextualized errors

// NewCredentialNotFoundError creates a credential not found error with provider context.
func NewCredentialNotFoundError(provider string, projectID string) error {
	return fmt.Errorf("%w: provider=%s project=%s", ErrCredentialNotFound, provider, projectID)
}

// NewCredentialExistsError creates a credential exists error.
func NewCredentialExistsError(provider string, projectID string) error {
	return fmt.Errorf("%w: provider=%s already configured for project=%s", ErrCredentialExists, provider, projectID)
}

// NewInvalidProviderError creates an invalid provider error.
func NewInvalidProviderError(provider string) error {
	return fmt.Errorf("%w: '%s' (must be 'openai' or 'anthropic')", ErrInvalidProvider, provider)
}

// NewAPIKeyValidationError creates an API key validation error with details.
func NewAPIKeyValidationError(provider string, details string) error {
	return fmt.Errorf("%w: %s - %s", ErrAPIKeyValidationFailed, provider, details)
}

// NewNoKeyConfiguredError creates a no key configured error.
func NewNoKeyConfiguredError(provider string) error {
	return fmt.Errorf("%w: %s (set via project settings or environment variable)", ErrNoKeyConfigured, provider)
}

// Error classification helpers

// IsNotFoundError checks if the error is a not-found error.
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrCredentialNotFound)
}

// IsValidationError checks if the error is a validation error.
func IsValidationError(err error) bool {
	return errors.Is(err, ErrInvalidProvider) ||
		errors.Is(err, ErrInvalidAPIKey) ||
		errors.Is(err, ErrAPIKeyValidationFailed) ||
		errors.Is(err, ErrInvalidBaseURL)
}

// IsEncryptionError checks if the error is an encryption-related error.
func IsEncryptionError(err error) bool {
	return errors.Is(err, ErrEncryptionFailed) ||
		errors.Is(err, ErrDecryptionFailed) ||
		errors.Is(err, ErrEncryptionKeyMissing)
}

// IsConflictError checks if the error is a conflict error.
func IsConflictError(err error) bool {
	return errors.Is(err, ErrCredentialExists)
}
