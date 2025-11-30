-- ClickHouse Rollback: enrich_exemplars_gauge
-- Created: 2025-11-28T12:27:24+05:30

ALTER TABLE otel_metrics_gauge
DROP COLUMN IF EXISTS exemplars_timestamp,
DROP COLUMN IF EXISTS exemplars_value,
DROP COLUMN IF EXISTS exemplars_filtered_attributes;
