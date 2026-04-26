// Chi-native handler tests. Reference template every migrated domain
// copies: stand up a chi.Router, register the handler, inject auth
// context via httpctx, drive operations through httptest, and
// assert wire shapes on both success and error paths.
package credentials_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	analyticsDomain "brokle/internal/core/domain/analytics"
	credentialsDomain "brokle/internal/core/domain/credentials"
	handler "brokle/internal/transport/http/handlers/credentials"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/response"
)

// ---- fake services -------------------------------------------------------

type fakeProviderCredentialService struct {
	listResp []*credentialsDomain.ProviderCredentialResponse
	listErr  error
}

func (f *fakeProviderCredentialService) Create(ctx context.Context, req *credentialsDomain.CreateCredentialRequest) (*credentialsDomain.ProviderCredentialResponse, error) {
	return nil, nil
}
func (f *fakeProviderCredentialService) Update(ctx context.Context, id uuid.UUID, orgID uuid.UUID, req *credentialsDomain.UpdateCredentialRequest) (*credentialsDomain.ProviderCredentialResponse, error) {
	return nil, nil
}
func (f *fakeProviderCredentialService) GetByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*credentialsDomain.ProviderCredentialResponse, error) {
	return nil, nil
}
func (f *fakeProviderCredentialService) GetByName(ctx context.Context, orgID uuid.UUID, name string) (*credentialsDomain.ProviderCredentialResponse, error) {
	return nil, nil
}
func (f *fakeProviderCredentialService) List(ctx context.Context, orgID uuid.UUID) ([]*credentialsDomain.ProviderCredentialResponse, error) {
	return f.listResp, f.listErr
}
func (f *fakeProviderCredentialService) Delete(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	return nil
}
func (f *fakeProviderCredentialService) GetDecryptedByID(ctx context.Context, credentialID uuid.UUID, orgID uuid.UUID) (*credentialsDomain.DecryptedKeyConfig, error) {
	return nil, nil
}
func (f *fakeProviderCredentialService) GetExecutionConfig(ctx context.Context, orgID uuid.UUID, credentialID uuid.UUID, adapter credentialsDomain.Provider) (*credentialsDomain.DecryptedKeyConfig, error) {
	return nil, nil
}
func (f *fakeProviderCredentialService) ValidateKey(ctx context.Context, adapter credentialsDomain.Provider, apiKey string, baseURL *string, config map[string]any) error {
	return nil
}
func (f *fakeProviderCredentialService) TestConnection(ctx context.Context, req *credentialsDomain.TestConnectionRequest) *credentialsDomain.TestConnectionResponse {
	return &credentialsDomain.TestConnectionResponse{Success: true}
}

type fakeModelCatalogService struct{}

func (fakeModelCatalogService) GetAvailableModels(ctx context.Context, orgID uuid.UUID) ([]*analyticsDomain.AvailableModel, error) {
	return []*analyticsDomain.AvailableModel{}, nil
}

// ---- helpers -------------------------------------------------------------

// newTestRouter mounts the credentials handler on a bare chi router
// with a user-context injector standing in for the production
// RequireAuth middleware. The handler now uses relative paths and
// expects to be mounted under /api/v1/organizations/{orgId} with
// RequireOrganizationAccess upstream — we simulate that by parsing
// the URL param and pinning orgID via httpctx.
func newTestRouter(t *testing.T, svc handler.CredentialService) (*chi.Mux, uuid.UUID) {
	t.Helper()
	userID := uuid.New()

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := httpctx.WithUserID(req.Context(), userID)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	r.Route("/api/v1/organizations/{orgId}", func(r chi.Router) {
		// Stand-in for RequireOrganizationAccess: parse orgId, pin via httpctx.
		// Returns 422 on bad UUID (matches production middleware envelope).
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				raw := chi.URLParam(req, "orgId")
				orgID, err := uuid.Parse(raw)
				if err != nil {
					response.WriteError(w, appErrors.NewValidationError(
						"Invalid orgId", "orgId must be a valid UUID",
						appErrors.WithParam("orgId"),
					))
					return
				}
				ctx := httpctx.WithOrganizationID(req.Context(), orgID)
				next.ServeHTTP(w, req.WithContext(ctx))
			})
		})
		h := handler.New(svc, fakeModelCatalogService{}, slog.Default())
		// Mirror routes from internal/server/routes.go credentials group.
		// Test URLs omit the trailing slash; register both forms for parity.
		r.Post("/credentials/ai", h.Create)
		r.Get("/credentials/ai", h.List)
		r.Get("/credentials/ai/{credentialId}", h.Get)
		r.Patch("/credentials/ai/{credentialId}", h.Update)
		r.Delete("/credentials/ai/{credentialId}", h.Delete)
		r.Post("/credentials/ai/test", h.TestConnection)
		r.Get("/credentials/ai/models", h.GetAvailableModels)
	})
	return r, userID
}

