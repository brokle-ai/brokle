-- ===================================
-- ROLLBACK: SIMPLIFY API KEYS SCHEMA
-- ===================================
-- This rollback migration restores the removed columns

-- Restore description column
ALTER TABLE api_keys ADD COLUMN description TEXT;

-- Restore scopes column
ALTER TABLE api_keys ADD COLUMN scopes JSON;

-- Restore rate_limit_rpm column with default value
ALTER TABLE api_keys ADD COLUMN rate_limit_rpm INTEGER DEFAULT 1000;

-- Restore default_environment column with default value
ALTER TABLE api_keys ADD COLUMN default_environment VARCHAR(40) DEFAULT 'default';

-- Recreate check constraint for environment name
ALTER TABLE api_keys ADD CONSTRAINT chk_environment_name CHECK (
    default_environment ~ '^[a-z0-9_-]+$' AND
    default_environment NOT LIKE 'brokle%' AND
    LENGTH(default_environment) <= 40 AND
    LENGTH(default_environment) > 0
);

-- Recreate index for default_environment
CREATE INDEX idx_api_keys_default_environment ON api_keys(default_environment);
