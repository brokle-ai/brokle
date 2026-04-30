package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	appErrors "brokle/pkg/errors"
)

// fakePermissionRepo is a minimal stub that implements just enough of
// PermissionRepository for the autoInjectFloorScope behaviour tests.
// Other methods panic if called — keeps the surface honest.
type fakePermissionRepo struct {
	byID   map[uuid.UUID]*authDomain.Permission
	byName map[string]*authDomain.Permission
}

func (f *fakePermissionRepo) GetByID(ctx context.Context, id uuid.UUID) (*authDomain.Permission, error) {
	if p, ok := f.byID[id]; ok {
		return p, nil
	}
	return nil, appErrors.NotFound("permission")
}

func (f *fakePermissionRepo) GetByName(ctx context.Context, name string) (*authDomain.Permission, error) {
	if p, ok := f.byName[name]; ok {
		return p, nil
	}
	return nil, appErrors.NotFound("permission")
}

// Unimplemented — must not be invoked by autoInjectFloorScope.
func (f *fakePermissionRepo) Create(context.Context, *authDomain.Permission) error {
	panic("unexpected Create")
}
func (f *fakePermissionRepo) Update(context.Context, *authDomain.Permission) error {
	panic("unexpected Update")
}
func (f *fakePermissionRepo) Delete(context.Context, uuid.UUID) error { panic("unexpected Delete") }
func (f *fakePermissionRepo) GetByResourceAction(context.Context, string, string) (*authDomain.Permission, error) {
	panic("unexpected GetByResourceAction")
}
func (f *fakePermissionRepo) GetAllPermissions(context.Context) ([]*authDomain.Permission, error) {
	panic("unexpected GetAllPermissions")
}
func (f *fakePermissionRepo) GetByResource(context.Context, string) ([]*authDomain.Permission, error) {
	panic("unexpected GetByResource")
}
func (f *fakePermissionRepo) GetByResourceActions(context.Context, []string) ([]*authDomain.Permission, error) {
	panic("unexpected GetByResourceActions")
}
func (f *fakePermissionRepo) GetByScopeLevel(context.Context, authDomain.ScopeLevel) ([]*authDomain.Permission, error) {
	panic("unexpected GetByScopeLevel")
}
func (f *fakePermissionRepo) GetByScopeLevelAndCategory(context.Context, authDomain.ScopeLevel, string) ([]*authDomain.Permission, error) {
	panic("unexpected GetByScopeLevelAndCategory")
}
func (f *fakePermissionRepo) GetByNames(context.Context, []string) ([]*authDomain.Permission, error) {
	panic("unexpected GetByNames")
}
func (f *fakePermissionRepo) GetByCategory(context.Context, string) ([]*authDomain.Permission, error) {
	panic("unexpected GetByCategory")
}
func (f *fakePermissionRepo) GetAvailableResources(context.Context) ([]string, error) {
	panic("unexpected GetAvailableResources")
}
func (f *fakePermissionRepo) GetActionsForResource(context.Context, string) ([]string, error) {
	panic("unexpected GetActionsForResource")
}
func (f *fakePermissionRepo) GetAvailableCategories(context.Context) ([]string, error) {
	panic("unexpected GetAvailableCategories")
}
func (f *fakePermissionRepo) GetPermissionsByRoleID(context.Context, uuid.UUID) ([]*authDomain.Permission, error) {
	panic("unexpected GetPermissionsByRoleID")
}
func (f *fakePermissionRepo) PermissionExists(context.Context, string, string) (bool, error) {
	panic("unexpected PermissionExists")
}
func (f *fakePermissionRepo) BulkPermissionExists(context.Context, []string) (map[string]bool, error) {
	panic("unexpected BulkPermissionExists")
}
func (f *fakePermissionRepo) BulkCreate(context.Context, []*authDomain.Permission) error {
	panic("unexpected BulkCreate")
}

// newRoleServiceForFloorScopeTest builds a RoleService whose only live
// dependency is the fake permission repo — the role/role-permission
// repos are nil because autoInjectFloorScope never touches them.
func newRoleServiceForFloorScopeTest(perms map[string]*authDomain.Permission) *RoleService {
	byID := make(map[uuid.UUID]*authDomain.Permission, len(perms))
	for _, p := range perms {
		byID[p.ID] = p
	}
	return &RoleService{
		permissionRepo: &fakePermissionRepo{byID: byID, byName: perms},
	}
}

func mkPerm(name string, scope authDomain.ScopeLevel) *authDomain.Permission {
	return &authDomain.Permission{
		ID:         uuid.New(),
		Name:       name,
		ScopeLevel: scope,
	}
}

