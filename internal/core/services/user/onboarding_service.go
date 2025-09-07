package user

import (
	"context"
	"encoding/json"
	"fmt"

	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
)

// onboardingService implements the user.OnboardingService interface for dynamic onboarding
type onboardingService struct {
	userRepo  user.Repository
	auditRepo auth.AuditLogRepository
}

// NewOnboardingService creates a new dynamic onboarding service instance
func NewOnboardingService(
	userRepo user.Repository,
	auditRepo auth.AuditLogRepository,
) user.OnboardingService {
	return &onboardingService{
		userRepo:  userRepo,
		auditRepo: auditRepo,
	}
}

// GetActiveQuestions retrieves all active onboarding questions
func (s *onboardingService) GetActiveQuestions(ctx context.Context) ([]*user.OnboardingQuestion, error) {
	return s.userRepo.GetActiveOnboardingQuestions(ctx)
}

// GetQuestionByID retrieves a specific onboarding question by ID
func (s *onboardingService) GetQuestionByID(ctx context.Context, id ulid.ULID) (*user.OnboardingQuestion, error) {
	return s.userRepo.GetOnboardingQuestionByID(ctx, id)
}

// CreateQuestion creates a new onboarding question
func (s *onboardingService) CreateQuestion(ctx context.Context, req *user.CreateOnboardingQuestionRequest) (*user.OnboardingQuestion, error) {
	// Validate request
	if req.Title == "" {
		return nil, fmt.Errorf("question title is required")
	}

	if req.QuestionType != "single_choice" && req.QuestionType != "multiple_choice" && 
	   req.QuestionType != "text" && req.QuestionType != "skip_optional" {
		return nil, fmt.Errorf("invalid question type")
	}

	if req.Step <= 0 {
		return nil, fmt.Errorf("step must be greater than 0")
	}

	// Create question
	question := user.NewOnboardingQuestion(req)
	
	err := s.userRepo.CreateOnboardingQuestion(ctx, question)
	if err != nil {
		return nil, fmt.Errorf("failed to create onboarding question: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(nil, nil, "onboarding.question_created", "onboarding", question.ID.String(),
		fmt.Sprintf(`{"title": "%s", "type": "%s", "step": %d}`, req.Title, req.QuestionType, req.Step), "", ""))

	return question, nil
}

// GetUserResponses retrieves all user responses with parsed values
func (s *onboardingService) GetUserResponses(ctx context.Context, userID ulid.ULID) ([]*user.UserOnboardingResponseData, error) {
	responses, err := s.userRepo.GetUserOnboardingResponses(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user responses: %w", err)
	}

	var result []*user.UserOnboardingResponseData
	for _, response := range responses {
		// Parse the JSON response value
		var responseValue interface{}
		if response.ResponseValue != nil {
			err := json.Unmarshal(response.ResponseValue, &responseValue)
			if err != nil {
				responseValue = string(response.ResponseValue) // Fallback to raw string
			}
		}

		result = append(result, &user.UserOnboardingResponseData{
			ID:            response.ID,
			UserID:        response.UserID,
			QuestionID:    response.QuestionID,
			ResponseValue: responseValue,
			Skipped:       response.Skipped,
		})
	}

	return result, nil
}

// SubmitResponse submits a single response to an onboarding question
func (s *onboardingService) SubmitResponse(ctx context.Context, userID, questionID ulid.ULID, responseValue interface{}, skipped bool) error {
	// Verify question exists
	question, err := s.userRepo.GetOnboardingQuestionByID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("question not found: %w", err)
	}

	// Verify user exists
	_, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Validate response if not skipped
	if !skipped {
		err = s.validateResponse(question, responseValue)
		if err != nil {
			return fmt.Errorf("invalid response: %w", err)
		}
	}

	// Create or update response
	response := user.NewUserOnboardingResponse(userID, questionID, responseValue, skipped)
	
	err = s.userRepo.UpsertUserOnboardingResponse(ctx, response)
	if err != nil {
		return fmt.Errorf("failed to save response: %w", err)
	}

	// Audit log
	actionType := "onboarding.response_submitted"
	if skipped {
		actionType = "onboarding.question_skipped"
	}
	
	metadata := fmt.Sprintf(`{"question_id": "%s", "skipped": %t}`, questionID.String(), skipped)
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, actionType, "onboarding", questionID.String(), metadata, "", ""))

	return nil
}

