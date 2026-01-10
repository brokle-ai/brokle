-- Rollback: relax_alert_threshold_constraint
-- Created: 2026-01-10T08:58:10+05:30
-- Restore original constraint (note: will fail if values >100 exist)

ALTER TABLE usage_alerts
DROP CONSTRAINT IF EXISTS usage_alerts_alert_threshold_check;

ALTER TABLE usage_alerts
ADD CONSTRAINT usage_alerts_alert_threshold_check
CHECK (alert_threshold >= 1 AND alert_threshold <= 100);
