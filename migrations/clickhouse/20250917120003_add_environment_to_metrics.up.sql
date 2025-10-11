-- ClickHouse Migration: add_environment_to_metrics
-- Created: 2025-10-11T23:36:00+05:30

-- Add environment column to metrics table
ALTER TABLE metrics
ADD COLUMN IF NOT EXISTS environment String DEFAULT 'default';
