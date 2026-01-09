-- Rollback: create_contracts_table
-- Created: 2026-01-06T15:34:48+05:30

DROP TRIGGER IF EXISTS update_contracts_updated_at ON contracts;
DROP INDEX IF EXISTS idx_contracts_active_org;
DROP INDEX IF EXISTS idx_contracts_dates;
DROP INDEX IF EXISTS idx_contracts_status;
DROP INDEX IF EXISTS idx_contracts_organization;
DROP TABLE IF EXISTS contracts;
