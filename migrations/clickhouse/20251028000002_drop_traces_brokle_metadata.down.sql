-- Rollback: Re-add brokle_metadata column to traces
ALTER TABLE traces ADD COLUMN IF NOT EXISTS brokle_metadata String DEFAULT '{}';
