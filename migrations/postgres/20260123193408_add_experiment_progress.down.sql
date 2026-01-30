-- Rollback: add_experiment_progress
-- Created: 2026-01-23T19:34:08+05:30

-- Remove progress index
DROP INDEX IF EXISTS idx_experiments_status_progress;

-- Remove progress columns
ALTER TABLE experiments
DROP COLUMN IF EXISTS total_items,
DROP COLUMN IF EXISTS completed_items,
DROP COLUMN IF EXISTS failed_items;
