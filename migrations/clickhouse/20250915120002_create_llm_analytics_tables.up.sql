-- Create ClickHouse tables for LLM observability analytics and real-time metrics

-- Extend existing request_logs table with observability fields
ALTER TABLE request_logs ADD COLUMN trace_id String DEFAULT '' AFTER id;
ALTER TABLE request_logs ADD COLUMN observation_id String DEFAULT '' AFTER trace_id;
ALTER TABLE request_logs ADD COLUMN session_id String DEFAULT '' AFTER observation_id;

-- Create dedicated LLM analytics table for high-performance observability queries
CREATE TABLE IF NOT EXISTS llm_analytics (
    -- Timestamp and identifiers
    timestamp DateTime64(3) DEFAULT now64(),
    trace_id String,
    observation_id String,
    session_id String DEFAULT '',

    -- Organization hierarchy
    organization_id String,
    project_id String,
    user_id String DEFAULT '',

    -- Request classification
    provider LowCardinality(String),
    model LowCardinality(String),
    observation_type LowCardinality(String), -- 'llm', 'span', 'event', 'generation', etc.
    endpoint_type LowCardinality(String) DEFAULT '', -- 'chat', 'completion', 'embedding'

    -- Performance metrics
    latency_ms UInt32,
    prompt_tokens UInt32,
    completion_tokens UInt32,
    total_tokens UInt32,

    -- Cost tracking
    input_cost Float64,
    output_cost Float64,
    total_cost Float64,

    -- Quality metrics
    quality_score Float32,

    -- Status and error tracking
    status LowCardinality(String), -- 'success', 'error', 'timeout', 'cancelled'
    error_message String DEFAULT '',
    error_code String DEFAULT '',

    -- Routing and optimization
    routing_strategy LowCardinality(String) DEFAULT '',
    fallback_used UInt8 DEFAULT 0,
    cache_hit UInt8 DEFAULT 0,

    -- Additional metadata
    model_version String DEFAULT '',
    sdk_version String DEFAULT '',
    environment String DEFAULT 'production',

    -- Partitioning date
    created_date Date DEFAULT toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, organization_id, project_id)
SETTINGS index_granularity = 8192
TTL toDateTime(timestamp) + INTERVAL 365 DAY;

-- Create materialized view to automatically populate llm_analytics from request_logs
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_llm_analytics_from_request_logs
TO llm_analytics
AS SELECT
    timestamp,
    trace_id,
    observation_id,
    session_id,
    organization_id,
    project_id,
    user_id,
    provider,
    model,
    'llm' as observation_type,
    method as endpoint_type,
    latency,
    input_tokens as prompt_tokens,
    output_tokens as completion_tokens,
    total_tokens,
    CASE WHEN input_tokens > 0 THEN cost * (input_tokens / nullIf(total_tokens, 0)) ELSE 0 END as input_cost,
    CASE WHEN output_tokens > 0 THEN cost * (output_tokens / nullIf(total_tokens, 0)) ELSE 0 END as output_cost,
    cost as total_cost,
    quality_score,
    CASE
        WHEN error != '' THEN 'error'
        WHEN latency > 30000 THEN 'timeout'
        ELSE 'success'
    END as status,
    error as error_message,
    '' as error_code,
    '' as routing_strategy,
    0 as fallback_used,
    cached as cache_hit,
    '' as model_version,
    '' as sdk_version,
    environment,
    created_date
FROM request_logs
WHERE trace_id != '';

