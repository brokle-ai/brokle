-- ClickHouse Migration: create_sessions
-- Created: 2025-10-19
-- Description: Sessions table for grouping traces into user journeys

CREATE TABLE IF NOT EXISTS sessions (
    -- Identifiers
    id String,
    project_id String,

    -- Session metadata
    user_id Nullable(String),
    metadata Map(LowCardinality(String), String),

    -- Feature flags
    bookmarked UInt8 DEFAULT 0,
    public UInt8 DEFAULT 0,

    -- Timestamps
    created_at DateTime64(3) DEFAULT now64(),

    -- ReplacingMergeTree fields for update support
    version UInt32,
    event_ts DateTime64(3),
    is_deleted UInt8 DEFAULT 0,

    -- Bloom filter indexes for fast lookups
    INDEX idx_id id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_user_id user_id TYPE bloom_filter(0.001) GRANULARITY 1

) ENGINE = ReplacingMergeTree(event_ts, is_deleted)
PARTITION BY toYYYYMM(created_at)
ORDER BY (project_id, toDate(created_at), id)
TTL toDateTime(created_at) + INTERVAL 365 DAY
SETTINGS index_granularity = 8192;

