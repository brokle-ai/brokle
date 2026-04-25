package observability

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// TelemetryEventRequest is the wire shape for a single telemetry event
// arriving via the OTLP HTTP ingress. Each event carries its own type
// (span / event / log / metric) and a free-form payload that the OTLP
// converter pipeline reshapes into the corresponding domain entity.
type TelemetryEventRequest struct {
	Payload   map[string]any     `json:"payload"`
	Timestamp *time.Time         `json:"timestamp,omitempty"`
	SpanID    string             `json:"span_id"`
	TraceID   string             `json:"trace_id"`
	EventType TelemetryEventType `json:"event_type"`
	EventID   uuid.UUID          `json:"event_id"`
}

// Validate enforces the cross-field constraints the OTLP converter
// can't express in struct tags alone.
func (e *TelemetryEventRequest) Validate() error {
	if e.EventType == TelemetryEventTypeSpan && e.SpanID == "" {
		return errors.New("span events must have non-empty span_id")
	}
	if e.TraceID == "" {
		return errors.New("trace_id is required for all events")
	}
	return nil
}

// ----- Telemetry health & metrics surfaces ---------------------------
//
// Used by the telemetry health worker and the /metrics scrape path.
// The shapes are intentionally narrow — only the fields downstream
// dashboards actually plot.

// RedisHealthStatus reports the health of the deduplication-store
// Redis connection used by the telemetry pipeline.
type RedisHealthStatus struct {
	LastError   *string       `json:"last_error,omitempty"`
	LatencyMs   float64       `json:"latency_ms"`
	MemoryUsage int64         `json:"memory_usage_bytes"`
	Connections int           `json:"connections"`
	Uptime      time.Duration `json:"uptime"`
	Available   bool          `json:"available"`
}

// DeduplicationStats reports the per-project hit rate of the
// telemetry deduplication cache. Captured on a sampling basis by the
// telemetry stream consumer.
type DeduplicationStats struct {
	ProjectID         uuid.UUID `json:"project_id"`
	TotalChecks       int64     `json:"total_checks"`
	CacheHits         int64     `json:"cache_hits"`
	CacheMisses       int64     `json:"cache_misses"`
	DatabaseFallbacks int64     `json:"database_fallbacks"`
	DuplicatesFound   int64     `json:"duplicates_found"`
	CacheHitRate      float64   `json:"cache_hit_rate"`
	FallbackRate      float64   `json:"fallback_rate"`
	AverageLatencyMs  float64   `json:"average_latency_ms"`
}

// TelemetryHealthStatus is the rolled-up health view of the telemetry
// ingestion pipeline (DB + Redis dedup + processing queue). Surfaced
// by /readyz and the ops dashboard.
type TelemetryHealthStatus struct {
	Database              *DatabaseHealth    `json:"database"`
	Redis                 *RedisHealthStatus `json:"redis"`
	ProcessingQueue       *QueueHealth       `json:"processing_queue"`
	ActiveWorkers         int                `json:"active_workers"`
	AverageProcessingTime float64            `json:"average_processing_time_ms"`
	ThroughputPerMinute   float64            `json:"throughput_per_minute"`
	ErrorRate             float64            `json:"error_rate"`
	Healthy               bool               `json:"healthy"`
}

// DatabaseHealth captures the ClickHouse pool's runtime connection
// state for the telemetry pipeline.
type DatabaseHealth struct {
	Connected         bool    `json:"connected"`
	LatencyMs         float64 `json:"latency_ms"`
	ActiveConnections int     `json:"active_connections"`
	MaxConnections    int     `json:"max_connections"`
}

// QueueHealth captures the per-queue Redis-stream state for the
// telemetry pipeline (size, processing rate, oldest-message age).
type QueueHealth struct {
	Size             int64   `json:"size"`
	ProcessingRate   float64 `json:"processing_rate"`
	AverageWaitTime  float64 `json:"average_wait_time_ms"`
	OldestMessageAge float64 `json:"oldest_message_age_ms"`
}

// TelemetryMetrics is the aggregate counter snapshot of the telemetry
// pipeline since the last reset. Consumed by /metrics and the ops
// dashboard's batch-processing panel.
type TelemetryMetrics struct {
	TotalBatches          int64   `json:"total_batches"`
	CompletedBatches      int64   `json:"completed_batches"`
	FailedBatches         int64   `json:"failed_batches"`
	ProcessingBatches     int64   `json:"processing_batches"`
	TotalEvents           int64   `json:"total_events"`
	ProcessedEvents       int64   `json:"processed_events"`
	FailedEvents          int64   `json:"failed_events"`
	DuplicateEvents       int64   `json:"duplicate_events"`
	AverageEventsPerBatch float64 `json:"average_events_per_batch"`
	ThroughputPerSecond   float64 `json:"throughput_per_second"`
	SuccessRate           float64 `json:"success_rate"`
	DeduplicationRate     float64 `json:"deduplication_rate"`
}

// TelemetryPerformanceStats reports the rolling-window latency and
// throughput stats for the telemetry pipeline (P95/P99, cache hit
// rate, retry rate) over a configurable TimeWindow.
type TelemetryPerformanceStats struct {
	TimeWindow           time.Duration `json:"time_window"`
	TotalRequests        int64         `json:"total_requests"`
	SuccessfulRequests   int64         `json:"successful_requests"`
	AverageLatencyMs     float64       `json:"average_latency_ms"`
	P95LatencyMs         float64       `json:"p95_latency_ms"`
	P99LatencyMs         float64       `json:"p99_latency_ms"`
	ThroughputPerSecond  float64       `json:"throughput_per_second"`
	PeakThroughput       float64       `json:"peak_throughput"`
	CacheHitRate         float64       `json:"cache_hit_rate"`
	DatabaseFallbackRate float64       `json:"database_fallback_rate"`
	ErrorRate            float64       `json:"error_rate"`
	RetryRate            float64       `json:"retry_rate"`
}
