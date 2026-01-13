-- Rollback: create_dataset_versions
-- Created: 2026-01-11T21:30:53+05:30

-- Remove column from datasets table first (due to foreign key)
ALTER TABLE datasets DROP COLUMN IF EXISTS current_version_id;

-- Drop indexes
DROP INDEX IF EXISTS idx_datasets_current_version_id;
DROP INDEX IF EXISTS idx_dataset_item_versions_item_id;
DROP INDEX IF EXISTS idx_dataset_item_versions_version_id;
DROP INDEX IF EXISTS idx_dataset_versions_created_at;
DROP INDEX IF EXISTS idx_dataset_versions_dataset_id;

-- Drop tables (join table first due to foreign keys)
DROP TABLE IF EXISTS dataset_item_versions;
DROP TABLE IF EXISTS dataset_versions;
