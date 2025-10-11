-- ClickHouse Migration: create_llm_analytics
-- Created: 2025-10-11T23:36:00+05:30

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
TTL toDateTime(timestamp) + INTERVAL 365 DAY
SETTINGS index_granularity = 8192;
