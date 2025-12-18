-- Rollback: Drop LLM provider credentials table and enum

DROP TABLE IF EXISTS llm_provider_credentials;
DROP TYPE IF EXISTS llm_provider;
