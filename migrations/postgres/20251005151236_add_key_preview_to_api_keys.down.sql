-- ===================================
-- ROLLBACK: ADD KEY_PREVIEW TO API_KEYS
-- ===================================

-- Drop key_preview column
ALTER TABLE api_keys DROP COLUMN IF EXISTS key_preview;
