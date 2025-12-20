-- Rollback: expand_llm_providers
-- Created: 2025-12-18T21:36:44+05:30
-- WARNING: This rollback cannot remove enum values (PostgreSQL limitation)
-- Enum values 'azure', 'gemini', 'openrouter', 'custom' will remain

-- NOTE: The partial index is dropped in add_llm_custom_provider_index.down.sql

-- Remove the new columns
ALTER TABLE llm_provider_credentials
DROP COLUMN IF EXISTS config,
DROP COLUMN IF EXISTS headers,
DROP COLUMN IF EXISTS provider_name;

-- Note: ALTER TYPE ... REMOVE VALUE is not supported in PostgreSQL
-- The enum values will remain but won't cause issues as no data should reference them after rollback
