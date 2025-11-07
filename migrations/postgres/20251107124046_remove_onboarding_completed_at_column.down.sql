-- Rollback: remove_onboarding_completed_at_column
-- Created: 2025-11-07T12:40:46+05:30

-- Restore onboarding_completed_at column for rollback
-- Note: Data will be lost and not restored
ALTER TABLE users ADD COLUMN onboarding_completed_at TIMESTAMP WITH TIME ZONE;
