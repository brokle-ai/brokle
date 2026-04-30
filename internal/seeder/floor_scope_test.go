package seeder_test

// Floor-scope invariant test: every built-in role in seeds/roles.yaml
// that grants ANY project-scoped permission MUST also grant
// `projects:read`. Mirrors GitHub's `metadata: read` floor-scope
// mechanic — a role with `datasets:read` (or any other project-scoped
// permission) but without `projects:read` is structurally invalid:
// dashboard navigation hits `GET /api/v1/projects/{projectId}` to
// resolve project context for every page load, and that endpoint is
// gated by `projects:read`. Without the floor, the user gets 403 on
// navigation and can't reach the feature pages their other
// permissions grant.
//
// The set of "project-scoped permissions" is derived dynamically from
// `internal/server/routes.go`: every `middleware.RequirePermission(...)`
// (and `RequireAny|All`) call that appears inside the
// `r.Route("/api/v1/projects/{projectId}", func(r chi.Router) { ... })`
// block. This makes routes.go the single source of truth — adding a
// new project-scoped route automatically extends the floor-scope set.
//
// See CLAUDE.md 2026-04-30 (project-rbac) for the rationale.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// projectRouteBlockMarker is the literal call that opens the project-
// scoped chi.Route block in internal/server/routes.go. The test fails
// loudly if this marker is renamed so we don't silently lose the
// project-scope detection.
const projectRouteBlockMarker = `r.Route("/api/v1/projects/{projectId}",`

// extractProjectScopedPermissions scans routes.go for the project route
// block and returns every permission string referenced by a
// RequirePermission/Any/All call inside it. Brace-balanced extraction so
// nested chi.Group / chi.Route blocks within the project tree are
// included, and sibling routes outside the project tree are excluded.
func extractProjectScopedPermissions(t *testing.T, routesPath string) []string {
	t.Helper()
	body, err := os.ReadFile(routesPath)
	if err != nil {
		t.Fatalf("read %s: %v", routesPath, err)
	}
	src := string(body)

	start := strings.Index(src, projectRouteBlockMarker)
	if start < 0 {
		t.Fatalf("project route marker %q not found in %s — has the route been renamed?",
			projectRouteBlockMarker, routesPath)
	}

	// Find the opening `{` of the route CALLBACK, not the `{projectId}`
	// chi placeholder inside the URL pattern. The chi.Router callback
	// signature is the unambiguous anchor.
	const callbackAnchor = `func(r chi.Router) {`
	cb := strings.Index(src[start:], callbackAnchor)
	if cb < 0 {
		t.Fatalf("project route marker present but no `%s` callback follows", callbackAnchor)
	}
	openBrace := start + cb + len(callbackAnchor) - 1 // index of the `{` itself

	depth := 1
	end := -1
	for i := openBrace + 1; i < len(src) && end < 0; i++ {
		switch src[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = i
			}
		}
	}
	if end < 0 {
		t.Fatalf("project route block has no matching close brace")
	}

	block := []byte(src[openBrace:end])
	seen := make(map[string]struct{})
	for _, m := range permissionLiteralRE.FindAllSubmatch(block, -1) {
		for _, q := range quotedRE.FindAllSubmatch(m[1], -1) {
			seen[string(q[1])] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// loadSeededRoles parses seeds/roles.yaml into a map of role-name →
// permission set. Skips roles whose `permissions` list is empty.
func loadSeededRoles(t *testing.T, yamlPath string) map[string]map[string]struct{} {
	t.Helper()
	body, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("read %s: %v", yamlPath, err)
	}
	var doc struct {
		Roles []struct {
			Name        string   `yaml:"name"`
			Permissions []string `yaml:"permissions"`
		} `yaml:"roles"`
	}
	if err := yaml.Unmarshal(body, &doc); err != nil {
		t.Fatalf("parse %s: %v", yamlPath, err)
	}
	out := make(map[string]map[string]struct{}, len(doc.Roles))
	for _, r := range doc.Roles {
		set := make(map[string]struct{}, len(r.Permissions))
		for _, p := range r.Permissions {
			set[p] = struct{}{}
		}
		out[r.Name] = set
	}
	return out
}

// loadSeededPermissionScopes parses seeds/permissions.yaml into a map
// of permission name → declared scope ("project" / "organization").
func loadSeededPermissionScopes(t *testing.T, yamlPath string) map[string]string {
	t.Helper()
	body, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("read %s: %v", yamlPath, err)
	}
	var doc struct {
		Permissions []struct {
			Name  string `yaml:"name"`
			Scope string `yaml:"scope"`
		} `yaml:"permissions"`
	}
	if err := yaml.Unmarshal(body, &doc); err != nil {
		t.Fatalf("parse %s: %v", yamlPath, err)
	}
	out := make(map[string]string, len(doc.Permissions))
	for _, p := range doc.Permissions {
		out[p.Name] = p.Scope
	}
	return out
}

