-- Remove config column from prompt_versions

ALTER TABLE prompt_versions DROP COLUMN IF EXISTS config;
