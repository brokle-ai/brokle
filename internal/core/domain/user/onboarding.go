// Package user provides onboarding functionality for user domain.
//
// The onboarding domain handles dynamic onboarding questions and user responses.
// It provides flexible question types and progress tracking capabilities.
package user

import (
	"encoding/json"
	"time"

	"brokle/pkg/ulid"
)

// OnboardingQuestion represents a dynamic onboarding question
type OnboardingQuestion struct {
	ID           ulid.ULID `json:"id" gorm:"type:char(26);primaryKey"`
	Step         int       `json:"step" gorm:"not null"`
	QuestionType string    `json:"question_type" gorm:"size:50;not null"` // single_choice, multiple_choice, text, skip_optional
	Title        string    `json:"title" gorm:"type:text;not null"`
	Description  string    `json:"description,omitempty" gorm:"type:text"`
	IsRequired   bool      `json:"is_required" gorm:"default:true"`
	Options      json.RawMessage `json:"options,omitempty" gorm:"type:jsonb"` // For choice questions
	DisplayOrder int       `json:"display_order" gorm:"default:0"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserOnboardingResponse represents a user's response to an onboarding question
type UserOnboardingResponse struct {
	ID            ulid.ULID `json:"id" gorm:"type:char(26);primaryKey"`
	UserID        ulid.ULID `json:"user_id" gorm:"type:char(26);not null"`
	QuestionID    ulid.ULID `json:"question_id" gorm:"type:char(26);not null"`
	ResponseValue json.RawMessage `json:"response_value" gorm:"type:jsonb;not null"` // Flexible storage
	Skipped       bool      `json:"skipped" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at"`

	// Relations
	User     User               `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Question OnboardingQuestion `json:"question,omitempty" gorm:"foreignKey:QuestionID"`
}

// OnboardingProgressStatus represents the user's onboarding progress
type OnboardingProgressStatus struct {
	TotalQuestions      int  `json:"total_questions"`
	CompletedQuestions  int  `json:"completed_questions"`
	SkippedQuestions    int  `json:"skipped_questions"`
	RemainingQuestions  int  `json:"remaining_questions"`
	IsComplete          bool `json:"is_complete"`
	CurrentStep         int  `json:"current_step"`
}

// UserOnboardingResponseData represents processed response data for service layer
type UserOnboardingResponseData struct {
	ID            ulid.ULID   `json:"id"`
	UserID        ulid.ULID   `json:"user_id"`
	QuestionID    ulid.ULID   `json:"question_id"`
	ResponseValue interface{} `json:"response_value"`
	Skipped       bool        `json:"skipped"`
}

// CreateOnboardingQuestionRequest represents the data needed to create a new question
type CreateOnboardingQuestionRequest struct {
	Step         int      `json:"step" validate:"required,min=1"`
	QuestionType string   `json:"question_type" validate:"required,oneof=single_choice multiple_choice text skip_optional"`
	Title        string   `json:"title" validate:"required,min=1"`
	Description  string   `json:"description,omitempty"`
	IsRequired   bool     `json:"is_required"`
	Options      []string `json:"options,omitempty"`
	DisplayOrder int      `json:"display_order"`
}

// SubmitOnboardingResponseRequest represents a response submission
type SubmitOnboardingResponseRequest struct {
	QuestionID    ulid.ULID   `json:"question_id" validate:"required"`
	ResponseValue interface{} `json:"response_value,omitempty"`
	Skipped       bool        `json:"skipped"`
}

// Table name methods for GORM
func (OnboardingQuestion) TableName() string      { return "onboarding_questions" }
func (UserOnboardingResponse) TableName() string { return "user_onboarding_responses" }

// OnboardingQuestion methods

// GetOptionsAsStrings converts the JSONB options to string array
func (q *OnboardingQuestion) GetOptionsAsStrings() ([]string, error) {
	if q.Options == nil {
		return nil, nil
	}
	
	var options []string
	err := json.Unmarshal(q.Options, &options)
	if err != nil {
		return nil, err
	}
	return options, nil
}

// SetOptionsFromStrings sets the options from a string array
func (q *OnboardingQuestion) SetOptionsFromStrings(options []string) error {
	if options == nil {
		q.Options = nil
		return nil
	}
	
	data, err := json.Marshal(options)
	if err != nil {
		return err
	}
	q.Options = json.RawMessage(data)
	return nil
}

// IsChoiceQuestion returns true if the question type is single_choice or multiple_choice
func (q *OnboardingQuestion) IsChoiceQuestion() bool {
	return q.QuestionType == "single_choice" || q.QuestionType == "multiple_choice"
}

// IsTextQuestion returns true if the question type is text
func (q *OnboardingQuestion) IsTextQuestion() bool {
	return q.QuestionType == "text"
}

// IsSkippable returns true if the question can be skipped
func (q *OnboardingQuestion) IsSkippable() bool {
	return !q.IsRequired || q.QuestionType == "skip_optional"
}

// UserOnboardingResponse methods

// SetResponseString sets a simple string response
func (r *UserOnboardingResponse) SetResponseString(value string) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	r.ResponseValue = json.RawMessage(data)
	return nil
}

// GetResponseString gets the response as a string
func (r *UserOnboardingResponse) GetResponseString() (string, error) {
	if r.ResponseValue == nil {
		return "", nil
	}
	
	var value string
	err := json.Unmarshal(r.ResponseValue, &value)
	if err != nil {
		return "", err
	}
	return value, nil
}

// SetResponseStrings sets a string array response (for multiple choice)
func (r *UserOnboardingResponse) SetResponseStrings(values []string) error {
	data, err := json.Marshal(values)
	if err != nil {
		return err
	}
	r.ResponseValue = json.RawMessage(data)
	return nil
}

// GetResponseStrings gets the response as a string array
func (r *UserOnboardingResponse) GetResponseStrings() ([]string, error) {
	if r.ResponseValue == nil {
		return nil, nil
	}
	
	var values []string
	err := json.Unmarshal(r.ResponseValue, &values)
	if err != nil {
		return nil, err
	}
	return values, nil
}

// SetResponseObject sets a complex object response
func (r *UserOnboardingResponse) SetResponseObject(value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	r.ResponseValue = json.RawMessage(data)
	return nil
}

// GetResponseObject gets the response as an interface{}
func (r *UserOnboardingResponse) GetResponseObject() (interface{}, error) {
	if r.ResponseValue == nil {
		return nil, nil
	}
	
	var value interface{}
	err := json.Unmarshal(r.ResponseValue, &value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// NewOnboardingQuestion creates a new onboarding question with default values
func NewOnboardingQuestion(req *CreateOnboardingQuestionRequest) *OnboardingQuestion {
	question := &OnboardingQuestion{
		ID:           ulid.New(),
		Step:         req.Step,
		QuestionType: req.QuestionType,
		Title:        req.Title,
		Description:  req.Description,
		IsRequired:   req.IsRequired,
		DisplayOrder: req.DisplayOrder,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	if len(req.Options) > 0 {
		_ = question.SetOptionsFromStrings(req.Options)
	}
	
	return question
}

// NewUserOnboardingResponse creates a new user onboarding response
func NewUserOnboardingResponse(userID, questionID ulid.ULID, responseValue interface{}, skipped bool) *UserOnboardingResponse {
	response := &UserOnboardingResponse{
		ID:         ulid.New(),
		UserID:     userID,
		QuestionID: questionID,
		Skipped:    skipped,
		CreatedAt:  time.Now(),
	}
	
	if responseValue != nil {
		_ = response.SetResponseObject(responseValue)
	} else {
		// For skipped questions or null responses, store empty JSON object
		response.ResponseValue = json.RawMessage("{}")
	}
	
	return response
}