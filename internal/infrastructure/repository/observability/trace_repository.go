package observability

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/ulid"
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

// Create inserts a new trace into ClickHouse
func (r *traceRepository) Create(ctx context.Context, trace *observability.Trace) error {
	// Set version and event_ts for new traces
	// Only set version to 1 if it's currently 0 (new record)
	// This allows Update() to increment version without being reset
	if trace.Version == 0 {
		trace.Version = 1
	}
	trace.EventTs = time.Now()

	query := `
		INSERT INTO traces (
			id, project_id, session_id, parent_trace_id, name, user_id,
			timestamp, input, output, metadata, tags, environment, release,
			version, event_ts, is_deleted
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.db.Exec(ctx, query,
		trace.ID.String(),
		trace.ProjectID.String(),
		ulidPtrToString(trace.SessionID),
		ulidPtrToString(trace.ParentTraceID),
		trace.Name,
		ulidPtrToString(trace.UserID),
		trace.Timestamp,
		trace.Input,
		trace.Output,
		trace.Metadata,
		trace.Tags,
		trace.Environment,
		trace.Release,
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

	// Same INSERT query as Create - ClickHouse will handle merging based on ORDER BY
	return r.Create(ctx, trace)
}

// Delete performs soft deletion by inserting a record with is_deleted = true
func (r *traceRepository) Delete(ctx context.Context, id ulid.ULID) error {
	query := `
		INSERT INTO traces
		SELECT
			id, project_id, session_id, parent_trace_id, name, user_id,
			timestamp, input, output, metadata, tags, environment, release,
			version + 1 as version,
			now64() as event_ts,
			1 as is_deleted
		FROM traces
		WHERE id = ? AND is_deleted = 0
		ORDER BY event_ts DESC
		LIMIT 1
	`

	return r.db.Exec(ctx, query, id.String())
}

// GetByID retrieves a trace by its ID (returns latest version)
func (r *traceRepository) GetByID(ctx context.Context, id ulid.ULID) (*observability.Trace, error) {
	query := `
		SELECT
			id, project_id, session_id, parent_trace_id, name, user_id,
			timestamp, input, output, metadata, tags, environment, release,
			version, event_ts, is_deleted
		FROM traces
		WHERE id = ? AND is_deleted = 0
		ORDER BY event_ts DESC
		LIMIT 1
	`

	var trace observability.Trace
	var (
		idStr, projectID, sessionID, parentTraceID, userID, input, output, release *string
		metadata                                                                    map[string]string
		tags                                                                        []string
		version                                                                     uint32
		eventTs                                                                     time.Time
		isDeleted                                                                   uint8
	)

	err := r.db.QueryRow(ctx, query, id.String()).Scan(
		&idStr,
		&projectID,
		&sessionID,
		&parentTraceID,
		&trace.Name,
		&userID,
		&trace.Timestamp,
		&input,
		&output,
		&metadata,
		&tags,
		&trace.Environment,
		&release,
		&version,
		&eventTs,
		&isDeleted,
	)

	if err != nil {
		return nil, fmt.Errorf("get trace by id: %w", err)
	}

	// Parse ULIDs
	if idStr != nil {
		parsedID, _ := ulid.Parse(*idStr)
		trace.ID = parsedID
	}
	if projectID != nil {
		parsedProjID, _ := ulid.Parse(*projectID)
		trace.ProjectID = parsedProjID
	}
	trace.SessionID = stringToUlidPtr(sessionID)
	trace.ParentTraceID = stringToUlidPtr(parentTraceID)
	trace.UserID = stringToUlidPtr(userID)
	trace.Input = input
	trace.Output = output
	trace.Metadata = metadata
	trace.Tags = tags
	trace.Release = release
	trace.Version = version
	trace.EventTs = eventTs
	trace.IsDeleted = isDeleted != 0

	return &trace, nil
}

// GetByProjectID retrieves traces by project ID with optional filters
func (r *traceRepository) GetByProjectID(ctx context.Context, projectID ulid.ULID, filter *observability.TraceFilter) ([]*observability.Trace, error) {
	query := `
		SELECT
			id, project_id, session_id, parent_trace_id, name, user_id,
			timestamp, input, output, metadata, tags, environment, release,
			version, event_ts, is_deleted
		FROM traces
		WHERE project_id = ? AND is_deleted = 0
	`

	args := []interface{}{projectID.String()}

	// Apply filters
	if filter != nil {
		if filter.SessionID != nil {
			query += " AND session_id = ?"
			args = append(args, filter.SessionID.String())
		}
		if filter.UserID != nil {
			query += " AND user_id = ?"
			args = append(args, filter.UserID.String())
		}
		if filter.ParentID != nil {
			query += " AND parent_trace_id = ?"
			args = append(args, filter.ParentID.String())
		}
		if filter.Environment != nil {
			query += " AND environment = ?"
			args = append(args, *filter.Environment)
		}
		if filter.StartTime != nil {
			query += " AND timestamp >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND timestamp <= ?"
			args = append(args, *filter.EndTime)
		}
	}

	// Order by timestamp descending (most recent first)
	query += " ORDER BY timestamp DESC"

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

	var traces []*observability.Trace
	for rows.Next() {
		trace, err := r.scanTrace(rows)
		if err != nil {
			return nil, fmt.Errorf("scan trace: %w", err)
		}
		traces = append(traces, trace)
	}

	return traces, rows.Err()
}

// GetBySessionID retrieves all traces in a session
func (r *traceRepository) GetBySessionID(ctx context.Context, sessionID ulid.ULID) ([]*observability.Trace, error) {
	query := `
		SELECT
			id, project_id, session_id, parent_trace_id, name, user_id,
			timestamp, input, output, metadata, tags, environment, release,
			version, event_ts, is_deleted
		FROM traces
		WHERE session_id = ? AND is_deleted = 0
		ORDER BY timestamp ASC
	`

	rows, err := r.db.Query(ctx, query, sessionID.String())
	if err != nil {
		return nil, fmt.Errorf("query traces by session: %w", err)
	}
	defer rows.Close()

	var traces []*observability.Trace
	for rows.Next() {
		trace, err := r.scanTrace(rows)
		if err != nil {
			return nil, fmt.Errorf("scan trace: %w", err)
		}
		traces = append(traces, trace)
	}

	return traces, rows.Err()
}

// GetChildren retrieves child traces of a parent trace
func (r *traceRepository) GetChildren(ctx context.Context, parentTraceID ulid.ULID) ([]*observability.Trace, error) {
	query := `
		SELECT
			id, project_id, session_id, parent_trace_id, name, user_id,
			timestamp, input, output, metadata, tags, environment, release,
			version, event_ts, is_deleted
		FROM traces
		WHERE parent_trace_id = ? AND is_deleted = 0
		ORDER BY timestamp ASC
	`

	rows, err := r.db.Query(ctx, query, parentTraceID.String())
	if err != nil {
		return nil, fmt.Errorf("query child traces: %w", err)
	}
	defer rows.Close()

	var traces []*observability.Trace
	for rows.Next() {
		trace, err := r.scanTrace(rows)
		if err != nil {
			return nil, fmt.Errorf("scan trace: %w", err)
		}
		traces = append(traces, trace)
	}

	return traces, rows.Err()
}

// GetByUserID retrieves traces by user ID
func (r *traceRepository) GetByUserID(ctx context.Context, userID ulid.ULID, filter *observability.TraceFilter) ([]*observability.Trace, error) {
	query := `
		SELECT
			id, project_id, session_id, parent_trace_id, name, user_id,
			timestamp, input, output, metadata, tags, environment, release,
			version, event_ts, is_deleted
		FROM traces
		WHERE user_id = ? AND is_deleted = 0
	`

	args := []interface{}{userID.String()}

	if filter != nil {
		if filter.StartTime != nil {
			query += " AND timestamp >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND timestamp <= ?"
			args = append(args, *filter.EndTime)
		}
	}

	query += " ORDER BY timestamp DESC"

	if filter != nil && filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query traces by user: %w", err)
	}
	defer rows.Close()

	var traces []*observability.Trace
	for rows.Next() {
		trace, err := r.scanTrace(rows)
		if err != nil {
			return nil, fmt.Errorf("scan trace: %w", err)
		}
		traces = append(traces, trace)
	}

	return traces, rows.Err()
}

// GetWithObservations retrieves a trace with all its observations (not implemented yet - requires join)
func (r *traceRepository) GetWithObservations(ctx context.Context, id ulid.ULID) (*observability.Trace, error) {
	// TODO: Implement after ObservationRepository is created
	return r.GetByID(ctx, id)
}

// GetWithScores retrieves a trace with all its scores (not implemented yet - requires join)
func (r *traceRepository) GetWithScores(ctx context.Context, id ulid.ULID) (*observability.Trace, error) {
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
			id, project_id, session_id, parent_trace_id, name, user_id,
			timestamp, input, output, metadata, tags, environment, release,
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

		err = batch.Append(
			trace.ID.String(),
			trace.ProjectID.String(),
			ulidPtrToString(trace.SessionID),
			ulidPtrToString(trace.ParentTraceID),
			trace.Name,
			ulidPtrToString(trace.UserID),
			trace.Timestamp,
			trace.Input,
			trace.Output,
			trace.Metadata,
			trace.Tags,
			trace.Environment,
			trace.Release,
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
			args = append(args, filter.SessionID.String())
		}
		if filter.UserID != nil {
			query += " AND user_id = ?"
			args = append(args, filter.UserID.String())
		}
		if filter.StartTime != nil {
			query += " AND timestamp >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND timestamp <= ?"
			args = append(args, *filter.EndTime)
		}
	}

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

// Helper function to scan a trace from query rows
func (r *traceRepository) scanTrace(rows driver.Rows) (*observability.Trace, error) {
	var trace observability.Trace
	var (
		idStr, projectID, sessionID, parentTraceID, userID, input, output, release *string
		metadata                                                                    map[string]string
		tags                                                                        []string
		version                                                                     uint32
		eventTs                                                                     time.Time
		isDeleted                                                                   uint8
	)

	err := rows.Scan(
		&idStr,
		&projectID,
		&sessionID,
		&parentTraceID,
		&trace.Name,
		&userID,
		&trace.Timestamp,
		&input,
		&output,
		&metadata,
		&tags,
		&trace.Environment,
		&release,
		&version,
		&eventTs,
		&isDeleted,
	)

	if err != nil {
		return nil, err
	}

	// Parse ULIDs
	if idStr != nil {
		parsedID, _ := ulid.Parse(*idStr)
		trace.ID = parsedID
	}
	if projectID != nil {
		parsedProjID, _ := ulid.Parse(*projectID)
		trace.ProjectID = parsedProjID
	}
	trace.SessionID = stringToUlidPtr(sessionID)
	trace.ParentTraceID = stringToUlidPtr(parentTraceID)
	trace.UserID = stringToUlidPtr(userID)
	trace.Input = input
	trace.Output = output
	trace.Metadata = metadata
	trace.Tags = tags
	trace.Release = release
	trace.Version = version
	trace.EventTs = eventTs
	trace.IsDeleted = isDeleted != 0

	return &trace, nil
}

// Helper functions

func ulidPtrToString(u *ulid.ULID) *string {
	if u == nil {
		return nil
	}
	s := u.String()
	return &s
}

func stringToUlidPtr(s *string) *ulid.ULID {
	if s == nil || *s == "" {
		return nil
	}
	u, err := ulid.Parse(*s)
	if err != nil {
		return nil
	}
	return &u
}

func boolToUint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
