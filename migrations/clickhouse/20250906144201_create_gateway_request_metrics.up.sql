-- ClickHouse Migration: create_gateway_request_metrics
-- Created: 2025-10-11T23:36:00+05:30

-- Gateway request metrics table for detailed request tracking
CREATE TABLE IF NOT EXISTS gateway_request_metrics (
    id String,
    request_id String,
    organization_id String,
    user_id Nullable(String),
    environment LowCardinality(String),
    provider_id String,
    provider_name LowCardinality(String),
    model_id String,
    model_name LowCardinality(String),
    request_type LowCardinality(String),
    method LowCardinality(String),
    endpoint String,
    status LowCardinality(String),
    status_code UInt16,
    duration UInt64, -- nanoseconds
    input_tokens UInt32,
    output_tokens UInt32,
    total_tokens UInt32,
    estimated_cost Float64,
    actual_cost Float64,
    currency LowCardinality(String),
    routing_reason String,
    cache_hit Bool,
    error String,
    metadata String,
    timestamp DateTime64(3)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (organization_id, timestamp, request_id)
TTL toDateTime(timestamp) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;
