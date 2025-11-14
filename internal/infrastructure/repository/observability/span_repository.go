package observability

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/core/domain/observability"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type spanRepository struct {
	db clickhouse.Conn
}

// NewSpanRepository creates a new span repository instance
func NewSpanRepository(db clickhouse.Conn) observability.SpanRepository {
	return &spanRepository{db: db}
}

// Create inserts a new OTEL span into ClickHouse
func (r *spanRepository) Create(ctx context.Context, span *observability.Span) error {
	// Set updated timestamp
	span.UpdatedAt = time.Now()

	// Calculate duration if not set
	span.CalculateDuration()

	query := `
		INSERT INTO spans (
			span_id, trace_id, parent_span_id, project_id,
			span_name, span_kind, start_time, end_time, duration_ms,
			status_code, status_message,
			span_attributes, resource_attributes,
			input, output,
			events_timestamp, events_name, events_attributes,
			links_trace_id, links_span_id, links_attributes,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.db.Exec(ctx, query,
		span.SpanID, // Renamed from ID
		span.TraceID,
		span.ParentSpanID,
		span.ProjectID,
		span.SpanName,   // Renamed from Name
		span.SpanKind,   // Now UInt8 (0-5)
		span.StartTime,
		span.EndTime,
		span.DurationMs,
		span.StatusCode, // Now UInt8 (0-2)
		span.StatusMessage,
		span.SpanAttributes,     // All attributes (gen_ai.*, brokle.*, custom)
		span.ResourceAttributes, // Resource-level attributes
		span.Input,
		span.Output,
		span.EventsTimestamp,  // OTEL Events arrays
		span.EventsName,
		span.EventsAttributes,
		span.LinksTraceID, // OTEL Links arrays
		span.LinksSpanID,
		span.LinksAttributes,
		span.CreatedAt,
		span.UpdatedAt,
		// Note: Materialized columns (gen_ai_*, brokle_*) are NOT inserted - ClickHouse computes them
		// Removed 19 columns: type, level, model_name, provider, internal_model_id, model_parameters,
		//                     provided_usage_details, usage_details, provided_cost_details, cost_details,
		//                     total_cost, prompt_id, prompt_name, prompt_version, version, metadata,
		//                     event_ts, is_deleted, attributes (split into span_attributes + resource_attributes)
	)
}

// Update performs an update by inserting new data (MergeTree will handle deduplication)
func (r *spanRepository) Update(ctx context.Context, span *observability.Span) error {
	// Set updated timestamp
	span.UpdatedAt = time.Now()

	// Calculate duration if not set
	span.CalculateDuration()

	// MergeTree pattern: INSERT new version, ClickHouse merges based on ORDER BY (span_id, updated_at)
	return r.Create(ctx, span)
}

// Delete performs hard deletion (MergeTree supports lightweight deletes with DELETE mutation)
func (r *spanRepository) Delete(ctx context.Context, id string) error {
	// MergeTree lightweight DELETE (async mutation, eventually consistent)
	query := `ALTER TABLE spans DELETE WHERE span_id = ?`
	return r.db.Exec(ctx, query, id)
}

// GetByID retrieves a span by its OTEL span_id
func (r *spanRepository) GetByID(ctx context.Context, id string) (*observability.Span, error) {
	query := `
		SELECT
			span_id, trace_id, parent_span_id, project_id,
			span_name, span_kind, start_time, end_time, duration_ms,
			status_code, status_message,
			span_attributes, resource_attributes,
			input, output,
			events_timestamp, events_name, events_attributes,
			links_trace_id, links_span_id, links_attributes,
			created_at, updated_at,
			gen_ai_operation_name, gen_ai_provider_name, gen_ai_request_model,
			gen_ai_usage_input_tokens, gen_ai_usage_output_tokens,
			brokle_span_type, brokle_span_level,
			brokle_cost_input, brokle_cost_output, brokle_cost_total,
			brokle_prompt_id, brokle_prompt_name, brokle_prompt_version,
			brokle_internal_model_id
		FROM spans
		WHERE span_id = ?
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, id)
	return r.scanSpanRow(row)
}

