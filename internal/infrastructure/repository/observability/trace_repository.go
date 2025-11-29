package observability

import (
	"context"
	"fmt"
	"strings"

	"brokle/internal/core/domain/observability"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/shopspring/decimal"
)

type traceRepository struct {
	db clickhouse.Conn
}

// NewTraceRepository creates a new trace repository instance
func NewTraceRepository(db clickhouse.Conn) observability.TraceRepository {
	return &traceRepository{db: db}
}

// GetRootSpan retrieves the root span for a trace (parent_span_id IS NULL)
// OTEL-Native: Root spans represent traces in OTLP
func (r *traceRepository) GetRootSpan(ctx context.Context, traceID string) (*observability.Span, error) {
	// Reuse span repository's select fields and scan logic
	query := `
		SELECT ` + spanSelectFields + `
		FROM otel_traces
		WHERE trace_id = ?
		  AND parent_span_id IS NULL
		  AND deleted_at IS NULL
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, traceID)
	return ScanSpanRow(row)
}

// GetTraceMetrics calculates aggregated trace-level metrics on-demand
// Uses GROUP BY query for real-time aggregation
func (r *traceRepository) GetTraceMetrics(ctx context.Context, traceID string) (*observability.TraceMetrics, error) {
	query := `
		SELECT
			trace_id,
			anyIf(span_id, parent_span_id IS NULL) as root_span_id,
			anyIf(project_id, parent_span_id IS NULL) as project_id,
			min(start_time) as trace_start,
			max(end_time) as trace_end,
			max(end_time) - min(start_time) as trace_duration_nano,

			-- Cost and usage aggregations
			sum(total_cost) as total_cost,
			sum(usage_details['input']) as total_input_tokens,
			sum(usage_details['output']) as total_output_tokens,
			sum(usage_details['total']) as total_tokens,

			-- Span metrics
			count() as span_count,
			countIf(has_error = true) as error_span_count,
			max(has_error) as has_error,

			-- Root span metadata (materialized columns for fast access)
			anyIf(service_name, parent_span_id IS NULL) as service_name,
			anyIf(model_name, parent_span_id IS NULL) as model_name,
			anyIf(provider_name, parent_span_id IS NULL) as provider_name,
			anyIf(span_attributes['user.id'], parent_span_id IS NULL) as user_id,
			anyIf(span_attributes['session.id'], parent_span_id IS NULL) as session_id
		FROM otel_traces
		WHERE trace_id = ?
		  AND deleted_at IS NULL
		GROUP BY trace_id
	`

	row := r.db.QueryRow(ctx, query, traceID)

	var metrics observability.TraceMetrics
	var totalCostFloat float64

	err := row.Scan(
		&metrics.TraceID,
		&metrics.RootSpanID,
		&metrics.ProjectID,
		&metrics.StartTime,
		&metrics.EndTime,
		&metrics.Duration,
		&totalCostFloat,
		&metrics.InputTokens,
		&metrics.OutputTokens,
		&metrics.TotalTokens,
		&metrics.SpanCount,
		&metrics.ErrorSpanCount,
		&metrics.HasError,
		&metrics.ServiceName,
		&metrics.ModelName,
		&metrics.ProviderName,
		&metrics.UserID,
		&metrics.SessionID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan trace metrics: %w", err)
	}

	// Convert float to decimal for cost
	metrics.TotalCost = decimal.NewFromFloat(totalCostFloat)

	return &metrics, nil
}

// CalculateTotalCost calculates the total cost for a trace
func (r *traceRepository) CalculateTotalCost(ctx context.Context, traceID string) (float64, error) {
	query := `
		SELECT sum(total_cost) as total
		FROM otel_traces
		WHERE trace_id = ?
		  AND deleted_at IS NULL
	`

	var total float64
	err := r.db.QueryRow(ctx, query, traceID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total cost: %w", err)
	}

	return total, nil
}

// CountSpans counts the number of spans in a trace
func (r *traceRepository) CountSpans(ctx context.Context, traceID string) (int64, error) {
	query := `
		SELECT count() as span_count
		FROM otel_traces
		WHERE trace_id = ?
		  AND deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query, traceID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count spans: %w", err)
	}

	return count, nil
}

