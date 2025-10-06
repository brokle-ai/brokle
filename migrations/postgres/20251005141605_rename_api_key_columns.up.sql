-- ===================================
-- RENAME API KEY COLUMNS
-- ===================================
-- Rename columns to match project-scoped key architecture:
-- - key_hash → secret_hash (more accurate naming)
-- - key_prefix → key_id (stores full identifier, not just prefix)

-- Step 1: Rename columns
ALTER TABLE api_keys RENAME COLUMN key_hash TO secret_hash;
ALTER TABLE api_keys RENAME COLUMN key_prefix TO key_id;

-- Step 2: Update key_id size to support full identifier (bk_proj_{project_id})
ALTER TABLE api_keys ALTER COLUMN key_id TYPE VARCHAR(100);

-- Step 3: Drop old indexes
DROP INDEX IF EXISTS idx_api_keys_key_hash;
DROP INDEX IF EXISTS idx_api_keys_key_prefix;

-- Step 4: Create new indexes with updated column names
CREATE INDEX idx_api_keys_secret_hash ON api_keys(secret_hash);
CREATE INDEX idx_api_keys_key_id ON api_keys(key_id);
