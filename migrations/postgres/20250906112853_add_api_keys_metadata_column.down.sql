-- Rollback: add_api_keys_metadata_column
-- Created: 2025-09-06T11:28:53+05:30

-- Remove metadata column from api_keys table
ALTER TABLE api_keys DROP COLUMN IF EXISTS metadata;

