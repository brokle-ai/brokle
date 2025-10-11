-- ClickHouse Migration: add_environment_to_traces
-- Created: 2025-10-11T23:36:00+05:30

-- Add environment column to traces table
ALTER TABLE traces
ADD COLUMN IF NOT EXISTS environment String DEFAULT 'default';
