-- ===================================
-- ROLLBACK: REPLACE KEY_ID WITH KEY_HASH
-- ===================================
-- Restore key_id and secret_hash columns

-- Step 1: Drop key_hash index
DROP INDEX IF EXISTS idx_api_keys_key_hash;

-- Step 2: Add back key_id column
ALTER TABLE api_keys ADD COLUMN key_id VARCHAR(100);

-- Step 3: Rename key_hash back to secret_hash
ALTER TABLE api_keys RENAME COLUMN key_hash TO secret_hash;

-- Step 4: Restore indexes
CREATE INDEX idx_api_keys_secret_hash ON api_keys(secret_hash);
CREATE INDEX idx_api_keys_key_id ON api_keys(key_id);
