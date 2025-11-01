-- Migration: Refactor onboarding to two-step signup
-- Add role and referral_source to users
-- Drop slug columns from organizations and projects
-- Drop onboarding tables

-- Add new user fields (nullable first for safe rollback)
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS referral_source VARCHAR(100);

-- Backfill existing users
UPDATE users SET role = 'Unknown' WHERE role IS NULL;

-- Make role NOT NULL after backfill
ALTER TABLE users ALTER COLUMN role SET NOT NULL;

-- Drop slug columns from organizations and projects (use ULIDs only)
ALTER TABLE organizations DROP COLUMN IF EXISTS slug;
ALTER TABLE projects DROP COLUMN IF EXISTS slug;

-- Check for foreign key dependencies before dropping
-- Run this query manually if needed to verify:
-- SELECT conname, conrelid::regclass
-- FROM pg_constraint
-- WHERE confrelid IN ('onboarding_questions'::regclass, 'onboarding_responses'::regclass);

-- Drop onboarding tables
DROP TABLE IF EXISTS onboarding_responses CASCADE;
DROP TABLE IF EXISTS onboarding_questions CASCADE;

-- Keep onboarding_completed_at for backward compatibility
COMMENT ON COLUMN users.onboarding_completed_at IS
'Legacy field - always set to NOW() on signup. Kept for backward compatibility.';
