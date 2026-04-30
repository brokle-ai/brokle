// Tests for AddCreatorGrant (system-initiated creator-attribution
// path) + AddMember preservation of the redundant-grant guard
// (operator path). These validate the round-26 split that resolves
// the regression where round-25's redundant-grant guard blocked
// CreateProject for owner/admin org roles.
package auth

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
)

// configurableProjectMemberRepo satisfies authDomain.ProjectMemberRepository.
// Test bodies set the response fields per case; un-implemented methods
// panic so an unintended caller surfaces in test output.
type configurableProjectMemberRepo struct {
	isMember     bool
	isMemberErr  error
	createRows   int64
	createErr    error
	createCalled int
}

func (s *configurableProjectMemberRepo) IsMember(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return s.isMember, s.isMemberErr
}
func (s *configurableProjectMemberRepo) Create(context.Context, *authDomain.ProjectMember) (int64, error) {
	s.createCalled++
	return s.createRows, s.createErr
}
func (s *configurableProjectMemberRepo) ListUserEffectivePermissionsInScope(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]string, error) {
	panic("unimplemented")
}
func (s *configurableProjectMemberRepo) GetByUserAndProject(context.Context, uuid.UUID, uuid.UUID) (*authDomain.ProjectMember, error) {
	panic("unimplemented")
}
func (s *configurableProjectMemberRepo) UpdateRole(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
	panic("unimplemented")
}
func (s *configurableProjectMemberRepo) Delete(context.Context, uuid.UUID, uuid.UUID) error {
	panic("unimplemented")
}
func (s *configurableProjectMemberRepo) DeleteAllInOrgForUser(context.Context, uuid.UUID, uuid.UUID) error {
	panic("unimplemented")
}
func (s *configurableProjectMemberRepo) ListByProject(context.Context, uuid.UUID) ([]*authDomain.ProjectMember, error) {
	panic("unimplemented")
}
func (s *configurableProjectMemberRepo) ListByUser(context.Context, uuid.UUID) ([]*authDomain.ProjectMember, error) {
	panic("unimplemented")
}
func (s *configurableProjectMemberRepo) GetMemberCount(context.Context, uuid.UUID) (int, error) {
	panic("unimplemented")
}

// configurableOrgMemberRepo satisfies authDomain.OrganizationMemberRepository.
type configurableOrgMemberRepo struct {
	exists       bool
	existsErr    error
	member       *authDomain.OrganizationMember
	memberErr    error
}

func (s *configurableOrgMemberRepo) Exists(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return s.exists, s.existsErr
}
func (s *configurableOrgMemberRepo) GetByUserAndOrganization(context.Context, uuid.UUID, uuid.UUID) (*authDomain.OrganizationMember, error) {
	return s.member, s.memberErr
}
func (s *configurableOrgMemberRepo) Create(context.Context, *authDomain.OrganizationMember) error {
	panic("unimplemented")
}
func (s *configurableOrgMemberRepo) Update(context.Context, *authDomain.OrganizationMember) error {
	panic("unimplemented")
}
func (s *configurableOrgMemberRepo) Delete(context.Context, uuid.UUID, uuid.UUID) error {
	panic("unimplemented")
}
func (s *configurableOrgMemberRepo) GetByUserID(context.Context, uuid.UUID) ([]*authDomain.OrganizationMember, error) {
	panic("unimplemented")
}
func (s *configurableOrgMemberRepo) GetByOrganizationID(context.Context, uuid.UUID) ([]*authDomain.OrganizationMember, error) {
	panic("unimplemented")
}
func (s *configurableOrgMemberRepo) GetByRole(context.Context, uuid.UUID) ([]*authDomain.OrganizationMember, error) {
	panic("unimplemented")
}
func (s *configurableOrgMemberRepo) GetUserEffectivePermissions(context.Context, uuid.UUID) ([]string, error) {
	panic("unimplemented")
}
func (s *configurableOrgMemberRepo) CheckUserPermissions(context.Context, uuid.UUID, []string) (map[string]bool, error) {
	panic("unimplemented")
}
func (s *configurableOrgMemberRepo) GetUserPermissionsInOrganization(context.Context, uuid.UUID, uuid.UUID) ([]string, error) {
	panic("unimplemented")
}
func (s *configurableOrgMemberRepo) GetActiveMembers(context.Context, uuid.UUID) ([]*authDomain.OrganizationMember, error) {
	panic("unimplemented")
}
func (s *configurableOrgMemberRepo) UpdateMemberRole(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
	panic("unimplemented")
}
func (s *configurableOrgMemberRepo) BulkCreate(context.Context, []*authDomain.OrganizationMember) error {
	panic("unimplemented")
}
func (s *configurableOrgMemberRepo) BulkUpdateRoles(context.Context, []authDomain.MemberRoleUpdate) error {
	panic("unimplemented")
}
func (s *configurableOrgMemberRepo) GetMemberCount(context.Context, uuid.UUID) (int, error) {
	panic("unimplemented")
}
func (s *configurableOrgMemberRepo) GetMembersByRole(context.Context, uuid.UUID) (map[string]int, error) {
	panic("unimplemented")
}

