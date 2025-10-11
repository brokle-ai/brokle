-- ClickHouse Migration: create_llm_provider_health
-- Created: 2025-10-11T23:36:00+05:30

-- Create llm_provider_health table for provider health monitoring
CREATE TABLE IF NOT EXISTS llm_provider_health (
    timestamp DateTime64(3) DEFAULT now64(),
    provider LowCardinality(String),
    model String DEFAULT '',
    region String DEFAULT '',
    success_rate Float32,
    average_latency Float32,
    p95_latency Float32,
    p99_latency Float32,
    error_rate Float32,
    timeout_rate Float32,
    requests_per_minute Float32,
    requests_per_hour Float32,
    average_cost_per_request Float64,
    average_cost_per_token Float64,
    sample_size UInt32,
    evaluation_period_minutes UInt16,
    health_score Float32, -- 0-1 composite health score
    is_healthy UInt8,
    created_date Date DEFAULT toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, provider, model)
TTL toDateTime(timestamp) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;
