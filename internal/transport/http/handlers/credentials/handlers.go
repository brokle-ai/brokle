// Package credentials exposes /api/v1/organizations/{orgId}/credentials/ai
// operations — AI provider credential CRUD + connection-test + available-
// models discovery. Dashboard plane, RequireAuth.
package credentials

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	analyticsDomain "brokle/internal/core/domain/analytics"
	credentialsDomain "brokle/internal/core/domain/credentials"
	"brokle/internal/transport/http/httpctx"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

// CredentialService is the narrow method set the handler consumes.
// Declared at the consumption site (Go idiom: accept interfaces) so tests
// can pass a fake and prod wiring passes *credentialsService.ProviderCredentialService.
type CredentialService interface {
	Create(ctx context.Context, req *credentialsDomain.CreateCredentialRequest) (*credentialsDomain.ProviderCredentialResponse, error)
	Update(ctx context.Context, id, orgID uuid.UUID, req *credentialsDomain.UpdateCredentialRequest) (*credentialsDomain.ProviderCredentialResponse, error)
	GetByID(ctx context.Context, id, orgID uuid.UUID) (*credentialsDomain.ProviderCredentialResponse, error)
	List(ctx context.Context, orgID uuid.UUID) ([]*credentialsDomain.ProviderCredentialResponse, error)
	Delete(ctx context.Context, id, orgID uuid.UUID) error
	TestConnection(ctx context.Context, req *credentialsDomain.TestConnectionRequest) *credentialsDomain.TestConnectionResponse
}

// ModelCatalog is the narrow method set used by the handler for model discovery.
type ModelCatalog interface {
	GetAvailableModels(ctx context.Context, orgID uuid.UUID) ([]*analyticsDomain.AvailableModel, error)
}

type Handler struct {
	svc     CredentialService
	catalog ModelCatalog
	logger  *slog.Logger
}

// New constructs a Handler with all required services.
func New(svc CredentialService, catalog ModelCatalog, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, catalog: catalog, logger: logger}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	orgID := httpctx.MustGetOrganizationID(r.Context())

	var body createCredentialBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	userID := httpctx.MustGetUserID(r.Context())
	cred, err := h.svc.Create(r.Context(), &credentialsDomain.CreateCredentialRequest{
		OrganizationID: orgID,
		Name:           body.Name,
		Adapter:        credentialsDomain.Provider(body.Adapter),
		APIKey:         body.APIKey,
		BaseURL:        body.BaseURL,
		Config:         body.Config,
		CustomModels:   body.CustomModels,
		Headers:        body.Headers,
		CreatedBy:      &userID,
	})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, cred)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	orgID := httpctx.MustGetOrganizationID(r.Context())
	creds, err := h.svc.List(r.Context(), orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, creds)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	orgID := httpctx.MustGetOrganizationID(r.Context())
	credID, err := request.URLParamUUID(r, "credentialId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	cred, err := h.svc.GetByID(r.Context(), credID, orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, cred)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	orgID := httpctx.MustGetOrganizationID(r.Context())
	credID, err := request.URLParamUUID(r, "credentialId")
	if err != nil {
		response.WriteError(w, err)
		return
	}

	var body updateCredentialBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	cred, err := h.svc.Update(r.Context(), credID, orgID, &credentialsDomain.UpdateCredentialRequest{
		Name:         body.Name,
		APIKey:       body.APIKey,
		BaseURL:      body.BaseURL,
		Config:       body.Config,
		CustomModels: body.CustomModels,
		Headers:      body.Headers,
	})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, cred)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID := httpctx.MustGetOrganizationID(r.Context())
	credID, err := request.URLParamUUID(r, "credentialId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.svc.Delete(r.Context(), credID, orgID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

// testConnection returns a structured TestConnectionResponse on both
// success and probe-level failure so the caller can render field-
// level feedback. Only input-validation / auth failures emit the
// standard error envelope.
func (h *Handler) TestConnection(w http.ResponseWriter, r *http.Request) {

	var body testConnectionBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	result := h.svc.TestConnection(r.Context(), &credentialsDomain.TestConnectionRequest{
		Adapter: credentialsDomain.Provider(body.Adapter),
		APIKey:  body.APIKey,
		BaseURL: body.BaseURL,
		Config:  body.Config,
		Headers: body.Headers,
	})
	response.Success(w, result)
}

func (h *Handler) GetAvailableModels(w http.ResponseWriter, r *http.Request) {
	orgID := httpctx.MustGetOrganizationID(r.Context())
	models, err := h.catalog.GetAvailableModels(r.Context(), orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, models)
}
