-- ClickHouse Migration: create_gateway_usage_daily_mv
-- Created: 2025-10-12T00:00:00+05:30

-- Materialized view for daily gateway usage aggregation
-- Populates gateway_usage_metrics table with daily data
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
    1 as request_count,
    if(status_code < 400, 1, 0) as success_count,
    if(status_code >= 400, 1, 0) as error_count,
    input_tokens as total_input_tokens,
    output_tokens as total_output_tokens,
    total_tokens,
    actual_cost as total_cost,
    currency,
    duration / 1000000.0 as avg_duration,
    duration as min_duration,
    duration as max_duration,
    if(cache_hit, 1.0, 0.0) as cache_hit_rate,
    now() as timestamp
FROM gateway_request_metrics;
