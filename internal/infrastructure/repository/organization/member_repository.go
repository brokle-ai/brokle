package organization

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	orgDomain "brokle/internal/core/domain/organization"
	"brokle/internal/infrastructure/db"
	"brokle/internal/infrastructure/db/gen"
	appErrors "brokle/pkg/errors"
)

// memberRepository is the pgx+sqlc implementation of orgDomain.MemberRepository.
// organization_members has a composite PK (user_id, organization_id); the
// single-ID Delete / GetByID methods on the interface are intentionally
// rejected because they cannot address a row unambiguously.
type memberRepository struct {
	tm *db.TxManager
}

// NewMemberRepository returns the pgx-backed repository.
func NewMemberRepository(tm *db.TxManager) orgDomain.MemberRepository {
	return &memberRepository{tm: tm}
}

func (r *memberRepository) Create(ctx context.Context, m *orgDomain.Member) error {
	now := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = now
	}
	if m.JoinedAt.IsZero() {
		m.JoinedAt = now
	}
	// UPSERT-WHERE absorbs the (user_id, organization_id) PK conflict
	// structurally: a fresh row is INSERTed; a soft-deleted row is
	// restored (deleted_at = NULL, fresh role/invited_by/joined_at/
	// updated_at); an active row is left untouched (rows = 0). See the
	// SQL query comment for the three reachable outcomes.
	rows, err := r.tm.Queries(ctx).CreateMember(ctx, gen.CreateMemberParams{
		UserID:         m.UserID,
		OrganizationID: m.OrganizationID,
		RoleID:         m.RoleID,
		JoinedAt:       m.JoinedAt,
		InvitedBy:      m.InvitedBy,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	})
	if err != nil {
		return appErrors.Internal("create member", err,
			appErrors.WithOp("repo.organization_member.create"))
	}
	if rows == 0 {
		// Active (non-soft-deleted) row already exists on the PK — the
		// caller raced a concurrent add. AlreadyExists matches the
		// pre-UPSERT contract; service-layer pre-checks (IsMember)
		// catch the common case fast.
		return appErrors.AlreadyExists("member",
			appErrors.WithOp("repo.organization_member.create"))
	}
	return nil
}

// GetByID is not supported — members are addressed by composite key.
func (r *memberRepository) GetByID(ctx context.Context, _ uuid.UUID) (*orgDomain.Member, error) {
	return nil, appErrors.NotImplemented("member", "get member by single ID not supported (composite key)", appErrors.WithOp("repo.organization_member.get_by_id"))
}

func (r *memberRepository) GetByUserAndOrganization(ctx context.Context, userID, orgID uuid.UUID) (*orgDomain.Member, error) {
	row, err := r.tm.Queries(ctx).GetMemberByUserAndOrg(ctx, gen.GetMemberByUserAndOrgParams{
		UserID:         userID,
		OrganizationID: orgID,
	})
	if err != nil {
		if db.IsNoRows(err) {
			return nil, appErrors.NotFound("member", appErrors.WithOp("repo.organization_member.get_by_user_and_org"))
		}
		return nil, fmt.Errorf("get member (user=%s org=%s): %w", userID, orgID, err)
	}
	return memberFromRow(&row), nil
}

func (r *memberRepository) Update(ctx context.Context, m *orgDomain.Member) error {
	if err := r.tm.Queries(ctx).UpdateMember(ctx, gen.UpdateMemberParams{
		UserID:         m.UserID,
		OrganizationID: m.OrganizationID,
		RoleID:         m.RoleID,
		InvitedBy:      m.InvitedBy,
	}); err != nil {
		return fmt.Errorf("update member (user=%s org=%s): %w", m.UserID, m.OrganizationID, err)
	}
	return nil
}

// Delete is not supported — members are addressed by composite key.
// Use DeleteByUserAndOrg instead.
func (r *memberRepository) Delete(ctx context.Context, _ uuid.UUID) error {
	return appErrors.NotImplemented("member", "delete member by single ID not supported (composite key)", appErrors.WithOp("repo.organization_member.delete"))
}