// GetByTraceID retrieves all spans for a trace
func (r *spanRepository) GetByTraceID(ctx context.Context, traceID string) ([]*observability.Span, error) {
	query := `
		SELECT
			span_id, trace_id, parent_span_id, project_id,
			span_name, span_kind, start_time, end_time, duration_ms,
			status_code, status_message,
			span_attributes, resource_attributes,
			input, output,
			events_timestamp, events_name, events_attributes,
			links_trace_id, links_span_id, links_attributes,
			created_at, updated_at,
			gen_ai_operation_name, gen_ai_provider_name, gen_ai_request_model,
			gen_ai_usage_input_tokens, gen_ai_usage_output_tokens,
			brokle_span_type, brokle_span_level,
			brokle_cost_input, brokle_cost_output, brokle_cost_total,
			brokle_prompt_id, brokle_prompt_name, brokle_prompt_version,
			brokle_internal_model_id
		FROM spans
		WHERE trace_id = ?
		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(ctx, query, traceID)
	if err != nil {
		return nil, fmt.Errorf("query spans by trace: %w", err)
	}
	defer rows.Close()

	return r.scanSpans(rows)
}

// GetRootSpan retrieves the root span for a trace (parent_span_id IS NULL)
func (r *spanRepository) GetRootSpan(ctx context.Context, traceID string) (*observability.Span, error) {
	query := `
		SELECT
			span_id, trace_id, parent_span_id, project_id,
			span_name, span_kind, start_time, end_time, duration_ms,
			status_code, status_message,
			span_attributes, resource_attributes,
			input, output,
			events_timestamp, events_name, events_attributes,
			links_trace_id, links_span_id, links_attributes,
			created_at, updated_at,
			gen_ai_operation_name, gen_ai_provider_name, gen_ai_request_model,
			gen_ai_usage_input_tokens, gen_ai_usage_output_tokens,
			brokle_span_type, brokle_span_level,
			brokle_cost_input, brokle_cost_output, brokle_cost_total,
			brokle_prompt_id, brokle_prompt_name, brokle_prompt_version,
			brokle_internal_model_id
		FROM spans
		WHERE trace_id = ? AND parent_span_id IS NULL
		ORDER BY start_time DESC
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, traceID)
	return r.scanSpanRow(row)
}

