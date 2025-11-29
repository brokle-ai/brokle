-- ClickHouse Migration: add_schema_urls_logs
-- Created: 2025-11-28T12:27:14+05:30

ALTER TABLE otel_logs
ADD COLUMN resource_schema_url Nullable(String) CODEC(ZSTD(1)),
ADD COLUMN scope_schema_url Nullable(String) CODEC(ZSTD(1));
