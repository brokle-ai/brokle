package observability

import (
	"encoding/json"
	"time"

	"brokle/pkg/ulid"
)

// Trace represents an OTEL trace (root span) with trace-level context
type Trace struct {
	// OTEL identifiers
	ID        string `json:"id" db:"id"`                 // OTEL trace_id (32 hex chars)
	ProjectID string `json:"project_id" db:"project_id"` // Brokle project context

	// Trace metadata
	Name      string  `json:"name" db:"name"`
	UserID    *string `json:"user_id,omitempty" db:"user_id"`
	SessionID *string `json:"session_id,omitempty" db:"session_id"` // Virtual session (attribute only, not FK)

	// Timing
	StartTime  time.Time  `json:"start_time" db:"start_time"`
	EndTime    *time.Time `json:"end_time,omitempty" db:"end_time"`
	DurationMs *uint32    `json:"duration_ms,omitempty" db:"duration_ms"`

	// OTEL status
	StatusCode    string  `json:"status_code" db:"status_code"` // OK, ERROR, UNSET
	StatusMessage *string `json:"status_message,omitempty" db:"status_message"`

	// OTEL attributes (JSON string for flexible key-value pairs)
	Attributes string `json:"attributes" db:"attributes"`

	// Input/Output (trace-level data)
	Input  *string `json:"input,omitempty" db:"input"`
	Output *string `json:"output,omitempty" db:"output"`

	// Blob storage (S3 offloading for large payloads >10KB)
	InputBlobStorageID  *string `json:"input_blob_storage_id,omitempty" db:"input_blob_storage_id"`
	OutputBlobStorageID *string `json:"output_blob_storage_id,omitempty" db:"output_blob_storage_id"`

	// Preview fields (ALWAYS populated - type-aware, adaptive 300-800 chars)
	InputPreview  *string `json:"input_preview,omitempty" db:"input_preview"`
	OutputPreview *string `json:"output_preview,omitempty" db:"output_preview"`

	// Metadata and tags
	Metadata map[string]string `json:"metadata" db:"metadata"`
	Tags     []string          `json:"tags" db:"tags"`

	// OTEL resource attributes
	Environment    string  `json:"environment" db:"environment"`
	ServiceName    *string `json:"service_name,omitempty" db:"service_name"`
	ServiceVersion *string `json:"service_version,omitempty" db:"service_version"`
	Release        *string `json:"release,omitempty" db:"release"`

	// Aggregate metrics (calculated from observations)
	TotalCost        *float64 `json:"total_cost,omitempty" db:"total_cost"`
	TotalTokens      *uint32  `json:"total_tokens,omitempty" db:"total_tokens"`
	ObservationCount *uint32  `json:"observation_count,omitempty" db:"observation_count"`

	// Flags (moved from sessions table)
	Bookmarked bool `json:"bookmarked" db:"bookmarked"`
	Public     bool `json:"public" db:"public"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// ReplacingMergeTree fields
	Version   uint32    `json:"version" db:"version"`
	EventTs   time.Time `json:"event_ts" db:"event_ts"`
	IsDeleted bool      `json:"is_deleted" db:"is_deleted"`

	// Populated from joins (not in ClickHouse)
	Observations []*Observation `json:"observations,omitempty" db:"-"`
	Scores       []*Score       `json:"scores,omitempty" db:"-"`
}

// Observation represents an OTEL span with Gen AI semantic conventions and Brokle extensions
type Observation struct {
	// OTEL identifiers
	ID                  string  `json:"id" db:"id"`                                                 // OTEL span_id (16 hex chars)
	TraceID             string  `json:"trace_id" db:"trace_id"`                                     // OTEL trace_id
	ParentObservationID *string `json:"parent_observation_id,omitempty" db:"parent_observation_id"` // NULL for root spans
	ProjectID           string  `json:"project_id" db:"project_id"`

	// Span data
	Name       string     `json:"name" db:"name"`
	SpanKind   string     `json:"span_kind" db:"span_kind"` // OTEL: INTERNAL, SERVER, CLIENT, PRODUCER, CONSUMER
	Type       string     `json:"type" db:"type"`           // Brokle: span, generation, event, tool, agent, chain
	StartTime  time.Time  `json:"start_time" db:"start_time"`
	EndTime    *time.Time `json:"end_time,omitempty" db:"end_time"`
	DurationMs *uint32    `json:"duration_ms,omitempty" db:"duration_ms"`

	// OTEL status
	StatusCode    string  `json:"status_code" db:"status_code"` // OK, ERROR, UNSET
	StatusMessage *string `json:"status_message,omitempty" db:"status_message"`

	// OTEL attributes (JSON string for flexible key-value pairs)
	Attributes string `json:"attributes" db:"attributes"`

	// Input/Output
	Input  *string `json:"input,omitempty" db:"input"`
	Output *string `json:"output,omitempty" db:"output"`

	// Preview fields (ALWAYS populated - type-aware, adaptive 300-800 chars)
	InputPreview  *string `json:"input_preview,omitempty" db:"input_preview"`
	OutputPreview *string `json:"output_preview,omitempty" db:"output_preview"`

	Metadata map[string]string `json:"metadata" db:"metadata"`
	Level    string            `json:"level" db:"level"` // DEBUG, INFO, WARNING, ERROR, DEFAULT

	// Universal model fields
	ModelName       *string `json:"model_name,omitempty" db:"model_name"`
	Provider        string  `json:"provider" db:"provider"`
	InternalModelID *string `json:"internal_model_id,omitempty" db:"internal_model_id"`
	ModelParameters *string `json:"model_parameters,omitempty" db:"model_parameters"` // JSON string

	// Usage & Cost Maps (Pattern: provided + calculated)
	ProvidedUsageDetails map[string]uint64  `json:"provided_usage_details,omitempty" db:"provided_usage_details"`
	UsageDetails         map[string]uint64  `json:"usage_details,omitempty" db:"usage_details"`
	ProvidedCostDetails  map[string]float64 `json:"provided_cost_details,omitempty" db:"provided_cost_details"`
	CostDetails          map[string]float64 `json:"cost_details,omitempty" db:"cost_details"`
	TotalCost            *float64           `json:"total_cost,omitempty" db:"total_cost"`

	// Prompt management
	PromptID      *string `json:"prompt_id,omitempty" db:"prompt_id"`
	PromptName    *string `json:"prompt_name,omitempty" db:"prompt_name"`
	PromptVersion *uint16 `json:"prompt_version,omitempty" db:"prompt_version"`

	// Blob storage references (Brokle innovation)
	InputBlobStorageID  *string `json:"input_blob_storage_id,omitempty" db:"input_blob_storage_id"`
	OutputBlobStorageID *string `json:"output_blob_storage_id,omitempty" db:"output_blob_storage_id"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// ReplacingMergeTree fields
	Version   uint32    `json:"version" db:"version"`
	EventTs   time.Time `json:"event_ts" db:"event_ts"`
	IsDeleted bool      `json:"is_deleted" db:"is_deleted"`

	// Populated from joins (not in ClickHouse)
	Scores            []*Score       `json:"scores,omitempty" db:"-"`
	ChildObservations []*Observation `json:"child_observations,omitempty" db:"-"`
}

