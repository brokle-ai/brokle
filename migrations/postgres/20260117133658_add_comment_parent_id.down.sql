-- Rollback: add_comment_parent_id
-- Created: 2026-01-17T13:36:58+05:30

DROP INDEX IF EXISTS idx_trace_comments_parent;

ALTER TABLE trace_comments
DROP CONSTRAINT IF EXISTS fk_trace_comments_parent;

ALTER TABLE trace_comments
DROP COLUMN IF EXISTS parent_id;
