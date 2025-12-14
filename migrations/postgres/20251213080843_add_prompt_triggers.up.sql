-- Migration: add_prompt_triggers
-- Created: 2025-12-13T08:08:43+05:30

-- ===================================
-- ADD PROMPT MANAGEMENT TRIGGERS
-- ===================================
-- Adds database triggers for automatic data management per design doc (db-schema.md)

-- Function: Auto-update updated_at on prompts table
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger: Apply to prompts table
CREATE TRIGGER trg_prompts_updated_at
    BEFORE UPDATE ON prompts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Function: Enforce label uniqueness per prompt
-- Prevents same label name from existing on multiple versions of the same prompt
CREATE OR REPLACE FUNCTION enforce_label_uniqueness()
RETURNS TRIGGER AS $$
DECLARE
    existing_count INTEGER;
BEGIN
    -- Check if label already exists on another version of this prompt
    SELECT COUNT(*) INTO existing_count
    FROM prompt_labels
    WHERE prompt_id = NEW.prompt_id
      AND name = NEW.name
      AND id != COALESCE(NEW.id, '00000000000000000000000000');

    IF existing_count > 0 THEN
        RAISE EXCEPTION 'Label "%" already exists on another version of this prompt', NEW.name;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger: Apply to prompt_labels table
CREATE TRIGGER trg_enforce_label_uniqueness
    BEFORE INSERT OR UPDATE ON prompt_labels
    FOR EACH ROW
    EXECUTE FUNCTION enforce_label_uniqueness();
