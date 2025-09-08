-- Rollback scope-based custom roles to original template-only design

-- Drop scope-based constraints and indexes
DROP INDEX IF EXISTS idx_roles_organization_scope;
DROP INDEX IF EXISTS idx_roles_system_scope;
ALTER TABLE roles DROP CONSTRAINT IF EXISTS unique_role_scope_name;
ALTER TABLE roles DROP CONSTRAINT IF EXISTS chk_roles_scope_type;
ALTER TABLE roles DROP CONSTRAINT IF EXISTS chk_roles_scope_consistency;

-- Organization template roles should already be organization scope
-- Only revert any truly system roles if they were created

-- Remove scope_id column
ALTER TABLE roles DROP COLUMN IF EXISTS scope_id;

-- Restore original unique constraint
ALTER TABLE roles ADD CONSTRAINT roles_name_scope_type_key UNIQUE (name, scope_type);