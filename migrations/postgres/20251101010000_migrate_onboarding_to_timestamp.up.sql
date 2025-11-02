-- Migrate onboarding from boolean to timestamp
-- Single field: onboarding_completed_at (timestamp)
-- Logic: NULL = not completed, NOT NULL = completed

-- Step 1: Add onboarding_completed_at column
ALTER TABLE users
ADD COLUMN IF NOT EXISTS onboarding_completed_at TIMESTAMP WITH TIME ZONE;

-- Step 2: Migrate existing data
-- For users who have onboarding_completed = true, set timestamp to updated_at
UPDATE users
SET onboarding_completed_at = updated_at
WHERE onboarding_completed = true;

-- Step 3: Drop old boolean column
ALTER TABLE users
DROP COLUMN IF EXISTS onboarding_completed;

-- Step 4: Add index for queries
CREATE INDEX IF NOT EXISTS idx_users_onboarding_completed_at ON users(onboarding_completed_at);

-- Step 5: Add comment
COMMENT ON COLUMN users.onboarding_completed_at IS 'Timestamp when user completed onboarding. NULL = not completed';
