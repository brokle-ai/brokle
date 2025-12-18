-- Rollback: add_prompt_constraints
-- Created: 2025-12-13T08:06:25+05:30

-- ===================================
-- ROLLBACK PROMPT MANAGEMENT CHECK CONSTRAINTS
-- ===================================
-- Removes validation constraints added in the up migration

-- Drop constraints in reverse order
ALTER TABLE prompt_protected_labels DROP CONSTRAINT IF EXISTS prompt_protected_labels_format;
ALTER TABLE prompt_labels DROP CONSTRAINT IF EXISTS prompt_labels_name_length;
ALTER TABLE prompt_labels DROP CONSTRAINT IF EXISTS prompt_labels_name_format;
ALTER TABLE prompt_versions DROP CONSTRAINT IF EXISTS prompt_versions_commit_message_length;
ALTER TABLE prompt_versions DROP CONSTRAINT IF EXISTS prompt_versions_version_positive;
ALTER TABLE prompts DROP CONSTRAINT IF EXISTS prompts_name_format;
