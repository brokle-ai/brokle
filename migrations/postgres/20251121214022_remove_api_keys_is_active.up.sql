-- Migration: remove_api_keys_is_active
-- Created: 2025-11-21T21:40:22+05:30
--
-- Remove is_active column from api_keys table
-- Status is now determined by:
--   - deleted_at IS NULL (active) vs deleted_at IS NOT NULL (deleted)
--   - expires_at < NOW() (expired)

ALTER TABLE api_keys DROP COLUMN IF EXISTS is_active;
