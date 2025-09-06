-- Migration: rename_sessions_to_user_sessions
-- Created: 2025-09-06T11:29:48+05:30

-- Step 1: Add new columns to sessions table
ALTER TABLE sessions ADD COLUMN device_info JSONB;
ALTER TABLE sessions ADD COLUMN revoked_at TIMESTAMP WITH TIME ZONE;

-- Step 2: Rename the table
ALTER TABLE sessions RENAME TO user_sessions;

-- Step 3: Drop the old deleted_at column (soft delete approach removed)
ALTER TABLE user_sessions DROP COLUMN deleted_at;

-- Step 4: Update indexes to reflect new table name
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_token;
DROP INDEX IF EXISTS idx_sessions_refresh_token;
DROP INDEX IF EXISTS idx_sessions_is_active;
DROP INDEX IF EXISTS idx_sessions_expires_at;

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token ON user_sessions(token);
CREATE INDEX idx_user_sessions_refresh_token ON user_sessions(refresh_token);
CREATE INDEX idx_user_sessions_is_active ON user_sessions(is_active);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX idx_user_sessions_revoked_at ON user_sessions(revoked_at);

-- Step 5: Update trigger name
DROP TRIGGER IF EXISTS update_sessions_updated_at ON user_sessions;
CREATE TRIGGER update_user_sessions_updated_at BEFORE UPDATE ON user_sessions 
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

