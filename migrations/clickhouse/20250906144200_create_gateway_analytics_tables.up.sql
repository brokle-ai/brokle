-- Create ClickHouse tables for gateway analytics
-- Migration: 20250906144200_create_gateway_analytics_tables.up.sql

-- Gateway request metrics table for detailed request tracking
CREATE TABLE IF NOT EXISTS gateway_request_metrics (
    id String,
    request_id String,
    organization_id String,
    user_id Nullable(String),
    environment LowCardinality(String),
    provider_id String,
    provider_name LowCardinality(String),
    model_id String,
    model_name LowCardinality(String),
    request_type LowCardinality(String),
    method LowCardinality(String),
    endpoint String,
    status LowCardinality(String),
    status_code UInt16,
    duration UInt64, -- nanoseconds
    input_tokens UInt32,
    output_tokens UInt32,
    total_tokens UInt32,
    estimated_cost Float64,
    actual_cost Float64,
    currency LowCardinality(String),
    routing_reason String,
    cache_hit Bool,
    error String,
    metadata String,
    timestamp DateTime64(3)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (organization_id, timestamp, request_id)
TTL timestamp + INTERVAL 90 DAY -- Keep data for 90 days
SETTINGS index_granularity = 8192;

-- Gateway usage metrics table for aggregated usage statistics
CREATE TABLE IF NOT EXISTS gateway_usage_metrics (
    id String,
    organization_id String,
    environment LowCardinality(String),
    provider_id String,
    model_id String,
    request_type LowCardinality(String),
    period LowCardinality(String), -- hourly, daily, monthly
    period_start DateTime,
    period_end DateTime,
    request_count UInt64,
    success_count UInt64,
    error_count UInt64,
    total_input_tokens UInt64,
    total_output_tokens UInt64,
    total_tokens UInt64,
    total_cost Float64,
    currency LowCardinality(String),
    avg_duration Float64, -- average duration in milliseconds
    min_duration UInt64, -- nanoseconds
    max_duration UInt64, -- nanoseconds
    cache_hit_rate Float64,
    timestamp DateTime64(3)
) ENGINE = ReplacingMergeTree(timestamp)
PARTITION BY (toYYYYMM(period_start), period)
ORDER BY (organization_id, provider_id, model_id, period_start, period)
TTL period_start + INTERVAL 180 DAY -- Keep aggregated data for 180 days
SETTINGS index_granularity = 8192;

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
TTL timestamp + INTERVAL 365 DAY -- Keep cost data for 1 year
SETTINGS index_granularity = 8192;

-- Create materialized views for real-time aggregations

-- Hourly usage aggregation
CREATE MATERIALIZED VIEW IF NOT EXISTS gateway_usage_hourly_mv
TO gateway_usage_metrics
AS SELECT
    generateUUIDv4() as id,
    organization_id,
    environment,
    provider_id,
    model_id,
    request_type,
    'hourly' as period,
    toStartOfHour(timestamp) as period_start,
    toStartOfHour(timestamp) + INTERVAL 1 HOUR as period_end,
    count() as request_count,
    countIf(status_code < 400) as success_count,
    countIf(status_code >= 400) as error_count,
    sum(input_tokens) as total_input_tokens,
    sum(output_tokens) as total_output_tokens,
    sum(total_tokens) as total_tokens,
    sum(actual_cost) as total_cost,
    any(currency) as currency,
    avg(duration) / 1000000.0 as avg_duration, -- convert to milliseconds
    min(duration) as min_duration,
    max(duration) as max_duration,
    countIf(cache_hit = true) / count() as cache_hit_rate,
    now() as timestamp
FROM gateway_request_metrics
WHERE timestamp >= now() - INTERVAL 2 HOUR -- Process recent data
GROUP BY 
    organization_id,
    environment, 
    provider_id,
    model_id,
    request_type,
    toStartOfHour(timestamp);

-- Daily usage aggregation
CREATE MATERIALIZED VIEW IF NOT EXISTS gateway_usage_daily_mv
TO gateway_usage_metrics
AS SELECT
    generateUUIDv4() as id,
    organization_id,
    environment,
    provider_id,
    model_id,
    request_type,
    'daily' as period,
    toStartOfDay(timestamp) as period_start,
    toStartOfDay(timestamp) + INTERVAL 1 DAY as period_end,
    count() as request_count,
    countIf(status_code < 400) as success_count,
    countIf(status_code >= 400) as error_count,
    sum(input_tokens) as total_input_tokens,
    sum(output_tokens) as total_output_tokens,
    sum(total_tokens) as total_tokens,
    sum(actual_cost) as total_cost,
    any(currency) as currency,
    avg(duration) / 1000000.0 as avg_duration, -- convert to milliseconds
    min(duration) as min_duration,
    max(duration) as max_duration,
    countIf(cache_hit = true) / count() as cache_hit_rate,
    now() as timestamp
FROM gateway_request_metrics
WHERE timestamp >= now() - INTERVAL 25 HOUR -- Process recent data
GROUP BY 
    organization_id,
    environment,
    provider_id,
    model_id,
    request_type,
    toStartOfDay(timestamp);

-- Add indexes for better query performance
-- Note: ClickHouse doesn't have traditional indexes, but we can add more ORDER BY dimensions

-- Create dictionary for provider/model lookups (optional optimization)
-- This would be populated from PostgreSQL gateway_providers and gateway_models tables
-- CREATE DICTIONARY IF NOT EXISTS gateway_providers_dict (
--     id String,
--     name String,
--     type String
-- )
-- PRIMARY KEY id
-- SOURCE(POSTGRESQL(
--     host 'postgres'
--     port 5432
--     user 'brokle'
--     password 'password'
--     db 'brokle'
--     table 'gateway_providers'
-- ))
-- LIFETIME(MIN 300 MAX 3600)
-- LAYOUT(HASHED());

-- Add comments for documentation
-- Note: ClickHouse doesn't support table/column comments in the same way as PostgreSQL