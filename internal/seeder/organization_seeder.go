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
		// Create organization entity (no slug - use ULID only)
		orgEntity := &organization.Organization{
			ID:                 ulid.New(),
			Name:               orgSeed.Name,
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
			return fmt.Errorf("failed to create organization %s: %w", orgSeed.Name, err)
		}

		// Store organization ID for later reference (use name as key)
		entityMaps.Organizations[orgSeed.Name] = orgEntity.ID

		if verbose {
			log.Printf("   ✅ Created organization: %s (ID: %s) - Plan: %s", orgEntity.Name, orgEntity.ID.String(), orgEntity.Plan)
		}
	}

	if verbose {
		log.Printf("✅ Organizations seeded successfully")
	}
	return nil
}