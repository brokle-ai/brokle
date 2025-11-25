-- Rollback: Restore exploded link arrays from Nested type

ALTER TABLE spans DROP COLUMN links;

-- Restore old exploded arrays (for rollback only)
ALTER TABLE spans
  ADD COLUMN links_trace_id Array(String) CODEC(ZSTD(1)),
  ADD COLUMN links_span_id Array(String) CODEC(ZSTD(1)),
  ADD COLUMN links_trace_state Array(String) CODEC(ZSTD(1)),
  ADD COLUMN links_attributes Array(Map(LowCardinality(String), String)) CODEC(ZSTD(1))
