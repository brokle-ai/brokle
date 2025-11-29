-- ClickHouse Migration: add_schema_urls_metrics_gauge
-- Created: 2025-11-28T12:27:11+05:30

ALTER TABLE otel_metrics_gauge
ADD COLUMN resource_schema_url Nullable(String) CODEC(ZSTD(1)),
ADD COLUMN scope_schema_url Nullable(String) CODEC(ZSTD(1));
