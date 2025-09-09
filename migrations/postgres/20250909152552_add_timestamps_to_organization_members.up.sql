-- Add missing standard GORM timestamp fields to organization_members table
-- This fixes the schema mismatch between database and domain models

-- Add missing timestamp columns
ALTER TABLE organization_members 
ADD COLUMN created_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

-- Set default values for existing records
-- Use joined_at as the created_at timestamp for existing members
UPDATE organization_members 
SET created_at = joined_at, 
    updated_at = joined_at 
WHERE created_at IS NULL;

-- Set NOT NULL constraints and defaults for future records
ALTER TABLE organization_members 
ALTER COLUMN created_at SET NOT NULL,
ALTER COLUMN created_at SET DEFAULT NOW(),
ALTER COLUMN updated_at SET NOT NULL,
ALTER COLUMN updated_at SET DEFAULT NOW();

-- Add performance index for deleted_at (GORM soft delete queries)
CREATE INDEX idx_organization_members_deleted_at ON organization_members(deleted_at);

-- Add updated_at trigger for automatic timestamp updates
CREATE TRIGGER update_organization_members_updated_at 
    BEFORE UPDATE ON organization_members 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();