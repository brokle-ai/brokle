-- Remove is_locked index and column from dashboards table
DROP INDEX IF EXISTS idx_dashboards_is_locked;
ALTER TABLE dashboards DROP COLUMN IF EXISTS is_locked;
