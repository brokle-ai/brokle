-- ============================================================================
-- Rollback: Re-add Trace Aggregation Columns
-- ============================================================================

-- Re-add aggregation columns (if rollback needed)
ALTER TABLE traces
  ADD COLUMN total_cost Nullable(Decimal(18, 9)) AFTER output,
  ADD COLUMN total_tokens Nullable(UInt32) AFTER total_cost,
  ADD COLUMN span_count Nullable(UInt32) AFTER total_tokens;
