-- OTEL traces table (root spans only)
CREATE TABLE IF NOT EXISTS traces (
    -- Identifiers
    id String,
    project_id String,

    -- Metadata
    name String,
    user_id Nullable(String),
    session_id Nullable(String),

    -- Timing
    start_time DateTime64(3),
    end_time Nullable(DateTime64(3)),
    duration_ms Nullable(UInt32),

    -- Status
    status_code String,
    status_message Nullable(String),

    -- Data
    attributes String CODEC(ZSTD(1)),
    input Nullable(String) CODEC(ZSTD(3)),
    output Nullable(String) CODEC(ZSTD(3)),
    metadata Map(LowCardinality(String), String),
    tags Array(String),

    -- Resource attributes
    environment LowCardinality(String) DEFAULT 'default',
    service_name Nullable(String),
    service_version Nullable(String),
    release Nullable(String),

    -- Aggregate metrics
    total_cost Nullable(Decimal64(12)),
    total_tokens Nullable(UInt32),
    observation_count Nullable(UInt32),

    -- Flags
    bookmarked Bool DEFAULT false,
    public Bool DEFAULT false,

    -- Timestamps
    created_at DateTime64(3) DEFAULT now64(),
    updated_at DateTime64(3) DEFAULT now64(),

    -- ReplacingMergeTree
    version UInt32,
    event_ts DateTime64(3),
    is_deleted UInt8 DEFAULT 0,

    -- Indexes
    INDEX idx_id id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_project_id project_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_user_id user_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_session_id session_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_tags tags TYPE bloom_filter(0.01) GRANULARITY 1

) ENGINE = ReplacingMergeTree(event_ts, is_deleted)
PARTITION BY toYYYYMM(start_time)
ORDER BY (project_id, toDate(start_time), id)
TTL toDateTime(start_time) + INTERVAL 365 DAY
SETTINGS index_granularity = 8192;
