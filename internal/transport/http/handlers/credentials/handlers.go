// Package credentials exposes /api/v1/organizations/{orgId}/credentials/ai
// operations — AI provider credential CRUD + connection-test + available-
// models discovery. Dashboard plane, RequireAuth.
package credentials

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
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

type handler struct {
	svc     CredentialService
	catalog ModelCatalog
	logger  *slog.Logger
}

// RegisterRoutes registers every credential operation on the provided
// chi router. Expected mount context: the authed dashboard group
// (RequireAuth + LimitByUser already applied).
func RegisterRoutes(
	r chi.Router,
	svc CredentialService,
	catalog ModelCatalog,
	logger *slog.Logger,
) {
	h := &handler{svc: svc, catalog: catalog, logger: logger}

	r.Route("/api/v1/organizations/{orgId}/credentials/ai", func(r chi.Router) {
		r.Post("/", h.create)
		r.Get("/", h.list)
		r.Get("/{credentialId}", h.get)
		r.Patch("/{credentialId}", h.update)
		r.Delete("/{credentialId}", h.delete)
		r.Post("/test", h.testConnection)
		r.Get("/models", h.getAvailableModels)
	})
}

// ---------------------------------------------------------------------------

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}

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

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	creds, err := h.svc.List(r.Context(), orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, creds)
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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

func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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
func (h *handler) testConnection(w http.ResponseWriter, r *http.Request) {
	if _, err := request.URLParamUUID(r, "orgId"); err != nil {
		response.WriteError(w, err)
		return
	}

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

func (h *handler) getAvailableModels(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	models, err := h.catalog.GetAvailableModels(r.Context(), orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, models)
}
