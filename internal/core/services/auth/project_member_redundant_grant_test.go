// Tests for ProjectMemberService.rejectIfRedundantGrant — the
// API-boundary guard that prevents project_members assignments which
// would have no effect under the round-24 additive resolver (a
// proposed role whose project-tier scopes are a subset of the user's
// org-role projection contributes zero new permissions).
package auth

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
)

// stubOrgMemberRepo answers GetByUserAndOrganization with a fixed
// OrganizationMember row. Other methods panic so an unintended caller
// shows up in test output.
type stubOrgMemberRepo struct {
	member *authDomain.OrganizationMember
	err    error
}

func (s *stubOrgMemberRepo) GetByUserAndOrganization(context.Context, uuid.UUID, uuid.UUID) (*authDomain.OrganizationMember, error) {
	return s.member, s.err
}
func (s *stubOrgMemberRepo) Create(context.Context, *authDomain.OrganizationMember) error {
	panic("unimplemented")
}
func (s *stubOrgMemberRepo) Update(context.Context, *authDomain.OrganizationMember) error {
	panic("unimplemented")
}
func (s *stubOrgMemberRepo) Delete(context.Context, uuid.UUID, uuid.UUID) error {
	panic("unimplemented")
}
func (s *stubOrgMemberRepo) GetByUserID(context.Context, uuid.UUID) ([]*authDomain.OrganizationMember, error) {
	panic("unimplemented")
}
func (s *stubOrgMemberRepo) GetByOrganizationID(context.Context, uuid.UUID) ([]*authDomain.OrganizationMember, error) {
	panic("unimplemented")
}
func (s *stubOrgMemberRepo) GetByRole(context.Context, uuid.UUID) ([]*authDomain.OrganizationMember, error) {
	panic("unimplemented")
}
func (s *stubOrgMemberRepo) Exists(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	panic("unimplemented")
}
func (s *stubOrgMemberRepo) GetUserEffectivePermissions(context.Context, uuid.UUID) ([]string, error) {
	panic("unimplemented")
}
func (s *stubOrgMemberRepo) CheckUserPermissions(context.Context, uuid.UUID, []string) (map[string]bool, error) {
	panic("unimplemented")
}
func (s *stubOrgMemberRepo) GetUserPermissionsInOrganization(context.Context, uuid.UUID, uuid.UUID) ([]string, error) {
	panic("unimplemented")
}
func (s *stubOrgMemberRepo) GetActiveMembers(context.Context, uuid.UUID) ([]*authDomain.OrganizationMember, error) {
	panic("unimplemented")
}
func (s *stubOrgMemberRepo) UpdateMemberRole(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
	panic("unimplemented")
}
func (s *stubOrgMemberRepo) GetMemberCount(context.Context, uuid.UUID) (int, error) {
	panic("unimplemented")
}
func (s *stubOrgMemberRepo) GetMembersByRole(context.Context, uuid.UUID) (map[string]int, error) {
	panic("unimplemented")
}

// stubRoleRepoByID returns a different perm slice depending on which
// role ID is passed to GetRolePermissions. Used to model the
// org-role-vs-proposed-role distinction the redundant-grant check
// depends on.
type stubRoleRepoByID struct {
	permsByRoleID map[uuid.UUID][]*authDomain.Permission
}

func (s *stubRoleRepoByID) GetRolePermissions(_ context.Context, roleID uuid.UUID) ([]*authDomain.Permission, error) {
	if perms, ok := s.permsByRoleID[roleID]; ok {
		return perms, nil
	}
	return nil, errors.New("unexpected role id in stub")
}

func (s *stubRoleRepoByID) Create(context.Context, *authDomain.Role) error {
	panic("unimplemented")
}
func (s *stubRoleRepoByID) GetByID(context.Context, uuid.UUID) (*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *stubRoleRepoByID) GetByNameAndScope(context.Context, string, string) (*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *stubRoleRepoByID) Update(context.Context, *authDomain.Role) error {
	panic("unimplemented")
}
func (s *stubRoleRepoByID) Delete(context.Context, uuid.UUID) error {
	panic("unimplemented")
}
func (s *stubRoleRepoByID) GetByScopeType(context.Context, string) ([]*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *stubRoleRepoByID) ListRoles(context.Context) ([]*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *stubRoleRepoByID) GetSystemRoles(context.Context) ([]*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *stubRoleRepoByID) GetCustomRolesByScopeID(context.Context, string, uuid.UUID) ([]*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *stubRoleRepoByID) GetByNameScopeAndID(context.Context, string, string, *uuid.UUID) (*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *stubRoleRepoByID) GetCustomRolesByOrganization(context.Context, uuid.UUID) ([]*authDomain.Role, error) {
	panic("unimplemented")
}
func (s *stubRoleRepoByID) AssignRolePermissions(context.Context, uuid.UUID, []uuid.UUID, *uuid.UUID) error {
	panic("unimplemented")
}
func (s *stubRoleRepoByID) RevokeRolePermissions(context.Context, uuid.UUID, []uuid.UUID) error {
	panic("unimplemented")
}
func (s *stubRoleRepoByID) UpdateRolePermissions(context.Context, uuid.UUID, []uuid.UUID, *uuid.UUID) error {
	panic("unimplemented")
}
func (s *stubRoleRepoByID) GetRoleStatistics(context.Context) (*authDomain.RoleStatistics, error) {
	panic("unimplemented")
}
// projectPerm builds a project-tier Permission with the given name.
func projectPerm(name string) *authDomain.Permission {
	return &authDomain.Permission{Name: name, ScopeLevel: authDomain.ScopeLevelProject}
}

