-- ClickHouse Migration: fix_token_materialized_columns
-- Created: 2025-11-14T20:17:09+05:30

-- Fix materialized columns: Use direct JSON access (no conversion functions)
-- Root cause: span_attributes is JSON type, toInt32OrNull() expects String
-- Proper fix: Direct access like other columns (gen_ai_response_model pattern)

-- Drop and re-add columns with direct JSON access
ALTER TABLE spans
    DROP COLUMN gen_ai_usage_input_tokens,
    DROP COLUMN gen_ai_usage_output_tokens,
    ADD COLUMN gen_ai_usage_input_tokens Nullable(Int32) MATERIALIZED
        span_attributes.`gen_ai.usage.input_tokens` CODEC(ZSTD(1)),
    ADD COLUMN gen_ai_usage_output_tokens Nullable(Int32) MATERIALIZED
        span_attributes.`gen_ai.usage.output_tokens` CODEC(ZSTD(1));
