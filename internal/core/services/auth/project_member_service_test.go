package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
)

// stubProjectMemberRepo is a minimal fake. Only the methods exercised by
// these tests return real values; everything else panics so a new caller
// shows up in test output instead of silently no-op'ing.
type stubProjectMemberRepo struct {
	effectivePerms []string
	effectiveErr   error
}

func (s *stubProjectMemberRepo) ListUserEffectivePermissionsInScope(ctx context.Context, userID, orgID, projectID uuid.UUID) ([]string, error) {
	return s.effectivePerms, s.effectiveErr
}

func (s *stubProjectMemberRepo) Create(context.Context, *authDomain.ProjectMember) (int64, error) {
	panic("unimplemented")
}
func (s *stubProjectMemberRepo) GetByUserAndProject(context.Context, uuid.UUID, uuid.UUID) (*authDomain.ProjectMember, error) {
	panic("unimplemented")
}
func (s *stubProjectMemberRepo) UpdateRole(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
	panic("unimplemented")
}
func (s *stubProjectMemberRepo) Delete(context.Context, uuid.UUID, uuid.UUID) error {
	panic("unimplemented")
}
func (s *stubProjectMemberRepo) DeleteAllInOrgForUser(context.Context, uuid.UUID, uuid.UUID) error {
	panic("unimplemented")
}
func (s *stubProjectMemberRepo) ListByProject(context.Context, uuid.UUID) ([]*authDomain.ProjectMember, error) {
	panic("unimplemented")
}
func (s *stubProjectMemberRepo) ListByUser(context.Context, uuid.UUID) ([]*authDomain.ProjectMember, error) {
	panic("unimplemented")
}
func (s *stubProjectMemberRepo) IsMember(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	panic("unimplemented")
}
func (s *stubProjectMemberRepo) GetMemberCount(context.Context, uuid.UUID) (int, error) {
	panic("unimplemented")
}

