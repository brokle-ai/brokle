package evaluation

import (
	"time"

	"brokle/pkg/ulid"

	"github.com/lib/pq"
)

// RuleStatus represents the current state of an evaluation rule.
type RuleStatus string

const (
	RuleStatusActive   RuleStatus = "active"
	RuleStatusInactive RuleStatus = "inactive"
	RuleStatusPaused   RuleStatus = "paused"
)

// RuleTrigger defines when an evaluation rule is executed.
type RuleTrigger string

const (
	RuleTriggerOnSpanComplete RuleTrigger = "on_span_complete"
)

// TargetScope defines the scope of evaluation (span or trace level).
type TargetScope string

const (
	TargetScopeSpan  TargetScope = "span"
	TargetScopeTrace TargetScope = "trace"
)

// ScorerType defines the type of scorer used for evaluation.
type ScorerType string

const (
	ScorerTypeLLM     ScorerType = "llm"
	ScorerTypeBuiltin ScorerType = "builtin"
	ScorerTypeRegex   ScorerType = "regex"
)

// FilterClause represents a single filter condition for matching spans.
type FilterClause struct {
	Field    string      `json:"field"`    // e.g., "input", "output", "metadata.key", "span_kind"
	Operator string      `json:"operator"` // equals, not_equals, contains, gt, lt, is_empty
	Value    interface{} `json:"value"`
}

// VariableMap defines how to extract a variable from span data.
type VariableMap struct {
	VariableName string `json:"variable_name"` // Template variable: {input}, {output}
	Source       string `json:"source"`        // span_input, span_output, span_metadata, trace_input
	JSONPath     string `json:"json_path"`     // Optional: "messages[0].content", "data.result"
}

// EvaluationRule defines an automated evaluation rule for scoring spans.
type EvaluationRule struct {
	ID              ulid.ULID      `json:"id" gorm:"type:char(26);primaryKey"`
	ProjectID       ulid.ULID      `json:"project_id" gorm:"type:char(26);not null;index"`
	Name            string         `json:"name" gorm:"type:varchar(100);not null"`
	Description     *string        `json:"description,omitempty" gorm:"type:text"`
	Status          RuleStatus     `json:"status" gorm:"type:varchar(20);not null;default:'inactive'"`
	TriggerType     RuleTrigger    `json:"trigger_type" gorm:"type:varchar(30);not null;default:'on_span_complete'"`
	TargetScope     TargetScope    `json:"target_scope" gorm:"type:varchar(20);not null;default:'span'"`
	Filter          []FilterClause `json:"filter" gorm:"type:jsonb;serializer:json;not null;default:'[]'"`
	SpanNames       pq.StringArray `json:"span_names" gorm:"type:text[];default:'{}'"`
	SamplingRate    float64        `json:"sampling_rate" gorm:"type:decimal(5,4);not null;default:1.0"`
	ScorerType      ScorerType     `json:"scorer_type" gorm:"type:varchar(20);not null"`
	ScorerConfig    map[string]any `json:"scorer_config" gorm:"type:jsonb;serializer:json;not null"`
	VariableMapping []VariableMap  `json:"variable_mapping" gorm:"type:jsonb;serializer:json;not null;default:'[]'"`
	CreatedBy       *string        `json:"created_by,omitempty" gorm:"type:char(26)"`
	CreatedAt       time.Time      `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"not null;autoUpdateTime"`
}

func (EvaluationRule) TableName() string {
	return "evaluation_rules"
}

