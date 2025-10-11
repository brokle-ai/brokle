-- ClickHouse Migration: create_cost_analytics
-- Created: 2025-10-11T23:36:00+05:30

-- Cost analytics table for billing
CREATE TABLE IF NOT EXISTS cost_analytics (
    date Date,
    hour UInt8,
    project_id String,
    environment String DEFAULT 'production',
    provider String,
    model String,
    -- Cost breakdown
    total_cost_usd Float64,
    input_cost_usd Float64,
    output_cost_usd Float64,
    cache_savings_usd Float64 DEFAULT 0.0,
    -- Token usage
    total_input_tokens UInt64,
    total_output_tokens UInt64,
    cached_tokens UInt64 DEFAULT 0,
    -- Request counts
    total_requests UInt32,
    cached_requests UInt32 DEFAULT 0,
    failed_requests UInt32
) ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, hour, project_id, provider)
TTL date + INTERVAL 730 DAY
SETTINGS index_granularity = 8192;
