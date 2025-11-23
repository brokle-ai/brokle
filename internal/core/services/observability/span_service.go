package observability

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/observability"
	appErrors "brokle/pkg/errors"
)

// SpanService implements business logic for OTEL span (span) management
type SpanService struct {
	spanRepo  observability.SpanRepository
	traceRepo observability.TraceRepository
	scoreRepo observability.ScoreRepository
	logger    *logrus.Logger
}

// NewSpanService creates a new span service instance
func NewSpanService(
	spanRepo observability.SpanRepository,
	traceRepo observability.TraceRepository,
	scoreRepo observability.ScoreRepository,
	logger *logrus.Logger,
) *SpanService {
	return &SpanService{
		spanRepo:  spanRepo,
		traceRepo: traceRepo,
		scoreRepo: scoreRepo,
		logger:    logger,
	}
}

// CreateSpan creates a new OTEL span (span) with validation
func (s *SpanService) CreateSpan(ctx context.Context, span *observability.Span) error {
	// Validate required fields
	if span.TraceID == "" {
		return appErrors.NewValidationError("trace_id is required", "span must be linked to a trace")
	}
	if span.ProjectID == "" {
		return appErrors.NewValidationError("project_id is required", "span must have a valid project_id")
	}
	if span.SpanName == "" {
		return appErrors.NewValidationError("span_name is required", "span name cannot be empty")
	}
	if span.SpanID == "" {
		return appErrors.NewValidationError("span_id is required", "span must have OTEL span_id")
	}

	// Validate OTEL span_id format (16 hex chars)
	if len(span.SpanID) != 16 {
		return appErrors.NewValidationError("invalid span_id", "OTEL span_id must be 16 hex characters")
	}

	// Note: We do NOT validate trace existence here for async processing (eventual consistency)
	// Trace may still be in-flight when span arrives
	// ClickHouse ReplacingMergeTree handles eventual consistency gracefully
	// Note: We also do NOT validate parent span existence here
	// Async processing means parent may arrive after children - eventual consistency model
	// Database foreign key relationships will be preserved in final merged state

	// Set defaults for required fields
	if span.StatusCode == 0 {
		span.StatusCode = observability.StatusCodeUnset // UInt8: 0
	}
	if span.SpanKind == 0 {
		span.SpanKind = observability.SpanKindInternal // UInt8: 1
	}
	if span.Attributes == nil {
		span.Attributes = make(map[string]interface{})
	}
	if span.Metadata == nil {
		span.Metadata = make(map[string]interface{})
	}
	if span.CreatedAt.IsZero() {
		span.CreatedAt = time.Now()
	}

	// Note: All OTEL and Brokle span attributes stored in attributes JSON:
	// - brokle.span.type, brokle.span.level, brokle.span.version
	// - gen_ai.provider.name, gen_ai.request.model, gen_ai.usage.*
	// Materialized columns extract frequently-filtered attributes: model_name, provider_name, span_type, level, version

	// Note: Usage and cost data stored in dedicated Map columns:
	// - usage_details: Flexible token tracking (input, output, cache, audio, etc.)
	// - cost_details: Flexible cost breakdown by usage type
	// - pricing_snapshot: Audit trail for billing dispute resolution
	// - total_cost: Pre-computed total for fast SUM() aggregation

	// Calculate duration if not set
	span.CalculateDuration()

	// Store span directly in ClickHouse (ZSTD compression handles all sizes)
	if err := s.spanRepo.Create(ctx, span); err != nil {
		return appErrors.NewInternalError("failed to create span", err)
	}

	// Note: Trace aggregations (total_cost, total_tokens, span_count) calculated on-demand
	// Industry standard pattern: Query-time aggregation from spans using materialized columns

	return nil
}

// UpdateSpan updates an existing span
func (s *SpanService) UpdateSpan(ctx context.Context, span *observability.Span) error {
	// Validate span exists
	existing, err := s.spanRepo.GetByID(ctx, span.SpanID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError("span " + span.SpanID)
		}
		return appErrors.NewInternalError("failed to get span", err)
	}

	// Merge non-zero fields from incoming span into existing
	mergeSpanFields(existing, span)

	// Note: Version field removed (only exists in traces, not spans in OTEL schema)

	// Calculate duration if end time updated
	existing.CalculateDuration()

	// Update span directly in ClickHouse (ZSTD compression handles all sizes)
	if err := s.spanRepo.Update(ctx, existing); err != nil {
		return appErrors.NewInternalError("failed to update span", err)
	}

	return nil
}

