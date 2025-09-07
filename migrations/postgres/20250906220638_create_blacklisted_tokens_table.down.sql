-- ===================================
-- ROLLBACK BLACKLISTED TOKENS TABLE
-- ===================================
-- This rollback drops the blacklisted_tokens table.

-- Drop the blacklisted_tokens table
DROP TABLE IF EXISTS blacklisted_tokens CASCADE;