-- Create llm_quality_analytics table for quality score analytics
CREATE TABLE IF NOT EXISTS llm_quality_analytics (
    -- Timestamp and identifiers
    timestamp DateTime64(3) DEFAULT now64(),
    trace_id String,
    observation_id String,

    -- Organization hierarchy
    organization_id String,
    project_id String,

    -- Score information
    score_name LowCardinality(String),
    score_value Float64,
    string_value String DEFAULT '',
    data_type LowCardinality(String), -- 'NUMERIC', 'CATEGORICAL', 'BOOLEAN'

    -- Evaluation metadata
    score_source LowCardinality(String), -- 'auto', 'human', 'eval', 'api'
    evaluator_name String DEFAULT '',
    evaluator_version String DEFAULT '',

    -- Context information
    provider String DEFAULT '',
    model String DEFAULT '',
    observation_type String DEFAULT '',

    -- Partitioning date
    created_date Date DEFAULT toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, score_name, organization_id, project_id)
SETTINGS index_granularity = 8192
TTL toDateTime(timestamp) + INTERVAL 180 DAY;

-- Create llm_trace_analytics table for trace-level analytics
CREATE TABLE IF NOT EXISTS llm_trace_analytics (
    -- Timestamp and identifiers
    timestamp DateTime64(3) DEFAULT now64(),
    trace_id String,
    session_id String DEFAULT '',

    -- Organization hierarchy
    organization_id String,
    project_id String,
    user_id String DEFAULT '',

    -- Trace metadata
    trace_name String,
    total_observations UInt32,
    completed_observations UInt32,

    -- Performance aggregates
    total_latency_ms UInt32,
    max_latency_ms UInt32,
    min_latency_ms UInt32,
    avg_latency_ms Float32,

    -- Cost aggregates
    total_cost Float64,
    total_tokens UInt32,
    total_prompt_tokens UInt32,
    total_completion_tokens UInt32,

    -- Quality aggregates
    avg_quality_score Float32,
    min_quality_score Float32,
    max_quality_score Float32,
    quality_score_count UInt32,

    -- Provider and model distribution
    provider_list Array(String),
    model_list Array(String),
    observation_types Array(String),

    -- Status tracking
    error_count UInt32,
    success_count UInt32,

    -- Partitioning date
    created_date Date DEFAULT toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, organization_id, project_id)
SETTINGS index_granularity = 8192
TTL toDateTime(timestamp) + INTERVAL 180 DAY;

-- Create llm_provider_health table for provider health monitoring
CREATE TABLE IF NOT EXISTS llm_provider_health (
    -- Timestamp and identifiers
    timestamp DateTime64(3) DEFAULT now64(),

    -- Provider information
    provider LowCardinality(String),
    model String DEFAULT '',
    region String DEFAULT '',

    -- Health metrics
    success_rate Float32,
    average_latency Float32,
    p95_latency Float32,
    p99_latency Float32,
    error_rate Float32,
    timeout_rate Float32,

    -- Throughput metrics
    requests_per_minute Float32,
    requests_per_hour Float32,

    -- Cost metrics
    average_cost_per_request Float64,
    average_cost_per_token Float64,

    -- Sample size
    sample_size UInt32,
    evaluation_period_minutes UInt16,

    -- Health status
    health_score Float32, -- 0-1 composite health score
    is_healthy UInt8,

    -- Partitioning date
    created_date Date DEFAULT toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, provider, model)
SETTINGS index_granularity = 8192
TTL toDateTime(timestamp) + INTERVAL 90 DAY;

-- Create real-time aggregation tables for dashboard performance

-- Minute-level aggregations for real-time metrics
CREATE TABLE IF NOT EXISTS llm_metrics_1min (
    -- Time bucket
    time_bucket DateTime,

    -- Grouping dimensions
    organization_id String,
    project_id String,
    provider String,
    model String,

    -- Aggregated metrics
    request_count UInt64,
    success_count UInt64,
    error_count UInt64,

    -- Latency metrics
    avg_latency Float32,
    p95_latency Float32,
    max_latency UInt32,

    -- Token metrics
    total_tokens UInt64,
    avg_tokens Float32,

    -- Cost metrics
    total_cost Float64,
    avg_cost Float64,

    -- Quality metrics
    avg_quality_score Float32,
    quality_scores_count UInt32
) ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(time_bucket)
ORDER BY (time_bucket, organization_id, project_id, provider, model)
SETTINGS index_granularity = 8192
TTL toDateTime(time_bucket) + INTERVAL 7 DAY;

