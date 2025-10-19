-- ClickHouse Migration: create_scores
-- Created: 2025-10-19
-- Description: Quality scores table with multi-level support (trace, observation, session)

CREATE TABLE IF NOT EXISTS scores (
    -- Identifiers (at least one of trace_id, observation_id, or session_id must be set)
    id String,
    project_id String,
    trace_id Nullable(String),
    observation_id Nullable(String),
    session_id Nullable(String),

    -- Score data
    name String,
    value Nullable(Float64),
    string_value Nullable(String),
    data_type LowCardinality(String), -- NUMERIC, CATEGORICAL, BOOLEAN

    -- Source and metadata
    source LowCardinality(String), -- API, AUTO, HUMAN, EVAL
    comment Nullable(String) CODEC(ZSTD(1)),

    -- Evaluator information
    evaluator_name Nullable(String),
    evaluator_version Nullable(String),
    evaluator_config Map(LowCardinality(String), String),

    -- Author tracking (for HUMAN source)
    author_user_id Nullable(String),

    -- Timestamp
    timestamp DateTime64(3) DEFAULT now64(),

    -- ReplacingMergeTree fields for update support
    version UInt32,
    event_ts DateTime64(3),
    is_deleted UInt8 DEFAULT 0,

    -- Bloom filter indexes for fast lookups
    INDEX idx_id id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_trace_id trace_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_observation_id observation_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_session_id session_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_name name TYPE bloom_filter(0.01) GRANULARITY 1

) ENGINE = ReplacingMergeTree(event_ts, is_deleted)
PARTITION BY toYYYYMM(timestamp)
ORDER BY (project_id, toDate(timestamp), name, id)
TTL toDateTime(timestamp) + INTERVAL 365 DAY
SETTINGS index_granularity = 8192;

