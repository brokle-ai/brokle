-- ===================================
-- SIMPLIFY API KEYS SCHEMA
-- ===================================
-- This migration removes unnecessary fields from api_keys table:
-- - description: Not needed for most use cases
-- - scopes: Permissions should be controlled at org RBAC level
-- - rate_limit_rpm: Rate limiting handled globally at project/org tier level
-- - default_environment: Environments now handled as tags via X-Environment header

-- Remove description column
ALTER TABLE api_keys DROP COLUMN IF EXISTS description;

-- Remove scopes column
ALTER TABLE api_keys DROP COLUMN IF EXISTS scopes;

-- Remove rate_limit_rpm column
ALTER TABLE api_keys DROP COLUMN IF EXISTS rate_limit_rpm;

-- Drop index for default_environment (must be dropped before column)
DROP INDEX IF EXISTS idx_api_keys_default_environment;

-- Drop check constraint for environment name
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS chk_environment_name;

-- Remove default_environment column
ALTER TABLE api_keys DROP COLUMN IF EXISTS default_environment;