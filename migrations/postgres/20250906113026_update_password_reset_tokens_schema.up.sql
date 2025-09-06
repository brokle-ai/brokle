-- Migration: update_password_reset_tokens_schema
-- Created: 2025-09-06T11:30:26+05:30

-- Replace used BOOLEAN with used_at TIMESTAMP to track when token was used
ALTER TABLE password_reset_tokens DROP COLUMN used;
ALTER TABLE password_reset_tokens ADD COLUMN used_at TIMESTAMP WITH TIME ZONE;

