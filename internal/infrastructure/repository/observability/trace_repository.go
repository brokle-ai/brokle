package observability

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/core/domain/observability"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type traceRepository struct {
	db clickhouse.Conn
}

// NewTraceRepository creates a new trace repository instance
func NewTraceRepository(db clickhouse.Conn) observability.TraceRepository {
	return &traceRepository{db: db}
}

// Create inserts a new OTEL trace into ClickHouse
func (r *traceRepository) Create(ctx context.Context, trace *observability.Trace) error {
	// Set version and event_ts for new traces
	if trace.Version == 0 {
		trace.Version = 1
	}
	trace.EventTs = time.Now()
	trace.UpdatedAt = time.Now()

	// Calculate duration if not set
	trace.CalculateDuration()

	query := `
		INSERT INTO traces (
			id, project_id, name, user_id, session_id,
			start_time, end_time, duration_ms, status_code, status_message,
			attributes, input, output, input_blob_storage_id, output_blob_storage_id,
			input_preview, output_preview, metadata, tags,
			environment, service_name, service_version, release,
			total_cost, total_tokens, observation_count,
			bookmarked, public, created_at, updated_at,
			version, event_ts, is_deleted
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.db.Exec(ctx, query,
		trace.ID,
		trace.ProjectID,
		trace.Name,
		trace.UserID,
		trace.SessionID,
		trace.StartTime,
		trace.EndTime,
		trace.DurationMs,
		trace.StatusCode,
		trace.StatusMessage,
		trace.Attributes,
		trace.Input,
		trace.Output,
		trace.InputBlobStorageID,
		trace.OutputBlobStorageID,
		trace.InputPreview,
		trace.OutputPreview,
		trace.Metadata,
		trace.Tags,
		trace.Environment,
		trace.ServiceName,
		trace.ServiceVersion,
		trace.Release,
		trace.TotalCost,
		trace.TotalTokens,
		trace.ObservationCount,
		boolToUint8(trace.Bookmarked),
		boolToUint8(trace.Public),
		trace.CreatedAt,
		trace.UpdatedAt,
		trace.Version,
		trace.EventTs,
		boolToUint8(trace.IsDeleted),
	)
}

// Update performs an update using ReplacingMergeTree pattern (insert with higher version)
func (r *traceRepository) Update(ctx context.Context, trace *observability.Trace) error {
	// ReplacingMergeTree pattern: increment version and update event_ts
	trace.Version++
	trace.EventTs = time.Now()
	trace.UpdatedAt = time.Now()

	// Calculate duration if not set
	trace.CalculateDuration()

	// Same INSERT query as Create - ClickHouse will handle merging based on ORDER BY
	return r.Create(ctx, trace)
}

// Delete performs soft deletion by inserting a record with is_deleted = true
func (r *traceRepository) Delete(ctx context.Context, id string) error {
	query := `
		INSERT INTO traces
		SELECT
			id, project_id, name, user_id, session_id,
			start_time, end_time, duration_ms, status_code, status_message,
			attributes, input, output, input_blob_storage_id, output_blob_storage_id,
			input_preview, output_preview, metadata, tags,
			environment, service_name, service_version, release,
			total_cost, total_tokens, observation_count,
			bookmarked, public, created_at, updated_at,
			version + 1 as version,
			now64() as event_ts,
			1 as is_deleted
		FROM traces
		WHERE id = ? AND is_deleted = 0
		ORDER BY event_ts DESC
		LIMIT 1
	`

	return r.db.Exec(ctx, query, id)
}

// GetByID retrieves a trace by its OTEL trace_id (returns latest version)
func (r *traceRepository) GetByID(ctx context.Context, id string) (*observability.Trace, error) {
	query := `
		SELECT
			id, project_id, name, user_id, session_id,
			start_time, end_time, duration_ms, status_code, status_message,
			attributes, input, output, input_blob_storage_id, output_blob_storage_id,
			input_preview, output_preview, metadata, tags,
			environment, service_name, service_version, release,
			total_cost, total_tokens, observation_count,
			bookmarked, public, created_at, updated_at,
			version, event_ts, is_deleted
		FROM traces
		WHERE id = ? AND is_deleted = 0
		ORDER BY event_ts DESC
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, id)
	return r.scanTraceRow(row)
}