// SetSpanCost sets cost details for a span
// Note: Costs stored in cost_details Map and total_cost Decimal
func (s *SpanService) SetSpanCost(ctx context.Context, spanID string, inputCost, outputCost float64) error {
	span, err := s.spanRepo.GetByID(ctx, spanID)
	if err != nil {
		return appErrors.NewNotFoundError("span " + spanID)
	}

	// Store in cost_details Map (as decimal.Decimal for precision)
	if span.CostDetails == nil {
		span.CostDetails = make(map[string]decimal.Decimal)
	}
	span.CostDetails["input"] = decimal.NewFromFloat(inputCost)
	span.CostDetails["output"] = decimal.NewFromFloat(outputCost)
	span.CostDetails["total"] = decimal.NewFromFloat(inputCost + outputCost)

	// Update total_cost for fast SUM() queries
	totalCost := decimal.NewFromFloat(inputCost + outputCost)
	span.TotalCost = &totalCost

	// Update span
	if err := s.spanRepo.Update(ctx, span); err != nil {
		return appErrors.NewInternalError("failed to update span cost", err)
	}

	return nil
}

// SetSpanUsage sets usage details for a span
// Note: Usage tokens stored in attributes JSON (gen_ai.usage.*) and usage_details Map
func (s *SpanService) SetSpanUsage(ctx context.Context, spanID string, promptTokens, completionTokens uint32) error {
	span, err := s.spanRepo.GetByID(ctx, spanID)
	if err != nil {
		return appErrors.NewNotFoundError("span " + spanID)
	}

	// Initialize attributes if needed
	if span.Attributes == nil {
		span.Attributes = make(map[string]interface{})
	}

	// Store in attributes JSON (for ClickHouse JSON queries)
	span.Attributes["gen_ai.usage.input_tokens"] = strconv.FormatUint(uint64(promptTokens), 10)
	span.Attributes["gen_ai.usage.output_tokens"] = strconv.FormatUint(uint64(completionTokens), 10)

	// Also store in usage_details Map (for aggregation queries)
	if span.UsageDetails == nil {
		span.UsageDetails = make(map[string]uint64)
	}
	span.UsageDetails["input"] = uint64(promptTokens)
	span.UsageDetails["output"] = uint64(completionTokens)
	span.UsageDetails["total"] = uint64(promptTokens + completionTokens)

	// Update span
	if err := s.spanRepo.Update(ctx, span); err != nil {
		return appErrors.NewInternalError("failed to update span usage", err)
	}

	return nil
}

// mergeSpanFields merges non-zero fields from src into dst
func mergeSpanFields(dst *observability.Span, src *observability.Span) {
	// Update core fields only if non-zero
	if src.SpanName != "" {
		dst.SpanName = src.SpanName
	}
	if src.SpanKind != 0 {
		dst.SpanKind = src.SpanKind
	}
	if !src.StartTime.IsZero() {
		dst.StartTime = src.StartTime
	}
	if src.EndTime != nil {
		dst.EndTime = src.EndTime
	}
	if src.StatusCode != 0 {
		dst.StatusCode = src.StatusCode
	}
	if src.StatusMessage != nil {
		dst.StatusMessage = src.StatusMessage
	}

	// Attribute and metadata fields (maps)
	if src.Attributes != nil && len(src.Attributes) > 0 {
		dst.Attributes = src.Attributes
	}
	if src.Metadata != nil && len(src.Metadata) > 0 {
		dst.Metadata = src.Metadata
	}

	// Usage and cost maps
	if src.UsageDetails != nil && len(src.UsageDetails) > 0 {
		dst.UsageDetails = src.UsageDetails
	}
	if src.CostDetails != nil && len(src.CostDetails) > 0 {
		dst.CostDetails = src.CostDetails
	}
	if src.PricingSnapshot != nil && len(src.PricingSnapshot) > 0 {
		dst.PricingSnapshot = src.PricingSnapshot
	}
	if src.TotalCost != nil {
		dst.TotalCost = src.TotalCost
	}

	// Input/Output
	if src.Input != nil {
		dst.Input = src.Input
	}
	if src.Output != nil {
		dst.Output = src.Output
	}

	// Events/Links arrays
	if len(src.EventsTimestamp) > 0 {
		dst.EventsTimestamp = src.EventsTimestamp
		dst.EventsName = src.EventsName
		dst.EventsAttributes = src.EventsAttributes
	}
	if len(src.LinksTraceID) > 0 {
		dst.LinksTraceID = src.LinksTraceID
		dst.LinksSpanID = src.LinksSpanID
		dst.LinksAttributes = src.LinksAttributes
	}

	// Note: Old dedicated fields (ModelName, Provider, InternalModelID, ModelParameters,
	// Usage/Cost maps, Prompt fields, Metadata, Type, Level) are now stored in
	// span_attributes JSON with proper namespaces and extracted to materialized columns.
	// No need to merge them here - they're part of span_attributes JSON.
}

// DeleteSpan soft deletes a span
func (s *SpanService) DeleteSpan(ctx context.Context, id string) error {
	// Validate span exists
	_, err := s.spanRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError("span " + id)
		}
		return appErrors.NewInternalError("failed to get span", err)
	}

	// Delete span
	if err := s.spanRepo.Delete(ctx, id); err != nil {
		return appErrors.NewInternalError("failed to delete span", err)
	}

	return nil
}

