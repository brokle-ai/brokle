-- Migration: add_experiment_configs (rollback)
-- Created: 2026-01-23

-- Remove source column and config_id from experiments
DROP INDEX IF EXISTS idx_experiments_source;
ALTER TABLE experiments DROP COLUMN IF EXISTS source;
ALTER TABLE experiments DROP COLUMN IF EXISTS config_id;

-- Remove experiment_configs table
DROP INDEX IF EXISTS idx_experiment_configs_dataset;
DROP INDEX IF EXISTS idx_experiment_configs_prompt;
DROP INDEX IF EXISTS idx_experiment_configs_experiment;
DROP TABLE IF EXISTS experiment_configs;
