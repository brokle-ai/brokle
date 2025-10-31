package user

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/core/domain/user"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

// OnboardingHandler handles onboarding-related HTTP requests
type OnboardingHandler struct {
	config            *config.Config
	logger            *logrus.Logger
	onboardingService user.OnboardingService
	userService       user.UserService
}

// NewOnboardingHandler creates a new onboarding handler
func NewOnboardingHandler(
	config *config.Config, 
	logger *logrus.Logger, 
	onboardingService user.OnboardingService,
	userService user.UserService,
) *OnboardingHandler {
	return &OnboardingHandler{
		config:            config,
		logger:            logger,
		onboardingService: onboardingService,
		userService:       userService,
	}
}

// Request/Response types for onboarding operations

// OnboardingQuestionsResponse represents a single onboarding question with user response
// @Description Complete onboarding question information including user's current response
type OnboardingQuestionsResponse struct {
	ID           string      `json:"id" example:"01K4FHGHT3XX9WFM293QPZ5G9V" description:"Question unique identifier"`
	Step         int         `json:"step" example:"1" description:"Question step number"`
	QuestionType string      `json:"question_type" example:"single_choice" description:"Question type (single_choice, multiple_choice, text, skip_optional)"`
	Title        string      `json:"title" example:"What is your primary role?" description:"Question title"`
	Description  string      `json:"description,omitempty" example:"This helps us customize your experience" description:"Question description"`
	IsRequired   bool        `json:"is_required" example:"true" description:"Whether response is required"`
	Options      []string    `json:"options,omitempty" example:"[\"Developer\", \"Manager\", \"Analyst\"]" description:"Available options for choice questions"`
	UserAnswer   *string     `json:"user_answer,omitempty" swaggertype:"string" example:"Developer" description:"User's current answer (string for single choice/text, JSON string for multiple choice)"`
	IsSkipped    bool        `json:"is_skipped" example:"false" description:"Whether user skipped this question"`
}

// SubmitResponseRequest represents a single response submission
// @Description Request body for submitting a response to an onboarding question
type SubmitResponseRequest struct {
	QuestionID    string      `json:"question_id" binding:"required" example:"01K4FHGHT3XX9WFM293QPZ5G9V" description:"Question ID to respond to"`
	ResponseValue interface{} `json:"response_value,omitempty" swaggertype:"string" example:"Developer" description:"Response value (string for single choice/text, JSON string for multiple choice)"`
	Skipped       bool        `json:"skipped" example:"false" description:"Whether to skip this question"`
}

// SubmitResponsesRequest represents multiple response submissions
// @Description Request body for submitting multiple onboarding responses at once
type SubmitResponsesRequest struct {
	Responses []SubmitResponseRequest `json:"responses" binding:"required" description:"Array of responses to submit"`
}

// OnboardingStatusResponse represents the user's onboarding progress
// @Description User's current onboarding progress and completion status
type OnboardingStatusResponse struct {
	TotalQuestions      int  `json:"total_questions" example:"5" description:"Total number of onboarding questions"`
	CompletedQuestions  int  `json:"completed_questions" example:"3" description:"Number of completed questions"`
	SkippedQuestions    int  `json:"skipped_questions" example:"1" description:"Number of skipped questions"`
	RemainingQuestions  int  `json:"remaining_questions" example:"1" description:"Number of remaining questions"`
	OnboardingCompleted bool `json:"onboarding_completed" example:"false" description:"Whether onboarding is completed (computed from onboarding_completed_at)"`
	CurrentStep         int  `json:"current_step" example:"2" description:"Current step number"`
}

// GetQuestions handles GET /api/v1/onboarding/questions
// @Summary Get onboarding questions
// @Description Retrieves all active onboarding questions with user's current responses
// @Tags Onboarding
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=[]OnboardingQuestionsResponse} "Questions retrieved successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/onboarding/questions [get]
func (h *OnboardingHandler) GetQuestions(c *gin.Context) {
	// Get user ID from middleware (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in request context")
		response.Unauthorized(c, "Authentication required")
		return
	}

	userID, ok := userIDInterface.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		response.InternalServerError(c, "Authentication error")
		return
	}

	// Get all active questions
	questions, err := h.onboardingService.GetActiveQuestions(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get onboarding questions")
		response.InternalServerError(c, "Failed to retrieve questions")
		return
	}

	// Get user responses
	userResponses, err := h.onboardingService.GetUserResponses(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user responses")
		response.InternalServerError(c, "Failed to retrieve user responses")
		return
	}

	// Create response map for quick lookup
	responseMap := make(map[ulid.ULID]*user.UserOnboardingResponseData)
	for _, resp := range userResponses {
		responseMap[resp.QuestionID] = resp
	}

	// Build response
	var questionsResponse []OnboardingQuestionsResponse
	for _, question := range questions {
		options, _ := question.GetOptionsAsStrings()
		
		questionResp := OnboardingQuestionsResponse{
			ID:           question.ID.String(),
			Step:         question.Step,
			QuestionType: question.QuestionType,
			Title:        question.Title,
			Description:  question.Description,
			IsRequired:   question.IsRequired,
			Options:      options,
		}

		// Add user response if exists
		if userResp, exists := responseMap[question.ID]; exists {
			// Convert response value to string for Swagger compatibility
			if userResp.ResponseValue != nil {
				if responseStr, ok := userResp.ResponseValue.(string); ok {
					questionResp.UserAnswer = &responseStr
				} else {
					// For non-string responses (arrays, objects), serialize to JSON
					responseBytes, _ := json.Marshal(userResp.ResponseValue)
					responseStr := string(responseBytes)
					questionResp.UserAnswer = &responseStr
				}
			}
			questionResp.IsSkipped = userResp.Skipped
		}

		questionsResponse = append(questionsResponse, questionResp)
	}

	h.logger.WithField("user_id", userID).WithField("question_count", len(questionsResponse)).Info("Retrieved onboarding questions")
	response.Success(c, questionsResponse)
}