func TestAutoInjectFloorScope_InjectsWhenProjectScopedPresentAndFloorMissing(t *testing.T) {
	floor := mkPerm("projects:read", authDomain.ScopeLevelProject)
	datasets := mkPerm("datasets:read", authDomain.ScopeLevelProject)
	billing := mkPerm("billing:read", authDomain.ScopeLevelOrganization)

	svc := newRoleServiceForFloorScopeTest(map[string]*authDomain.Permission{
		floor.Name:    floor,
		datasets.Name: datasets,
		billing.Name:  billing,
	})

	in := []uuid.UUID{datasets.ID, billing.ID}
	out, err := svc.autoInjectFloorScope(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != len(in)+1 {
		t.Fatalf("expected floor injected (len %d → %d), got len %d", len(in), len(in)+1, len(out))
	}
	if out[len(out)-1] != floor.ID {
		t.Errorf("expected floor permission %s appended at the end, got %s", floor.ID, out[len(out)-1])
	}
}

func TestAutoInjectFloorScope_NoOpWhenFloorAlreadyPresent(t *testing.T) {
	floor := mkPerm("projects:read", authDomain.ScopeLevelProject)
	datasets := mkPerm("datasets:read", authDomain.ScopeLevelProject)

	svc := newRoleServiceForFloorScopeTest(map[string]*authDomain.Permission{
		floor.Name:    floor,
		datasets.Name: datasets,
	})

	in := []uuid.UUID{datasets.ID, floor.ID}
	out, err := svc.autoInjectFloorScope(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != len(in) {
		t.Errorf("expected no injection (floor already present), got len %d → %d", len(in), len(out))
	}
}

func TestAutoInjectFloorScope_NoOpWhenOnlyOrgScopedPermissions(t *testing.T) {
	billing := mkPerm("billing:read", authDomain.ScopeLevelOrganization)
	settings := mkPerm("settings:read", authDomain.ScopeLevelOrganization)

	svc := newRoleServiceForFloorScopeTest(map[string]*authDomain.Permission{
		billing.Name:  billing,
		settings.Name: settings,
	})

	in := []uuid.UUID{billing.ID, settings.ID}
	out, err := svc.autoInjectFloorScope(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != len(in) {
		t.Errorf("expected no injection (org-only set), got len %d → %d", len(in), len(out))
	}
}

func TestAutoInjectFloorScope_EmptyInputIsEmptyOutput(t *testing.T) {
	svc := newRoleServiceForFloorScopeTest(map[string]*authDomain.Permission{})
	out, err := svc.autoInjectFloorScope(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty output for empty input, got %v", out)
	}
}

func TestAutoInjectFloorScope_NoOpForOrgTierOnly(t *testing.T) {
	// A custom role granting ONLY org-tier verbs (e.g., a role intended
	// for org-member management via `org_members:*`) MUST NOT trigger
	// floor-scope auto-injection of `projects:read`. Floor injection
	// fires only when a project-tier permission is granted; org-tier
	// permissions resolve against the org role independently of project
	// context. Pinned by tagging org-rooted verbs `scope: organization`
	// in seeds/permissions.yaml.
	orgMembersRead := mkPerm("org_members:read", authDomain.ScopeLevelOrganization)
	orgMembersUpdate := mkPerm("org_members:update", authDomain.ScopeLevelOrganization)
	orgMembersRemove := mkPerm("org_members:remove", authDomain.ScopeLevelOrganization)
	orgMembersInvite := mkPerm("org_members:invite", authDomain.ScopeLevelOrganization)

	svc := newRoleServiceForFloorScopeTest(map[string]*authDomain.Permission{
		orgMembersRead.Name:   orgMembersRead,
		orgMembersUpdate.Name: orgMembersUpdate,
		orgMembersRemove.Name: orgMembersRemove,
		orgMembersInvite.Name: orgMembersInvite,
	})

	in := []uuid.UUID{orgMembersRead.ID, orgMembersUpdate.ID, orgMembersRemove.ID, orgMembersInvite.ID}
	out, err := svc.autoInjectFloorScope(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != len(in) {
		t.Errorf("expected no injection on org-tier-only role, got len %d → %d (over-injection regression)", len(in), len(out))
	}
}

func TestAutoInjectFloorScope_RejectsUnknownPermissionID(t *testing.T) {
	svc := newRoleServiceForFloorScopeTest(map[string]*authDomain.Permission{})
	_, err := svc.autoInjectFloorScope(context.Background(), []uuid.UUID{uuid.New()})
	if err == nil {
		t.Fatalf("expected InvalidParam for unknown id, got nil")
	}
	if !appErrors.IsReason(err, appErrors.ReasonInvalidInput) {
		t.Errorf("expected ReasonInvalidInput error, got %v", err)
	}
}

// containsPermissionByName reports whether the result list resolves
// (via the fake repo) to at least one Permission with the given name.
// The invariant tests below assert post-condition shape — "the role
// ends up with the floor permission" — without coupling to the
// internal trigger map / loop ordering of autoInjectFloorScope.
func containsPermissionByName(t *testing.T, repo *fakePermissionRepo, ids []uuid.UUID, name string) bool {
	t.Helper()
	for _, id := range ids {
		p, ok := repo.byID[id]
		if !ok {
			continue
		}
		if p.Name == name {
			return true
		}
	}
	return false
}

// TestAutoInjectFloorScope_InjectsForOrgProjectsList locks the
// invariant: a role granting `org_projects:list` (without the floor)
// must come out of autoInjectFloorScope carrying `projects:read`.
// Industry convention: list implies read. Closes the gap where
// `CanDiscoverProject` short-circuits on `:list` while route gates
// require the per-project floor (route layer / discovery alignment).
func TestAutoInjectFloorScope_InjectsForOrgProjectsList(t *testing.T) {
	orgList := mkPerm("org_projects:list", authDomain.ScopeLevelOrganization)
	floor := mkPerm("projects:read", authDomain.ScopeLevelProject)
	repo := &fakePermissionRepo{
		byID: map[uuid.UUID]*authDomain.Permission{
			orgList.ID: orgList,
			floor.ID:   floor,
		},
		byName: map[string]*authDomain.Permission{
			orgList.Name: orgList,
			floor.Name:   floor,
		},
	}
	svc := &RoleService{permissionRepo: repo}

	out, err := svc.autoInjectFloorScope(context.Background(), []uuid.UUID{orgList.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsPermissionByName(t, repo, out, "projects:read") {
		t.Errorf("invariant violated: role with `org_projects:list` must carry `projects:read` after auto-injection; got %v", out)
	}
}

// TestAutoInjectFloorScope_InjectsForOrgProjectsAdmin locks the same
// invariant for the org-tier admin permission. Admin at org level
// trivially implies read at project level.
func TestAutoInjectFloorScope_InjectsForOrgProjectsAdmin(t *testing.T) {
	orgAdmin := mkPerm("org_projects:admin", authDomain.ScopeLevelOrganization)
	floor := mkPerm("projects:read", authDomain.ScopeLevelProject)
	repo := &fakePermissionRepo{
		byID: map[uuid.UUID]*authDomain.Permission{
			orgAdmin.ID: orgAdmin,
			floor.ID:    floor,
		},
		byName: map[string]*authDomain.Permission{
			orgAdmin.Name: orgAdmin,
			floor.Name:    floor,
		},
	}
	svc := &RoleService{permissionRepo: repo}

	out, err := svc.autoInjectFloorScope(context.Background(), []uuid.UUID{orgAdmin.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsPermissionByName(t, repo, out, "projects:read") {
		t.Errorf("invariant violated: role with `org_projects:admin` must carry `projects:read` after auto-injection; got %v", out)
	}
}

// TestAutoInjectFloorScope_NoOpForOrgProjectsCreateOnly pins the
// deliberate exclusion of `org_projects:create` from the trigger set.
// A create-only role gets per-project `projects:read` via the
// creator-override row written at CreateProject time (round 22). Auto-
// injecting at role-creation time would broaden the user's read access
// to EVERY project in the org via inheritance — over-broad for
// "create-only" semantics. The override row correctly limits
// visibility to projects the user actually created.
func TestAutoInjectFloorScope_NoOpForOrgProjectsCreateOnly(t *testing.T) {
	orgCreate := mkPerm("org_projects:create", authDomain.ScopeLevelOrganization)
	repo := &fakePermissionRepo{
		byID:   map[uuid.UUID]*authDomain.Permission{orgCreate.ID: orgCreate},
		byName: map[string]*authDomain.Permission{orgCreate.Name: orgCreate},
	}
	svc := &RoleService{permissionRepo: repo}

	out, err := svc.autoInjectFloorScope(context.Background(), []uuid.UUID{orgCreate.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if containsPermissionByName(t, repo, out, "projects:read") {
		t.Errorf("invariant violated: `org_projects:create` MUST NOT auto-inject `projects:read` (would broaden access beyond creator-override semantics); got %v", out)
	}
}
