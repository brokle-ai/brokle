package observability

import (
	"encoding/json"
	"time"

	"brokle/pkg/ulid"
)

// Trace represents a complete operation trace with hierarchy and session support
type Trace struct {
	// Identifiers
	ID            ulid.ULID  `json:"id" db:"id"`
	ProjectID     ulid.ULID  `json:"project_id" db:"project_id"`
	SessionID     *ulid.ULID `json:"session_id,omitempty" db:"session_id"`
	ParentTraceID *ulid.ULID `json:"parent_trace_id,omitempty" db:"parent_trace_id"`

	// Basic information
	Name      string     `json:"name" db:"name"`
	UserID    *ulid.ULID `json:"user_id,omitempty" db:"user_id"`
	Timestamp time.Time  `json:"timestamp" db:"timestamp"`

	// Data (stored as JSON strings in ClickHouse with ZSTD compression)
	Input  *string `json:"input,omitempty" db:"input"`
	Output *string `json:"output,omitempty" db:"output"`

	// Metadata and tags
	Metadata map[string]string `json:"metadata" db:"metadata"`
	Tags     []string          `json:"tags" db:"tags"`

	// Environment and versioning
	Environment string  `json:"environment" db:"environment"`
	Release     *string `json:"release,omitempty" db:"release"`

	// ReplacingMergeTree fields
	Version   uint32    `json:"version" db:"version"`
	EventTs   time.Time `json:"event_ts" db:"event_ts"`
	IsDeleted bool      `json:"is_deleted" db:"is_deleted"`

	// Populated from joins (not in ClickHouse)
	Observations []*Observation `json:"observations,omitempty" db:"-"`
	Scores       []*Score       `json:"scores,omitempty" db:"-"`
}

// Observation represents a span/event/generation within a trace
type Observation struct {
	// Identifiers
	ID                  ulid.ULID  `json:"id" db:"id"`
	TraceID             ulid.ULID  `json:"trace_id" db:"trace_id"`
	ParentObservationID *ulid.ULID `json:"parent_observation_id,omitempty" db:"parent_observation_id"`
	ProjectID           ulid.ULID  `json:"project_id" db:"project_id"`

	// Observation metadata
	Type      ObservationType `json:"type" db:"type"`
	Name      string          `json:"name" db:"name"`
	StartTime time.Time       `json:"start_time" db:"start_time"`
	EndTime   *time.Time      `json:"end_time,omitempty" db:"end_time"`

	// Model information
	Model           *string           `json:"model,omitempty" db:"model"`
	ModelParameters map[string]string `json:"model_parameters" db:"model_parameters"`

	// Data (stored as JSON strings in ClickHouse with ZSTD compression)
	Input    *string           `json:"input,omitempty" db:"input"`
	Output   *string           `json:"output,omitempty" db:"output"`
	Metadata map[string]string `json:"metadata" db:"metadata"`

	// Cost tracking (keys: input, output, total - all in USD)
	CostDetails map[string]float64 `json:"cost_details" db:"cost_details"`

	// Token usage (keys: prompt_tokens, completion_tokens, total_tokens)
	UsageDetails map[string]uint64 `json:"usage_details" db:"usage_details"`

	// Status and logging
	Level         ObservationLevel `json:"level" db:"level"`
	StatusMessage *string          `json:"status_message,omitempty" db:"status_message"`

	// Completion tracking for streaming responses
	CompletionStartTime *time.Time `json:"completion_start_time,omitempty" db:"completion_start_time"`
	TimeToFirstTokenMs  *uint32    `json:"time_to_first_token_ms,omitempty" db:"time_to_first_token_ms"`

	// ReplacingMergeTree fields
	Version   uint32    `json:"version" db:"version"`
	EventTs   time.Time `json:"event_ts" db:"event_ts"`
	IsDeleted bool      `json:"is_deleted" db:"is_deleted"`

	// Populated from joins (not in ClickHouse)
	Scores            []*Score       `json:"scores,omitempty" db:"-"`
	ChildObservations []*Observation `json:"child_observations,omitempty" db:"-"`
}

