-- ClickHouse Migration: rename_brokle_version_to_span_version (rollback)
-- Purpose: Restore brokle_version column with original MATERIALIZED expression

-- Step 1: Drop the new index
ALTER TABLE otel_traces DROP INDEX IF EXISTS idx_span_version;

-- Step 2: Drop the new column
ALTER TABLE otel_traces DROP COLUMN IF EXISTS span_version;

-- Step 3: Restore the original column
ALTER TABLE otel_traces ADD COLUMN IF NOT EXISTS brokle_version LowCardinality(String) MATERIALIZED span_attributes['brokle.version'] CODEC(ZSTD(1));

-- Step 4: Restore the original index
ALTER TABLE otel_traces ADD INDEX IF NOT EXISTS idx_brokle_version brokle_version TYPE bloom_filter(0.01) GRANULARITY 1;
