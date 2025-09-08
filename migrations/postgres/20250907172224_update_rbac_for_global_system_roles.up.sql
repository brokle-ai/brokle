-- Update RBAC System for Global System Roles
-- This migration enables global system roles (organization_id = NULL) while maintaining organization-specific custom roles

-- Step 1: Allow NULL organization_id for global system roles
ALTER TABLE roles ALTER COLUMN organization_id DROP NOT NULL;

-- Step 2: Drop existing unique constraint that requires organization_id
DROP INDEX IF EXISTS idx_roles_unique_org_name;

-- Step 3: Add constraint to ensure proper role organization assignment
-- System roles must be global (org_id = NULL), custom roles must be org-specific (org_id != NULL)
ALTER TABLE roles ADD CONSTRAINT system_role_organization_check 
CHECK (
    (is_system_role = true AND organization_id IS NULL) OR     -- Global system roles
    (is_system_role = false AND organization_id IS NOT NULL)   -- Org-specific custom roles
);

-- Step 4: Add unique constraint for global system role names
-- Ensures we can't have duplicate global system roles
CREATE UNIQUE INDEX unique_global_system_role 
ON roles (name) WHERE (is_system_role = true AND organization_id IS NULL);

-- Step 5: Add unique constraint for org-specific role names
-- Ensures we can't have duplicate role names within the same organization  
CREATE UNIQUE INDEX unique_org_role
ON roles (organization_id, name) WHERE (organization_id IS NOT NULL);

-- Step 6: Update permissions table for resource:action model
-- Add new columns for resource:action permission format
ALTER TABLE permissions ADD COLUMN resource VARCHAR(50);
ALTER TABLE permissions ADD COLUMN action VARCHAR(50);

-- Step 7: Update existing permissions to resource:action format
-- Convert simple permission names to resource:action format
UPDATE permissions SET 
    resource = CASE 
        WHEN name LIKE 'users.%' THEN 'users'
        WHEN name LIKE 'organizations.%' THEN 'organizations'
        WHEN name LIKE 'projects.%' THEN 'projects'
        WHEN name LIKE 'environments.%' THEN 'environments'
        WHEN name LIKE 'billing.%' THEN 'billing'
        WHEN name LIKE 'analytics.%' THEN 'analytics'
        WHEN name LIKE 'api_keys.%' THEN 'api_keys'
        ELSE 'system'
    END,
    action = CASE
        WHEN name LIKE '%.create' THEN 'create'
        WHEN name LIKE '%.read' THEN 'read'
        WHEN name LIKE '%.update' THEN 'update'
        WHEN name LIKE '%.delete' THEN 'delete'
        WHEN name LIKE '%.manage' THEN 'manage'
        WHEN name LIKE '%.admin' THEN 'admin'
        ELSE 'access'
    END
WHERE resource IS NULL OR action IS NULL;

-- Step 8: Add constraints for resource:action format
ALTER TABLE permissions ALTER COLUMN resource SET NOT NULL;
ALTER TABLE permissions ALTER COLUMN action SET NOT NULL;

-- Step 9: Add unique constraint for resource:action combinations
ALTER TABLE permissions ADD CONSTRAINT unique_resource_action UNIQUE(resource, action);

-- Step 10: Add performance indexes for efficient RBAC queries
-- Optimize system role lookups
CREATE INDEX idx_roles_system_global 
ON roles (is_system_role, organization_id) WHERE is_system_role = true;

-- Optimize permission lookups by resource
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_resource_action ON permissions(resource, action);

-- Optimize role-permission relationship queries
CREATE INDEX idx_role_permissions_lookup ON role_permissions(role_id);

-- Optimize organization member role lookups
CREATE INDEX IF NOT EXISTS idx_org_members_role_lookup
ON organization_members (user_id, organization_id, role_id);

-- Step 11: Add updated_at triggers for new columns
-- Ensure proper timestamp management for audit trails
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Migration completed successfully
-- Next step: Update domain entities and service implementations to use new schema