package observability

import (
	"encoding/json"
	"time"

	"brokle/pkg/ulid"
)

// Trace represents an OTEL trace (root span) with trace-level context
type Trace struct {
	// OTEL identifiers
	TraceID   string `json:"trace_id" db:"trace_id"`     // OTEL trace_id (32 hex chars) - renamed from ID
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
	StatusCode    uint8   `json:"status_code" db:"status_code"`         // OTEL enum: 0=UNSET, 1=OK, 2=ERROR
	StatusMessage *string `json:"status_message,omitempty" db:"status_message"`

	// OTEL resource attributes (JSON string with all resource-level attributes)
	ResourceAttributes string `json:"resource_attributes" db:"resource_attributes"`

	// Input/Output (trace-level data stored in ClickHouse with ZSTD compression)
	Input  *string `json:"input,omitempty" db:"input"`
	Output *string `json:"output,omitempty" db:"output"`

	// Tags for categorization
	Tags []string `json:"tags" db:"tags"`

	// OTEL resource attributes (extracted for common queries)
	Environment    string  `json:"environment" db:"environment"`
	ServiceName    *string `json:"service_name,omitempty" db:"service_name"`
	ServiceVersion *string `json:"service_version,omitempty" db:"service_version"`
	Release        *string `json:"release,omitempty" db:"release"`

	// Note: Aggregate metrics (total_cost, total_tokens, span_count) removed
	// Following industry standard (Langfuse/Datadog/Honeycomb pattern):
	// Aggregations calculated on-demand from spans table using materialized columns
	// Performance: 10-50ms for 1000 spans (ClickHouse columnar aggregation)

	// Flags (moved from sessions table)
	Bookmarked bool `json:"bookmarked" db:"bookmarked"`
	Public     bool `json:"public" db:"public"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Application versioning (experiment tracking)
	Version *string `json:"version,omitempty" db:"version"`

	// Populated from joins (not in ClickHouse)
	Spans  []*Span  `json:"spans,omitempty" db:"-"`
	Scores []*Score `json:"scores,omitempty" db:"-"`
}

// Span represents an OTEL span with Gen AI semantic conventions and Brokle extensions
type Span struct {
	// OTEL identifiers
	SpanID       string  `json:"span_id" db:"span_id"`                         // OTEL span_id (16 hex chars) - renamed from ID
	TraceID      string  `json:"trace_id" db:"trace_id"`                       // OTEL trace_id
	ParentSpanID *string `json:"parent_span_id,omitempty" db:"parent_span_id"` // NULL for root spans
	ProjectID    string  `json:"project_id" db:"project_id"`

	// Span data
	SpanName   string     `json:"span_name" db:"span_name"` // OTEL span name - renamed from Name
	SpanKind   uint8      `json:"span_kind" db:"span_kind"` // OTEL enum: 0=UNSPECIFIED, 1=INTERNAL, 2=SERVER, 3=CLIENT, 4=PRODUCER, 5=CONSUMER
	StartTime  time.Time  `json:"start_time" db:"start_time"`
	EndTime    *time.Time `json:"end_time,omitempty" db:"end_time"`
	DurationMs *uint32    `json:"duration_ms,omitempty" db:"duration_ms"`

	// OTEL status
	StatusCode    uint8   `json:"status_code" db:"status_code"` // OTEL enum: 0=UNSET, 1=OK, 2=ERROR
	StatusMessage *string `json:"status_message,omitempty" db:"status_message"`

	// OTEL attributes (JSON string with all span-level attributes)
	// Stores: gen_ai.*, brokle.*, and custom attributes
	SpanAttributes string `json:"span_attributes" db:"span_attributes"`

	// OTEL resource attributes (JSON string with resource-level context)
	ResourceAttributes string `json:"resource_attributes" db:"resource_attributes"`

	// Input/Output (stored in ClickHouse with ZSTD compression)
	Input  *string `json:"input,omitempty" db:"input"`
	Output *string `json:"output,omitempty" db:"output"`

	// OTEL Events (OTEL spec) - arrays for event tracking
	EventsTimestamp  []time.Time               `json:"events_timestamp,omitempty" db:"events_timestamp"`
	EventsName       []string                  `json:"events_name,omitempty" db:"events_name"`
	EventsAttributes []string                  `json:"events_attributes,omitempty" db:"events_attributes"` // JSON strings

	// OTEL Links (OTEL spec) - arrays for span linking
	LinksTraceID    []string `json:"links_trace_id,omitempty" db:"links_trace_id"`
	LinksSpanID     []string `json:"links_span_id,omitempty" db:"links_span_id"`
	LinksAttributes []string `json:"links_attributes,omitempty" db:"links_attributes"` // JSON strings

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// ===== MATERIALIZED COLUMNS (Read-only, computed by ClickHouse) =====
	// Gen AI attributes (OTEL 1.38+ conventions)
	GenAIOperationName       *string  `json:"gen_ai_operation_name,omitempty" db:"gen_ai_operation_name"`
	GenAIProviderName        *string  `json:"gen_ai_provider_name,omitempty" db:"gen_ai_provider_name"`
	GenAIRequestModel        *string  `json:"gen_ai_request_model,omitempty" db:"gen_ai_request_model"`
	GenAIRequestMaxTokens    *uint16  `json:"gen_ai_request_max_tokens,omitempty" db:"gen_ai_request_max_tokens"`
	GenAIRequestTemperature  *float32 `json:"gen_ai_request_temperature,omitempty" db:"gen_ai_request_temperature"`
	GenAIRequestTopP         *float32 `json:"gen_ai_request_top_p,omitempty" db:"gen_ai_request_top_p"`
	GenAIUsageInputTokens    *uint32  `json:"gen_ai_usage_input_tokens,omitempty" db:"gen_ai_usage_input_tokens"`
	GenAIUsageOutputTokens   *uint32  `json:"gen_ai_usage_output_tokens,omitempty" db:"gen_ai_usage_output_tokens"`

	// Brokle attributes (custom extensions)
	BrokleSpanType        *string  `json:"brokle_span_type,omitempty" db:"brokle_span_type"`       // span, generation, event, tool, agent, chain
	BrokleSpanLevel       *string  `json:"brokle_span_level,omitempty" db:"brokle_span_level"`     // DEBUG, INFO, WARNING, ERROR
	BrokleCostInput       *float64 `json:"brokle_cost_input,omitempty" db:"brokle_cost_input"`     // Decimal(18,9) extracted from STRING
	BrokleCostOutput      *float64 `json:"brokle_cost_output,omitempty" db:"brokle_cost_output"`   // Decimal(18,9) extracted from STRING
	BrokleCostTotal       *float64 `json:"brokle_cost_total,omitempty" db:"brokle_cost_total"`     // Decimal(18,9) extracted from STRING
	BroklePromptID        *string  `json:"brokle_prompt_id,omitempty" db:"brokle_prompt_id"`
	BroklePromptName      *string  `json:"brokle_prompt_name,omitempty" db:"brokle_prompt_name"`
	BroklePromptVersion   *uint16  `json:"brokle_prompt_version,omitempty" db:"brokle_prompt_version"`
	BrokleInternalModelID *string  `json:"brokle_internal_model_id,omitempty" db:"brokle_internal_model_id"`

	// Populated from joins (not in ClickHouse)
	Scores     []*Score `json:"scores,omitempty" db:"-"`
	ChildSpans []*Span  `json:"child_spans,omitempty" db:"-"`
}

// Score represents a quality evaluation score linked to traces and spans
type Score struct {
	// Identifiers
	ID        string `json:"id" db:"id"`
	ProjectID string `json:"project_id" db:"project_id"`
	TraceID   string `json:"trace_id" db:"trace_id"` // OTEL trace_id
	SpanID    string `json:"span_id" db:"span_id"`   // OTEL span_id

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

	// Application versioning (experiment tracking)
	Version *string `json:"version,omitempty" db:"version"`
}

// BlobStorageFileLog represents a reference to S3-stored large payload
// Used when payload > 10KB threshold
type BlobStorageFileLog struct {
	// Identifiers
	ID        string `json:"id" db:"id"`
	ProjectID string `json:"project_id" db:"project_id"`

	// Entity reference
	EntityType string `json:"entity_type" db:"entity_type"` // 'trace', 'span', 'score'
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

	// Application versioning (experiment tracking)
	Version *string `json:"version,omitempty" db:"version"`
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

// OTEL SpanKind enum values (UInt8 in ClickHouse)
const (
	SpanKindUnspecified uint8 = 0 // SPAN_KIND_UNSPECIFIED
	SpanKindInternal    uint8 = 1 // SPAN_KIND_INTERNAL
	SpanKindServer      uint8 = 2 // SPAN_KIND_SERVER
	SpanKindClient      uint8 = 3 // SPAN_KIND_CLIENT
	SpanKindProducer    uint8 = 4 // SPAN_KIND_PRODUCER
	SpanKindConsumer    uint8 = 5 // SPAN_KIND_CONSUMER
)

// SpanKind string constants for backwards compatibility
const (
	SpanKindUnspecifiedStr = "UNSPECIFIED"
	SpanKindInternalStr    = "INTERNAL"
	SpanKindServerStr      = "SERVER"
	SpanKindClientStr      = "CLIENT"
	SpanKindProducerStr    = "PRODUCER"
	SpanKindConsumerStr    = "CONSUMER"
)

// Brokle span type constants (stored in attributes as brokle.span.type)
const (
	SpanTypeSpan       = "span"
	SpanTypeGeneration = "generation"
	SpanTypeEvent      = "event"
	SpanTypeTool       = "tool"
	SpanTypeAgent      = "agent"
	SpanTypeChain      = "chain"
	SpanTypeRetrieval  = "retrieval"
	SpanTypeEmbedding  = "embedding"
)

// OTEL StatusCode enum values (UInt8 in ClickHouse)
const (
	StatusCodeUnset uint8 = 0 // STATUS_CODE_UNSET
	StatusCodeOK    uint8 = 1 // STATUS_CODE_OK
	StatusCodeError uint8 = 2 // STATUS_CODE_ERROR
)

// StatusCode string constants for backwards compatibility
const (
	StatusCodeUnsetStr = "UNSET"
	StatusCodeOKStr    = "OK"
	StatusCodeErrorStr = "ERROR"
)

// Span level constants
const (
	SpanLevelDebug   = "DEBUG"
	SpanLevelInfo    = "INFO"
	SpanLevelWarning = "WARNING"
	SpanLevelError   = "ERROR"
	SpanLevelDefault = "DEFAULT"
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

// UnmarshalJSON implements custom JSON unmarshaling for Span
// Handles input/output fields that may be strings, objects, or arrays from SDK
func (s *Span) UnmarshalJSON(data []byte) error {
	// Create a temporary struct with json.RawMessage for input/output
	type Alias Span
	aux := &struct {
		Input  json.RawMessage `json:"input,omitempty"`
		Output json.RawMessage `json:"output,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle input field
	if len(aux.Input) > 0 {
		s.Input = normalizeJSONField(aux.Input)
	}

	// Handle output field
	if len(aux.Output) > 0 {
		s.Output = normalizeJSONField(aux.Output)
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

// ===== Span Helper Methods =====

// IsCompleted checks if the span has ended
func (s *Span) IsCompleted() bool {
	return s.EndTime != nil
}

// HasParent checks if this span has a parent span
func (s *Span) HasParent() bool {
	return s.ParentSpanID != nil && *s.ParentSpanID != ""
}

// IsRootSpan checks if this is a root span (no parent)
func (s *Span) IsRootSpan() bool {
	return s.ParentSpanID == nil || *s.ParentSpanID == ""
}

// CalculateDuration calculates and sets the duration if not already set
func (s *Span) CalculateDuration() {
	if s.EndTime != nil && s.DurationMs == nil {
		duration := uint32(s.EndTime.Sub(s.StartTime).Milliseconds())
		s.DurationMs = &duration
	}
}

// GetTotalCost returns the total cost from materialized BrokleCostTotal field
func (s *Span) GetTotalCost() float64 {
	// Use materialized column from ClickHouse
	if s.BrokleCostTotal != nil {
		return *s.BrokleCostTotal
	}
	// Fallback: calculate from input + output materialized columns
	var inputCost, outputCost float64
	if s.BrokleCostInput != nil {
		inputCost = *s.BrokleCostInput
	}
	if s.BrokleCostOutput != nil {
		outputCost = *s.BrokleCostOutput
	}
	return inputCost + outputCost
}

// GetTotalTokens returns the total tokens from materialized Gen AI usage fields
func (s *Span) GetTotalTokens() uint64 {
	var total uint64
	if s.GenAIUsageInputTokens != nil {
		total += uint64(*s.GenAIUsageInputTokens)
	}
	if s.GenAIUsageOutputTokens != nil {
		total += uint64(*s.GenAIUsageOutputTokens)
	}
	return total
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
// OTEL Enum Converters
// ==================================

// ConvertStatusCodeToEnum converts a string status code to OTEL enum (UInt8)
func ConvertStatusCodeToEnum(statusStr string) uint8 {
	switch statusStr {
	case StatusCodeOKStr:
		return StatusCodeOK
	case StatusCodeErrorStr:
		return StatusCodeError
	case StatusCodeUnsetStr, "":
		return StatusCodeUnset
	default:
		return StatusCodeUnset
	}
}

// ConvertStatusCodeToString converts OTEL enum (UInt8) to string
func ConvertStatusCodeToString(statusCode uint8) string {
	switch statusCode {
	case StatusCodeOK:
		return StatusCodeOKStr
	case StatusCodeError:
		return StatusCodeErrorStr
	case StatusCodeUnset:
		return StatusCodeUnsetStr
	default:
		return StatusCodeUnsetStr
	}
}

// ConvertSpanKindToEnum converts a string span kind to OTEL enum (UInt8)
func ConvertSpanKindToEnum(kindStr string) uint8 {
	switch kindStr {
	case SpanKindInternalStr:
		return SpanKindInternal
	case SpanKindServerStr:
		return SpanKindServer
	case SpanKindClientStr:
		return SpanKindClient
	case SpanKindProducerStr:
		return SpanKindProducer
	case SpanKindConsumerStr:
		return SpanKindConsumer
	case SpanKindUnspecifiedStr, "":
		return SpanKindUnspecified
	default:
		return SpanKindUnspecified
	}
}

// ConvertSpanKindToString converts OTEL enum (UInt8) to string
func ConvertSpanKindToString(spanKind uint8) string {
	switch spanKind {
	case SpanKindInternal:
		return SpanKindInternalStr
	case SpanKindServer:
		return SpanKindServerStr
	case SpanKindClient:
		return SpanKindClientStr
	case SpanKindProducer:
		return SpanKindProducerStr
	case SpanKindConsumer:
		return SpanKindConsumerStr
	case SpanKindUnspecified:
		return SpanKindUnspecifiedStr
	default:
		return SpanKindUnspecifiedStr
	}
}

// ==================================
// Telemetry Types
// ==================================

// TelemetryEventType represents the type of telemetry event
type TelemetryEventType string

const (
	// Structured observability (immutable events only)
	TelemetryEventTypeTrace        TelemetryEventType = "trace"
	TelemetryEventTypeSession      TelemetryEventType = "session"
	TelemetryEventTypeSpan         TelemetryEventType = "span"
	TelemetryEventTypeQualityScore TelemetryEventType = "quality_score"
)

// TelemetryEventDeduplication represents a deduplication entry for telemetry events
// Internal type used by deduplication repository implementation
type TelemetryEventDeduplication struct {
	EventID     string    `json:"event_id" db:"event_id"`     // Composite OTLP ID: "trace_id:span_id"
	BatchID     ulid.ULID `json:"batch_id" db:"batch_id"`     // Brokle batch ULID
	ProjectID   ulid.ULID `json:"project_id" db:"project_id"` // Brokle project ULID
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

	if d.EventID == "" {
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
