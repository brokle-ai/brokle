-- Rollback: enhance_invitations_security
-- Created: 2026-01-10T08:03:51+05:30
-- WARNING: This will remove security enhancements

-- Make token_hash nullable again (before dropping)
ALTER TABLE user_invitations ALTER COLUMN token_hash DROP NOT NULL;

-- Restore the token column that was dropped
ALTER TABLE user_invitations ADD COLUMN IF NOT EXISTS token VARCHAR(255);

-- Drop the new indexes
DROP INDEX IF EXISTS idx_invitations_token_hash;
DROP INDEX IF EXISTS idx_invitations_email_org_pending;

-- Remove the new columns
ALTER TABLE user_invitations DROP COLUMN IF EXISTS token_hash;
ALTER TABLE user_invitations DROP COLUMN IF EXISTS token_preview;
ALTER TABLE user_invitations DROP COLUMN IF EXISTS message;
ALTER TABLE user_invitations DROP COLUMN IF EXISTS resent_count;
ALTER TABLE user_invitations DROP COLUMN IF EXISTS resent_at;
ALTER TABLE user_invitations DROP COLUMN IF EXISTS accepted_by_id;
ALTER TABLE user_invitations DROP COLUMN IF EXISTS revoked_at;
ALTER TABLE user_invitations DROP COLUMN IF EXISTS revoked_by_id;

-- Recreate the old plaintext token unique constraint (creates both constraint and index)
ALTER TABLE user_invitations ADD CONSTRAINT user_invitations_token_key UNIQUE (token);
