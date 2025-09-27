package seeder

import (
	"context"
	"fmt"
	"log"

	"brokle/internal/core/domain/organization"
	"brokle/pkg/ulid"
)

// ProjectSeeder handles seeding of project data
type ProjectSeeder struct {
	projectRepo organization.ProjectRepository
}

// NewProjectSeeder creates a new ProjectSeeder instance
func NewProjectSeeder(projectRepo organization.ProjectRepository) *ProjectSeeder {
	return &ProjectSeeder{
		projectRepo: projectRepo,
	}
}

// SeedProjects seeds projects from the provided seed data
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

		if verbose {
			log.Printf("   ✅ Created project: %s (%s)", project.Name, projectSlug)
		}
	}

	if verbose {
		log.Printf("✅ Projects seeded successfully")
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