// ---- tests ---------------------------------------------------------------

// Happy path: list returns the domain slice as the raw JSON body —
// Stripe/OpenAI-style, no envelope, no `data` wrapper.
func TestListCredentials_Returns200WithBody(t *testing.T) {
	orgID := uuid.New()
	credID := uuid.New()
	svc := &fakeProviderCredentialService{
		listResp: []*credentialsDomain.ProviderCredentialResponse{
			{
				ID:             credID,
				OrganizationID: orgID,
				Name:           "prod-openai",
				Adapter:        credentialsDomain.Provider("openai"),
				KeyPreview:     "sk-***wxyz",
			},
		},
	}

	r, _ := newTestRouter(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/"+orgID.String()+"/credentials/ai", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var items []credentialsDomain.ProviderCredentialResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &items))
	require.Len(t, items, 1)
	assert.Equal(t, credID, items[0].ID)
	assert.Equal(t, "prod-openai", items[0].Name)
}

// Error path: a malformed UUID on the path triggers the
// pkg/request.URLParamUUID validation. The wire body must be the
// canonical Stripe-style `{error: {...}}` envelope with Param set.
func TestListCredentials_InvalidOrgID_EmitsErrorEnvelope(t *testing.T) {
	r, _ := newTestRouter(t, &fakeProviderCredentialService{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/not-a-uuid/credentials/ai", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var env response.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	require.NotNil(t, env.Error, "ErrorResponse.error must be populated on error paths")
	assert.Equal(t, "validation_error", env.Error.Type)
	assert.Equal(t, "orgId", env.Error.Param)
	assert.Contains(t, env.Error.Message, "Invalid orgId")
}

// Request-body validation: missing required fields + out-of-range
// values surface as a 422 with a populated per-field Errors[] array.
// Locks the go-playground/validator → pkg/request.DecodeJSON →
// pkg/response.WriteError wire path end-to-end.
func TestCreateCredential_ValidationErrors_EmitsPerFieldDiagnostics(t *testing.T) {
	orgID := uuid.New()
	r, _ := newTestRouter(t, &fakeProviderCredentialService{})

	// Empty body — every required field is missing.
	body := `{}`
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/organizations/"+orgID.String()+"/credentials/ai",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var env response.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	require.NotEmpty(t, env.Error.Errors, "per-field Errors[] must be populated")

	// Field paths must use the JSON tag names (api_key, not APIKey).
	seen := map[string]bool{}
	for _, d := range env.Error.Errors {
		seen[d.Location] = true
	}
	for _, want := range []string{"body.name", "body.adapter", "body.api_key"} {
		assert.True(t, seen[want], "expected %q in Errors[]; got keys: %v", want, seen)
	}
}

// Unknown fields must be rejected (DisallowUnknownFields), not
// silently ignored. Client typos are surfaced instead of dropped.
func TestCreateCredential_UnknownField_Rejected(t *testing.T) {
	orgID := uuid.New()
	r, _ := newTestRouter(t, &fakeProviderCredentialService{})

	body := `{"name":"x","adapter":"openai","api_key":"sk-verylongkey","surprise":"hi"}`
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/organizations/"+orgID.String()+"/credentials/ai",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var env response.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	assert.Equal(t, "surprise", env.Error.Param)
}

// Delete returns 204 with no body per CLAUDE.md's NoContent rule.
func TestDeleteCredential_Returns204NoContent(t *testing.T) {
	orgID := uuid.New()
	credID := uuid.New()
	r, _ := newTestRouter(t, &fakeProviderCredentialService{})

	req := httptest.NewRequest(http.MethodDelete,
		"/api/v1/organizations/"+orgID.String()+"/credentials/ai/"+credID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.Bytes(), "204 must carry no body")
}
