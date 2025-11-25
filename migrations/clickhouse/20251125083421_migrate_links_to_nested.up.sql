-- Migrate links from exploded arrays to Nested type (OTLP Collector standard)
-- Zero users = clean migration, no data preservation needed

ALTER TABLE spans
  DROP COLUMN links_trace_id,
  DROP COLUMN links_span_id,
  DROP COLUMN links_trace_state,
  DROP COLUMN links_attributes,
  ADD COLUMN links Nested (
    trace_id String,
    span_id String,
    trace_state String,
    attributes Map(LowCardinality(String), String),
    dropped_attributes_count UInt32
  ) CODEC(ZSTD(1))
