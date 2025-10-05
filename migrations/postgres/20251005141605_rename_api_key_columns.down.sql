-- ===================================
-- ROLLBACK: RENAME API KEY COLUMNS
-- ===================================
-- Restore original column names

-- Step 1: Drop new indexes
DROP INDEX IF EXISTS idx_api_keys_secret_hash;
DROP INDEX IF EXISTS idx_api_keys_key_id;

-- Step 2: Recreate old indexes
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_key_prefix ON api_keys(key_prefix);

-- Step 3: Revert column size
ALTER TABLE api_keys ALTER COLUMN key_id TYPE VARCHAR(8);

-- Step 4: Rename columns back
ALTER TABLE api_keys RENAME COLUMN secret_hash TO key_hash;
ALTER TABLE api_keys RENAME COLUMN key_id TO key_prefix;
