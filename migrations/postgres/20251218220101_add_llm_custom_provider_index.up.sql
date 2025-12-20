-- Migration: add_llm_custom_provider_index
-- Created: 2025-12-18T22:01:01+05:30
-- Creates partial index for custom LLM providers
-- Separated from enum expansion due to PostgreSQL limitation (SQLSTATE 55P04):
-- Cannot use newly-added enum values in the same transaction.

CREATE UNIQUE INDEX IF NOT EXISTS idx_llm_credentials_custom_name
ON llm_provider_credentials(project_id, provider_name)
WHERE provider = 'custom' AND provider_name IS NOT NULL;

-- Documentation for columns added in expand_llm_providers migration
COMMENT ON COLUMN llm_provider_credentials.config IS 'Provider-specific configuration (JSONB). Azure: deployment_id, api_version. Gemini: location. Custom: models list.';
COMMENT ON COLUMN llm_provider_credentials.headers IS 'Encrypted custom HTTP headers (JSON string). Used for proxy authentication or custom endpoints.';
COMMENT ON COLUMN llm_provider_credentials.provider_name IS 'Custom provider name (only for custom provider type). Allows multiple custom providers per project.';
