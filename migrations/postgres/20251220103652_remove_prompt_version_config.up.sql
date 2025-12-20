-- Remove model config from prompt versions
-- Config is no longer needed as prompts execute via playground only

ALTER TABLE prompt_versions DROP COLUMN IF EXISTS config;
