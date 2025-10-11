-- ClickHouse Migration: create_gateway_cost_metrics
-- Created: 2025-10-11T23:36:00+05:30

-- Gateway cost metrics table for billing and cost tracking
CREATE TABLE IF NOT EXISTS gateway_cost_metrics (
    id String,
    request_id String,
    organization_id String,
    environment LowCardinality(String),
    provider_id String,
    model_id String,
    request_type LowCardinality(String),
    input_tokens UInt32,
    output_tokens UInt32,
    total_tokens UInt32,
    input_cost Float64,
    output_cost Float64,
    total_cost Float64,
    estimated_cost Float64,
    cost_difference Float64, -- actual - estimated
    currency LowCardinality(String),
    billing_tier LowCardinality(String),
    discount_applied Float64,
    timestamp DateTime64(3)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (organization_id, timestamp, request_id)
TTL toDateTime(timestamp) + INTERVAL 365 DAY
SETTINGS index_granularity = 8192;
