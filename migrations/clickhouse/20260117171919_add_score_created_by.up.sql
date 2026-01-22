-- Add created_by column to scores table for audit trail
-- Tracks which user created the score (for human annotations)

ALTER TABLE scores ADD COLUMN IF NOT EXISTS created_by Nullable(String) CODEC(ZSTD(1));