// GetByProjectID retrieves traces by project ID with optional filters
func (r *traceRepository) GetByProjectID(ctx context.Context, projectID string, filter *observability.TraceFilter) ([]*observability.Trace, error) {
	query := `
		SELECT
			id, project_id, name, user_id, session_id,
			start_time, end_time, duration_ms, status_code, status_message,
			attributes, input, output, input_blob_storage_id, output_blob_storage_id,
			input_preview, output_preview, metadata, tags,
			environment, service_name, service_version, release,
			total_cost, total_tokens, observation_count,
			bookmarked, public, created_at, updated_at,
			version, event_ts, is_deleted
		FROM traces
		WHERE project_id = ? AND is_deleted = 0
	`

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
			query += " AND service_name = ?"
			args = append(args, *filter.ServiceName)
		}
		if filter.StatusCode != nil {
			query += " AND status_code = ?"
			args = append(args, *filter.StatusCode)
		}
		if filter.Bookmarked != nil {
			query += " AND bookmarked = ?"
			args = append(args, boolToUint8(*filter.Bookmarked))
		}
		if filter.Public != nil {
			query += " AND public = ?"
			args = append(args, boolToUint8(*filter.Public))
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

	// Order by start_time descending (most recent first)
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
		return nil, fmt.Errorf("query traces by project: %w", err)
	}
	defer rows.Close()

	return r.scanTraces(rows)
}

// GetBySessionID retrieves all traces in a virtual session
func (r *traceRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*observability.Trace, error) {
	query := `
		SELECT
			id, project_id, name, user_id, session_id,
			start_time, end_time, duration_ms, status_code, status_message,
			attributes, input, output, input_blob_storage_id, output_blob_storage_id,
			input_preview, output_preview, metadata, tags,
			environment, service_name, service_version, release,
			total_cost, total_tokens, observation_count,
			bookmarked, public, created_at, updated_at,
			version, event_ts, is_deleted
		FROM traces
		WHERE session_id = ? AND is_deleted = 0
		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("query traces by session: %w", err)
	}
	defer rows.Close()

	return r.scanTraces(rows)
}

// GetByUserID retrieves traces by user ID
func (r *traceRepository) GetByUserID(ctx context.Context, userID string, filter *observability.TraceFilter) ([]*observability.Trace, error) {
	query := `
		SELECT
			id, project_id, name, user_id, session_id,
			start_time, end_time, duration_ms, status_code, status_message,
			attributes, input, output, input_blob_storage_id, output_blob_storage_id,
			input_preview, output_preview, metadata, tags,
			environment, service_name, service_version, release,
			total_cost, total_tokens, observation_count,
			bookmarked, public, created_at, updated_at,
			version, event_ts, is_deleted
		FROM traces
		WHERE user_id = ? AND is_deleted = 0
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

	query += " ORDER BY start_time DESC"

	if filter != nil && filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query traces by user: %w", err)
	}
	defer rows.Close()

	return r.scanTraces(rows)
}

