package observability

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"brokle/pkg/ulid"
)

// Trace represents an OTEL trace (root span) with trace-level context
type Trace struct {
	StartTime     time.Time  `json:"start_time" db:"start_time"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UserID        *string    `json:"user_id,omitempty" db:"user_id"`
	SessionID     *string    `json:"session_id,omitempty" db:"session_id"`
	EndTime       *time.Time `json:"end_time,omitempty" db:"end_time"`
	Duration      *uint64    `json:"duration,omitempty" db:"duration"` // Nanoseconds (OTLP spec)
	StatusMessage *string    `json:"status_message,omitempty" db:"status_message"`
	Input         *string    `json:"input,omitempty" db:"input"`
	Output        *string    `json:"output,omitempty" db:"output"`
	Environment   string     `json:"environment" db:"environment"`
	TraceID       string     `json:"trace_id" db:"trace_id"`
	Name          string     `json:"name" db:"name"`
	ProjectID     string     `json:"project_id" db:"project_id"`
	Tags          []string   `json:"tags" db:"tags"`

	// Aggregations (calculated from spans - ReplacingMergeTree)
	TotalCost   *decimal.Decimal `json:"total_cost,omitempty" db:"total_cost"`
	TotalTokens *uint32          `json:"total_tokens,omitempty" db:"total_tokens"`
	SpanCount   *uint32          `json:"span_count,omitempty" db:"span_count"`

	// Modern: JSON Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty" db:"metadata"`

	// Versioning & Release (Materialized from metadata)
	Release *string `json:"release,omitempty" db:"release"`
	Version *string `json:"version,omitempty" db:"version"`

	// Soft Delete
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	Spans      []*Span  `json:"spans,omitempty" db:"-"`
	Scores     []*Score `json:"scores,omitempty" db:"-"`
	Bookmarked bool     `json:"bookmarked" db:"bookmarked"`
	Public     bool     `json:"public" db:"public"`
	StatusCode uint8    `json:"status_code" db:"status_code"`
}

// Span represents an OTEL span with Gen AI semantic conventions and Brokle extensions
type Span struct {
	StartTime     time.Time  `json:"start_time" db:"start_time"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	EndTime       *time.Time `json:"end_time,omitempty" db:"end_time"`
	Duration      *uint64    `json:"duration,omitempty" db:"duration"` // Nanoseconds (OTLP spec)
	StatusMessage *string    `json:"status_message,omitempty" db:"status_message"`
	ParentSpanID  *string    `json:"parent_span_id,omitempty" db:"parent_span_id"`

	// W3C Trace Context
	TraceState *string `json:"trace_state,omitempty" db:"trace_state"`

	// Input/Output
	Input  *string `json:"input,omitempty" db:"input"`
	Output *string `json:"output,omitempty" db:"output"`

	TraceID   string `json:"trace_id" db:"trace_id"`
	SpanName  string `json:"span_name" db:"span_name"`
	SpanID    string `json:"span_id" db:"span_id"`
	ProjectID string `json:"project_id" db:"project_id"`
	// OTEL Events (Array of maps for performance)
	EventsTimestamp    []time.Time              `json:"events_timestamp,omitempty" db:"events_timestamp"`
	EventsName         []string                 `json:"events_name,omitempty" db:"events_name"`
	EventsAttributes   []map[string]interface{} `json:"events_attributes,omitempty" db:"events_attributes"`
	EventsDroppedCount []uint32                 `json:"events_dropped_attributes_count,omitempty" db:"events_dropped_attributes_count"`

	// OTEL Links (Array of maps with TraceState)
	LinksTraceID      []string                 `json:"links_trace_id,omitempty" db:"links_trace_id"`
	LinksSpanID       []string                 `json:"links_span_id,omitempty" db:"links_span_id"`
	LinksTraceState   []string                 `json:"links_trace_state,omitempty" db:"links_trace_state"`
	LinksAttributes   []map[string]interface{} `json:"links_attributes,omitempty" db:"links_attributes"`
	LinksDroppedCount []uint32                 `json:"links_dropped_attributes_count,omitempty" db:"links_dropped_attributes_count"`

	// ============================================
	// Modern: JSON Attributes + Usage/Cost Maps
	// ============================================
	Attributes map[string]interface{} `json:"attributes,omitempty" db:"attributes"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" db:"metadata"`

	UsageDetails    map[string]uint64      `json:"usage_details,omitempty" db:"usage_details"`
	CostDetails     map[string]decimal.Decimal `json:"cost_details,omitempty" db:"cost_details"`
	PricingSnapshot map[string]decimal.Decimal `json:"pricing_snapshot,omitempty" db:"pricing_snapshot"`
	TotalCost       *decimal.Decimal       `json:"total_cost,omitempty" db:"total_cost"`

	// ============================================
	// A/B Testing & Versioning
	// ============================================
	Version             *string    `json:"version,omitempty" db:"version"`
	CompletionStartTime *time.Time `json:"completion_start_time,omitempty" db:"completion_start_time"`

	// ============================================
	// Materialized Columns (from attributes JSON)
	// For fast filtering/sorting + API display
	// ============================================
	ModelName    *string `json:"model_name,omitempty" db:"-"`    // Materialized from attributes["gen_ai.request.model"]
	ProviderName *string `json:"provider_name,omitempty" db:"-"` // Materialized from attributes["gen_ai.provider.name"]
	SpanType     *string `json:"span_type,omitempty" db:"-"`     // Materialized from attributes["brokle.span.type"]
	Level        *string `json:"level,omitempty" db:"-"`         // Materialized from attributes["brokle.span.level"]

	// ============================================
	// Soft Delete
	// ============================================
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	Scores     []*Score `json:"scores,omitempty" db:"-"`
	ChildSpans []*Span  `json:"child_spans,omitempty" db:"-"`
	StatusCode uint8    `json:"status_code" db:"status_code"`
	SpanKind   uint8    `json:"span_kind" db:"span_kind"`
}

