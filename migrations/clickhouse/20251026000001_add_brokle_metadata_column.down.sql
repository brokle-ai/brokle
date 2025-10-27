-- Rollback: Remove brokle_metadata column from observations
ALTER TABLE observations DROP COLUMN IF EXISTS brokle_metadata;