-- Rollback: Remove timestamp fields from organization_members table

-- Drop the updated_at trigger
DROP TRIGGER IF EXISTS update_organization_members_updated_at ON organization_members;

-- Drop the deleted_at index
DROP INDEX IF EXISTS idx_organization_members_deleted_at;

-- Remove the added timestamp columns
ALTER TABLE organization_members 
DROP COLUMN IF EXISTS created_at,
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS deleted_at;