// Score represents a quality evaluation score
type Score struct {
	// Identifiers (at least one of trace_id, observation_id, or session_id must be set)
	ID            ulid.ULID  `json:"id" db:"id"`
	ProjectID     ulid.ULID  `json:"project_id" db:"project_id"`
	TraceID       *ulid.ULID `json:"trace_id,omitempty" db:"trace_id"`
	ObservationID *ulid.ULID `json:"observation_id,omitempty" db:"observation_id"`
	SessionID     *ulid.ULID `json:"session_id,omitempty" db:"session_id"`

	// Score data
	Name        string        `json:"name" db:"name"`
	Value       *float64      `json:"value,omitempty" db:"value"`
	StringValue *string       `json:"string_value,omitempty" db:"string_value"`
	DataType    ScoreDataType `json:"data_type" db:"data_type"`

	// Source and metadata
	Source  ScoreSource `json:"source" db:"source"`
	Comment *string     `json:"comment,omitempty" db:"comment"`

	// Evaluator information
	EvaluatorName    *string           `json:"evaluator_name,omitempty" db:"evaluator_name"`
	EvaluatorVersion *string           `json:"evaluator_version,omitempty" db:"evaluator_version"`
	EvaluatorConfig  map[string]string `json:"evaluator_config" db:"evaluator_config"`

	// Author tracking (for HUMAN source)
	AuthorUserID *ulid.ULID `json:"author_user_id,omitempty" db:"author_user_id"`

	// Timestamp
	Timestamp time.Time `json:"timestamp" db:"timestamp"`

	// ReplacingMergeTree fields
	Version   uint32    `json:"version" db:"version"`
	EventTs   time.Time `json:"event_ts" db:"event_ts"`
	IsDeleted bool      `json:"is_deleted" db:"is_deleted"`
}

// Session represents a user journey grouping multiple traces
type Session struct {
	// Identifiers
	ID        ulid.ULID  `json:"id" db:"id"`
	ProjectID ulid.ULID  `json:"project_id" db:"project_id"`
	UserID    *ulid.ULID `json:"user_id,omitempty" db:"user_id"`

	// Session metadata
	Metadata map[string]string `json:"metadata" db:"metadata"`

	// Feature flags
	Bookmarked bool `json:"bookmarked" db:"bookmarked"`
	Public     bool `json:"public" db:"public"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// ReplacingMergeTree fields
	Version   uint32    `json:"version" db:"version"`
	EventTs   time.Time `json:"event_ts" db:"event_ts"`
	IsDeleted bool      `json:"is_deleted" db:"is_deleted"`

	// Populated from joins (not in ClickHouse)
	Traces []*Trace `json:"traces,omitempty" db:"-"`
	Scores []*Score `json:"scores,omitempty" db:"-"`
}

// UpdateSessionRequest represents fields that can be updated in a session
// Uses pointers to distinguish between "not sent" (nil) and "explicitly set"
type UpdateSessionRequest struct {
	UserID     *ulid.ULID        `json:"user_id,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Bookmarked *bool             `json:"bookmarked,omitempty"` // nil = not sent, preserve existing
	Public     *bool             `json:"public,omitempty"`     // nil = not sent, preserve existing
}

// ObservationType defines the type of observation
type ObservationType string

const (
	ObservationTypeLLM        ObservationType = "LLM"
	ObservationTypeSpan       ObservationType = "SPAN"
	ObservationTypeEvent      ObservationType = "EVENT"
	ObservationTypeGeneration ObservationType = "GENERATION"
	ObservationTypeRetrieval  ObservationType = "RETRIEVAL"
	ObservationTypeEmbedding  ObservationType = "EMBEDDING"
	ObservationTypeAgent      ObservationType = "AGENT"
	ObservationTypeTool       ObservationType = "TOOL"
	ObservationTypeChain      ObservationType = "CHAIN"
)

// ObservationLevel defines the log level for observations
type ObservationLevel string

const (
	ObservationLevelDebug   ObservationLevel = "DEBUG"
	ObservationLevelInfo    ObservationLevel = "INFO"
	ObservationLevelWarn    ObservationLevel = "WARN"
	ObservationLevelError   ObservationLevel = "ERROR"
	ObservationLevelDefault ObservationLevel = "DEFAULT"
)

// ScoreDataType defines the data type of a quality score
type ScoreDataType string

const (
	ScoreDataTypeNumeric     ScoreDataType = "NUMERIC"
	ScoreDataTypeCategorical ScoreDataType = "CATEGORICAL"
	ScoreDataTypeBoolean     ScoreDataType = "BOOLEAN"
)

// ScoreSource defines the source of a quality score
type ScoreSource string

const (
	ScoreSourceAPI   ScoreSource = "API"
	ScoreSourceAuto  ScoreSource = "AUTO"
	ScoreSourceHuman ScoreSource = "HUMAN"
	ScoreSourceEval  ScoreSource = "EVAL"
)

// ===== Custom JSON Unmarshaling =====

