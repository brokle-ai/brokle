-- Rollback: clean_models_pricing_langfuse_pattern
-- Created: 2025-11-22T00:10:40+05:30

-- Drop new tables
DROP TABLE IF EXISTS provider_prices CASCADE;
DROP TABLE IF EXISTS provider_models CASCADE;
