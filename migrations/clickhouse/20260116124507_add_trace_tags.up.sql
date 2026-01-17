-- Add tags column for user-managed trace tagging
-- Distinct from OTLP instrumentation data - these are platform-level tags
ALTER TABLE otel_traces ADD COLUMN IF NOT EXISTS
    tags Array(LowCardinality(String)) DEFAULT [] CODEC(ZSTD(1));

-- Bloom filter index for efficient containment queries (e.g., has('tag'))
ALTER TABLE otel_traces ADD INDEX IF NOT EXISTS
    idx_tags tags TYPE bloom_filter(0.01) GRANULARITY 1;
