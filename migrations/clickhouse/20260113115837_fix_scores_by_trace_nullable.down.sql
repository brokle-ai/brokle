-- ClickHouse Rollback: fix_scores_by_trace_nullable
-- Created: 2026-01-13T11:58:37+05:30
-- Revert to original view without trace_id filter

-- Drop fixed view
DROP VIEW IF EXISTS scores_by_trace;

-- Recreate original view (without trace_id filter)
-- WARNING: This will fail on NULL trace_id inserts if trace_id is nullable
CREATE MATERIALIZED VIEW scores_by_trace
ENGINE = AggregatingMergeTree()
PARTITION BY project_id
ORDER BY (project_id, trace_id, name)
SETTINGS index_granularity = 8192
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
WHERE value IS NOT NULL
GROUP BY project_id, trace_id, name;
