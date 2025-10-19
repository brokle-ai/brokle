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

type sessionRepository struct {
	db clickhouse.Conn
}

// NewSessionRepository creates a new session repository instance
func NewSessionRepository(db clickhouse.Conn) observability.SessionRepository {
	return &sessionRepository{db: db}
}

// Create inserts a new session into ClickHouse
func (r *sessionRepository) Create(ctx context.Context, session *observability.Session) error {
	// Set version and event_ts for new sessions
	// Only set version to 1 if it's currently 0 (new record)
	// This allows Update() to increment version without being reset
	if session.Version == 0 {
		session.Version = 1
	}
	session.EventTs = time.Now()

	query := `
		INSERT INTO sessions (
			id, project_id, user_id, metadata, bookmarked, public,
			created_at, version, event_ts, is_deleted
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.db.Exec(ctx, query,
		session.ID.String(),
		session.ProjectID.String(),
		ulidPtrToString(session.UserID),
		session.Metadata,
		boolToUint8(session.Bookmarked),
		boolToUint8(session.Public),
		session.CreatedAt,
		session.Version,
		session.EventTs,
		boolToUint8(session.IsDeleted),
	)
}

// Update performs an update using ReplacingMergeTree pattern (insert with higher version)
func (r *sessionRepository) Update(ctx context.Context, session *observability.Session) error {
	// ReplacingMergeTree pattern: increment version and update event_ts
	session.Version++
	session.EventTs = time.Now()

	// Same INSERT query as Create - ClickHouse will handle merging
	return r.Create(ctx, session)
}

// Delete performs soft deletion by inserting a record with is_deleted = true
func (r *sessionRepository) Delete(ctx context.Context, id ulid.ULID) error {
	query := `
		INSERT INTO sessions
		SELECT
			id, project_id, user_id, metadata, bookmarked, public, created_at,
			version + 1 as version,
			now64() as event_ts,
			1 as is_deleted
		FROM sessions
		WHERE id = ? AND is_deleted = 0
		ORDER BY event_ts DESC
		LIMIT 1
	`

	return r.db.Exec(ctx, query, id.String())
}

// GetByID retrieves a session by its ID (returns latest version)
func (r *sessionRepository) GetByID(ctx context.Context, id ulid.ULID) (*observability.Session, error) {
	query := `
		SELECT
			id, project_id, user_id, metadata, bookmarked, public,
			created_at, version, event_ts, is_deleted
		FROM sessions
		WHERE id = ? AND is_deleted = 0
		ORDER BY event_ts DESC
		LIMIT 1
	`

	var session observability.Session
	var (
		idStr, projectID, userID *string
		metadata                 map[string]string
		bookmarked, public       uint8
		createdAt, eventTs       time.Time
		version, isDeleted       uint32
	)

	err := r.db.QueryRow(ctx, query, id.String()).Scan(
		&idStr,
		&projectID,
		&userID,
		&metadata,
		&bookmarked,
		&public,
		&createdAt,
		&version,
		&eventTs,
		&isDeleted,
	)

	if err != nil {
		return nil, fmt.Errorf("get session by id: %w", err)
	}

	// Parse ULIDs
	if idStr != nil {
		parsedID, _ := ulid.Parse(*idStr)
		session.ID = parsedID
	}
	if projectID != nil {
		parsedProjectID, _ := ulid.Parse(*projectID)
		session.ProjectID = parsedProjectID
	}
	session.UserID = stringToUlidPtr(userID)
	session.Metadata = metadata
	session.Bookmarked = bookmarked != 0
	session.Public = public != 0
	session.CreatedAt = createdAt
	session.Version = version
	session.EventTs = eventTs
	session.IsDeleted = isDeleted != 0

	return &session, nil
}

// GetByProjectID retrieves sessions by project ID with optional filters
func (r *sessionRepository) GetByProjectID(ctx context.Context, projectID ulid.ULID, filter *observability.SessionFilter) ([]*observability.Session, error) {
	query := `
		SELECT
			id, project_id, user_id, metadata, bookmarked, public,
			created_at, version, event_ts, is_deleted
		FROM sessions
		WHERE project_id = ? AND is_deleted = 0
	`

	args := []interface{}{projectID.String()}

	// Apply filters
	if filter != nil {
		if filter.UserID != nil {
			query += " AND user_id = ?"
			args = append(args, filter.UserID.String())
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
			query += " AND created_at >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND created_at <= ?"
			args = append(args, *filter.EndTime)
		}
	}

	// Order by created_at descending (most recent first)
	query += " ORDER BY created_at DESC"

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
		return nil, fmt.Errorf("query sessions by project: %w", err)
	}
	defer rows.Close()

	var sessions []*observability.Session
	for rows.Next() {
		session, err := r.scanSession(rows)
		if err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// GetByUserID retrieves sessions by user ID
func (r *sessionRepository) GetByUserID(ctx context.Context, userID ulid.ULID) ([]*observability.Session, error) {
	query := `
		SELECT
			id, project_id, user_id, metadata, bookmarked, public,
			created_at, version, event_ts, is_deleted
		FROM sessions
		WHERE user_id = ? AND is_deleted = 0
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID.String())
	if err != nil {
		return nil, fmt.Errorf("query sessions by user: %w", err)
	}
	defer rows.Close()

	var sessions []*observability.Session
	for rows.Next() {
		session, err := r.scanSession(rows)
		if err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// GetWithTraces retrieves a session with all its traces (not implemented yet - requires join)
func (r *sessionRepository) GetWithTraces(ctx context.Context, id ulid.ULID) (*observability.Session, error) {
	// TODO: Implement after TraceRepository integration
	// For now, just return the session without traces
	return r.GetByID(ctx, id)
}

// Count returns the count of sessions matching the filter
func (r *sessionRepository) Count(ctx context.Context, filter *observability.SessionFilter) (int64, error) {
	query := "SELECT count() FROM sessions WHERE is_deleted = 0"
	args := []interface{}{}

	if filter != nil {
		if filter.UserID != nil {
			query += " AND user_id = ?"
			args = append(args, filter.UserID.String())
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
			query += " AND created_at >= ?"
			args = append(args, *filter.StartTime)
		}
		if filter.EndTime != nil {
			query += " AND created_at <= ?"
			args = append(args, *filter.EndTime)
		}
	}

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

// Helper function to scan a session from query rows
func (r *sessionRepository) scanSession(rows driver.Rows) (*observability.Session, error) {
	var session observability.Session
	var (
		idStr, projectID, userID *string
		metadata                 map[string]string
		bookmarked, public       uint8
		createdAt, eventTs       time.Time
		version, isDeleted       uint32
	)

	err := rows.Scan(
		&idStr,
		&projectID,
		&userID,
		&metadata,
		&bookmarked,
		&public,
		&createdAt,
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
		session.ID = parsedID
	}
	if projectID != nil {
		parsedProjectID, _ := ulid.Parse(*projectID)
		session.ProjectID = parsedProjectID
	}
	session.UserID = stringToUlidPtr(userID)
	session.Metadata = metadata
	session.Bookmarked = bookmarked != 0
	session.Public = public != 0
	session.CreatedAt = createdAt
	session.Version = version
	session.EventTs = eventTs
	session.IsDeleted = isDeleted != 0

	return &session, nil
}
