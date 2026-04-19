// Huma operation-layer test for the credentials handler. Exercises
// the `humax` harness pattern every migrated domain can copy:
//
//   - stand up a humatest.TestAPI via humax.NewAPI (installs the
//     APIResponse / ErrorResponse envelope override)
//   - inject the authenticated-user context the handlers read via
//     httpctx (in production this comes from middleware.RequireAuth)
//   - drive operations by path + method, inspect the envelope on both
//     success and error paths
//
// Reference template — mirror when adding coverage for other domains.
package credentials_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	analyticsDomain "brokle/internal/core/domain/analytics"
	credentialsDomain "brokle/internal/core/domain/credentials"
	"brokle/internal/testing/humax"
	handler "brokle/internal/transport/http/handlers/credentials"
	"brokle/internal/transport/http/httpctx"
	"brokle/pkg/response"
)

// ---- fake services ---------------------------------------------------
//
// Hand-rolled fakes rather than testify/mock — the surface is small
// and assertions care about exactly one or two methods per test.

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

// ---- helpers ---------------------------------------------------------

func newTestAPI(t *testing.T, svc credentialsDomain.ProviderCredentialService) humatest.TestAPI {
	t.Helper()
	api := humax.NewAPI(t)
	handler.RegisterRoutes(api, svc, fakeModelCatalogService{}, slog.Default())
	return api
}

// ---- tests -----------------------------------------------------------

// Happy path: a request with a populated auth context surfaces the
// service response. The handler's Body is passed through verbatim —
// the Brokle success-path wire shape is owned by the handler, not the
// error envelope. SchemaLinkTransformer decorates success responses
// with `$schema` + `Link: rel=describedBy` since the response type is
// the handler's DTO (not the statusError wrapper).
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

	api := newTestAPI(t, svc)
	ctx := httpctx.WithUserID(context.Background(), uuid.New())

	resp := api.GetCtx(ctx, "/api/v1/organizations/"+orgID.String()+"/credentials/ai")
	require.Equal(t, http.StatusOK, resp.Code)

	var items []credentialsDomain.ProviderCredentialResponse
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &items))
	require.Len(t, items, 1)
	assert.Equal(t, credID, items[0].ID)
	assert.Equal(t, "prod-openai", items[0].Name)
}

// Error path: a malformed UUID on the path triggers Huma's
// pipeline-level validation. The response MUST be our ErrorResponse
// envelope at application/json content-type — not Huma's default
// problem+json shape. Field access uses natural Go names
// (env.Error.Message), mirroring the JSON wire keys exactly.
//
// Note on $schema decoration: error responses are wrapped in the
// internal statusError type at runtime, while the OpenAPI spec's
// response Schema ref points at ErrorResponse (via SchemaProvider).
// SchemaLinkTransformer keys its decoration table by the ref type, so
// the runtime-vs-registered type mismatch means errors skip
// decoration. This is the intended trade-off of the clean DTO split;
// SDK consumers read /openapi.json for error shapes directly.
func TestListCredentials_InvalidOrgID_EmitsErrorEnvelope(t *testing.T) {
	api := newTestAPI(t, &fakeProviderCredentialService{})
	ctx := httpctx.WithUserID(context.Background(), uuid.New())

	resp := api.GetCtx(ctx, "/api/v1/organizations/not-a-uuid/credentials/ai")
	require.Equal(t, http.StatusUnprocessableEntity, resp.Code)
	assert.Equal(t, "application/json", resp.Header().Get("Content-Type"),
		"error responses must stay application/json — problem+json regression")

	var env response.ErrorResponse
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &env))
	assert.False(t, env.Success, "ErrorResponse.success must be false on error paths")
	require.NotNil(t, env.Error, "ErrorResponse.error must be populated on error paths")
	assert.NotEmpty(t, env.Error.Type, "error.type must be set (closed enum)")
	assert.NotEmpty(t, env.Error.Message, "error.message must be set (human-readable)")

	// Per-field Huma validation details must survive the lift from
	// `errs ...error` into ErrorResponse.Error.Errors.
	require.NotEmpty(t, env.Error.Errors, "per-field Errors[] must be populated from huma.ErrorDetailer")
	found := false
	for _, d := range env.Error.Errors {
		if d.Location == "path.orgId" {
			found = true
			break
		}
	}
	assert.True(t, found, "per-field Errors[] must include an entry for the offending path param")

	// $schema decoration is intentionally absent from error bodies —
	// guard the invariant so future changes don't silently reintroduce
	// noise (e.g. by dropping SchemaProvider on statusError).
	var probe map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &probe))
	_, hasSchema := probe["$schema"]
	assert.False(t, hasSchema,
		"$schema must be absent on error responses — statusError must stay out of SchemaLinkTransformer's type table")
}

