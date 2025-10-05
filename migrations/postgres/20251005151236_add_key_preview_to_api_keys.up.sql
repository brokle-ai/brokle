-- ===================================
-- ADD KEY_PREVIEW TO API_KEYS
-- ===================================
-- Add key_preview column to store display version of API key

-- Add key_preview column
ALTER TABLE api_keys ADD COLUMN key_preview VARCHAR(50);

-- For existing keys, create preview from key_id (temporary - will be updated on next use)
UPDATE api_keys SET key_preview = CONCAT(SUBSTRING(key_id FROM 1 FOR 8), '...', SUBSTRING(key_id FROM LENGTH(key_id)-3));

-- Make it NOT NULL after populating
ALTER TABLE api_keys ALTER COLUMN key_preview SET NOT NULL;
