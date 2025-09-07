package seeder

import (
	"context"
	"fmt"
	"log"

	"brokle/internal/core/domain/organization"
	"brokle/pkg/ulid"
)

// ProjectSeeder handles seeding of project and environment data
type ProjectSeeder struct {
	projectRepo     organization.ProjectRepository
	environmentRepo organization.EnvironmentRepository
}

// NewProjectSeeder creates a new ProjectSeeder instance
func NewProjectSeeder(projectRepo organization.ProjectRepository, environmentRepo organization.EnvironmentRepository) *ProjectSeeder {
	return &ProjectSeeder{
		projectRepo:     projectRepo,
		environmentRepo: environmentRepo,
	}
}

// SeedProjects seeds projects and their environments from the provided seed data
func (ps *ProjectSeeder) SeedProjects(ctx context.Context, projectSeeds []ProjectSeed, entityMaps *EntityMaps, verbose bool) error {
	if verbose {
		log.Printf("📂 Seeding %d projects...", len(projectSeeds))
	}

	for _, projectSeed := range projectSeeds {
		// Get organization ID
		orgID, ok := entityMaps.Organizations[projectSeed.OrganizationSlug]
		if !ok {
			return fmt.Errorf("organization %s not found for project %s", projectSeed.OrganizationSlug, projectSeed.Name)
		}

		// Check if project already exists by getting all projects and filtering
		// Since there's no GetByNameAndOrganization method, we'll use a different approach
		projectSlug := generateSlug(projectSeed.Name)
		existing, err := ps.projectRepo.GetBySlug(ctx, orgID, projectSlug)
		if err == nil && existing != nil {
			if verbose {
				log.Printf("   Project %s already exists, skipping", projectSeed.Name)
			}
			projectKey := fmt.Sprintf("%s:%s", projectSeed.OrganizationSlug, projectSeed.Name)
			entityMaps.Projects[projectKey] = existing.ID
			
			// Still need to process environments for existing project
			if err := ps.seedEnvironments(ctx, existing.ID, projectSeed.Environments, entityMaps, verbose); err != nil {
				return fmt.Errorf("failed to seed environments for existing project %s: %w", projectSeed.Name, err)
			}
			continue
		}

		// Project slug already generated above

		// Create project entity
		project := &organization.Project{
			ID:             ulid.New(),
			OrganizationID: orgID,
			Name:           projectSeed.Name,
			Slug:           projectSlug,
			Description:    projectSeed.Description,
		}

		// Create project in database
		if err := ps.projectRepo.Create(ctx, project); err != nil {
			return fmt.Errorf("failed to create project %s: %w", projectSeed.Name, err)
		}

		// Store project ID for later reference
		projectKey := fmt.Sprintf("%s:%s", projectSeed.OrganizationSlug, projectSeed.Name)
		entityMaps.Projects[projectKey] = project.ID

		// Seed environments for this project
		if err := ps.seedEnvironments(ctx, project.ID, projectSeed.Environments, entityMaps, verbose); err != nil {
			return fmt.Errorf("failed to seed environments for project %s: %w", projectSeed.Name, err)
		}

		if verbose {
			log.Printf("   ✅ Created project: %s (%s) with %d environments", project.Name, projectSlug, len(projectSeed.Environments))
		}
	}

	if verbose {
		log.Printf("✅ Projects seeded successfully")
	}
	return nil
}

// seedEnvironments seeds environments for a specific project
func (ps *ProjectSeeder) seedEnvironments(ctx context.Context, projectID ulid.ULID, envSeeds []EnvironmentSeed, entityMaps *EntityMaps, verbose bool) error {
	for _, envSeed := range envSeeds {
		// Generate environment slug if not provided
		envSlug := envSeed.Slug
		if envSlug == "" {
			envSlug = generateSlug(envSeed.Name)
		}

		// Check if environment already exists
		existing, err := ps.environmentRepo.GetBySlug(ctx, projectID, envSlug)
		if err == nil && existing != nil {
			if verbose {
				log.Printf("     Environment %s already exists, skipping", envSeed.Name)
			}
			envKey := fmt.Sprintf("%s:%s", projectID.String(), envSeed.Name)
			entityMaps.Environments[envKey] = existing.ID
			continue
		}

		// Create environment entity
		environment := &organization.Environment{
			ID:        ulid.New(),
			ProjectID: projectID,
			Name:      envSeed.Name,
			Slug:      envSlug,
		}

		// Create environment in database
		if err := ps.environmentRepo.Create(ctx, environment); err != nil {
			return fmt.Errorf("failed to create environment %s: %w", envSeed.Name, err)
		}

		// Store environment ID for later reference
		envKey := fmt.Sprintf("%s:%s", projectID.String(), envSeed.Name)
		entityMaps.Environments[envKey] = environment.ID

		if verbose {
			log.Printf("     ✅ Created environment: %s (%s)", environment.Name, environment.Slug)
		}
	}

	return nil
}

// generateSlug creates a URL-friendly slug from a name
func generateSlug(name string) string {
	// Simple slug generation - convert to lowercase and replace spaces with hyphens
	// In a real implementation, you might want to use a proper slug library
	slug := ""
	for _, r := range name {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			slug += string(r)
		case r >= 'A' && r <= 'Z':
			slug += string(r + 32) // Convert to lowercase
		case r == ' ' || r == '_':
			if len(slug) > 0 && slug[len(slug)-1] != '-' {
				slug += "-"
			}
		}
	}
	
	// Remove trailing hyphens
	for len(slug) > 0 && slug[len(slug)-1] == '-' {
		slug = slug[:len(slug)-1]
	}
	
	if slug == "" {
		slug = "unnamed"
	}
	
	return slug
}