// DeleteByUserAndOrg soft-deletes the organization_members row.
// Cascade-cleanup of project_members overrides is the application
// service's responsibility (see invitation/member-removal flows).
// Single-statement persistence per Vernon's DDD.
func (r *memberRepository) DeleteByUserAndOrg(ctx context.Context, orgID, userID uuid.UUID) error {
	if err := r.tm.Queries(ctx).SoftDeleteMemberByUserAndOrg(ctx, gen.SoftDeleteMemberByUserAndOrgParams{
		UserID:         userID,
		OrganizationID: orgID,
	}); err != nil {
		return fmt.Errorf("soft-delete member (user=%s org=%s): %w", userID, orgID, err)
	}
	return nil
}

func (r *memberRepository) GetByOrganizationID(ctx context.Context, orgID uuid.UUID) ([]*orgDomain.Member, error) {
	rows, err := r.tm.Queries(ctx).ListMembersByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list members for org %s: %w", orgID, err)
	}
	return membersFromRows(rows), nil
}

func (r *memberRepository) UpdateMemberRole(ctx context.Context, orgID, userID, roleID uuid.UUID) error {
	if err := r.tm.Queries(ctx).UpdateMemberRole(ctx, gen.UpdateMemberRoleParams{
		UserID:         userID,
		OrganizationID: orgID,
		RoleID:         roleID,
	}); err != nil {
		return fmt.Errorf("update member role (user=%s org=%s): %w", userID, orgID, err)
	}
	return nil
}

func (r *memberRepository) GetMemberRole(ctx context.Context, userID, orgID uuid.UUID) (uuid.UUID, error) {
	id, err := r.tm.Queries(ctx).GetMemberRoleID(ctx, gen.GetMemberRoleIDParams{
		UserID:         userID,
		OrganizationID: orgID,
	})
	if err != nil {
		if db.IsNoRows(err) {
			return uuid.Nil, appErrors.NotFound("member", appErrors.WithOp("repo.organization_member.get_member_role"))
		}
		return uuid.Nil, fmt.Errorf("get member role (user=%s org=%s): %w", userID, orgID, err)
	}
	return id, nil
}

func (r *memberRepository) CountByOrganizationAndRole(ctx context.Context, orgID, roleID uuid.UUID) (int, error) {
	n, err := r.tm.Queries(ctx).CountMembersByOrganizationAndRole(ctx, gen.CountMembersByOrganizationAndRoleParams{
		OrganizationID: orgID,
		RoleID:         roleID,
	})
	if err != nil {
		return 0, fmt.Errorf("count members (org=%s role=%s): %w", orgID, roleID, err)
	}
	return int(n), nil
}

func (r *memberRepository) IsMember(ctx context.Context, userID, orgID uuid.UUID) (bool, error) {
	ok, err := r.tm.Queries(ctx).IsMember(ctx, gen.IsMemberParams{
		UserID:         userID,
		OrganizationID: orgID,
	})
	if err != nil {
		return false, fmt.Errorf("is-member check (user=%s org=%s): %w", userID, orgID, err)
	}
	return ok, nil
}

func (r *memberRepository) GetMemberCount(ctx context.Context, orgID uuid.UUID) (int, error) {
	n, err := r.tm.Queries(ctx).CountMembersByOrganization(ctx, orgID)
	if err != nil {
		return 0, fmt.Errorf("count members for org %s: %w", orgID, err)
	}
	return int(n), nil
}

// ----- gen ↔ domain boundary -----------------------------------------

func memberFromRow(row *gen.OrganizationMember) *orgDomain.Member {
	return &orgDomain.Member{
		UserID:         row.UserID,
		OrganizationID: row.OrganizationID,
		RoleID:         row.RoleID,
		JoinedAt:       row.JoinedAt,
		InvitedBy:      row.InvitedBy,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
		DeletedAt:      row.DeletedAt,
	}
}

func membersFromRows(rows []gen.OrganizationMember) []*orgDomain.Member {
	out := make([]*orgDomain.Member, 0, len(rows))
	for i := range rows {
		out = append(out, memberFromRow(&rows[i]))
	}
	return out
}
