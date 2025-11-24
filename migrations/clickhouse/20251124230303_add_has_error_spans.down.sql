-- Rollback: Remove has_error column from spans
ALTER TABLE spans DROP COLUMN IF EXISTS has_error
