package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	"brokle/internal/infrastructure/db"
	"brokle/internal/infrastructure/db/gen"
	appErrors "brokle/pkg/errors"
)

// organizationMemberRepository is the pgx+sqlc implementation of
// authDomain.OrganizationMemberRepository. The GORM-era repository
// eagerly Preloaded Role on every read; handlers fetch roles via
// roleService instead, so the preload is dropped.
type organizationMemberRepository struct {
	tm *db.TxManager
}

// NewOrganizationMemberRepository returns the pgx-backed repository.
func NewOrganizationMemberRepository(tm *db.TxManager) authDomain.OrganizationMemberRepository {
	return &organizationMemberRepository{tm: tm}
}

// ----- CRUD ----------------------------------------------------------

func (r *organizationMemberRepository) Create(ctx context.Context, m *authDomain.OrganizationMember) error {
	now := time.Now()
	if m.JoinedAt.IsZero() {
		m.JoinedAt = now
	}
	// UPSERT-WHERE absorbs the (user_id, organization_id) PK conflict
	// structurally — see member.sql CreateMember query comment.
	rows, err := r.tm.Queries(ctx).CreateMember(ctx, gen.CreateMemberParams{
		UserID:         m.UserID,
		OrganizationID: m.OrganizationID,
		RoleID:         m.RoleID,
		JoinedAt:       m.JoinedAt,
		InvitedBy:      m.InvitedBy,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	if err != nil {
		return appErrors.Internal("create organization member", err,
			appErrors.WithOp("repo.organization_member.create"))
	}
	if rows == 0 {
		return appErrors.AlreadyExists("organization_member",
			appErrors.WithOp("repo.organization_member.create"))
	}
	return nil
}

func (r *organizationMemberRepository) GetByUserAndOrganization(ctx context.Context, userID, orgID uuid.UUID) (*authDomain.OrganizationMember, error) {
	row, err := r.tm.Queries(ctx).GetMemberByUserAndOrg(ctx, gen.GetMemberByUserAndOrgParams{
		UserID:         userID,
		OrganizationID: orgID,
	})
	if err != nil {
		if db.IsNoRows(err) {
			return nil, appErrors.NotFound("organization_member", appErrors.WithOp("repo.organization_member.get_by_user_and_org"))
		}
		return nil, fmt.Errorf("get member (user=%s org=%s): %w", userID, orgID, err)
	}
	return authMemberFromRow(&row), nil
}

func (r *organizationMemberRepository) Update(ctx context.Context, m *authDomain.OrganizationMember) error {
	if err := r.tm.Queries(ctx).UpdateMember(ctx, gen.UpdateMemberParams{
		UserID:         m.UserID,
		OrganizationID: m.OrganizationID,
		RoleID:         m.RoleID,
		InvitedBy:      m.InvitedBy,
	}); err != nil {
		return fmt.Errorf("update organization member (user=%s org=%s): %w", m.UserID, m.OrganizationID, err)
	}
	return nil
}

// Delete soft-deletes the organization_members row. The org-removal
// flow ALSO needs to cascade-clean project_members overrides for
// this user — that's the application service's responsibility (see
// OrganizationMemberService.RemoveMember). This repo method does ONE
// thing per Vernon's DDD: persistence per aggregate; atomicity owned
// by the use case.
func (r *organizationMemberRepository) Delete(ctx context.Context, userID, orgID uuid.UUID) error {
	if err := r.tm.Queries(ctx).SoftDeleteMemberByUserAndOrg(ctx, gen.SoftDeleteMemberByUserAndOrgParams{
		UserID:         userID,
		OrganizationID: orgID,
	}); err != nil {
		return fmt.Errorf("delete organization member (user=%s org=%s): %w", userID, orgID, err)
	}
	return nil
}

// ----- Membership queries -------------------------------------------

func (r *organizationMemberRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*authDomain.OrganizationMember, error) {
	rows, err := r.tm.Queries(ctx).ListMembersByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list memberships for user %s: %w", userID, err)
	}
	return authMembersFromRows(rows), nil
}

func (r *organizationMemberRepository) GetByOrganizationID(ctx context.Context, orgID uuid.UUID) ([]*authDomain.OrganizationMember, error) {
	rows, err := r.tm.Queries(ctx).ListMembersByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list members for org %s: %w", orgID, err)
	}
	return authMembersFromRows(rows), nil
}

func (r *organizationMemberRepository) GetByRole(ctx context.Context, roleID uuid.UUID) ([]*authDomain.OrganizationMember, error) {
	rows, err := r.tm.Queries(ctx).ListMembersByRole(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("list members for role %s: %w", roleID, err)
	}
	return authMembersFromRows(rows), nil
}

