-- Add brokle_metadata column to traces table
-- This column stores Brokle-specific system metadata

ALTER TABLE traces
ADD COLUMN IF NOT EXISTS brokle_metadata String DEFAULT '{}';