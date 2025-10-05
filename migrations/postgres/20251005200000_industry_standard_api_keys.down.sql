-- Rollback: Industry-Standard API Keys Migration
-- Reverts unique index on key_hash to non-unique

-- Drop unique index and recreate as non-unique
DROP INDEX IF EXISTS idx_api_keys_key_hash;
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);

-- Remove comments
COMMENT ON COLUMN api_keys.key_hash IS NULL;
COMMENT ON COLUMN api_keys.key_preview IS NULL;
COMMENT ON COLUMN api_keys.project_id IS NULL;
COMMENT ON TABLE api_keys IS NULL;
