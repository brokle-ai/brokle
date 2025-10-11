-- AI Gateway ClickHouse Analytics Migration
-- Enhances existing analytics tables with gateway-specific tracking

-- Check if request_logs table exists and add gateway columns
-- Note: ClickHouse doesn't support traditional ALTER TABLE ADD COLUMN with NOT NULL constraints on existing data
-- We'll create a new table structure and handle data migration

-- Enhance request_logs table with gateway analytics
CREATE TABLE IF NOT EXISTS request_logs_gateway (
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
    -- Gateway-specific columns (NEW)
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
    request_hash Nullable(String), -- For caching
    completion_id Nullable(String), -- Provider's completion ID
    metadata String DEFAULT '{}' -- JSON string for additional data
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (project_id, timestamp)
TTL timestamp + INTERVAL 60 DAY
SETTINGS index_granularity = 8192;

-- Enhance ai_routing_metrics table with more detailed tracking
CREATE TABLE IF NOT EXISTS ai_routing_metrics_enhanced (
    id String,
    timestamp DateTime64(3),
    project_id String,
    environment String DEFAULT 'production',
    model_requested String,
    model_used String,
    provider_selected String,
    routing_strategy String,
    routing_reason String, -- Why this provider was selected
    latency_ms UInt32,
    cost_usd Float64,
    input_tokens UInt32,
    output_tokens UInt32,
    total_tokens UInt32,
    success Bool,
    error_code Nullable(String),
    error_message Nullable(String),
    -- Enhanced routing metrics (NEW)
    fallback_triggered Bool DEFAULT false,
    primary_provider String DEFAULT '',
    fallback_provider Nullable(String),
    providers_tried Array(String) DEFAULT [],
    provider_latencies Array(UInt32) DEFAULT [], -- Latency for each provider tried
    provider_costs Array(Float64) DEFAULT [], -- Cost estimates for each provider
    quality_scores Array(Float32) DEFAULT [], -- Quality scores for each provider
    load_balancing_weight Float32 DEFAULT 1.0,
    cache_hit Bool DEFAULT false,
    cache_key Nullable(String),
    user_tier Nullable(String), -- For routing decisions based on user tier
    request_priority UInt8 DEFAULT 5, -- 1-10 priority scale
    rate_limit_applied Bool DEFAULT false,
    quota_remaining Nullable(UInt32),
    metadata String DEFAULT '{}' -- JSON string for additional routing metadata
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (project_id, provider_selected, timestamp)
TTL timestamp + INTERVAL 365 DAY -- Keep routing data longer for analytics
SETTINGS index_granularity = 8192;

-- Create provider performance aggregation table
CREATE TABLE IF NOT EXISTS provider_performance_metrics (
    date Date,
    hour UInt8, -- 0-23
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
TTL date + INTERVAL 90 DAY -- Keep hourly aggregations for 90 days
SETTINGS index_granularity = 8192;

-- Create materialized view to populate provider performance metrics
CREATE MATERIALIZED VIEW IF NOT EXISTS provider_performance_mv
TO provider_performance_metrics AS
SELECT
    toDate(timestamp) as date,
    toHour(timestamp) as hour,
    project_id,
    provider,
    model,
    environment,
    count() as total_requests,
    sum(if(status_code < 400, 1, 0)) as successful_requests,
    sum(if(status_code >= 400, 1, 0)) as failed_requests,
    avg(provider_latency_ms) as avg_latency_ms,
    quantile(0.95)(provider_latency_ms) as p95_latency_ms,
    quantile(0.99)(provider_latency_ms) as p99_latency_ms,
    min(provider_latency_ms) as min_latency_ms,
    max(provider_latency_ms) as max_latency_ms,
    sum(input_tokens) as total_input_tokens,
    sum(output_tokens) as total_output_tokens,
    sum(total_tokens) as total_tokens,
    avg(total_tokens) as avg_tokens_per_request,
    sum(cost_usd) as total_cost_usd,
    avg(cost_usd) as avg_cost_per_request,
    if(sum(total_tokens) > 0, sum(cost_usd) / (sum(total_tokens) / 1000.0), 0) as cost_per_1k_tokens,
    avg(quality_score) as avg_quality_score,
    avg(if(cache_hit, 1, 0)) as cache_hit_rate,
    avg(if(status_code >= 400, 1, 0)) as error_rate,
    avg(if(status_code = 408, 1, 0)) as timeout_rate,
    avg(if(status_code = 429, 1, 0)) as rate_limit_rate,
    sum(if(fallback_triggered = false, 1, 0)) as primary_selection_count,
    sum(if(fallback_triggered = true, 1, 0)) as fallback_selection_count,
    1.0 as load_balancing_weight
FROM request_logs_gateway
WHERE provider != '' AND model != ''
GROUP BY date, hour, project_id, provider, model, environment;

-- Create model usage analytics table
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
TTL date + INTERVAL 180 DAY -- Keep daily aggregations for 6 months
SETTINGS index_granularity = 8192;

-- Create cost analytics table for billing
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
TTL date + INTERVAL 730 DAY -- Keep cost data for 2 years for billing
SETTINGS index_granularity = 8192;

-- Add indexes for common query patterns
-- Note: ClickHouse doesn't have traditional indexes, but we can optimize with proper ORDER BY and PARTITION BY

-- Comments for documentation
-- request_logs_gateway: Enhanced request logs with AI gateway metrics
-- ai_routing_metrics_enhanced: Detailed routing decisions and performance
-- provider_performance_metrics: Hourly aggregated provider performance data
-- model_usage_analytics: Daily model usage analytics for insights
-- cost_analytics: Detailed cost tracking for billing and optimization