-- Create metrics table for real-time platform metrics
CREATE TABLE IF NOT EXISTS metrics (
    timestamp DateTime64(3) DEFAULT now64(),
    metric_name LowCardinality(String),
    metric_value Float64,
    labels Map(String, String),
    organization_id String,
    project_id String,
    user_id String,
    provider String DEFAULT '',
    model String DEFAULT '',
    environment String DEFAULT 'production',
    created_date Date DEFAULT toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, metric_name, organization_id)
TTL toDateTime(timestamp) + INTERVAL 90 DAY;
