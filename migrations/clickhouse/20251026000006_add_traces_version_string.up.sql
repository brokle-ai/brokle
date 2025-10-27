-- Add version column back as Nullable(String) for application versioning

ALTER TABLE traces ADD COLUMN IF NOT EXISTS version Nullable(String);