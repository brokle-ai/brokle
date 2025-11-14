-- ============================================================================
-- Remove Trace Aggregation Columns - Industry Standard Pattern
-- ============================================================================
-- Purpose: Remove denormalized aggregation columns from traces table
-- Rationale: Calculate on-demand from spans (Langfuse/Datadog/Honeycomb pattern)
-- Performance: ClickHouse materialized column aggregation = 10-50ms for 1000 spans
-- Migration: Brokle v2.0 → Industry standard architecture
-- ============================================================================

-- Remove denormalized aggregation columns
-- Aggregations will be calculated on-demand from spans table using materialized columns
ALTER TABLE traces
  DROP COLUMN total_cost,
  DROP COLUMN total_tokens,
  DROP COLUMN span_count;

-- Note: Aggregations now calculated in API layer:
-- SELECT
--   SUM(brokle_cost_total) as total_cost,
--   SUM(gen_ai_usage_input_tokens + gen_ai_usage_output_tokens) as total_tokens,
--   COUNT(*) as span_count
-- FROM spans WHERE trace_id = ?
