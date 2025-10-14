-- Create telemetry_batches table for SDK observability batch tracking
CREATE TABLE IF NOT EXISTS telemetry_batches
(
    id                  String COMMENT 'ULID identifier for the batch',
    project_id          String COMMENT 'Project identifier',
    environment         String DEFAULT 'default' COMMENT 'Environment tag (production, staging, etc.)',
    status              LowCardinality(String) COMMENT 'Batch status (completed, failed, partial, processing)',
    total_events        UInt32 COMMENT 'Total number of events in batch',
    processed_events    UInt32 COMMENT 'Number of successfully processed events',
    failed_events       UInt32 COMMENT 'Number of failed events',
    processing_time_ms  UInt32 COMMENT 'Total processing time in milliseconds',
    metadata            String COMMENT 'JSON metadata about the batch',
    timestamp           DateTime64(3) COMMENT 'Batch creation timestamp',
    processed_at        DateTime64(3) COMMENT 'Batch processing completion timestamp',
    INDEX idx_project_id project_id TYPE bloom_filter() GRANULARITY 4,
    INDEX idx_environment environment TYPE bloom_filter() GRANULARITY 4,
    INDEX idx_status status TYPE bloom_filter() GRANULARITY 4,
    INDEX idx_timestamp timestamp TYPE minmax GRANULARITY 1
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (project_id, timestamp, id)
TTL toDateTime(timestamp) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192
COMMENT 'Telemetry batch records for SDK observability tracking with 90-day retention';
