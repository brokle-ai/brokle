package organization

import (
	"context"
	"time"

	orgDomain "brokle/internal/core/domain/organization"
	"brokle/pkg/ulid"
	appErrors "brokle/pkg/errors"
)

// environmentService implements the orgDomain.EnvironmentService interface
type environmentService struct {
	envRepo     orgDomain.EnvironmentRepository
	projectRepo orgDomain.ProjectRepository
	memberRepo  orgDomain.MemberRepository
}

// NewEnvironmentService creates a new environment service instance
func NewEnvironmentService(
	envRepo orgDomain.EnvironmentRepository,
	projectRepo orgDomain.ProjectRepository,
	memberRepo orgDomain.MemberRepository,
) orgDomain.EnvironmentService {
	return &environmentService{
		envRepo:     envRepo,
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
	}
}

// CreateEnvironment creates a new environment in a project
func (s *environmentService) CreateEnvironment(ctx context.Context, projectID ulid.ULID, req *orgDomain.CreateEnvironmentRequest) (*orgDomain.Environment, error) {
	// Verify project exists
	_, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Project not found")
	}

	// Check if environment with slug already exists in project
	existing, _ := s.envRepo.GetBySlug(ctx, projectID, req.Slug)
	if existing != nil {
		return nil, appErrors.NewConflictError("Environment with this slug already exists in project")
	}

	// Create environment
	environment := orgDomain.NewEnvironment(projectID, req.Name, req.Slug)
	err = s.envRepo.Create(ctx, environment)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to create environment", err)
	}


	return environment, nil
}

// GetEnvironment retrieves an environment by ID
func (s *environmentService) GetEnvironment(ctx context.Context, envID ulid.ULID) (*orgDomain.Environment, error) {
	return s.envRepo.GetByID(ctx, envID)
}

// GetEnvironmentBySlug retrieves an environment by project and slug
func (s *environmentService) GetEnvironmentBySlug(ctx context.Context, projectID ulid.ULID, slug string) (*orgDomain.Environment, error) {
	return s.envRepo.GetBySlug(ctx, projectID, slug)
}

// UpdateEnvironment updates environment details
func (s *environmentService) UpdateEnvironment(ctx context.Context, envID ulid.ULID, req *orgDomain.UpdateEnvironmentRequest) error {
	environment, err := s.envRepo.GetByID(ctx, envID)
	if err != nil {
		return appErrors.NewNotFoundError("Environment not found")
	}

	// Update fields if provided
	if req.Name != nil {
		environment.Name = *req.Name
	}

	environment.UpdatedAt = time.Now()

	err = s.envRepo.Update(ctx, environment)
	if err != nil {
		return appErrors.NewInternalError("Failed to update environment", err)
	}


	return nil
}

// DeleteEnvironment soft deletes an environment
func (s *environmentService) DeleteEnvironment(ctx context.Context, envID ulid.ULID) error {
	// Verify environment exists before deletion
	_, err := s.envRepo.GetByID(ctx, envID)
	if err != nil {
		return appErrors.NewNotFoundError("Environment not found")
	}

	err = s.envRepo.Delete(ctx, envID)
	if err != nil {
		return appErrors.NewInternalError("Failed to delete environment", err)
	}


	return nil
}

// GetEnvironmentsByProject retrieves all environments for a project
func (s *environmentService) GetEnvironmentsByProject(ctx context.Context, projectID ulid.ULID) ([]*orgDomain.Environment, error) {
	return s.envRepo.GetByProjectID(ctx, projectID)
}

// GetEnvironmentCount returns the number of environments in a project
func (s *environmentService) GetEnvironmentCount(ctx context.Context, projectID ulid.ULID) (int, error) {
	environments, err := s.envRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return 0, appErrors.NewInternalError("Failed to get environments", err)
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
		environment := orgDomain.NewEnvironment(projectID, env.name, env.slug)
		err := s.envRepo.Create(ctx, environment)
		if err != nil {
			return appErrors.NewInternalError("Failed to create "+env.name+" environment", err)
		}
	}

	return nil
}

// CanUserAccessEnvironment checks if user can access an environment
func (s *environmentService) CanUserAccessEnvironment(ctx context.Context, userID, envID ulid.ULID) (bool, error) {
	// Get environment to find project
	environment, err := s.envRepo.GetByID(ctx, envID)
	if err != nil {
		return false, appErrors.NewNotFoundError("Environment not found")
	}

	// Get project to find organization
	project, err := s.projectRepo.GetByID(ctx, environment.ProjectID)
	if err != nil {
		return false, appErrors.NewNotFoundError("Project not found")
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
		return appErrors.NewForbiddenError("User does not have access to this environment")
	}
	return nil
}

// GetEnvironmentOrganization returns the organization ID for an environment
func (s *environmentService) GetEnvironmentOrganization(ctx context.Context, envID ulid.ULID) (ulid.ULID, error) {
	// Get environment to find project
	environment, err := s.envRepo.GetByID(ctx, envID)
	if err != nil {
		return ulid.ULID{}, appErrors.NewNotFoundError("Environment not found")
	}

	// Get project to find organization
	project, err := s.projectRepo.GetByID(ctx, environment.ProjectID)
	if err != nil {
		return ulid.ULID{}, appErrors.NewNotFoundError("Project not found")
	}

	return project.OrganizationID, nil
}