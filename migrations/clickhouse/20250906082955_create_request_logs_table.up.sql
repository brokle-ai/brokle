-- Create request_logs table for API request logging
CREATE TABLE IF NOT EXISTS request_logs (
    timestamp DateTime64(3) DEFAULT now64(),
    request_id String,
    method LowCardinality(String),
    path String,
    status_code UInt16,
    duration_ms UInt32,
    request_size UInt32,
    response_size UInt32,
    user_id String,
    organization_id String,
    ip_address IPv4,
    user_agent String,
    created_date Date DEFAULT toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, status_code, organization_id)
TTL toDateTime(timestamp) + INTERVAL 60 DAY;

