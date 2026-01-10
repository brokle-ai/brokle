-- Rollback: create_contract_history_table
-- Created: 2026-01-06T15:37:49+05:30

DROP INDEX IF EXISTS idx_contract_history_changed_at;
DROP INDEX IF EXISTS idx_contract_history_contract;
DROP TABLE IF EXISTS contract_history;
