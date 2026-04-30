package auth

import (
	"testing"

	"github.com/google/uuid"
)

// TestCanDiscoverProject_OrgListAuthority asserts that an org-level
// `org_projects:list` permission is sufficient to discover every
// project, including ones the user has no project-tier perms on
// (GitHub org-admin shape — admins see the full list even before
// being granted explicit per-project access).
func TestCanDiscoverProject_OrgListAuthority(t *testing.T) {
	projA := uuid.New()
	projB := uuid.New()
	unknown := uuid.New()
	o := &EffectiveOrgScopes{
		OrganizationID: uuid.New(),
		Scopes:         []string{PermOrgProjectsList},
		Projects: map[uuid.UUID]EffectiveProjectScopes{
			projA: {ProjectID: projA, Scopes: []string{PermProjectsRead}},
			projB: {ProjectID: projB, Scopes: []string{}}, // no per-project perms
		},
	}
	cases := []struct {
		name string
		id   uuid.UUID
	}{
		{"project_with_floor", projA},
		{"project_without_floor", projB},
		{"project_unknown_to_aggregator", unknown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !o.CanDiscoverProject(tc.id) {
				t.Fatalf("expected discoverable; org_projects:list grants visibility")
			}
		})
	}
}

// TestCanDiscoverProject_PerProjectFloor asserts that without
// `org_projects:list`, the per-project `projects:read` floor scope
// gates discovery.
func TestCanDiscoverProject_PerProjectFloor(t *testing.T) {
	projWithFloor := uuid.New()
	projWithContentOnly := uuid.New() // hypothetical pre-floor-injection state
	projUnknown := uuid.New()
	o := &EffectiveOrgScopes{
		OrganizationID: uuid.New(),
		Scopes:         []string{}, // no org_projects:list
		Projects: map[uuid.UUID]EffectiveProjectScopes{
			projWithFloor:       {ProjectID: projWithFloor, Scopes: []string{PermProjectsRead, "traces:read"}},
			projWithContentOnly: {ProjectID: projWithContentOnly, Scopes: []string{"traces:read"}},
		},
	}

	if !o.CanDiscoverProject(projWithFloor) {
		t.Fatalf("project with projects:read should be discoverable")
	}
	if o.CanDiscoverProject(projWithContentOnly) {
		t.Fatalf("project without projects:read floor must not be discoverable (defensive: shouldn't happen post-auto-injection)")
	}
	if o.CanDiscoverProject(projUnknown) {
		t.Fatalf("project absent from scope tree must not be discoverable")
	}
}

// TestCanDiscoverProject_NoAccess asserts that a user with no relevant
// permissions sees no projects.
func TestCanDiscoverProject_NoAccess(t *testing.T) {
	o := &EffectiveOrgScopes{
		OrganizationID: uuid.New(),
		Scopes:         []string{"members:read", "billing:read"}, // non-discovery org-tier perms
		Projects:       map[uuid.UUID]EffectiveProjectScopes{},
	}
	if o.CanDiscoverProject(uuid.New()) {
		t.Fatalf("custom org-only role without discovery perms must produce empty discoverable set")
	}
}

// TestCanDiscoverProject_NilReceiver guards the handler-side path
// where the aggregator may not have an entry for an org (e.g.,
// resolution failure → empty map). The predicate must return false
// without panicking.
func TestCanDiscoverProject_NilReceiver(t *testing.T) {
	var o *EffectiveOrgScopes
	if o.CanDiscoverProject(uuid.New()) {
		t.Fatalf("nil receiver must return false")
	}
}
