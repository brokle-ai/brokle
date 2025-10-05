-- ===================================
-- ROLLBACK SCOPE-BASED CUSTOM ROLES
-- ===================================
-- Restore state after 20250908140000_normalized_rbac_schema was applied
-- (before scope_id was added for custom roles)

-- Drop scope-based constraints and indexes added by this migration
DROP INDEX IF EXISTS idx_roles_organization_scope;
DROP INDEX IF EXISTS idx_roles_system_scope;
ALTER TABLE roles DROP CONSTRAINT IF EXISTS unique_role_scope_name;
ALTER TABLE roles DROP CONSTRAINT IF EXISTS chk_roles_scope_type;
ALTER TABLE roles DROP CONSTRAINT IF EXISTS chk_roles_scope_consistency;

-- Remove scope_id column added by this migration
ALTER TABLE roles DROP COLUMN IF EXISTS scope_id;

-- Restore the original unique constraint from 20250908140000
ALTER TABLE roles ADD CONSTRAINT roles_name_scope_type_key UNIQUE (name, scope_type);