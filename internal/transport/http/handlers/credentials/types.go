package credentials

import (
	analyticsDomain "brokle/internal/core/domain/analytics"
	credentialsDomain "brokle/internal/core/domain/credentials"
)

// Huma operation types for the credentials package.

// ----- create ----------------------------------------------------------

type CreateCredentialInput struct {
	OrgID string `path:"orgId" format:"uuid" doc:"Organization the credential belongs to"`
	Body  createCredentialBody
}

type createCredentialBody struct {
	Name         string            `json:"name" minLength:"1" maxLength:"100" doc:"Human-readable credential name, unique within the organization"`
	Adapter      string            `json:"adapter" enum:"openai,anthropic,azure,gemini,openrouter,custom" doc:"Provider adapter"`
	APIKey       string            `json:"api_key" minLength:"10" doc:"Provider API key — encrypted at rest on storage"`
	BaseURL      *string           `json:"base_url,omitempty" doc:"Override the provider base URL (required for azure / custom)"`
	Config       map[string]any    `json:"config,omitempty" doc:"Adapter-specific config (e.g. azure deployment_id)"`
	CustomModels []string          `json:"custom_models,omitempty" doc:"Custom model identifiers — required for custom adapter"`
	Headers      map[string]string `json:"headers,omitempty" doc:"Extra HTTP headers sent on every provider request"`
}

type CreateCredentialOutput struct {
	Body *credentialsDomain.ProviderCredentialResponse
}

// ----- list ------------------------------------------------------------

type ListCredentialsInput struct {
	OrgID string `path:"orgId" format:"uuid"`
}

type ListCredentialsOutput struct {
	Body []*credentialsDomain.ProviderCredentialResponse
}

// ----- get -------------------------------------------------------------

type GetCredentialInput struct {
	OrgID        string `path:"orgId" format:"uuid"`
	CredentialID string `path:"credentialId" format:"uuid"`
}

type GetCredentialOutput struct {
	Body *credentialsDomain.ProviderCredentialResponse
}

// ----- update ----------------------------------------------------------

type UpdateCredentialInput struct {
	OrgID        string `path:"orgId" format:"uuid"`
	CredentialID string `path:"credentialId" format:"uuid"`
	Body         updateCredentialBody
}

type updateCredentialBody struct {
	Name         *string        `json:"name,omitempty" minLength:"1" maxLength:"100"`
	APIKey       *string        `json:"api_key,omitempty" minLength:"10"`
	BaseURL      *string        `json:"base_url,omitempty"`
	Config       map[string]any `json:"config,omitempty"`
	CustomModels []string       `json:"custom_models,omitempty"`
	// Headers is a pointer-to-map so an explicit empty object clears
	// all headers, while omitting the field leaves them untouched.
	// Matches the domain UpdateCredentialRequest contract.
	Headers *map[string]string `json:"headers,omitempty"`
}

type UpdateCredentialOutput struct {
	Body *credentialsDomain.ProviderCredentialResponse
}

// ----- delete ----------------------------------------------------------

type DeleteCredentialInput struct {
	OrgID        string `path:"orgId" format:"uuid"`
	CredentialID string `path:"credentialId" format:"uuid"`
}

type DeleteCredentialOutput struct{}

// ----- test-connection -------------------------------------------------

type TestConnectionInput struct {
	OrgID string `path:"orgId" format:"uuid"`
	Body  testConnectionBody
}

type testConnectionBody struct {
	Adapter string            `json:"adapter" enum:"openai,anthropic,azure,gemini,openrouter,custom"`
	APIKey  string            `json:"api_key" minLength:"1" doc:"API key to probe — not persisted"`
	BaseURL *string           `json:"base_url,omitempty"`
	Config  map[string]any    `json:"config,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type TestConnectionOutput struct {
	Body *credentialsDomain.TestConnectionResponse
}

// ----- get-available-models --------------------------------------------

type GetAvailableModelsInput struct {
	OrgID string `path:"orgId" format:"uuid"`
}

type GetAvailableModelsOutput struct {
	Body []*analyticsDomain.AvailableModel
}