-- Hour-level aggregations for trend analysis
CREATE TABLE IF NOT EXISTS llm_metrics_1hour (
    -- Time bucket
    time_bucket DateTime,

    -- Grouping dimensions
    organization_id String,
    project_id String,
    provider String,
    model String,

    -- Aggregated metrics
    request_count UInt64,
    success_count UInt64,
    error_count UInt64,

    -- Latency metrics
    avg_latency Float32,
    p95_latency Float32,
    p99_latency Float32,
    max_latency UInt32,

    -- Token metrics
    total_tokens UInt64,
    avg_tokens Float32,

    -- Cost metrics
    total_cost Float64,
    avg_cost Float64,
    cost_per_token Float64,

    -- Quality metrics
    avg_quality_score Float32,
    quality_scores_count UInt32,

    -- Additional analysis metrics
    unique_traces UInt32,
    unique_users UInt32
) ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(time_bucket)
ORDER BY (time_bucket, organization_id, project_id, provider, model)
SETTINGS index_granularity = 8192
TTL toDateTime(time_bucket) + INTERVAL 30 DAY;

-- Create materialized views for automatic aggregations

-- 1-minute aggregations
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_llm_metrics_1min
TO llm_metrics_1min
AS SELECT
    toStartOfMinute(timestamp) as time_bucket,
    organization_id,
    project_id,
    provider,
    model,
    count() as request_count,
    countIf(status = 'success') as success_count,
    countIf(status = 'error') as error_count,
    avg(latency_ms) as avg_latency,
    quantile(0.95)(latency_ms) as p95_latency,
    max(latency_ms) as max_latency,
    sum(total_tokens) as total_tokens,
    avg(total_tokens) as avg_tokens,
    sum(total_cost) as total_cost,
    avg(total_cost) as avg_cost,
    avg(quality_score) as avg_quality_score,
    countIf(quality_score > 0) as quality_scores_count
FROM llm_analytics
WHERE timestamp >= now() - INTERVAL 1 HOUR
GROUP BY time_bucket, organization_id, project_id, provider, model;

-- 1-hour aggregations
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_llm_metrics_1hour
TO llm_metrics_1hour
AS SELECT
    toStartOfHour(timestamp) as time_bucket,
    organization_id,
    project_id,
    provider,
    model,
    count() as request_count,
    countIf(status = 'success') as success_count,
    countIf(status = 'error') as error_count,
    avg(latency_ms) as avg_latency,
    quantile(0.95)(latency_ms) as p95_latency,
    quantile(0.99)(latency_ms) as p99_latency,
    max(latency_ms) as max_latency,
    sum(total_tokens) as total_tokens,
    avg(total_tokens) as avg_tokens,
    sum(total_cost) as total_cost,
    avg(total_cost) as avg_cost,
    sum(total_cost) / nullIf(sum(total_tokens), 0) as cost_per_token,
    avg(quality_score) as avg_quality_score,
    countIf(quality_score > 0) as quality_scores_count,
    uniqExact(trace_id) as unique_traces,
    uniqExact(user_id) as unique_users
FROM llm_analytics
WHERE timestamp >= now() - INTERVAL 24 HOUR
GROUP BY time_bucket, organization_id, project_id, provider, model;

-- Create indexes for optimal query performance
-- Note: ClickHouse uses ORDER BY as primary index, additional indexes are for specific use cases

-- Create ttl expressions for data lifecycle management
-- Analytics data: 1 year retention
-- Quality analytics: 6 months retention
-- Trace analytics: 6 months retention
-- Provider health: 3 months retention
-- Real-time metrics: 1 week (1min) and 1 month (1hour) retention

-- Add table comments for documentation
-- Note: ClickHouse doesn't support table comments in older versions, these are for reference