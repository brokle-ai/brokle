-- Migration: add_experiment_progress
-- Created: 2026-01-23T19:34:08+05:30

-- Add progress tracking columns to experiments table
ALTER TABLE experiments
ADD COLUMN total_items INTEGER NOT NULL DEFAULT 0,
ADD COLUMN completed_items INTEGER NOT NULL DEFAULT 0,
ADD COLUMN failed_items INTEGER NOT NULL DEFAULT 0;

-- Add index for efficient progress queries on running experiments
CREATE INDEX idx_experiments_status_progress ON experiments(status) WHERE status = 'running';
