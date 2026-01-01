package evaluation

import (
	"time"

	"brokle/pkg/ulid"
)

// ExecutionStatus represents the current state of a rule execution.
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
)

// TriggerType defines how the evaluation was initiated.
type TriggerType string

const (
	TriggerTypeAutomatic TriggerType = "automatic"
	TriggerTypeManual    TriggerType = "manual"
)

// RuleExecution tracks the execution history of an evaluation rule.
// Inspired by Langfuse's JobExecution and Opik's automation rule logs.
type RuleExecution struct {
	ID           ulid.ULID       `json:"id" gorm:"type:char(26);primaryKey"`
	RuleID       ulid.ULID       `json:"rule_id" gorm:"type:char(26);not null;index"`
	ProjectID    ulid.ULID       `json:"project_id" gorm:"type:char(26);not null;index"`
	Status       ExecutionStatus `json:"status" gorm:"type:varchar(20);not null"`
	TriggerType  TriggerType     `json:"trigger_type" gorm:"type:varchar(20);not null;default:'automatic'"`
	SpansMatched int             `json:"spans_matched" gorm:"not null;default:0"`
	SpansScored  int             `json:"spans_scored" gorm:"not null;default:0"`
	ErrorsCount  int             `json:"errors_count" gorm:"not null;default:0"`
	ErrorMessage *string         `json:"error_message,omitempty" gorm:"type:text"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	DurationMs   *int            `json:"duration_ms,omitempty"`
	Metadata     map[string]any  `json:"metadata" gorm:"type:jsonb;serializer:json;default:'{}'"`
	CreatedAt    time.Time       `json:"created_at" gorm:"not null;autoCreateTime"`
}

func (RuleExecution) TableName() string {
	return "evaluation_rule_executions"
}

// NewRuleExecution creates a new rule execution record.
func NewRuleExecution(ruleID, projectID ulid.ULID, triggerType TriggerType) *RuleExecution {
	now := time.Now()
	return &RuleExecution{
		ID:           ulid.New(),
		RuleID:       ruleID,
		ProjectID:    projectID,
		Status:       ExecutionStatusPending,
		TriggerType:  triggerType,
		SpansMatched: 0,
		SpansScored:  0,
		ErrorsCount:  0,
		Metadata:     make(map[string]any),
		CreatedAt:    now,
	}
}

// Start marks the execution as running.
func (e *RuleExecution) Start() {
	now := time.Now()
	e.Status = ExecutionStatusRunning
	e.StartedAt = &now
}

// Complete marks the execution as successfully completed with counts.
func (e *RuleExecution) Complete(spansMatched, spansScored, errorsCount int) {
	now := time.Now()
	e.Status = ExecutionStatusCompleted
	e.SpansMatched = spansMatched
	e.SpansScored = spansScored
	e.ErrorsCount = errorsCount
	e.CompletedAt = &now

	if e.StartedAt != nil {
		durationMs := int(now.Sub(*e.StartedAt).Milliseconds())
		e.DurationMs = &durationMs
	}
}

// Fail marks the execution as failed with an error message.
func (e *RuleExecution) Fail(errorMessage string) {
	now := time.Now()
	e.Status = ExecutionStatusFailed
	e.ErrorMessage = &errorMessage
	e.CompletedAt = &now

	if e.StartedAt != nil {
		durationMs := int(now.Sub(*e.StartedAt).Milliseconds())
		e.DurationMs = &durationMs
	}
}

// Cancel marks the execution as cancelled.
func (e *RuleExecution) Cancel() {
	now := time.Now()
	e.Status = ExecutionStatusCancelled
	e.CompletedAt = &now

	if e.StartedAt != nil {
		durationMs := int(now.Sub(*e.StartedAt).Milliseconds())
		e.DurationMs = &durationMs
	}
}

// SetMetadata adds contextual metadata to the execution.
func (e *RuleExecution) SetMetadata(key string, value any) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]any)
	}
	e.Metadata[key] = value
}

// IsTerminal returns true if the execution is in a final state.
func (e *RuleExecution) IsTerminal() bool {
	switch e.Status {
	case ExecutionStatusCompleted, ExecutionStatusFailed, ExecutionStatusCancelled:
		return true
	default:
		return false
	}
}

// Response types

type RuleExecutionResponse struct {
	ID           string          `json:"id"`
	RuleID       string          `json:"rule_id"`
	ProjectID    string          `json:"project_id"`
	Status       ExecutionStatus `json:"status"`
	TriggerType  TriggerType     `json:"trigger_type"`
	SpansMatched int             `json:"spans_matched"`
	SpansScored  int             `json:"spans_scored"`
	ErrorsCount  int             `json:"errors_count"`
	ErrorMessage *string         `json:"error_message,omitempty"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	DurationMs   *int            `json:"duration_ms,omitempty"`
	Metadata     map[string]any  `json:"metadata,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

func (e *RuleExecution) ToResponse() *RuleExecutionResponse {
	metadata := e.Metadata
	if metadata == nil {
		metadata = make(map[string]any)
	}

	return &RuleExecutionResponse{
		ID:           e.ID.String(),
		RuleID:       e.RuleID.String(),
		ProjectID:    e.ProjectID.String(),
		Status:       e.Status,
		TriggerType:  e.TriggerType,
		SpansMatched: e.SpansMatched,
		SpansScored:  e.SpansScored,
		ErrorsCount:  e.ErrorsCount,
		ErrorMessage: e.ErrorMessage,
		StartedAt:    e.StartedAt,
		CompletedAt:  e.CompletedAt,
		DurationMs:   e.DurationMs,
		Metadata:     metadata,
		CreatedAt:    e.CreatedAt,
	}
}

type RuleExecutionListResponse struct {
	Executions []*RuleExecutionResponse `json:"executions"`
	Total      int64                    `json:"total"`
	Page       int                      `json:"page"`
	Limit      int                      `json:"limit"`
}

// Filter for listing executions
type ExecutionFilter struct {
	Status      *ExecutionStatus
	TriggerType *TriggerType
}

// TriggerOptions for manual evaluation trigger
type TriggerOptions struct {
	TimeRangeStart *time.Time `json:"time_range_start,omitempty"` // Optional: start of time range to evaluate
	TimeRangeEnd   *time.Time `json:"time_range_end,omitempty"`   // Optional: end of time range to evaluate
	SpanIDs        []string   `json:"span_ids,omitempty"`         // Optional: specific spans to evaluate
	SampleLimit    int        `json:"sample_limit,omitempty"`     // Optional: max spans to process (default: 1000)
}

// TriggerResponse returned when triggering manual evaluation
type TriggerResponse struct {
	ExecutionID string `json:"execution_id"`
	SpansQueued int    `json:"spans_queued"`
	Message     string `json:"message"`
}
