-- Migration: Industry-Standard API Keys (Pure Random)
-- Migrates from project-scoped keys (bk_proj_{project_id}_{secret})
-- to pure random keys (bk_{random_secret}) following GitHub/Stripe pattern
--
-- Key Changes:
-- - API key format: bk_{40_char_random_secret}
-- - SHA-256 hashing (not bcrypt - deterministic for O(1) lookup)
-- - Direct hash lookup via unique index on key_hash
-- - O(1) validation performance
-- - No project ID embedded in key (stored in database)
--
-- Technical Note:
-- We use SHA-256 instead of bcrypt for API keys because:
-- - Bcrypt is non-deterministic (same input = different hashes each time)
-- - Bcrypt is designed for passwords (user authentication)
-- - SHA-256 is deterministic (same input = same hash), enabling O(1) lookup
-- - SHA-256 is industry standard for API keys (GitHub, Stripe, OpenAI)

-- Add unique index on key_hash for O(1) direct hash lookup
-- This enables industry-standard validation: sha256(incoming_key) -> lookup -> validate
DROP INDEX IF EXISTS idx_api_keys_key_hash;
CREATE UNIQUE INDEX idx_api_keys_key_hash ON api_keys(key_hash);

-- Add database comments documenting new format
COMMENT ON COLUMN api_keys.key_hash IS 'SHA-256 hash of full API key. Format: bk_{40_char_random}. Deterministic hashing enables O(1) validation.';
COMMENT ON COLUMN api_keys.key_preview IS 'Preview of API key for display. Format: bk_xxxx...yyyy (first 7 + last 4 chars). Follows GitHub pattern.';
COMMENT ON COLUMN api_keys.project_id IS 'Project this API key belongs to. Retrieved from database after hash validation.';

-- Add table comment
COMMENT ON TABLE api_keys IS 'API keys for SDK authentication. Uses industry-standard pure random format (bk_{random}) with SHA-256 hashing.';
