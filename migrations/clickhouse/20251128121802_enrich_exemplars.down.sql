-- ClickHouse Rollback: enrich_exemplars
-- Created: 2025-11-28T12:18:02+05:30

ALTER TABLE otel_metrics_sum
DROP COLUMN IF EXISTS exemplars_timestamp,
DROP COLUMN IF EXISTS exemplars_value,
DROP COLUMN IF EXISTS exemplars_filtered_attributes;
