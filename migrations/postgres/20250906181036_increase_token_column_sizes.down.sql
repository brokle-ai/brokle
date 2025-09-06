-- Rollback: increase_token_column_sizes
-- Created: 2025-09-06T18:10:36+05:30
-- Purpose: Rollback token column sizes to original VARCHAR(255)

-- Revert token column sizes back to VARCHAR(255)
-- WARNING: This may truncate data if tokens are longer than 255 characters
ALTER TABLE user_sessions ALTER COLUMN token TYPE VARCHAR(255);
ALTER TABLE user_sessions ALTER COLUMN refresh_token TYPE VARCHAR(255);

-- Revert other token tables
ALTER TABLE password_reset_tokens ALTER COLUMN token TYPE VARCHAR(255);
ALTER TABLE email_verification_tokens ALTER COLUMN token TYPE VARCHAR(255);
ALTER TABLE user_invitations ALTER COLUMN token TYPE VARCHAR(255);

