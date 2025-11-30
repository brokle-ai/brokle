-- ClickHouse Migration: add_schema_urls_metrics_exp_hist
-- Created: 2025-11-28T12:27:13+05:30

ALTER TABLE otel_metrics_exponential_histogram
ADD COLUMN resource_schema_url Nullable(String) CODEC(ZSTD(1)),
ADD COLUMN scope_schema_url Nullable(String) CODEC(ZSTD(1));
