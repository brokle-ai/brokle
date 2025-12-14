-- Migration: add_label_version_ownership_constraint
-- Created: 2025-12-13T19:06:34+05:30

-- ===================================
-- ADD LABEL-VERSION OWNERSHIP CONSTRAINT (TRIGGER-BASED)
-- ===================================
-- PostgreSQL does not support subqueries in CHECK constraints.
-- Use a trigger to ensure labels only point to versions from the same prompt.

-- Function to validate version belongs to the same prompt
CREATE OR REPLACE FUNCTION validate_label_version_ownership()
RETURNS TRIGGER AS $$
BEGIN
    -- Check if the version_id belongs to the prompt_id
    IF NOT EXISTS (
        SELECT 1 FROM prompt_versions
        WHERE id = NEW.version_id AND prompt_id = NEW.prompt_id
    ) THEN
        RAISE EXCEPTION 'Label version_id (%) does not belong to prompt_id (%)',
            NEW.version_id, NEW.prompt_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to enforce constraint on INSERT and UPDATE
CREATE TRIGGER trg_validate_label_version_ownership
    BEFORE INSERT OR UPDATE ON prompt_labels
    FOR EACH ROW
    EXECUTE FUNCTION validate_label_version_ownership();

-- Comment explaining the constraint
COMMENT ON FUNCTION validate_label_version_ownership() IS
    'Validates that a label version_id belongs to the same prompt as prompt_id, preventing cross-prompt corruption';
