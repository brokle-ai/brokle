-- Add brokle_metadata column to observations table
-- This column stores Brokle-specific system metadata (routing, cache, governance)
-- The existing metadata column is repurposed for pure OTEL metadata (resource + scope)

ALTER TABLE observations
ADD COLUMN IF NOT EXISTS brokle_metadata String DEFAULT '{}';