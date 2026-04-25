package credentials

import (
	"github.com/google/uuid"
)

type CreateCredentialRequest struct {
	// OrganizationID is set from the URL path parameter
	OrganizationID uuid.UUID `json:"-"`

	// Name is the user-defined unique identifier for this configuration
	// e.g., "OpenAI Production", "Claude Development"
	Name string `json:"name" validate:"required,min=1,max=100"`

	// Adapter type (openai, anthropic, azure, gemini, openrouter, custom)
	Adapter Provider `json:"adapter" validate:"required"`

	// APIKey is the plaintext API key (only sent during create/update, never returned)
	APIKey string `json:"api_key" validate:"required,min=10"`

	// BaseURL is an optional custom endpoint (Azure OpenAI, proxy, etc.)
	// Required for Azure and Custom providers
	BaseURL *string `json:"base_url,omitempty" validate:"omitempty,url"`

	// Config is provider-specific configuration
	// Azure: {"deployment_id": "...", "api_version": "..."}
	// Gemini: {"location": "..."}
	Config map[string]any `json:"config,omitempty"`

	// CustomModels are user-defined model IDs
	// For standard providers: optional fine-tuned models (e.g., "ft:gpt-4o:my-org")
	// For custom provider: required list of available models (e.g., "llama-3.1", "mistral-7b")
	CustomModels []string `json:"custom_models,omitempty"`

	// Headers are custom HTTP headers (encrypted at rest, never returned)
	// Used for proxy authentication or custom endpoints
	Headers map[string]string `json:"headers,omitempty"`

	// CreatedBy is set from the auth context
	CreatedBy *uuid.UUID `json:"-"`
}

type UpdateCredentialRequest struct {
	// Name can be updated (unique within organization)
	Name *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`

	// APIKey is optional during update (only update if provided)
	APIKey *string `json:"api_key,omitempty" validate:"omitempty,min=10"`

	// BaseURL is an optional custom endpoint
	BaseURL *string `json:"base_url,omitempty" validate:"omitempty,url"`

	// Config is provider-specific configuration
	Config map[string]any `json:"config,omitempty"`

	// CustomModels are user-defined model IDs
	CustomModels []string `json:"custom_models,omitempty"`

	// Headers are custom HTTP headers
	// Pointer type allows distinguishing between:
	// - nil: don't change headers (omitted from request)
	// - empty map: clear headers (explicitly set to {})
	// - non-empty map: set new headers
	Headers *map[string]string `json:"headers,omitempty"`
}

type TestConnectionRequest struct {
	Adapter Provider          `json:"adapter" validate:"required"`
	APIKey  string            `json:"api_key" validate:"required"`
	BaseURL *string           `json:"base_url,omitempty"`
	Config  map[string]any    `json:"config,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type TestConnectionResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// Service-level interface removed; constructor `NewProviderCredentialService`
// in `internal/core/services/credentials/` returns the concrete
// *ProviderCredentialService directly. See CLAUDE.md Mandatory Development
// Rules → "Service constructors return concrete types".
