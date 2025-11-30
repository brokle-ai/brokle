-- ============================================================================
-- Rollback OTEL-Native Schema Migration
-- ============================================================================
-- Purpose: Drop otel_traces table (for testing/emergency rollback)
-- Warning: This will drop all spans data!
-- ============================================================================

DROP TABLE IF EXISTS otel_traces;

-- Note: Old schema (spans, traces) not recreated
-- If rollback needed, restore from backup or re-run old migrations
