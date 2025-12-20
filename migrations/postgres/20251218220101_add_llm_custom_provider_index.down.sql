-- Rollback: add_llm_custom_provider_index
-- Created: 2025-12-18T22:01:01+05:30

DROP INDEX IF EXISTS idx_llm_credentials_custom_name;
