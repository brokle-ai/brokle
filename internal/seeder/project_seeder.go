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
		// Get organization ID (use name as key instead of slug)
		orgID, ok := entityMaps.Organizations[projectSeed.OrganizationName]
		if !ok {
			return fmt.Errorf("organization %s not found for project %s", projectSeed.OrganizationName, projectSeed.Name)
		}

		// Create project entity (no slug - use ULID only)
		project := &organization.Project{
			ID:             ulid.New(),
			OrganizationID: orgID,
			Name:           projectSeed.Name,
			Description:    projectSeed.Description,
		}

		// Create project in database
		if err := ps.projectRepo.Create(ctx, project); err != nil {
			return fmt.Errorf("failed to create project %s: %w", projectSeed.Name, err)
		}

		// Store project ID for later reference (use name as key)
		projectKey := fmt.Sprintf("%s:%s", projectSeed.OrganizationName, projectSeed.Name)
		entityMaps.Projects[projectKey] = project.ID

		if verbose {
			log.Printf("   ✅ Created project: %s (ID: %s)", project.Name, project.ID.String())
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