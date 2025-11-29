-- ClickHouse Migration: add_schema_urls_metrics_sum
-- Created: 2025-11-28T12:26:56+05:30
-- Purpose: Add OTLP schema URL fields to otel_metrics_sum

ALTER TABLE otel_metrics_sum
ADD COLUMN resource_schema_url Nullable(String) CODEC(ZSTD(1)),
ADD COLUMN scope_schema_url Nullable(String) CODEC(ZSTD(1));
