package observability

import (
	"time"

	"brokle/pkg/ulid"
)

// Trace represents a complete LLM operation trace
type Trace struct {
	ID                ulid.ULID              `json:"id" db:"id"`
	ProjectID         ulid.ULID              `json:"project_id" db:"project_id"`
	SessionID         *ulid.ULID             `json:"session_id,omitempty" db:"session_id"`
	ExternalTraceID   string                 `json:"external_trace_id" db:"external_trace_id"`
	ParentTraceID     *ulid.ULID             `json:"parent_trace_id,omitempty" db:"parent_trace_id"`
	Name              string                 `json:"name" db:"name"`
	UserID            *ulid.ULID             `json:"user_id,omitempty" db:"user_id"`
	Tags              map[string]interface{} `json:"tags" db:"tags"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	Observations      []Observation          `json:"observations,omitempty" db:"-"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// Observation represents a single observation within a trace (LLM call, span, event)
type Observation struct {
	ID                    ulid.ULID              `json:"id" db:"id"`
	TraceID               ulid.ULID              `json:"trace_id" db:"trace_id"`
	ExternalObservationID string                 `json:"external_observation_id" db:"external_observation_id"`
	ParentObservationID   *ulid.ULID             `json:"parent_observation_id,omitempty" db:"parent_observation_id"`
	Type                  ObservationType        `json:"type" db:"type"`
	Name                  string                 `json:"name" db:"name"`
	StartTime             time.Time              `json:"start_time" db:"start_time"`
	EndTime               *time.Time             `json:"end_time,omitempty" db:"end_time"`
	Level                 ObservationLevel       `json:"level" db:"level"`
	StatusMessage         *string                `json:"status_message,omitempty" db:"status_message"`
	Version               *string                `json:"version,omitempty" db:"version"`
	Model                 *string                `json:"model,omitempty" db:"model"`
	Provider              *string                `json:"provider,omitempty" db:"provider"`
	Input                 map[string]interface{} `json:"input,omitempty" db:"input"`
	Output                map[string]interface{} `json:"output,omitempty" db:"output"`
	ModelParameters       map[string]interface{} `json:"model_parameters" db:"model_parameters"`
	PromptTokens          int                    `json:"prompt_tokens" db:"prompt_tokens"`
	CompletionTokens      int                    `json:"completion_tokens" db:"completion_tokens"`
	TotalTokens           int                    `json:"total_tokens" db:"total_tokens"`
	InputCost             *float64               `json:"input_cost,omitempty" db:"input_cost"`
	OutputCost            *float64               `json:"output_cost,omitempty" db:"output_cost"`
	TotalCost             *float64               `json:"total_cost,omitempty" db:"total_cost"`
	LatencyMs             *int                   `json:"latency_ms,omitempty" db:"latency_ms"`
	QualityScore          *float64               `json:"quality_score,omitempty" db:"quality_score"`
	QualityScores         []QualityScore         `json:"quality_scores,omitempty" db:"-"`
	CreatedAt             time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at" db:"updated_at"`
}

// QualityScore represents a quality evaluation score for a trace or observation
type QualityScore struct {
	ID                ulid.ULID     `json:"id" db:"id"`
	TraceID           ulid.ULID     `json:"trace_id" db:"trace_id"`
	ObservationID     *ulid.ULID    `json:"observation_id,omitempty" db:"observation_id"`
	ScoreName         string        `json:"score_name" db:"score_name"`
	ScoreValue        *float64      `json:"score_value,omitempty" db:"score_value"`
	StringValue       *string       `json:"string_value,omitempty" db:"string_value"`
	DataType          ScoreDataType `json:"data_type" db:"data_type"`
	Source            ScoreSource   `json:"source" db:"source"`
	EvaluatorName     *string       `json:"evaluator_name,omitempty" db:"evaluator_name"`
	EvaluatorVersion  *string       `json:"evaluator_version,omitempty" db:"evaluator_version"`
	Comment           *string       `json:"comment,omitempty" db:"comment"`
	AuthorUserID      *ulid.ULID    `json:"author_user_id,omitempty" db:"author_user_id"`
	CreatedAt         time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at" db:"updated_at"`
}

// ObservationType defines the type of observation
type ObservationType string

const (
	ObservationTypeLLM        ObservationType = "llm"
	ObservationTypeSpan       ObservationType = "span"
	ObservationTypeEvent      ObservationType = "event"
	ObservationTypeGeneration ObservationType = "generation"
	ObservationTypeRetrieval  ObservationType = "retrieval"
	ObservationTypeEmbedding  ObservationType = "embedding"
	ObservationTypeAgent      ObservationType = "agent"
	ObservationTypeTool       ObservationType = "tool"
	ObservationTypeChain      ObservationType = "chain"
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
	ScoreSourceAPI    ScoreSource = "API"
	ScoreSourceAuto   ScoreSource = "AUTO"
	ScoreSourceHuman  ScoreSource = "HUMAN"
	ScoreSourceEval   ScoreSource = "EVAL"
)

// TraceStats represents aggregated statistics for a trace
type TraceStats struct {
	TraceID              ulid.ULID `json:"trace_id"`
	TotalObservations    int       `json:"total_observations"`
	TotalLatencyMs       int       `json:"total_latency_ms"`
	TotalTokens          int       `json:"total_tokens"`
	TotalCost            float64   `json:"total_cost"`
	AverageQualityScore  *float64  `json:"average_quality_score,omitempty"`
	ErrorCount           int       `json:"error_count"`
	LLMObservationCount  int       `json:"llm_observation_count"`
	ProviderDistribution map[string]int `json:"provider_distribution"`
	ModelDistribution    map[string]int `json:"model_distribution"`
}

// ObservationStats represents aggregated statistics for observations
type ObservationStats struct {
	TotalCount           int64   `json:"total_count"`
	AverageLatencyMs     float64 `json:"average_latency_ms"`
	P95LatencyMs         float64 `json:"p95_latency_ms"`
	P99LatencyMs         float64 `json:"p99_latency_ms"`
	TotalCost            float64 `json:"total_cost"`
	AverageCostPerToken  float64 `json:"average_cost_per_token"`
	AverageQualityScore  float64 `json:"average_quality_score"`
	ErrorRate            float64 `json:"error_rate"`
	ThroughputPerMinute  float64 `json:"throughput_per_minute"`
}

// ValidationError represents a domain validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return e.Message
}

