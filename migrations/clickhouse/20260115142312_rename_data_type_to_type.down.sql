-- ClickHouse Migration: rename_data_type_to_type (rollback)
-- Created: 2026-01-15T14:23:12+05:30

-- Revert column rename
ALTER TABLE scores RENAME COLUMN type TO data_type;
