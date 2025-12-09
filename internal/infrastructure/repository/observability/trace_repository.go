package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/pagination"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/shopspring/decimal"
)

type traceRepository struct {
	db clickhouse.Conn
}

func NewTraceRepository(db clickhouse.Conn) observability.TraceRepository {
	return &traceRepository{db: db}
}

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

func ScanSpanRow(row driver.Row) (*observability.Span, error) {
	var span observability.Span

	var eventsTimestamps []time.Time
	var eventsNames []string
	var eventsAttributes []map[string]string

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
		&span.Duration,
		&span.CompletionStartTime,
		&span.StatusCode,
		&span.StatusMessage,
		&span.HasError,
		&span.Input,
		&span.Output,
		&span.ResourceAttributes,
		&span.SpanAttributes,
		&span.ScopeName,
		&span.ScopeVersion,
		&span.ScopeAttributes,
		&span.ResourceSchemaURL,
		&span.ScopeSchemaURL,
		&span.UsageDetails,
		&span.CostDetails,
		&span.PricingSnapshot,
		&span.TotalCost,
		&eventsTimestamps,
		&eventsNames,
		&eventsAttributes,
		&linksTraceIDs,
		&linksSpanIDs,
		&linksTraceStates,
		&linksAttributes,
		&span.Version,
		&span.DeletedAt,
		&span.ModelName,
		&span.ProviderName,
		&span.SpanType,
		&span.Level,
		&span.ServiceName,
	)

	if err != nil {
		return nil, fmt.Errorf("scan span: %w", err)
	}

	span.Events = convertArraysToEvents(eventsTimestamps, eventsNames, eventsAttributes)
	span.Links = convertArraysToLinks(linksTraceIDs, linksSpanIDs, linksTraceStates, linksAttributes)

	return &span, nil
}

func (r *traceRepository) scanSpans(rows driver.Rows) ([]*observability.Span, error) {
	spans := make([]*observability.Span, 0)

	for rows.Next() {
		var span observability.Span

		var eventsTimestamps []time.Time
		var eventsNames []string
		var eventsAttributes []map[string]string

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
			&span.Duration,
			&span.CompletionStartTime,
			&span.StatusCode,
			&span.StatusMessage,
			&span.HasError,
			&span.Input,
			&span.Output,
			&span.ResourceAttributes,
			&span.SpanAttributes,
			&span.ScopeName,
			&span.ScopeVersion,
			&span.ScopeAttributes,
			&span.ResourceSchemaURL,
			&span.ScopeSchemaURL,
			&span.UsageDetails,
			&span.CostDetails,
			&span.PricingSnapshot,
			&span.TotalCost,
			&eventsTimestamps,
			&eventsNames,
			&eventsAttributes,
			&linksTraceIDs,
			&linksSpanIDs,
			&linksTraceStates,
			&linksAttributes,
			&span.Version,
			&span.DeletedAt,
			&span.ModelName,
			&span.ProviderName,
			&span.SpanType,
			&span.Level,
			&span.ServiceName,
		)

		if err != nil {
			return nil, fmt.Errorf("scan span: %w", err)
		}

		span.Events = convertArraysToEvents(eventsTimestamps, eventsNames, eventsAttributes)
		span.Links = convertArraysToLinks(linksTraceIDs, linksSpanIDs, linksTraceStates, linksAttributes)

		spans = append(spans, &span)
	}

	return spans, rows.Err()
}

// OTLP spans are immutable - no Update method per OTEL specification
func (r *traceRepository) InsertSpan(ctx context.Context, span *observability.Span) error {
	span.CalculateDuration()

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
		span.Duration,
		span.CompletionStartTime,
		span.StatusCode,
		span.StatusMessage,
		span.Input,
		span.Output,
		span.ResourceAttributes,
		span.SpanAttributes,
		span.ScopeName,
		span.ScopeVersion,
		span.ScopeAttributes,
		span.ResourceSchemaURL,
		span.ScopeSchemaURL,
		span.UsageDetails,
		span.CostDetails,
		span.PricingSnapshot,
		span.TotalCost,
		eventsTimestamps,
		eventsNames,
		eventsAttributes,
		linksTraceIDs,
		linksSpanIDs,
		linksTraceStates,
		linksAttributes,
		span.DeletedAt,
	)
}

