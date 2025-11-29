-- ClickHouse Rollback: add_schema_urls_metrics_histogram
-- Created: 2025-11-28T12:27:10+05:30

ALTER TABLE otel_metrics_histogram
DROP COLUMN IF EXISTS resource_schema_url,
DROP COLUMN IF EXISTS scope_schema_url;
