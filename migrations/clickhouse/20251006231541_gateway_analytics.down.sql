-- AI Gateway ClickHouse Analytics Down Migration
-- Drops all gateway analytics tables and materialized views

-- Drop materialized view first
DROP VIEW IF EXISTS provider_performance_mv;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS cost_analytics;
DROP TABLE IF EXISTS model_usage_analytics;
DROP TABLE IF EXISTS provider_performance_metrics;
DROP TABLE IF EXISTS ai_routing_metrics_enhanced;
DROP TABLE IF EXISTS request_logs_gateway;