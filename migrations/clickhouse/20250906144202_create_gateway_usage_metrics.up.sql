-- ClickHouse Migration: create_gateway_usage_metrics
-- Created: 2025-10-11T23:36:00+05:30

-- Gateway usage metrics table for aggregated usage statistics
CREATE TABLE IF NOT EXISTS gateway_usage_metrics (
    id String,
    organization_id String,
    environment LowCardinality(String),
    provider_id String,
    model_id String,
    request_type LowCardinality(String),
    period LowCardinality(String), -- hourly, daily, monthly
    period_start DateTime,
    period_end DateTime,
    request_count UInt64,
    success_count UInt64,
    error_count UInt64,
    total_input_tokens UInt64,
    total_output_tokens UInt64,
    total_tokens UInt64,
    total_cost Float64,
    currency LowCardinality(String),
    avg_duration Float64, -- average duration in milliseconds
    min_duration UInt64, -- nanoseconds
    max_duration UInt64, -- nanoseconds
    cache_hit_rate Float64,
    timestamp DateTime64(3)
) ENGINE = ReplacingMergeTree(timestamp)
PARTITION BY (toYYYYMM(period_start), period)
ORDER BY (organization_id, provider_id, model_id, period_start, period)
TTL period_start + INTERVAL 180 DAY
SETTINGS index_granularity = 8192;
