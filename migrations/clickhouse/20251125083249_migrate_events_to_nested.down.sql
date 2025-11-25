-- Rollback: Restore exploded event arrays from Nested type

ALTER TABLE spans DROP COLUMN events;

-- Restore old exploded arrays (for rollback only)
ALTER TABLE spans
  ADD COLUMN events_timestamp Array(DateTime64(9)) CODEC(ZSTD(1)),
  ADD COLUMN events_name Array(LowCardinality(String)) CODEC(ZSTD(1)),
  ADD COLUMN events_attributes Array(Map(LowCardinality(String), String)) CODEC(ZSTD(1))
