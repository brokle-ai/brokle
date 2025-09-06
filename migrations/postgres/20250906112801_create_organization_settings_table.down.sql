-- Rollback: create_organization_settings_table
-- Created: 2025-09-06T11:28:01+05:30

-- Drop trigger
DROP TRIGGER IF EXISTS update_organization_settings_updated_at ON organization_settings;

-- Drop indexes
DROP INDEX IF EXISTS idx_organization_settings_organization_id;
DROP INDEX IF EXISTS idx_organization_settings_key;

-- Drop table
DROP TABLE IF EXISTS organization_settings;

