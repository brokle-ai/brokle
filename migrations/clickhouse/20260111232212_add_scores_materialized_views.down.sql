-- ClickHouse Rollback: add_scores_materialized_views
-- Created: 2026-01-11T23:22:12+05:30
--
-- Drop materialized views for scores analytics

DROP VIEW IF EXISTS scores_source_distribution;
DROP VIEW IF EXISTS scores_by_trace;
DROP VIEW IF EXISTS scores_by_experiment;
DROP VIEW IF EXISTS scores_daily_summary;
