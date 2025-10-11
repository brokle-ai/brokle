-- Drop ClickHouse gateway analytics tables
-- Migration: 20250906144200_create_gateway_analytics_tables.down.sql

-- Drop materialized views first (they depend on tables)
DROP VIEW IF EXISTS gateway_usage_daily_mv;
DROP VIEW IF EXISTS gateway_usage_hourly_mv;

-- Drop main analytics tables
DROP TABLE IF EXISTS gateway_cost_metrics;
DROP TABLE IF EXISTS gateway_usage_metrics;
DROP TABLE IF EXISTS gateway_request_metrics;