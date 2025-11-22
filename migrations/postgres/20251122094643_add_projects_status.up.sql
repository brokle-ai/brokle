-- Add status column to projects table (active/archived only)
ALTER TABLE projects
ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active';

-- Add check constraint for valid values (only active/archived)
ALTER TABLE projects
ADD CONSTRAINT projects_status_check
CHECK (status IN ('active', 'archived'));

-- Add index for filtering by status
CREATE INDEX idx_projects_status ON projects(status);

-- Add composite index for common queries (org + status + soft delete)
CREATE INDEX idx_projects_org_status_deleted
ON projects(organization_id, status, deleted_at);