// Score represents a quality evaluation score linked to traces and spans
type Score struct {
	Timestamp        time.Time         `json:"timestamp" db:"timestamp"`
	Comment          *string           `json:"comment,omitempty" db:"comment"`
	Version          *string           `json:"version,omitempty" db:"version"`
	AuthorUserID     *string           `json:"author_user_id,omitempty" db:"author_user_id"`
	EvaluatorConfig  map[string]string `json:"evaluator_config" db:"evaluator_config"`
	Value            *float64          `json:"value,omitempty" db:"value"`
	StringValue      *string           `json:"string_value,omitempty" db:"string_value"`
	EvaluatorVersion *string           `json:"evaluator_version,omitempty" db:"evaluator_version"`
	EvaluatorName    *string           `json:"evaluator_name,omitempty" db:"evaluator_name"`
	Name             string            `json:"name" db:"name"`
	Source           string            `json:"source" db:"source"`
	DataType         string            `json:"data_type" db:"data_type"`
	ID               string            `json:"id" db:"id"`
	SpanID           string            `json:"span_id" db:"span_id"`
	TraceID          string            `json:"trace_id" db:"trace_id"`
	ProjectID        string            `json:"project_id" db:"project_id"`
}

// BlobStorageFileLog represents a reference to S3-stored large payload
// Used when payload > 10KB threshold
type BlobStorageFileLog struct {
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	FileSizeBytes *uint64   `json:"file_size_bytes,omitempty" db:"file_size_bytes"`
	Version       *string   `json:"version,omitempty" db:"version"`
	Compression   *string   `json:"compression,omitempty" db:"compression"`
	ContentType   *string   `json:"content_type,omitempty" db:"content_type"`
	EntityID      string    `json:"entity_id" db:"entity_id"`
	BucketPath    string    `json:"bucket_path" db:"bucket_path"`
	BucketName    string    `json:"bucket_name" db:"bucket_name"`
	EventID       string    `json:"event_id" db:"event_id"`
	ID            string    `json:"id" db:"id"`
	EntityType    string    `json:"entity_type" db:"entity_type"`
	ProjectID     string    `json:"project_id" db:"project_id"`
}

