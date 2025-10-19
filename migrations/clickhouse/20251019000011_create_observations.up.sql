-- ClickHouse Migration: create_observations
-- Created: 2025-10-19
-- Description: Observations table with hierarchical support, typed spans, and cost tracking

CREATE TABLE IF NOT EXISTS observations (
    -- Identifiers
    id String,
    trace_id String,
    parent_observation_id Nullable(String),
    project_id String,

    -- Observation metadata
    type LowCardinality(String), -- SPAN, EVENT, GENERATION, LLM, TOOL, AGENT, CHAIN, EMBEDDING, RETRIEVAL
    name String,
    start_time DateTime64(3),
    end_time Nullable(DateTime64(3)),

    -- Model information
    model Nullable(String),
    model_parameters Map(LowCardinality(String), String),

    -- Data (compressed for storage efficiency)
    input Nullable(String) CODEC(ZSTD(3)),
    output Nullable(String) CODEC(ZSTD(3)),
    metadata Map(LowCardinality(String), String),

    -- Cost tracking (provided by user OR calculated by system)
    -- Keys: input, output, total (all in USD)
    cost_details Map(LowCardinality(String), Decimal64(12)),

    -- Token usage tracking
    -- Keys: prompt_tokens, completion_tokens, total_tokens
    usage_details Map(LowCardinality(String), UInt64),

    -- Status and logging
    level LowCardinality(String) DEFAULT 'DEFAULT', -- DEBUG, INFO, WARN, ERROR, DEFAULT
    status_message Nullable(String),

    -- Completion tracking for streaming responses
    completion_start_time Nullable(DateTime64(3)),
    time_to_first_token_ms Nullable(UInt32),

    -- ReplacingMergeTree fields for update support
    version UInt32,
    event_ts DateTime64(3),
    is_deleted UInt8 DEFAULT 0,

    -- Bloom filter indexes for fast lookups
    INDEX idx_id id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_trace_id trace_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_parent_id parent_observation_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_model model TYPE bloom_filter(0.01) GRANULARITY 1

) ENGINE = ReplacingMergeTree(event_ts, is_deleted)
PARTITION BY toYYYYMM(start_time)
ORDER BY (project_id, type, toDate(start_time), trace_id, id)
TTL toDateTime(start_time) + INTERVAL 365 DAY
SETTINGS index_granularity = 8192;

