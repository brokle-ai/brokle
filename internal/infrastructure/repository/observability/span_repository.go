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

// Create inserts a new OTEL span (span) into ClickHouse
func (r *spanRepository) Create(ctx context.Context, span *observability.Span) error {
	// Set event_ts for ReplacingMergeTree deduplication
	span.EventTs = time.Now()
	span.UpdatedAt = time.Now()

	// Calculate duration if not set
	span.CalculateDuration()

	query := `
		INSERT INTO spans (
			id, trace_id, parent_span_id, project_id,
			name, span_kind, type, start_time, end_time, duration_ms,
			status_code, status_message,
			attributes, input, output, metadata, level,
			model_name, provider, internal_model_id, model_parameters,
			provided_usage_details, usage_details,
			provided_cost_details, cost_details, total_cost,
			prompt_id, prompt_name, prompt_version,
			created_at, updated_at,
			version, event_ts, is_deleted
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.db.Exec(ctx, query,
		span.ID,
		span.TraceID,
		span.ParentSpanID,
		span.ProjectID,
		span.Name,
		span.SpanKind,
		span.Type,
		span.StartTime,
		span.EndTime,
		span.DurationMs,
		span.StatusCode,
		span.StatusMessage,
		span.Attributes,
		span.Input,
		span.Output,
		span.Metadata,
		span.Level,
		span.ModelName,
		span.Provider,
		span.InternalModelID,
		span.ModelParameters,
		span.ProvidedUsageDetails,
		span.UsageDetails,
		span.ProvidedCostDetails,
		span.CostDetails,
		span.TotalCost,
		span.PromptID,
		span.PromptName,
		span.PromptVersion,
		span.CreatedAt,
		span.UpdatedAt,
		span.Version,
		span.EventTs,
		boolToUint8(span.IsDeleted),
	)
}

// Update performs an update using ReplacingMergeTree pattern (insert with higher version)
func (r *spanRepository) Update(ctx context.Context, span *observability.Span) error {
	// ReplacingMergeTree pattern: increment version and update event_ts
	span.EventTs = time.Now()
	span.UpdatedAt = time.Now()

	// Calculate duration if not set
	span.CalculateDuration()

	// Same INSERT query as Create - ClickHouse will handle merging
	return r.Create(ctx, span)
}

// Delete performs soft deletion by inserting a record with is_deleted = true
func (r *spanRepository) Delete(ctx context.Context, id string) error {
	query := `
		INSERT INTO spans
		SELECT
			id, trace_id, parent_span_id, project_id,
			name, span_kind, type, start_time, end_time, duration_ms,
			status_code, status_message,
			attributes, input, output, metadata, level,
			model_name, provider, internal_model_id, model_parameters,
			provided_usage_details, usage_details,
			provided_cost_details, cost_details, total_cost,
			prompt_id, prompt_name, prompt_version,
			created_at, updated_at,
			version,
			now64() as event_ts,
			1 as is_deleted
		FROM spans
		WHERE id = ? AND is_deleted = 0
		ORDER BY event_ts DESC
		LIMIT 1
	`

	return r.db.Exec(ctx, query, id)
}

// GetByID retrieves a span by its OTEL span_id (returns latest version)
func (r *spanRepository) GetByID(ctx context.Context, id string) (*observability.Span, error) {
	query := `
		SELECT
			id, trace_id, parent_span_id, project_id,
			name, span_kind, type, start_time, end_time, duration_ms,
			status_code, status_message,
			attributes, input, output, metadata, level,
			model_name, provider, internal_model_id, model_parameters,
			provided_usage_details, usage_details,
			provided_cost_details, cost_details, total_cost,
			prompt_id, prompt_name, prompt_version,
			created_at, updated_at,
			version, event_ts, is_deleted
		FROM spans
		WHERE id = ? AND is_deleted = 0
		ORDER BY event_ts DESC
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, id)
	return r.scanSpanRow(row)
}

