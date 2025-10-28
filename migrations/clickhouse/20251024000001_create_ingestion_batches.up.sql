-- Ingestion batches tracking table
CREATE TABLE IF NOT EXISTS ingestion_batches (
    id String,
    project_id String,
    status LowCardinality(String),
    total_events UInt32,
    processed_events UInt32,
    failed_events UInt32,
    processing_time_ms UInt32,
    metadata String,
    timestamp DateTime64(3),
    processed_at DateTime64(3)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (project_id, timestamp, id)
TTL toDateTime(timestamp) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;
