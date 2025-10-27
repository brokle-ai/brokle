-- Rollback: Remove brokle_metadata column from traces
ALTER TABLE traces DROP COLUMN IF EXISTS brokle_metadata;