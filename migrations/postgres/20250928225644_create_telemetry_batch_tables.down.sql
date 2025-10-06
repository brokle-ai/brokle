-- Rollback: create_telemetry_batch_tables
-- Created: 2025-09-28T22:56:44+05:30
--
-- Drop telemetry batch processing tables and related objects

-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_telemetry_batches_completion ON telemetry_batches;

-- Drop functions
DROP FUNCTION IF EXISTS update_telemetry_updated_at();

-- Drop indexes (they'll be dropped with tables, but explicit for clarity)
DROP INDEX IF EXISTS idx_telemetry_batches_project_id;
DROP INDEX IF EXISTS idx_telemetry_batches_status;
DROP INDEX IF EXISTS idx_telemetry_batches_created_at;
DROP INDEX IF EXISTS idx_telemetry_events_batch_id;
DROP INDEX IF EXISTS idx_telemetry_events_event_type;
DROP INDEX IF EXISTS idx_telemetry_events_created_at;
DROP INDEX IF EXISTS idx_telemetry_dedup_expires_at;
DROP INDEX IF EXISTS idx_telemetry_dedup_project_id;
DROP INDEX IF EXISTS idx_telemetry_batches_metadata;
DROP INDEX IF EXISTS idx_telemetry_events_payload;
DROP INDEX IF EXISTS idx_telemetry_batches_processing;
DROP INDEX IF EXISTS idx_telemetry_events_unprocessed;
DROP INDEX IF EXISTS idx_telemetry_events_failed;

-- Drop tables in dependency order
DROP TABLE IF EXISTS telemetry_event_deduplication;
DROP TABLE IF EXISTS telemetry_events;
DROP TABLE IF EXISTS telemetry_batches;

