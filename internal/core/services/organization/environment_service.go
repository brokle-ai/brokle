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

// environmentService implements the organization.EnvironmentService interface
type environmentService struct {
	envRepo     organization.EnvironmentRepository
	projectRepo organization.ProjectRepository
	memberRepo  organization.MemberRepository
	auditRepo   auth.AuditLogRepository
}

// NewEnvironmentService creates a new environment service instance
func NewEnvironmentService(
	envRepo organization.EnvironmentRepository,
	projectRepo organization.ProjectRepository,
	memberRepo organization.MemberRepository,
	auditRepo auth.AuditLogRepository,
) organization.EnvironmentService {
	return &environmentService{
		envRepo:     envRepo,
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
		auditRepo:   auditRepo,
	}
}

// CreateEnvironment creates a new environment in a project
func (s *environmentService) CreateEnvironment(ctx context.Context, projectID ulid.ULID, req *organization.CreateEnvironmentRequest) (*organization.Environment, error) {
	// Verify project exists
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Check if environment with slug already exists in project
	existing, _ := s.envRepo.GetBySlug(ctx, projectID, req.Slug)
	if existing != nil {
		return nil, errors.New("environment with this slug already exists in project")
	}

	// Create environment
	environment := organization.NewEnvironment(projectID, req.Name, req.Slug)
	err = s.envRepo.Create(ctx, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(nil, &project.OrganizationID, "environment.created", "environment", environment.ID.String(),
		fmt.Sprintf(`{"name": "%s", "slug": "%s", "project": "%s"}`, environment.Name, environment.Slug, project.Name), "", ""))

	return environment, nil
}

// GetEnvironment retrieves an environment by ID
func (s *environmentService) GetEnvironment(ctx context.Context, envID ulid.ULID) (*organization.Environment, error) {
	return s.envRepo.GetByID(ctx, envID)
}

// GetEnvironmentBySlug retrieves an environment by project and slug
func (s *environmentService) GetEnvironmentBySlug(ctx context.Context, projectID ulid.ULID, slug string) (*organization.Environment, error) {
	return s.envRepo.GetBySlug(ctx, projectID, slug)
}

// UpdateEnvironment updates environment details
func (s *environmentService) UpdateEnvironment(ctx context.Context, envID ulid.ULID, req *organization.UpdateEnvironmentRequest) error {
	environment, err := s.envRepo.GetByID(ctx, envID)
	if err != nil {
		return fmt.Errorf("environment not found: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		environment.Name = *req.Name
	}

	environment.UpdatedAt = time.Now()

	err = s.envRepo.Update(ctx, environment)
	if err != nil {
		return fmt.Errorf("failed to update environment: %w", err)
	}

	// Get project for audit log
	project, _ := s.projectRepo.GetByID(ctx, environment.ProjectID)
	var orgID *ulid.ULID
	if project != nil {
		orgID = &project.OrganizationID
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(nil, orgID, "environment.updated", "environment", envID.String(), "", "", ""))

	return nil
}

// DeleteEnvironment soft deletes an environment
func (s *environmentService) DeleteEnvironment(ctx context.Context, envID ulid.ULID) error {
	environment, err := s.envRepo.GetByID(ctx, envID)
	if err != nil {
		return fmt.Errorf("environment not found: %w", err)
	}

	err = s.envRepo.Delete(ctx, envID)
	if err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}

	// Get project for audit log
	project, _ := s.projectRepo.GetByID(ctx, environment.ProjectID)
	var orgID *ulid.ULID
	if project != nil {
		orgID = &project.OrganizationID
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(nil, orgID, "environment.deleted", "environment", envID.String(),
		fmt.Sprintf(`{"name": "%s"}`, environment.Name), "", ""))

	return nil
}

// GetEnvironmentsByProject retrieves all environments for a project
func (s *environmentService) GetEnvironmentsByProject(ctx context.Context, projectID ulid.ULID) ([]*organization.Environment, error) {
	return s.envRepo.GetByProjectID(ctx, projectID)
}

// GetEnvironmentCount returns the number of environments in a project
func (s *environmentService) GetEnvironmentCount(ctx context.Context, projectID ulid.ULID) (int, error) {
	environments, err := s.envRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get environments: %w", err)
	}
	return len(environments), nil
}

// CreateDefaultEnvironments creates default environments for a project
func (s *environmentService) CreateDefaultEnvironments(ctx context.Context, projectID ulid.ULID) error {
	// Create default environments: development, staging, production
	environments := []struct {
		name, slug string
	}{
		{"Development", "dev"},
		{"Staging", "staging"},
		{"Production", "prod"},
	}

	for _, env := range environments {
		environment := organization.NewEnvironment(projectID, env.name, env.slug)
		err := s.envRepo.Create(ctx, environment)
		if err != nil {
			return fmt.Errorf("failed to create %s environment: %w", env.name, err)
		}
	}

	return nil
}

// CanUserAccessEnvironment checks if user can access an environment
func (s *environmentService) CanUserAccessEnvironment(ctx context.Context, userID, envID ulid.ULID) (bool, error) {
	// Get environment to find project
	environment, err := s.envRepo.GetByID(ctx, envID)
	if err != nil {
		return false, fmt.Errorf("environment not found: %w", err)
	}

	// Get project to find organization
	project, err := s.projectRepo.GetByID(ctx, environment.ProjectID)
	if err != nil {
		return false, fmt.Errorf("project not found: %w", err)
	}

	// Check if user is a member of the organization
	return s.memberRepo.IsMember(ctx, userID, project.OrganizationID)
}

// ValidateEnvironmentAccess validates if user can access an environment (throws error if not)
func (s *environmentService) ValidateEnvironmentAccess(ctx context.Context, userID, envID ulid.ULID) error {
	canAccess, err := s.CanUserAccessEnvironment(ctx, userID, envID)
	if err != nil {
		return err
	}
	if !canAccess {
		return errors.New("user does not have access to this environment")
	}
	return nil
}

// GetEnvironmentOrganization returns the organization ID for an environment
func (s *environmentService) GetEnvironmentOrganization(ctx context.Context, envID ulid.ULID) (ulid.ULID, error) {
	// Get environment to find project
	environment, err := s.envRepo.GetByID(ctx, envID)
	if err != nil {
		return ulid.ULID{}, fmt.Errorf("environment not found: %w", err)
	}

	// Get project to find organization
	project, err := s.projectRepo.GetByID(ctx, environment.ProjectID)
	if err != nil {
		return ulid.ULID{}, fmt.Errorf("project not found: %w", err)
	}

	return project.OrganizationID, nil
}