// SubmitMultipleResponses submits multiple responses at once
func (s *onboardingService) SubmitMultipleResponses(ctx context.Context, userID ulid.ULID, responses []*user.SubmitOnboardingResponseRequest) error {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Process each response
	for _, resp := range responses {
		err = s.SubmitResponse(ctx, userID, resp.QuestionID, resp.ResponseValue, resp.Skipped)
		if err != nil {
			return fmt.Errorf("failed to submit response for question %s: %w", resp.QuestionID.String(), err)
		}
	}

	return nil
}

// GetOnboardingStatus retrieves the user's current onboarding progress
func (s *onboardingService) GetOnboardingStatus(ctx context.Context, userID ulid.ULID) (*user.OnboardingProgressStatus, error) {
	// Get total questions
	totalQuestions, err := s.userRepo.GetActiveOnboardingQuestionCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get question count: %w", err)
	}

	// Get user responses
	responses, err := s.userRepo.GetUserOnboardingResponses(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user responses: %w", err)
	}

	// Calculate status
	completedQuestions := 0
	skippedQuestions := 0
	currentStep := 1

	for _, response := range responses {
		if response.Skipped {
			skippedQuestions++
		} else {
			completedQuestions++
		}
	}

	answeredQuestions := len(responses)
	remainingQuestions := totalQuestions - answeredQuestions
	isComplete := answeredQuestions >= totalQuestions

	// Calculate current step (next unanswered question step)
	if !isComplete {
		nextQuestion, err := s.userRepo.GetNextUnansweredQuestion(ctx, userID)
		if err == nil && nextQuestion != nil {
			currentStep = nextQuestion.Step
		}
	}

	return &user.OnboardingProgressStatus{
		TotalQuestions:      totalQuestions,
		CompletedQuestions:  completedQuestions,
		SkippedQuestions:    skippedQuestions,
		RemainingQuestions:  remainingQuestions,
		IsComplete:          isComplete,
		CurrentStep:         currentStep,
	}, nil
}

// CompleteOnboarding marks the user's onboarding as completed
func (s *onboardingService) CompleteOnboarding(ctx context.Context, userID ulid.ULID) error {
	err := s.userRepo.CompleteOnboarding(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to complete onboarding: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.onboarding_completed", "onboarding", userID.String(), "", "", ""))

	return nil
}

// validateResponse validates a response against the question requirements
func (s *onboardingService) validateResponse(question *user.OnboardingQuestion, responseValue interface{}) error {
	if question.IsRequired && responseValue == nil {
		return fmt.Errorf("response is required for this question")
	}

	if responseValue == nil {
		return nil // Optional question with no response is valid
	}

	switch question.QuestionType {
	case "single_choice", "multiple_choice":
		// Validate against available options
		options, err := question.GetOptionsAsStrings()
		if err != nil {
			return fmt.Errorf("failed to parse question options: %w", err)
		}

		if len(options) == 0 {
			return fmt.Errorf("question has no valid options")
		}

		if question.QuestionType == "single_choice" {
			// Single choice should be a string
			responseStr, ok := responseValue.(string)
			if !ok {
				return fmt.Errorf("single choice response must be a string")
			}
			
			// Check if response is in options
			for _, option := range options {
				if option == responseStr {
					return nil
				}
			}
			return fmt.Errorf("response is not a valid option")
		} else {
			// Multiple choice should be an array
			responseArray, ok := responseValue.([]interface{})
			if !ok {
				return fmt.Errorf("multiple choice response must be an array")
			}
			
			// Check if all responses are valid options
			for _, resp := range responseArray {
				respStr, ok := resp.(string)
				if !ok {
					return fmt.Errorf("all multiple choice responses must be strings")
				}
				
				found := false
				for _, option := range options {
					if option == respStr {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("response '%s' is not a valid option", respStr)
				}
			}
		}

	case "text":
		// Text response should be a string
		_, ok := responseValue.(string)
		if !ok {
			return fmt.Errorf("text response must be a string")
		}

	case "skip_optional":
		// Skip optional questions can have any response or be skipped
		break
	}

	return nil
}