-- Add has_error semantic flag for clean error queries (no magic numbers)
-- Materialized from status_code: true when ERROR (2), false otherwise (0,1)
ALTER TABLE traces ADD COLUMN
    has_error Bool MATERIALIZED status_code = 2 CODEC(ZSTD(1))
