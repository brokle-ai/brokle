-- ClickHouse Migration: create_ai_routing_metrics
-- Created: 2025-10-11T23:36:00+05:30

-- AI routing metrics table with detailed tracking
CREATE TABLE IF NOT EXISTS ai_routing_metrics (
    id String,
    timestamp DateTime64(3),
    project_id String,
    environment String DEFAULT 'production',
    model_requested String,
    model_used String,
    provider_selected String,
    routing_strategy String,
    routing_reason String,
    latency_ms UInt32,
    cost_usd Float64,
    input_tokens UInt32,
    output_tokens UInt32,
    total_tokens UInt32,
    success Bool,
    error_code Nullable(String),
    error_message Nullable(String),
    -- Enhanced routing metrics
    fallback_triggered Bool DEFAULT false,
    primary_provider String DEFAULT '',
    fallback_provider Nullable(String),
    providers_tried Array(String) DEFAULT [],
    provider_latencies Array(UInt32) DEFAULT [],
    provider_costs Array(Float64) DEFAULT [],
    quality_scores Array(Float32) DEFAULT [],
    load_balancing_weight Float32 DEFAULT 1.0,
    cache_hit Bool DEFAULT false,
    cache_key Nullable(String),
    user_tier Nullable(String),
    request_priority UInt8 DEFAULT 5,
    rate_limit_applied Bool DEFAULT false,
    quota_remaining Nullable(UInt32),
    metadata String DEFAULT '{}'
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (project_id, provider_selected, timestamp)
TTL toDateTime(timestamp) + INTERVAL 365 DAY
SETTINGS index_granularity = 8192;
