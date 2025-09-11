-- Rollback: Remove added fields from user_invitations table

-- Drop indexes
DROP INDEX IF EXISTS idx_user_invitations_deleted_at;
DROP INDEX IF EXISTS idx_user_invitations_invited_by_id;

-- Drop foreign key constraint
ALTER TABLE user_invitations
DROP CONSTRAINT IF EXISTS fk_user_invitations_invited_by_id;

-- Remove added columns
ALTER TABLE user_invitations
DROP COLUMN IF EXISTS deleted_at,
DROP COLUMN IF EXISTS invited_by_id;