package observability

import (
	"context"
	"time"

	"brokle/pkg/pagination"
	"brokle/pkg/ulid"
)

// TraceRepository defines the interface for trace data access (ClickHouse)
type TraceRepository interface {
	// Basic operations (ReplacingMergeTree pattern)
	Create(ctx context.Context, trace *Trace) error
	Update(ctx context.Context, trace *Trace) error // Inserts with higher version
	Delete(ctx context.Context, id string) error    // Soft delete (OTEL trace_id)
	GetByID(ctx context.Context, id string) (*Trace, error)

	// Queries
	GetByProjectID(ctx context.Context, projectID string, filter *TraceFilter) ([]*Trace, error)
	GetBySessionID(ctx context.Context, sessionID string) ([]*Trace, error) // Virtual session analytics
	GetByUserID(ctx context.Context, userID string, filter *TraceFilter) ([]*Trace, error)

	// With relations
	GetWithSpans(ctx context.Context, id string) (*Trace, error)
	GetWithScores(ctx context.Context, id string) (*Trace, error)

	// Batch operations
	CreateBatch(ctx context.Context, traces []*Trace) error

	// Count
	Count(ctx context.Context, filter *TraceFilter) (int64, error)
}

// SpanRepository defines the interface for span data access (ClickHouse)
type SpanRepository interface {
	// Basic operations (ReplacingMergeTree pattern)
	Create(ctx context.Context, span *Span) error
	Update(ctx context.Context, span *Span) error
	Delete(ctx context.Context, id string) error // Soft delete (OTEL span_id)
	GetByID(ctx context.Context, id string) (*Span, error)

	// Queries
	GetByTraceID(ctx context.Context, traceID string) ([]*Span, error)
	GetRootSpan(ctx context.Context, traceID string) (*Span, error) // Get span with parent_span_id IS NULL
	GetChildren(ctx context.Context, parentSpanID string) ([]*Span, error)
	GetTreeByTraceID(ctx context.Context, traceID string) ([]*Span, error) // Recursive tree

	// Filters
	GetByFilter(ctx context.Context, filter *SpanFilter) ([]*Span, error)

	// Batch operations
	CreateBatch(ctx context.Context, spans []*Span) error

	// Count
	Count(ctx context.Context, filter *SpanFilter) (int64, error)
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
	GetBySpanID(ctx context.Context, spanID string) ([]*Score, error)

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
	// Scope
	ProjectID string // Required for scoping queries to project

	// Domain filters
	UserID      *string
	SessionID   *string // Virtual session filtering
	StartTime   *time.Time
	EndTime     *time.Time
	Tags        []string
	Environment *string
	ServiceName *string
	StatusCode  *string
	Bookmarked  *bool
	Public      *bool

	// Pagination (embedded for DRY)
	pagination.Params
}

// SpanFilter represents filters for span queries
type SpanFilter struct {
	// Scope
	ProjectID string // Required for scoping queries to project

	// Domain filters
	TraceID      *string
	ParentID     *string
	Type         *string
	SpanKind     *string
	Model        *string
	StartTime    *time.Time
	EndTime      *time.Time
	MinLatencyMs *uint32
	MaxLatencyMs *uint32
	MinCost      *float64
	MaxCost      *float64
	Level        *string
	IsCompleted  *bool

	// Pagination (embedded for DRY)
	pagination.Params
}

// ScoreFilter represents filters for score queries
type ScoreFilter struct {
	// Scope
	ProjectID string // Required for scoping queries to project

	// Domain filters
	TraceID       *string
	SpanID        *string
	Name          *string
	Source        *string
	DataType      *string
	EvaluatorName *string
	MinValue      *float64
	MaxValue      *float64
	StartTime     *time.Time
	EndTime       *time.Time

	// Pagination (embedded for DRY)
	pagination.Params
}

// BlobStorageFilter represents filters for blob storage queries
type BlobStorageFilter struct {
	// Domain filters
	EntityType   *string
	StartTime    *time.Time
	EndTime      *time.Time
	MinSizeBytes *uint64
	MaxSizeBytes *uint64

	// Pagination (embedded for DRY)
	pagination.Params
}

// TelemetryDeduplicationRepository defines methods for telemetry deduplication
type TelemetryDeduplicationRepository interface {
	// Atomic claim operations for deduplication
	ClaimEvents(ctx context.Context, projectID ulid.ULID, batchID ulid.ULID, dedupIDs []string, ttl time.Duration) (claimed []string, duplicates []string, err error)
	ReleaseEvents(ctx context.Context, dedupIDs []string) error

	// Individual operations
	CheckDuplicate(ctx context.Context, dedupID string) (bool, error)
	RegisterEvent(ctx context.Context, dedupID string, batchID ulid.ULID, projectID ulid.ULID, ttl time.Duration) error
	Exists(ctx context.Context, dedupID string) (bool, error)
	Create(ctx context.Context, dedup *TelemetryEventDeduplication) error
	Delete(ctx context.Context, dedupID string) error

	// Batch operations
	CheckBatchDuplicates(ctx context.Context, dedupIDs []string) ([]string, error)
	CreateBatch(ctx context.Context, dedups []*TelemetryEventDeduplication) error

	// Statistics
	CountByProjectID(ctx context.Context, projectID ulid.ULID) (int64, error)
}

// ModelRepository defines the interface for model pricing data access (PostgreSQL)
// Used for cost calculation during telemetry ingestion
type ModelRepository interface {
	// FindByModelName finds pricing for a model using regex pattern matching
	// Implements Langfuse pattern: project-scoped with global fallback
	// Returns nil if no pricing found (allows ingestion to continue without costs)
	FindByModelName(ctx context.Context, modelName, projectID string) (*Model, error)

	// FindByID retrieves pricing by ID
	GetByID(ctx context.Context, id string) (*Model, error)

	// CRUD operations for pricing management
	Create(ctx context.Context, pricing *Model) error
	Update(ctx context.Context, pricing *Model) error
	Delete(ctx context.Context, id string) error // Soft delete (set end_date)

	// List operations
	List(ctx context.Context, filter *ModelFilter) ([]*Model, error)
	Count(ctx context.Context, filter *ModelFilter) (int64, error)

	// Historical pricing queries
	FindHistoricalPricing(ctx context.Context, modelName, projectID string, timestamp time.Time) (*Model, error)
}

// ModelFilter represents filters for model pricing queries
type ModelFilter struct {
	// Domain filters
	ProjectID    *string
	Provider     *string
	ModelName    *string
	IsActive     *bool // Filter by active pricing (end_date IS NULL)
	IsDeprecated *bool
	StartDate    *time.Time
	EndDate      *time.Time

	// Pagination (embedded for DRY)
	pagination.Params
}
