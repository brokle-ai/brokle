-- ClickHouse Migration: create_provider_performance_metrics
-- Created: 2025-10-11T23:36:00+05:30

-- Provider performance aggregation table
CREATE TABLE IF NOT EXISTS provider_performance_metrics (
    date Date,
    hour UInt8,
    project_id String,
    provider String,
    model String,
    environment String DEFAULT 'production',
    -- Performance metrics
    total_requests UInt32,
    successful_requests UInt32,
    failed_requests UInt32,
    avg_latency_ms Float32,
    p95_latency_ms Float32,
    p99_latency_ms Float32,
    min_latency_ms UInt32,
    max_latency_ms UInt32,
    -- Token usage
    total_input_tokens UInt64,
    total_output_tokens UInt64,
    total_tokens UInt64,
    avg_tokens_per_request Float32,
    -- Cost metrics
    total_cost_usd Float64,
    avg_cost_per_request Float64,
    cost_per_1k_tokens Float64,
    -- Quality metrics
    avg_quality_score Nullable(Float32),
    cache_hit_rate Float32 DEFAULT 0.0,
    error_rate Float32,
    timeout_rate Float32,
    rate_limit_rate Float32,
    -- Routing metrics
    primary_selection_count UInt32 DEFAULT 0,
    fallback_selection_count UInt32 DEFAULT 0,
    load_balancing_weight Float32 DEFAULT 1.0
) ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, hour, project_id, provider, model)
TTL date + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;
