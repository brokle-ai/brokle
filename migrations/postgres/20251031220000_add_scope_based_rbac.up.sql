-- ========================================
-- SCOPE-BASED RBAC MIGRATION
-- ========================================
-- Adds schema changes for scope-based RBAC
-- Data seeding is handled by Go seeder (seeds/dev.yaml)
--
-- This migration ONLY adds columns and indexes
-- No INSERT statements (clean separation of schema vs data)
-- ========================================

-- Step 1: Add new columns to permissions table
ALTER TABLE permissions
ADD COLUMN IF NOT EXISTS scope_level VARCHAR(20) NOT NULL DEFAULT 'organization',
ADD COLUMN IF NOT EXISTS category VARCHAR(50);

-- Step 2: Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_permissions_scope_level ON permissions(scope_level);
CREATE INDEX IF NOT EXISTS idx_permissions_category ON permissions(category);

-- Step 3: Add helpful comments
COMMENT ON COLUMN permissions.scope_level IS 'Scope level: organization (org-wide), project (project-specific), or global (platform admin)';
COMMENT ON COLUMN permissions.category IS 'Permission category for grouping: organization, members, billing, projects, observability, gateway, etc.';

-- ========================================
-- Migration Complete
-- ========================================
--
-- Next Steps:
-- 1. Clean existing data (manual):
--    docker exec -it brokle-postgres psql -U postgres -d brokle -c "TRUNCATE role_permissions, organization_members, permissions, roles CASCADE;"
--
-- 2. Run migration (schema changes):
--    make migrate-up
--
-- 3. Seed data (roles, scopes, users, orgs):
--    make seed-dev
--
-- 4. Verify scopes seeded correctly:
--    docker exec -it brokle-postgres psql -U postgres -d brokle -c "SELECT * FROM v_role_scope_mappings LIMIT 20;"
--
-- Expected Results After Seeding:
-- - 63 scopes total (40 organization, 20 project)
-- - 4 roles: owner (63 scopes), admin (61), developer (30), viewer (15)
--
-- ========================================
