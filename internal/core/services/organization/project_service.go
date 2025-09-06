package organization

import (
	"context"
	"errors"
	"fmt"
	"time"

	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/organization"
	"brokle/pkg/ulid"
)

// projectService implements the organization.ProjectService interface
type projectService struct {
	projectRepo organization.ProjectRepository
	orgRepo     organization.OrganizationRepository
	memberRepo  organization.MemberRepository
	auditRepo   auth.AuditLogRepository
}

// NewProjectService creates a new project service instance
func NewProjectService(
	projectRepo organization.ProjectRepository,
	orgRepo organization.OrganizationRepository,
	memberRepo organization.MemberRepository,
	auditRepo auth.AuditLogRepository,
) organization.ProjectService {
	return &projectService{
		projectRepo: projectRepo,
		orgRepo:     orgRepo,
		memberRepo:  memberRepo,
		auditRepo:   auditRepo,
	}
}

// CreateProject creates a new project in an organization
func (s *projectService) CreateProject(ctx context.Context, orgID ulid.ULID, req *organization.CreateProjectRequest) (*organization.Project, error) {
	// Verify organization exists
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Check if project with slug already exists in organization
	existing, _ := s.projectRepo.GetBySlug(ctx, orgID, req.Slug)
	if existing != nil {
		return nil, errors.New("project with this slug already exists in organization")
	}

	// Create project
	project := organization.NewProject(orgID, req.Name, req.Slug, req.Description)
	err = s.projectRepo.Create(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(nil, &orgID, "project.created", "project", project.ID.String(),
		fmt.Sprintf(`{"name": "%s", "slug": "%s"}`, project.Name, project.Slug), "", ""))

	return project, nil
}

// GetProject retrieves a project by ID
func (s *projectService) GetProject(ctx context.Context, projectID ulid.ULID) (*organization.Project, error) {
	return s.projectRepo.GetByID(ctx, projectID)
}

// GetProjectBySlug retrieves a project by organization and slug
func (s *projectService) GetProjectBySlug(ctx context.Context, orgID ulid.ULID, slug string) (*organization.Project, error) {
	return s.projectRepo.GetBySlug(ctx, orgID, slug)
}

// UpdateProject updates project details
func (s *projectService) UpdateProject(ctx context.Context, projectID ulid.ULID, req *organization.UpdateProjectRequest) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
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
		return fmt.Errorf("failed to update project: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(nil, &project.OrganizationID, "project.updated", "project", projectID.String(), "", "", ""))

	return nil
}

// DeleteProject soft deletes a project
func (s *projectService) DeleteProject(ctx context.Context, projectID ulid.ULID) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	err = s.projectRepo.Delete(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(nil, &project.OrganizationID, "project.deleted", "project", projectID.String(),
		fmt.Sprintf(`{"name": "%s"}`, project.Name), "", ""))

	return nil
}

// GetProjectsByOrganization retrieves all projects for an organization
func (s *projectService) GetProjectsByOrganization(ctx context.Context, orgID ulid.ULID) ([]*organization.Project, error) {
	return s.projectRepo.GetByOrganizationID(ctx, orgID)
}

// GetProjectCount returns the number of projects in an organization
func (s *projectService) GetProjectCount(ctx context.Context, orgID ulid.ULID) (int, error) {
	projects, err := s.projectRepo.GetByOrganizationID(ctx, orgID)
	if err != nil {
		return 0, fmt.Errorf("failed to get projects: %w", err)
	}
	return len(projects), nil
}

// CanUserAccessProject checks if user can access a project
func (s *projectService) CanUserAccessProject(ctx context.Context, userID, projectID ulid.ULID) (bool, error) {
	// Get project to find organization
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return false, fmt.Errorf("project not found: %w", err)
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
		return errors.New("user does not have access to this project")
	}
	return nil
}