// orgPerm builds an org-tier Permission with the given name. Used to
// confirm that org-tier permissions in either role do NOT participate
// in the subset check (the projection set considers only project-tier
// scopes).
func orgPerm(name string) *authDomain.Permission {
	return &authDomain.Permission{Name: name, ScopeLevel: authDomain.ScopeLevelOrganization}
}

func TestRejectIfRedundantGrant(t *testing.T) {
	orgRoleID := uuid.New()
	proposedRoleID := uuid.New()
	userID := uuid.New()
	orgID := uuid.New()

	cases := []struct {
		name         string
		orgRolePerms []*authDomain.Permission
		propRolePerms []*authDomain.Permission
		wantConflict bool
	}{
		{
			name:         "exact_equal_project_tier_scopes",
			orgRolePerms: []*authDomain.Permission{projectPerm("projects:read"), projectPerm("traces:read")},
			propRolePerms: []*authDomain.Permission{projectPerm("projects:read"), projectPerm("traces:read")},
			wantConflict: true, // proposed ⊆ org → no-op
		},
		{
			name:         "demotion_attempt_strict_subset",
			orgRolePerms: []*authDomain.Permission{projectPerm("projects:read"), projectPerm("projects:write"), projectPerm("projects:delete")},
			propRolePerms: []*authDomain.Permission{projectPerm("projects:read")},
			wantConflict: true, // viewer-shape ⊊ admin-shape → would silently no-op
		},
		{
			name:         "genuine_elevation_disjoint",
			orgRolePerms: []*authDomain.Permission{projectPerm("projects:read")},
			propRolePerms: []*authDomain.Permission{projectPerm("projects:read"), projectPerm("projects:write"), projectPerm("traces:delete")},
			wantConflict: false,
		},
		{
			name:         "mixed_overlap_with_new_scopes",
			orgRolePerms: []*authDomain.Permission{projectPerm("projects:read")},
			propRolePerms: []*authDomain.Permission{projectPerm("projects:read"), projectPerm("traces:write")},
			wantConflict: false, // overlap on read but adds traces:write
		},
		{
			name:         "empty_proposed_is_vacuous_subset",
			orgRolePerms: []*authDomain.Permission{projectPerm("projects:read")},
			propRolePerms: []*authDomain.Permission{},
			wantConflict: true,
		},
		{
			name:         "org_role_has_no_project_perms_proposed_does",
			orgRolePerms: []*authDomain.Permission{orgPerm("billing:read")},
			propRolePerms: []*authDomain.Permission{projectPerm("projects:read")},
			wantConflict: false, // empty org-projection; any project-tier scope is genuine elevation
		},
		{
			name:         "both_empty_at_project_tier",
			orgRolePerms: []*authDomain.Permission{orgPerm("billing:read")},
			propRolePerms: []*authDomain.Permission{orgPerm("members:read")},
			wantConflict: true, // proposed adds nothing project-tier (vacuous)
		},
		{
			name:         "org_tier_perms_do_not_count_for_subset",
			orgRolePerms: []*authDomain.Permission{projectPerm("projects:read"), orgPerm("billing:read")},
			propRolePerms: []*authDomain.Permission{projectPerm("projects:read"), orgPerm("members:read")},
			wantConflict: true, // project-tier sets are equal; the org-tier difference is irrelevant
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewProjectMemberService(
				nil,
				&stubOrgMemberRepo{member: &authDomain.OrganizationMember{
					UserID:         userID,
					OrganizationID: orgID,
					RoleID:         orgRoleID,
				}},
				&stubRoleRepoByID{permsByRoleID: map[uuid.UUID][]*authDomain.Permission{
					orgRoleID:      tc.orgRolePerms,
					proposedRoleID: tc.propRolePerms,
				}},
			)

			err := svc.rejectIfRedundantGrant(context.Background(), userID, orgID, proposedRoleID)
			if tc.wantConflict {
				if err == nil {
					t.Fatalf("expected Conflict, got nil")
				}
				if !strings.Contains(err.Error(), "equivalent or broader") {
					t.Errorf("expected canonical 'equivalent or broader' message, got %q", err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("expected nil, got %v", err)
				}
			}
		})
	}
}
