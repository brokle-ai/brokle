-- Revert playground_sessions table

DROP INDEX IF EXISTS idx_playground_sessions_tags;
DROP INDEX IF EXISTS idx_playground_sessions_expires;
DROP INDEX IF EXISTS idx_playground_sessions_project_saved;
DROP TABLE IF EXISTS playground_sessions;
