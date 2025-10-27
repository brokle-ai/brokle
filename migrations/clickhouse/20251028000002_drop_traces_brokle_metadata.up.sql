-- Drop brokle_metadata column from traces table
-- Brokle-specific data is stored in attributes column with brokle.* namespace (OTEL-native)

ALTER TABLE traces DROP COLUMN IF EXISTS brokle_metadata;
