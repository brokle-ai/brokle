-- ===================================
-- REMOVE ENVIRONMENT COLUMNS FROM ANALYTICS TABLES
-- ===================================
-- This migration removes environment columns from all ClickHouse analytics tables
-- to rollback the Langfuse-style environment tags.

-- Remove environment column from request_logs table
ALTER TABLE request_logs DROP COLUMN IF EXISTS environment;

-- Remove environment column from metrics table (if exists)
ALTER TABLE metrics DROP COLUMN IF EXISTS environment;

-- Remove environment column from ai_routing_metrics table (if exists)
ALTER TABLE ai_routing_metrics DROP COLUMN IF EXISTS environment;

-- Remove environment column from traces table (if exists)
ALTER TABLE traces DROP COLUMN IF EXISTS environment;

-- Remove environment column from events table (if exists)
ALTER TABLE events DROP COLUMN IF EXISTS environment;