// Model represents an LLM/API model with pricing information (PostgreSQL)
// Used for cost calculation via internal_model_id lookup
type Model struct {
	UpdatedAt               time.Time  `json:"updated_at" db:"updated_at"`
	CreatedAt               time.Time  `json:"created_at" db:"created_at"`
	StartDate               *time.Time `json:"start_date,omitempty" db:"start_date"`
	ProjectID               *string    `json:"project_id,omitempty" db:"project_id"`
	TokenizerConfig         *string    `json:"tokenizer_config,omitempty" db:"tokenizer_config"`
	InputPrice              *float64   `json:"input_price,omitempty" db:"input_price"`
	OutputPrice             *float64   `json:"output_price,omitempty" db:"output_price"`
	TotalPrice              *float64   `json:"total_price,omitempty" db:"total_price"`
	TokenizerID             *string    `json:"tokenizer_id,omitempty" db:"tokenizer_id"`
	EndDate                 *time.Time `json:"end_date,omitempty" db:"end_date"`
	Provider                string     `json:"provider" db:"provider"`
	Unit                    string     `json:"unit" db:"unit"`
	ID                      string     `json:"id" db:"id"`
	MatchPattern            string     `json:"match_pattern" db:"match_pattern"`
	ModelName               string     `json:"model_name" db:"model_name"`
	BatchDiscountPercentage float64    `json:"batch_discount_percentage" db:"batch_discount_percentage"`
	CacheReadMultiplier     float64    `json:"cache_read_multiplier" db:"cache_read_multiplier"`
	CacheWriteMultiplier    float64    `json:"cache_write_multiplier" db:"cache_write_multiplier"`
	IsDeprecated            bool       `json:"is_deprecated" db:"is_deprecated"`
}

// Model business logic methods

// IsActive checks if pricing is currently active
func (m *Model) IsActive() bool {
	if m.IsDeprecated {
		return false
	}

	now := time.Now()

	// Check start date
	if m.StartDate != nil && now.Before(*m.StartDate) {
		return false
	}

	// Check end date (NULL = active)
	if m.EndDate != nil && now.After(*m.EndDate) {
		return false
	}

	return true
}

// IsGlobalPricing checks if this is global (non-project-specific) pricing
func (m *Model) IsGlobalPricing() bool {
	return m.ProjectID == nil
}

// CalculateInputCost calculates cost for input tokens
func (m *Model) CalculateInputCost(inputTokens int32, cacheHit bool) float64 {
	if m.InputPrice == nil {
		return 0.0
	}

	// Base cost (per-1M tokens)
	cost := (float64(inputTokens) / 1_000_000.0) * *m.InputPrice

	// Apply caching multiplier
	if cacheHit && m.CacheReadMultiplier > 0 {
		cost *= m.CacheReadMultiplier
	}

	return cost
}

// CalculateOutputCost calculates cost for output tokens
func (m *Model) CalculateOutputCost(outputTokens int32) float64 {
	if m.OutputPrice == nil {
		return 0.0
	}

	// Base cost (per-1M tokens)
	return (float64(outputTokens) / 1_000_000.0) * *m.OutputPrice
}

// CalculateTotalCost calculates total cost with optional batch discount
func (m *Model) CalculateTotalCost(inputTokens, outputTokens int32, cacheHit, batchMode bool) float64 {
	inputCost := m.CalculateInputCost(inputTokens, cacheHit)
	outputCost := m.CalculateOutputCost(outputTokens)
	totalCost := inputCost + outputCost

	// Apply batch discount
	if batchMode && m.BatchDiscountPercentage > 0 {
		totalCost *= (1.0 - m.BatchDiscountPercentage/100.0)
	}

	return totalCost
}

