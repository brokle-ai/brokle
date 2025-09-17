-- ===================================
-- MIGRATE FROM API KEYS TO KEY PAIRS
-- Migration: Replace api_keys table with key_pairs table for public+secret key authentication
-- ===================================

-- Create key_pairs table with public+secret key authentication model
CREATE TABLE key_pairs (
    id CHAR(26) PRIMARY KEY,
    user_id CHAR(26) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id CHAR(26) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    project_id CHAR(26) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    environment_id CHAR(26) REFERENCES environments(id) ON DELETE CASCADE,

    -- Key pair identification
    name VARCHAR(255) NOT NULL,

    -- Public key (pk_projectId_random) - stored in plain text for lookup
    public_key VARCHAR(255) NOT NULL UNIQUE,

    -- Secret key hash (sk_random hashed) - never store plain text
    secret_key_hash VARCHAR(255) NOT NULL UNIQUE,
    secret_key_prefix VARCHAR(8) NOT NULL DEFAULT 'sk_', -- Always 'sk_' for validation

    -- Scoping and permissions
    scopes JSON, -- ['gateway:read', 'analytics:read', etc.]

    -- Rate limiting and usage controls
    rate_limit_rpm INTEGER DEFAULT 1000,

    -- Status and lifecycle
    is_active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,

    -- Metadata for enterprise features
    metadata JSONB,

    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- ===================================
-- INDEXES FOR KEY_PAIRS TABLE
-- ===================================

-- Primary authentication lookup indexes (most critical for performance)
CREATE INDEX idx_key_pairs_public_key ON key_pairs(public_key);
CREATE INDEX idx_key_pairs_secret_key_hash ON key_pairs(secret_key_hash);
CREATE INDEX idx_key_pairs_secret_key_prefix ON key_pairs(secret_key_prefix);

-- Relationship indexes for foreign key lookups
CREATE INDEX idx_key_pairs_user_id ON key_pairs(user_id);
CREATE INDEX idx_key_pairs_organization_id ON key_pairs(organization_id);
CREATE INDEX idx_key_pairs_project_id ON key_pairs(project_id);
CREATE INDEX idx_key_pairs_environment_id ON key_pairs(environment_id);

-- Status and lifecycle indexes for filtering and cleanup
CREATE INDEX idx_key_pairs_is_active ON key_pairs(is_active);
CREATE INDEX idx_key_pairs_deleted_at ON key_pairs(deleted_at);
CREATE INDEX idx_key_pairs_expires_at ON key_pairs(expires_at);
CREATE INDEX idx_key_pairs_last_used_at ON key_pairs(last_used_at);

-- Performance optimization indexes
CREATE INDEX idx_key_pairs_name ON key_pairs(name);
CREATE INDEX idx_key_pairs_rate_limit_rpm ON key_pairs(rate_limit_rpm);

-- Composite indexes for common query patterns
CREATE INDEX idx_key_pairs_active_lookup ON key_pairs(is_active, deleted_at, expires_at) WHERE is_active = true AND deleted_at IS NULL;
CREATE INDEX idx_key_pairs_project_active ON key_pairs(project_id, is_active) WHERE is_active = true AND deleted_at IS NULL;
CREATE INDEX idx_key_pairs_org_active ON key_pairs(organization_id, is_active) WHERE is_active = true AND deleted_at IS NULL;

-- ===================================
-- UPDATED_AT TRIGGER FOR KEY_PAIRS
-- ===================================

-- Add updated_at trigger for key_pairs table
CREATE TRIGGER update_key_pairs_updated_at
    BEFORE UPDATE ON key_pairs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ===================================
-- DATA MIGRATION (IF NEEDED)
-- ===================================

-- Note: Since this is early development phase and user explicitly stated no backward compatibility needed,
-- we're not migrating data from api_keys to key_pairs. The old api_keys table will be dropped.
-- If data migration is needed later, it would go here.

-- ===================================
-- DROP OLD API_KEYS TABLE AND RELATED OBJECTS
-- ===================================

-- Drop triggers first
DROP TRIGGER IF EXISTS update_api_keys_updated_at ON api_keys;

-- Drop indexes
DROP INDEX IF EXISTS idx_api_keys_user_id;
DROP INDEX IF EXISTS idx_api_keys_organization_id;
DROP INDEX IF EXISTS idx_api_keys_project_id;
DROP INDEX IF EXISTS idx_api_keys_environment_id;
DROP INDEX IF EXISTS idx_api_keys_key_hash;
DROP INDEX IF EXISTS idx_api_keys_key_prefix;
DROP INDEX IF EXISTS idx_api_keys_is_active;
DROP INDEX IF EXISTS idx_api_keys_deleted_at;

-- Drop the api_keys table
DROP TABLE IF EXISTS api_keys;

-- ===================================
-- CONSTRAINTS VALIDATION
-- ===================================

-- Add check constraints for key format validation
ALTER TABLE key_pairs ADD CONSTRAINT chk_public_key_format
    CHECK (public_key ~ '^pk_[0-9A-Z]{26}_[a-zA-Z0-9]+$');

ALTER TABLE key_pairs ADD CONSTRAINT chk_secret_key_prefix
    CHECK (secret_key_prefix = 'sk_');

ALTER TABLE key_pairs ADD CONSTRAINT chk_name_not_empty
    CHECK (LENGTH(TRIM(name)) > 0);

ALTER TABLE key_pairs ADD CONSTRAINT chk_rate_limit_positive
    CHECK (rate_limit_rpm > 0);

-- Ensure expires_at is in the future when set
ALTER TABLE key_pairs ADD CONSTRAINT chk_expires_at_future
    CHECK (expires_at IS NULL OR expires_at > created_at);

-- ===================================
-- COMMENTS FOR DOCUMENTATION
-- ===================================

COMMENT ON TABLE key_pairs IS 'Public+Secret key pairs for API authentication replacing the old api_keys system';
COMMENT ON COLUMN key_pairs.public_key IS 'Public key in format pk_projectId_random - safe to expose';
COMMENT ON COLUMN key_pairs.secret_key_hash IS 'Hashed secret key in format sk_random - never store plain text';
COMMENT ON COLUMN key_pairs.secret_key_prefix IS 'Always sk_ - used for validation';
COMMENT ON COLUMN key_pairs.scopes IS 'JSON array of permissions like ["gateway:read", "analytics:read"]';
COMMENT ON COLUMN key_pairs.metadata IS 'JSONB metadata for enterprise features and custom attributes';