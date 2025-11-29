-- ClickHouse Rollback: add_schema_urls_logs
-- Created: 2025-11-28T12:27:14+05:30

ALTER TABLE otel_logs
DROP COLUMN IF EXISTS resource_schema_url,
DROP COLUMN IF EXISTS scope_schema_url;