// Validate validates model pricing data
// Includes ReDoS protection for regex patterns
func (m *Model) Validate() []ValidationError {
	var errors []ValidationError

	// Required fields
	if m.ModelName == "" {
		errors = append(errors, ValidationError{
			Field:   "model_name",
			Message: "model name is required",
		})
	}

	if m.MatchPattern == "" {
		errors = append(errors, ValidationError{
			Field:   "match_pattern",
			Message: "match pattern is required",
		})
	}

	if m.Provider == "" {
		errors = append(errors, ValidationError{
			Field:   "provider",
			Message: "provider is required",
		})
	}

	if m.Unit == "" {
		errors = append(errors, ValidationError{
			Field:   "unit",
			Message: "pricing unit is required",
		})
	}

	// At least one price must be set
	if m.InputPrice == nil && m.OutputPrice == nil && m.TotalPrice == nil {
		errors = append(errors, ValidationError{
			Field:   "pricing",
			Message: "at least one price (input/output/total) is required",
		})
	}

	// Validate non-negative prices
	if m.InputPrice != nil && *m.InputPrice < 0 {
		errors = append(errors, ValidationError{
			Field:   "input_price",
			Message: "must be non-negative",
		})
	}

	if m.OutputPrice != nil && *m.OutputPrice < 0 {
		errors = append(errors, ValidationError{
			Field:   "output_price",
			Message: "must be non-negative",
		})
	}

	if m.TotalPrice != nil && *m.TotalPrice < 0 {
		errors = append(errors, ValidationError{
			Field:   "total_price",
			Message: "must be non-negative",
		})
	}

	// Validate regex pattern (ReDoS protection)
	if m.MatchPattern != "" {
		// Test regex compilation
		if _, err := regexp.Compile(m.MatchPattern); err != nil {
			errors = append(errors, ValidationError{
				Field:   "match_pattern",
				Message: fmt.Sprintf("invalid regex pattern: %v", err),
			})
		}

		// Pattern complexity checks (prevent ReDoS attacks)
		if len(m.MatchPattern) > 200 {
			errors = append(errors, ValidationError{
				Field:   "match_pattern",
				Message: "pattern too long (max 200 characters)",
			})
		}

		if strings.Count(m.MatchPattern, "*") > 10 {
			errors = append(errors, ValidationError{
				Field:   "match_pattern",
				Message: "pattern too complex (max 10 wildcards)",
			})
		}
	}

	// Validate temporal constraints
	if m.StartDate != nil && m.EndDate != nil && !m.EndDate.After(*m.StartDate) {
		errors = append(errors, ValidationError{
			Field:   "end_date",
			Message: "end date must be after start date",
		})
	}

	// Validate multipliers
	if m.CacheWriteMultiplier < 0 {
		errors = append(errors, ValidationError{
			Field:   "cache_write_multiplier",
			Message: "must be non-negative",
		})
	}

	if m.CacheReadMultiplier < 0 || m.CacheReadMultiplier > 1.0 {
		errors = append(errors, ValidationError{
			Field:   "cache_read_multiplier",
			Message: "must be between 0 and 1",
		})
	}

	if m.BatchDiscountPercentage < 0 || m.BatchDiscountPercentage > 100 {
		errors = append(errors, ValidationError{
			Field:   "batch_discount_percentage",
			Message: "must be between 0 and 100",
		})
	}

	return errors
}

// CostBreakdown represents detailed cost calculation result
type CostBreakdown struct {
	CacheSavings *float64 `json:"cache_savings,omitempty"`
	BatchSavings *float64 `json:"batch_savings,omitempty"`
	InputCost    string   `json:"input_cost"`
	OutputCost   string   `json:"output_cost"`
	TotalCost    string   `json:"total_cost"`
	Currency     string   `json:"currency"`
	ModelName    string   `json:"model_name"`
	Provider     string   `json:"provider"`
	InputTokens  int32    `json:"input_tokens"`
	OutputTokens int32    `json:"output_tokens"`
	CacheHit     bool     `json:"cache_hit"`
	BatchMode    bool     `json:"batch_mode"`
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
		*Alias
		Input  json.RawMessage `json:"input,omitempty"`
		Output json.RawMessage `json:"output,omitempty"`
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
		*Alias
		Input  json.RawMessage `json:"input,omitempty"`
		Output json.RawMessage `json:"output,omitempty"`
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
	if t.EndTime != nil && t.Duration == nil {
		duration := uint64(t.EndTime.Sub(t.StartTime).Nanoseconds())
		t.Duration = &duration
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
	if s.EndTime != nil && s.Duration == nil {
		duration := uint64(s.EndTime.Sub(s.StartTime).Nanoseconds())
		s.Duration = &duration
	}
}

// GetTotalCost returns the total cost from TotalCost field or CostDetails map
func (s *Span) GetTotalCost() decimal.Decimal {
	// Use pre-computed total_cost field
	if s.TotalCost != nil {
		return *s.TotalCost
	}
	// Fallback: sum from cost_details map
	total := decimal.Zero
	if s.CostDetails != nil {
		for _, cost := range s.CostDetails {
			total = total.Add(cost)
		}
	}
	return total
}

// GetTotalTokens returns the total tokens from usage_details map
func (s *Span) GetTotalTokens() uint64 {
	// Return from usage_details["total"] if available
	if s.UsageDetails != nil {
		if total, ok := s.UsageDetails["total"]; ok {
			return total
		}
		// Fallback: sum input + output
		var sum uint64
		if input, ok := s.UsageDetails["input"]; ok {
			sum += input
		}
		if output, ok := s.UsageDetails["output"]; ok {
			sum += output
		}
		return sum
	}
	return 0
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
	FirstSeenAt time.Time `json:"first_seen_at" db:"first_seen_at"`
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at"`
	EventID     string    `json:"event_id" db:"event_id"`
	BatchID     ulid.ULID `json:"batch_id" db:"batch_id"`
	ProjectID   ulid.ULID `json:"project_id" db:"project_id"`
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
