-- Migration: provider_infrastructure_refactor
-- Created: 2025-12-20T09:43:09+05:30
-- Consolidates: rename_llm_to_ai_providers + add_provider_to_provider_models
--               + add_custom_models + rename_ai_to_provider + refactor + bugfix
-- Renames directly from llm_* to provider_* (skipping intermediate ai_* naming)

-- ============================================
-- Phase 1: Rename directly (skip ai_* intermediate)
-- ============================================
ALTER TYPE llm_provider RENAME TO provider;
ALTER TABLE llm_provider_credentials RENAME TO provider_credentials;
ALTER INDEX idx_llm_provider_credentials_project_id RENAME TO idx_provider_credentials_project_id;
-- Note: idx_llm_credentials_custom_name is NOT renamed because it depends on provider_name column
-- which is dropped in Phase 3. The index is auto-dropped when the column is dropped.

-- ============================================
-- Phase 2: Add custom_models column
-- ============================================
ALTER TABLE provider_credentials ADD COLUMN custom_models TEXT[] DEFAULT '{}';
COMMENT ON COLUMN provider_credentials.custom_models IS
  'User-defined model IDs for fine-tuned models, private deployments, or custom provider models.';

-- ============================================
-- Phase 3: Refactor for multiple configs
-- ============================================
-- Drop both old constraints (unique_project_provider was original name)
ALTER TABLE provider_credentials DROP CONSTRAINT IF EXISTS llm_provider_credentials_project_id_provider_key;
ALTER TABLE provider_credentials DROP CONSTRAINT IF EXISTS unique_project_provider;

-- Rename provider → adapter (API protocol type)
ALTER TABLE provider_credentials RENAME COLUMN provider TO adapter;

-- Remove provider_name (replaced by name column)
ALTER TABLE provider_credentials DROP COLUMN IF EXISTS provider_name;

-- Add name as unique identifier per project
ALTER TABLE provider_credentials ADD COLUMN name VARCHAR(100) NOT NULL DEFAULT 'Default';
ALTER TABLE provider_credentials ALTER COLUMN name DROP DEFAULT;

-- New unique constraint
ALTER TABLE provider_credentials ADD CONSTRAINT provider_credentials_project_id_name_key UNIQUE(project_id, name);

-- Index for adapter filtering
CREATE INDEX idx_provider_credentials_adapter ON provider_credentials(adapter);

-- ============================================
-- Phase 4: Provider models (independent table)
-- ============================================
ALTER TABLE provider_models ADD COLUMN provider VARCHAR(50);
UPDATE provider_models SET provider =
  CASE
    WHEN tokenizer_id = 'openai' THEN 'openai'
    WHEN tokenizer_id = 'claude' THEN 'anthropic'
    WHEN tokenizer_id = 'gemini' THEN 'gemini'
    ELSE 'unknown'
  END;
ALTER TABLE provider_models ALTER COLUMN provider SET NOT NULL;
ALTER TABLE provider_models ADD COLUMN display_name VARCHAR(255);
CREATE INDEX idx_provider_models_provider ON provider_models(provider);
