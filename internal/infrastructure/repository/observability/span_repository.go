package observability

import (
	"context"
	"encoding/json"
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

// marshalJSON converts map[string]interface{} to JSON string for ClickHouse JSON columns
// ClickHouse JSON columns require string, []byte, or *clickhouse.JSON types
func marshalJSON(m map[string]interface{}) string {
	if m == nil || len(m) == 0 {
		return "{}"
	}
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		// Log error but return empty JSON to prevent batch failure
		return "{}"
	}
	return string(jsonBytes)
}

// spanSelectFields defines the SELECT clause for span queries (reused across all queries)
// Uses OTLP-standard naming (resource_attributes, span_attributes, scope fields)
const spanSelectFields = `
	span_id, trace_id, parent_span_id, trace_state, project_id,
	span_name, span_kind, start_time, end_time, duration_nano, completion_start_time,
	status_code, status_message, has_error,
	input, output,
	resource_attributes, span_attributes, scope_name, scope_version, scope_attributes,
	resource_schema_url, scope_schema_url,
	usage_details, cost_details, pricing_snapshot, total_cost,
	events_timestamp, events_name, events_attributes,
	links_trace_id, links_span_id, links_trace_state, links_attributes,
	brokle_version, deleted_at,
	model_name, provider_name, span_type, span_level,
	service_name
`

// ==================================
// Array Conversion Helpers
// ==================================

// convertEventsToArrays converts SpanEvent slice to exploded arrays for INSERT
func convertEventsToArrays(events []observability.SpanEvent) (
	timestamps []time.Time,
	names []string,
	attributes []map[string]string,
) {
	if len(events) == 0 {
		return []time.Time{}, []string{}, []map[string]string{}
	}

	timestamps = make([]time.Time, len(events))
	names = make([]string, len(events))
	attributes = make([]map[string]string, len(events))

	for i, event := range events {
		timestamps[i] = event.Timestamp
		names[i] = event.Name
		attrs := make(map[string]string)
		for k, v := range event.Attributes {
			attrs[k] = fmt.Sprint(v)
		}
		attributes[i] = attrs
	}
	return
}

// convertLinksToArrays converts SpanLink slice to exploded arrays for INSERT
func convertLinksToArrays(links []observability.SpanLink) (
	traceIDs []string,
	spanIDs []string,
	traceStates []string,
	attributes []map[string]string,
) {
	if len(links) == 0 {
		return []string{}, []string{}, []string{}, []map[string]string{}
	}

	traceIDs = make([]string, len(links))
	spanIDs = make([]string, len(links))
	traceStates = make([]string, len(links))
	attributes = make([]map[string]string, len(links))

	for i, link := range links {
		traceIDs[i] = link.TraceID
		spanIDs[i] = link.SpanID
		traceStates[i] = link.TraceState
		attrs := make(map[string]string)
		for k, v := range link.Attributes {
			attrs[k] = fmt.Sprint(v)
		}
		attributes[i] = attrs
	}
	return
}

// convertArraysToEvents converts exploded arrays back to SpanEvent slice for SELECT
func convertArraysToEvents(
	timestamps []time.Time,
	names []string,
	attributes []map[string]string,
) []observability.SpanEvent {
	if len(timestamps) == 0 {
		return nil
	}

	events := make([]observability.SpanEvent, len(timestamps))
	for i := range timestamps {
		// Attributes are already map[string]string from ClickHouse
		var attrs map[string]string
		if i < len(attributes) && attributes[i] != nil {
			attrs = attributes[i]
		} else {
			attrs = make(map[string]string)
		}
		events[i] = observability.SpanEvent{
			Timestamp:  timestamps[i],
			Name:       names[i],
			Attributes: attrs,
		}
	}
	return events
}

// convertArraysToLinks converts exploded arrays back to SpanLink slice for SELECT
func convertArraysToLinks(
	traceIDs []string,
	spanIDs []string,
	traceStates []string,
	attributes []map[string]string,
) []observability.SpanLink {
	if len(traceIDs) == 0 {
		return nil
	}

	links := make([]observability.SpanLink, len(traceIDs))
	for i := range traceIDs {
		// Attributes are already map[string]string from ClickHouse
		var attrs map[string]string
		if i < len(attributes) && attributes[i] != nil {
			attrs = attributes[i]
		} else {
			attrs = make(map[string]string)
		}
		links[i] = observability.SpanLink{
			TraceID:    traceIDs[i],
			SpanID:     spanIDs[i],
			TraceState: traceStates[i],
			Attributes: attrs,
		}
	}
	return links
}

