-- Rollback: create_volume_discount_tiers_table
-- Created: 2026-01-06T15:35:44+05:30

DROP INDEX IF EXISTS idx_volume_tiers_no_overlap;
DROP INDEX IF EXISTS idx_volume_tiers_dimension;
DROP INDEX IF EXISTS idx_volume_tiers_contract;
DROP TABLE IF EXISTS volume_discount_tiers;
