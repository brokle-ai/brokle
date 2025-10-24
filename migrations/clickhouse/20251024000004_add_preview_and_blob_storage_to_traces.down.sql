-- Rollback: Remove preview and blob storage columns from traces table

ALTER TABLE traces
    DROP COLUMN IF EXISTS input_preview,
    DROP COLUMN IF EXISTS output_preview,
    DROP COLUMN IF EXISTS input_blob_storage_id,
    DROP COLUMN IF EXISTS output_blob_storage_id;
