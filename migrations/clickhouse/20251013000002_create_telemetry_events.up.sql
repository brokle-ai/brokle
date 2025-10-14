-- Create telemetry_events table for SDK observability event tracking
CREATE TABLE IF NOT EXISTS telemetry_events
(
    id              String COMMENT 'ULID identifier for the event',
    batch_id        String COMMENT 'ULID identifier for the parent batch',
    project_id      String COMMENT 'Project identifier',
    environment     String DEFAULT 'default' COMMENT 'Environment tag (production, staging, etc.)',
    event_type      LowCardinality(String) COMMENT 'Event type (trace_create, observation_create, etc.)',
    event_data      String COMMENT 'JSON event payload data',
    timestamp       DateTime64(3) COMMENT 'Event creation timestamp',
    retry_count     UInt8 DEFAULT 0 COMMENT 'Number of retry attempts',
    processed_at    DateTime64(3) COMMENT 'Event processing timestamp',
    INDEX idx_batch_id batch_id TYPE bloom_filter() GRANULARITY 4,
    INDEX idx_project_id project_id TYPE bloom_filter() GRANULARITY 4,
    INDEX idx_environment environment TYPE bloom_filter() GRANULARITY 4,
    INDEX idx_event_type event_type TYPE bloom_filter() GRANULARITY 4,
    INDEX idx_timestamp timestamp TYPE minmax GRANULARITY 1
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (project_id, timestamp, batch_id, id)
TTL toDateTime(timestamp) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192
COMMENT 'Telemetry event records for SDK observability with 90-day retention';
