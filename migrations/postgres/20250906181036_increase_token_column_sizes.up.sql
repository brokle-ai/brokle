-- Migration: increase_token_column_sizes
-- Created: 2025-09-06T18:10:36+05:30
-- Purpose: Increase token column sizes to accommodate longer JWT tokens

-- Increase token column sizes in user_sessions table
-- JWT tokens can be 300-500+ characters, so using TEXT for unlimited length
ALTER TABLE user_sessions ALTER COLUMN token TYPE TEXT;
ALTER TABLE user_sessions ALTER COLUMN refresh_token TYPE TEXT;

-- Also update other token tables for consistency
ALTER TABLE password_reset_tokens ALTER COLUMN token TYPE TEXT;
ALTER TABLE email_verification_tokens ALTER COLUMN token TYPE TEXT;
ALTER TABLE user_invitations ALTER COLUMN token TYPE TEXT;

