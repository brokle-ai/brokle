-- ===================================================
-- ROLLBACK USER-WIDE TIMESTAMP BLACKLISTING MIGRATION
-- ===================================================
-- This migration removes user-wide timestamp blacklisting support
-- from the blacklisted_tokens table.

-- Drop the specialized indexes
DROP INDEX IF EXISTS idx_blacklisted_tokens_user_timestamp;
DROP INDEX IF EXISTS idx_blacklisted_tokens_token_type;

-- Remove the new columns
ALTER TABLE blacklisted_tokens 
DROP COLUMN IF EXISTS blacklist_timestamp,
DROP COLUMN IF EXISTS token_type;