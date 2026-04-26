package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	"brokle/internal/infrastructure/db"
	"brokle/internal/infrastructure/db/gen"
)

// projectMemberRepository is the pgx+sqlc implementation of
// authDomain.ProjectMemberRepository. Mirrors organizationMemberRepository
// at a smaller scope: project_members has no soft-delete, no invited_by,
// no created_at/updated_at, and only one optional status (active).
type projectMemberRepository struct {
	tm *db.TxManager
}

// NewProjectMemberRepository returns the pgx-backed repository.
func NewProjectMemberRepository(tm *db.TxManager) authDomain.ProjectMemberRepository {
	return &projectMemberRepository{tm: tm}
}

// ----- CRUD ----------------------------------------------------------

func (r *projectMemberRepository) Create(ctx context.Context, m *authDomain.ProjectMember) error {
	if m.JoinedAt.IsZero() {
		m.JoinedAt = time.Now()
	}
	if m.Status == "" {
		m.Status = authDomain.MemberStatusActive
	}
	if err := r.tm.Queries(ctx).CreateProjectMember(ctx, gen.CreateProjectMemberParams{
		UserID:    m.UserID,
		ProjectID: m.ProjectID,
		RoleID:    m.RoleID,
		Status:    m.Status,
		JoinedAt:  m.JoinedAt,
	}); err != nil {
		return fmt.Errorf("create project member (user=%s project=%s): %w", m.UserID, m.ProjectID, err)
	}
	return nil
}

func (r *projectMemberRepository) GetByUserAndProject(ctx context.Context, userID, projectID uuid.UUID) (*authDomain.ProjectMember, error) {
	row, err := r.tm.Queries(ctx).GetProjectMemberByUserAndProject(ctx, gen.GetProjectMemberByUserAndProjectParams{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		if db.IsNoRows(err) {
			return nil, fmt.Errorf("get project member (user=%s project=%s): %w", userID, projectID, authDomain.ErrNotFound)
		}
		return nil, fmt.Errorf("get project member (user=%s project=%s): %w", userID, projectID, err)
	}
	return projectMemberFromRow(&row), nil
}

func (r *projectMemberRepository) UpdateRole(ctx context.Context, userID, projectID, roleID uuid.UUID) error {
	if err := r.tm.Queries(ctx).UpdateProjectMemberRole(ctx, gen.UpdateProjectMemberRoleParams{
		UserID:    userID,
		ProjectID: projectID,
		RoleID:    roleID,
	}); err != nil {
		return fmt.Errorf("update project member role (user=%s project=%s): %w", userID, projectID, err)
	}
	return nil
}

func (r *projectMemberRepository) Delete(ctx context.Context, userID, projectID uuid.UUID) error {
	if err := r.tm.Queries(ctx).DeleteProjectMember(ctx, gen.DeleteProjectMemberParams{
		UserID:    userID,
		ProjectID: projectID,
	}); err != nil {
		return fmt.Errorf("delete project member (user=%s project=%s): %w", userID, projectID, err)
	}
	return nil
}

// ----- Membership queries -------------------------------------------

func (r *projectMemberRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*authDomain.ProjectMember, error) {
	rows, err := r.tm.Queries(ctx).ListProjectMembersByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project members for project %s: %w", projectID, err)
	}
	return projectMembersFromRows(rows), nil
}

func (r *projectMemberRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*authDomain.ProjectMember, error) {
	rows, err := r.tm.Queries(ctx).ListProjectMembersByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list project memberships for user %s: %w", userID, err)
	}
	return projectMembersFromRows(rows), nil
}

func (r *projectMemberRepository) IsMember(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	ok, err := r.tm.Queries(ctx).IsProjectMember(ctx, gen.IsProjectMemberParams{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		return false, fmt.Errorf("check project member existence (user=%s project=%s): %w", userID, projectID, err)
	}
	return ok, nil
}

func (r *projectMemberRepository) GetMemberCount(ctx context.Context, projectID uuid.UUID) (int, error) {
	n, err := r.tm.Queries(ctx).CountProjectMembersByProject(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("count project members for project %s: %w", projectID, err)
	}
	return int(n), nil
}

// ----- Effective-permission resolution -----------------------------------

// ListUserEffectivePermissionsInScope returns the union of permissions the
// user holds for (orgID, optional projectID), applying Langfuse MAX
// semantics. Pass uuid.Nil as projectID to resolve org-only.
func (r *projectMemberRepository) ListUserEffectivePermissionsInScope(ctx context.Context, userID, orgID, projectID uuid.UUID) ([]string, error) {
	perms, err := r.tm.Queries(ctx).ListUserEffectivePermissionsInScope(ctx, gen.ListUserEffectivePermissionsInScopeParams{
		UserID:         userID,
		OrganizationID: orgID,
		ProjectID:      projectID,
	})
	if err != nil {
		return nil, fmt.Errorf("list effective permissions in scope (user=%s org=%s project=%s): %w", userID, orgID, projectID, err)
	}
	return perms, nil
}

// ----- gen ↔ domain boundary ----------------------------------------

func projectMemberFromRow(row *gen.ProjectMember) *authDomain.ProjectMember {
	return &authDomain.ProjectMember{
		UserID:    row.UserID,
		ProjectID: row.ProjectID,
		RoleID:    row.RoleID,
		Status:    row.Status,
		JoinedAt:  row.JoinedAt,
	}
}

func projectMembersFromRows(rows []gen.ProjectMember) []*authDomain.ProjectMember {
	out := make([]*authDomain.ProjectMember, 0, len(rows))
	for i := range rows {
		out = append(out, projectMemberFromRow(&rows[i]))
	}
	return out
}
