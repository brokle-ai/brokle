-- OTEL observations table (all spans)
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

    -- Data
    attributes String CODEC(ZSTD(1)),
    input Nullable(String) CODEC(ZSTD(3)),
    output Nullable(String) CODEC(ZSTD(3)),
    metadata Map(LowCardinality(String), String),
    level LowCardinality(String) DEFAULT 'DEFAULT',

    -- Gen AI semantic conventions
    gen_ai_system Nullable(String),
    gen_ai_model Nullable(String),
    gen_ai_prompt Nullable(String) CODEC(ZSTD(3)),
    gen_ai_completion Nullable(String) CODEC(ZSTD(3)),
    gen_ai_prompt_tokens Nullable(UInt32),
    gen_ai_completion_tokens Nullable(UInt32),
    gen_ai_total_tokens Nullable(UInt32),

    -- Brokle extensions
    brokle_routing_provider Nullable(String),
    brokle_routing_strategy Nullable(String),
    brokle_cost_total Nullable(Decimal64(12)),
    brokle_cost_input Nullable(Decimal64(12)),
    brokle_cost_output Nullable(Decimal64(12)),
    brokle_cache_hit Nullable(UInt8),
    brokle_cache_similarity Nullable(Float32),
    brokle_governance_passed Nullable(UInt8),
    brokle_governance_policy Nullable(String),

    -- Streaming metrics
    completion_start_time Nullable(DateTime64(3)),
    time_to_first_token_ms Nullable(UInt32),

    -- Prompt management
    prompt_id Nullable(String),
    prompt_name Nullable(String),
    prompt_version Nullable(UInt16),

    -- Blob storage references
    input_blob_storage_id Nullable(String),
    output_blob_storage_id Nullable(String),

    -- Usage and cost maps
    provided_usage_details Map(LowCardinality(String), UInt64),
    usage_details Map(LowCardinality(String), UInt64),
    provided_cost_details Map(LowCardinality(String), Decimal(18, 12)),
    cost_details Map(LowCardinality(String), Decimal(18, 12)),

    -- Model configuration
    model_parameters Nullable(String),
    internal_model_id Nullable(String),
    provider String DEFAULT '',

    -- Timestamps
    created_at DateTime64(3) DEFAULT now64(),
    updated_at DateTime64(3) DEFAULT now64(),

    -- ReplacingMergeTree
    version UInt32,
    event_ts DateTime64(3),
    is_deleted UInt8 DEFAULT 0,

    -- Indexes
    INDEX idx_id id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_trace_id trace_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_parent_observation_id parent_observation_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_project_id project_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_type type TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_gen_ai_model gen_ai_model TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_prompt_id prompt_id TYPE bloom_filter(0.001) GRANULARITY 1

) ENGINE = ReplacingMergeTree(version)
PARTITION BY toYYYYMM(start_time)
ORDER BY (project_id, trace_id, type, toDate(start_time), id)
TTL toDateTime(start_time) + INTERVAL 365 DAY
SETTINGS index_granularity = 8192;
