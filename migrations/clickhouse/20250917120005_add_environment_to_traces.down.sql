-- ClickHouse Rollback: add_environment_to_traces
-- Created: 2025-10-11T23:36:00+05:30

ALTER TABLE traces
DROP COLUMN IF EXISTS environment;
