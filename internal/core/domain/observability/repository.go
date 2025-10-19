package observability

import (
	"context"
	"time"

	"brokle/pkg/ulid"
)

// TraceRepository defines the interface for trace data access (ClickHouse)
type TraceRepository interface {
	// Basic operations (ReplacingMergeTree pattern)
	Create(ctx context.Context, trace *Trace) error
	Update(ctx context.Context, trace *Trace) error // Inserts with higher version
	Delete(ctx context.Context, id ulid.ULID) error // Soft delete
	GetByID(ctx context.Context, id ulid.ULID) (*Trace, error)

	// Queries
	GetByProjectID(ctx context.Context, projectID ulid.ULID, filter *TraceFilter) ([]*Trace, error)
	GetBySessionID(ctx context.Context, sessionID ulid.ULID) ([]*Trace, error)
	GetChildren(ctx context.Context, parentTraceID ulid.ULID) ([]*Trace, error)
	GetByUserID(ctx context.Context, userID ulid.ULID, filter *TraceFilter) ([]*Trace, error)

	// With relations
	GetWithObservations(ctx context.Context, id ulid.ULID) (*Trace, error)
	GetWithScores(ctx context.Context, id ulid.ULID) (*Trace, error)

	// Batch operations
	CreateBatch(ctx context.Context, traces []*Trace) error

	// Count
	Count(ctx context.Context, filter *TraceFilter) (int64, error)
}

// ObservationRepository defines the interface for observation data access (ClickHouse)
type ObservationRepository interface {
	// Basic operations (ReplacingMergeTree pattern)
	Create(ctx context.Context, obs *Observation) error
	Update(ctx context.Context, obs *Observation) error
	Delete(ctx context.Context, id ulid.ULID) error
	GetByID(ctx context.Context, id ulid.ULID) (*Observation, error)

	// Queries
	GetByTraceID(ctx context.Context, traceID ulid.ULID) ([]*Observation, error)
	GetChildren(ctx context.Context, parentObservationID ulid.ULID) ([]*Observation, error)
	GetTreeByTraceID(ctx context.Context, traceID ulid.ULID) ([]*Observation, error) // Recursive tree

	// Filters
	GetByFilter(ctx context.Context, filter *ObservationFilter) ([]*Observation, error)

	// Batch operations
	CreateBatch(ctx context.Context, observations []*Observation) error

	// Count
	Count(ctx context.Context, filter *ObservationFilter) (int64, error)
}

// ScoreRepository defines the interface for score data access (ClickHouse)
type ScoreRepository interface {
	// Basic operations (ReplacingMergeTree pattern)
	Create(ctx context.Context, score *Score) error
	Update(ctx context.Context, score *Score) error
	Delete(ctx context.Context, id ulid.ULID) error
	GetByID(ctx context.Context, id ulid.ULID) (*Score, error)

	// Queries
	GetByTraceID(ctx context.Context, traceID ulid.ULID) ([]*Score, error)
	GetByObservationID(ctx context.Context, observationID ulid.ULID) ([]*Score, error)
	GetBySessionID(ctx context.Context, sessionID ulid.ULID) ([]*Score, error)

	// Filters
	GetByFilter(ctx context.Context, filter *ScoreFilter) ([]*Score, error)

	// Batch operations
	CreateBatch(ctx context.Context, scores []*Score) error

	// Count
	Count(ctx context.Context, filter *ScoreFilter) (int64, error)
}

// SessionRepository defines the interface for session data access (ClickHouse)
type SessionRepository interface {
	// Basic operations (ReplacingMergeTree pattern)
	Create(ctx context.Context, session *Session) error
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, id ulid.ULID) error
	GetByID(ctx context.Context, id ulid.ULID) (*Session, error)

	// Queries
	GetByProjectID(ctx context.Context, projectID ulid.ULID, filter *SessionFilter) ([]*Session, error)
	GetByUserID(ctx context.Context, userID ulid.ULID) ([]*Session, error)

	// With relations
	GetWithTraces(ctx context.Context, id ulid.ULID) (*Session, error)

	// Count
	Count(ctx context.Context, filter *SessionFilter) (int64, error)
}

// Filter types

// TraceFilter represents filters for trace queries
type TraceFilter struct {
	UserID      *ulid.ULID
	SessionID   *ulid.ULID
	ParentID    *ulid.ULID
	StartTime   *time.Time
	EndTime     *time.Time
	Tags        []string
	Environment *string
	Limit       int
	Offset      int
}

// ObservationFilter represents filters for observation queries
type ObservationFilter struct {
	TraceID         *ulid.ULID
	ParentID        *ulid.ULID
	Type            *ObservationType
	Model           *string
	StartTime       *time.Time
	EndTime         *time.Time
	MinLatencyMs    *uint32
	MaxLatencyMs    *uint32
	MinCost         *float64
	MaxCost         *float64
	Level           *ObservationLevel
	IsCompleted     *bool
	Limit           int
	Offset          int
}

// ScoreFilter represents filters for score queries
type ScoreFilter struct {
	TraceID         *ulid.ULID
	ObservationID   *ulid.ULID
	SessionID       *ulid.ULID
	Name            *string
	Source          *ScoreSource
	DataType        *ScoreDataType
	EvaluatorName   *string
	MinValue        *float64
	MaxValue        *float64
	StartTime       *time.Time
	EndTime         *time.Time
	Limit           int
	Offset          int
}

// SessionFilter represents filters for session queries
type SessionFilter struct {
	UserID      *ulid.ULID
	Bookmarked  *bool
	Public      *bool
	StartTime   *time.Time
	EndTime     *time.Time
	Limit       int
	Offset      int
}


// TelemetryDeduplicationRepository defines methods for telemetry deduplication
type TelemetryDeduplicationRepository interface {
	// Atomic claim operations for deduplication
	ClaimEvents(ctx context.Context, projectID, batchID ulid.ULID, eventIDs []ulid.ULID, ttl time.Duration) (claimed []ulid.ULID, duplicates []ulid.ULID, err error)
	ReleaseEvents(ctx context.Context, eventIDs []ulid.ULID) error

	// Individual operations
	CheckDuplicate(ctx context.Context, eventID ulid.ULID) (bool, error)
	RegisterEvent(ctx context.Context, eventID, batchID, projectID ulid.ULID, ttl time.Duration) error
	Exists(ctx context.Context, eventID ulid.ULID) (bool, error)
	Create(ctx context.Context, dedup *TelemetryEventDeduplication) error
	Delete(ctx context.Context, eventID ulid.ULID) error

	// Batch operations
	CheckBatchDuplicates(ctx context.Context, eventIDs []ulid.ULID) ([]ulid.ULID, error)
	CreateBatch(ctx context.Context, dedups []*TelemetryEventDeduplication) error

	// Statistics
	CountByProjectID(ctx context.Context, projectID ulid.ULID) (int64, error)
}

// TelemetryAnalyticsRepository defines methods for storing telemetry data in ClickHouse
type TelemetryAnalyticsRepository interface {
	// Telemetry event operations
	InsertTelemetryEvent(ctx context.Context, event *TelemetryEvent) error
	InsertTelemetryEventsBatch(ctx context.Context, events []*TelemetryEvent) error

	// Telemetry batch operations
	InsertTelemetryBatch(ctx context.Context, batch *TelemetryBatch) error
	InsertTelemetryBatchesBatch(ctx context.Context, batches []*TelemetryBatch) error

	// Telemetry metric operations
	InsertTelemetryMetric(ctx context.Context, metric *TelemetryMetric) error
	InsertTelemetryMetricsBatch(ctx context.Context, metrics []*TelemetryMetric) error
}
