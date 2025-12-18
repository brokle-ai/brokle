-- Drop prompt_protected_labels table and all indexes
DROP INDEX IF EXISTS idx_prompt_protected_labels_project_id;
DROP INDEX IF EXISTS idx_prompt_protected_labels_project_label;
DROP TABLE IF EXISTS prompt_protected_labels;
