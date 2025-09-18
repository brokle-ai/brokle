-- ===================================
-- REMOVE ENVIRONMENTS & UPDATE API KEYS
-- ===================================
-- This migration removes the Environment entity entirely and implements
-- a Langfuse-style approach where environments are just metadata tags.

-- Step 1: Remove environment_id from api_keys and add default_environment
-- First, let's backup any existing environment associations
CREATE TEMP TABLE temp_api_key_env AS
SELECT ak.id, ak.environment_id, e.name as env_name, e.slug as env_slug
FROM api_keys ak
LEFT JOIN environments e ON ak.environment_id = e.id
WHERE ak.environment_id IS NOT NULL;

-- Remove the foreign key constraint and environment_id column
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS api_keys_environment_id_fkey;
ALTER TABLE api_keys DROP COLUMN IF EXISTS environment_id;

-- Make project_id required (it was optional before)
ALTER TABLE api_keys ALTER COLUMN project_id SET NOT NULL;

-- Add default_environment column
ALTER TABLE api_keys ADD COLUMN default_environment VARCHAR(40) DEFAULT 'default';

-- Update existing API keys to use environment names from the temp table
UPDATE api_keys
SET default_environment = COALESCE(temp.env_name, 'default')
FROM temp_api_key_env temp
WHERE api_keys.id = temp.id;

-- Add constraint for environment name validation (Langfuse rules)
ALTER TABLE api_keys ADD CONSTRAINT chk_environment_name
CHECK (
    default_environment ~ '^[a-z0-9_-]+$'
    AND default_environment NOT LIKE 'langfuse%'
    AND length(default_environment) <= 40
    AND length(default_environment) > 0
);

-- Step 2: Drop the environments table
DROP TABLE IF EXISTS environments CASCADE;

-- Step 3: Clean up any orphaned references in other tables (if any exist)
-- This is a safety measure in case there are other tables referencing environments

-- Add index on default_environment for better query performance
CREATE INDEX idx_api_keys_default_environment ON api_keys(default_environment);

-- Add comment to document the change
COMMENT ON COLUMN api_keys.default_environment IS 'Default environment tag for API requests (Langfuse-style), can be overridden per request';