// sharedVerbResources lists org-tier resource prefixes that LEGITIMATELY
// decorate routes inside the project URL tree. The verb resolves at the
// org tier (org-rooted GOVERNANCE — managing role bindings, admin
// configuration), but the routing surface lives under
// /api/v1/projects/{projectId} for URL-shape reasons (the resource
// being administered is project-scoped data, even though the capability
// to administer is an org-rooted decision).
//
// Currently:
//
//   - `org_members` decorates /api/v1/projects/{projectId}/members*
//     (project_members override management). Tagging these verbs
//     project-tier would subject them to the OVERRIDE-aware resolver
//     and create a lockout: a restrictive project_members override
//     would strip the caller's ability to UNDO it. Org-tier resolution
//     is required so role-binding management remains an org admin
//     prerogative regardless of any project_members override the admin
//     might carry. See CLAUDE.md 2026-04-30 (project-rbac).
//
// History: rounds 8 / 19 used `sharedVerbResources` as a family-level
// workaround for the dual-scope problem (verbs that surface at both
// org and project tiers tagged organization-wide). Round 20 retired
// that workaround by minting per-tier verb pairs (`org_projects:*` +
// `projects:*`, `org_members:*` + `members:*`). Round 22 emptied this
// map. Round N (this one) re-populates it for a structurally different
// reason: an org-tier verb that legitimately decorates a project-URL
// route because the CAPABILITY is org-rooted even though the URL
// surface is project-scoped. The exemption is correct here because the
// per-tier mint puts the verb at the right tier; the test just needs
// to know the routing accident.
//
// Add a new prefix here only when (a) the verb is org-rooted and tagged
// `scope: organization` AND (b) it legitimately decorates a route under
// the project URL tree because the URL surface is project-scoped data
// administered via an org-rooted capability. Document the rationale
// inline.
var sharedVerbResources = map[string]struct{}{
	"org_members": {},
}

// TestPermissionScopesMatchRoutes asserts the data-driven floor-scope
// source-of-truth invariant: every permission appearing inside the
// project route block in `routes.go` MUST be tagged `scope: project` in
// `seeds/permissions.yaml`, UNLESS its resource prefix is in
// `sharedVerbResources` (in which case `scope: organization` is correct
// to prevent floor-scope over-injection).
//
// This is the data-side complement of TestRolesCarryProjectsReadFloorScope
// — the seeder reads the YAML scope into the `permissions.scope_level`
// column, and the role service queries that column to drive auto-
// injection of `projects:read`. If YAML and routes.go drift, auto-
// injection silently misclassifies and the floor breaks.
func TestPermissionScopesMatchRoutes(t *testing.T) {
	root := repoRoot(t)

	projectRoutePerms := extractProjectScopedPermissions(t, filepath.Join(root, "internal", "server", "routes.go"))
	if len(projectRoutePerms) == 0 {
		t.Fatalf("extracted zero project-scoped permissions from routes.go")
	}

	scopes := loadSeededPermissionScopes(t, filepath.Join(root, "seeds", "permissions.yaml"))
	if len(scopes) == 0 {
		t.Fatalf("loaded zero permissions from seeds/permissions.yaml")
	}

	// 1. Every permission in routes.go's project block must be tagged
	//    scope: project in YAML, UNLESS the resource prefix is in the
	//    shared-verb exemption set (org-rooted family that also surfaces
	//    at project URLs — e.g. `members:*`).
	for _, p := range projectRoutePerms {
		got, present := scopes[p]
		if !present {
			t.Errorf("routes.go uses permission %q under the project route block but it is missing from seeds/permissions.yaml", p)
			continue
		}
		resource, _, _ := strings.Cut(p, ":")
		_, isShared := sharedVerbResources[resource]
		if isShared {
			if got != "organization" {
				t.Errorf("permission %q is in the shared-verb family %q (org-rooted; appears at project URLs via override semantics) — must be `scope: organization` to prevent floor-scope over-injection, got `scope: %s`", p, resource, got)
			}
			continue
		}
		if got != "project" {
			t.Errorf("permission %q is gated under the project route block in routes.go but is tagged `scope: %s` in seeds/permissions.yaml — must be `scope: project` (or add %q to sharedVerbResources if it's intentionally org-rooted)", p, got, resource)
		}
	}

	// 2. Every YAML scope must be one of the known values.
	for name, scope := range scopes {
		switch scope {
		case "project", "organization":
			// ok
		default:
			t.Errorf("permission %q has unknown scope %q in seeds/permissions.yaml — must be `project` or `organization`", name, scope)
		}
	}
}

