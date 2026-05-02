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

// Create executes the UPSERT-WHERE statement defined by the
// CreateProjectMember query. Returns the affected-row count: 1 means
// the row was inserted (fresh add) or updated in-place (orphan repair);
// 0 means the row exists AND is visible (active org membership), so the
// UPSERT WHERE-clause refused to overwrite — caller's intended add is a
// duplicate or lost a concurrent race. See the SQL query comment block
// + CLAUDE.md 2026-04-30 (project-rbac).
func (r *projectMemberRepository) Create(ctx context.Context, m *authDomain.ProjectMember) (int64, error) {
	if m.JoinedAt.IsZero() {
		m.JoinedAt = time.Now()
	}
	rows, err := r.tm.Queries(ctx).CreateProjectMember(ctx, gen.CreateProjectMemberParams{
		UserID:    m.UserID,
		ProjectID: m.ProjectID,
		RoleID:    m.RoleID,
		JoinedAt:  m.JoinedAt,
	})
	if err != nil {
		// The UPSERT-WHERE clause means a unique violation should be
		// unreachable on the create path (Postgres serializes the
		// upsert decision at the unique index). Keep the classifier
		// for defense in depth so any future write path that still
		// uses plain INSERT gets the right wire shape.
		if appErrors.IsUniqueViolation(err) {
			return 0, appErrors.AlreadyExists("project_member",
				appErrors.WithOp("repo.project_member.create"))
		}
		return 0, appErrors.Internal("create project member", err,
			appErrors.WithOp("repo.project_member.create"))
	}
	return rows, nil
}

func (r *projectMemberRepository) GetByUserAndProject(ctx context.Context, userID, projectID uuid.UUID) (*authDomain.ProjectMember, error) {
	row, err := r.tm.Queries(ctx).GetProjectMemberByUserAndProject(ctx, gen.GetProjectMemberByUserAndProjectParams{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		if db.IsNoRows(err) {
			return nil, appErrors.NotFound("project_member", appErrors.WithOp("repo.project_member.get_by_user_and_project"))
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

// DeleteAllInOrgForUser cascades the org-removal to every
// project_members row this user holds within the org's projects.
// Single-statement persistence operation; the caller (application
// service) wraps this + OrganizationMemberRepository.Delete in
// WithinTransaction for atomicity. Schema note: project_members has
// no organization_id column, so the link is resolved via
// projects.organization_id.
func (r *projectMemberRepository) DeleteAllInOrgForUser(ctx context.Context, userID, orgID uuid.UUID) error {
	if err := r.tm.Queries(ctx).DeleteProjectMembersByUserAndOrg(ctx, gen.DeleteProjectMembersByUserAndOrgParams{
		UserID:         userID,
		OrganizationID: orgID,
	}); err != nil {
		return fmt.Errorf("cascade-delete project_members (user=%s org=%s): %w", userID, orgID, err)
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

// ListUserEffectivePermissionsInScope returns the permissions the user
// holds for (orgID, optional projectID), applying additive semantics:
// the org role's project-tier projection UNIONs with any
// project_members override grant; both branches require active org
// membership. Pass uuid.Nil as projectID to skip the project layer.
// See ADR-0001 for the migration history (round 11 OVERRIDE → round 24
// additive).
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
