-- ClickHouse Migration: create_llm_metrics_1hour
-- Created: 2025-10-11T23:36:00+05:30

-- Hour-level aggregations for trend analysis
CREATE TABLE IF NOT EXISTS llm_metrics_1hour (
    time_bucket DateTime,
    organization_id String,
    project_id String,
    provider String,
    model String,
    request_count UInt64,
    success_count UInt64,
    error_count UInt64,
    avg_latency Float32,
    p95_latency Float32,
    p99_latency Float32,
    max_latency UInt32,
    total_tokens UInt64,
    avg_tokens Float32,
    total_cost Float64,
    avg_cost Float64,
    cost_per_token Float64,
    avg_quality_score Float32,
    quality_scores_count UInt32,
    unique_traces UInt32,
    unique_users UInt32
) ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(time_bucket)
ORDER BY (time_bucket, organization_id, project_id, provider, model)
TTL toDateTime(time_bucket) + INTERVAL 30 DAY
SETTINGS index_granularity = 8192;