// ValidateTrace validates a trace entity
func (t *Trace) Validate() []ValidationError {
	var errors []ValidationError

	if t.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "trace name is required",
		})
	}

	if t.ExternalTraceID == "" {
		errors = append(errors, ValidationError{
			Field:   "external_trace_id",
			Message: "external trace ID is required",
		})
	}

	return errors
}

// ValidateObservation validates an observation entity
func (o *Observation) Validate() []ValidationError {
	var errors []ValidationError

	if o.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "observation name is required",
		})
	}

	if o.Type == "" {
		errors = append(errors, ValidationError{
			Field:   "type",
			Message: "observation type is required",
		})
	}

	if !o.isValidObservationType() {
		errors = append(errors, ValidationError{
			Field:   "type",
			Message: "invalid observation type",
		})
	}

	if o.ExternalObservationID == "" {
		errors = append(errors, ValidationError{
			Field:   "external_observation_id",
			Message: "external observation ID is required",
		})
	}

	if o.StartTime.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "start_time",
			Message: "start time is required",
		})
	}

	return errors
}

// isValidObservationType checks if the observation type is valid
func (o *Observation) isValidObservationType() bool {
	validTypes := []ObservationType{
		ObservationTypeLLM,
		ObservationTypeSpan,
		ObservationTypeEvent,
		ObservationTypeGeneration,
		ObservationTypeRetrieval,
		ObservationTypeEmbedding,
		ObservationTypeAgent,
		ObservationTypeTool,
		ObservationTypeChain,
	}

	for _, validType := range validTypes {
		if o.Type == validType {
			return true
		}
	}
	return false
}

// CalculateLatency calculates the latency in milliseconds
func (o *Observation) CalculateLatency() *int {
	if o.EndTime == nil || o.StartTime.IsZero() {
		return nil
	}

	latencyMs := int(o.EndTime.Sub(o.StartTime).Milliseconds())
	return &latencyMs
}

// IsCompleted checks if the observation is completed
func (o *Observation) IsCompleted() bool {
	return o.EndTime != nil && !o.EndTime.IsZero()
}

// ValidateQualityScore validates a quality score entity
func (q *QualityScore) Validate() []ValidationError {
	var errors []ValidationError

	if q.ScoreName == "" {
		errors = append(errors, ValidationError{
			Field:   "score_name",
			Message: "score name is required",
		})
	}

	if q.DataType == ScoreDataTypeNumeric && q.ScoreValue == nil {
		errors = append(errors, ValidationError{
			Field:   "score_value",
			Message: "numeric score value is required for numeric data type",
		})
	}

	if q.DataType == ScoreDataTypeCategorical && q.StringValue == nil {
		errors = append(errors, ValidationError{
			Field:   "string_value",
			Message: "string value is required for categorical data type",
		})
	}

	if q.DataType == ScoreDataTypeBoolean && q.ScoreValue == nil {
		errors = append(errors, ValidationError{
			Field:   "score_value",
			Message: "score value is required for boolean data type (0.0 or 1.0)",
		})
	}

	return errors
}

// IsNumeric checks if the score is numeric
func (q *QualityScore) IsNumeric() bool {
	return q.DataType == ScoreDataTypeNumeric
}

// IsCategorical checks if the score is categorical
func (q *QualityScore) IsCategorical() bool {
	return q.DataType == ScoreDataTypeCategorical
}

// IsBoolean checks if the score is boolean
func (q *QualityScore) IsBoolean() bool {
	return q.DataType == ScoreDataTypeBoolean
}

// QualityScoreAggregation represents aggregated quality score data
type QualityScoreAggregation struct {
	ScoreName string   `json:"score_name"`
	DataType  ScoreDataType `json:"data_type"`
	Count     int64    `json:"count"`
	AvgValue  *float64 `json:"avg_value,omitempty"`
	MinValue  *float64 `json:"min_value,omitempty"`
	MaxValue  *float64 `json:"max_value,omitempty"`
}