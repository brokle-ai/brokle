-- Migration: rename_data_type_to_type
-- Created: 2026-01-15T14:20:29+05:30
-- Purpose: Rename data_type column to type in score_configs table for cleaner API naming

-- Rename the column
ALTER TABLE score_configs RENAME COLUMN data_type TO type;

-- Update the CHECK constraint
ALTER TABLE score_configs DROP CONSTRAINT IF EXISTS score_configs_data_type_check;
ALTER TABLE score_configs ADD CONSTRAINT score_configs_type_check
    CHECK (type IN ('NUMERIC', 'CATEGORICAL', 'BOOLEAN'));

-- Update documentation
COMMENT ON COLUMN score_configs.type IS 'Score type: NUMERIC (float with optional min/max), CATEGORICAL (string from predefined list), BOOLEAN (0 or 1)';
