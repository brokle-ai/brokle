-- Migration: add_api_keys_metadata_column
-- Created: 2025-09-06T11:28:53+05:30

-- Add metadata JSONB column to api_keys table for flexible key metadata storage
ALTER TABLE api_keys ADD COLUMN metadata JSONB;

