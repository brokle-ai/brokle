-- ClickHouse Migration: create_llm_trace_analytics
-- Created: 2025-10-11T23:36:00+05:30

-- Create llm_trace_analytics table for trace-level analytics
CREATE TABLE IF NOT EXISTS llm_trace_analytics (
    timestamp DateTime64(3) DEFAULT now64(),
    trace_id String,
    session_id String DEFAULT '',
    organization_id String,
    project_id String,
    user_id String DEFAULT '',
    trace_name String,
    total_observations UInt32,
    completed_observations UInt32,
    total_latency_ms UInt32,
    max_latency_ms UInt32,
    min_latency_ms UInt32,
    avg_latency_ms Float32,
    total_cost Float64,
    total_tokens UInt32,
    total_prompt_tokens UInt32,
    total_completion_tokens UInt32,
    avg_quality_score Float32,
    min_quality_score Float32,
    max_quality_score Float32,
    quality_score_count UInt32,
    provider_list Array(String),
    model_list Array(String),
    observation_types Array(String),
    error_count UInt32,
    success_count UInt32,
    created_date Date DEFAULT toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, organization_id, project_id)
TTL toDateTime(timestamp) + INTERVAL 180 DAY
SETTINGS index_granularity = 8192;
