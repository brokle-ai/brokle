-- Migration: drop_telemetry_tables
-- Created: 2025-10-16T22:27:30+05:30
--
-- Drop PostgreSQL telemetry tables as part of Redis Streams migration
-- These tables are no longer needed as telemetry processing is now async via Redis Streams

-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_telemetry_batches_completion ON telemetry_batches;

-- Drop functions
DROP FUNCTION IF EXISTS update_telemetry_updated_at();

-- Drop indexes explicitly for documentation (they'll be dropped with tables automatically)
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

-- Drop tables in dependency order (children first, then parents)
DROP TABLE IF EXISTS telemetry_event_deduplication;
DROP TABLE IF EXISTS telemetry_events;
DROP TABLE IF EXISTS telemetry_batches;
