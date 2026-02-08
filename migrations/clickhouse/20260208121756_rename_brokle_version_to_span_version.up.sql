-- ClickHouse Migration: rename_brokle_version_to_span_version
-- Purpose: Rename brokle_version column to span_version and update MATERIALIZED expression
--          from span_attributes['brokle.version'] to span_attributes['brokle.span.version']
-- Note: Cannot use simple RENAME COLUMN because the MATERIALIZED expression also changes.

-- Step 1: Drop the old bloom filter index
ALTER TABLE otel_traces DROP INDEX IF EXISTS idx_brokle_version;

-- Step 2: Drop the old materialized column
ALTER TABLE otel_traces DROP COLUMN IF EXISTS brokle_version;

-- Step 3: Add the new materialized column with updated name and expression
ALTER TABLE otel_traces ADD COLUMN IF NOT EXISTS span_version LowCardinality(String) MATERIALIZED span_attributes['brokle.span.version'] CODEC(ZSTD(1));

-- Step 4: Add the new bloom filter index
ALTER TABLE otel_traces ADD INDEX IF NOT EXISTS idx_span_version span_version TYPE bloom_filter(0.01) GRANULARITY 1;
