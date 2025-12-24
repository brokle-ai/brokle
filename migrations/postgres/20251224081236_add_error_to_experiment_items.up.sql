-- Migration: add_error_to_experiment_items
-- Created: 2025-12-24T08:12:36+05:30

-- Add error column to experiment_items table to store task execution errors
ALTER TABLE experiment_items ADD COLUMN error TEXT;
