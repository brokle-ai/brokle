-- Rollback: update_password_reset_tokens_schema
-- Created: 2025-09-06T11:30:26+05:30

-- Revert back to used BOOLEAN from used_at TIMESTAMP
-- Convert existing data: if used_at is not null, set used = true
ALTER TABLE password_reset_tokens ADD COLUMN used BOOLEAN DEFAULT FALSE;
UPDATE password_reset_tokens SET used = TRUE WHERE used_at IS NOT NULL;
ALTER TABLE password_reset_tokens DROP COLUMN used_at;

