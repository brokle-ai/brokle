-- Rollback: add_project_members_fk_and_indexes
-- Created: 2026-04-26T14:32:23+05:30

ALTER TABLE project_members DROP CONSTRAINT IF EXISTS project_members_project_id_fkey;
