package credentials

import (
	"context"

	"brokle/pkg/ulid"
)

// ----------------------------
// Request DTOs
// ----------------------------

// CreateCredentialRequest represents a request to create or update a provider credential.
type CreateCredentialRequest struct {
	// ProjectID is set from the URL path parameter
	ProjectID ulid.ULID `json:"-"`

	// Provider must be "openai" or "anthropic"
	Provider LLMProvider `json:"provider" validate:"required"`

	// APIKey is the plaintext API key (only sent during create/update, never returned)
	APIKey string `json:"api_key" validate:"required,min=10"`

	// BaseURL is an optional custom endpoint (Azure OpenAI, proxy, etc.)
	BaseURL *string `json:"base_url,omitempty" validate:"omitempty,url"`

	// CreatedBy is set from the auth context
	CreatedBy *ulid.ULID `json:"-"`
}

// ----------------------------
// Service Interface
// ----------------------------

// LLMProviderCredentialService defines the service interface for LLM provider credential management.
type LLMProviderCredentialService interface {
	// CreateOrUpdate creates a new credential or updates an existing one for the project/provider.
	// The API key is validated with the provider before storing.
	// Returns the safe response (no encrypted data).
	CreateOrUpdate(ctx context.Context, req *CreateCredentialRequest) (*LLMProviderCredentialResponse, error)

	// Get retrieves a credential by project and provider.
	// Returns the safe response (no encrypted data, only masked key preview).
	Get(ctx context.Context, projectID ulid.ULID, provider LLMProvider) (*LLMProviderCredentialResponse, error)

	// List retrieves all credentials for a project.
	// Returns safe responses (no encrypted data).
	List(ctx context.Context, projectID ulid.ULID) ([]*LLMProviderCredentialResponse, error)

	// Delete removes a credential by project and provider.
	Delete(ctx context.Context, projectID ulid.ULID, provider LLMProvider) error

	// GetDecrypted retrieves the decrypted key configuration.
	// This is ONLY for internal use during prompt execution.
	// Returns ErrCredentialNotFound if no credential exists.
	GetDecrypted(ctx context.Context, projectID ulid.ULID, provider LLMProvider) (*DecryptedKeyConfig, error)

	// GetExecutionConfig returns the key configuration for LLM execution.
	// Tries user-provided key first, falls back to environment variable configuration.
	// Returns ErrNoKeyConfigured if neither is available.
	GetExecutionConfig(ctx context.Context, projectID ulid.ULID, provider LLMProvider) (*DecryptedKeyConfig, error)

	// ValidateKey validates an API key with the provider without storing it.
	// Makes a lightweight API call to verify the key works.
	ValidateKey(ctx context.Context, provider LLMProvider, apiKey string, baseURL *string) error
}