// configurableRoleRepo answers GetByID and GetRolePermissions per
// role-id keys; the rest panic.
type configurableRoleRepo struct {
	rolesByID       map[uuid.UUID]*authDomain.Role
	permsByRoleID   map[uuid.UUID][]*authDomain.Permission
	permsByRoleErr  map[uuid.UUID]error
}

func (s *configurableRoleRepo) GetByID(_ context.Context, id uuid.UUID) (*authDomain.Role, error) {
	if r, ok := s.rolesByID[id]; ok {
		return r, nil
	}
	return nil, errors.New("unexpected role id in stub")
}
func (s *configurableRoleRepo) GetRolePermissions(_ context.Context, id uuid.UUID) ([]*authDomain.Permission, error) {
	if err, ok := s.permsByRoleErr[id]; ok {
		return nil, err
	}
	if perms, ok := s.permsByRoleID[id]; ok {
		return perms, nil
	}
	return nil, errors.New("unexpected role id in stub")
}
func (s *configurableRoleRepo) Create(context.Context, *authDomain.Role) error {
	panic("unimplemented")
}
func (s *configurableRoleRepo) GetByNameAndScope(context.Context, string, string) (*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *configurableRoleRepo) Update(context.Context, *authDomain.Role) error {
	panic("unimplemented")
}
func (s *configurableRoleRepo) Delete(context.Context, uuid.UUID) error {
	panic("unimplemented")
}
func (s *configurableRoleRepo) GetByScopeType(context.Context, string) ([]*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *configurableRoleRepo) ListRoles(context.Context) ([]*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *configurableRoleRepo) GetSystemRoles(context.Context) ([]*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *configurableRoleRepo) GetCustomRolesByScopeID(context.Context, string, uuid.UUID) ([]*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *configurableRoleRepo) GetByNameScopeAndID(context.Context, string, string, *uuid.UUID) (*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *configurableRoleRepo) GetCustomRolesByOrganization(context.Context, uuid.UUID) ([]*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *configurableRoleRepo) AssignRolePermissions(context.Context, uuid.UUID, []uuid.UUID, *uuid.UUID) error {
	panic("unimplemented")
}
func (s *configurableRoleRepo) RevokeRolePermissions(context.Context, uuid.UUID, []uuid.UUID) error {
	panic("unimplemented")
}
func (s *configurableRoleRepo) UpdateRolePermissions(context.Context, uuid.UUID, []uuid.UUID, *uuid.UUID) error {
	panic("unimplemented")
}
func (s *configurableRoleRepo) GetRoleStatistics(context.Context) (*authDomain.RoleStatistics, error) {
	panic("unimplemented")
}
func (s *configurableRoleRepo) BulkCreate(context.Context, []*authDomain.Role) error {
	panic("unimplemented")
}

// builtinAdminRole returns a Role shaped like the seeded `admin`
// template (scope_type=organization, scope_id=NULL). This is exactly
// the role ProjectService.CreateProject resolves and assigns.
func builtinAdminRole(id uuid.UUID) *authDomain.Role {
	return &authDomain.Role{
		ID:        id,
		Name:      "admin",
		ScopeType: authDomain.ScopeOrganization,
		ScopeID:   nil,
	}
}

// adminTemplateProjectPerms is the project-tier scope set the seeded
// admin template carries — the full project lifecycle
// `{projects:read, projects:write, projects:admin, projects:delete}`.
// admin and owner share the same project-tier projection; owner stays
// distinct at the ORG tier via `organizations:delete` (not by
// withholding project deletion). See seeds/roles.yaml admin block
// for the rationale.
func adminTemplateProjectPerms() []*authDomain.Permission {
	return []*authDomain.Permission{
		projectPerm("projects:read"),
		projectPerm("projects:write"),
		projectPerm("projects:admin"),
		projectPerm("projects:delete"),
	}
}

