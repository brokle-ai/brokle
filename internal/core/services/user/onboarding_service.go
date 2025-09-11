package user

import (
	"context"
	"encoding/json"

	userDomain "brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
	appErrors "brokle/pkg/errors"
)

// onboardingService implements the user.OnboardingService interface for dynamic onboarding
type onboardingService struct {
	userRepo userDomain.Repository
}

// NewOnboardingService creates a new dynamic onboarding service instance
func NewOnboardingService(
	userRepo userDomain.Repository,
) userDomain.OnboardingService {
	return &onboardingService{
		userRepo: userRepo,
	}
}

// GetActiveQuestions retrieves all active onboarding questions
func (s *onboardingService) GetActiveQuestions(ctx context.Context) ([]*userDomain.OnboardingQuestion, error) {
	return s.userRepo.GetActiveOnboardingQuestions(ctx)
}

// GetQuestionByID retrieves a specific onboarding question by ID
func (s *onboardingService) GetQuestionByID(ctx context.Context, id ulid.ULID) (*userDomain.OnboardingQuestion, error) {
	return s.userRepo.GetOnboardingQuestionByID(ctx, id)
}

// CreateQuestion creates a new onboarding question
func (s *onboardingService) CreateQuestion(ctx context.Context, req *userDomain.CreateOnboardingQuestionRequest) (*userDomain.OnboardingQuestion, error) {
	// Validate request
	if req.Title == "" {
		return nil, appErrors.NewValidationError("title", "Question title is required")
	}

	if req.QuestionType != "single_choice" && req.QuestionType != "multiple_choice" && 
	   req.QuestionType != "text" && req.QuestionType != "skip_optional" {
		return nil, appErrors.NewValidationError("question_type", "Invalid question type")
	}

	if req.Step <= 0 {
		return nil, appErrors.NewValidationError("step", "Step must be greater than 0")
	}

	// Create question
	question := userDomain.NewOnboardingQuestion(req)
	
	err := s.userRepo.CreateOnboardingQuestion(ctx, question)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to create onboarding question", err)
	}

	return question, nil
}

// GetUserResponses retrieves all user responses with parsed values
func (s *onboardingService) GetUserResponses(ctx context.Context, userID ulid.ULID) ([]*userDomain.UserOnboardingResponseData, error) {
	responses, err := s.userRepo.GetUserOnboardingResponses(ctx, userID)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to get user responses", err)
	}

	var result []*userDomain.UserOnboardingResponseData
	for _, response := range responses {
		// Parse the JSON response value
		var responseValue interface{}
		if response.ResponseValue != nil {
			err := json.Unmarshal(response.ResponseValue, &responseValue)
			if err != nil {
				responseValue = string(response.ResponseValue) // Fallback to raw string
			}
		}

		result = append(result, &userDomain.UserOnboardingResponseData{
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
		return appErrors.NewNotFoundError("Question not found")
	}

	// Verify user exists
	_, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return appErrors.NewNotFoundError("User not found")
	}

	// Validate response if not skipped
	if !skipped {
		err = s.validateResponse(question, responseValue)
		if err != nil {
			return err // validateResponse already returns AppError
		}
	}

	// Create or update response
	response := userDomain.NewUserOnboardingResponse(userID, questionID, responseValue, skipped)
	
	err = s.userRepo.UpsertUserOnboardingResponse(ctx, response)
	if err != nil {
		return appErrors.NewInternalError("Failed to save response", err)
	}

	return nil
}

// SubmitMultipleResponses submits multiple responses at once
func (s *onboardingService) SubmitMultipleResponses(ctx context.Context, userID ulid.ULID, responses []*userDomain.SubmitOnboardingResponseRequest) error {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return appErrors.NewNotFoundError("User not found")
	}

	// Process each response
	for _, resp := range responses {
		err = s.SubmitResponse(ctx, userID, resp.QuestionID, resp.ResponseValue, resp.Skipped)
		if err != nil {
			return appErrors.NewInternalError("Failed to submit response for question "+resp.QuestionID.String(), err)
		}
	}

	return nil
}

// GetOnboardingStatus retrieves the user's current onboarding progress
func (s *onboardingService) GetOnboardingStatus(ctx context.Context, userID ulid.ULID) (*userDomain.OnboardingProgressStatus, error) {
	// Get total questions
	totalQuestions, err := s.userRepo.GetActiveOnboardingQuestionCount(ctx)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to get question count", err)
	}

	// Get user responses
	responses, err := s.userRepo.GetUserOnboardingResponses(ctx, userID)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to get user responses", err)
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

	return &userDomain.OnboardingProgressStatus{
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
		return appErrors.NewInternalError("Failed to complete onboarding", err)
	}

	return nil
}

// validateResponse validates a response against the question requirements
func (s *onboardingService) validateResponse(question *userDomain.OnboardingQuestion, responseValue interface{}) error {
	if question.IsRequired && responseValue == nil {
		return appErrors.NewValidationError("response", "Response is required for this question")
	}

	if responseValue == nil {
		return nil // Optional question with no response is valid
	}

	switch question.QuestionType {
	case "single_choice", "multiple_choice":
		// Validate against available options
		options, err := question.GetOptionsAsStrings()
		if err != nil {
			return appErrors.NewInternalError("Failed to parse question options", err)
		}

		if len(options) == 0 {
			return appErrors.NewValidationError("options", "Question has no valid options")
		}

		if question.QuestionType == "single_choice" {
			// Single choice should be a string
			responseStr, ok := responseValue.(string)
			if !ok {
				return appErrors.NewValidationError("response", "Single choice response must be a string")
			}
			
			// Check if response is in options
			for _, option := range options {
				if option == responseStr {
					return nil
				}
			}
			return appErrors.NewValidationError("response", "Response is not a valid option")
		} else {
			// Multiple choice should be an array
			responseArray, ok := responseValue.([]interface{})
			if !ok {
				return appErrors.NewValidationError("response", "Multiple choice response must be an array")
			}
			
			// Check if all responses are valid options
			for _, resp := range responseArray {
				respStr, ok := resp.(string)
				if !ok {
					return appErrors.NewValidationError("response", "All multiple choice responses must be strings")
				}
				
				found := false
				for _, option := range options {
					if option == respStr {
						found = true
						break
					}
				}
				if !found {
					return appErrors.NewValidationError("response", "Response '"+respStr+"' is not a valid option")
				}
			}
		}

	case "text":
		// Text response should be a string
		_, ok := responseValue.(string)
		if !ok {
			return appErrors.NewValidationError("response", "Text response must be a string")
		}

	case "skip_optional":
		// Skip optional questions can have any response or be skipped
		break
	}

	return nil
}