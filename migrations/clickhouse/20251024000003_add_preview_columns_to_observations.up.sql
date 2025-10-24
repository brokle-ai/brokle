-- Add preview columns to observations table
-- Preview columns are ALWAYS populated (even when content is offloaded to S3)
-- Adaptive sizing: 300-800 chars based on content type (JSON, text, markdown, errors)
-- Note: Adding columns AFTER their corresponding input/output columns for better organization

ALTER TABLE observations
    ADD COLUMN IF NOT EXISTS input_preview Nullable(String) AFTER input,
    ADD COLUMN IF NOT EXISTS output_preview Nullable(String) AFTER output;
