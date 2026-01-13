-- Make trace_id and span_id nullable for experiment-only scores
-- Experiment scores don't have traces, so these fields should be optional

-- Step 1: Drop existing indexes that reference these columns
ALTER TABLE scores DROP INDEX IF EXISTS idx_trace_id;
ALTER TABLE scores DROP INDEX IF EXISTS idx_span_id;

-- Step 2: Modify columns to be nullable
ALTER TABLE scores MODIFY COLUMN trace_id Nullable(String);
ALTER TABLE scores MODIFY COLUMN span_id Nullable(String);

-- Step 3: Recreate indexes on nullable columns
ALTER TABLE scores ADD INDEX idx_trace_id trace_id TYPE bloom_filter(0.001) GRANULARITY 1;
ALTER TABLE scores ADD INDEX idx_span_id span_id TYPE bloom_filter(0.001) GRANULARITY 1;