// GetWithObservations retrieves a trace with all its observations (requires join)
func (r *traceRepository) GetWithObservations(ctx context.Context, id string) (*observability.Trace, error) {
	// TODO: Implement after ObservationRepository is created
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
			id, project_id, name, user_id, session_id,
			start_time, end_time, duration_ms, status_code, status_message,
			attributes, input, output, input_blob_storage_id, output_blob_storage_id,
			input_preview, output_preview, metadata, tags,
			environment, service_name, service_version, release,
			total_cost, total_tokens, observation_count,
			bookmarked, public, created_at, updated_at,
			version, event_ts, is_deleted
		)
	`)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, trace := range traces {
		// Set version and event_ts for new traces
		if trace.Version == 0 {
			trace.Version = 1
			trace.EventTs = time.Now()
		}
		if trace.UpdatedAt.IsZero() {
			trace.UpdatedAt = time.Now()
		}

		// Calculate duration if not set
		trace.CalculateDuration()

		err = batch.Append(
			trace.ID,
			trace.ProjectID,
			trace.Name,
			trace.UserID,
			trace.SessionID,
			trace.StartTime,
			trace.EndTime,
			trace.DurationMs,
			trace.StatusCode,
			trace.StatusMessage,
			trace.Attributes,
			trace.Input,
			trace.Output,
			trace.InputBlobStorageID,
			trace.OutputBlobStorageID,
			trace.InputPreview,
			trace.OutputPreview,
			trace.Metadata,
			trace.Tags,
			trace.Environment,
			trace.ServiceName,
			trace.ServiceVersion,
			trace.Release,
			trace.TotalCost,
			trace.TotalTokens,
			trace.ObservationCount,
			boolToUint8(trace.Bookmarked),
			boolToUint8(trace.Public),
			trace.CreatedAt,
			trace.UpdatedAt,
			trace.Version,
			trace.EventTs,
			boolToUint8(trace.IsDeleted),
		)
		if err != nil {
			return fmt.Errorf("append to batch: %w", err)
		}
	}

	return batch.Send()
}

// Count returns the count of traces matching the filter
func (r *traceRepository) Count(ctx context.Context, filter *observability.TraceFilter) (int64, error) {
	query := "SELECT count() FROM traces WHERE is_deleted = 0"
	args := []interface{}{}

	if filter != nil {
		if filter.SessionID != nil {
			query += " AND session_id = ?"
			args = append(args, *filter.SessionID)
		}
		if filter.UserID != nil {
			query += " AND user_id = ?"
			args = append(args, *filter.UserID)
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

// Helper function to scan a single trace from query row
func (r *traceRepository) scanTraceRow(row driver.Row) (*observability.Trace, error) {
	var trace observability.Trace
	var bookmarked, public, isDeleted uint8

	err := row.Scan(
		&trace.ID,
		&trace.ProjectID,
		&trace.Name,
		&trace.UserID,
		&trace.SessionID,
		&trace.StartTime,
		&trace.EndTime,
		&trace.DurationMs,
		&trace.StatusCode,
		&trace.StatusMessage,
		&trace.Attributes,
		&trace.Input,
		&trace.Output,
		&trace.InputBlobStorageID,
		&trace.OutputBlobStorageID,
		&trace.InputPreview,
		&trace.OutputPreview,
		&trace.Metadata,
		&trace.Tags,
		&trace.Environment,
		&trace.ServiceName,
		&trace.ServiceVersion,
		&trace.Release,
		&trace.TotalCost,
		&trace.TotalTokens,
		&trace.ObservationCount,
		&bookmarked,
		&public,
		&trace.CreatedAt,
		&trace.UpdatedAt,
		&trace.Version,
		&trace.EventTs,
		&isDeleted,
	)

	if err != nil {
		return nil, fmt.Errorf("scan trace: %w", err)
	}

	trace.Bookmarked = bookmarked != 0
	trace.Public = public != 0
	trace.IsDeleted = isDeleted != 0

	return &trace, nil
}

// Helper function to scan traces from query rows
func (r *traceRepository) scanTraces(rows driver.Rows) ([]*observability.Trace, error) {
	var traces []*observability.Trace

	for rows.Next() {
		var trace observability.Trace
		var bookmarked, public, isDeleted uint8

		err := rows.Scan(
			&trace.ID,
			&trace.ProjectID,
			&trace.Name,
			&trace.UserID,
			&trace.SessionID,
			&trace.StartTime,
			&trace.EndTime,
			&trace.DurationMs,
			&trace.StatusCode,
			&trace.StatusMessage,
			&trace.Attributes,
			&trace.Input,
			&trace.Output,
			&trace.InputBlobStorageID,
			&trace.OutputBlobStorageID,
			&trace.InputPreview,
			&trace.OutputPreview,
			&trace.Metadata,
			&trace.Tags,
			&trace.Environment,
			&trace.ServiceName,
			&trace.ServiceVersion,
			&trace.Release,
			&trace.TotalCost,
			&trace.TotalTokens,
			&trace.ObservationCount,
			&bookmarked,
			&public,
			&trace.CreatedAt,
			&trace.UpdatedAt,
			&trace.Version,
			&trace.EventTs,
			&isDeleted,
		)

		if err != nil {
			return nil, fmt.Errorf("scan trace: %w", err)
		}

		trace.Bookmarked = bookmarked != 0
		trace.Public = public != 0
		trace.IsDeleted = isDeleted != 0

		traces = append(traces, &trace)
	}

	return traces, rows.Err()
}

// Helper function
func boolToUint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