func NewEvaluationRule(projectID ulid.ULID, name string, scorerType ScorerType, scorerConfig map[string]any) *EvaluationRule {
	now := time.Now()
	return &EvaluationRule{
		ID:              ulid.New(),
		ProjectID:       projectID,
		Name:            name,
		Status:          RuleStatusInactive,
		TriggerType:     RuleTriggerOnSpanComplete,
		TargetScope:     TargetScopeSpan,
		Filter:          []FilterClause{},
		SpanNames:       []string{},
		SamplingRate:    1.0,
		ScorerType:      scorerType,
		ScorerConfig:    scorerConfig,
		VariableMapping: []VariableMap{},
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (r *EvaluationRule) Validate() []ValidationError {
	var errors []ValidationError

	if r.Name == "" {
		errors = append(errors, ValidationError{Field: "name", Message: "name is required"})
	}
	if len(r.Name) > 100 {
		errors = append(errors, ValidationError{Field: "name", Message: "name must be 100 characters or less"})
	}

	switch r.Status {
	case RuleStatusActive, RuleStatusInactive, RuleStatusPaused:
	default:
		errors = append(errors, ValidationError{Field: "status", Message: "invalid status, must be active, inactive, or paused"})
	}

	switch r.TriggerType {
	case RuleTriggerOnSpanComplete:
	default:
		errors = append(errors, ValidationError{Field: "trigger_type", Message: "invalid trigger type"})
	}

	switch r.TargetScope {
	case TargetScopeSpan, TargetScopeTrace:
	default:
		errors = append(errors, ValidationError{Field: "target_scope", Message: "invalid target scope, must be span or trace"})
	}

	if r.SamplingRate < 0.0 || r.SamplingRate > 1.0 {
		errors = append(errors, ValidationError{Field: "sampling_rate", Message: "sampling rate must be between 0 and 1"})
	}

	switch r.ScorerType {
	case ScorerTypeLLM, ScorerTypeBuiltin, ScorerTypeRegex:
	default:
		errors = append(errors, ValidationError{Field: "scorer_type", Message: "invalid scorer type, must be llm, builtin, or regex"})
	}

	if r.ScorerConfig == nil {
		errors = append(errors, ValidationError{Field: "scorer_config", Message: "scorer_config is required"})
	}

	return errors
}

// Request/Response types

type CreateEvaluationRuleRequest struct {
	Name            string         `json:"name" binding:"required,min=1,max=100"`
	Description     *string        `json:"description,omitempty"`
	Status          *RuleStatus    `json:"status,omitempty"`
	TriggerType     *RuleTrigger   `json:"trigger_type,omitempty"`
	TargetScope     *TargetScope   `json:"target_scope,omitempty"`
	Filter          []FilterClause `json:"filter,omitempty"`
	SpanNames       []string       `json:"span_names,omitempty"`
	SamplingRate    *float64       `json:"sampling_rate,omitempty"`
	ScorerType      ScorerType     `json:"scorer_type" binding:"required,oneof=llm builtin regex"`
	ScorerConfig    map[string]any `json:"scorer_config" binding:"required"`
	VariableMapping []VariableMap  `json:"variable_mapping,omitempty"`
}

type UpdateEvaluationRuleRequest struct {
	Name            *string        `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description     *string        `json:"description,omitempty"`
	Status          *RuleStatus    `json:"status,omitempty" binding:"omitempty,oneof=active inactive paused"`
	TriggerType     *RuleTrigger   `json:"trigger_type,omitempty"`
	TargetScope     *TargetScope   `json:"target_scope,omitempty" binding:"omitempty,oneof=span trace"`
	Filter          []FilterClause `json:"filter,omitempty"`
	SpanNames       []string       `json:"span_names,omitempty"`
	SamplingRate    *float64       `json:"sampling_rate,omitempty"`
	ScorerType      *ScorerType    `json:"scorer_type,omitempty" binding:"omitempty,oneof=llm builtin regex"`
	ScorerConfig    map[string]any `json:"scorer_config,omitempty"`
	VariableMapping []VariableMap  `json:"variable_mapping,omitempty"`
}

type EvaluationRuleResponse struct {
	ID              string         `json:"id"`
	ProjectID       string         `json:"project_id"`
	Name            string         `json:"name"`
	Description     *string        `json:"description,omitempty"`
	Status          RuleStatus     `json:"status"`
	TriggerType     RuleTrigger    `json:"trigger_type"`
	TargetScope     TargetScope    `json:"target_scope"`
	Filter          []FilterClause `json:"filter"`
	SpanNames       []string       `json:"span_names"`
	SamplingRate    float64        `json:"sampling_rate"`
	ScorerType      ScorerType     `json:"scorer_type"`
	ScorerConfig    map[string]any `json:"scorer_config"`
	VariableMapping []VariableMap  `json:"variable_mapping"`
	CreatedBy       *string        `json:"created_by,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

func (r *EvaluationRule) ToResponse() *EvaluationRuleResponse {
	var createdBy *string
	if r.CreatedBy != nil {
		createdBy = r.CreatedBy
	}

	spanNames := r.SpanNames
	if spanNames == nil {
		spanNames = []string{}
	}

	filter := r.Filter
	if filter == nil {
		filter = []FilterClause{}
	}

	variableMapping := r.VariableMapping
	if variableMapping == nil {
		variableMapping = []VariableMap{}
	}

	return &EvaluationRuleResponse{
		ID:              r.ID.String(),
		ProjectID:       r.ProjectID.String(),
		Name:            r.Name,
		Description:     r.Description,
		Status:          r.Status,
		TriggerType:     r.TriggerType,
		TargetScope:     r.TargetScope,
		Filter:          filter,
		SpanNames:       spanNames,
		SamplingRate:    r.SamplingRate,
		ScorerType:      r.ScorerType,
		ScorerConfig:    r.ScorerConfig,
		VariableMapping: variableMapping,
		CreatedBy:       createdBy,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

// RuleFilter for listing rules with pagination.
type RuleFilter struct {
	Status     *RuleStatus
	ScorerType *ScorerType
	Search     *string
}

// LLM Scorer Configuration Types

type LLMMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"` // Template with {input}, {output}, {expected}
}

type OutputField struct {
	Name        string   `json:"name"`                  // Score name: "relevance", "coherence"
	Type        string   `json:"type"`                  // numeric, categorical, boolean
	Description string   `json:"description,omitempty"` // What to evaluate
	MinValue    *float64 `json:"min_value,omitempty"`
	MaxValue    *float64 `json:"max_value,omitempty"`
	Categories  []string `json:"categories,omitempty"` // For categorical: ["good", "bad"]
}

type LLMScorerConfig struct {
	CredentialID   string        `json:"credential_id"`   // Project's AI credential
	Model          string        `json:"model"`           // gpt-4o, claude-3-5-sonnet
	Messages       []LLMMessage  `json:"messages"`        // System + User messages
	Temperature    float64       `json:"temperature"`     // 0.0-1.0
	ResponseFormat string        `json:"response_format"` // json, text
	OutputSchema   []OutputField `json:"output_schema"`   // Expected output structure
}

// Builtin Scorer Configuration Types

type BuiltinScorerConfig struct {
	ScorerName string         `json:"scorer_name"` // contains, json_valid, length_check
	Config     map[string]any `json:"config"`      // Scorer-specific configuration
}

// Regex Scorer Configuration Types

type RegexScorerConfig struct {
	Pattern      string  `json:"pattern"`                 // Regex pattern
	ScoreName    string  `json:"score_name"`              // Name for the generated score
	MatchScore   float64 `json:"match_score,omitempty"`   // Score when pattern matches (default 1.0)
	NoMatchScore float64 `json:"no_match_score,omitempty"` // Score when pattern doesn't match (default 0.0)
	CaptureGroup *int    `json:"capture_group,omitempty"` // Capture group to use for value extraction
}
