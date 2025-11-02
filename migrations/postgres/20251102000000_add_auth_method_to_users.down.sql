-- Rollback: Remove auth method tracking from users

-- Drop constraints and indexes (in reverse order)
DROP INDEX IF EXISTS idx_users_oauth_provider_unique;
ALTER TABLE users DROP CONSTRAINT IF EXISTS check_oauth_provider_consistency;
DROP INDEX IF EXISTS idx_users_oauth_provider;
DROP INDEX IF EXISTS idx_users_auth_method;

-- Remove columns
ALTER TABLE users DROP COLUMN IF EXISTS oauth_provider_id;
ALTER TABLE users DROP COLUMN IF EXISTS oauth_provider;
ALTER TABLE users DROP COLUMN IF EXISTS auth_method;

-- Restore password as NOT NULL (requires default for rollback safety)
-- Note: This will fail if OAuth users exist with NULL password
-- Manual intervention required if rolling back with OAuth users in database
ALTER TABLE users ALTER COLUMN password SET DEFAULT '';
UPDATE users SET password = '' WHERE password IS NULL;
ALTER TABLE users ALTER COLUMN password SET NOT NULL;
ALTER TABLE users ALTER COLUMN password DROP DEFAULT;
