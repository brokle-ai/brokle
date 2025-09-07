-- Rollback RBAC System Global System Roles Update
-- This migration rolls back changes to revert to organization-specific roles only

-- Step 1: Remove performance indexes
DROP INDEX IF EXISTS idx_org_members_role_lookup;
DROP INDEX IF EXISTS idx_role_permissions_lookup;
DROP INDEX IF EXISTS idx_permissions_resource_action;
DROP INDEX IF EXISTS idx_permissions_resource;
DROP INDEX IF EXISTS idx_roles_system_global;

-- Step 2: Remove resource:action unique constraint
ALTER TABLE permissions DROP CONSTRAINT IF EXISTS unique_resource_action;

-- Step 3: Remove resource and action columns from permissions
ALTER TABLE permissions DROP COLUMN IF EXISTS action;
ALTER TABLE permissions DROP COLUMN IF EXISTS resource;

-- Step 4: Remove unique constraints for role names
DROP INDEX IF EXISTS unique_org_role;
DROP INDEX IF EXISTS unique_global_system_role;

-- Step 5: Remove system role organization constraint
ALTER TABLE roles DROP CONSTRAINT IF EXISTS system_role_organization_check;

-- Step 6: Re-add original unique constraint for organization roles
CREATE UNIQUE INDEX idx_roles_unique_org_name ON roles(organization_id, name);

-- Step 7: Make organization_id required again
ALTER TABLE roles ALTER COLUMN organization_id SET NOT NULL;

-- Step 8: Clean up any system roles that may have been created with NULL organization_id
-- Note: This assumes no important system roles exist, adjust as needed
DELETE FROM roles WHERE organization_id IS NULL AND is_system_role = true;

-- Rollback completed - RBAC system reverted to organization-specific roles only