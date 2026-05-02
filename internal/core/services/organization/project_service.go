package organization

import (
	"context"
	"time"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/common"
	orgDomain "brokle/internal/core/domain/organization"
	authService "brokle/internal/core/services/auth"
	appErrors "brokle/pkg/errors"
)

// ProjectService manages project lifecycle within an organization.
type ProjectService struct {
	projectRepo    orgDomain.ProjectRepository
	orgRepo        orgDomain.OrganizationRepository
	memberRepo     orgDomain.MemberRepository
	roles          *authService.RoleService
	projectMembers *authService.ProjectMemberService
	tx             common.Transactor
}

// NewProjectService creates a new project service instance
func NewProjectService(
	projectRepo orgDomain.ProjectRepository,
	orgRepo orgDomain.OrganizationRepository,
	memberRepo orgDomain.MemberRepository,
	roles *authService.RoleService,
	projectMembers *authService.ProjectMemberService,
	tx common.Transactor,
) *ProjectService {
	return &ProjectService{
		projectRepo:    projectRepo,
		orgRepo:        orgRepo,
		memberRepo:     memberRepo,
		roles:          roles,
		projectMembers: projectMembers,
		tx:             tx,
	}
}

// CreateProject creates a project and grants the creator project admin
// via a project_members admin grant row. Mirrors the GitHub / GitLab /
// Linear / Notion convention: the creator owns what they create
// regardless of their org-tier role.
//
// This is required for Brokle's per-tier verb minting (round 19) where
// a custom role can hold `org_projects:create` without `org_projects:list`
// or any project-tier perm. Without this grant, the creator cannot
// discover the project they just made — `mapOrgsWithProjects` filters
// it out (`org_projects:list` absent + `projects:read` floor absent).
//
// Granting the seeded `admin` template (scope_type=organization,
// scope_id=NULL) yields full project-tier perms including the
// `projects:read` floor; under the round-24 additive resolver this
// UNIONs with the creator's org-role projection — never reduces
// existing access. validateProjectAssignableRole accepts the template
// (scope_id=NULL satisfies the cross-tenant check).
func (s *ProjectService) CreateProject(
	ctx context.Context,
	orgID, userID uuid.UUID,
	req *orgDomain.CreateProjectRequest,
) (*orgDomain.Project, error) {
	// Verify organization exists
	if _, err := s.orgRepo.GetByID(ctx, orgID); err != nil {
		return nil, appErrors.NotFound("organization")
	}

	// Resolve the seeded `admin` template ONCE before opening the tx,
	// so a missing-seed condition surfaces as a clean Internal error
	// instead of leaving a partial project row + tx rollback log noise.
	adminRole, err := s.roles.GetRoleByNameAndScope(ctx, "admin", authDomain.ScopeOrganization)
	if err != nil {
		return nil, appErrors.Internal("admin role template not found; reseed required", err)
	}

	project := orgDomain.NewProject(orgID, req.Name, req.Description)

	// Atomic create-and-grant. WithinTransaction is reentrant so a
	// caller already in an outer tx (none today, but possible future
	// composition) reuses it. Rollback on either failure: no orphan
	// project, no orphan membership, idempotent on retry.
	err = s.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := s.projectRepo.Create(ctx, project); err != nil {
			return appErrors.Internal("failed to create project", err)
		}
		// AddCreatorGrant (NOT AddMember) — the creator-attribution row
		// is system-initiated and intentionally bypasses the redundant-
		// grant guard. Owner / admin org roles already inherit
		// admin-equivalent project-tier scopes; the row is written for
		// creator-retention across future role changes (ADR-0001 §3),
		// not for permission elevation. AddMember would 409 here.
		if _, err := s.projectMembers.AddCreatorGrant(ctx, userID, project.ID, orgID, adminRole.ID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return project, nil
}

// GetProject retrieves a project by ID.
//
// The repository returns a self-describing *domain.Error (Reason=NotFound
// when the row is absent, Reason=Internal otherwise), so this service
// method is a correct passthrough. response.WriteError reads the Reason
// and emits the right HTTP status. No sentinel-translation discipline
// required here — the bug class that broke `RequireProjectAccess` after
// Phase 6 (404 → 500 because the service forgot to translate) is now
// structurally impossible. See CLAUDE.md error-taxonomy.
func (s *ProjectService) GetProject(ctx context.Context, projectID uuid.UUID) (*orgDomain.Project, error) {
	return s.projectRepo.GetByID(ctx, projectID)
}

// GetProjectBySlug retrieves a project by organization and slug.
// Same self-describing-error passthrough contract as GetProject.
func (s *ProjectService) GetProjectBySlug(ctx context.Context, orgID uuid.UUID, slug string) (*orgDomain.Project, error) {
	return s.projectRepo.GetBySlug(ctx, orgID, slug)
}

// UpdateProject updates project details
func (s *ProjectService) UpdateProject(ctx context.Context, projectID uuid.UUID, req *orgDomain.UpdateProjectRequest) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return appErrors.NotFound("project")
	}

	// Block updates to archived projects — state-machine rejection
	// (HTTP 409). Sibling sites in this file (ArchiveProject /
	// UnarchiveProject already-state checks) use Conflict for the
	// same reason.
	if project.IsArchived() {
		return appErrors.Conflict("project", "cannot update archived project; unarchive it first")
	}

	// Update fields if provided
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = req.Description
	}

	project.UpdatedAt = time.Now()

	err = s.projectRepo.Update(ctx, project)
	if err != nil {
		return appErrors.Internal("failed to update project", err)
	}

	return nil
}