func (r *organizationMemberRepository) Exists(ctx context.Context, userID, orgID uuid.UUID) (bool, error) {
	ok, err := r.tm.Queries(ctx).IsMember(ctx, gen.IsMemberParams{
		UserID:         userID,
		OrganizationID: orgID,
	})
	if err != nil {
		return false, fmt.Errorf("check member existence (user=%s org=%s): %w", userID, orgID, err)
	}
	return ok, nil
}

// ----- Permission queries -------------------------------------------

func (r *organizationMemberRepository) GetUserEffectivePermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	perms, err := r.tm.Queries(ctx).ListUserEffectivePermissionsGlobal(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list effective permissions for user %s: %w", userID, err)
	}
	return perms, nil
}

func (r *organizationMemberRepository) CheckUserPermissions(ctx context.Context, userID uuid.UUID, permissions []string) (map[string]bool, error) {
	if len(permissions) == 0 {
		return map[string]bool{}, nil
	}
	granted, err := r.tm.Queries(ctx).ListUserEffectivePermissionsGlobal(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list effective permissions for user %s: %w", userID, err)
	}
	grantedSet := make(map[string]struct{}, len(granted))
	for _, p := range granted {
		grantedSet[p] = struct{}{}
	}
	result := make(map[string]bool, len(permissions))
	for _, p := range permissions {
		_, ok := grantedSet[p]
		result[p] = ok
	}
	return result, nil
}

func (r *organizationMemberRepository) GetUserPermissionsInOrganization(ctx context.Context, userID, orgID uuid.UUID) ([]string, error) {
	perms, err := r.tm.Queries(ctx).ListUserEffectivePermissionsInOrg(ctx, gen.ListUserEffectivePermissionsInOrgParams{
		UserID:         userID,
		OrganizationID: orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("list permissions in org (user=%s org=%s): %w", userID, orgID, err)
	}
	return perms, nil
}

// GetActiveMembers lists non-soft-deleted members. Suspension is no
// longer modeled (deleted 2026-04-30) — "active" here means
// "deleted_at IS NULL".
func (r *organizationMemberRepository) GetActiveMembers(ctx context.Context, orgID uuid.UUID) ([]*authDomain.OrganizationMember, error) {
	rows, err := r.tm.Queries(ctx).ListActiveMembersByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list active members for org %s: %w", orgID, err)
	}
	return authMembersFromRows(rows), nil
}

// ----- Role management ----------------------------------------------

func (r *organizationMemberRepository) UpdateMemberRole(ctx context.Context, userID, orgID, roleID uuid.UUID) error {
	if err := r.tm.Queries(ctx).UpdateMemberRole(ctx, gen.UpdateMemberRoleParams{
		UserID:         userID,
		OrganizationID: orgID,
		RoleID:         roleID,
	}); err != nil {
		return fmt.Errorf("update member role (user=%s org=%s): %w", userID, orgID, err)
	}
	return nil
}

// ----- Statistics ---------------------------------------------------

func (r *organizationMemberRepository) GetMemberCount(ctx context.Context, orgID uuid.UUID) (int, error) {
	n, err := r.tm.Queries(ctx).CountActiveMembersByOrganization(ctx, orgID)
	if err != nil {
		return 0, fmt.Errorf("count active members for org %s: %w", orgID, err)
	}
	return int(n), nil
}

func (r *organizationMemberRepository) GetMembersByRole(ctx context.Context, orgID uuid.UUID) (map[string]int, error) {
	rows, err := r.tm.Queries(ctx).CountActiveMembersByRoleName(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("count members by role for org %s: %w", orgID, err)
	}
	out := make(map[string]int, len(rows))
	for _, row := range rows {
		out[row.RoleName] = int(row.Count)
	}
	return out, nil
}

// ----- gen ↔ domain boundary ----------------------------------------

func authMemberFromRow(row *gen.OrganizationMember) *authDomain.OrganizationMember {
	return &authDomain.OrganizationMember{
		UserID:         row.UserID,
		OrganizationID: row.OrganizationID,
		RoleID:         row.RoleID,
		JoinedAt:       row.JoinedAt,
		InvitedBy:      row.InvitedBy,
	}
}

func authMembersFromRows(rows []gen.OrganizationMember) []*authDomain.OrganizationMember {
	out := make([]*authDomain.OrganizationMember, 0, len(rows))
	for i := range rows {
		out = append(out, authMemberFromRow(&rows[i]))
	}
	return out
}
