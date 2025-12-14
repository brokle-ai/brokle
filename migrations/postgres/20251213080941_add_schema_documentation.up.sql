-- Migration: add_schema_documentation
-- Created: 2025-12-13T08:09:41+05:30

-- ===================================
-- ADD SCHEMA DOCUMENTATION COMMENTS
-- ===================================
-- Documents intentional deviations from design doc for clarity

-- Document dual FK structure in prompt_labels (design doc uses single FK)
-- Current implementation is BETTER: dual FK enables faster queries
COMMENT ON TABLE prompt_labels IS
  'Label-to-version mappings with denormalized foreign keys. Uses prompt_id + version_id (not prompt_version_id from design doc) for better query performance - enables fast "all labels for prompt" queries without JOIN.';

COMMENT ON COLUMN prompt_labels.prompt_id IS
  'Direct FK to prompt (enables O(1) lookup of all labels for a prompt without JOIN to versions table)';

COMMENT ON COLUMN prompt_labels.version_id IS
  'FK to specific version this label currently points to - can be changed atomically for instant rollback';

-- Document VARCHAR(50) vs VARCHAR(36) for label names
-- Allows for longer custom label names if needed in future
COMMENT ON COLUMN prompt_protected_labels.label_name IS
  'Protected label name requiring admin permissions to modify (VARCHAR(50) for future flexibility, design specifies VARCHAR(36))';
