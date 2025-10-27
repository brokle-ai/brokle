-- Change observations version column from UInt32 to Nullable(String)
-- For application versioning (experiment tracking)

ALTER TABLE observations DROP COLUMN IF EXISTS version;