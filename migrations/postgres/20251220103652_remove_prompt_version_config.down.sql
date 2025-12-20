-- Restore config column (will be empty for existing rows)
ALTER TABLE prompt_versions ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}';
