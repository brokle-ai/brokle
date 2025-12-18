-- Drop prompt_labels table and all indexes
DROP INDEX IF EXISTS idx_prompt_labels_name;
DROP INDEX IF EXISTS idx_prompt_labels_version_id;
DROP INDEX IF EXISTS idx_prompt_labels_prompt_name;
DROP TABLE IF EXISTS prompt_labels;
