package observability

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/pagination"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type spanRepository struct {
	db clickhouse.Conn
}

// spanSelectFields defines the SELECT clause for span queries (reused across all queries)
const spanSelectFields = `
	span_id, trace_id, parent_span_id, trace_state, project_id,
	span_name, span_kind, start_time, end_time, duration, completion_start_time,
	status_code, status_message, has_error,
	input, output,
	attributes, metadata,
	usage_details, cost_details, pricing_snapshot, total_cost,
	events_timestamp, events_name, events_attributes,
	links_trace_id, links_span_id, links_trace_state, links_attributes,
	version, deleted_at,
	created_at, updated_at,
	model_name, provider_name, span_type, level,
	service_name
`

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
			span_id, trace_id, parent_span_id, trace_state, project_id,
			span_name, span_kind, start_time, end_time, duration, completion_start_time,
			status_code, status_message,
			input, output,
			attributes, metadata,
			usage_details, cost_details, pricing_snapshot, total_cost,
			events_timestamp, events_name, events_attributes,
			links_trace_id, links_span_id, links_trace_state, links_attributes,
			deleted_at,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.db.Exec(ctx, query,
		span.SpanID,
		span.TraceID,
		span.ParentSpanID,
		span.TraceState,
		span.ProjectID,
		span.SpanName,
		span.SpanKind,
		span.StartTime,
		span.EndTime,
		span.Duration,
		span.CompletionStartTime,
		span.StatusCode,
		span.StatusMessage,
		span.Input,
		span.Output,
		span.Attributes,        // JSON type
		span.Metadata,          // JSON type
		span.UsageDetails,      // Map(String, UInt64)
		span.CostDetails,       // Map(String, Decimal)
		span.PricingSnapshot,   // Map(String, Decimal) - CRITICAL: audit trail
		span.TotalCost,         // Nullable(Decimal)
		span.EventsTimestamp,
		span.EventsName,
		span.EventsAttributes,  // Array(Map)
		span.LinksTraceID,
		span.LinksSpanID,
		span.LinksTraceState,
		span.LinksAttributes,   // Array(Map)
		// version, model_name, provider_name, span_type, level omitted - MATERIALIZED from attributes JSON
		span.DeletedAt,         // Soft delete
		span.CreatedAt,
		span.UpdatedAt,
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
	query := "SELECT " + spanSelectFields + " FROM spans WHERE span_id = ? AND deleted_at IS NULL LIMIT 1"

	row := r.db.QueryRow(ctx, query, id)
	return r.scanSpanRow(row)
}

// GetByTraceID retrieves all spans for a trace
func (r *spanRepository) GetByTraceID(ctx context.Context, traceID string) ([]*observability.Span, error) {
	query := "SELECT " + spanSelectFields + " FROM spans WHERE trace_id = ? AND deleted_at IS NULL ORDER BY start_time ASC"

	rows, err := r.db.Query(ctx, query, traceID)
	if err != nil {
		return nil, fmt.Errorf("query spans by trace: %w", err)
	}
	defer rows.Close()

	return r.scanSpans(rows)
}

// GetRootSpan retrieves the root span for a trace (parent_span_id IS NULL)
func (r *spanRepository) GetRootSpan(ctx context.Context, traceID string) (*observability.Span, error) {
	query := "SELECT " + spanSelectFields + " FROM spans WHERE trace_id = ? AND parent_span_id IS NULL AND deleted_at IS NULL ORDER BY start_time DESC LIMIT 1"
	row := r.db.QueryRow(ctx, query, traceID)
	return r.scanSpanRow(row)
}

