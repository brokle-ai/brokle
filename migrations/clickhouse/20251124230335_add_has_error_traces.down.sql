-- Rollback: Remove has_error column from traces
ALTER TABLE traces DROP COLUMN IF EXISTS has_error
