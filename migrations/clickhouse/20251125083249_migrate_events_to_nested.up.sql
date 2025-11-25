-- Migrate events from exploded arrays to Nested type (OTLP Collector standard)
-- Zero users = clean migration, no data preservation needed

ALTER TABLE spans
  DROP COLUMN events_timestamp,
  DROP COLUMN events_name,
  DROP COLUMN events_attributes,
  ADD COLUMN events Nested (
    timestamp DateTime64(9),
    name LowCardinality(String),
    attributes Map(LowCardinality(String), String),
    dropped_attributes_count UInt32
  ) CODEC(ZSTD(1))
