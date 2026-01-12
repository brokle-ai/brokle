-- ClickHouse Rollback: add_organization_id
-- Created: 2026-01-05T01:10:30+05:30
-- Purpose: Remove organization_id column from otel_traces and scores tables

-- Remove organization_id from scores table
ALTER TABLE scores DROP COLUMN IF EXISTS organization_id;

-- Remove organization_id from otel_traces table
ALTER TABLE otel_traces DROP COLUMN IF EXISTS organization_id;
