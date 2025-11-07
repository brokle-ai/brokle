-- Migration: remove_onboarding_completed_at_column
-- Created: 2025-11-07T12:40:46+05:30

-- Remove onboarding_completed_at column from users table
-- This column was part of the legacy onboarding system that has been completely removed
ALTER TABLE users DROP COLUMN IF EXISTS onboarding_completed_at;
