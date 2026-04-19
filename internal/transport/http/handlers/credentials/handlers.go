// Package credentials exposes /api/v1/organizations/{orgId}/credentials/ai
// operations — AI provider credential CRUD + connection-test + available-
// models discovery. Dashboard plane (apiAdmin). Every op requires
// RequireAuth.
package credentials

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	analyticsDomain "brokle/internal/core/domain/analytics"
	credentialsDomain "brokle/internal/core/domain/credentials"
	credentialsService "brokle/internal/core/services/credentials"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
)

type handler struct {
	svc      credentialsDomain.ProviderCredentialService
	catalog  credentialsService.ModelCatalogService
	logger   *slog.Logger
}

// RegisterRoutes registers every credential operation on apiAdmin.
func RegisterRoutes(
	api huma.API,
	svc credentialsDomain.ProviderCredentialService,
	catalog credentialsService.ModelCatalogService,
	logger *slog.Logger,
) {
	h := &handler{svc: svc, catalog: catalog, logger: logger}

	huma.Register(api, huma.Operation{
		OperationID:   "create-credential",
		Method:        http.MethodPost,
		Path:          "/api/v1/organizations/{orgId}/credentials/ai",
		Tags:          []string{"credentials"},
		Summary:       "Create an AI provider credential",
		Description:   "Each credential has a unique name within the organization. The API key is encrypted at rest; the response returns a masked preview.",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "list-credentials",
		Method:      http.MethodGet,
		Path:        "/api/v1/organizations/{orgId}/credentials/ai",
		Tags:        []string{"credentials"},
		Summary:     "List AI provider credentials for an organization",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID: "get-credential",
		Method:      http.MethodGet,
		Path:        "/api/v1/organizations/{orgId}/credentials/ai/{credentialId}",
		Tags:        []string{"credentials"},
		Summary:     "Get a specific AI provider credential",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.get)

	huma.Register(api, huma.Operation{
		OperationID: "update-credential",
		Method:      http.MethodPatch,
		Path:        "/api/v1/organizations/{orgId}/credentials/ai/{credentialId}",
		Tags:        []string{"credentials"},
		Summary:     "Update an AI provider credential (partial)",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.update)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-credential",
		Method:        http.MethodDelete,
		Path:          "/api/v1/organizations/{orgId}/credentials/ai/{credentialId}",
		Tags:          []string{"credentials"},
		Summary:       "Delete an AI provider credential",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.delete)

	huma.Register(api, huma.Operation{
		OperationID: "test-credential-connection",
		Method:      http.MethodPost,
		Path:        "/api/v1/organizations/{orgId}/credentials/ai/test",
		Tags:        []string{"credentials"},
		Summary:     "Test an AI provider connection without saving",
		Description: "Validates the provided API key + configuration against the provider. Returns a structured result rather than an error envelope so the caller can render field-level feedback.",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.testConnection)

	huma.Register(api, huma.Operation{
		OperationID: "get-available-models",
		Method:      http.MethodGet,
		Path:        "/api/v1/organizations/{orgId}/credentials/ai/models",
		Tags:        []string{"credentials"},
		Summary:     "Get available models derived from configured providers",
		Description: "Standard providers (openai, anthropic, …) return the default model set plus any custom models. Custom providers return only user-defined models.",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getAvailableModels)
}

// ----- shared input helpers --------------------------------------------

func parseOrg(orgIDStr string) (uuid.UUID, error) {
	id, err := uuid.Parse(orgIDStr)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid organization ID", "orgId must be a valid UUID")
	}
	return id, nil
}

func parseCred(credIDStr string) (uuid.UUID, error) {
	id, err := uuid.Parse(credIDStr)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid credential ID", "credentialId must be a valid UUID")
	}
	return id, nil
}

// ----- create ----------------------------------------------------------

type CreateInput struct {
	OrgID string `path:"orgId" format:"uuid" doc:"Organization the credential belongs to"`
	Body  createBody
}

type createBody struct {
	Name         string            `json:"name" minLength:"1" maxLength:"100" doc:"Human-readable credential name, unique within the organization"`
	Adapter      string            `json:"adapter" enum:"openai,anthropic,azure,gemini,openrouter,custom" doc:"Provider adapter"`
	APIKey       string            `json:"api_key" minLength:"10" doc:"Provider API key — encrypted at rest on storage"`
	BaseURL      *string           `json:"base_url,omitempty" doc:"Override the provider base URL (required for azure / custom)"`
	Config       map[string]any    `json:"config,omitempty" doc:"Adapter-specific config (e.g. azure deployment_id)"`
	CustomModels []string          `json:"custom_models,omitempty" doc:"Custom model identifiers — required for custom adapter"`
	Headers      map[string]string `json:"headers,omitempty" doc:"Extra HTTP headers sent on every provider request"`
}

type CreateOutput struct {
	Body *credentialsDomain.ProviderCredentialResponse
}

func (h *handler) create(ctx context.Context, in *CreateInput) (*CreateOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)

	cred, err := h.svc.Create(ctx, &credentialsDomain.CreateCredentialRequest{
		OrganizationID: orgID,
		Name:           in.Body.Name,
		Adapter:        credentialsDomain.Provider(in.Body.Adapter),
		APIKey:         in.Body.APIKey,
		BaseURL:        in.Body.BaseURL,
		Config:         in.Body.Config,
		CustomModels:   in.Body.CustomModels,
		Headers:        in.Body.Headers,
		CreatedBy:      &userID,
	})
	if err != nil {
		return nil, err
	}
	return &CreateOutput{Body: cred}, nil
}