// GetSpanByID retrieves a span by its OTEL span_id
func (s *SpanService) GetSpanByID(ctx context.Context, id string) (*observability.Span, error) {
	span, err := s.spanRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError("span " + id)
		}
		return nil, appErrors.NewInternalError("failed to get span", err)
	}

	return span, nil
}

// GetSpansByTraceID retrieves all spans for a trace
func (s *SpanService) GetSpansByTraceID(ctx context.Context, traceID string) ([]*observability.Span, error) {
	spans, err := s.spanRepo.GetByTraceID(ctx, traceID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get spans", err)
	}

	return spans, nil
}

// GetRootSpan retrieves the root span for a trace (parent_span_id IS NULL)
func (s *SpanService) GetRootSpan(ctx context.Context, traceID string) (*observability.Span, error) {
	rootSpan, err := s.spanRepo.GetRootSpan(ctx, traceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError("root span for trace " + traceID)
		}
		return nil, appErrors.NewInternalError("failed to get root span", err)
	}

	return rootSpan, nil
}

// GetSpanTreeByTraceID retrieves all spans in a tree structure
func (s *SpanService) GetSpanTreeByTraceID(ctx context.Context, traceID string) ([]*observability.Span, error) {
	spans, err := s.spanRepo.GetTreeByTraceID(ctx, traceID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get span tree", err)
	}

	return spans, nil
}

// GetChildSpans retrieves child spans of a parent
func (s *SpanService) GetChildSpans(ctx context.Context, parentSpanID string) ([]*observability.Span, error) {
	spans, err := s.spanRepo.GetChildren(ctx, parentSpanID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get child spans", err)
	}

	return spans, nil
}

// GetSpansByFilter retrieves spans by filter criteria
func (s *SpanService) GetSpansByFilter(ctx context.Context, filter *observability.SpanFilter) ([]*observability.Span, error) {
	spans, err := s.spanRepo.GetByFilter(ctx, filter)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get spans", err)
	}

	return spans, nil
}

// CreateSpanBatch creates multiple spans in a batch
func (s *SpanService) CreateSpanBatch(ctx context.Context, spans []*observability.Span) error {
	if len(spans) == 0 {
		return nil
	}

	// Validate all spans
	for i, span := range spans {
		if span.TraceID == "" {
			return appErrors.NewValidationError(fmt.Sprintf("span[%d].trace_id", i), "trace_id is required")
		}
		if span.ProjectID == "" {
			return appErrors.NewValidationError(fmt.Sprintf("span[%d].project_id", i), "project_id is required")
		}
		if span.SpanName == "" {
			return appErrors.NewValidationError(fmt.Sprintf("span[%d].span_name", i), "span_name is required")
		}
		if span.SpanID == "" {
			return appErrors.NewValidationError(fmt.Sprintf("span[%d].span_id", i), "OTEL span_id is required")
		}

		// Set defaults for required fields
		if span.StatusCode == 0 {
			span.StatusCode = observability.StatusCodeUnset // UInt8: 0
		}
		if span.SpanKind == 0 {
			span.SpanKind = observability.SpanKindInternal // UInt8: 1
		}
		if span.Attributes == nil {
			span.Attributes = make(map[string]interface{})
		}
		if span.Metadata == nil {
			span.Metadata = make(map[string]interface{})
		}
		if span.CreatedAt.IsZero() {
			span.CreatedAt = time.Now()
		}

		// Note: All OTEL and Brokle data stored in attributes JSON and usage/cost Maps
		// Materialized columns extract frequently-filtered attributes

		// Calculate duration
		span.CalculateDuration()
	}

	// Create batch
	if err := s.spanRepo.CreateBatch(ctx, spans); err != nil {
		return appErrors.NewInternalError("failed to create span batch", err)
	}

	return nil
}

// CountSpans returns the count of spans matching the filter
func (s *SpanService) CountSpans(ctx context.Context, filter *observability.SpanFilter) (int64, error) {
	count, err := s.spanRepo.Count(ctx, filter)
	if err != nil {
		return 0, appErrors.NewInternalError("failed to count spans", err)
	}

	return count, nil
}

// CalculateTraceCost calculates total cost for all spans in a trace
func (s *SpanService) CalculateTraceCost(ctx context.Context, traceID string) (float64, error) {
	spans, err := s.spanRepo.GetByTraceID(ctx, traceID)
	if err != nil {
		return 0, appErrors.NewInternalError("failed to get spans", err)
	}

	totalCost := decimal.Zero
	for _, span := range spans {
		totalCost = totalCost.Add(span.GetTotalCost())
	}

	return totalCost.InexactFloat64(), nil
}

// CalculateTraceTokens calculates total tokens for all spans in a trace
func (s *SpanService) CalculateTraceTokens(ctx context.Context, traceID string) (uint32, error) {
	spans, err := s.spanRepo.GetByTraceID(ctx, traceID)
	if err != nil {
		return 0, appErrors.NewInternalError("failed to get spans", err)
	}

	var totalTokens uint64
	for _, span := range spans {
		totalTokens += span.GetTotalTokens()
	}

	return uint32(totalTokens), nil
}
