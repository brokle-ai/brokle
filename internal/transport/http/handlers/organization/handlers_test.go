// Tests for the organization handler. Coverage focuses on the list
// endpoint's wire contract — CLAUDE.md gotcha #23 mandates the
// canonical `{data, pagination}` shape for every list endpoint. The
// old `{organizations, total, page, limit}` shape is a historical
// bug that broke frontend signup+login; these tests pin the correct
// wire contract.
package organization_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	orgDomain "brokle/internal/core/domain/organization"
	organizationService "brokle/internal/core/services/organization"
	handler "brokle/internal/transport/http/handlers/organization"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/response"
)

// ---- fake services ---------------------------------------------------
//
// Embedded-interface stubs — only the methods the list-organizations
// handler actually calls are implemented. Any other interface method
// invoked during the test panics on the embedded nil interface, which
// is the desired signal (means the handler is calling something
// outside the tested surface).

type fakeOrganizationService struct {
	*organizationService.OrganizationService // embedded: unused methods panic

	listResp []*orgDomain.Organization
	listErr  error
}

func (f *fakeOrganizationService) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*orgDomain.Organization, error) {
	return f.listResp, f.listErr
}

type fakeMemberService struct {
	*organizationService.MemberService
}

type fakeInvitationService struct {
	*organizationService.InvitationService
}

type fakeSettingsService struct {
	*organizationService.OrganizationSettingsService
}

// ---- helpers ---------------------------------------------------------

// newTestRouter mounts the organization handler on a bare chi router
// with a user-context injector standing in for the production
// RequireAuth middleware. Mounts both halves of the post-split
// handler:
//   - RegisterListRoutes — top-level /api/v1/organizations + invitations
//   - RegisterOrgRoutes  — relative paths under /api/v1/organizations/{orgId}
//     (a stand-in for RequireOrganizationAccess pins orgID into ctx).
func newTestRouter(t *testing.T, orgSvc handler.OrganizationService) (*chi.Mux, uuid.UUID) {
	t.Helper()
	userID := uuid.New()
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := httpctx.WithUserID(req.Context(), userID)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	h := handler.New(orgSvc, fakeMemberService{}, fakeInvitationService{}, fakeSettingsService{}, slog.Default())

	// Mirror list/invitation routes from internal/server/routes.go.
	r.Get("/api/v1/organizations", h.ListOrganizations)
	r.Post("/api/v1/organizations", h.CreateOrganization)
	r.Get("/api/v1/invitations", h.ListUserInvitations)
	r.Get("/api/v1/invitations/validate/{token}", h.ValidateInvitationToken)
	r.Post("/api/v1/invitations/accept", h.AcceptInvitation)
	r.Post("/api/v1/invitations/decline", h.DeclineInvitation)

	r.Route("/api/v1/organizations/{orgId}", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				orgID, err := uuid.Parse(chi.URLParam(req, "orgId"))
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
		// Mirror org-scoped routes from internal/server/routes.go.
		r.Get("/", h.GetOrganization)
		r.Patch("/", h.UpdateOrganization)
		r.Delete("/", h.DeleteOrganization)
		r.Get("/members", h.ListMembers)
		r.Delete("/members/{userId}", h.RemoveMember)
		r.Post("/invitations", h.CreateInvitation)
		r.Get("/invitations", h.ListPendingInvitations)
		r.Post("/invitations/{invitationId}/resend", h.ResendInvitation)
		r.Delete("/invitations/{invitationId}", h.RevokeInvitation)
		r.Get("/settings/", h.ListSettings)
		r.Post("/settings/", h.CreateSetting)
		r.Get("/settings/{key}", h.GetSetting)
		r.Put("/settings/{key}", h.UpdateSetting)
		r.Delete("/settings/{key}", h.DeleteSetting)
	})
	return r, userID
}

func makeOrgs(n int) []*orgDomain.Organization {
	out := make([]*orgDomain.Organization, n)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		out[i] = &orgDomain.Organization{
			ID:                 uuid.New(),
			Name:               "org-" + uuid.New().String()[:8],
			Plan:               "free",
			SubscriptionStatus: "active",
			CreatedAt:          base.Add(time.Duration(i) * time.Hour),
			UpdatedAt:          base.Add(time.Duration(i) * time.Hour),
		}
	}
	return out
}

