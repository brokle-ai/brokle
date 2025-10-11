-- ClickHouse Migration: create_request_logs
-- Created: 2025-10-11T23:36:00+05:30

-- Request logs table with AI gateway metrics
CREATE TABLE IF NOT EXISTS request_logs (
    id String,
    timestamp DateTime64(3),
    project_id String,
    environment String,
    user_id Nullable(String),
    api_key_id Nullable(String),
    method String,
    path String,
    status_code UInt16,
    response_time_ms UInt32,
    request_size_bytes UInt32,
    response_size_bytes UInt32,
    user_agent Nullable(String),
    ip_address Nullable(String),
    error_message Nullable(String),
    -- Gateway-specific columns
    provider String DEFAULT '',
    model String DEFAULT '',
    input_tokens UInt32 DEFAULT 0,
    output_tokens UInt32 DEFAULT 0,
    total_tokens UInt32 DEFAULT 0,
    cost_usd Float64 DEFAULT 0.0,
    routing_strategy String DEFAULT '',
    cache_hit Bool DEFAULT false,
    provider_latency_ms UInt32 DEFAULT 0,
    fallback_triggered Bool DEFAULT false,
    primary_provider Nullable(String),
    fallback_provider Nullable(String),
    quality_score Nullable(Float32),
    request_hash Nullable(String),
    completion_id Nullable(String),
    metadata String DEFAULT '{}'
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (project_id, timestamp)
TTL toDateTime(timestamp) + INTERVAL 60 DAY
SETTINGS index_granularity = 8192;
