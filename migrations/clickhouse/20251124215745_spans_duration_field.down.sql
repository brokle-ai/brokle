-- Rollback: Restore duration_ms from duration
ALTER TABLE spans
  DROP COLUMN duration,
  ADD COLUMN duration_ms Nullable(UInt32) CODEC(ZSTD(1)) AFTER end_time
