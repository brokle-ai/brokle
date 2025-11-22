DROP INDEX IF EXISTS idx_projects_org_status_deleted;
DROP INDEX IF EXISTS idx_projects_status;
ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_status_check;
ALTER TABLE projects DROP COLUMN IF EXISTS status;
