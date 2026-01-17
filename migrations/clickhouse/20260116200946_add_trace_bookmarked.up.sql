-- Add bookmarked column for trace favorites feature
ALTER TABLE otel_traces ADD COLUMN IF NOT EXISTS
    bookmarked Bool DEFAULT false CODEC(ZSTD(1));

ALTER TABLE otel_traces ADD INDEX IF NOT EXISTS
    idx_bookmarked bookmarked TYPE bloom_filter(0.01) GRANULARITY 1;
