-- Rollback: add_label_version_ownership_constraint
-- Created: 2025-12-13T19:06:34+05:30

-- ===================================
-- REMOVE LABEL-VERSION OWNERSHIP CONSTRAINT
-- ===================================
-- Removes the trigger-based constraint that ensures labels only point to versions from the same prompt

-- Drop the trigger first (depends on function)
DROP TRIGGER IF EXISTS trg_validate_label_version_ownership ON prompt_labels;

-- Drop the validation function
DROP FUNCTION IF EXISTS validate_label_version_ownership();
