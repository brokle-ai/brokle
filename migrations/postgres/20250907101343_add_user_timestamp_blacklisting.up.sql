-- ===================================================
-- USER-WIDE TIMESTAMP BLACKLISTING MIGRATION
-- ===================================================
-- This migration adds user-wide timestamp blacklisting support
-- to the blacklisted_tokens table for GDPR/SOC2 compliance.
--
-- This ensures that when a user revokes all sessions, ALL tokens
-- issued before a specific timestamp are immediately invalid.

-- Add new columns for user-wide timestamp blacklisting
ALTER TABLE blacklisted_tokens 
ADD COLUMN token_type VARCHAR(50) NOT NULL DEFAULT 'individual',
ADD COLUMN blacklist_timestamp BIGINT;

-- Create optimized index for user timestamp lookups
-- This index is critical for fast middleware token validation
CREATE INDEX idx_blacklisted_tokens_user_timestamp 
ON blacklisted_tokens(user_id, token_type, blacklist_timestamp) 
WHERE token_type = 'user_wide_timestamp';

-- Create general token_type index for query optimization
CREATE INDEX idx_blacklisted_tokens_token_type ON blacklisted_tokens(token_type);

-- Update existing records to have the default token_type
UPDATE blacklisted_tokens SET token_type = 'individual' WHERE token_type IS NULL;

-- Comments for documentation
COMMENT ON COLUMN blacklisted_tokens.token_type IS 'Type of blacklisting: individual (JTI-based) or user_wide_timestamp (all tokens before timestamp)';
COMMENT ON COLUMN blacklisted_tokens.blacklist_timestamp IS 'Unix timestamp for user-wide blacklisting - blocks all tokens issued before this time';