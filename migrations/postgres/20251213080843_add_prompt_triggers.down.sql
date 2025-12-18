-- Rollback: add_prompt_triggers
-- Created: 2025-12-13T08:08:43+05:30

-- ===================================
-- ROLLBACK PROMPT MANAGEMENT TRIGGERS
-- ===================================
-- Removes triggers and functions added in the up migration

-- Drop triggers first
DROP TRIGGER IF EXISTS trg_enforce_label_uniqueness ON prompt_labels;
DROP TRIGGER IF EXISTS trg_prompts_updated_at ON prompts;

-- Then drop functions
DROP FUNCTION IF EXISTS enforce_label_uniqueness();
DROP FUNCTION IF EXISTS update_updated_at();
