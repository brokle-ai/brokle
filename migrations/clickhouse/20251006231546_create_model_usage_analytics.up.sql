-- ClickHouse Migration: create_model_usage_analytics
-- Created: 2025-10-11T23:36:00+05:30

-- Model usage analytics table
CREATE TABLE IF NOT EXISTS model_usage_analytics (
    date Date,
    project_id String,
    model String,
    provider String,
    environment String DEFAULT 'production',
    -- Usage statistics
    request_count UInt32,
    unique_users UInt32,
    total_input_tokens UInt64,
    total_output_tokens UInt64,
    total_cost_usd Float64,
    avg_cost_per_request Float64,
    -- Performance metrics
    avg_response_time_ms Float32,
    p95_response_time_ms Float32,
    success_rate Float32,
    cache_hit_rate Float32,
    -- Quality metrics
    avg_quality_score Nullable(Float32)
) ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, project_id, model, provider)
TTL date + INTERVAL 180 DAY
SETTINGS index_granularity = 8192;
