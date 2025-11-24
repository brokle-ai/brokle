-- Rollback: Remove service_name column and index from spans table

ALTER TABLE spans DROP INDEX IF EXISTS idx_service_name;
ALTER TABLE spans DROP COLUMN IF EXISTS service_name
