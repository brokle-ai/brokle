-- Remove created_by column from scores table

ALTER TABLE scores DROP COLUMN IF EXISTS created_by;
