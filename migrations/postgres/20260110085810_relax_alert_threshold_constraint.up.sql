-- Migration: relax_alert_threshold_constraint
-- Created: 2026-01-10T08:58:10+05:30
-- Remove the overly restrictive upper bound on alert thresholds
-- Allows over-budget warnings (e.g., 150% threshold)

ALTER TABLE usage_alerts
DROP CONSTRAINT IF EXISTS usage_alerts_alert_threshold_check;

ALTER TABLE usage_alerts
ADD CONSTRAINT usage_alerts_alert_threshold_check
CHECK (alert_threshold >= 1);
