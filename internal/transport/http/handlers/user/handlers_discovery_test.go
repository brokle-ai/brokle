// Discovery-filter tests for /api/v1/users/me bootstrap. The
// permission resolution invariants live in the auth service tests
// (project_member_service_discovery_test.go); this file locks the
// transport-layer composition: given an org topology + scope tree,
// the response includes exactly the discoverable subset.
package user

import (
	"testing"
	"time"

	"github.com/google/uuid"

	orgDomain "brokle/internal/core/domain/organization"
	authService "brokle/internal/core/services/auth"
)

func TestMapOrgsWithProjects_DiscoveryFilter(t *testing.T) {
	orgID := uuid.New()
	projA := uuid.New()
	projB := uuid.New()
	projC := uuid.New()

	topology := []*orgDomain.OrganizationWithProjectsAndRole{{
		Organization: &orgDomain.Organization{
			ID:        orgID,
			Name:      "Acme",
			Plan:      "pro",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		RoleName: "viewer",
		Projects: []*orgDomain.Project{
			{ID: projA, Name: "A", OrganizationID: orgID, Status: "active"},
			{ID: projB, Name: "B", OrganizationID: orgID, Status: "active"},
			{ID: projC, Name: "C", OrganizationID: orgID, Status: "active"},
		},
	}}

	cases := []struct {
		name             string
		scopes           map[uuid.UUID]*authService.EffectiveOrgScopes
		expectedProjects []uuid.UUID
	}{
		{
			name: "owner_equivalent_full_access",
			scopes: map[uuid.UUID]*authService.EffectiveOrgScopes{
				orgID: {
					OrganizationID: orgID,
					Scopes:         []string{authService.PermOrgProjectsList},
					Projects: map[uuid.UUID]authService.EffectiveProjectScopes{
						projA: {ProjectID: projA, Scopes: []string{authService.PermProjectsRead}},
						projB: {ProjectID: projB, Scopes: []string{authService.PermProjectsRead}},
						projC: {ProjectID: projC, Scopes: []string{authService.PermProjectsRead}},
					},
				},
			},
			expectedProjects: []uuid.UUID{projA, projB, projC},
		},
		{
			name: "custom_org_only_role_no_project_access",
			scopes: map[uuid.UUID]*authService.EffectiveOrgScopes{
				orgID: {
					OrganizationID: orgID,
					Scopes:         []string{"members:read", "billing:read"},
					Projects: map[uuid.UUID]authService.EffectiveProjectScopes{
						projA: {ProjectID: projA, Scopes: []string{}},
						projB: {ProjectID: projB, Scopes: []string{}},
						projC: {ProjectID: projC, Scopes: []string{}},
					},
				},
			},
			expectedProjects: nil,
		},
		{
			name: "single_project_override_only",
			scopes: map[uuid.UUID]*authService.EffectiveOrgScopes{
				orgID: {
					OrganizationID: orgID,
					Scopes:         []string{},
					Projects: map[uuid.UUID]authService.EffectiveProjectScopes{
						projA: {ProjectID: projA, RoleName: "developer", Scopes: []string{authService.PermProjectsRead, "traces:read"}},
						projB: {ProjectID: projB, Scopes: []string{}},
						projC: {ProjectID: projC, Scopes: []string{}},
					},
				},
			},
			expectedProjects: []uuid.UUID{projA},
		},
		{
			name: "org_list_authority_overrides_missing_per_project_floor",
			scopes: map[uuid.UUID]*authService.EffectiveOrgScopes{
				orgID: {
					OrganizationID: orgID,
					Scopes:         []string{authService.PermOrgProjectsList},
					Projects: map[uuid.UUID]authService.EffectiveProjectScopes{
						// B carries no per-project perms; org-tier authority
						// still grants discovery.
						projA: {ProjectID: projA, Scopes: []string{authService.PermProjectsRead}},
						projB: {ProjectID: projB, Scopes: []string{}},
					},
				},
			},
			expectedProjects: []uuid.UUID{projA, projB, projC},
		},
		{
			name:             "aggregator_failure_emits_empty_projects",
			scopes:           map[uuid.UUID]*authService.EffectiveOrgScopes{},
			expectedProjects: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := mapOrgsWithProjects(topology, tc.scopes)
			if len(out) != 1 {
				t.Fatalf("expected 1 org in response, got %d", len(out))
			}
			got := projectIDs(out[0].Projects)
			if !sameSet(got, tc.expectedProjects) {
				t.Fatalf("project set mismatch:\n  got:  %v\n  want: %v", got, tc.expectedProjects)
			}
		})
	}
}

// TestMapOrgsWithProjects_FilterIsPerOrgIndependent verifies that
// discovery filtering applied to one org does not affect projects in
// another org under the same user (a custom role in org X must not
// leak / hide projects in org Y).
func TestMapOrgsWithProjects_FilterIsPerOrgIndependent(t *testing.T) {
	orgX := uuid.New()
	orgY := uuid.New()
	pX := uuid.New()
	pY := uuid.New()

	topology := []*orgDomain.OrganizationWithProjectsAndRole{
		{
			Organization: &orgDomain.Organization{ID: orgX, Name: "X", Plan: "pro"},
			RoleName:     "custom",
			Projects:     []*orgDomain.Project{{ID: pX, Name: "PX", OrganizationID: orgX}},
		},
		{
			Organization: &orgDomain.Organization{ID: orgY, Name: "Y", Plan: "pro"},
			RoleName:     "viewer",
			Projects:     []*orgDomain.Project{{ID: pY, Name: "PY", OrganizationID: orgY}},
		},
	}
	scopes := map[uuid.UUID]*authService.EffectiveOrgScopes{
		// org X: custom role with no project access → pX hidden.
		orgX: {
			OrganizationID: orgX,
			Scopes:         []string{"members:read"},
			Projects: map[uuid.UUID]authService.EffectiveProjectScopes{
				pX: {ProjectID: pX, Scopes: []string{}},
			},
		},
		// org Y: viewer-equivalent → pY visible.
		orgY: {
			OrganizationID: orgY,
			Scopes:         []string{authService.PermOrgProjectsList},
			Projects: map[uuid.UUID]authService.EffectiveProjectScopes{
				pY: {ProjectID: pY, Scopes: []string{authService.PermProjectsRead}},
			},
		},
	}

	out := mapOrgsWithProjects(topology, scopes)
	if len(out) != 2 {
		t.Fatalf("expected 2 orgs, got %d", len(out))
	}
	for _, o := range out {
		switch o.ID {
		case orgX:
			if len(o.Projects) != 0 {
				t.Fatalf("orgX must have empty projects under custom-role; got %d", len(o.Projects))
			}
		case orgY:
			if len(o.Projects) != 1 || o.Projects[0].ID != pY {
				t.Fatalf("orgY must include pY; got %v", projectIDs(o.Projects))
			}
		default:
			t.Fatalf("unexpected org id %s", o.ID)
		}
	}
}

func projectIDs(in []projectSummary) []uuid.UUID {
	if len(in) == 0 {
		return nil
	}
	out := make([]uuid.UUID, len(in))
	for i, p := range in {
		out[i] = p.ID
	}
	return out
}

func sameSet(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(a))
	for _, x := range a {
		seen[x] = struct{}{}
	}
	for _, y := range b {
		if _, ok := seen[y]; !ok {
			return false
		}
	}
	return true
}
