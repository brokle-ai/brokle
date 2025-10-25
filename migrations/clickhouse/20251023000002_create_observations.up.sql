-- OTEL observations table
CREATE TABLE IF NOT EXISTS observations (
    -- Identifiers
    id String,
    trace_id String,
    parent_observation_id Nullable(String),
    project_id String,

    -- Span data
    name String,
    span_kind LowCardinality(String),
    type LowCardinality(String),
    start_time DateTime64(3),
    end_time Nullable(DateTime64(3)),
    duration_ms Nullable(UInt32),

    -- Status
    status_code String,
    status_message Nullable(String),

    -- Data storage
    attributes String CODEC(ZSTD(1)),
    input Nullable(String) CODEC(ZSTD(3)),
    output Nullable(String) CODEC(ZSTD(3)),
    metadata Map(LowCardinality(String), String),
    level LowCardinality(String) DEFAULT 'DEFAULT',

    -- Universal model fields
    model_name Nullable(String),
    provider String DEFAULT '',
    internal_model_id Nullable(String),
    model_parameters Nullable(String),

    -- Usage & Cost
    provided_usage_details Map(LowCardinality(String), UInt64),
    usage_details Map(LowCardinality(String), UInt64),
    provided_cost_details Map(LowCardinality(String), Decimal(18, 12)),
    cost_details Map(LowCardinality(String), Decimal(18, 12)),
    total_cost Nullable(Decimal(18, 12)),

    -- Prompt management
    prompt_id Nullable(String),
    prompt_name Nullable(String),
    prompt_version Nullable(UInt16),

    -- System fields
    created_at DateTime64(3) DEFAULT now64(),
    updated_at DateTime64(3) DEFAULT now64(),
    version UInt32,
    event_ts DateTime64(3),
    is_deleted UInt8 DEFAULT 0,

    -- Indexes
    INDEX idx_id id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_trace_id trace_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_parent_observation_id parent_observation_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_project_id project_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_type type TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_provider provider TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_model_name model_name TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_prompt_id prompt_id TYPE bloom_filter(0.001) GRANULARITY 1

) ENGINE = ReplacingMergeTree(event_ts, is_deleted)
PARTITION BY toYYYYMM(start_time)
ORDER BY (project_id, type, toDate(start_time), id)
TTL toDateTime(start_time) + INTERVAL 365 DAY
SETTINGS index_granularity = 8192;
