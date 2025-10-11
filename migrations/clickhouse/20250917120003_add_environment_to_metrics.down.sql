-- ClickHouse Rollback: add_environment_to_metrics
-- Created: 2025-10-11T23:36:00+05:30

ALTER TABLE metrics
DROP COLUMN IF EXISTS environment;
