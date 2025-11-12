package observability

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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
	if span.Name == "" {
		return appErrors.NewValidationError("name is required", "span name cannot be empty")
	}
	if span.ID == "" {
		return appErrors.NewValidationError("id is required", "span must have OTEL span_id")
	}

	// Validate OTEL span_id format (16 hex chars)
	if len(span.ID) != 16 {
		return appErrors.NewValidationError("invalid span_id", "OTEL span_id must be 16 hex characters")
	}

	// Note: We do NOT validate trace existence here for async processing (eventual consistency)
	// Trace may still be in-flight when span arrives
	// ClickHouse ReplacingMergeTree handles eventual consistency gracefully
	// Note: We also do NOT validate parent span existence here
	// Async processing means parent may arrive after children - eventual consistency model
	// Database foreign key relationships will be preserved in final merged state

	// Set defaults
	if span.StatusCode == "" {
		span.StatusCode = observability.StatusCodeUnset
	}
	if span.SpanKind == "" {
		span.SpanKind = string(observability.SpanKindInternal)
	}
	if span.Type == "" {
		span.Type = observability.SpanTypeSpan
	}
	if span.Level == "" {
		span.Level = observability.SpanLevelDefault
	}
	if span.Attributes == "" {
		span.Attributes = "{}"
	}
	if span.Provider == "" {
		span.Provider = ""
	}
	if span.CreatedAt.IsZero() {
		span.CreatedAt = time.Now()
	}

	// Initialize maps if nil
	if span.Metadata == nil {
		span.Metadata = make(map[string]interface{})
	}
	if span.ProvidedUsageDetails == nil {
		span.ProvidedUsageDetails = make(map[string]uint64)
	}
	if span.UsageDetails == nil {
		span.UsageDetails = make(map[string]uint64)
	}
	if span.ProvidedCostDetails == nil {
		span.ProvidedCostDetails = make(map[string]float64)
	}
	if span.CostDetails == nil {
		span.CostDetails = make(map[string]float64)
	}

	// Calculate duration if not set
	span.CalculateDuration()

	// Store span directly in ClickHouse (ZSTD compression handles all sizes)
	if err := s.spanRepo.Create(ctx, span); err != nil {
		return appErrors.NewInternalError("failed to create span", err)
	}

	return nil
}

// UpdateSpan updates an existing span
func (s *SpanService) UpdateSpan(ctx context.Context, span *observability.Span) error {
	// Validate span exists
	existing, err := s.spanRepo.GetByID(ctx, span.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError(fmt.Sprintf("span %s", span.ID))
		}
		return appErrors.NewInternalError("failed to get span", err)
	}

	// Merge non-zero fields from incoming span into existing
	mergeSpanFields(existing, span)

	// Preserve version for increment in repository layer
	existing.Version = existing.Version

	// Calculate duration if end time updated
	existing.CalculateDuration()

	// Update span directly in ClickHouse (ZSTD compression handles all sizes)
	if err := s.spanRepo.Update(ctx, existing); err != nil {
		return appErrors.NewInternalError("failed to update span", err)
	}

	return nil
}

// SetSpanCost sets cost details for a span
func (s *SpanService) SetSpanCost(ctx context.Context, spanID string, inputCost, outputCost float64) error {
	span, err := s.spanRepo.GetByID(ctx, spanID)
	if err != nil {
		return appErrors.NewNotFoundError(fmt.Sprintf("span %s", spanID))
	}

	// Set cost details (updates both Maps and Brokle extension fields)
	span.SetCostDetails(inputCost, outputCost)

	// Update span
	if err := s.spanRepo.Update(ctx, span); err != nil {
		return appErrors.NewInternalError("failed to update span cost", err)
	}

	return nil
}

// SetSpanUsage sets usage details for a span
func (s *SpanService) SetSpanUsage(ctx context.Context, spanID string, promptTokens, completionTokens uint32) error {
	span, err := s.spanRepo.GetByID(ctx, spanID)
	if err != nil {
		return appErrors.NewNotFoundError(fmt.Sprintf("span %s", spanID))
	}

	// Set usage details (populates usage_details Map)
	span.SetUsageDetails(uint64(promptTokens), uint64(completionTokens))

	// Update span
	if err := s.spanRepo.Update(ctx, span); err != nil {
		return appErrors.NewInternalError("failed to update span usage", err)
	}

	return nil
}

