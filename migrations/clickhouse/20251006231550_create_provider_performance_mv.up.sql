-- ClickHouse Migration: create_provider_performance_mv
-- Created: 2025-10-12T00:00:00+05:30

-- Materialized view for provider performance metrics
-- Populates provider_performance_metrics (SummingMergeTree) with raw data
-- The engine will automatically aggregate the metrics
CREATE MATERIALIZED VIEW IF NOT EXISTS provider_performance_mv
TO provider_performance_metrics
AS SELECT
    toDate(timestamp) as date,
    toHour(timestamp) as hour,
    project_id,
    provider,
    model,
    environment,
    1 as total_requests,
    if(status_code < 400, 1, 0) as successful_requests,
    if(status_code >= 400, 1, 0) as failed_requests,
    toFloat32(provider_latency_ms) as avg_latency_ms,
    toFloat32(provider_latency_ms) as p95_latency_ms,
    toFloat32(provider_latency_ms) as p99_latency_ms,
    provider_latency_ms as min_latency_ms,
    provider_latency_ms as max_latency_ms,
    toUInt64(input_tokens) as total_input_tokens,
    toUInt64(output_tokens) as total_output_tokens,
    toUInt64(total_tokens) as total_tokens,
    toFloat32(total_tokens) as avg_tokens_per_request,
    cost_usd as total_cost_usd,
    cost_usd as avg_cost_per_request,
    if(total_tokens > 0, cost_usd / (toFloat64(total_tokens) / 1000.0), 0.0) as cost_per_1k_tokens,
    quality_score as avg_quality_score,
    if(cache_hit, 1.0, 0.0) as cache_hit_rate,
    if(status_code >= 400, 1.0, 0.0) as error_rate,
    if(status_code = 408, 1.0, 0.0) as timeout_rate,
    if(status_code = 429, 1.0, 0.0) as rate_limit_rate,
    if(fallback_triggered = false, 1, 0) as primary_selection_count,
    if(fallback_triggered = true, 1, 0) as fallback_selection_count,
    1.0 as load_balancing_weight
FROM request_logs
WHERE provider != '' AND model != '';
