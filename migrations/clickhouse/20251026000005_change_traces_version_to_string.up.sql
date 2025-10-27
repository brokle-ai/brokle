-- Change traces version column from UInt32 to Nullable(String)

ALTER TABLE traces DROP COLUMN IF EXISTS version;