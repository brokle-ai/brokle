-- Migration: rename_data_type_to_type (rollback)
-- Created: 2026-01-15T14:20:29+05:30

-- Revert column rename
ALTER TABLE score_configs RENAME COLUMN type TO data_type;

-- Revert CHECK constraint
ALTER TABLE score_configs DROP CONSTRAINT IF EXISTS score_configs_type_check;
ALTER TABLE score_configs ADD CONSTRAINT score_configs_data_type_check
    CHECK (data_type IN ('NUMERIC', 'CATEGORICAL', 'BOOLEAN'));

-- Revert documentation
COMMENT ON COLUMN score_configs.data_type IS 'Score data type: NUMERIC (float with optional min/max), CATEGORICAL (string from predefined list), BOOLEAN (0 or 1)';
