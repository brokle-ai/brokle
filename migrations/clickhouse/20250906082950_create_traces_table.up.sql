-- Create traces table for distributed tracing
CREATE TABLE IF NOT EXISTS traces (
    timestamp DateTime64(3) DEFAULT now64(),
    trace_id String,
    span_id String,
    parent_span_id String DEFAULT '',
    operation_name String,
    duration_ms UInt32,
    status_code UInt16,
    tags Map(String, String),
    user_id String,
    organization_id String,
    created_date Date DEFAULT toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, trace_id, span_id)
TTL toDateTime(timestamp) + INTERVAL 30 DAY;

