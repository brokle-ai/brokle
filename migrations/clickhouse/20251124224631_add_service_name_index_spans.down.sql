-- Remove bloom filter index on service_name
ALTER TABLE spans DROP INDEX IF EXISTS idx_service_name
