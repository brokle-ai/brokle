-- Replace duration_ms with duration (nanoseconds) for OTLP spec compliance
ALTER TABLE traces
  DROP COLUMN duration_ms,
  ADD COLUMN duration Nullable(UInt64) CODEC(ZSTD(1)) AFTER end_time
