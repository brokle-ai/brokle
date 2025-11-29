-- ClickHouse Migration: enrich_exemplars_exp_hist
-- Created: 2025-11-28T12:27:25+05:30

ALTER TABLE otel_metrics_exponential_histogram
ADD COLUMN exemplars_timestamp Array(DateTime64(9)) CODEC(ZSTD(1)),
ADD COLUMN exemplars_value Array(Float64) CODEC(ZSTD(1)),
ADD COLUMN exemplars_filtered_attributes Array(Map(LowCardinality(String), String)) CODEC(ZSTD(1));
