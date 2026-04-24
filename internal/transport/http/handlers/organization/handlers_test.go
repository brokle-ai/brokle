// Tests for the organization handler. The coverage focuses on the
// list endpoint's wire contract — CLAUDE.md gotcha #23 mandates the
// canonical `{data, pagination}` shape for every list endpoint, and
// the dashboard + SDK auth flows depend on it (/v1/organizations
// unwrapping via client.getPaginated). A regression to the old
// `{organizations, total, page, limit}` shape would break signup and
// login with "No organizations found for user", which is the exact
// bug these tests guard against.
package organization_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	orgDomain "brokle/internal/core/domain/organization"
	"brokle/internal/testing/humax"
	handler "brokle/internal/transport/http/handlers/organization"
	"brokle/internal/transport/http/httpctx"
)

// ---- fake services ---------------------------------------------------
//
// Embedded-interface stubs — only the methods the list-organizations
// handler actually calls are implemented. Any other interface method
// invoked during the test panics on the embedded nil interface, which
// is the desired signal (means the handler is calling something
// outside the tested surface).

type fakeOrganizationService struct {
	orgDomain.OrganizationService // embedded: unused methods panic

	listResp []*orgDomain.Organization
	listErr  error
}

func (f *fakeOrganizationService) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*orgDomain.Organization, error) {
	return f.listResp, f.listErr
}

type fakeMemberService struct {
	orgDomain.MemberService
}

type fakeInvitationService struct {
	orgDomain.InvitationService
}

type fakeSettingsService struct {
	orgDomain.OrganizationSettingsService
}

// ---- helpers ---------------------------------------------------------

func newTestAPI(t *testing.T, orgSvc orgDomain.OrganizationService) humatest.TestAPI {
	t.Helper()
	api := humax.NewAPI(t)
	handler.RegisterRoutes(api, orgSvc, fakeMemberService{}, fakeInvitationService{}, fakeSettingsService{}, slog.Default())
	return api
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

// ---- tests -----------------------------------------------------------

// TestListOrganizations_CanonicalShape — the list endpoint MUST emit
// the canonical `{data, pagination}` shape (CLAUDE.md gotcha #23).
// The old `{organizations, total, page, limit}` shape is a historical
// bug that broke frontend signup+login; this test pins the correct
// wire contract.
func TestListOrganizations_CanonicalShape(t *testing.T) {
	svc := &fakeOrganizationService{listResp: makeOrgs(2)}
	api := newTestAPI(t, svc)
	ctx := httpctx.WithUserID(context.Background(), uuid.New())

	resp := api.GetCtx(ctx, "/api/v1/organizations")
	require.Equal(t, http.StatusOK, resp.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &body))

	// Canonical keys present.
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

	// Legacy / wrong keys MUST NOT be present — this is the bug guard.
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
	api := newTestAPI(t, svc)
	ctx := httpctx.WithUserID(context.Background(), uuid.New())

	resp := api.GetCtx(ctx, "/api/v1/organizations")
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
	api := newTestAPI(t, svc)
	ctx := httpctx.WithUserID(context.Background(), uuid.New())

	resp := api.GetCtx(ctx, "/api/v1/organizations?page=2&limit=10")
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
