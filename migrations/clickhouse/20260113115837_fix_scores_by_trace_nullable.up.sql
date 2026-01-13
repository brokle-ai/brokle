-- ClickHouse Migration: fix_scores_by_trace_nullable
-- Created: 2026-01-13T11:58:37+05:30
-- Fix scores_by_trace materialized view to handle nullable trace_id
-- After making trace_id nullable, the view fails on NULL inserts

-- Drop existing view
DROP VIEW IF EXISTS scores_by_trace;

-- Recreate with trace_id IS NOT NULL filter and allow_nullable_key
-- The filter ensures only non-null trace_ids are included
-- allow_nullable_key is required because source column is now Nullable
-- Experiment-only scores (NULL trace_id) use scores_by_experiment view instead
CREATE MATERIALIZED VIEW scores_by_trace
ENGINE = AggregatingMergeTree()
PARTITION BY project_id
ORDER BY (project_id, trace_id, name)
SETTINGS index_granularity = 8192, allow_nullable_key = 1
POPULATE
AS SELECT
    project_id,
    trace_id,
    name,
    countState() AS count_state,
    sumState(value) AS sum_state,
    minState(value) AS min_state,
    maxState(value) AS max_state
FROM scores
WHERE value IS NOT NULL AND trace_id IS NOT NULL AND trace_id != ''
GROUP BY project_id, trace_id, name;
