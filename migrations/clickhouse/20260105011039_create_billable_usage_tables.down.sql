-- ClickHouse Rollback: create_billable_usage_tables
-- Created: 2026-01-05T01:10:39+05:30

-- Drop materialized views first (they depend on tables)
DROP VIEW IF EXISTS billable_usage_daily_mv;
DROP VIEW IF EXISTS billable_scores_hourly_mv;
DROP VIEW IF EXISTS billable_usage_hourly_mv;

-- Drop tables
DROP TABLE IF EXISTS billable_usage_daily;
DROP TABLE IF EXISTS billable_usage_hourly;