// mergeSpanFields merges non-zero fields from src into dst
func mergeSpanFields(dst *observability.Span, src *observability.Span) {
	// Update optional fields only if non-zero
	if src.Name != "" {
		dst.Name = src.Name
	}
	if src.SpanKind != "" {
		dst.SpanKind = src.SpanKind
	}
	if src.Type != "" {
		dst.Type = src.Type
	}
	if !src.StartTime.IsZero() {
		dst.StartTime = src.StartTime
	}
	if src.EndTime != nil {
		dst.EndTime = src.EndTime
	}
	if src.StatusCode != "" {
		dst.StatusCode = src.StatusCode
	}
	if src.StatusMessage != nil {
		dst.StatusMessage = src.StatusMessage
	}
	if src.Attributes != "" {
		dst.Attributes = src.Attributes
	}
	if src.Input != nil {
		dst.Input = src.Input
	}
	if src.Output != nil {
		dst.Output = src.Output
	}
	if src.Metadata != nil {
		dst.Metadata = src.Metadata
	}
	if src.Level != "" {
		dst.Level = src.Level
	}

	// Model fields
	if src.ModelName != nil {
		dst.ModelName = src.ModelName
	}
	if src.Provider != "" {
		dst.Provider = src.Provider
	}
	if src.InternalModelID != nil {
		dst.InternalModelID = src.InternalModelID
	}
	if src.ModelParameters != nil {
		dst.ModelParameters = src.ModelParameters
	}

	// Usage & Cost Maps
	if src.ProvidedUsageDetails != nil {
		dst.ProvidedUsageDetails = src.ProvidedUsageDetails
	}
	if src.UsageDetails != nil {
		dst.UsageDetails = src.UsageDetails
	}
	if src.ProvidedCostDetails != nil {
		dst.ProvidedCostDetails = src.ProvidedCostDetails
	}
	if src.CostDetails != nil {
		dst.CostDetails = src.CostDetails
	}
	if src.TotalCost != nil {
		dst.TotalCost = src.TotalCost
	}

	// Prompt management
	if src.PromptID != nil {
		dst.PromptID = src.PromptID
	}
	if src.PromptName != nil {
		dst.PromptName = src.PromptName
	}
	if src.PromptVersion != nil {
		dst.PromptVersion = src.PromptVersion
	}
}

// DeleteSpan soft deletes a span
func (s *SpanService) DeleteSpan(ctx context.Context, id string) error {
	// Validate span exists
	_, err := s.spanRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError(fmt.Sprintf("span %s", id))
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
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("span %s", id))
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
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("root span for trace %s", traceID))
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
		if span.Name == "" {
			return appErrors.NewValidationError(fmt.Sprintf("span[%d].name", i), "name is required")
		}
		if span.ID == "" {
			return appErrors.NewValidationError(fmt.Sprintf("span[%d].id", i), "OTEL span_id is required")
		}

		// Set defaults
		if span.StatusCode == "" {
			span.StatusCode = observability.StatusCodeUnset
		}
		if span.SpanKind == "" {
			span.SpanKind = string(observability.SpanKindInternal)
		}
		if span.Type == "" {
			span.Type = observability.SpanTypeSpan
		}
		if span.Level == "" {
			span.Level = observability.SpanLevelDefault
		}
		if span.Attributes == "" {
			span.Attributes = "{}"
		}
		if span.CreatedAt.IsZero() {
			span.CreatedAt = time.Now()
		}

		// Initialize maps if nil
		if span.Metadata == nil {
			span.Metadata = make(map[string]interface{})
		}
		if span.ProvidedUsageDetails == nil {
			span.ProvidedUsageDetails = make(map[string]uint64)
		}
		if span.UsageDetails == nil {
			span.UsageDetails = make(map[string]uint64)
		}
		if span.ProvidedCostDetails == nil {
			span.ProvidedCostDetails = make(map[string]float64)
		}
		if span.CostDetails == nil {
			span.CostDetails = make(map[string]float64)
		}

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

	var totalCost float64
	for _, span := range spans {
		totalCost += span.GetTotalCost()
	}

	return totalCost, nil
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