// Score represents a quality evaluation score linked to traces and observations
type Score struct {
	// Identifiers
	ID            string `json:"id" db:"id"`
	ProjectID     string `json:"project_id" db:"project_id"`
	TraceID       string `json:"trace_id" db:"trace_id"`             // OTEL trace_id
	ObservationID string `json:"observation_id" db:"observation_id"` // OTEL span_id

	// Score data
	Name        string   `json:"name" db:"name"`
	Value       *float64 `json:"value,omitempty" db:"value"`
	StringValue *string  `json:"string_value,omitempty" db:"string_value"`
	DataType    string   `json:"data_type" db:"data_type"` // NUMERIC, CATEGORICAL, BOOLEAN

	// Metadata
	Source  string  `json:"source" db:"source"` // API, ANNOTATION, EVAL
	Comment *string `json:"comment,omitempty" db:"comment"`

	// Evaluator information
	EvaluatorName    *string           `json:"evaluator_name,omitempty" db:"evaluator_name"`
	EvaluatorVersion *string           `json:"evaluator_version,omitempty" db:"evaluator_version"`
	EvaluatorConfig  map[string]string `json:"evaluator_config" db:"evaluator_config"`
	AuthorUserID     *string           `json:"author_user_id,omitempty" db:"author_user_id"`

	// Timestamp
	Timestamp time.Time `json:"timestamp" db:"timestamp"`

	// ReplacingMergeTree fields
	Version   uint32    `json:"version" db:"version"`
	EventTs   time.Time `json:"event_ts" db:"event_ts"`
	IsDeleted bool      `json:"is_deleted" db:"is_deleted"`
}

