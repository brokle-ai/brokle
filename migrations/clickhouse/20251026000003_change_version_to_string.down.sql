-- Rollback: Restore version column to UInt32 (database versioning)

ALTER TABLE observations DROP COLUMN IF EXISTS version;
ALTER TABLE observations ADD COLUMN IF NOT EXISTS version UInt32 DEFAULT 1;

ALTER TABLE traces DROP COLUMN IF EXISTS version;
ALTER TABLE traces ADD COLUMN IF NOT EXISTS version UInt32 DEFAULT 1;