package seeder

import (
	"context"
	"fmt"
	"log"

	"brokle/internal/core/domain/organization"
	"brokle/pkg/ulid"
)

// OrganizationSeeder handles seeding of organization data
type OrganizationSeeder struct {
	repo organization.OrganizationRepository
}

// NewOrganizationSeeder creates a new OrganizationSeeder instance
func NewOrganizationSeeder(repo organization.OrganizationRepository) *OrganizationSeeder {
	return &OrganizationSeeder{repo: repo}
}

// SeedOrganizations seeds organizations from the provided seed data
func (os *OrganizationSeeder) SeedOrganizations(ctx context.Context, orgSeeds []OrganizationSeed, entityMaps *EntityMaps, verbose bool) error {
	if verbose {
		log.Printf("🏢 Seeding %d organizations...", len(orgSeeds))
	}

	for _, orgSeed := range orgSeeds {
		// Check if organization already exists
		existing, err := os.repo.GetBySlug(ctx, orgSeed.Slug)
		if err == nil && existing != nil {
			if verbose {
				log.Printf("   Organization %s already exists, skipping", orgSeed.Slug)
			}
			entityMaps.Organizations[orgSeed.Slug] = existing.ID
			continue
		}

		// Create organization entity
		orgEntity := &organization.Organization{
			ID:                 ulid.New(),
			Name:               orgSeed.Name,
			Slug:               orgSeed.Slug,
			BillingEmail:       orgSeed.BillingEmail,
			Plan:               orgSeed.Plan,
			SubscriptionStatus: orgSeed.SubscriptionStatus,
		}

		// Set defaults if not provided
		if orgEntity.Plan == "" {
			orgEntity.Plan = "free"
		}
		if orgEntity.SubscriptionStatus == "" {
			orgEntity.SubscriptionStatus = "active"
		}

		// Create organization in database
		if err := os.repo.Create(ctx, orgEntity); err != nil {
			return fmt.Errorf("failed to create organization %s: %w", orgSeed.Slug, err)
		}

		// Store organization ID for later reference
		entityMaps.Organizations[orgSeed.Slug] = orgEntity.ID

		if verbose {
			log.Printf("   ✅ Created organization: %s (%s) - Plan: %s", orgEntity.Name, orgEntity.Slug, orgEntity.Plan)
		}
	}

	if verbose {
		log.Printf("✅ Organizations seeded successfully")
	}
	return nil
}