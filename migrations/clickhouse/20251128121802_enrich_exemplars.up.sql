-- ClickHouse Migration: enrich_exemplars
-- Created: 2025-11-28T12:18:02+05:30
-- Purpose: Add additional exemplar fields to otel_metrics_sum for better metric → trace correlation

ALTER TABLE otel_metrics_sum
ADD COLUMN exemplars_timestamp Array(DateTime64(9)) CODEC(ZSTD(1)),
ADD COLUMN exemplars_value Array(Float64) CODEC(ZSTD(1)),
ADD COLUMN exemplars_filtered_attributes Array(Map(LowCardinality(String), String)) CODEC(ZSTD(1));
