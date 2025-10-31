-- ========================================
-- SCOPE-BASED RBAC ROLLBACK
-- ========================================
-- Reverts scope-based RBAC migration
-- ========================================

-- Remove scope columns from permissions
ALTER TABLE permissions
DROP COLUMN IF EXISTS scope_level,
DROP COLUMN IF EXISTS category;

-- Drop indexes
DROP INDEX IF EXISTS idx_permissions_scope_level;
DROP INDEX IF EXISTS idx_permissions_category;

-- Note: We don't delete the scopes or roles here
-- because data cleanup is done manually via docker exec
-- This just removes the schema changes

-- ========================================
-- Rollback Complete
-- ========================================
--
-- To fully reset to old state:
-- 1. Run this down migration
-- 2. Clean data: docker exec -it brokle-postgres psql -U postgres -d brokle -c "TRUNCATE role_permissions, organization_members, permissions, roles CASCADE;"
-- 3. Run old migrations or seed old data
--
-- ========================================