// ----- list ------------------------------------------------------------

type ListInput struct {
	OrgID string `path:"orgId" format:"uuid"`
}

type ListOutput struct {
	Body []*credentialsDomain.ProviderCredentialResponse
}

func (h *handler) list(ctx context.Context, in *ListInput) (*ListOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	creds, err := h.svc.List(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return &ListOutput{Body: creds}, nil
}

// ----- get -------------------------------------------------------------

type GetInput struct {
	OrgID        string `path:"orgId" format:"uuid"`
	CredentialID string `path:"credentialId" format:"uuid"`
}

type GetOutput struct {
	Body *credentialsDomain.ProviderCredentialResponse
}

func (h *handler) get(ctx context.Context, in *GetInput) (*GetOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	credID, err := parseCred(in.CredentialID)
	if err != nil {
		return nil, err
	}
	cred, err := h.svc.GetByID(ctx, credID, orgID)
	if err != nil {
		return nil, err
	}
	return &GetOutput{Body: cred}, nil
}

// ----- update ----------------------------------------------------------

type UpdateInput struct {
	OrgID        string `path:"orgId" format:"uuid"`
	CredentialID string `path:"credentialId" format:"uuid"`
	Body         updateBody
}

type updateBody struct {
	Name         *string            `json:"name,omitempty" minLength:"1" maxLength:"100"`
	APIKey       *string            `json:"api_key,omitempty" minLength:"10"`
	BaseURL      *string            `json:"base_url,omitempty"`
	Config       map[string]any     `json:"config,omitempty"`
	CustomModels []string           `json:"custom_models,omitempty"`
	// Headers is a pointer-to-map so an explicit empty object clears
	// all headers, while omitting the field leaves them untouched.
	// Matches the domain UpdateCredentialRequest contract.
	Headers *map[string]string `json:"headers,omitempty"`
}

type UpdateOutput struct {
	Body *credentialsDomain.ProviderCredentialResponse
}

func (h *handler) update(ctx context.Context, in *UpdateInput) (*UpdateOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	credID, err := parseCred(in.CredentialID)
	if err != nil {
		return nil, err
	}

	cred, err := h.svc.Update(ctx, credID, orgID, &credentialsDomain.UpdateCredentialRequest{
		Name:         in.Body.Name,
		APIKey:       in.Body.APIKey,
		BaseURL:      in.Body.BaseURL,
		Config:       in.Body.Config,
		CustomModels: in.Body.CustomModels,
		Headers:      in.Body.Headers,
	})
	if err != nil {
		return nil, err
	}
	return &UpdateOutput{Body: cred}, nil
}

// ----- delete ----------------------------------------------------------

type DeleteInput struct {
	OrgID        string `path:"orgId" format:"uuid"`
	CredentialID string `path:"credentialId" format:"uuid"`
}

type DeleteOutput struct{}

func (h *handler) delete(ctx context.Context, in *DeleteInput) (*DeleteOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	credID, err := parseCred(in.CredentialID)
	if err != nil {
		return nil, err
	}
	if err := h.svc.Delete(ctx, credID, orgID); err != nil {
		return nil, err
	}
	return &DeleteOutput{}, nil
}

// ----- test-connection -------------------------------------------------
//
// Returns a structured TestConnectionResponse regardless of success/
// failure — the caller wants field-level feedback on connection
// problems, not an error envelope it has to decode differently from
// every other API call.

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

func (h *handler) testConnection(ctx context.Context, in *TestConnectionInput) (*TestConnectionOutput, error) {
	if _, err := parseOrg(in.OrgID); err != nil {
		return nil, err
	}
	result := h.svc.TestConnection(ctx, &credentialsDomain.TestConnectionRequest{
		Adapter: credentialsDomain.Provider(in.Body.Adapter),
		APIKey:  in.Body.APIKey,
		BaseURL: in.Body.BaseURL,
		Config:  in.Body.Config,
		Headers: in.Body.Headers,
	})
	return &TestConnectionOutput{Body: result}, nil
}

// ----- get-available-models --------------------------------------------

type GetAvailableModelsInput struct {
	OrgID string `path:"orgId" format:"uuid"`
}

type GetAvailableModelsOutput struct {
	Body []*analyticsDomain.AvailableModel
}

func (h *handler) getAvailableModels(ctx context.Context, in *GetAvailableModelsInput) (*GetAvailableModelsOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	models, err := h.catalog.GetAvailableModels(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return &GetAvailableModelsOutput{Body: models}, nil
}
