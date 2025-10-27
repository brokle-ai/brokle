-- Rollback: Re-add brokle_metadata column to observations
ALTER TABLE observations ADD COLUMN IF NOT EXISTS brokle_metadata String DEFAULT '{}';