// TestRolesCarryProjectsReadFloorScope asserts the GitHub-style floor-
// scope invariant: a built-in role that grants any project-scoped
// permission must also grant `projects:read`.
func TestRolesCarryProjectsReadFloorScope(t *testing.T) {
	root := repoRoot(t)

	projectScoped := extractProjectScopedPermissions(t, filepath.Join(root, "internal", "server", "routes.go"))
	if len(projectScoped) == 0 {
		t.Fatalf("extracted zero project-scoped permissions from routes.go — block detection may be broken")
	}

	// Sanity: `projects:read` itself must be in the project-scoped set
	// (it's the floor; if it were excluded the test would be vacuous).
	floorPresent := false
	for _, p := range projectScoped {
		if p == "projects:read" {
			floorPresent = true
			break
		}
	}
	if !floorPresent {
		t.Fatalf("`projects:read` is the floor-scope permission but is not present in the project-scoped set extracted from routes.go (have: %v)", projectScoped)
	}

	roles := loadSeededRoles(t, filepath.Join(root, "seeds", "roles.yaml"))
	if len(roles) == 0 {
		t.Fatalf("loaded zero roles from seeds/roles.yaml")
	}

	scopedSet := make(map[string]struct{}, len(projectScoped))
	for _, p := range projectScoped {
		scopedSet[p] = struct{}{}
	}

	for roleName, perms := range roles {
		hasAnyProjectScoped := false
		for p := range perms {
			if _, ok := scopedSet[p]; ok && p != "projects:read" {
				hasAnyProjectScoped = true
				break
			}
		}
		if !hasAnyProjectScoped {
			continue
		}
		if _, ok := perms["projects:read"]; !ok {
			t.Errorf("role %q grants project-scoped permissions but does NOT include the `projects:read` floor — dashboard navigation will 403", roleName)
		}
	}
}

// TestDiscoveryGatePermissionsExist locks the contract between the
// /users/me bootstrap discovery filter and seeds/permissions.yaml.
//
// The handler's CanDiscoverProject predicate
// (internal/core/services/auth/project_member_service.go) keys off two
// named constants — PermOrgProjectsList ("org_projects:list") and
// PermProjectsRead ("projects:read"). If either is renamed in YAML
// without updating the const, the predicate silently filters
// everything out (or nothing) without a compile error. This test
// fails loudly on that drift.
func TestDiscoveryGatePermissionsExist(t *testing.T) {
	root := repoRoot(t)
	scopes := loadSeededPermissionScopes(t, filepath.Join(root, "seeds", "permissions.yaml"))

	// Mirror the constants in
	// internal/core/services/auth/project_member_service.go. Hardcoded
	// strings here are deliberate — the test exists to catch drift
	// between the const VALUE and the YAML.
	wantOrgList := "org_projects:list"
	wantProjectsRead := "projects:read"

	if scope, ok := scopes[wantOrgList]; !ok {
		t.Errorf("discovery-gate permission %q (PermOrgProjectsList) is missing from seeds/permissions.yaml — /users/me filter would deny org-tier discovery for everyone", wantOrgList)
	} else if scope != "organization" {
		t.Errorf("discovery-gate permission %q must be `scope: organization` (org-tier authority); got %q", wantOrgList, scope)
	}

	if scope, ok := scopes[wantProjectsRead]; !ok {
		t.Errorf("discovery-gate permission %q (PermProjectsRead) is missing from seeds/permissions.yaml — /users/me filter would deny per-project discovery for everyone", wantProjectsRead)
	} else if scope != "project" {
		t.Errorf("discovery-gate permission %q must be `scope: project` (per-project floor); got %q", wantProjectsRead, scope)
	}
}

// TestOrgProjectsListImpliesProjectsRead pins the seeded-roles
// invariant: any built-in role granting `org_projects:list` or
// `org_projects:admin` MUST also grant `projects:read`. These org-tier
// verbs let a user list/admin every project in the org, but the
// per-project route guard (`GET /api/v1/projects/{projectId}` requires
// `projects:read`) means the user 403s on click-through without the
// floor.
//
// This is the seeded-data complement to
// `RoleService.autoInjectFloorScope` (round 26 trigger extension):
// auto-injection enforces the invariant for runtime-created custom
// roles, this test enforces it for the YAML-seeded built-ins.
//
// Hardcoded strings here are deliberate — the test exists to catch
// drift between the YAML and the runtime trigger map.
func TestOrgProjectsListImpliesProjectsRead(t *testing.T) {
	root := repoRoot(t)
	roles := loadSeededRoles(t, filepath.Join(root, "seeds", "roles.yaml"))

	const floor = "projects:read"
	triggerVerbs := []string{"org_projects:list", "org_projects:admin"}

	for roleName, perms := range roles {
		for _, trigger := range triggerVerbs {
			if _, has := perms[trigger]; !has {
				continue
			}
			if _, hasFloor := perms[floor]; !hasFloor {
				t.Errorf("role %q grants %q but does NOT include the %q floor — users assigned this role will 403 on `GET /api/v1/projects/{projectId}` despite the org-tier authority", roleName, trigger, floor)
			}
		}
	}
}
