-- Add blob storage support (for S3 offloading) and preview columns to traces table
-- Blob storage columns: Store S3 file IDs when content is offloaded (>10KB threshold)
-- Preview columns: ALWAYS populated (even when content is offloaded to S3)
-- Adaptive sizing: 300-800 chars based on content type (JSON, text, markdown, errors)
-- Column ordering: input → input_blob_storage_id → input_preview (grouping related fields)

ALTER TABLE traces
    ADD COLUMN IF NOT EXISTS input_blob_storage_id Nullable(String) AFTER input,
    ADD COLUMN IF NOT EXISTS output_blob_storage_id Nullable(String) AFTER output,
    ADD COLUMN IF NOT EXISTS input_preview Nullable(String) AFTER input_blob_storage_id,
    ADD COLUMN IF NOT EXISTS output_preview Nullable(String) AFTER output_blob_storage_id;
