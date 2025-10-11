-- ClickHouse Migration: create_llm_analytics_mv
-- Created: 2025-10-12T00:00:00+05:30

-- Materialized view to automatically populate llm_analytics from request_logs
-- Maps AI Gateway request logs to LLM analytics table for observability
CREATE MATERIALIZED VIEW IF NOT EXISTS llm_analytics_mv
TO llm_analytics
AS SELECT
    -- Timestamp and identifiers
    timestamp,
    '' as trace_id,  -- Will be populated by SDK telemetry later
    '' as observation_id,  -- Will be populated by SDK telemetry later
    '' as session_id,

    -- Organization hierarchy
    project_id as organization_id,  -- Using project_id as org_id for now
    project_id,
    ifNull(user_id, '') as user_id,

    -- Request classification
    provider,
    model,
    'llm' as observation_type,  -- Default type for gateway requests
    '' as endpoint_type,  -- Can be enhanced later based on path

    -- Performance metrics
    response_time_ms as latency_ms,
    input_tokens as prompt_tokens,
    output_tokens as completion_tokens,
    total_tokens,

    -- Cost tracking
    0.0 as input_cost,  -- Can be calculated later based on tokens
    0.0 as output_cost,  -- Can be calculated later based on tokens
    cost_usd as total_cost,

    -- Quality metrics
    ifNull(quality_score, 0.0) as quality_score,

    -- Status and error tracking
    if(status_code < 400, 'success', 'error') as status,
    ifNull(error_message, '') as error_message,
    '' as error_code,

    -- Routing and optimization
    routing_strategy,
    if(fallback_triggered, 1, 0) as fallback_used,
    if(cache_hit, 1, 0) as cache_hit,

    -- Additional metadata
    '' as model_version,
    '' as sdk_version,
    environment,

    -- Partitioning date
    toDate(timestamp) as created_date
FROM request_logs
WHERE provider != '' AND model != '';  -- Only process AI Gateway requests
