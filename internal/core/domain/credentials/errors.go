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

// Domain errors for credential management.
var (
	ErrCredentialNotFound = errors.New("credential not found")
	ErrCredentialExists   = errors.New("credential with this name already exists")
	ErrInvalidProvider    = errors.New("invalid adapter type")
	ErrNoKeyConfigured    = errors.New("no API key configured")
	ErrAdapterMismatch    = errors.New("credential adapter mismatch")
	ErrDecryptionFailed   = errors.New("failed to decrypt API key")
)

// NewInvalidAdapterError wraps ErrInvalidProvider with the offending
// adapter value, used by the credentials service when validating CRUD
// requests.
func NewInvalidAdapterError(adapter string) error {
	return fmt.Errorf("%w: '%s' (must be one of: openai, anthropic, azure, gemini, openrouter, custom)", ErrInvalidProvider, adapter)
}

// NewAdapterMismatchError wraps ErrAdapterMismatch with the expected
// vs actual adapter values. Used when the playground service resolves
// a credential whose stored adapter doesn't match the request adapter.
func NewAdapterMismatchError(expected, actual string) error {
	return fmt.Errorf("%w: expected '%s', credential uses '%s'", ErrAdapterMismatch, expected, actual)
}