// BlobStorageFileLog represents a reference to S3-stored large payload
// Used when payload > 10KB threshold
type BlobStorageFileLog struct {
	// Identifiers
	ID        string `json:"id" db:"id"`
	ProjectID string `json:"project_id" db:"project_id"`

	// Entity reference
	EntityType string `json:"entity_type" db:"entity_type"` // 'trace', 'observation', 'score'
	EntityID   string `json:"entity_id" db:"entity_id"`     // trace_id or span_id
	EventID    string `json:"event_id" db:"event_id"`       // Event ULID

	// Storage location
	BucketName string `json:"bucket_name" db:"bucket_name"`
	BucketPath string `json:"bucket_path" db:"bucket_path"`

	// Metadata
	FileSizeBytes *uint64 `json:"file_size_bytes,omitempty" db:"file_size_bytes"`
	ContentType   *string `json:"content_type,omitempty" db:"content_type"`
	Compression   *string `json:"compression,omitempty" db:"compression"` // 'gzip', 'zstd', null

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// ReplacingMergeTree fields
	Version   uint32    `json:"version" db:"version"`
	EventTs   time.Time `json:"event_ts" db:"event_ts"`
	IsDeleted bool      `json:"is_deleted" db:"is_deleted"`
}

// Model represents an LLM/API model with pricing information (PostgreSQL)
// Used for cost calculation via internal_model_id lookup
type Model struct {
	// Identifiers
	ID        string  `json:"id" db:"id"`
	ProjectID *string `json:"project_id,omitempty" db:"project_id"` // NULL = global model

	// Model identification
	ModelName    string `json:"model_name" db:"model_name"`       // gpt-4-turbo, claude-3-opus, etc.
	MatchPattern string `json:"match_pattern" db:"match_pattern"` // Regex for model aliases
	Provider     string `json:"provider" db:"provider"`           // openai, anthropic, google, etc.

	// Pricing (per 1k tokens by default)
	InputPrice  *float64 `json:"input_price,omitempty" db:"input_price"`
	OutputPrice *float64 `json:"output_price,omitempty" db:"output_price"`
	TotalPrice  *float64 `json:"total_price,omitempty" db:"total_price"` // Fallback for non-token pricing
	Unit        string   `json:"unit" db:"unit"`                         // TOKENS, CHARACTERS, REQUESTS, etc.

	// Versioning
	StartDate    *time.Time `json:"start_date,omitempty" db:"start_date"`
	IsDeprecated bool       `json:"is_deprecated" db:"is_deprecated"`

	// Tokenizer config (optional)
	TokenizerID     *string `json:"tokenizer_id,omitempty" db:"tokenizer_id"`
	TokenizerConfig *string `json:"tokenizer_config,omitempty" db:"tokenizer_config"` // JSONB

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// OTEL SpanKind constants
const (
	SpanKindInternal SpanKind = "INTERNAL"
	SpanKindServer   SpanKind = "SERVER"
	SpanKindClient   SpanKind = "CLIENT"
	SpanKindProducer SpanKind = "PRODUCER"
	SpanKindConsumer SpanKind = "CONSUMER"
)

type SpanKind string

// Brokle observation type constants (stored in attributes but also as dedicated field)
const (
	ObservationTypeSpan       = "span"
	ObservationTypeGeneration = "generation"
	ObservationTypeEvent      = "event"
	ObservationTypeTool       = "tool"
	ObservationTypeAgent      = "agent"
	ObservationTypeChain      = "chain"
	ObservationTypeRetrieval  = "retrieval"
	ObservationTypeEmbedding  = "embedding"
)

// OTEL StatusCode constants
const (
	StatusCodeUnset = "UNSET"
	StatusCodeOK    = "OK"
	StatusCodeError = "ERROR"
)

// Observation level constants
const (
	ObservationLevelDebug   = "DEBUG"
	ObservationLevelInfo    = "INFO"
	ObservationLevelWarning = "WARNING"
	ObservationLevelError   = "ERROR"
	ObservationLevelDefault = "DEFAULT"
)

// Score data type constants
const (
	ScoreDataTypeNumeric     = "NUMERIC"
	ScoreDataTypeCategorical = "CATEGORICAL"
	ScoreDataTypeBoolean     = "BOOLEAN"
)

// Score source constants
const (
	ScoreSourceAPI        = "API"
	ScoreSourceAnnotation = "ANNOTATION"
	ScoreSourceEval       = "EVAL"
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

// HasSession checks if this trace belongs to a virtual session
func (t *Trace) HasSession() bool {
	return t.SessionID != nil && *t.SessionID != ""
}

// IsCompleted checks if the trace has ended
func (t *Trace) IsCompleted() bool {
	return t.EndTime != nil
}

// CalculateDuration calculates and sets the duration if not already set
func (t *Trace) CalculateDuration() {
	if t.EndTime != nil && t.DurationMs == nil {
		duration := uint32(t.EndTime.Sub(t.StartTime).Milliseconds())
		t.DurationMs = &duration
	}
}

// ===== Observation Helper Methods =====

// IsCompleted checks if the observation has ended
func (o *Observation) IsCompleted() bool {
	return o.EndTime != nil
}

// HasParent checks if this observation has a parent observation
func (o *Observation) HasParent() bool {
	return o.ParentObservationID != nil && *o.ParentObservationID != ""
}

// IsRootSpan checks if this is a root span (no parent)
func (o *Observation) IsRootSpan() bool {
	return o.ParentObservationID == nil || *o.ParentObservationID == ""
}

// CalculateDuration calculates and sets the duration if not already set
func (o *Observation) CalculateDuration() {
	if o.EndTime != nil && o.DurationMs == nil {
		duration := uint32(o.EndTime.Sub(o.StartTime).Milliseconds())
		o.DurationMs = &duration
	}
}

// GetTotalCost returns the total cost from TotalCost field or calculated from cost details map
func (o *Observation) GetTotalCost() float64 {
	// Prefer denormalized TotalCost field
	if o.TotalCost != nil {
		return *o.TotalCost
	}
	// Fallback to cost details map
	if total, ok := o.CostDetails["total"]; ok {
		return total
	}
	// Calculate from input + output
	return o.CostDetails["input"] + o.CostDetails["output"]
}

// GetTotalTokens returns the total tokens from usage details map
func (o *Observation) GetTotalTokens() uint64 {
	// Check total in usage details
	if total, ok := o.UsageDetails["total"]; ok {
		return total
	}
	// Calculate from input + output
	return o.UsageDetails["input"] + o.UsageDetails["output"]
}

// SetCostDetails sets the cost details map with input, output, and total
func (o *Observation) SetCostDetails(inputCost, outputCost float64) {
	if o.CostDetails == nil {
		o.CostDetails = make(map[string]float64)
	}
	o.CostDetails["input"] = inputCost
	o.CostDetails["output"] = outputCost
	total := inputCost + outputCost
	o.CostDetails["total"] = total

	// Set denormalized field for fast queries
	o.TotalCost = &total
}

// SetUsageDetails sets the usage details map with input, output, and total tokens
func (o *Observation) SetUsageDetails(inputTokens, outputTokens uint64) {
	if o.UsageDetails == nil {
		o.UsageDetails = make(map[string]uint64)
	}
	o.UsageDetails["input"] = inputTokens
	o.UsageDetails["output"] = outputTokens
	o.UsageDetails["total"] = inputTokens + outputTokens
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

// ===== BlobStorageFileLog Helper Methods =====

// GetS3URI returns the full S3 URI
func (b *BlobStorageFileLog) GetS3URI() string {
	return "s3://" + b.BucketName + "/" + b.BucketPath
}

// IsCompressed checks if the content is compressed
func (b *BlobStorageFileLog) IsCompressed() bool {
	return b.Compression != nil && *b.Compression != ""
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