// UnmarshalJSON implements custom JSON unmarshaling for Trace
// Handles input/output fields that may be strings, objects, or arrays from SDK
func (t *Trace) UnmarshalJSON(data []byte) error {
	// Create a temporary struct with json.RawMessage for input/output
	type Alias Trace
	aux := &struct {
		Input  json.RawMessage `json:"input,omitempty"`
		Output json.RawMessage `json:"output,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle input field
	if len(aux.Input) > 0 {
		t.Input = normalizeJSONField(aux.Input)
	}

	// Handle output field
	if len(aux.Output) > 0 {
		t.Output = normalizeJSONField(aux.Output)
	}

	return nil
}

// UnmarshalJSON implements custom JSON unmarshaling for Observation
// Handles input/output fields that may be strings, objects, or arrays from SDK
func (o *Observation) UnmarshalJSON(data []byte) error {
	// Create a temporary struct with json.RawMessage for input/output
	type Alias Observation
	aux := &struct {
		Input  json.RawMessage `json:"input,omitempty"`
		Output json.RawMessage `json:"output,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(o),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle input field
	if len(aux.Input) > 0 {
		o.Input = normalizeJSONField(aux.Input)
	}

	// Handle output field
	if len(aux.Output) > 0 {
		o.Output = normalizeJSONField(aux.Output)
	}

	return nil
}

// normalizeJSONField converts a JSON field to a string
// If it's already a string, unwrap it; if it's an object/array, keep it as JSON
func normalizeJSONField(raw json.RawMessage) *string {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}

	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return &str
	}

	// If not a string, it's an object/array - keep as JSON string
	jsonStr := string(raw)
	return &jsonStr
}

// ===== Trace Helper Methods =====

// IsRootTrace checks if this trace has no parent
func (t *Trace) IsRootTrace() bool {
	return t.ParentTraceID == nil
}

// HasSession checks if this trace belongs to a session
func (t *Trace) HasSession() bool {
	return t.SessionID != nil
}

// ===== Observation Helper Methods =====

// CalculateLatencyMs calculates the latency in milliseconds
func (o *Observation) CalculateLatencyMs() *uint32 {
	if o.EndTime == nil {
		return nil
	}
	latency := uint32(o.EndTime.Sub(o.StartTime).Milliseconds())
	return &latency
}

// IsCompleted checks if the observation has ended
func (o *Observation) IsCompleted() bool {
	return o.EndTime != nil
}

// HasParent checks if this observation has a parent observation
func (o *Observation) HasParent() bool {
	return o.ParentObservationID != nil
}

// GetTotalCost returns the total cost, calculated from cost_details map
func (o *Observation) GetTotalCost() float64 {
	if total, ok := o.CostDetails["total"]; ok {
		return total
	}
	return o.CostDetails["input"] + o.CostDetails["output"]
}

// GetTotalTokens returns the total tokens, calculated from usage_details map
func (o *Observation) GetTotalTokens() uint64 {
	if total, ok := o.UsageDetails["total_tokens"]; ok {
		return total
	}
	return o.UsageDetails["prompt_tokens"] + o.UsageDetails["completion_tokens"]
}

// SetCostDetails sets the cost details with input, output, and calculated total
func (o *Observation) SetCostDetails(inputCost, outputCost float64) {
	if o.CostDetails == nil {
		o.CostDetails = make(map[string]float64)
	}
	o.CostDetails["input"] = inputCost
	o.CostDetails["output"] = outputCost
	o.CostDetails["total"] = inputCost + outputCost
}

// SetUsageDetails sets the usage details with prompt, completion, and calculated total tokens
func (o *Observation) SetUsageDetails(promptTokens, completionTokens uint64) {
	if o.UsageDetails == nil {
		o.UsageDetails = make(map[string]uint64)
	}
	o.UsageDetails["prompt_tokens"] = promptTokens
	o.UsageDetails["completion_tokens"] = completionTokens
	o.UsageDetails["total_tokens"] = promptTokens + completionTokens
}

// ===== Score Helper Methods =====

// GetScoreLevel returns a human-readable quality level based on the score
func (s *Score) GetScoreLevel() string {
	switch s.DataType {
	case ScoreDataTypeNumeric, ScoreDataTypeBoolean:
		if s.Value != nil {
			if *s.Value >= 0.8 {
				return "excellent"
			} else if *s.Value >= 0.6 {
				return "good"
			} else if *s.Value >= 0.4 {
				return "fair"
			}
			return "poor"
		}
	case ScoreDataTypeCategorical:
		if s.StringValue != nil {
			return *s.StringValue
		}
	}
	return "unknown"
}

// IsNumeric checks if the score is numeric
func (s *Score) IsNumeric() bool {
	return s.DataType == ScoreDataTypeNumeric
}

