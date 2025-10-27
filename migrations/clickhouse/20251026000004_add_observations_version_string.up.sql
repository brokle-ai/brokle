-- Add version column back as Nullable(String) for application versioning

ALTER TABLE observations ADD COLUMN IF NOT EXISTS version Nullable(String);