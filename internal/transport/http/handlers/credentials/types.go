// Request-body DTOs for the credentials handler. One struct per
// body shape; validation tags drive go-playground/validator/v10 via
// pkg/request.DecodeJSON. Response bodies are the domain
// ProviderCredentialResponse / TestConnectionResponse /
// analyticsDomain.AvailableModel types directly — no wrapper DTOs
// because every response is a single resource or a list of
// resources and the domain types already carry the wire tags.
package credentials

// createCredentialBody — POST /api/v1/organizations/{orgId}/credentials/ai
type createCredentialBody struct {
	Name         string            `json:"name"                    validate:"required,min=1,max=100"`
	Adapter      string            `json:"adapter"                 validate:"required,oneof=openai anthropic azure gemini openrouter custom"`
	APIKey       string            `json:"api_key"                 validate:"required,min=10"`
	BaseURL      *string           `json:"base_url,omitempty"      validate:"omitempty,url"`
	Config       map[string]any    `json:"config,omitempty"`
	CustomModels []string          `json:"custom_models,omitempty" validate:"omitempty,dive,min=1"`
	Headers      map[string]string `json:"headers,omitempty"`
}

// updateCredentialBody — PATCH /api/v1/organizations/{orgId}/credentials/ai/{credentialId}
//
// All fields are optional — absent = leave unchanged. Headers is a
// pointer-to-map so an explicit `{}` clears all headers (distinct
// from omitting the key).
type updateCredentialBody struct {
	Name         *string            `json:"name,omitempty"          validate:"omitempty,min=1,max=100"`
	APIKey       *string            `json:"api_key,omitempty"       validate:"omitempty,min=10"`
	BaseURL      *string            `json:"base_url,omitempty"      validate:"omitempty,url"`
	Config       map[string]any     `json:"config,omitempty"`
	CustomModels []string           `json:"custom_models,omitempty" validate:"omitempty,dive,min=1"`
	Headers      *map[string]string `json:"headers,omitempty"`
}

// testConnectionBody — POST /api/v1/organizations/{orgId}/credentials/ai/test
//
// APIKey uses a looser min=1 than createCredentialBody because some
// adapters accept short placeholder tokens during connectivity probes
// (e.g. a custom adapter pointing at a local proxy).
type testConnectionBody struct {
	Adapter string            `json:"adapter"            validate:"required,oneof=openai anthropic azure gemini openrouter custom"`
	APIKey  string            `json:"api_key"            validate:"required,min=1"`
	BaseURL *string           `json:"base_url,omitempty" validate:"omitempty,url"`
	Config  map[string]any    `json:"config,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}