// GetChildren retrieves child spans of a parent span
func (r *spanRepository) GetChildren(ctx context.Context, parentSpanID string) ([]*observability.Span, error) {
	query := "SELECT " + spanSelectFields + " FROM spans WHERE parent_span_id = ? AND deleted_at IS NULL ORDER BY start_time ASC"
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
		SELECT ` + spanSelectFields + `
		FROM spans
		WHERE 1=1
			AND deleted_at IS NULL
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
			query += " AND span_type = ?" // Use materialized column
			args = append(args, *filter.Type)
		}
		if filter.SpanKind != nil {
			query += " AND span_kind = ?"
			args = append(args, *filter.SpanKind)
		}
		if filter.Model != nil {
			query += " AND model_name = ?" // Use materialized column
			args = append(args, *filter.Model)
		}
		if filter.ServiceName != nil {
			query += " AND service_name = ?" // Use materialized column (5-10x faster than JSON extraction)
			args = append(args, *filter.ServiceName)
		}
		if filter.Level != nil {
			query += " AND level = ?" // Use materialized column
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
			// Convert milliseconds to nanoseconds for duration field
			query += " AND duration >= ?"
			args = append(args, uint64(*filter.MinLatencyMs)*1000000)
		}
		if filter.MaxLatencyMs != nil {
			// Convert milliseconds to nanoseconds for duration field
			query += " AND duration <= ?"
			args = append(args, uint64(*filter.MaxLatencyMs)*1000000)
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

	// Determine sort field and direction with SQL injection protection
	allowedSortFields := []string{"start_time", "end_time", "duration", "span_name", "level", "status_code", "created_at", "updated_at", "span_id"}
	sortField := "start_time" // default
	sortDir := "DESC"

	if filter != nil {
		// Validate sort field against whitelist
		if filter.Params.SortBy != "" {
			validated, err := pagination.ValidateSortField(filter.Params.SortBy, allowedSortFields)
			if err != nil {
				return nil, fmt.Errorf("invalid sort field: %w", err)
			}
			if validated != "" {
				sortField = validated
			}
		}
		if filter.Params.SortDir == "asc" {
			sortDir = "ASC"
		}
	}

	query += fmt.Sprintf(" ORDER BY %s %s, span_id %s", sortField, sortDir, sortDir)

	// Apply limit and offset for pagination
	limit := pagination.DefaultPageSize
	offset := 0
	if filter != nil {
		if filter.Params.Limit > 0 {
			limit = filter.Params.Limit
		}
		offset = filter.Params.GetOffset()
	}
	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

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
			span_id, trace_id, parent_span_id, trace_state, project_id,
			span_name, span_kind, start_time, end_time, duration, completion_start_time,
			status_code, status_message,
			input, output,
			attributes, metadata,
			usage_details, cost_details, pricing_snapshot, total_cost,
			events_timestamp, events_name, events_attributes,
			links_trace_id, links_span_id, links_trace_state, links_attributes,
			deleted_at,
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
			span.SpanID,
			span.TraceID,
			span.ParentSpanID,
			span.TraceState,
			span.ProjectID,
			span.SpanName,
			span.SpanKind,
			span.StartTime,
			span.EndTime,
			span.Duration,
			span.CompletionStartTime,
			span.StatusCode,
			span.StatusMessage,
			span.Input,
			span.Output,
			span.Attributes,        // JSON: All OTEL + Brokle attributes (source for materialized columns)
			span.Metadata,          // JSON: Resource attributes + scope
			span.UsageDetails,      // Map: Flexible token tracking
			span.CostDetails,       // Map: Flexible cost breakdown
			span.PricingSnapshot,   // Map: Audit trail
			span.TotalCost,         // Decimal: Pre-computed total
			span.EventsTimestamp,
			span.EventsName,
			span.EventsAttributes,
			span.LinksTraceID,
			span.LinksSpanID,
			span.LinksTraceState,
			span.LinksAttributes,
			// version, model_name, provider_name, span_type, level omitted - MATERIALIZED from attributes JSON
			span.DeletedAt,
			span.CreatedAt,
			span.UpdatedAt,
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
			query += " AND span_type = ?" // Use materialized column
			args = append(args, *filter.Type)
		}
		if filter.Level != nil {
			query += " AND level = ?" // Use materialized column
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
	}

	var count uint64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	return int64(count), err
}