// ArchiveProject archives a project (sets status to archived, read-only, reversible)
func (s *ProjectService) ArchiveProject(ctx context.Context, projectID uuid.UUID) error {
	// Verify project exists
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return appErrors.NotFound("project")
	}

	// Check if already archived
	if project.IsArchived() {
		return appErrors.Conflict("project", "project is already archived")
	}

	// Archive the project
	project.Archive()

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return appErrors.Internal("failed to archive project", err)
	}

	return nil
}

// UnarchiveProject unarchives a project (sets status back to active)
func (s *ProjectService) UnarchiveProject(ctx context.Context, projectID uuid.UUID) error {
	// Verify project exists
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return appErrors.NotFound("project")
	}

	// Check if not archived
	if project.IsActive() {
		return appErrors.Conflict("project", "project is already active")
	}

	// Unarchive the project
	project.Unarchive()

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return appErrors.Internal("failed to unarchive project", err)
	}

	return nil
}

// DeleteProject soft deletes a project
func (s *ProjectService) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	// Verify project exists before deletion
	_, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return appErrors.NotFound("project")
	}

	err = s.projectRepo.Delete(ctx, projectID)
	if err != nil {
		return appErrors.Internal("failed to delete project", err)
	}

	return nil
}

// GetProjectsByOrganization retrieves all projects for an organization
func (s *ProjectService) GetProjectsByOrganization(ctx context.Context, orgID uuid.UUID) ([]*orgDomain.Project, error) {
	return s.projectRepo.GetByOrganizationID(ctx, orgID)
}

// GetProjectCount returns the number of projects in an organization
func (s *ProjectService) GetProjectCount(ctx context.Context, orgID uuid.UUID) (int, error) {
	projects, err := s.projectRepo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return 0, appErrors.Internal("failed to get projects", err)
	}
	return len(projects), nil
}

// CanUserAccessProject reports whether the user is a member of the project's
// organization. Returns an AppError so callers can forward it via
// response.Error and get the right HTTP status: 404 when the project does not
// exist, 500 on infrastructure errors, nil otherwise.
func (s *ProjectService) CanUserAccessProject(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		// Repo returns self-describing *domain.Error (Reason=NotFound when
		// the row is absent); pass through and let response.WriteError emit
		// the right HTTP status.
		return false, err
	}

	isMember, err := s.memberRepo.IsMember(ctx, userID, project.OrganizationID)
	if err != nil {
		return false, appErrors.Internal("failed to check organization membership", err)
	}
	return isMember, nil
}

// ValidateProjectAccess validates if user can access a project (throws error if not)
func (s *ProjectService) ValidateProjectAccess(ctx context.Context, userID, projectID uuid.UUID) error {
	canAccess, err := s.CanUserAccessProject(ctx, userID, projectID)
	if err != nil {
		return err
	}
	if !canAccess {
		return appErrors.PermissionDenied("project", "user does not have access to this project")
	}
	return nil
}
