-- Rollback: add_prompt_indexes
-- Created: 2025-12-13T08:07:34+05:30

-- ===================================
-- ROLLBACK PROMPT MANAGEMENT PERFORMANCE INDEXES
-- ===================================
-- Removes indexes added in the up migration

-- Drop indexes in reverse order
DROP INDEX IF EXISTS idx_prompt_versions_variables;
DROP INDEX IF EXISTS idx_prompt_versions_config;
DROP INDEX IF EXISTS idx_prompt_versions_template;
DROP INDEX IF EXISTS idx_prompts_search;
