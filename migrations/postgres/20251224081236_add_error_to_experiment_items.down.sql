-- Rollback: add_error_to_experiment_items
-- Created: 2025-12-24T08:12:36+05:30

-- Remove error column from experiment_items table
ALTER TABLE experiment_items DROP COLUMN IF EXISTS error;