func (r *traceRepository) InsertSpanBatch(ctx context.Context, spans []*observability.Span) error {
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
		span.CalculateDuration()

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
			span.Duration,
			span.CompletionStartTime,
			span.StatusCode,
			span.StatusMessage,
			span.Input,
			span.Output,
			span.ResourceAttributes,
			span.SpanAttributes,
			span.ScopeName,
			span.ScopeVersion,
			span.ScopeAttributes,
			span.ResourceSchemaURL,
			span.ScopeSchemaURL,
			span.UsageDetails,
			span.CostDetails,
			span.PricingSnapshot,
			span.TotalCost,
			eventsTimestamps,
			eventsNames,
			eventsAttributes,
			linksTraceIDs,
			linksSpanIDs,
			linksTraceStates,
			linksAttributes,
			span.DeletedAt,
		)
		if err != nil {
			return fmt.Errorf("append to batch: %w", err)
		}
	}

	return batch.Send()
}

func (r *traceRepository) DeleteSpan(ctx context.Context, spanID string) error {
	query := `ALTER TABLE otel_traces DELETE WHERE span_id = ?`
	return r.db.Exec(ctx, query, spanID)
}

func (r *traceRepository) GetSpan(ctx context.Context, spanID string) (*observability.Span, error) {
	query := "SELECT " + spanSelectFields + " FROM otel_traces WHERE span_id = ? AND deleted_at IS NULL LIMIT 1"

	row := r.db.QueryRow(ctx, query, spanID)
	return ScanSpanRow(row)
}

func (r *traceRepository) GetSpansByTraceID(ctx context.Context, traceID string) ([]*observability.Span, error) {
	query := "SELECT " + spanSelectFields + " FROM otel_traces WHERE trace_id = ? AND deleted_at IS NULL ORDER BY start_time ASC"

	rows, err := r.db.Query(ctx, query, traceID)
	if err != nil {
		return nil, fmt.Errorf("query spans by trace: %w", err)
	}
	defer rows.Close()

	return r.scanSpans(rows)
}

func (r *traceRepository) GetSpanChildren(ctx context.Context, parentSpanID string) ([]*observability.Span, error) {
	query := "SELECT " + spanSelectFields + " FROM otel_traces WHERE parent_span_id = ? AND deleted_at IS NULL ORDER BY start_time ASC"
	rows, err := r.db.Query(ctx, query, parentSpanID)
	if err != nil {
		return nil, fmt.Errorf("query child spans: %w", err)
	}
	defer rows.Close()

	return r.scanSpans(rows)
}

func (r *traceRepository) GetSpanTree(ctx context.Context, traceID string) ([]*observability.Span, error) {
	return r.GetSpansByTraceID(ctx, traceID)
}