// ListTraces queries traces with filters and returns aggregated trace metrics
// Uses GROUP BY to aggregate across all spans in each trace (matching GetTraceMetrics pattern)
// IMPORTANT: Attribute filters (user_id, session_id, service_name) are applied via HAVING clause
// AFTER aggregation to ensure full trace metrics are calculated correctly.
func (r *traceRepository) ListTraces(ctx context.Context, filter *observability.TraceFilter) ([]*observability.TraceMetrics, error) {
	query := `
		SELECT
			trace_id,
			anyIf(span_id, parent_span_id IS NULL) as root_span_id,
			anyIf(project_id, parent_span_id IS NULL) as project_id,
			min(start_time) as trace_start,
			max(end_time) as trace_end,
			max(end_time) - min(start_time) as trace_duration_nano,

			-- Aggregated cost and usage across all spans
			sum(total_cost) as total_cost,
			sum(usage_details['input']) as input_tokens,
			sum(usage_details['output']) as output_tokens,
			sum(usage_details['total']) as total_tokens,

			-- Aggregated span metrics
			count() as span_count,
			countIf(has_error = true) as error_span_count,
			max(has_error) as has_error,

			-- Root span metadata (use anyIf to get from root span)
			anyIf(service_name, parent_span_id IS NULL) as service_name,
			anyIf(model_name, parent_span_id IS NULL) as model_name,
			anyIf(provider_name, parent_span_id IS NULL) as provider_name,
			anyIf(span_attributes['user.id'], parent_span_id IS NULL) as user_id,
			anyIf(span_attributes['session.id'], parent_span_id IS NULL) as session_id
		FROM otel_traces
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	havingClauses := []string{}
	havingArgs := []interface{}{}

	// Apply filters
	// WHERE clause: span-level filters that should apply BEFORE aggregation
	// HAVING clause: trace-level filters that should apply AFTER aggregation
	if filter != nil {
		// WHERE clause filters (safe before GROUP BY - all spans have these)
		if filter.ProjectID != "" {
			query += " AND project_id = ?"
			args = append(args, filter.ProjectID)
		}
		if filter.StartTime != nil {
			query += " AND start_time >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND start_time <= ?"
			args = append(args, *filter.EndTime)
		}

		// HAVING clause filters (applied after aggregation to preserve full trace metrics)
		// These filter on root span attributes to avoid excluding root span from aggregation
		if filter.UserID != nil {
			havingClauses = append(havingClauses, "user_id = ?")
			havingArgs = append(havingArgs, *filter.UserID)
		}
		if filter.SessionID != nil {
			havingClauses = append(havingClauses, "session_id = ?")
			havingArgs = append(havingArgs, *filter.SessionID)
		}
		if filter.ServiceName != nil {
			havingClauses = append(havingClauses, "service_name = ?")
			havingArgs = append(havingArgs, *filter.ServiceName)
		}
		if filter.StatusCode != nil {
			// Filter traces where root span has this status code
			havingClauses = append(havingClauses, "anyIf(status_code, parent_span_id IS NULL) = ?")
			havingArgs = append(havingArgs, *filter.StatusCode)
		}
	}

	// Group by trace_id to aggregate all spans per trace
	query += " GROUP BY trace_id"

	// Add HAVING clause if needed (filters applied after aggregation)
	if len(havingClauses) > 0 {
		query += " HAVING " + strings.Join(havingClauses, " AND ")
		args = append(args, havingArgs...)
	}

	// Order by trace start time (use the aggregated min)
	query += " ORDER BY trace_start DESC"

	if filter != nil && filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)

		offset := filter.GetOffset()
		if offset > 0 {
			query += " OFFSET ?"
			args = append(args, offset)
		}
	} else {
		query += " LIMIT 100" // Default limit
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list traces: %w", err)
	}
	defer rows.Close()

	var traces []*observability.TraceMetrics
	for rows.Next() {
		var trace observability.TraceMetrics
		var totalCostFloat float64

		err := rows.Scan(
			&trace.TraceID,
			&trace.RootSpanID,
			&trace.ProjectID,
			&trace.StartTime,
			&trace.EndTime,
			&trace.Duration,
			&totalCostFloat,
			&trace.InputTokens,
			&trace.OutputTokens,
			&trace.TotalTokens,
			&trace.SpanCount,
			&trace.ErrorSpanCount,
			&trace.HasError,
			&trace.ServiceName,
			&trace.ModelName,
			&trace.ProviderName,
			&trace.UserID,
			&trace.SessionID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trace: %w", err)
		}

		trace.TotalCost = decimal.NewFromFloat(totalCostFloat)
		traces = append(traces, &trace)
	}

	return traces, nil
}

// GetBySessionID retrieves traces by session ID (virtual session analytics)
func (r *traceRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*observability.TraceMetrics, error) {
	filter := &observability.TraceFilter{
		SessionID: &sessionID,
	}
	filter.Limit = 1000 // Higher limit for session analytics
	return r.ListTraces(ctx, filter)
}

// GetByUserID retrieves traces by user ID
func (r *traceRepository) GetByUserID(ctx context.Context, userID string, filter *observability.TraceFilter) ([]*observability.TraceMetrics, error) {
	if filter == nil {
		filter = &observability.TraceFilter{}
	}
	filter.UserID = &userID
	return r.ListTraces(ctx, filter)
}

// Count counts traces matching the filter
// IMPORTANT: Uses the same filtering semantics as ListTraces (subquery + GROUP BY + HAVING)
// to ensure count matches the actual number of traces returned by ListTraces.
func (r *traceRepository) Count(ctx context.Context, filter *observability.TraceFilter) (int64, error) {
	// Build inner query with same logic as ListTraces
	innerQuery := `
		SELECT trace_id
		FROM otel_traces
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	havingClauses := []string{}
	havingArgs := []interface{}{}

	// Apply filters (mirrors ListTraces exactly)
	if filter != nil {
		// WHERE clause filters (safe before GROUP BY - all spans have these)
		if filter.ProjectID != "" {
			innerQuery += " AND project_id = ?"
			args = append(args, filter.ProjectID)
		}
		if filter.StartTime != nil {
			innerQuery += " AND start_time >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			innerQuery += " AND start_time <= ?"
			args = append(args, *filter.EndTime)
		}

		// HAVING clause filters (applied after aggregation - mirrors ListTraces)
		if filter.UserID != nil {
			havingClauses = append(havingClauses, "anyIf(span_attributes['user.id'], parent_span_id IS NULL) = ?")
			havingArgs = append(havingArgs, *filter.UserID)
		}
		if filter.SessionID != nil {
			havingClauses = append(havingClauses, "anyIf(span_attributes['session.id'], parent_span_id IS NULL) = ?")
			havingArgs = append(havingArgs, *filter.SessionID)
		}
		if filter.ServiceName != nil {
			havingClauses = append(havingClauses, "anyIf(service_name, parent_span_id IS NULL) = ?")
			havingArgs = append(havingArgs, *filter.ServiceName)
		}
		if filter.StatusCode != nil {
			havingClauses = append(havingClauses, "anyIf(status_code, parent_span_id IS NULL) = ?")
			havingArgs = append(havingArgs, *filter.StatusCode)
		}
	}

	innerQuery += " GROUP BY trace_id"

	// Add HAVING clause if needed
	if len(havingClauses) > 0 {
		innerQuery += " HAVING " + strings.Join(havingClauses, " AND ")
		args = append(args, havingArgs...)
	}

	// Wrap in count query
	query := "SELECT count() FROM (" + innerQuery + ")"

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count traces: %w", err)
	}

	return count, nil
}
