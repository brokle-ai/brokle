-- ClickHouse Rollback: add_schema_urls_metrics_sum
-- Created: 2025-11-28T12:26:56+05:30

ALTER TABLE otel_metrics_sum
DROP COLUMN IF EXISTS resource_schema_url,
DROP COLUMN IF EXISTS scope_schema_url;
