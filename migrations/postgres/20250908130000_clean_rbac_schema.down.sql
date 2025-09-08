-- ===================================
-- CLEAN RBAC SCHEMA ROLLBACK
-- (Forward-only approach - minimal rollback)
-- ===================================

-- Note: This is a forward-only migration
-- Rollback is not fully supported by design

-- Drop new tables
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;

-- Drop new indexes and constraints
DROP INDEX IF EXISTS idx_roles_system_name;
DROP INDEX IF EXISTS idx_roles_scoped_name;
DROP INDEX IF EXISTS idx_roles_scope;
ALTER TABLE roles DROP CONSTRAINT IF EXISTS chk_scope_consistency;

-- Remove new columns
ALTER TABLE roles 
    DROP COLUMN IF EXISTS scope_type,
    DROP COLUMN IF EXISTS scope_id;

-- Warning: Original schema structure would need to be manually restored
-- This migration is designed to be forward-only