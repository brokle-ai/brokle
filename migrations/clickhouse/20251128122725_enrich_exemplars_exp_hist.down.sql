-- ClickHouse Rollback: enrich_exemplars_exp_hist
-- Created: 2025-11-28T12:27:25+05:30

ALTER TABLE otel_metrics_exponential_histogram
DROP COLUMN IF EXISTS exemplars_timestamp,
DROP COLUMN IF EXISTS exemplars_value,
DROP COLUMN IF EXISTS exemplars_filtered_attributes;
