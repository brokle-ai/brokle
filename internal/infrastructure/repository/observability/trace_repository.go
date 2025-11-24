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

type traceRepository struct {
	db clickhouse.Conn
}

// traceSelectFields defines the SELECT clause for trace queries (reused across all queries)
const traceSelectFields = `
	trace_id, project_id, name, user_id, session_id, version, release,
	tags, environment, metadata,
	start_time, end_time, duration,
	status_code, status_message,
	input, output,
	total_cost, total_tokens, span_count,
	bookmarked, public,
	service_name,
	created_at, updated_at, deleted_at
`

// NewTraceRepository creates a new trace repository instance
func NewTraceRepository(db clickhouse.Conn) observability.TraceRepository {
	return &traceRepository{db: db}
}

// Create inserts a new OTEL trace into ClickHouse
func (r *traceRepository) Create(ctx context.Context, trace *observability.Trace) error {
	// Set updated timestamp
	trace.UpdatedAt = time.Now()

	// Calculate duration if not set
	trace.CalculateDuration()

	query := `
		INSERT INTO traces (
			trace_id, project_id, name, user_id, session_id,
			tags, environment, metadata,
			start_time, end_time, duration,
			status_code, status_message,
			input, output,
			total_cost, total_tokens, span_count,
			bookmarked, public,
			created_at, updated_at, deleted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.db.Exec(ctx, query,
		trace.TraceID,
		trace.ProjectID,
		trace.Name,
		trace.UserID,
		trace.SessionID,
		// version and release omitted - MATERIALIZED from metadata JSON
		trace.Tags,
		trace.Environment,
		trace.Metadata,           // JSON: Contains brokle.version and brokle.release (auto-computed to materialized columns)
		trace.StartTime,
		trace.EndTime,
		trace.Duration,
		trace.StatusCode,
		trace.StatusMessage,
		trace.Input,
		trace.Output,
		trace.TotalCost,         // Pre-computed aggregation
		trace.TotalTokens,
		trace.SpanCount,
		trace.Bookmarked,
		trace.Public,
		trace.CreatedAt,
		trace.UpdatedAt,
		trace.DeletedAt,         // Soft delete
	)
}

// Update performs an update by inserting new data (MergeTree will handle deduplication)
func (r *traceRepository) Update(ctx context.Context, trace *observability.Trace) error {
	// Set updated timestamp
	trace.UpdatedAt = time.Now()

	// Calculate duration if not set
	trace.CalculateDuration()

	// MergeTree pattern: INSERT new version, ClickHouse merges based on ORDER BY (trace_id)
	return r.Create(ctx, trace)
}

// Delete performs hard deletion (MergeTree supports lightweight deletes with DELETE mutation)
func (r *traceRepository) Delete(ctx context.Context, id string) error {
	// MergeTree lightweight DELETE (async mutation, eventually consistent)
	query := `ALTER TABLE traces DELETE WHERE trace_id = ?`
	return r.db.Exec(ctx, query, id)
}

// GetByID retrieves a trace by its OTEL trace_id
func (r *traceRepository) GetByID(ctx context.Context, id string) (*observability.Trace, error) {
	query := "SELECT " + traceSelectFields + " FROM traces WHERE trace_id = ? AND deleted_at IS NULL LIMIT 1"
	row := r.db.QueryRow(ctx, query, id)
	return r.scanTraceRow(row)
}

// GetByProjectID retrieves traces by project ID with cursor-based pagination
func (r *traceRepository) GetByProjectID(ctx context.Context, projectID string, filter *observability.TraceFilter) ([]*observability.Trace, error) {
	query := "SELECT " + traceSelectFields + " FROM traces WHERE project_id = ? AND deleted_at IS NULL"
	args := []interface{}{projectID}

	// Apply filters
	if filter != nil {
		if filter.SessionID != nil {
			query += " AND session_id = ?"
			args = append(args, *filter.SessionID)
		}
		if filter.UserID != nil {
			query += " AND user_id = ?"
			args = append(args, *filter.UserID)
		}
		if filter.Environment != nil {
			query += " AND environment = ?"
			args = append(args, *filter.Environment)
		}
		if filter.ServiceName != nil {
			query += " AND service_name = ?" // Materialized column (5-10x faster than JSON extraction)
			args = append(args, *filter.ServiceName)
		}
		if filter.StatusCode != nil {
			query += " AND status_code = ?"
			args = append(args, *filter.StatusCode)
		}
		if filter.Bookmarked != nil {
			query += " AND bookmarked = ?"
			args = append(args, *filter.Bookmarked)
		}
		if filter.Public != nil {
			query += " AND public = ?"
			args = append(args, *filter.Public)
		}
		if filter.StartTime != nil {
			query += " AND start_time >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND start_time <= ?"
			args = append(args, *filter.EndTime)
		}
		if len(filter.Tags) > 0 {
			query += " AND hasAll(tags, ?)"
			args = append(args, filter.Tags)
		}

	}

	// Determine sort field and direction with SQL injection protection
	allowedSortFields := []string{"start_time", "end_time", "duration", "status_code", "created_at", "updated_at", "trace_id"}
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

	// Order by sort field and trace_id for stable ordering
	query += fmt.Sprintf(" ORDER BY %s %s, trace_id %s", sortField, sortDir, sortDir)

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
		return nil, fmt.Errorf("query traces by project: %w", err)
	}
	defer rows.Close()

	return r.scanTraces(rows)
}

// GetBySessionID retrieves all traces in a virtual session
func (r *traceRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*observability.Trace, error) {
	query := `
		SELECT ` + traceSelectFields + `
		FROM traces
		WHERE session_id = ?
			AND deleted_at IS NULL
		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("query traces by session: %w", err)
	}
	defer rows.Close()

	return r.scanTraces(rows)
}