// SubmitResponses handles POST /api/v1/onboarding/responses
// @Summary Submit onboarding responses
// @Description Submit answers to one or more onboarding questions
// @Tags Onboarding
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SubmitResponsesRequest true "Responses to submit"
// @Success 200 {object} response.APIResponse "Responses submitted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/onboarding/responses [post]
func (h *OnboardingHandler) SubmitResponses(c *gin.Context) {
	// Get user ID from middleware
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in request context")
		response.Unauthorized(c, "Authentication required")
		return
	}

	userID, ok := userIDInterface.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		response.InternalServerError(c, "Authentication error")
		return
	}

	var req SubmitResponsesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid response submission request")
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Convert to domain types
	var domainResponses []*user.SubmitOnboardingResponseRequest
	for _, respReq := range req.Responses {
		questionID, err := ulid.Parse(respReq.QuestionID)
		if err != nil {
			h.logger.WithError(err).WithField("question_id", respReq.QuestionID).Error("Invalid question ID format")
			response.BadRequest(c, "Invalid question ID format", err.Error())
			return
		}

		domainResponses = append(domainResponses, &user.SubmitOnboardingResponseRequest{
			QuestionID:    questionID,
			ResponseValue: respReq.ResponseValue,
			Skipped:       respReq.Skipped,
		})
	}

	err := h.onboardingService.SubmitMultipleResponses(c.Request.Context(), userID, domainResponses)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to submit onboarding responses")
		response.InternalServerError(c, "Failed to submit responses")
		return
	}

	h.logger.WithField("user_id", userID).WithField("response_count", len(req.Responses)).Info("Onboarding responses submitted successfully")
	response.Success(c, gin.H{"message": "Responses submitted successfully"})
}

// SkipQuestion handles POST /api/v1/onboarding/skip/{id}
// @Summary Skip onboarding question
// @Description Skip a specific onboarding question
// @Tags Onboarding
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question ID"
// @Success 200 {object} response.APIResponse "Question skipped successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid question ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/onboarding/skip/{id} [post]
func (h *OnboardingHandler) SkipQuestion(c *gin.Context) {
	// Get user ID from middleware
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in request context")
		response.Unauthorized(c, "Authentication required")
		return
	}

	userID, ok := userIDInterface.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		response.InternalServerError(c, "Authentication error")
		return
	}

	questionID, err := ulid.Parse(c.Param("id"))
	if err != nil {
		h.logger.WithError(err).WithField("question_id", c.Param("id")).Error("Invalid question ID format")
		response.BadRequest(c, "Invalid question ID format", err.Error())
		return
	}

	// Skip the question
	err = h.onboardingService.SubmitResponse(c.Request.Context(), userID, questionID, nil, true)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).WithField("question_id", questionID).Error("Failed to skip question")
		response.InternalServerError(c, "Failed to skip question")
		return
	}

	h.logger.WithField("user_id", userID).WithField("question_id", questionID).Info("Onboarding question skipped successfully")
	response.Success(c, gin.H{"message": "Question skipped successfully"})
}

// CompleteOnboarding handles POST /api/v1/onboarding/complete
// @Summary Complete onboarding
// @Description Mark the user's onboarding as completed
// @Tags Onboarding
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.APIResponse "Onboarding completed successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/onboarding/complete [post]
func (h *OnboardingHandler) CompleteOnboarding(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in request context")
		response.Unauthorized(c, "Authentication required")
		return
	}

	userID, ok := userIDInterface.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		response.InternalServerError(c, "Authentication error")
		return
	}

	err := h.onboardingService.CompleteOnboarding(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to complete onboarding")
		response.InternalServerError(c, "Failed to complete onboarding")
		return
	}

	h.logger.WithField("user_id", userID).Info("Onboarding completed successfully")
	response.Success(c, gin.H{"message": "Onboarding completed successfully"})
}

// GetStatus handles GET /api/v1/onboarding/status
// @Summary Get onboarding status
// @Description Get the user's current onboarding progress
// @Tags Onboarding
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=OnboardingStatusResponse} "Status retrieved successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/onboarding/status [get]
func (h *OnboardingHandler) GetStatus(c *gin.Context) {
	// Get user ID from middleware
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in request context")
		response.Unauthorized(c, "Authentication required")
		return
	}

	userID, ok := userIDInterface.(ulid.ULID)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		response.InternalServerError(c, "Authentication error")
		return
	}

	// Get onboarding status
	status, err := h.onboardingService.GetOnboardingStatus(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get onboarding status")
		response.InternalServerError(c, "Failed to retrieve onboarding status")
		return
	}

	statusResponse := OnboardingStatusResponse{
		TotalQuestions:      status.TotalQuestions,
		CompletedQuestions:  status.CompletedQuestions,
		SkippedQuestions:    status.SkippedQuestions,
		RemainingQuestions:  status.RemainingQuestions,
		OnboardingCompleted: status.IsComplete,
		CurrentStep:         status.CurrentStep,
	}

	h.logger.WithField("user_id", userID).Info("Retrieved onboarding status")
	response.Success(c, statusResponse)
}