-- Add optional model configuration to prompt versions
-- Config is optional JSONB: null means "no default config" (use playground/SDK overrides)

ALTER TABLE prompt_versions
  ADD COLUMN IF NOT EXISTS config JSONB DEFAULT NULL;

COMMENT ON COLUMN prompt_versions.config IS
  'Optional model config: {"model": "gpt-4", "temperature": 0.7, ...}. Null = no default.';