func (r *traceRepository) GetSpansByFilter(ctx context.Context, filter *observability.SpanFilter) ([]*observability.Span, error) {
	query := `
		SELECT ` + spanSelectFields + `
		FROM otel_traces
		WHERE 1=1
			AND deleted_at IS NULL
	`

	args := []interface{}{}

	if filter != nil {
		if filter.ProjectID != "" {
			query += " AND project_id = ?"
			args = append(args, filter.ProjectID)
		}
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
			query += " AND model_name = ?"
			args = append(args, *filter.Model)
		}
		if filter.ServiceName != nil {
			query += " AND service_name = ?"
			args = append(args, *filter.ServiceName)
		}
		if filter.Level != nil {
			query += " AND span_level = ?"
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
			query += " AND duration_nano >= ?"
			args = append(args, uint64(*filter.MinLatencyMs)*1000000)
		}
		if filter.MaxLatencyMs != nil {
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

	allowedSortFields := []string{"start_time", "end_time", "duration_nano", "span_name", "span_level", "status_code", "span_id"}
	sortField := "start_time" // default
	sortDir := "DESC"

	if filter != nil {
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

func (r *traceRepository) CountSpansByFilter(ctx context.Context, filter *observability.SpanFilter) (int64, error) {
	query := "SELECT count() FROM otel_traces WHERE 1=1 AND deleted_at IS NULL"
	args := []interface{}{}

	if filter != nil {
		if filter.ProjectID != "" {
			query += " AND project_id = ?"
			args = append(args, filter.ProjectID)
		}
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

func (r *traceRepository) GetRootSpan(ctx context.Context, traceID string) (*observability.Span, error) {
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

func (r *traceRepository) GetTraceSummary(ctx context.Context, traceID string) (*observability.TraceSummary, error) {
	query := `
		SELECT
			trace_id,
			anyIf(span_id, parent_span_id IS NULL) as root_span_id,
			anyIf(project_id, parent_span_id IS NULL) as root_project_id,
			anyIf(span_name, parent_span_id IS NULL) as root_span_name,
			min(start_time) as trace_start,
			maxOrNull(end_time) as trace_end,
			anyIf(duration_nano, parent_span_id IS NULL) as trace_duration_nano,

			-- Cost and usage aggregations
			toFloat64(sum(total_cost)) as total_cost,
			sum(usage_details['input']) as total_input_tokens,
			sum(usage_details['output']) as total_output_tokens,
			sum(usage_details['total']) as total_tokens,

			-- Span metrics
			toInt64(count()) as span_count,
			toInt64(countIf(has_error = true)) as error_span_count,
			max(has_error) as trace_has_error,
			anyIf(status_code, parent_span_id IS NULL) as root_status_code,

			-- Root span metadata (materialized columns for fast access)
			anyIf(service_name, parent_span_id IS NULL) as root_service_name,
			anyIf(model_name, parent_span_id IS NULL) as root_model_name,
			anyIf(provider_name, parent_span_id IS NULL) as root_provider_name,
			anyIf(span_attributes['user.id'], parent_span_id IS NULL) as root_user_id,
			anyIf(span_attributes['session.id'], parent_span_id IS NULL) as root_session_id
		FROM otel_traces
		WHERE trace_id = ?
		  AND deleted_at IS NULL
		GROUP BY trace_id
	`

	row := r.db.QueryRow(ctx, query, traceID)

	var summary observability.TraceSummary
	var totalCostFloat float64

	err := row.Scan(
		&summary.TraceID,
		&summary.RootSpanID,
		&summary.ProjectID,
		&summary.Name,
		&summary.StartTime,
		&summary.EndTime,
		&summary.Duration,
		&totalCostFloat,
		&summary.InputTokens,
		&summary.OutputTokens,
		&summary.TotalTokens,
		&summary.SpanCount,
		&summary.ErrorSpanCount,
		&summary.HasError,
		&summary.StatusCode,
		&summary.ServiceName,
		&summary.ModelName,
		&summary.ProviderName,
		&summary.UserID,
		&summary.SessionID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan trace summary: %w", err)
	}

	summary.TotalCost = decimal.NewFromFloat(totalCostFloat)

	return &summary, nil
}

// Attribute filters (user_id, session_id, service_name) use HAVING clause to preserve full trace metrics
func (r *traceRepository) ListTraces(ctx context.Context, filter *observability.TraceFilter) ([]*observability.TraceSummary, error) {
	query := `
		SELECT
			trace_id,
			anyIf(span_id, parent_span_id IS NULL) as root_span_id,
			anyIf(project_id, parent_span_id IS NULL) as root_project_id,
			anyIf(span_name, parent_span_id IS NULL) as root_span_name,
			min(start_time) as trace_start,
			maxOrNull(end_time) as trace_end,
			anyIf(duration_nano, parent_span_id IS NULL) as trace_duration_nano,

			-- Aggregated cost and usage across all spans
			toFloat64(sum(total_cost)) as total_cost,
			sum(usage_details['input']) as input_tokens,
			sum(usage_details['output']) as output_tokens,
			sum(usage_details['total']) as total_tokens,

			-- Aggregated span metrics
			toInt64(count()) as span_count,
			toInt64(countIf(has_error = true)) as error_span_count,
			max(has_error) as trace_has_error,
			anyIf(status_code, parent_span_id IS NULL) as root_status_code,

			-- Root span metadata (use anyIf to get from root span)
			anyIf(service_name, parent_span_id IS NULL) as root_service_name,
			anyIf(model_name, parent_span_id IS NULL) as root_model_name,
			anyIf(provider_name, parent_span_id IS NULL) as root_provider_name,
			anyIf(span_attributes['user.id'], parent_span_id IS NULL) as root_user_id,
			anyIf(span_attributes['session.id'], parent_span_id IS NULL) as root_session_id
		FROM otel_traces
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	havingClauses := []string{}
	havingArgs := []interface{}{}

	if filter != nil {
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

		if filter.UserID != nil {
			havingClauses = append(havingClauses, "root_user_id = ?")
			havingArgs = append(havingArgs, *filter.UserID)
		}
		if filter.SessionID != nil {
			havingClauses = append(havingClauses, "root_session_id = ?")
			havingArgs = append(havingArgs, *filter.SessionID)
		}
		if filter.ServiceName != nil {
			havingClauses = append(havingClauses, "root_service_name = ?")
			havingArgs = append(havingArgs, *filter.ServiceName)
		}
		if filter.StatusCode != nil {
			havingClauses = append(havingClauses, "anyIf(status_code, parent_span_id IS NULL) = ?")
			havingArgs = append(havingArgs, *filter.StatusCode)
		}

		// Advanced filters
		if filter.ModelName != nil {
			havingClauses = append(havingClauses, "root_model_name = ?")
			havingArgs = append(havingArgs, *filter.ModelName)
		}
		if filter.ProviderName != nil {
			havingClauses = append(havingClauses, "root_provider_name = ?")
			havingArgs = append(havingArgs, *filter.ProviderName)
		}
		if filter.MinCost != nil {
			havingClauses = append(havingClauses, "total_cost >= ?")
			havingArgs = append(havingArgs, *filter.MinCost)
		}
		if filter.MaxCost != nil {
			havingClauses = append(havingClauses, "total_cost <= ?")
			havingArgs = append(havingArgs, *filter.MaxCost)
		}
		if filter.MinTokens != nil {
			havingClauses = append(havingClauses, "total_tokens >= ?")
			havingArgs = append(havingArgs, *filter.MinTokens)
		}
		if filter.MaxTokens != nil {
			havingClauses = append(havingClauses, "total_tokens <= ?")
			havingArgs = append(havingArgs, *filter.MaxTokens)
		}
		if filter.MinDuration != nil {
			havingClauses = append(havingClauses, "trace_duration_nano >= ?")
			havingArgs = append(havingArgs, *filter.MinDuration)
		}
		if filter.MaxDuration != nil {
			havingClauses = append(havingClauses, "trace_duration_nano <= ?")
			havingArgs = append(havingArgs, *filter.MaxDuration)
		}
		if filter.HasError != nil && *filter.HasError {
			havingClauses = append(havingClauses, "trace_has_error = ?")
			havingArgs = append(havingArgs, true)
		}
	}

	query += " GROUP BY trace_id"

	if len(havingClauses) > 0 {
		query += " HAVING " + strings.Join(havingClauses, " AND ")
		args = append(args, havingArgs...)
	}

	allowedSortFields := []string{
		"trace_start", "trace_end", "trace_duration_nano",
		"total_cost", "input_tokens", "output_tokens", "total_tokens",
		"span_count", "error_span_count", "service_name", "model_name",
	}
	sortField := "trace_start"
	sortDir := "DESC"

	if filter != nil {
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

	query += fmt.Sprintf(" ORDER BY %s %s, trace_id %s", sortField, sortDir, sortDir)

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

	var traces []*observability.TraceSummary
	for rows.Next() {
		var trace observability.TraceSummary
		var totalCostFloat float64

		err := rows.Scan(
			&trace.TraceID,
			&trace.RootSpanID,
			&trace.ProjectID,
			&trace.Name,
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
			&trace.StatusCode,
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

func (r *traceRepository) CountTraces(ctx context.Context, filter *observability.TraceFilter) (int64, error) {
	innerQuery := `
		SELECT
			trace_id,
			toFloat64(sum(total_cost)) as total_cost,
			sum(usage_details['total']) as total_tokens,
			if(maxOrNull(end_time) IS NOT NULL, toUInt64(maxOrNull(end_time) - min(start_time)), NULL) as trace_duration_nano,
			max(has_error) as trace_has_error,
			anyIf(service_name, parent_span_id IS NULL) as root_service_name,
			anyIf(model_name, parent_span_id IS NULL) as root_model_name,
			anyIf(provider_name, parent_span_id IS NULL) as root_provider_name,
			anyIf(span_attributes['user.id'], parent_span_id IS NULL) as root_user_id,
			anyIf(span_attributes['session.id'], parent_span_id IS NULL) as root_session_id
		FROM otel_traces
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	havingClauses := []string{}
	havingArgs := []interface{}{}

	if filter != nil {
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

		if filter.UserID != nil {
			havingClauses = append(havingClauses, "root_user_id = ?")
			havingArgs = append(havingArgs, *filter.UserID)
		}
		if filter.SessionID != nil {
			havingClauses = append(havingClauses, "root_session_id = ?")
			havingArgs = append(havingArgs, *filter.SessionID)
		}
		if filter.ServiceName != nil {
			havingClauses = append(havingClauses, "root_service_name = ?")
			havingArgs = append(havingArgs, *filter.ServiceName)
		}
		if filter.StatusCode != nil {
			havingClauses = append(havingClauses, "anyIf(status_code, parent_span_id IS NULL) = ?")
			havingArgs = append(havingArgs, *filter.StatusCode)
		}

		// Advanced filters
		if filter.ModelName != nil {
			havingClauses = append(havingClauses, "root_model_name = ?")
			havingArgs = append(havingArgs, *filter.ModelName)
		}
		if filter.ProviderName != nil {
			havingClauses = append(havingClauses, "root_provider_name = ?")
			havingArgs = append(havingArgs, *filter.ProviderName)
		}
		if filter.MinCost != nil {
			havingClauses = append(havingClauses, "total_cost >= ?")
			havingArgs = append(havingArgs, *filter.MinCost)
		}
		if filter.MaxCost != nil {
			havingClauses = append(havingClauses, "total_cost <= ?")
			havingArgs = append(havingArgs, *filter.MaxCost)
		}
		if filter.MinTokens != nil {
			havingClauses = append(havingClauses, "total_tokens >= ?")
			havingArgs = append(havingArgs, *filter.MinTokens)
		}
		if filter.MaxTokens != nil {
			havingClauses = append(havingClauses, "total_tokens <= ?")
			havingArgs = append(havingArgs, *filter.MaxTokens)
		}
		if filter.MinDuration != nil {
			havingClauses = append(havingClauses, "trace_duration_nano >= ?")
			havingArgs = append(havingArgs, *filter.MinDuration)
		}
		if filter.MaxDuration != nil {
			havingClauses = append(havingClauses, "trace_duration_nano <= ?")
			havingArgs = append(havingArgs, *filter.MaxDuration)
		}
		if filter.HasError != nil && *filter.HasError {
			havingClauses = append(havingClauses, "trace_has_error = ?")
			havingArgs = append(havingArgs, true)
		}
	}

	innerQuery += " GROUP BY trace_id"

	if len(havingClauses) > 0 {
		innerQuery += " HAVING " + strings.Join(havingClauses, " AND ")
		args = append(args, havingArgs...)
	}

	query := "SELECT toInt64(count()) FROM (" + innerQuery + ")"

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count traces: %w", err)
	}

	return count, nil
}

func (r *traceRepository) CountSpansInTrace(ctx context.Context, traceID string) (int64, error) {
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

func (r *traceRepository) DeleteTrace(ctx context.Context, traceID string) error {
	query := `ALTER TABLE otel_traces DELETE WHERE trace_id = ?`
	return r.db.Exec(ctx, query, traceID)
}

func (r *traceRepository) GetTracesBySessionID(ctx context.Context, sessionID string) ([]*observability.TraceSummary, error) {
	filter := &observability.TraceFilter{
		SessionID: &sessionID,
	}
	filter.Limit = 1000 // Higher limit for session analytics
	return r.ListTraces(ctx, filter)
}

func (r *traceRepository) GetTracesByUserID(ctx context.Context, userID string, filter *observability.TraceFilter) ([]*observability.TraceSummary, error) {
	if filter == nil {
		filter = &observability.TraceFilter{}
	}
	filter.UserID = &userID
	return r.ListTraces(ctx, filter)
}

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

func (r *traceRepository) CalculateTotalTokens(ctx context.Context, traceID string) (uint64, error) {
	query := `
		SELECT sum(usage_details['total']) as total
		FROM otel_traces
		WHERE trace_id = ?
		  AND deleted_at IS NULL
	`

	var total uint64
	err := r.db.QueryRow(ctx, query, traceID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total tokens: %w", err)
	}

	return total, nil
}

// GetFilterOptions returns available filter values for the traces filter UI
// This queries distinct values from actual trace data for dropdowns and range sliders
func (r *traceRepository) GetFilterOptions(ctx context.Context, projectID string) (*observability.TraceFilterOptions, error) {
	// Query to get distinct values and ranges from trace data
	// We aggregate at trace level (GROUP BY trace_id) to get trace-level values
	query := `
		SELECT
			arrayDistinct(groupArray(root_model_name)) as models,
			arrayDistinct(groupArray(root_provider_name)) as providers,
			arrayDistinct(groupArray(root_service_name)) as services,
			arrayDistinct(groupArray(root_deployment_environment)) as environments,
			arrayDistinct(groupArray(root_user_id)) as users,
			arrayDistinct(groupArray(root_session_id)) as sessions,
			minOrNull(total_cost) as min_cost,
			maxOrNull(total_cost) as max_cost,
			minOrNull(total_tokens) as min_tokens,
			maxOrNull(total_tokens) as max_tokens,
			minOrNull(trace_duration_nano) as min_duration,
			maxOrNull(trace_duration_nano) as max_duration
		FROM (
			SELECT
				trace_id,
				anyIf(model_name, parent_span_id IS NULL) as root_model_name,
				anyIf(provider_name, parent_span_id IS NULL) as root_provider_name,
				anyIf(service_name, parent_span_id IS NULL) as root_service_name,
				anyIf(deployment_environment, parent_span_id IS NULL) as root_deployment_environment,
				anyIf(span_attributes['user.id'], parent_span_id IS NULL) as root_user_id,
				anyIf(span_attributes['session.id'], parent_span_id IS NULL) as root_session_id,
				toFloat64(sum(total_cost)) as total_cost,
				sum(usage_details['total']) as total_tokens,
				if(maxOrNull(end_time) IS NOT NULL, toUInt64(maxOrNull(end_time) - min(start_time)), NULL) as trace_duration_nano
			FROM otel_traces
			WHERE project_id = ? AND deleted_at IS NULL
			GROUP BY trace_id
		)
	`

	var (
		models       []string
		providers    []string
		services     []string
		environments []string
		users        []string
		sessions     []string
		minCost      *float64
		maxCost      *float64
		minTokens    *uint64
		maxTokens    *uint64
		minDuration  *uint64
		maxDuration  *uint64
	)

	err := r.db.QueryRow(ctx, query, projectID).Scan(
		&models,
		&providers,
		&services,
		&environments,
		&users,
		&sessions,
		&minCost,
		&maxCost,
		&minTokens,
		&maxTokens,
		&minDuration,
		&maxDuration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get filter options: %w", err)
	}

	// Filter out empty strings from arrays
	models = filterEmptyStrings(models)
	providers = filterEmptyStrings(providers)
	services = filterEmptyStrings(services)
	environments = filterEmptyStrings(environments)
	users = filterEmptyStrings(users)
	sessions = filterEmptyStrings(sessions)

	options := &observability.TraceFilterOptions{
		Models:       models,
		Providers:    providers,
		Services:     services,
		Environments: environments,
		Users:        users,
		Sessions:     sessions,
	}

	// Set cost range if we have data
	if minCost != nil && maxCost != nil {
		options.CostRange = &observability.Range{
			Min: *minCost,
			Max: *maxCost,
		}
	}

	// Set token range if we have data
	if minTokens != nil && maxTokens != nil {
		options.TokenRange = &observability.Range{
			Min: float64(*minTokens),
			Max: float64(*maxTokens),
		}
	}

	// Set duration range if we have data
	if minDuration != nil && maxDuration != nil {
		options.DurationRange = &observability.Range{
			Min: float64(*minDuration),
			Max: float64(*maxDuration),
		}
	}

	return options, nil
}

// filterEmptyStrings removes empty strings from slice
func filterEmptyStrings(slice []string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}