// GetByUserID retrieves traces by user ID with cursor-based pagination
func (r *traceRepository) GetByUserID(ctx context.Context, userID string, filter *observability.TraceFilter) ([]*observability.Trace, error) {
	query := `
		SELECT ` + traceSelectFields + `
		FROM traces
		WHERE user_id = ?
			AND deleted_at IS NULL
	`

	args := []interface{}{userID}

	if filter != nil {
		if filter.StartTime != nil {
			query += " AND start_time >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND start_time <= ?"
			args = append(args, *filter.EndTime)
		}

	}

	// Determine sort field and direction with SQL injection protection
	allowedSortFields := []string{"start_time", "end_time", "duration", "status_code", "created_at", "updated_at", "trace_id"}
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

	query += fmt.Sprintf(" ORDER BY %s %s, trace_id %s", sortField, sortDir, sortDir)

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
		return nil, fmt.Errorf("query traces by user: %w", err)
	}
	defer rows.Close()

	return r.scanTraces(rows)
}

// GetWithSpans retrieves a trace with all its spans (requires join)
func (r *traceRepository) GetWithSpans(ctx context.Context, id string) (*observability.Trace, error) {
	// TODO: Implement after SpanRepository is created
	return r.GetByID(ctx, id)
}

// GetWithScores retrieves a trace with all its scores (requires join)
func (r *traceRepository) GetWithScores(ctx context.Context, id string) (*observability.Trace, error) {
	// TODO: Implement after ScoreRepository is created
	return r.GetByID(ctx, id)
}

