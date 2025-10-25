-- Quality scores table
CREATE TABLE IF NOT EXISTS scores (
    -- Identifiers
    id String,
    project_id String,
    trace_id String,
    observation_id String,

    -- Score data
    name String,
    value Nullable(Float64),
    string_value Nullable(String),
    data_type String,

    -- Metadata
    source String,
    comment Nullable(String) CODEC(ZSTD(1)),

    -- Evaluator
    evaluator_name Nullable(String),
    evaluator_version Nullable(String),
    evaluator_config Map(LowCardinality(String), String),
    author_user_id Nullable(String),

    -- Timestamps
    timestamp DateTime64(3) DEFAULT now64(),
    version UInt32,
    event_ts DateTime64(3),
    is_deleted UInt8 DEFAULT 0,

    -- Indexes
    INDEX idx_trace_id trace_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_observation_id observation_id TYPE bloom_filter(0.001) GRANULARITY 1

) ENGINE = ReplacingMergeTree(event_ts, is_deleted)
PARTITION BY toYYYYMM(timestamp)
ORDER BY (project_id, toDate(timestamp), name, id)
TTL toDateTime(timestamp) + INTERVAL 365 DAY;