func fetch(t *testing.T, r *chi.Mux, url string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// ---- tests -----------------------------------------------------------

// TestListOrganizations_CanonicalShape — the list endpoint MUST emit
// the canonical `{data, pagination}` shape (CLAUDE.md gotcha #23).
// The old `{organizations, total, page, limit}` shape is a historical
// bug that broke frontend signup+login; this test pins the correct
// wire contract.
func TestListOrganizations_CanonicalShape(t *testing.T) {
	svc := &fakeOrganizationService{listResp: makeOrgs(2)}
	r, _ := newTestRouter(t, svc)

	resp := fetch(t, r, "/api/v1/organizations")
	require.Equal(t, http.StatusOK, resp.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &body))

	data, hasData := body["data"].([]any)
	require.True(t, hasData, "response must expose `data` array (canonical list shape)")
	assert.Len(t, data, 2)

	pagination, hasPagination := body["pagination"].(map[string]any)
	require.True(t, hasPagination, "response must expose `pagination` object")
	assert.Equal(t, float64(1), pagination["page"])
	assert.Equal(t, float64(20), pagination["limit"])
	assert.Equal(t, float64(2), pagination["total"])
	assert.Equal(t, float64(1), pagination["total_pages"])
	assert.Equal(t, false, pagination["has_next"])
	assert.Equal(t, false, pagination["has_prev"])

	_, hasLegacyOrgs := body["organizations"]
	assert.False(t, hasLegacyOrgs, "response must NOT contain the legacy `organizations` key (CLAUDE.md gotcha #23 violation)")
	_, hasFlatTotal := body["total"]
	assert.False(t, hasFlatTotal, "pagination metadata must live under `pagination`, not at the top level")
}

// TestListOrganizations_Empty — empty membership returns `data: []`,
// not null. Frontend's client.getPaginated treats missing pagination
// as a hard error; a null data array would break consumers.
func TestListOrganizations_Empty(t *testing.T) {
	svc := &fakeOrganizationService{listResp: []*orgDomain.Organization{}}
	r, _ := newTestRouter(t, svc)

	resp := fetch(t, r, "/api/v1/organizations")
	require.Equal(t, http.StatusOK, resp.Code)

	var body struct {
		Data       []map[string]any `json:"data"`
		Pagination map[string]any   `json:"pagination"`
	}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &body))

	assert.NotNil(t, body.Data, "empty data must serialise as [] not null")
	assert.Len(t, body.Data, 0)
	require.NotNil(t, body.Pagination)
	assert.Equal(t, float64(0), body.Pagination["total"])
	assert.Equal(t, float64(0), body.Pagination["total_pages"])
	assert.Equal(t, false, body.Pagination["has_next"])
	assert.Equal(t, false, body.Pagination["has_prev"])
}

// TestListOrganizations_Pagination — pagination math is surfaced
// correctly through response.BuildPagination for a multi-page dataset.
func TestListOrganizations_Pagination(t *testing.T) {
	svc := &fakeOrganizationService{listResp: makeOrgs(30)}
	r, _ := newTestRouter(t, svc)

	resp := fetch(t, r, "/api/v1/organizations?page=2&limit=10")
	require.Equal(t, http.StatusOK, resp.Code)

	var body struct {
		Data       []map[string]any `json:"data"`
		Pagination map[string]any   `json:"pagination"`
	}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &body))

	assert.Len(t, body.Data, 10, "page 2 of 30 with limit=10 must return 10 items")
	require.NotNil(t, body.Pagination)
	assert.Equal(t, float64(2), body.Pagination["page"])
	assert.Equal(t, float64(10), body.Pagination["limit"])
	assert.Equal(t, float64(30), body.Pagination["total"])
	assert.Equal(t, float64(3), body.Pagination["total_pages"])
	assert.Equal(t, true, body.Pagination["has_next"])
	assert.Equal(t, true, body.Pagination["has_prev"])
}
