-- ClickHouse Migration: create_llm_metrics_1hour_mv
-- Created: 2025-10-12T00:00:00+05:30

-- Materialized view for 1-hour LLM metrics aggregation
-- Populates llm_metrics_1hour (SummingMergeTree) with raw data
-- The engine will automatically sum the metrics
CREATE MATERIALIZED VIEW IF NOT EXISTS llm_metrics_1hour_mv
TO llm_metrics_1hour
AS SELECT
    toStartOfHour(timestamp) as time_bucket,
    organization_id,
    project_id,
    provider,
    model,
    toUInt64(1) as request_count,
    toUInt64(if(status = 'success', 1, 0)) as success_count,
    toUInt64(if(status = 'error', 1, 0)) as error_count,
    toFloat32(latency_ms) as avg_latency,
    toFloat32(latency_ms) as p95_latency,
    toFloat32(latency_ms) as p99_latency,
    latency_ms as max_latency,
    toUInt64(total_tokens) as total_tokens,
    toFloat32(total_tokens) as avg_tokens,
    total_cost,
    total_cost as avg_cost,
    if(total_tokens > 0, total_cost / toFloat64(total_tokens), 0.0) as cost_per_token,
    toFloat32(quality_score) as avg_quality_score,
    toUInt32(if(quality_score > 0, 1, 0)) as quality_scores_count,
    toUInt32(1) as unique_traces,
    toUInt32(if(user_id != '', 1, 0)) as unique_users
FROM llm_analytics;
