-- Migration: Add unique partial index for alert deduplication
-- This migration handles existing duplicate alerts by auto-resolving older ones.
--
-- Why duplicates can exist legitimately:
-- 1. Window expiration: After 24h deduplication window expires, a new alert can be
--    created while the old one is still unresolved
-- 2. Historical concurrency: Before this fix, race conditions could create duplicates
-- 3. Long-lived alerts: User never resolved old alerts, new ones triggered after window

-- Step 1: Auto-resolve older duplicate unresolved alerts
-- For each (budget_id, alert_threshold, dimension) combination with multiple
-- unresolved alerts, keep only the most recent one and mark others as resolved.
WITH duplicates AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY budget_id, alert_threshold, dimension
               ORDER BY triggered_at DESC
           ) as rn
    FROM usage_alerts
    WHERE status != 'resolved'
),
older_duplicates AS (
    SELECT id FROM duplicates WHERE rn > 1
)
UPDATE usage_alerts
SET status = 'resolved',
    resolved_at = NOW()
WHERE id IN (SELECT id FROM older_duplicates);

-- Step 2: Create the unique partial index
-- Now safe because duplicates have been resolved.
-- This prevents race conditions where concurrent workers might create
-- duplicate alerts before the hasRecentAlert check can detect them.
CREATE UNIQUE INDEX idx_usage_alerts_dedup
ON usage_alerts (budget_id, alert_threshold, dimension)
WHERE status != 'resolved';
