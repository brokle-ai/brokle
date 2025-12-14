-- Drop prompt_versions table and all indexes
DROP INDEX IF EXISTS idx_prompt_versions_created_by;
DROP INDEX IF EXISTS idx_prompt_versions_prompt_created;
DROP INDEX IF EXISTS idx_prompt_versions_prompt_version;
DROP TABLE IF EXISTS prompt_versions;
