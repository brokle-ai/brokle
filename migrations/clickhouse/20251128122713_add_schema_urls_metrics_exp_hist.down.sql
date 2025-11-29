-- ClickHouse Rollback: add_schema_urls_metrics_exp_hist
-- Created: 2025-11-28T12:27:13+05:30

ALTER TABLE otel_metrics_exponential_histogram
DROP COLUMN IF EXISTS resource_schema_url,
DROP COLUMN IF EXISTS scope_schema_url;
