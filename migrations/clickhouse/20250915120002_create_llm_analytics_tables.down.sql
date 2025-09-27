-- Drop ClickHouse tables and views for LLM observability analytics

-- Drop materialized views first
DROP VIEW IF EXISTS mv_llm_metrics_1hour;
DROP VIEW IF EXISTS mv_llm_metrics_1min;
DROP VIEW IF EXISTS mv_llm_analytics_from_request_logs;

-- Drop aggregation tables
DROP TABLE IF EXISTS llm_metrics_1hour;
DROP TABLE IF EXISTS llm_metrics_1min;

-- Drop analytics tables
DROP TABLE IF EXISTS llm_provider_health;
DROP TABLE IF EXISTS llm_trace_analytics;
DROP TABLE IF EXISTS llm_quality_analytics;
DROP TABLE IF EXISTS llm_analytics;

-- Remove columns added to existing request_logs table
ALTER TABLE request_logs DROP COLUMN IF EXISTS session_id;
ALTER TABLE request_logs DROP COLUMN IF EXISTS observation_id;
ALTER TABLE request_logs DROP COLUMN IF EXISTS trace_id;