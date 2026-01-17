-- Remove tags index and column
ALTER TABLE otel_traces DROP INDEX IF EXISTS idx_tags;
ALTER TABLE otel_traces DROP COLUMN IF EXISTS tags;