// CreateBatch inserts multiple traces in a single batch
func (r *traceRepository) CreateBatch(ctx context.Context, traces []*observability.Trace) error {
	if len(traces) == 0 {
		return nil
	}

	batch, err := r.db.PrepareBatch(ctx, `
		INSERT INTO traces (
			trace_id, project_id, name, user_id, session_id,
			tags, environment, metadata,
			start_time, end_time, duration,
			status_code, status_message,
			input, output,
			total_cost, total_tokens, span_count,
			bookmarked, public,
			created_at, updated_at, deleted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, trace := range traces {
		// Set timestamp defaults
		if trace.UpdatedAt.IsZero() {
			trace.UpdatedAt = time.Now()
		}

		// Calculate duration if not set
		trace.CalculateDuration()

		err = batch.Append(
			trace.TraceID,
			trace.ProjectID,
			trace.Name,
			trace.UserID,
			trace.SessionID,
			trace.Tags,
			trace.Environment,
			trace.Metadata,     // JSON: Contains brokle.release and brokle.version (materialized columns auto-computed)
			trace.StartTime,
			trace.EndTime,
			trace.Duration,
			trace.StatusCode,
			trace.StatusMessage,
			trace.Input,
			trace.Output,
			trace.TotalCost,
			trace.TotalTokens,
			trace.SpanCount,
			trace.Bookmarked,
			trace.Public,
			trace.CreatedAt,
			trace.UpdatedAt,
			trace.DeletedAt,
		)
		if err != nil {
			return fmt.Errorf("append to batch: %w", err)
		}
	}

	return batch.Send()
}

// Count returns the count of traces matching the filter
func (r *traceRepository) Count(ctx context.Context, filter *observability.TraceFilter) (int64, error) {
	query := "SELECT count() FROM traces WHERE 1=1"
	args := []interface{}{}

	if filter != nil {
		// CRITICAL: Filter by project_id first
		if filter.ProjectID != "" {
			query += " AND project_id = ?"
			args = append(args, filter.ProjectID)
		}

		// Apply ALL the same filters as GetByProjectID
		if filter.SessionID != nil {
			query += " AND session_id = ?"
			args = append(args, *filter.SessionID)
		}
		if filter.UserID != nil {
			query += " AND user_id = ?"
			args = append(args, *filter.UserID)
		}
		if filter.Environment != nil {
			query += " AND environment = ?"
			args = append(args, *filter.Environment)
		}
		if filter.ServiceName != nil {
			query += " AND service_name = ?" // Materialized column (5-10x faster than JSON extraction)
			args = append(args, *filter.ServiceName)
		}
		if filter.StatusCode != nil {
			query += " AND status_code = ?"
			args = append(args, *filter.StatusCode)
		}
		if filter.Bookmarked != nil {
			query += " AND bookmarked = ?"
			args = append(args, *filter.Bookmarked)
		}
		if filter.Public != nil {
			query += " AND public = ?"
			args = append(args, *filter.Public)
		}
		if filter.StartTime != nil {
			query += " AND start_time >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND start_time <= ?"
			args = append(args, *filter.EndTime)
		}
		if len(filter.Tags) > 0 {
			query += " AND hasAll(tags, ?)"
			args = append(args, filter.Tags)
		}
	}

	var count uint64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	return int64(count), err
}

// Helper function to scan a single trace from query row
func (r *traceRepository) scanTraceRow(row driver.Row) (*observability.Trace, error) {
	var trace observability.Trace

	err := row.Scan(
		&trace.TraceID,
		&trace.ProjectID,
		&trace.Name,
		&trace.UserID,
		&trace.SessionID,
		&trace.Version,           // Materialized from metadata.brokle.version
		&trace.Release,           // Materialized from metadata.brokle.release
		&trace.Tags,
		&trace.Environment,
		&trace.Metadata,          // JSON type
		&trace.StartTime,
		&trace.EndTime,
		&trace.Duration,
		&trace.StatusCode,
		&trace.StatusMessage,
		&trace.Input,
		&trace.Output,
		&trace.TotalCost,        // Pre-computed aggregation
		&trace.TotalTokens,
		&trace.SpanCount,
		&trace.Bookmarked,
		&trace.Public,
		&trace.ServiceName,      // Materialized from metadata.resourceAttributes.service.name
		&trace.CreatedAt,
		&trace.UpdatedAt,
		&trace.DeletedAt,        // Soft delete
	)

	if err != nil {
		return nil, fmt.Errorf("scan trace: %w", err)
	}

	return &trace, nil
}

// Helper function to scan traces from query rows
func (r *traceRepository) scanTraces(rows driver.Rows) ([]*observability.Trace, error) {
	traces := make([]*observability.Trace, 0) // Initialize empty slice to return [] instead of nil

	for rows.Next() {
		var trace observability.Trace

		err := rows.Scan(
			&trace.TraceID,
			&trace.ProjectID,
			&trace.Name,
			&trace.UserID,
			&trace.SessionID,
			&trace.Version,           // Materialized from metadata.brokle.version
			&trace.Release,           // Materialized from metadata.brokle.release
			&trace.Tags,
			&trace.Environment,
			&trace.Metadata,          // JSON type
			&trace.StartTime,
			&trace.EndTime,
			&trace.Duration,
			&trace.StatusCode,
			&trace.StatusMessage,
			&trace.Input,
			&trace.Output,
			&trace.TotalCost,        // Pre-computed aggregation
			&trace.TotalTokens,
			&trace.SpanCount,
			&trace.Bookmarked,
			&trace.Public,
			&trace.ServiceName,      // Materialized from metadata.resourceAttributes.service.name
			&trace.CreatedAt,
			&trace.UpdatedAt,
			&trace.DeletedAt,        // Soft delete
		)

		if err != nil {
			return nil, fmt.Errorf("scan trace: %w", err)
		}

		traces = append(traces, &trace)
	}

	return traces, rows.Err()
}
