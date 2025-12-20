-- Rollback: provider_infrastructure_refactor
-- Created: 2025-12-20T09:43:09+05:30

-- ============================================
-- Phase 4: Rollback provider_models
-- ============================================
DROP INDEX IF EXISTS idx_provider_models_provider;
ALTER TABLE provider_models DROP COLUMN IF EXISTS display_name;
ALTER TABLE provider_models DROP COLUMN IF EXISTS provider;

-- ============================================
-- Phase 3: Rollback refactor
-- ============================================
DROP INDEX IF EXISTS idx_provider_credentials_adapter;
ALTER TABLE provider_credentials DROP CONSTRAINT IF EXISTS provider_credentials_project_id_name_key;
ALTER TABLE provider_credentials DROP COLUMN IF EXISTS name;
ALTER TABLE provider_credentials ADD COLUMN provider_name VARCHAR(100);
ALTER TABLE provider_credentials RENAME COLUMN adapter TO provider;
ALTER TABLE provider_credentials ADD CONSTRAINT unique_project_provider UNIQUE(project_id, provider);
-- Recreate the partial index for custom providers (depends on provider_name column)
CREATE UNIQUE INDEX IF NOT EXISTS idx_llm_credentials_custom_name
ON provider_credentials(project_id, provider_name)
WHERE provider = 'custom' AND provider_name IS NOT NULL;

-- ============================================
-- Phase 2: Rollback custom_models
-- ============================================
ALTER TABLE provider_credentials DROP COLUMN IF EXISTS custom_models;

-- ============================================
-- Phase 1: Rollback rename (back to llm_*)
-- ============================================
ALTER INDEX idx_provider_credentials_project_id RENAME TO idx_llm_provider_credentials_project_id;
-- Note: idx_llm_credentials_custom_name was recreated in Phase 3 above, no rename needed
ALTER TABLE provider_credentials RENAME TO llm_provider_credentials;
ALTER TYPE provider RENAME TO llm_provider;
