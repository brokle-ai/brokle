-- ClickHouse Migration: rename_data_type_to_type
-- Created: 2026-01-15T14:23:12+05:30
-- Purpose: Rename data_type column to type in scores table for cleaner API naming

-- Rename the column (ClickHouse 23.7+ supports RENAME COLUMN)
ALTER TABLE scores RENAME COLUMN data_type TO type;
