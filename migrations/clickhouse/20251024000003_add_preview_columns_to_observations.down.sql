-- Rollback: Drop preview columns from observations table

ALTER TABLE observations
    DROP COLUMN IF EXISTS input_preview,
    DROP COLUMN IF EXISTS output_preview;