// IsCategorical checks if the score is categorical
func (s *Score) IsCategorical() bool {
	return s.DataType == ScoreDataTypeCategorical
}

// IsBoolean checks if the score is boolean
func (s *Score) IsBoolean() bool {
	return s.DataType == ScoreDataTypeBoolean
}

// ==================================
// Telemetry Types
// ==================================

// TelemetryEventType represents the type of telemetry event
type TelemetryEventType string

const (
	// Generic events
	TelemetryEventTypeEvent TelemetryEventType = "event"

	// Structured observability (immutable events only)
	TelemetryEventTypeTrace        TelemetryEventType = "trace"
	TelemetryEventTypeSession      TelemetryEventType = "session"
	TelemetryEventTypeObservation  TelemetryEventType = "observation"
	TelemetryEventTypeQualityScore TelemetryEventType = "quality_score"
)

// TelemetryEventDeduplication represents a deduplication entry for telemetry events
type TelemetryEventDeduplication struct {
	EventID     ulid.ULID `json:"event_id" db:"event_id"`
	BatchID     ulid.ULID `json:"batch_id" db:"batch_id"`
	ProjectID   ulid.ULID `json:"project_id" db:"project_id"`
	FirstSeenAt time.Time `json:"first_seen_at" db:"first_seen_at"`
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at"`
}

// IsExpired checks if the deduplication entry has expired
func (d *TelemetryEventDeduplication) IsExpired() bool {
	return time.Now().After(d.ExpiresAt)
}

// TimeUntilExpiry returns the duration until the entry expires
func (d *TelemetryEventDeduplication) TimeUntilExpiry() time.Duration {
	return time.Until(d.ExpiresAt)
}

// Validate checks if the deduplication entry is valid
func (d *TelemetryEventDeduplication) Validate() []ValidationError {
	var errors []ValidationError

	if d.EventID.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "event_id",
			Message: "event_id is required",
		})
	}
	if d.BatchID.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "batch_id",
			Message: "batch_id is required",
		})
	}
	if d.ProjectID.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "project_id",
			Message: "project_id is required",
		})
	}

	return errors
}

// BatchStatus defines the status of a telemetry batch
type BatchStatus string

const (
	BatchStatusPending    BatchStatus = "pending"
	BatchStatusProcessing BatchStatus = "processing"
	BatchStatusCompleted  BatchStatus = "completed"
	BatchStatusFailed     BatchStatus = "failed"
)

// TelemetryBatch represents a batch of telemetry events
type TelemetryBatch struct {
	ID               ulid.ULID              `json:"id" db:"id"`
	ProjectID        ulid.ULID              `json:"project_id" db:"project_id"`
	Environment      string                 `json:"environment,omitempty" db:"environment"`
	BatchMetadata    map[string]interface{} `json:"batch_metadata" db:"batch_metadata"`
	TotalEvents      int                    `json:"total_events" db:"total_events"`
	ProcessedEvents  int                    `json:"processed_events" db:"processed_events"`
	FailedEvents     int                    `json:"failed_events" db:"failed_events"`
	Status           BatchStatus            `json:"status" db:"status"`
	ProcessingTimeMs *int                   `json:"processing_time_ms,omitempty" db:"processing_time_ms"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
	Events           []TelemetryEvent       `json:"events,omitempty" db:"-"`
}

// TelemetryEvent represents an individual telemetry event within a batch
type TelemetryEvent struct {
	ID           ulid.ULID              `json:"id" db:"id"`
	BatchID      ulid.ULID              `json:"batch_id" db:"batch_id"`
	ProjectID    ulid.ULID              `json:"project_id" db:"project_id"`
	Environment  string                 `json:"environment,omitempty" db:"environment"`
	EventType    TelemetryEventType     `json:"event_type" db:"event_type"`
	EventPayload map[string]interface{} `json:"event_payload" db:"event_payload"`
	ProcessedAt  *time.Time             `json:"processed_at,omitempty" db:"processed_at"`
	ErrorMessage *string                `json:"error_message,omitempty" db:"error_message"`
	RetryCount   int                    `json:"retry_count" db:"retry_count"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// TelemetryMetric represents a telemetry metric for performance and analytics tracking
type TelemetryMetric struct {
	ProjectID   ulid.ULID              `json:"project_id" db:"project_id"`
	Environment string                 `json:"environment,omitempty" db:"environment"`
	MetricName  string                 `json:"metric_name" db:"metric_name"`
	MetricType  string                 `json:"metric_type" db:"metric_type"`
	MetricValue float64                `json:"metric_value" db:"metric_value"`
	Labels      map[string]interface{} `json:"labels,omitempty" db:"labels"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty" db:"processed_at"`
}
