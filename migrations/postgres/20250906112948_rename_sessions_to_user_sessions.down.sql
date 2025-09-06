-- Rollback: rename_sessions_to_user_sessions
-- Created: 2025-09-06T11:29:48+05:30

-- Step 1: Drop trigger
DROP TRIGGER IF EXISTS update_user_sessions_updated_at ON user_sessions;

-- Step 2: Drop new indexes
DROP INDEX IF EXISTS idx_user_sessions_user_id;
DROP INDEX IF EXISTS idx_user_sessions_token;
DROP INDEX IF EXISTS idx_user_sessions_refresh_token;
DROP INDEX IF EXISTS idx_user_sessions_is_active;
DROP INDEX IF EXISTS idx_user_sessions_expires_at;
DROP INDEX IF EXISTS idx_user_sessions_revoked_at;

-- Step 3: Add back deleted_at column with GORM DeletedAt structure
ALTER TABLE user_sessions ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;
CREATE INDEX idx_user_sessions_deleted_at ON user_sessions(deleted_at);

-- Step 4: Rename table back to sessions
ALTER TABLE user_sessions RENAME TO sessions;

-- Step 5: Remove new columns
ALTER TABLE sessions DROP COLUMN IF EXISTS device_info;
ALTER TABLE sessions DROP COLUMN IF EXISTS revoked_at;

-- Step 6: Recreate original indexes
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token);
CREATE INDEX idx_sessions_is_active ON sessions(is_active);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Step 7: Recreate original trigger
CREATE TRIGGER update_sessions_updated_at BEFORE UPDATE ON sessions 
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

