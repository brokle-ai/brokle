-- ClickHouse Migration: add_environment_to_events
-- Created: 2025-10-11T23:36:00+05:30

-- Add environment column to events table
ALTER TABLE events
ADD COLUMN IF NOT EXISTS environment String DEFAULT 'default';
