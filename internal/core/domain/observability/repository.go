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
	Delete(ctx context.Context, id string) error // Soft delete (OTEL trace_id)
	GetByID(ctx context.Context, id string) (*Trace, error)

	// Queries
	GetByProjectID(ctx context.Context, projectID string, filter *TraceFilter) ([]*Trace, error)
	GetBySessionID(ctx context.Context, sessionID string) ([]*Trace, error) // Virtual session analytics
	GetByUserID(ctx context.Context, userID string, filter *TraceFilter) ([]*Trace, error)

	// With relations
	GetWithObservations(ctx context.Context, id string) (*Trace, error)
	GetWithScores(ctx context.Context, id string) (*Trace, error)

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
	Delete(ctx context.Context, id string) error // Soft delete (OTEL span_id)
	GetByID(ctx context.Context, id string) (*Observation, error)

	// Queries
	GetByTraceID(ctx context.Context, traceID string) ([]*Observation, error)
	GetRootSpan(ctx context.Context, traceID string) (*Observation, error) // Get span with parent_observation_id IS NULL
	GetChildren(ctx context.Context, parentObservationID string) ([]*Observation, error)
	GetTreeByTraceID(ctx context.Context, traceID string) ([]*Observation, error) // Recursive tree

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
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*Score, error)

	// Queries
	GetByTraceID(ctx context.Context, traceID string) ([]*Score, error)
	GetByObservationID(ctx context.Context, observationID string) ([]*Score, error)

	// Filters
	GetByFilter(ctx context.Context, filter *ScoreFilter) ([]*Score, error)

	// Batch operations
	CreateBatch(ctx context.Context, scores []*Score) error

	// Count
	Count(ctx context.Context, filter *ScoreFilter) (int64, error)
}

// BlobStorageRepository defines the interface for blob storage file log data access (ClickHouse)
type BlobStorageRepository interface {
	// Basic operations (ReplacingMergeTree pattern)
	Create(ctx context.Context, blob *BlobStorageFileLog) error
	Update(ctx context.Context, blob *BlobStorageFileLog) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*BlobStorageFileLog, error)

	// Queries
	GetByEntityID(ctx context.Context, entityType, entityID string) ([]*BlobStorageFileLog, error)
	GetByProjectID(ctx context.Context, projectID string, filter *BlobStorageFilter) ([]*BlobStorageFileLog, error)

	// Count
	Count(ctx context.Context, filter *BlobStorageFilter) (int64, error)
}

// Filter types

// TraceFilter represents filters for trace queries
type TraceFilter struct {
	UserID        *string
	SessionID     *string // Virtual session filtering
	StartTime     *time.Time
	EndTime       *time.Time
	Tags          []string
	Environment   *string
	ServiceName   *string
	StatusCode    *string
	Bookmarked    *bool
	Public        *bool
	Limit         int
	Offset        int
}

// ObservationFilter represents filters for observation queries
type ObservationFilter struct {
	TraceID         *string
	ParentID        *string
	Type            *string
	SpanKind        *string
	Model           *string
	StartTime       *time.Time
	EndTime         *time.Time
	MinLatencyMs    *uint32
	MaxLatencyMs    *uint32
	MinCost         *float64
	MaxCost         *float64
	Level           *string
	IsCompleted     *bool
	Limit           int
	Offset          int
}

// ScoreFilter represents filters for score queries
type ScoreFilter struct {
	TraceID         *string
	ObservationID   *string
	Name            *string
	Source          *string
	DataType        *string
	EvaluatorName   *string
	MinValue        *float64
	MaxValue        *float64
	StartTime       *time.Time
	EndTime         *time.Time
	Limit           int
	Offset          int
}

// BlobStorageFilter represents filters for blob storage queries
type BlobStorageFilter struct {
	EntityType    *string
	StartTime     *time.Time
	EndTime       *time.Time
	MinSizeBytes  *uint64
	MaxSizeBytes  *uint64
	Limit         int
	Offset        int
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

