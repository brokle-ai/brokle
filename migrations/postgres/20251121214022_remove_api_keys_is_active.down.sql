-- Rollback: remove_api_keys_is_active
-- Created: 2025-11-21T21:40:22+05:30
--
-- Re-add is_active column to api_keys table
-- Default to true for all existing keys

ALTER TABLE api_keys ADD COLUMN is_active BOOLEAN DEFAULT true NOT NULL;
