-- Add is_locked column to dashboards table
ALTER TABLE dashboards ADD COLUMN is_locked BOOLEAN NOT NULL DEFAULT false;

-- Create partial index for efficient locked dashboard queries
CREATE INDEX idx_dashboards_is_locked ON dashboards(is_locked) WHERE deleted_at IS NULL;
