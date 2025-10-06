-- ===================================
-- REMOVE ORGANIZATION_ID FROM API_KEYS
-- ===================================
-- API keys are project-scoped, organization can be derived via project
-- Removes redundant organization_id column to simplify schema

-- Drop foreign key constraint first
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS api_keys_organization_id_fkey;

-- Drop the organization_id column
ALTER TABLE api_keys DROP COLUMN IF EXISTS organization_id;

-- Add comment documenting the change
COMMENT ON TABLE api_keys IS 'Project-scoped API keys. Organization derived via projects.organization_id';
