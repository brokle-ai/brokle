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

func (h *handler) create(ctx context.Context, in *CreateCredentialInput) (*CreateCredentialOutput, error) {
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
	return &CreateCredentialOutput{Body: cred}, nil
}

// ----- list ------------------------------------------------------------

func (h *handler) list(ctx context.Context, in *ListCredentialsInput) (*ListCredentialsOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	creds, err := h.svc.List(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return &ListCredentialsOutput{Body: creds}, nil
}

// ----- get -------------------------------------------------------------

func (h *handler) get(ctx context.Context, in *GetCredentialInput) (*GetCredentialOutput, error) {
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
	return &GetCredentialOutput{Body: cred}, nil
}

// ----- update ----------------------------------------------------------

func (h *handler) update(ctx context.Context, in *UpdateCredentialInput) (*UpdateCredentialOutput, error) {
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
	return &UpdateCredentialOutput{Body: cred}, nil
}

// ----- delete ----------------------------------------------------------

func (h *handler) delete(ctx context.Context, in *DeleteCredentialInput) (*DeleteCredentialOutput, error) {
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
	return &DeleteCredentialOutput{}, nil
}

// ----- test-connection -------------------------------------------------
//
// Returns a structured TestConnectionResponse regardless of success/
// failure — the caller wants field-level feedback on connection
// problems, not an error envelope it has to decode differently from
// every other API call.

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