// Helper function to scan a single span from query row
func (r *spanRepository) scanSpanRow(row driver.Row) (*observability.Span, error) {
	var span observability.Span

	err := row.Scan(
		&span.SpanID,
		&span.TraceID,
		&span.ParentSpanID,
		&span.TraceState,
		&span.ProjectID,
		&span.SpanName,
		&span.SpanKind,
		&span.StartTime,
		&span.EndTime,
		&span.Duration,
		&span.CompletionStartTime,
		&span.StatusCode,
		&span.StatusMessage,
		&span.HasError,
		&span.Input,
		&span.Output,
		&span.Attributes,           // JSON: All OTEL + Brokle attributes
		&span.Metadata,             // JSON: Resource attributes + scope
		&span.UsageDetails,         // Map: Flexible token tracking
		&span.CostDetails,          // Map: Flexible cost breakdown
		&span.PricingSnapshot,      // Map: Audit trail
		&span.TotalCost,            // Decimal: Pre-computed total
		&span.EventsTimestamp,
		&span.EventsName,
		&span.EventsAttributes,     // Array(Map)
		&span.LinksTraceID,
		&span.LinksSpanID,
		&span.LinksTraceState,
		&span.LinksAttributes,      // Array(Map)
		&span.Version,              // Materialized from attributes.brokle.span.version
		&span.DeletedAt,
		&span.CreatedAt,
		&span.UpdatedAt,
		&span.ModelName,            // Materialized from attributes (for filtering + API display)
		&span.ProviderName,         // Materialized from attributes (for filtering + API display)
		&span.SpanType,             // Materialized from attributes (for filtering + API display)
		&span.Level,                // Materialized from attributes (for filtering/sorting + API display)
		&span.ServiceName,          // Materialized from metadata.resourceAttributes.service.name (OTLP REQUIRED)
	)

	if err != nil {
		return nil, fmt.Errorf("scan span: %w", err)
	}

	return &span, nil
}

// Helper function to scan spans from query rows
func (r *spanRepository) scanSpans(rows driver.Rows) ([]*observability.Span, error) {
	spans := make([]*observability.Span, 0) // Initialize empty slice to return [] instead of nil

	for rows.Next() {
		var span observability.Span

		err := rows.Scan(
			&span.SpanID,
			&span.TraceID,
			&span.ParentSpanID,
			&span.TraceState,
			&span.ProjectID,
			&span.SpanName,
			&span.SpanKind,
			&span.StartTime,
			&span.EndTime,
			&span.Duration,
			&span.CompletionStartTime,
			&span.StatusCode,
			&span.StatusMessage,
			&span.HasError,
			&span.Input,
			&span.Output,
			&span.Attributes,           // JSON: All OTEL + Brokle attributes
			&span.Metadata,             // JSON: Resource attributes + scope
			&span.UsageDetails,         // Map: Flexible token tracking
			&span.CostDetails,          // Map: Flexible cost breakdown
			&span.PricingSnapshot,      // Map: Audit trail
			&span.TotalCost,            // Decimal: Pre-computed total
			&span.EventsTimestamp,
			&span.EventsName,
			&span.EventsAttributes,     // Array(Map)
			&span.LinksTraceID,
			&span.LinksSpanID,
			&span.LinksTraceState,
			&span.LinksAttributes,      // Array(Map)
			&span.Version,              // Materialized from attributes.brokle.span.version
			&span.DeletedAt,
			&span.CreatedAt,
			&span.UpdatedAt,
			&span.ModelName,            // Materialized from attributes (for filtering + API display)
			&span.ProviderName,         // Materialized from attributes (for filtering + API display)
			&span.SpanType,             // Materialized from attributes (for filtering + API display)
			&span.Level,                // Materialized from attributes (for filtering/sorting + API display)
			&span.ServiceName,          // Materialized from metadata.resourceAttributes.service.name (OTLP REQUIRED)
		)

		if err != nil {
			return nil, fmt.Errorf("scan span: %w", err)
		}

		spans = append(spans, &span)
	}

	return spans, rows.Err()
}
