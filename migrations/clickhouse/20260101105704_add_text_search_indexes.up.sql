-- ClickHouse Migration: add_text_search_indexes
-- Created: 2026-01-01T10:57:04+05:30
-- Purpose: Add full-text search capability on input/output fields

-- Add truncated preview columns for efficient text search
-- These are MATERIALIZED columns that automatically extract the first 10000 characters
-- for faster text search operations without scanning full content

ALTER TABLE otel_traces
    ADD COLUMN input_preview String
    MATERIALIZED substring(coalesce(input, ''), 1, 10000)
    CODEC(ZSTD(1));

ALTER TABLE otel_traces
    ADD COLUMN output_preview String
    MATERIALIZED substring(coalesce(output, ''), 1, 10000)
    CODEC(ZSTD(1));

-- Add token bloom filter indexes for efficient full-text search
-- tokenbf_v1 tokenizes text and uses bloom filters for fast substring matching
-- Parameters: (filter_size=10240, hash_functions=3, seed=0)
-- - filter_size: Number of bits in bloom filter (larger = fewer false positives)
-- - hash_functions: Number of hash functions (3 is optimal for most cases)
-- - seed: Random seed for hash functions (0 for deterministic)

ALTER TABLE otel_traces
    ADD INDEX idx_input_tokens input_preview
    TYPE tokenbf_v1(10240, 3, 0)
    GRANULARITY 1;

ALTER TABLE otel_traces
    ADD INDEX idx_output_tokens output_preview
    TYPE tokenbf_v1(10240, 3, 0)
    GRANULARITY 1;

-- Materialize the new columns for existing data
-- This may take time for large tables but ensures search works on historical data
ALTER TABLE otel_traces MATERIALIZE COLUMN input_preview;
ALTER TABLE otel_traces MATERIALIZE COLUMN output_preview;

-- Force index materialization on existing data
ALTER TABLE otel_traces MATERIALIZE INDEX idx_input_tokens;
ALTER TABLE otel_traces MATERIALIZE INDEX idx_output_tokens;