// OpenAPI spec regression: every operation's error responses must ref
// `#/components/schemas/ErrorResponse`, never `statusError` or any other
// wrapper leakage. SDK codegen (hey-api, openapi-python-client, fern)
// reads these refs directly; if SchemaProvider stops working or the
// wrapper gets registered as its own component, error types in
// generated SDKs will break in subtle ways. This test guards the
// invariant at the spec level.
func TestOpenAPI_ErrorResponsesRefErrorResponseComponent(t *testing.T) {
	api := newTestAPI(t, &fakeProviderCredentialService{})

	// Components: ErrorResponse must be registered; statusError must not.
	spec := api.OpenAPI()
	require.NotNil(t, spec)
	require.NotNil(t, spec.Components)
	require.NotNil(t, spec.Components.Schemas)

	names := make([]string, 0)
	for name := range spec.Components.Schemas.Map() {
		names = append(names, name)
	}
	assert.Contains(t, names, "ErrorResponse",
		"OpenAPI components must include ErrorResponse schema — SchemaProvider regression")
	for _, n := range names {
		assert.NotContains(t, strings.ToLower(n), "statuserror",
			"statusError wrapper must not leak into OpenAPI components: saw %q", n)
	}

	// Every registered operation's 4xx/5xx responses must ref ErrorResponse.
	const expectedRef = "#/components/schemas/ErrorResponse"
	checked := 0
	for path, pi := range spec.Paths {
		for method, op := range map[string]*struct {
			Responses map[string]struct {
				Content map[string]struct {
					Schema *struct{ Ref string }
				}
			}
		}{} {
			_ = path
			_ = method
			_ = op
		}
		// The spec.Paths mapping exposes *huma.PathItem with typed method
		// fields; a simple serialize + reparse is the most robust way to
		// walk every operation regardless of method name.
		raw, err := json.Marshal(pi)
		require.NoError(t, err)
		var pathItem map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(raw, &pathItem))
		for method, opRaw := range pathItem {
			if method == "parameters" || method == "summary" || method == "description" {
				continue
			}
			var op struct {
				Responses map[string]struct {
					Content map[string]struct {
						Schema struct {
							Ref string `json:"$ref"`
						} `json:"schema"`
					} `json:"content"`
				} `json:"responses"`
			}
			if err := json.Unmarshal(opRaw, &op); err != nil {
				// not an operation object (e.g. x-* extensions) — skip
				continue
			}
			for code, resp := range op.Responses {
				// Error responses — every 4xx/5xx plus "default".
				if code == "default" || (len(code) > 0 && (code[0] == '4' || code[0] == '5')) {
					for ct, mt := range resp.Content {
						if mt.Schema.Ref == "" {
							continue
						}
						assert.Equal(t, expectedRef, mt.Schema.Ref,
							"%s %s %s %s: error-response $ref must point at ErrorResponse, got %q",
							method, path, code, ct, mt.Schema.Ref)
						checked++
					}
				}
			}
		}
	}
	require.Greater(t, checked, 0, "must have inspected at least one error-response ref")
}
