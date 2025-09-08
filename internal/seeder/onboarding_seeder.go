package seeder

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
)

// OnboardingSeeder handles seeding of onboarding question data
type OnboardingSeeder struct {
	repo user.Repository
}

// NewOnboardingSeeder creates a new OnboardingSeeder instance
func NewOnboardingSeeder(repo user.Repository) *OnboardingSeeder {
	return &OnboardingSeeder{repo: repo}
}

// SeedOnboardingQuestions seeds onboarding questions from the provided seed data
func (os *OnboardingSeeder) SeedOnboardingQuestions(ctx context.Context, questionSeeds []OnboardingSeed, entityMaps *EntityMaps, verbose bool) error {
	if verbose {
		log.Printf("❓ Seeding %d onboarding questions...", len(questionSeeds))
	}

	for _, questionSeed := range questionSeeds {
		// Check if question already exists for this step
		// Since we don't have a GetByStep method, we'll just attempt to create and handle duplicates
		
		// Convert options to JSON
		var optionsJSON json.RawMessage
		if len(questionSeed.Options) > 0 {
			optionsBytes, err := json.Marshal(questionSeed.Options)
			if err != nil {
				return fmt.Errorf("failed to marshal options for step %d: %w", questionSeed.Step, err)
			}
			optionsJSON = optionsBytes
		}

		// Create onboarding question entity
		question := &user.OnboardingQuestion{
			ID:           ulid.New(),
			Step:         questionSeed.Step,
			QuestionType: questionSeed.QuestionType,
			Title:        questionSeed.Title,
			Description:  questionSeed.Description,
			IsRequired:   questionSeed.IsRequired,
			Options:      optionsJSON,
			DisplayOrder: questionSeed.DisplayOrder,
			IsActive:     questionSeed.IsActive,
		}

		// Set defaults
		if question.DisplayOrder == 0 {
			question.DisplayOrder = questionSeed.Step // Use step as default display order
		}
		if !question.IsActive {
			question.IsActive = true // Default to active if not specified
		}

		// Validate question type
		validTypes := map[string]bool{
			"single_choice":   true,
			"multiple_choice": true,
			"text":           true,
			"skip_optional":  true,
		}
		if !validTypes[question.QuestionType] {
			return fmt.Errorf("invalid question type: %s for step %d", question.QuestionType, question.Step)
		}

		// Validate that choice questions have options
		if (question.QuestionType == "single_choice" || question.QuestionType == "multiple_choice") && len(question.Options) == 0 {
			return fmt.Errorf("choice question at step %d missing options", question.Step)
		}

		// Create onboarding question in database
		if err := os.repo.CreateOnboardingQuestion(ctx, question); err != nil {
			// If it fails due to duplicate, just log a warning and continue
			if verbose {
				log.Printf("   ⚠️  Could not create onboarding question for step %d (may already exist): %v", questionSeed.Step, err)
			}
			continue
		}

		if verbose {
			requiredText := "optional"
			if question.IsRequired {
				requiredText = "required"
			}
			log.Printf("   ✅ Created onboarding question: Step %d - %s (%s, %s)", 
				question.Step, 
				question.Title, 
				question.QuestionType, 
				requiredText,
			)
		}
	}

	if verbose {
		log.Printf("✅ Onboarding questions seeded successfully")
	}
	return nil
}