// GetByTraceID retrieves all spans for a trace
func (r *spanRepository) GetByTraceID(ctx context.Context, traceID string) ([]*observability.Span, error) {
	query := `
		SELECT
			id, trace_id, parent_span_id, project_id,
			name, span_kind, type, start_time, end_time, duration_ms,
			status_code, status_message,
			attributes, input, output, metadata, level,
			model_name, provider, internal_model_id, model_parameters,
			provided_usage_details, usage_details,
			provided_cost_details, cost_details, total_cost,
			prompt_id, prompt_name, prompt_version,
			created_at, updated_at,
			version, event_ts, is_deleted
		FROM spans
		WHERE trace_id = ? AND is_deleted = 0
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
			id, trace_id, parent_span_id, project_id,
			name, span_kind, type, start_time, end_time, duration_ms,
			status_code, status_message,
			attributes, input, output, metadata, level,
			model_name, provider, internal_model_id, model_parameters,
			provided_usage_details, usage_details,
			provided_cost_details, cost_details, total_cost,
			prompt_id, prompt_name, prompt_version,
			created_at, updated_at,
			version, event_ts, is_deleted
		FROM spans
		WHERE trace_id = ? AND parent_span_id IS NULL AND is_deleted = 0
		ORDER BY event_ts DESC
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, traceID)
	return r.scanSpanRow(row)
}

// GetChildren retrieves child spans of a parent span
func (r *spanRepository) GetChildren(ctx context.Context, parentSpanID string) ([]*observability.Span, error) {
	query := `
		SELECT
			id, trace_id, parent_span_id, project_id,
			name, span_kind, type, start_time, end_time, duration_ms,
			status_code, status_message,
			attributes, input, output, metadata, level,
			model_name, provider, internal_model_id, model_parameters,
			provided_usage_details, usage_details,
			provided_cost_details, cost_details, total_cost,
			prompt_id, prompt_name, prompt_version,
			created_at, updated_at,
			version, event_ts, is_deleted
		FROM spans
		WHERE parent_span_id = ? AND is_deleted = 0
		ORDER BY start_time ASC
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
			id, trace_id, parent_span_id, project_id,
			name, span_kind, type, start_time, end_time, duration_ms,
			status_code, status_message,
			attributes, input, output, metadata, level,
			model_name, provider, internal_model_id, model_parameters,
			provided_usage_details, usage_details,
			provided_cost_details, cost_details, total_cost,
			prompt_id, prompt_name, prompt_version,
			created_at, updated_at,
			version, event_ts, is_deleted
		FROM spans
		WHERE is_deleted = 0
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
			query += " AND type = ?"
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
			id, trace_id, parent_span_id, project_id,
			name, span_kind, type, start_time, end_time, duration_ms,
			status_code, status_message,
			attributes, input, output, metadata, level,
			model_name, provider, internal_model_id, model_parameters,
			provided_usage_details, usage_details,
			provided_cost_details, cost_details, total_cost,
			prompt_id, prompt_name, prompt_version,
			created_at, updated_at,
			version, event_ts, is_deleted
		)
	`)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, span := range spans {
		// Set event_ts for ReplacingMergeTree
		if span.EventTs.IsZero() {
			span.EventTs = time.Now()
		}
		if span.UpdatedAt.IsZero() {
			span.UpdatedAt = time.Now()
		}

		// Calculate duration if not set
		span.CalculateDuration()

		err = batch.Append(
			span.ID,
			span.TraceID,
			span.ParentSpanID,
			span.ProjectID,
			span.Name,
			span.SpanKind,
			span.Type,
			span.StartTime,
			span.EndTime,
			span.DurationMs,
			span.StatusCode,
			span.StatusMessage,
			span.Attributes,
			span.Input,
			span.Output,
			span.Metadata,
			span.Level,
			span.ModelName,
			span.Provider,
			span.InternalModelID,
			span.ModelParameters,
			span.ProvidedUsageDetails,
			span.UsageDetails,
			span.ProvidedCostDetails,
			span.CostDetails,
			span.TotalCost,
			span.PromptID,
			span.PromptName,
			span.PromptVersion,
			span.CreatedAt,
			span.UpdatedAt,
			span.Version,
			span.EventTs,
			boolToUint8(span.IsDeleted),
		)
		if err != nil {
			return fmt.Errorf("append to batch: %w", err)
		}
	}

	return batch.Send()
}

// Count returns the count of spans matching the filter
func (r *spanRepository) Count(ctx context.Context, filter *observability.SpanFilter) (int64, error) {
	query := "SELECT count() FROM spans WHERE is_deleted = 0"
	args := []interface{}{}

	if filter != nil {
		if filter.TraceID != nil {
			query += " AND trace_id = ?"
			args = append(args, *filter.TraceID)
		}
		if filter.Type != nil {
			query += " AND type = ?"
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
	var isDeleted uint8

	err := row.Scan(
		&span.ID,
		&span.TraceID,
		&span.ParentSpanID,
		&span.ProjectID,
		&span.Name,
		&span.SpanKind,
		&span.Type,
		&span.StartTime,
		&span.EndTime,
		&span.DurationMs,
		&span.StatusCode,
		&span.StatusMessage,
		&span.Attributes,
		&span.Input,
		&span.Output,
		&span.Metadata,
		&span.Level,
		&span.ModelName,
		&span.Provider,
		&span.InternalModelID,
		&span.ModelParameters,
		&span.ProvidedUsageDetails,
		&span.UsageDetails,
		&span.ProvidedCostDetails,
		&span.CostDetails,
		&span.TotalCost,
		&span.PromptID,
		&span.PromptName,
		&span.PromptVersion,
		&span.CreatedAt,
		&span.UpdatedAt,
		&span.Version,
		&span.EventTs,
		&isDeleted,
	)

	if err != nil {
		return nil, fmt.Errorf("sca span: %w", err)
	}

	span.IsDeleted = isDeleted != 0

	return &span, nil
}

// Helper function to scan spans from query rows
func (r *spanRepository) scanSpans(rows driver.Rows) ([]*observability.Span, error) {
	var spans []*observability.Span

	for rows.Next() {
		var span observability.Span
		var isDeleted uint8

		err := rows.Scan(
			&span.ID,
			&span.TraceID,
			&span.ParentSpanID,
			&span.ProjectID,
			&span.Name,
			&span.SpanKind,
			&span.Type,
			&span.StartTime,
			&span.EndTime,
			&span.DurationMs,
			&span.StatusCode,
			&span.StatusMessage,
			&span.Attributes,
			&span.Input,
			&span.Output,
			&span.Metadata,
			&span.Level,
			&span.ModelName,
			&span.Provider,
			&span.InternalModelID,
			&span.ModelParameters,
			&span.ProvidedUsageDetails,
			&span.UsageDetails,
			&span.ProvidedCostDetails,
			&span.CostDetails,
			&span.TotalCost,
			&span.PromptID,
			&span.PromptName,
			&span.PromptVersion,
			&span.CreatedAt,
			&span.UpdatedAt,
			&span.Version,
			&span.EventTs,
			&isDeleted,
		)

		if err != nil {
			return nil, fmt.Errorf("sca span: %w", err)
		}

		span.IsDeleted = isDeleted != 0

		spans = append(spans, &span)
	}

	return spans, rows.Err()
}
