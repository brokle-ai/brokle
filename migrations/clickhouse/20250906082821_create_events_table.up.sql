-- Create events table for business and system events
CREATE TABLE IF NOT EXISTS events (
    timestamp DateTime64(3) DEFAULT now64(),
    event_id String,
    event_type LowCardinality(String),
    event_data String,
    user_id String,
    organization_id String,
    project_id String,
    source String DEFAULT 'api',
    severity LowCardinality(String) DEFAULT 'info',
    created_date Date DEFAULT toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, event_type, organization_id)
TTL toDateTime(timestamp) + INTERVAL 180 DAY;

