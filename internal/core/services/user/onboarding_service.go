package user

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
)

// onboardingService implements the user.OnboardingService interface
type onboardingService struct {
	userRepo  user.Repository
	auditRepo auth.AuditLogRepository
}

// NewOnboardingService creates a new onboarding service instance
func NewOnboardingService(
	userRepo user.Repository,
	auditRepo auth.AuditLogRepository,
) user.OnboardingService {
	return &onboardingService{
		userRepo:  userRepo,
		auditRepo: auditRepo,
	}
}

// GetOnboardingStatus retrieves user's onboarding progress
func (s *onboardingService) GetOnboardingStatus(ctx context.Context, userID ulid.ULID) (*user.OnboardingStatus, error) {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Define the standard onboarding flow
	allSteps := []user.OnboardingStep{
		user.OnboardingStepProfile,
		user.OnboardingStepPreferences,
		user.OnboardingStepOrganization,
		user.OnboardingStepProject,
		user.OnboardingStepComplete,
	}

	// Determine completed steps based on user data
	completedSteps := []user.OnboardingStep{}
	stepProgress := make(map[user.OnboardingStep]bool)

	// Check profile completion
	if existingUser.FirstName != "" && existingUser.LastName != "" {
		completedSteps = append(completedSteps, user.OnboardingStepProfile)
		stepProgress[user.OnboardingStepProfile] = true
	}

	// Check preferences (assume completed if user exists for now)
	completedSteps = append(completedSteps, user.OnboardingStepPreferences)
	stepProgress[user.OnboardingStepPreferences] = true

	// Check if onboarding is fully completed
	isCompleted := existingUser.OnboardingCompleted
	if isCompleted {
		completedSteps = allSteps
		for _, step := range allSteps {
			stepProgress[step] = true
		}
	}

	// Determine current step
	var currentStep user.OnboardingStep
	if !isCompleted {
		if len(completedSteps) < len(allSteps) {
			currentStep = allSteps[len(completedSteps)]
		} else {
			currentStep = user.OnboardingStepComplete
		}
	} else {
		currentStep = user.OnboardingStepComplete
	}

	// Calculate completion rate
	completionRate := (len(completedSteps) * 100) / len(allSteps)

	status := &user.OnboardingStatus{
		UserID:          userID,
		IsCompleted:     isCompleted,
		CompletedSteps:  completedSteps,
		CurrentStep:     currentStep,
		TotalSteps:      len(allSteps),
		CompletionRate:  completionRate,
		StepProgress:    stepProgress,
		StartedAt:       nil, // Would be set if we tracked start time
		CompletedAt:     nil, // Would be set if we tracked completion time
	}

	if isCompleted {
		completedAt := existingUser.CreatedAt.Format(time.RFC3339)
		status.CompletedAt = &completedAt
	}

	return status, nil
}

// CompleteOnboardingStep marks a specific onboarding step as completed
func (s *onboardingService) CompleteOnboardingStep(ctx context.Context, userID ulid.ULID, step user.OnboardingStep) error {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// For now, just create audit log since we don't store individual step progress
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.onboarding_step_completed", "onboarding", userID.String(),
		fmt.Sprintf(`{"step": "%s"}`, string(step)), "", ""))

	// Check if this is the final step
	if step == user.OnboardingStepComplete {
		return s.CompleteOnboarding(ctx, userID)
	}

	return nil
}

// CompleteOnboarding marks the entire onboarding process as completed
func (s *onboardingService) CompleteOnboarding(ctx context.Context, userID ulid.ULID) error {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if existingUser.OnboardingCompleted {
		return nil // Already completed
	}

	existingUser.OnboardingCompleted = true
	existingUser.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, existingUser)
	if err != nil {
		return fmt.Errorf("failed to complete onboarding: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.onboarding_completed", "onboarding", userID.String(), "", "", ""))

	return nil
}

// IsOnboardingCompleted checks if onboarding is completed
func (s *onboardingService) IsOnboardingCompleted(ctx context.Context, userID ulid.ULID) (bool, error) {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("user not found: %w", err)
	}

	return existingUser.OnboardingCompleted, nil
}

// RestartOnboarding restarts the onboarding process
func (s *onboardingService) RestartOnboarding(ctx context.Context, userID ulid.ULID) error {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	existingUser.OnboardingCompleted = false
	existingUser.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, existingUser)
	if err != nil {
		return fmt.Errorf("failed to restart onboarding: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.onboarding_restarted", "onboarding", userID.String(), "", "", ""))

	return nil
}

// GetOnboardingFlow returns the onboarding flow for a specific user type
func (s *onboardingService) GetOnboardingFlow(ctx context.Context, userType user.UserType) (*user.OnboardingFlow, error) {
	// Define flows based on user type
	var steps []user.OnboardingStep
	var optional []user.OnboardingStep

	switch userType {
	case user.UserTypeDeveloper:
		steps = []user.OnboardingStep{
			user.OnboardingStepProfile,
			user.OnboardingStepPreferences,
			user.OnboardingStepProject,
			user.OnboardingStepIntegration,
			user.OnboardingStepComplete,
		}
		optional = []user.OnboardingStep{
			user.OnboardingStepOrganization,
		}
	case user.UserTypeManager:
		steps = []user.OnboardingStep{
			user.OnboardingStepProfile,
			user.OnboardingStepOrganization,
			user.OnboardingStepProject,
			user.OnboardingStepComplete,
		}
		optional = []user.OnboardingStep{
			user.OnboardingStepPreferences,
			user.OnboardingStepIntegration,
		}
	case user.UserTypeAdmin:
		steps = []user.OnboardingStep{
			user.OnboardingStepProfile,
			user.OnboardingStepOrganization,
			user.OnboardingStepProject,
			user.OnboardingStepIntegration,
			user.OnboardingStepComplete,
		}
		optional = []user.OnboardingStep{
			user.OnboardingStepPreferences,
		}
	default:
		// Default flow for UserTypeAnalyst and others
		steps = []user.OnboardingStep{
			user.OnboardingStepProfile,
			user.OnboardingStepPreferences,
			user.OnboardingStepOrganization,
			user.OnboardingStepProject,
			user.OnboardingStepComplete,
		}
		optional = []user.OnboardingStep{
			user.OnboardingStepIntegration,
		}
	}

	return &user.OnboardingFlow{
		UserType: userType,
		Steps:    steps,
		Optional: optional,
	}, nil
}

// UpdateOnboardingPreferences updates user's onboarding preferences
func (s *onboardingService) UpdateOnboardingPreferences(ctx context.Context, userID ulid.ULID, req *user.UpdateOnboardingPreferencesRequest) error {
	// For now, just create audit log since onboarding preferences aren't fully implemented
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.onboarding_preferences_updated", "onboarding", userID.String(), "", "", ""))

	return nil
}