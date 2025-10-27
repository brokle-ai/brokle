-- Drop brokle_metadata column from observations table
-- Brokle-specific data is stored in attributes column with brokle.* namespace (OTEL-native)

ALTER TABLE observations DROP COLUMN IF EXISTS brokle_metadata;