// GetChildren retrieves child spans of a parent span
func (r *spanRepository) GetChildren(ctx context.Context, parentSpanID string) ([]*observability.Span, error) {
	query := `
		SELECT
			span_id, trace_id, parent_span_id, project_id,
			span_name, span_kind, start_time, end_time, duration_ms,
			status_code, status_message,
			span_attributes, resource_attributes,
			input, output,
			events_timestamp, events_name, events_attributes,
			links_trace_id, links_span_id, links_attributes,
			created_at, updated_at,
			gen_ai_operation_name, gen_ai_provider_name, gen_ai_request_model,
			gen_ai_usage_input_tokens, gen_ai_usage_output_tokens,
			brokle_span_type, brokle_span_level,
			brokle_cost_input, brokle_cost_output, brokle_cost_total,
			brokle_prompt_id, brokle_prompt_name, brokle_prompt_version,
			brokle_internal_model_id
		FROM spans
		WHERE parent_span_id = ?		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(ctx, query, parentSpanID)
	if err != nil {
		return nil, fmt.Errorf("query child spans: %w", err)
	}
	defer rows.Close()

	return r.scanSpans(rows)
}

// GetTreeByTraceID retrieves all spans for a trace (recursive tree)
func (r *spanRepository) GetTreeByTraceID(ctx context.Context, traceID string) ([]*observability.Span, error) {
	// Return all spans in start_time order (building tree is done in service layer)
	return r.GetByTraceID(ctx, traceID)
}

// GetByFilter retrieves spans by filter criteria
func (r *spanRepository) GetByFilter(ctx context.Context, filter *observability.SpanFilter) ([]*observability.Span, error) {
	query := `
		SELECT
			span_id, trace_id, parent_span_id, project_id,
			span_name, span_kind, start_time, end_time, duration_ms,
			status_code, status_message,
			span_attributes, resource_attributes,
			input, output,
			events_timestamp, events_name, events_attributes,
			links_trace_id, links_span_id, links_attributes,
			created_at, updated_at,
			gen_ai_operation_name, gen_ai_provider_name, gen_ai_request_model,
			gen_ai_usage_input_tokens, gen_ai_usage_output_tokens,
			brokle_span_type, brokle_span_level,
			brokle_cost_input, brokle_cost_output, brokle_cost_total,
			brokle_prompt_id, brokle_prompt_name, brokle_prompt_version,
			brokle_internal_model_id
		FROM spans
		WHERE 1=1
	`

	args := []interface{}{}

	// Apply filters
	if filter != nil {
		if filter.TraceID != nil {
			query += " AND trace_id = ?"
			args = append(args, *filter.TraceID)
		}
		if filter.ParentID != nil {
			query += " AND parent_span_id = ?"
			args = append(args, *filter.ParentID)
		}
		if filter.Type != nil {
			query += " AND brokle_span_type = ?" // Use materialized column
			args = append(args, *filter.Type)
		}
		if filter.SpanKind != nil {
			query += " AND span_kind = ?"
			args = append(args, *filter.SpanKind)
		}
		if filter.Model != nil {
			query += " AND model_name = ?"
			args = append(args, *filter.Model)
		}
		if filter.Level != nil {
			query += " AND level = ?"
			args = append(args, *filter.Level)
		}
		if filter.StartTime != nil {
			query += " AND start_time >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND start_time <= ?"
			args = append(args, *filter.EndTime)
		}
		if filter.MinLatencyMs != nil {
			query += " AND duration_ms >= ?"
			args = append(args, *filter.MinLatencyMs)
		}
		if filter.MaxLatencyMs != nil {
			query += " AND duration_ms <= ?"
			args = append(args, *filter.MaxLatencyMs)
		}
		if filter.MinCost != nil {
			query += " AND total_cost >= ?"
			args = append(args, *filter.MinCost)
		}
		if filter.MaxCost != nil {
			query += " AND total_cost <= ?"
			args = append(args, *filter.MaxCost)
		}
		if filter.IsCompleted != nil {
			if *filter.IsCompleted {
				query += " AND end_time IS NOT NULL"
			} else {
				query += " AND end_time IS NULL"
			}
		}
	}

	// Order by start_time descending
	query += " ORDER BY start_time DESC"

	// Apply limit and offset
	if filter != nil {
		if filter.Limit > 0 {
			query += " LIMIT ?"
			args = append(args, filter.Limit)
		}
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query spans by filter: %w", err)
	}
	defer rows.Close()

	return r.scanSpans(rows)
}

// CreateBatch inserts multiple spans in a single batch
func (r *spanRepository) CreateBatch(ctx context.Context, spans []*observability.Span) error {
	if len(spans) == 0 {
		return nil
	}

	batch, err := r.db.PrepareBatch(ctx, `
		INSERT INTO spans (
			span_id, trace_id, parent_span_id, project_id,
			span_name, span_kind, start_time, end_time, duration_ms,
			status_code, status_message,
			span_attributes, resource_attributes,
			input, output,
			events_timestamp, events_name, events_attributes,
			links_trace_id, links_span_id, links_attributes,
			created_at, updated_at
		)
	`)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, span := range spans {
		// Set timestamp defaults
		if span.UpdatedAt.IsZero() {
			span.UpdatedAt = time.Now()
		}

		// Calculate duration if not set
		span.CalculateDuration()

		err = batch.Append(
			span.SpanID, // Renamed from ID
			span.TraceID,
			span.ParentSpanID,
			span.ProjectID,
			span.SpanName,   // Renamed from Name
			span.SpanKind,   // Now UInt8
			span.StartTime,
			span.EndTime,
			span.DurationMs,
			span.StatusCode, // Now UInt8
			span.StatusMessage,
			span.SpanAttributes,     // All attributes JSON
			span.ResourceAttributes, // Resource attributes JSON
			span.Input,
			span.Output,
			span.EventsTimestamp,  // OTEL Events
			span.EventsName,
			span.EventsAttributes,
			span.LinksTraceID, // OTEL Links
			span.LinksSpanID,
			span.LinksAttributes,
			span.CreatedAt,
			span.UpdatedAt,
			// Note: Materialized columns NOT inserted (computed by ClickHouse)
		)
		if err != nil {
			return fmt.Errorf("append to batch: %w", err)
		}
	}

	return batch.Send()
}

// Count returns the count of spans matching the filter
func (r *spanRepository) Count(ctx context.Context, filter *observability.SpanFilter) (int64, error) {
	query := "SELECT count() FROM spans"
	args := []interface{}{}

	if filter != nil {
		if filter.TraceID != nil {
			query += " AND trace_id = ?"
			args = append(args, *filter.TraceID)
		}
		if filter.Type != nil {
			query += " AND brokle_span_type = ?" // Use materialized column
			args = append(args, *filter.Type)
		}
		if filter.StartTime != nil {
			query += " AND start_time >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND start_time <= ?"
			args = append(args, *filter.EndTime)
		}
	}

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

// Helper function to scan a single span from query row
func (r *spanRepository) scanSpanRow(row driver.Row) (*observability.Span, error) {
	var span observability.Span

	err := row.Scan(
		&span.SpanID, // Renamed from ID
		&span.TraceID,
		&span.ParentSpanID,
		&span.ProjectID,
		&span.SpanName,   // Renamed from Name
		&span.SpanKind,   // Now UInt8
		&span.StartTime,
		&span.EndTime,
		&span.DurationMs,
		&span.StatusCode, // Now UInt8
		&span.StatusMessage,
		&span.SpanAttributes,     // All attributes JSON
		&span.ResourceAttributes, // Resource attributes JSON
		&span.Input,
		&span.Output,
		&span.EventsTimestamp,  // OTEL Events
		&span.EventsName,
		&span.EventsAttributes,
		&span.LinksTraceID, // OTEL Links
		&span.LinksSpanID,
		&span.LinksAttributes,
		&span.CreatedAt,
		&span.UpdatedAt,
		// Materialized columns (read-only, computed by ClickHouse)
		&span.GenAIOperationName,
		&span.GenAIProviderName,
		&span.GenAIRequestModel,
		&span.GenAIRequestMaxTokens,
		&span.GenAIRequestTemperature,
		&span.GenAIRequestTopP,
		&span.GenAIUsageInputTokens,
		&span.GenAIUsageOutputTokens,
		&span.BrokleSpanType,
		&span.BrokleSpanLevel,
		&span.BrokleCostInput,
		&span.BrokleCostOutput,
		&span.BrokleCostTotal,
		&span.BroklePromptID,
		&span.BroklePromptName,
		&span.BroklePromptVersion,
		&span.BrokleInternalModelID,
	)

	if err != nil {
		return nil, fmt.Errorf("scan span: %w", err)
	}

	return &span, nil
}

// Helper function to scan spans from query rows
func (r *spanRepository) scanSpans(rows driver.Rows) ([]*observability.Span, error) {
	var spans []*observability.Span

	for rows.Next() {
		var span observability.Span

		err := rows.Scan(
			&span.SpanID, // Renamed from ID
			&span.TraceID,
			&span.ParentSpanID,
			&span.ProjectID,
			&span.SpanName,   // Renamed from Name
			&span.SpanKind,   // Now UInt8
			&span.StartTime,
			&span.EndTime,
			&span.DurationMs,
			&span.StatusCode, // Now UInt8
			&span.StatusMessage,
			&span.SpanAttributes,     // All attributes JSON
			&span.ResourceAttributes, // Resource attributes JSON
			&span.Input,
			&span.Output,
			&span.EventsTimestamp,  // OTEL Events
			&span.EventsName,
			&span.EventsAttributes,
			&span.LinksTraceID, // OTEL Links
			&span.LinksSpanID,
			&span.LinksAttributes,
			&span.CreatedAt,
			&span.UpdatedAt,
			// Materialized columns
			&span.GenAIOperationName,
			&span.GenAIProviderName,
			&span.GenAIRequestModel,
			&span.GenAIRequestMaxTokens,
			&span.GenAIRequestTemperature,
			&span.GenAIRequestTopP,
			&span.GenAIUsageInputTokens,
			&span.GenAIUsageOutputTokens,
			&span.BrokleSpanType,
			&span.BrokleSpanLevel,
			&span.BrokleCostInput,
			&span.BrokleCostOutput,
			&span.BrokleCostTotal,
			&span.BroklePromptID,
			&span.BroklePromptName,
			&span.BroklePromptVersion,
			&span.BrokleInternalModelID,
		)

		if err != nil {
			return nil, fmt.Errorf("scan span: %w", err)
		}

		spans = append(spans, &span)
	}

	return spans, rows.Err()
}
