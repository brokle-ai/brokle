-- Migration: expand_llm_providers
-- Created: 2025-12-18T21:36:44+05:30
-- Expand LLM Provider Credentials for additional providers
-- Adds Azure, Gemini, OpenRouter, and Custom provider support

-- Expand provider enum with new values
ALTER TYPE llm_provider ADD VALUE IF NOT EXISTS 'azure';
ALTER TYPE llm_provider ADD VALUE IF NOT EXISTS 'gemini';
ALTER TYPE llm_provider ADD VALUE IF NOT EXISTS 'openrouter';
ALTER TYPE llm_provider ADD VALUE IF NOT EXISTS 'custom';

-- Add new columns for provider-specific configuration
ALTER TABLE llm_provider_credentials
ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}',
ADD COLUMN IF NOT EXISTS headers TEXT,
ADD COLUMN IF NOT EXISTS provider_name VARCHAR(100);

-- NOTE: The partial index for custom providers is created in a separate migration
-- (add_llm_custom_provider_index) due to PostgreSQL limitation SQLSTATE 55P04:
-- Cannot use newly-added enum values in the same transaction.
