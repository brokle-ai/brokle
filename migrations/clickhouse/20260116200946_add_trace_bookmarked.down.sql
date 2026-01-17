-- Remove bookmarked column and index
ALTER TABLE otel_traces DROP INDEX IF EXISTS idx_bookmarked;
ALTER TABLE otel_traces DROP COLUMN IF EXISTS bookmarked;
