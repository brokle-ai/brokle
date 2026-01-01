-- ClickHouse Migration Rollback: add_text_search_indexes
-- Created: 2026-01-01T10:57:04+05:30
-- Purpose: Remove full-text search indexes and preview columns

-- Drop the indexes first
ALTER TABLE otel_traces DROP INDEX IF EXISTS idx_input_tokens;
ALTER TABLE otel_traces DROP INDEX IF EXISTS idx_output_tokens;

-- Drop the preview columns
ALTER TABLE otel_traces DROP COLUMN IF EXISTS input_preview;
ALTER TABLE otel_traces DROP COLUMN IF EXISTS output_preview;
