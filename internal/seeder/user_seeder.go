package seeder

import (
	"context"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"

	"brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
)

// UserSeeder handles seeding of user data
type UserSeeder struct {
	repo user.Repository
}

// NewUserSeeder creates a new UserSeeder instance
func NewUserSeeder(repo user.Repository) *UserSeeder {
	return &UserSeeder{repo: repo}
}

// SeedUsers seeds users from the provided seed data
func (us *UserSeeder) SeedUsers(ctx context.Context, userSeeds []UserSeed, entityMaps *EntityMaps, verbose bool) error {
	if verbose {
		log.Printf("👥 Seeding %d users...", len(userSeeds))
	}

	for _, userSeed := range userSeeds {
		// Check if user already exists
		existing, err := us.repo.GetByEmail(ctx, userSeed.Email)
		if err == nil && existing != nil {
			if verbose {
				log.Printf("   User %s already exists, skipping", userSeed.Email)
			}
			entityMaps.Users[userSeed.Email] = existing.ID
			continue
		}

		// Hash password
		hashedPassword, err := us.hashPassword(userSeed.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password for user %s: %w", userSeed.Email, err)
		}

		// Create user entity
		userEntity := &user.User{
			ID:              ulid.New(),
			Email:           userSeed.Email,
			FirstName:       userSeed.FirstName,
			LastName:        userSeed.LastName,
			Password:        hashedPassword,
			IsEmailVerified: userSeed.EmailVerified,
			IsActive:        userSeed.IsActive,
			Timezone:        userSeed.Timezone,
			Language:        userSeed.Language,
		}

		// Set defaults if not provided
		if userEntity.Timezone == "" {
			userEntity.Timezone = "UTC"
		}
		if userEntity.Language == "" {
			userEntity.Language = "en"
		}

		// Create user in database
		if err := us.repo.Create(ctx, userEntity); err != nil {
			return fmt.Errorf("failed to create user %s: %w", userSeed.Email, err)
		}

		// Store user ID for later reference
		entityMaps.Users[userSeed.Email] = userEntity.ID

		if verbose {
			log.Printf("   ✅ Created user: %s %s (%s)", userEntity.FirstName, userEntity.LastName, userEntity.Email)
		}
	}

	if verbose {
		log.Printf("✅ Users seeded successfully")
	}
	return nil
}

// hashPassword hashes a password using bcrypt
func (us *UserSeeder) hashPassword(password string) (string, error) {
	if password == "" {
		// Use a default password for development
		password = "password123"
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}
