-- Revert to non-nullable (convert NULLs to empty strings first)

-- Step 1: Drop indexes
ALTER TABLE scores DROP INDEX IF EXISTS idx_trace_id;
ALTER TABLE scores DROP INDEX IF EXISTS idx_span_id;

-- Step 2: Convert NULLs to empty strings
ALTER TABLE scores UPDATE trace_id = '' WHERE trace_id IS NULL;
ALTER TABLE scores UPDATE span_id = '' WHERE span_id IS NULL;

-- Step 3: Wait for mutations to complete (ClickHouse processes async)
-- Note: In production, verify mutations completed before proceeding

-- Step 4: Modify columns back to non-nullable
ALTER TABLE scores MODIFY COLUMN trace_id String;
ALTER TABLE scores MODIFY COLUMN span_id String;

-- Step 5: Recreate indexes
ALTER TABLE scores ADD INDEX idx_trace_id trace_id TYPE bloom_filter(0.001) GRANULARITY 1;
ALTER TABLE scores ADD INDEX idx_span_id span_id TYPE bloom_filter(0.001) GRANULARITY 1;