// TestAddCreatorGrant_BypassesRedundantGuard exercises the system
// path that ProjectService.CreateProject takes. Owner/admin/developer
// /custom-create-only all succeed because the redundant-grant guard
// is intentionally NOT applied here — the row is creator-attribution,
// not permission elevation.
func TestAddCreatorGrant_BypassesRedundantGuard(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()
	orgRoleID := uuid.New()
	adminRoleID := uuid.New()

	cases := []struct {
		name         string
		orgRolePerms []*authDomain.Permission
	}{
		{
			// owner and admin now share the same project-tier projection
			// (admin gained projects:delete in round 27); the case still
			// exercises a distinct org role from `admin_org_role_equals_admin_template`
			// to confirm AddCreatorGrant succeeds regardless of org-role
			// identity, even when the org-role's project-tier set equals
			// the creator-grant template's.
			name: "owner_org_role_equals_admin_template_set",
			orgRolePerms: []*authDomain.Permission{
				projectPerm("projects:read"),
				projectPerm("projects:write"),
				projectPerm("projects:delete"),
				projectPerm("projects:admin"),
			},
		},
		{
			name:         "admin_org_role_equals_admin_template",
			orgRolePerms: adminTemplateProjectPerms(),
		},
		{
			name: "developer_org_role_admin_template_elevates",
			orgRolePerms: []*authDomain.Permission{
				projectPerm("projects:read"),
				projectPerm("projects:write"),
			},
		},
		{
			name: "custom_create_only_role_no_project_tier_perms",
			orgRolePerms: []*authDomain.Permission{
				orgPerm("org_projects:create"),
				orgPerm("org_members:read"),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			projMemberRepo := &configurableProjectMemberRepo{
				isMember:   false, // no existing project membership
				createRows: 1,     // UPSERT inserts/updates exactly one row
			}
			svc := NewProjectMemberService(
				projMemberRepo,
				&configurableOrgMemberRepo{
					exists: true,
					member: &authDomain.OrganizationMember{
						UserID:         userID,
						OrganizationID: orgID,
						RoleID:         orgRoleID,
					},
				},
				&configurableRoleRepo{
					rolesByID: map[uuid.UUID]*authDomain.Role{
						adminRoleID: builtinAdminRole(adminRoleID),
					},
					permsByRoleID: map[uuid.UUID][]*authDomain.Permission{
						orgRoleID:   tc.orgRolePerms,
						adminRoleID: adminTemplateProjectPerms(),
					},
				},
			)

			member, err := svc.AddCreatorGrant(context.Background(), userID, projectID, orgID, adminRoleID)
			if err != nil {
				t.Fatalf("AddCreatorGrant must succeed (creator-attribution row, guard bypassed); got %v", err)
			}
			if member == nil {
				t.Fatal("expected non-nil project_member result")
			}
			if projMemberRepo.createCalled != 1 {
				t.Errorf("expected exactly one Create call (UPSERT), got %d", projMemberRepo.createCalled)
			}
		})
	}
}

// TestAddCreatorGrant_StillRejectsCrossTenantRole asserts that
// validateProjectAssignableRole runs even on the system path —
// AddCreatorGrant is not a free pass, just a redundant-grant-guard
// bypass.
func TestAddCreatorGrant_StillRejectsCrossTenantRole(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()
	orgRoleID := uuid.New()
	foreignOrg := uuid.New()
	foreignRoleID := uuid.New()

	svc := NewProjectMemberService(
		&configurableProjectMemberRepo{},
		&configurableOrgMemberRepo{
			exists: true,
			member: &authDomain.OrganizationMember{RoleID: orgRoleID},
		},
		&configurableRoleRepo{
			rolesByID: map[uuid.UUID]*authDomain.Role{
				foreignRoleID: {
					ID:        foreignRoleID,
					ScopeType: authDomain.ScopeOrganization,
					ScopeID:   &foreignOrg, // belongs to a different org
				},
			},
		},
	)

	_, err := svc.AddCreatorGrant(context.Background(), userID, projectID, orgID, foreignRoleID)
	if err == nil {
		t.Fatal("expected cross-tenant rejection, got nil")
	}
	if !strings.Contains(err.Error(), "does not belong to this organization") {
		t.Errorf("expected cross-tenant scope error, got %q", err.Error())
	}
}

// TestAddMember_KeepsRedundantGuard confirms the operator path is
// unchanged: assigning an org admin a role whose project-tier scopes
// ⊆ their org-role projection still returns 409. This is the round-25
// invariant we deliberately preserved on the operator surface.
func TestAddMember_KeepsRedundantGuard(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()
	orgRoleID := uuid.New()
	adminRoleID := uuid.New()

	projMemberRepo := &configurableProjectMemberRepo{}
	svc := NewProjectMemberService(
		projMemberRepo,
		&configurableOrgMemberRepo{
			exists: true,
			member: &authDomain.OrganizationMember{
				UserID:         userID,
				OrganizationID: orgID,
				RoleID:         orgRoleID,
			},
		},
		&configurableRoleRepo{
			rolesByID: map[uuid.UUID]*authDomain.Role{
				adminRoleID: builtinAdminRole(adminRoleID),
			},
			permsByRoleID: map[uuid.UUID][]*authDomain.Permission{
				// org admin already has admin's project-tier scope set
				orgRoleID:   adminTemplateProjectPerms(),
				adminRoleID: adminTemplateProjectPerms(),
			},
		},
	)

	_, err := svc.AddMember(context.Background(), userID, projectID, orgID, adminRoleID)
	if err == nil {
		t.Fatal("expected 409 redundant-grant rejection, got nil")
	}
	if !strings.Contains(err.Error(), "equivalent or broader") {
		t.Errorf("expected canonical redundant-grant message, got %q", err.Error())
	}
	// Critical invariant: Create must NOT be called when the guard fires.
	if projMemberRepo.createCalled != 0 {
		t.Errorf("expected Create to NOT run when guard rejects, got %d calls", projMemberRepo.createCalled)
	}
}
