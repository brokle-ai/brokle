-- ClickHouse Migration: create_traces
-- Created: 2025-10-19
-- Description: Core traces table with ReplacingMergeTree for updates and hierarchical support

CREATE TABLE IF NOT EXISTS traces (
    -- Identifiers
    id String,
    project_id String,
    session_id Nullable(String),
    parent_trace_id Nullable(String),

    -- Basic information
    name String,
    user_id Nullable(String),
    timestamp DateTime64(3) DEFAULT now64(),

    -- Data (compressed for storage efficiency)
    input Nullable(String) CODEC(ZSTD(3)),
    output Nullable(String) CODEC(ZSTD(3)),
    metadata Map(LowCardinality(String), String),
    tags Array(String),

    -- Environment and versioning
    environment LowCardinality(String) DEFAULT 'production',
    release Nullable(String),

    -- ReplacingMergeTree fields for update support
    version UInt32,
    event_ts DateTime64(3),
    is_deleted UInt8 DEFAULT 0,

    -- Bloom filter indexes for fast lookups
    INDEX idx_id id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_session_id session_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_user_id user_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_metadata_keys mapKeys(metadata) TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_tags tags TYPE bloom_filter(0.01) GRANULARITY 1

) ENGINE = ReplacingMergeTree(event_ts, is_deleted)
PARTITION BY toYYYYMM(timestamp)
ORDER BY (project_id, toDate(timestamp), id)
TTL toDateTime(timestamp) + INTERVAL 365 DAY
SETTINGS index_granularity = 8192
