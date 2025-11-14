-- ClickHouse Rollback: fix_token_materialized_columns
-- Created: 2025-11-14T20:17:09+05:30

-- Rollback: restore original (broken) columns with toInt32OrNull
ALTER TABLE spans
    DROP COLUMN gen_ai_usage_input_tokens,
    DROP COLUMN gen_ai_usage_output_tokens,
    ADD COLUMN gen_ai_usage_input_tokens Nullable(Int32) MATERIALIZED
        toInt32OrNull(span_attributes.`gen_ai.usage.input_tokens`) CODEC(ZSTD(1)),
    ADD COLUMN gen_ai_usage_output_tokens Nullable(Int32) MATERIALIZED
        toInt32OrNull(span_attributes.`gen_ai.usage.output_tokens`) CODEC(ZSTD(1));
