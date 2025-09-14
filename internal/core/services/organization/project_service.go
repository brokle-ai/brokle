package organization

import (
	"context"
	"time"

	orgDomain "brokle/internal/core/domain/organization"
	"brokle/pkg/ulid"
	appErrors "brokle/pkg/errors"
)

// projectService implements the orgDomain.ProjectService interface
type projectService struct {
	projectRepo orgDomain.ProjectRepository
	orgRepo     orgDomain.OrganizationRepository
	memberRepo  orgDomain.MemberRepository
}

// NewProjectService creates a new project service instance
func NewProjectService(
	projectRepo orgDomain.ProjectRepository,
	orgRepo orgDomain.OrganizationRepository,
	memberRepo orgDomain.MemberRepository,
) orgDomain.ProjectService {
	return &projectService{
		projectRepo: projectRepo,
		orgRepo:     orgRepo,
		memberRepo:  memberRepo,
	}
}

// CreateProject creates a new project in an organization
func (s *projectService) CreateProject(ctx context.Context, orgID ulid.ULID, req *orgDomain.CreateProjectRequest) (*orgDomain.Project, error) {
	// Verify organization exists
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Organization not found")
	}

	// Check if project with slug already exists in organization
	existing, _ := s.projectRepo.GetBySlug(ctx, orgID, req.Slug)
	if existing != nil {
		return nil, appErrors.NewConflictError("Project with this slug already exists in organization")
	}

	// Create project
	project := orgDomain.NewProject(orgID, req.Name, req.Slug, req.Description)
	err = s.projectRepo.Create(ctx, project)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to create project", err)
	}


	return project, nil
}

// GetProject retrieves a project by ID
func (s *projectService) GetProject(ctx context.Context, projectID ulid.ULID) (*orgDomain.Project, error) {
	return s.projectRepo.GetByID(ctx, projectID)
}

// GetProjectBySlug retrieves a project by organization and slug
func (s *projectService) GetProjectBySlug(ctx context.Context, orgID ulid.ULID, slug string) (*orgDomain.Project, error) {
	return s.projectRepo.GetBySlug(ctx, orgID, slug)
}

// UpdateProject updates project details
func (s *projectService) UpdateProject(ctx context.Context, projectID ulid.ULID, req *orgDomain.UpdateProjectRequest) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return appErrors.NewNotFoundError("Project not found")
	}

	// Update fields if provided
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}

	project.UpdatedAt = time.Now()

	err = s.projectRepo.Update(ctx, project)
	if err != nil {
		return appErrors.NewInternalError("Failed to update project", err)
	}


	return nil
}

// DeleteProject soft deletes a project
func (s *projectService) DeleteProject(ctx context.Context, projectID ulid.ULID) error {
	// Verify project exists before deletion
	_, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return appErrors.NewNotFoundError("Project not found")
	}

	err = s.projectRepo.Delete(ctx, projectID)
	if err != nil {
		return appErrors.NewInternalError("Failed to delete project", err)
	}


	return nil
}

// GetProjectsByOrganization retrieves all projects for an organization
func (s *projectService) GetProjectsByOrganization(ctx context.Context, orgID ulid.ULID) ([]*orgDomain.Project, error) {
	return s.projectRepo.GetByOrganizationID(ctx, orgID)
}

// GetProjectCount returns the number of projects in an organization
func (s *projectService) GetProjectCount(ctx context.Context, orgID ulid.ULID) (int, error) {
	projects, err := s.projectRepo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return 0, appErrors.NewInternalError("Failed to get projects", err)
	}
	return len(projects), nil
}

// CanUserAccessProject checks if user can access a project
func (s *projectService) CanUserAccessProject(ctx context.Context, userID, projectID ulid.ULID) (bool, error) {
	// Get project to find organization
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return false, appErrors.NewNotFoundError("Project not found")
	}

	// Check if user is a member of the organization
	return s.memberRepo.IsMember(ctx, userID, project.OrganizationID)
}

// ValidateProjectAccess validates if user can access a project (throws error if not)
func (s *projectService) ValidateProjectAccess(ctx context.Context, userID, projectID ulid.ULID) error {
	canAccess, err := s.CanUserAccessProject(ctx, userID, projectID)
	if err != nil {
		return err
	}
	if !canAccess {
		return appErrors.NewForbiddenError("User does not have access to this project")
	}
	return nil
}