// TestCheckUserPermissionsInScope_MapsResults exercises the result-shape
// contract: a permission appears in the result map iff the user holds it,
// even when the input list contains permissions the user does NOT hold
// (which must come back as false rather than missing keys).
func TestCheckUserPermissionsInScope_MapsResults(t *testing.T) {
	svc := NewProjectMemberService(
		&stubProjectMemberRepo{
			effectivePerms: []string{"traces:read", "scores:write"},
		},
		nil, // orgMemberRepo unused on this path
		nil, // roleRepo unused on this path
	)

	got, err := svc.CheckUserPermissionsInScope(
		context.Background(),
		uuid.New(), uuid.New(), uuid.New(),
		[]string{"traces:read", "scores:write", "projects:delete"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[string]bool{
		"traces:read":     true,
		"scores:write":    true,
		"projects:delete": false,
	}
	for perm, expected := range want {
		if got[perm] != expected {
			t.Errorf("permission %q: got %v, want %v", perm, got[perm], expected)
		}
	}
	if len(got) != len(want) {
		t.Errorf("expected %d entries, got %d (%v)", len(want), len(got), got)
	}
}

// TestCheckUserPermissionsInScope_EmptyRequestShortCircuits guards the
// no-op path: callers passing an empty permission slice should not
// trigger a DB lookup. We assert this by configuring the stub repo to
// return an error and checking it never gets called.
func TestCheckUserPermissionsInScope_EmptyRequestShortCircuits(t *testing.T) {
	svc := NewProjectMemberService(
		&stubProjectMemberRepo{effectiveErr: errors.New("repo must not be called")},
		nil, nil,
	)

	got, err := svc.CheckUserPermissionsInScope(
		context.Background(),
		uuid.New(), uuid.New(), uuid.New(),
		[]string{},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

// TestCheckUserPermissionsInScope_RepoErrorPropagates ensures repo
// failures surface as 500 AppError (not silently dropped to false).
func TestCheckUserPermissionsInScope_RepoErrorPropagates(t *testing.T) {
	svc := NewProjectMemberService(
		&stubProjectMemberRepo{effectiveErr: errors.New("connection refused")},
		nil, nil,
	)

	_, err := svc.CheckUserPermissionsInScope(
		context.Background(),
		uuid.New(), uuid.New(), uuid.New(),
		[]string{"traces:read"},
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// stubRoleRepoForFloorCheck is a minimal RoleRepository fake for the
// validateProjectAssignableRole floor-check tests. Only GetRolePermissions
// returns real values; everything else panics so an unintended caller
// shows up in test output.
type stubRoleRepoForFloorCheck struct {
	perms    []*authDomain.Permission
	permsErr error
}

func (s *stubRoleRepoForFloorCheck) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*authDomain.Permission, error) {
	return s.perms, s.permsErr
}

func (s *stubRoleRepoForFloorCheck) Create(context.Context, *authDomain.Role) error {
	panic("unimplemented")
}
func (s *stubRoleRepoForFloorCheck) GetByID(context.Context, uuid.UUID) (*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *stubRoleRepoForFloorCheck) GetByNameAndScope(context.Context, string, string) (*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *stubRoleRepoForFloorCheck) Update(context.Context, *authDomain.Role) error {
	panic("unimplemented")
}
func (s *stubRoleRepoForFloorCheck) Delete(context.Context, uuid.UUID) error {
	panic("unimplemented")
}
func (s *stubRoleRepoForFloorCheck) GetByScopeType(context.Context, string) ([]*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *stubRoleRepoForFloorCheck) ListRoles(context.Context) ([]*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *stubRoleRepoForFloorCheck) GetSystemRoles(context.Context) ([]*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *stubRoleRepoForFloorCheck) GetCustomRolesByScopeID(context.Context, string, uuid.UUID) ([]*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *stubRoleRepoForFloorCheck) GetByNameScopeAndID(context.Context, string, string, *uuid.UUID) (*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *stubRoleRepoForFloorCheck) GetCustomRolesByOrganization(context.Context, uuid.UUID) ([]*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *stubRoleRepoForFloorCheck) AssignRolePermissions(context.Context, uuid.UUID, []uuid.UUID, *uuid.UUID) error {
	panic("unimplemented")
}
func (s *stubRoleRepoForFloorCheck) RevokeRolePermissions(context.Context, uuid.UUID, []uuid.UUID) error {
	panic("unimplemented")
}
func (s *stubRoleRepoForFloorCheck) UpdateRolePermissions(context.Context, uuid.UUID, []uuid.UUID, *uuid.UUID) error {
	panic("unimplemented")
}
func (s *stubRoleRepoForFloorCheck) GetRoleStatistics(context.Context) (*authDomain.RoleStatistics, error) {
	panic("unimplemented")
}
func (s *stubRoleRepoForFloorCheck) BulkCreate(context.Context, []*authDomain.Role) error {
	panic("unimplemented")
}

// orgScopedRole returns a built-in-shape role (ScopeID == nil so the
// existing scope check accepts it for any org).
func orgScopedRole() *authDomain.Role {
	return &authDomain.Role{
		ID:        uuid.New(),
		ScopeType: authDomain.ScopeOrganization,
		ScopeID:   nil,
	}
}

// TestValidateProjectAssignableRole_AcceptsBuiltinTemplate — happy path:
// a built-in org-scoped role template (ScopeID == nil) is accepted for
// any project. Cross-tenant safety relies on scope_id matching the
// project's org for org-custom roles, satisfied trivially for templates.
func TestValidateProjectAssignableRole_AcceptsBuiltinTemplate(t *testing.T) {
	svc := NewProjectMemberService(nil, nil, &stubRoleRepoForFloorCheck{})

	err := svc.validateProjectAssignableRole(context.Background(), orgScopedRole(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("expected nil error for built-in template, got %v", err)
	}
}

// TestValidateProjectAssignableRole_AcceptsRoleWithoutFloor — under the
// round-24 additive resolver the lockout vector does not exist
// (org-projection always unions in regardless of override content), so
// a role lacking `projects:read` is no longer rejected. ADR-0001 §
// "Decision" / "Considered alternatives" has the rationale.
func TestValidateProjectAssignableRole_AcceptsRoleWithoutFloor(t *testing.T) {
	svc := NewProjectMemberService(nil, nil, &stubRoleRepoForFloorCheck{
		perms: []*authDomain.Permission{
			{Name: "billing:read"},
			{Name: "billing:manage"},
		},
	})

	err := svc.validateProjectAssignableRole(context.Background(), orgScopedRole(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("expected nil error under additive resolver, got %v", err)
	}
}

// TestValidateProjectAssignableRole_AcceptsEmptyRole — degenerate
// shape: a role with zero permissions is also accepted under additive
// (the override grants nothing; the user keeps their org-projection
// unchanged). No-op assignments are legal.
func TestValidateProjectAssignableRole_AcceptsEmptyRole(t *testing.T) {
	svc := NewProjectMemberService(nil, nil, &stubRoleRepoForFloorCheck{
		perms: []*authDomain.Permission{},
	})

	err := svc.validateProjectAssignableRole(context.Background(), orgScopedRole(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("expected nil error for empty role under additive resolver, got %v", err)
	}
}

// TestValidateProjectAssignableRole_RejectsCrossTenantOrgRole — the
// scope_id check is still load-bearing under additive: an org-scoped
// role with scope_id set must match the target org. Cross-tenant role
// injection is closed at the assignment boundary regardless of
// resolver semantics.
func TestValidateProjectAssignableRole_RejectsCrossTenantOrgRole(t *testing.T) {
	otherOrg := uuid.New()
	role := &authDomain.Role{
		ID:        uuid.New(),
		ScopeType: authDomain.ScopeOrganization,
		ScopeID:   &otherOrg,
	}
	// permsErr would normally surface — but the validator no longer
	// calls GetRolePermissions, so the stub error must remain quiescent.
	svc := NewProjectMemberService(nil, nil, &stubRoleRepoForFloorCheck{
		permsErr: errors.New("must not be called"),
	})

	err := svc.validateProjectAssignableRole(context.Background(), role, uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected scope-mismatch error, got nil")
	}
	if !contains(err.Error(), "does not belong to this organization") {
		t.Errorf("expected scope-check error, got %q", err.Error())
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
