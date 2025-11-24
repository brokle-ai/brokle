-- Rollback: Remove service_name column and index from traces table

ALTER TABLE traces DROP INDEX IF EXISTS idx_service_name;
ALTER TABLE traces DROP COLUMN IF EXISTS service_name
