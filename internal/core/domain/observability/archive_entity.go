package observability

import "time"

// RawTelemetryRecord represents a raw telemetry event for S3 archival.
// Stores full event JSON for replay capability and vendor migration.
type RawTelemetryRecord struct {
	RecordID    string    `parquet:"record_id" json:"record_id"`
	ProjectID   string    `parquet:"project_id" json:"project_id"`
	SignalType  string    `parquet:"signal_type" json:"signal_type"` // traces, metrics, logs, genai
	Timestamp   time.Time `parquet:"timestamp,timestamp(microsecond)" json:"timestamp"`
	TraceID     string    `parquet:"trace_id" json:"trace_id"`
	SpanID      string    `parquet:"span_id" json:"span_id"`
	SpanJSONRaw string    `parquet:"span_json_raw" json:"span_json_raw"` // Full event as JSON (sufficient for replay)
	ArchivedAt  time.Time `parquet:"archived_at,timestamp(microsecond)" json:"archived_at"`
}

// SignalType constants for archive records
const (
	SignalTypeTraces  = "traces"
	SignalTypeMetrics = "metrics"
	SignalTypeLogs    = "logs"
	SignalTypeGenAI   = "genai"
)

// EntityType constants for blob_storage_file_log
const (
	EntityTypeArchiveBatch = "archive_batch"
)

// ArchiveBatchResult represents the result of archiving a batch to S3.
type ArchiveBatchResult struct {
	S3Path        string    `json:"s3_path"`
	BucketName    string    `json:"bucket_name"`
	RecordCount   int       `json:"record_count"`
	FileSizeBytes int64     `json:"file_size_bytes"`
	ArchivedAt    time.Time `json:"archived_at"`
}
