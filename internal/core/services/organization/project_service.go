package organization

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	orgDomain "brokle/internal/core/domain/organization"
	appErrors "brokle/pkg/errors"
)

// ProjectService manages project lifecycle within an organization.
type ProjectService struct {
	projectRepo orgDomain.ProjectRepository
	orgRepo     orgDomain.OrganizationRepository
	memberRepo  orgDomain.MemberRepository
}

// NewProjectService creates a new project service instance
func NewProjectService(
	projectRepo orgDomain.ProjectRepository,
	orgRepo orgDomain.OrganizationRepository,
	memberRepo orgDomain.MemberRepository,
) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		orgRepo:     orgRepo,
		memberRepo:  memberRepo,
	}
}

// CreateProject creates a new project in an organization
func (s *ProjectService) CreateProject(ctx context.Context, orgID uuid.UUID, req *orgDomain.CreateProjectRequest) (*orgDomain.Project, error) {
	// Verify organization exists
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("organization not found")
	}

	// Create project (no slug - use UUID only)
	project := orgDomain.NewProject(orgID, req.Name, req.Description)
	err = s.projectRepo.Create(ctx, project)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to create project", err)
	}

	return project, nil
}

// GetProject retrieves a project by ID
func (s *ProjectService) GetProject(ctx context.Context, projectID uuid.UUID) (*orgDomain.Project, error) {
	return s.projectRepo.GetByID(ctx, projectID)
}

// GetProjectBySlug retrieves a project by organization and slug
func (s *ProjectService) GetProjectBySlug(ctx context.Context, orgID uuid.UUID, slug string) (*orgDomain.Project, error) {
	return s.projectRepo.GetBySlug(ctx, orgID, slug)
}

// UpdateProject updates project details
func (s *ProjectService) UpdateProject(ctx context.Context, projectID uuid.UUID, req *orgDomain.UpdateProjectRequest) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return appErrors.NewNotFoundError("project not found")
	}

	// Block updates to archived projects
	if project.IsArchived() {
		return appErrors.NewBadRequestError("cannot update archived project", "Unarchive the project first")
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
		return appErrors.NewInternalError("failed to update project", err)
	}

	return nil
}

// ArchiveProject archives a project (sets status to archived, read-only, reversible)
func (s *ProjectService) ArchiveProject(ctx context.Context, projectID uuid.UUID) error {
	// Verify project exists
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return appErrors.NewNotFoundError("project not found")
	}

	// Check if already archived
	if project.IsArchived() {
		return appErrors.NewBadRequestError("project is already archived", "")
	}

	// Archive the project
	project.Archive()

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return appErrors.NewInternalError("failed to archive project", err)
	}

	return nil
}

// UnarchiveProject unarchives a project (sets status back to active)
func (s *ProjectService) UnarchiveProject(ctx context.Context, projectID uuid.UUID) error {
	// Verify project exists
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return appErrors.NewNotFoundError("project not found")
	}

	// Check if not archived
	if project.IsActive() {
		return appErrors.NewBadRequestError("project is already active", "")
	}

	// Unarchive the project
	project.Unarchive()

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return appErrors.NewInternalError("failed to unarchive project", err)
	}

	return nil
}

// DeleteProject soft deletes a project
func (s *ProjectService) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	// Verify project exists before deletion
	_, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return appErrors.NewNotFoundError("project not found")
	}

	err = s.projectRepo.Delete(ctx, projectID)
	if err != nil {
		return appErrors.NewInternalError("failed to delete project", err)
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
		return 0, appErrors.NewInternalError("failed to get projects", err)
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
		if errors.Is(err, orgDomain.ErrProjectNotFound) {
			return false, appErrors.NewNotFoundError("project")
		}
		return false, appErrors.NewInternalError("failed to get project", err)
	}

	isMember, err := s.memberRepo.IsMember(ctx, userID, project.OrganizationID)
	if err != nil {
		return false, appErrors.NewInternalError("failed to check organization membership", err)
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
		return appErrors.NewForbiddenError("user does not have access to this project")
	}
	return nil
}
