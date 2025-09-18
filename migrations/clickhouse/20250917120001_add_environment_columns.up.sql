-- ===================================
-- ADD ENVIRONMENT COLUMNS TO ANALYTICS TABLES
-- ===================================
-- This migration adds environment columns to all ClickHouse analytics tables
-- to support Langfuse-style environment tags.

-- Add environment column to request_logs table
ALTER TABLE request_logs ADD COLUMN IF NOT EXISTS environment String DEFAULT 'default';

-- Add environment column to metrics table (if exists)
ALTER TABLE metrics ADD COLUMN IF NOT EXISTS environment String DEFAULT 'default';

-- Add environment column to ai_routing_metrics table (if exists)
ALTER TABLE ai_routing_metrics ADD COLUMN IF NOT EXISTS environment String DEFAULT 'default';

-- Add environment column to traces table (if exists)
ALTER TABLE traces ADD COLUMN IF NOT EXISTS environment String DEFAULT 'default';

-- Add environment column to events table (if exists)
ALTER TABLE events ADD COLUMN IF NOT EXISTS environment String DEFAULT 'default';

-- Update existing data to have 'default' environment
-- (This is safe since we're adding with DEFAULT, but being explicit)

-- Set all existing records to 'default' environment
ALTER TABLE request_logs UPDATE environment = 'default' WHERE environment = '';
ALTER TABLE metrics UPDATE environment = 'default' WHERE environment = '';
ALTER TABLE ai_routing_metrics UPDATE environment = 'default' WHERE environment = '';
ALTER TABLE traces UPDATE environment = 'default' WHERE environment = '';
ALTER TABLE events UPDATE environment = 'default' WHERE environment = '';

-- Add comments to document the purpose
ALTER TABLE request_logs COMMENT COLUMN environment 'Environment tag for request (Langfuse-style): default, production, staging, development, etc.';
ALTER TABLE metrics COMMENT COLUMN environment 'Environment tag for metric (Langfuse-style): default, production, staging, development, etc.';
ALTER TABLE ai_routing_metrics COMMENT COLUMN environment 'Environment tag for AI routing decision (Langfuse-style): default, production, staging, development, etc.';
ALTER TABLE traces COMMENT COLUMN environment 'Environment tag for trace (Langfuse-style): default, production, staging, development, etc.';
ALTER TABLE events COMMENT COLUMN environment 'Environment tag for event (Langfuse-style): default, production, staging, development, etc.';