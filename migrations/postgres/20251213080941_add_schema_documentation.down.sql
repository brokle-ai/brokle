-- Rollback: add_schema_documentation
-- Created: 2025-12-13T08:09:41+05:30

-- ===================================
-- ROLLBACK SCHEMA DOCUMENTATION COMMENTS
-- ===================================
-- Removes documentation comments added in the up migration

-- Remove table and column comments
COMMENT ON TABLE prompt_labels IS NULL;
COMMENT ON COLUMN prompt_labels.prompt_id IS NULL;
COMMENT ON COLUMN prompt_labels.version_id IS NULL;
COMMENT ON COLUMN prompt_protected_labels.label_name IS NULL;
