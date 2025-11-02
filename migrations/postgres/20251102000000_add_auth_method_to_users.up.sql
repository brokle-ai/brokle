-- Migration: Add auth method tracking to users
-- Adds auth_method, oauth_provider, and oauth_provider_id columns to distinguish
-- between password-based and OAuth-based authentication

-- Add auth method tracking
ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_method VARCHAR(20) DEFAULT 'password';
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_provider_id VARCHAR(255);

-- Create indexes for efficient OAuth lookups
CREATE INDEX IF NOT EXISTS idx_users_auth_method ON users(auth_method);
CREATE INDEX IF NOT EXISTS idx_users_oauth_provider ON users(oauth_provider, oauth_provider_id);

-- Backfill existing users (assume password-based auth)
UPDATE users SET auth_method = 'password' WHERE auth_method IS NULL OR auth_method = '';

-- Make password nullable for OAuth users (who don't have passwords)
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;

-- Add helpful comments
COMMENT ON COLUMN users.auth_method IS 'Authentication method: password | oauth';
COMMENT ON COLUMN users.oauth_provider IS 'OAuth provider name: google | github | etc';
COMMENT ON COLUMN users.oauth_provider_id IS 'Unique ID from OAuth provider';

-- Add database constraints for data consistency
-- Ensure OAuth users have provider info, password users don't
ALTER TABLE users ADD CONSTRAINT check_oauth_provider_consistency
CHECK (
    (auth_method = 'oauth' AND oauth_provider IS NOT NULL AND oauth_provider_id IS NOT NULL) OR
    (auth_method = 'password')
);

-- Ensure unique OAuth provider + provider_id combinations (prevent duplicate OAuth accounts)
CREATE UNIQUE INDEX idx_users_oauth_provider_unique
ON users(oauth_provider, oauth_provider_id)
WHERE auth_method = 'oauth' AND oauth_provider IS NOT NULL;

