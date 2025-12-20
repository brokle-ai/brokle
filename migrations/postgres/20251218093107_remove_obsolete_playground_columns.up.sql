-- Remove obsolete playground_sessions columns
-- The new architecture uses 'windows' array instead of 'template',
-- and all sessions are saved (no ephemeral sessions with expiration)

-- Remove obsolete indexes first
DROP INDEX IF EXISTS idx_playground_sessions_project_saved;
DROP INDEX IF EXISTS idx_playground_sessions_expires;
DROP INDEX IF EXISTS idx_playground_sessions_tags;

-- Drop columns no longer needed in new architecture
ALTER TABLE playground_sessions DROP COLUMN IF EXISTS template;
ALTER TABLE playground_sessions DROP COLUMN IF EXISTS template_type;
ALTER TABLE playground_sessions DROP COLUMN IF EXISTS is_saved;
ALTER TABLE playground_sessions DROP COLUMN IF EXISTS expires_at;

-- Create new simplified index for listing sessions
CREATE INDEX IF NOT EXISTS idx_playground_sessions_project_used
    ON playground_sessions(project_id, last_used_at DESC);

-- Create new GIN index for tags (no is_saved filter)
CREATE INDEX IF NOT EXISTS idx_playground_sessions_tags
    ON playground_sessions USING GIN(tags);
