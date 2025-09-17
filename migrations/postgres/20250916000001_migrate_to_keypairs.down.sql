-- ===================================
-- ROLLBACK: MIGRATE FROM KEY PAIRS BACK TO API KEYS
-- Migration Rollback: Restore api_keys table and remove key_pairs table
-- WARNING: This will lose all key_pairs data
-- ===================================

-- ===================================
-- RECREATE API_KEYS TABLE
-- ===================================

-- Recreate the original api_keys table structure
CREATE TABLE api_keys (
    id CHAR(26) PRIMARY KEY,
    user_id CHAR(26) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id CHAR(26) REFERENCES organizations(id) ON DELETE CASCADE,
    project_id CHAR(26) REFERENCES projects(id) ON DELETE CASCADE,
    environment_id CHAR(26) REFERENCES environments(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    key_prefix VARCHAR(8) NOT NULL,
    scopes JSON,
    is_active BOOLEAN DEFAULT TRUE,
    rate_limit_rpm INTEGER DEFAULT 1000,
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- ===================================
-- RECREATE API_KEYS INDEXES
-- ===================================

-- Recreate all original indexes for api_keys table
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_organization_id ON api_keys(organization_id);
CREATE INDEX idx_api_keys_project_id ON api_keys(project_id);
CREATE INDEX idx_api_keys_environment_id ON api_keys(environment_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_key_prefix ON api_keys(key_prefix);
CREATE INDEX idx_api_keys_is_active ON api_keys(is_active);
CREATE INDEX idx_api_keys_deleted_at ON api_keys(deleted_at);

-- ===================================
-- RECREATE UPDATED_AT TRIGGER FOR API_KEYS
-- ===================================

-- Recreate updated_at trigger for api_keys table
CREATE TRIGGER update_api_keys_updated_at
    BEFORE UPDATE ON api_keys
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ===================================
-- DROP KEY_PAIRS TABLE AND RELATED OBJECTS
-- ===================================

-- Drop check constraints first
ALTER TABLE key_pairs DROP CONSTRAINT IF EXISTS chk_public_key_format;
ALTER TABLE key_pairs DROP CONSTRAINT IF EXISTS chk_secret_key_prefix;
ALTER TABLE key_pairs DROP CONSTRAINT IF EXISTS chk_name_not_empty;
ALTER TABLE key_pairs DROP CONSTRAINT IF EXISTS chk_rate_limit_positive;
ALTER TABLE key_pairs DROP CONSTRAINT IF EXISTS chk_expires_at_future;

-- Drop trigger
DROP TRIGGER IF EXISTS update_key_pairs_updated_at ON key_pairs;

-- Drop indexes
DROP INDEX IF EXISTS idx_key_pairs_public_key;
DROP INDEX IF EXISTS idx_key_pairs_secret_key_hash;
DROP INDEX IF EXISTS idx_key_pairs_secret_key_prefix;
DROP INDEX IF EXISTS idx_key_pairs_user_id;
DROP INDEX IF EXISTS idx_key_pairs_organization_id;
DROP INDEX IF EXISTS idx_key_pairs_project_id;
DROP INDEX IF EXISTS idx_key_pairs_environment_id;
DROP INDEX IF EXISTS idx_key_pairs_is_active;
DROP INDEX IF EXISTS idx_key_pairs_deleted_at;
DROP INDEX IF EXISTS idx_key_pairs_expires_at;
DROP INDEX IF EXISTS idx_key_pairs_last_used_at;
DROP INDEX IF EXISTS idx_key_pairs_name;
DROP INDEX IF EXISTS idx_key_pairs_rate_limit_rpm;
DROP INDEX IF EXISTS idx_key_pairs_active_lookup;
DROP INDEX IF EXISTS idx_key_pairs_project_active;
DROP INDEX IF EXISTS idx_key_pairs_org_active;

-- Drop the key_pairs table
DROP TABLE IF EXISTS key_pairs;