-- Rollback onboarding timestamp migration

-- Step 1: Add back boolean column
ALTER TABLE users
ADD COLUMN IF NOT EXISTS onboarding_completed BOOLEAN DEFAULT FALSE;

-- Step 2: Migrate data back
UPDATE users
SET onboarding_completed = true
WHERE onboarding_completed_at IS NOT NULL;

-- Step 3: Drop timestamp column
ALTER TABLE users
DROP COLUMN IF EXISTS onboarding_completed_at;

-- Step 4: Drop index
DROP INDEX IF EXISTS idx_users_onboarding_completed_at;