// NewSpanRepository creates a new span repository instance
func NewSpanRepository(db clickhouse.Conn) observability.SpanRepository {
	return &spanRepository{db: db}
}

// Create inserts a new OTEL span into ClickHouse
func (r *spanRepository) Create(ctx context.Context, span *observability.Span) error {
	// Calculate duration if not set
	span.CalculateDuration()

	// Convert events and links to exploded arrays
	eventsTimestamps, eventsNames, eventsAttributes := convertEventsToArrays(span.Events)
	linksTraceIDs, linksSpanIDs, linksTraceStates, linksAttributes := convertLinksToArrays(span.Links)

	query := `
		INSERT INTO otel_traces (
			span_id, trace_id, parent_span_id, trace_state, project_id,
			span_name, span_kind, start_time, end_time, duration_nano, completion_start_time,
			status_code, status_message,
			input, output,
			resource_attributes, span_attributes, scope_name, scope_version, scope_attributes,
			resource_schema_url, scope_schema_url,
			usage_details, cost_details, pricing_snapshot, total_cost,
			events_timestamp, events_name, events_attributes,
			links_trace_id, links_span_id, links_trace_state, links_attributes,
			deleted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		span.Duration,              // Maps to duration_nano column
		span.CompletionStartTime,
		span.StatusCode,
		span.StatusMessage,
		span.Input,
		span.Output,
		span.ResourceAttributes, // Map(String, String) - direct insert
		span.SpanAttributes,     // Map(String, String) - direct insert
		span.ScopeName,
		span.ScopeVersion,
		span.ScopeAttributes,
		span.ResourceSchemaURL,
		span.ScopeSchemaURL,
		span.UsageDetails, // Map(String, UInt64)
		span.CostDetails,           // Map(String, Decimal)
		span.PricingSnapshot,       // Map(String, Decimal) - CRITICAL: audit trail
		span.TotalCost,             // Nullable(Decimal)
		eventsTimestamps,           // Array: events_timestamp
		eventsNames,                // Array: events_name
		eventsAttributes,           // Array: events_attributes
		linksTraceIDs,              // Array: links_trace_id
		linksSpanIDs,               // Array: links_span_id
		linksTraceStates,           // Array: links_trace_state
		linksAttributes,            // Array: links_attributes
		// version, model_name, provider_name, span_type, span_level omitted - MATERIALIZED from attributes JSON
		span.DeletedAt, // Soft delete
	)
}

// Update performs an update by inserting new data (MergeTree will handle deduplication)
func (r *spanRepository) Update(ctx context.Context, span *observability.Span) error {
	// Calculate duration if not set
	span.CalculateDuration()

	// MergeTree pattern: INSERT new version, ClickHouse merges based on ORDER BY
	return r.Create(ctx, span)
}

// Delete performs hard deletion (MergeTree supports lightweight deletes with DELETE mutation)
func (r *spanRepository) Delete(ctx context.Context, id string) error {
	// MergeTree lightweight DELETE (async mutation, eventually consistent)
	query := `ALTER TABLE otel_traces DELETE WHERE span_id = ?`
	return r.db.Exec(ctx, query, id)
}

// GetByID retrieves a span by its OTEL span_id
func (r *spanRepository) GetByID(ctx context.Context, id string) (*observability.Span, error) {
	query := "SELECT " + spanSelectFields + " FROM otel_traces WHERE span_id = ? AND deleted_at IS NULL LIMIT 1"

	row := r.db.QueryRow(ctx, query, id)
	return ScanSpanRow(row)
}

// GetByTraceID retrieves all spans for a trace
func (r *spanRepository) GetByTraceID(ctx context.Context, traceID string) ([]*observability.Span, error) {
	query := "SELECT " + spanSelectFields + " FROM otel_traces WHERE trace_id = ? AND deleted_at IS NULL ORDER BY start_time ASC"

	rows, err := r.db.Query(ctx, query, traceID)
	if err != nil {
		return nil, fmt.Errorf("query spans by trace: %w", err)
	}
	defer rows.Close()

	return r.scanSpans(rows)
}

// GetRootSpan retrieves the root span for a trace (parent_span_id IS NULL)
func (r *spanRepository) GetRootSpan(ctx context.Context, traceID string) (*observability.Span, error) {
	query := "SELECT " + spanSelectFields + " FROM otel_traces WHERE trace_id = ? AND parent_span_id IS NULL AND deleted_at IS NULL ORDER BY start_time DESC LIMIT 1"
	row := r.db.QueryRow(ctx, query, traceID)
	return ScanSpanRow(row)
}

// GetChildren retrieves child spans of a parent span
func (r *spanRepository) GetChildren(ctx context.Context, parentSpanID string) ([]*observability.Span, error) {
	query := "SELECT " + spanSelectFields + " FROM otel_traces WHERE parent_span_id = ? AND deleted_at IS NULL ORDER BY start_time ASC"
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
		FROM otel_traces
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
			query += " AND span_level = ?" // Use materialized column
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
			// Convert milliseconds to nanoseconds for duration_nano column
			query += " AND duration_nano >= ?"
			args = append(args, uint64(*filter.MinLatencyMs)*1000000)
		}
		if filter.MaxLatencyMs != nil {
			// Convert milliseconds to nanoseconds for duration_nano column
			query += " AND duration_nano <= ?"
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
	allowedSortFields := []string{"start_time", "end_time", "duration_nano", "span_name", "span_level", "status_code", "span_id"}
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
		INSERT INTO otel_traces (
			span_id, trace_id, parent_span_id, trace_state, project_id,
			span_name, span_kind, start_time, end_time, duration_nano, completion_start_time,
			status_code, status_message,
			input, output,
			resource_attributes, span_attributes, scope_name, scope_version, scope_attributes,
			resource_schema_url, scope_schema_url,
			usage_details, cost_details, pricing_snapshot, total_cost,
			events_timestamp, events_name, events_attributes,
			links_trace_id, links_span_id, links_trace_state, links_attributes,
			deleted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, span := range spans {
		// Calculate duration if not set
		span.CalculateDuration()

		// Convert events and links to exploded arrays
		eventsTimestamps, eventsNames, eventsAttributes := convertEventsToArrays(span.Events)
		linksTraceIDs, linksSpanIDs, linksTraceStates, linksAttributes := convertLinksToArrays(span.Links)

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
			span.Duration,              // Maps to duration_nano column
			span.CompletionStartTime,
			span.StatusCode,
			span.StatusMessage,
			span.Input,
			span.Output,
			span.ResourceAttributes, // Map: OTLP resource attributes
			span.SpanAttributes,     // Map: OTLP span attributes
			span.ScopeName,
			span.ScopeVersion,
			span.ScopeAttributes,
			span.ResourceSchemaURL,
			span.ScopeSchemaURL,
			span.UsageDetails, // Map: Flexible token tracking
			span.CostDetails,           // Map: Flexible cost breakdown
			span.PricingSnapshot,       // Map: Audit trail
			span.TotalCost,             // Decimal: Pre-computed total
			eventsTimestamps,           // Array: events_timestamp
			eventsNames,                // Array: events_name
			eventsAttributes,           // Array: events_attributes
			linksTraceIDs,              // Array: links_trace_id
			linksSpanIDs,               // Array: links_span_id
			linksTraceStates,           // Array: links_trace_state
			linksAttributes,            // Array: links_attributes
			// version, model_name, provider_name, span_type, span_level omitted - MATERIALIZED from attributes JSON
			span.DeletedAt,
		)
		if err != nil {
			return fmt.Errorf("append to batch: %w", err)
		}
	}

	return batch.Send()
}

// Count returns the count of spans matching the filter
func (r *spanRepository) Count(ctx context.Context, filter *observability.SpanFilter) (int64, error) {
	query := "SELECT count() FROM otel_traces WHERE 1=1"
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
			query += " AND span_level = ?" // Use materialized column
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
// ScanSpanRow scans a single span row from the database
// Exported for use by trace_repository.go when querying root spans
func ScanSpanRow(row driver.Row) (*observability.Span, error) {
	var span observability.Span

	// Intermediate arrays for events (exploded columns)
	var eventsTimestamps []time.Time
	var eventsNames []string
	var eventsAttributes []map[string]string

	// Intermediate arrays for links (exploded columns)
	var linksTraceIDs []string
	var linksSpanIDs []string
	var linksTraceStates []string
	var linksAttributes []map[string]string

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
		&span.Duration,             // Now from duration_nano column
		&span.CompletionStartTime,
		&span.StatusCode,
		&span.StatusMessage,
		&span.HasError,
		&span.Input,
		&span.Output,
		&span.ResourceAttributes, // Map: OTLP resource attributes
		&span.SpanAttributes,     // Map: OTLP span attributes
		&span.ScopeName,
		&span.ScopeVersion,
		&span.ScopeAttributes,
		&span.ResourceSchemaURL,
		&span.ScopeSchemaURL,
		&span.UsageDetails, // Map: Flexible token tracking
		&span.CostDetails,          // Map: Flexible cost breakdown
		&span.PricingSnapshot,      // Map: Audit trail
		&span.TotalCost,            // Decimal: Pre-computed total
		&eventsTimestamps,          // Array: events_timestamp
		&eventsNames,               // Array: events_name
		&eventsAttributes,          // Array: events_attributes
		&linksTraceIDs,             // Array: links_trace_id
		&linksSpanIDs,              // Array: links_span_id
		&linksTraceStates,          // Array: links_trace_state
		&linksAttributes,           // Array: links_attributes
		&span.Version,              // Materialized from attributes.brokle.span.version
		&span.DeletedAt,
		&span.ModelName,            // Materialized from attributes (for filtering + API display)
		&span.ProviderName,         // Materialized from attributes (for filtering + API display)
		&span.SpanType,             // Materialized from attributes (for filtering + API display)
		&span.Level,                // Materialized from attributes (for filtering/sorting + API display)
		&span.ServiceName,          // Materialized from metadata.resourceAttributes.service.name (OTLP REQUIRED)
	)

	if err != nil {
		return nil, fmt.Errorf("scan span: %w", err)
	}

	// Convert arrays to domain types
	span.Events = convertArraysToEvents(eventsTimestamps, eventsNames, eventsAttributes)
	span.Links = convertArraysToLinks(linksTraceIDs, linksSpanIDs, linksTraceStates, linksAttributes)

	return &span, nil
}

// Helper function to scan spans from query rows
func (r *spanRepository) scanSpans(rows driver.Rows) ([]*observability.Span, error) {
	spans := make([]*observability.Span, 0) // Initialize empty slice to return [] instead of nil

	for rows.Next() {
		var span observability.Span

		// Intermediate arrays for events (exploded columns)
		var eventsTimestamps []time.Time
		var eventsNames []string
		var eventsAttributes []map[string]string

		// Intermediate arrays for links (exploded columns)
		var linksTraceIDs []string
		var linksSpanIDs []string
		var linksTraceStates []string
		var linksAttributes []map[string]string

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
			&span.Duration,             // Now from duration_nano column
			&span.CompletionStartTime,
			&span.StatusCode,
			&span.StatusMessage,
			&span.HasError,
			&span.Input,
			&span.Output,
			&span.ResourceAttributes, // Map: OTLP resource attributes
			&span.SpanAttributes,     // Map: OTLP span attributes
			&span.ScopeName,
			&span.ScopeVersion,
			&span.ScopeAttributes,
			&span.ResourceSchemaURL,
			&span.ScopeSchemaURL,
			&span.UsageDetails, // Map: Flexible token tracking
			&span.CostDetails,          // Map: Flexible cost breakdown
			&span.PricingSnapshot,      // Map: Audit trail
			&span.TotalCost,            // Decimal: Pre-computed total
			&eventsTimestamps,          // Array: events_timestamp
			&eventsNames,               // Array: events_name
			&eventsAttributes,          // Array: events_attributes
			&linksTraceIDs,             // Array: links_trace_id
			&linksSpanIDs,              // Array: links_span_id
			&linksTraceStates,          // Array: links_trace_state
			&linksAttributes,           // Array: links_attributes
			&span.Version,              // Materialized from attributes.brokle.span.version
			&span.DeletedAt,
			&span.ModelName,            // Materialized from attributes (for filtering + API display)
			&span.ProviderName,         // Materialized from attributes (for filtering + API display)
			&span.SpanType,             // Materialized from attributes (for filtering + API display)
			&span.Level,                // Materialized from attributes (for filtering/sorting + API display)
			&span.ServiceName,          // Materialized from metadata.resourceAttributes.service.name (OTLP REQUIRED)
		)

		if err != nil {
			return nil, fmt.Errorf("scan span: %w", err)
		}

		// Convert arrays to domain types
		span.Events = convertArraysToEvents(eventsTimestamps, eventsNames, eventsAttributes)
		span.Links = convertArraysToLinks(linksTraceIDs, linksSpanIDs, linksTraceStates, linksAttributes)

		spans = append(spans, &span)
	}

	return spans, rows.Err()
}
