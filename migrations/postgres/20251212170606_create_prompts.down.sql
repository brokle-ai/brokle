-- Drop prompts table and all indexes
DROP INDEX IF EXISTS idx_prompts_created_at;
DROP INDEX IF EXISTS idx_prompts_tags;
DROP INDEX IF EXISTS idx_prompts_project_id;
DROP INDEX IF EXISTS idx_prompts_project_name;
DROP TABLE IF EXISTS prompts;
