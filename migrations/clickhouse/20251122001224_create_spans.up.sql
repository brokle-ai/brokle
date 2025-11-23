CREATE TABLE IF NOT EXISTS spans (
    -- ============================================
    -- OTEL Core Identity
    -- ============================================
    span_id String CODEC(ZSTD(1)),
    trace_id String CODEC(ZSTD(1)),
    parent_span_id Nullable(String) CODEC(ZSTD(1)),
    trace_state Nullable(String) CODEC(ZSTD(1)),

    -- ============================================
    -- Multi-Tenancy
    -- ============================================
    project_id String CODEC(ZSTD(1)),

    -- ============================================
    -- Span Metadata
    -- ============================================
    span_name String CODEC(ZSTD(1)),
    span_kind UInt8 CODEC(ZSTD(1)),

    -- ============================================
    -- Timing
    -- ============================================
    start_time DateTime64(9) CODEC(Delta(8), ZSTD(1)),
    end_time Nullable(DateTime64(9)) CODEC(Delta(8), ZSTD(1)),
    duration_ms Nullable(UInt32) CODEC(ZSTD(1)),

    completion_start_time Nullable(DateTime64(9)) CODEC(Delta(8), ZSTD(1)),

    -- ============================================
    -- Status
    -- ============================================
    status_code UInt8 CODEC(ZSTD(1)),
    status_message Nullable(String) CODEC(ZSTD(1)),

    -- ============================================
    -- I/O (Separate - Too Large for JSON)
    -- ============================================
    input Nullable(String) CODEC(ZSTD(3)),
    output Nullable(String) CODEC(ZSTD(3)),

    -- ============================================
    -- MODERN: JSON Attributes (9-10x Faster than Map)
    -- ============================================
    attributes JSON,

    metadata JSON,

    -- ============================================
    -- FLEXIBLE: Usage/Cost Maps
    -- ============================================
    usage_details Map(LowCardinality(String), UInt64) CODEC(ZSTD(1)),

    cost_details Map(LowCardinality(String), Decimal(18,12)) CODEC(ZSTD(1)),

    pricing_snapshot Map(LowCardinality(String), Decimal(18,12)) CODEC(ZSTD(1)),

    total_cost Nullable(Decimal(18,12)) CODEC(ZSTD(1)),

    -- ============================================
    -- Materialized: Only for Filters (Need Indexes)
    -- ============================================
    model_name LowCardinality(String) MATERIALIZED
        attributes.gen_ai.request.model CODEC(ZSTD(1)),

    provider_name LowCardinality(String) MATERIALIZED
        attributes.gen_ai.provider.name CODEC(ZSTD(1)),

    span_type LowCardinality(String) MATERIALIZED
        attributes.brokle.span.type CODEC(ZSTD(1)),

    -- ============================================
    -- OTEL Events/Links
    -- ============================================
    events_timestamp Array(DateTime64(9)) CODEC(ZSTD(1)),
    events_name Array(LowCardinality(String)) CODEC(ZSTD(1)),
    events_attributes Array(Map(LowCardinality(String), String)) CODEC(ZSTD(1)),

    links_trace_id Array(String) CODEC(ZSTD(1)),
    links_span_id Array(String) CODEC(ZSTD(1)),
    links_trace_state Array(String) CODEC(ZSTD(1)),
    links_attributes Array(Map(LowCardinality(String), String)) CODEC(ZSTD(1)),

    -- ============================================
    -- A/B Testing & Versioning (Materialized from attributes)
    -- ============================================
    version LowCardinality(String) MATERIALIZED JSONExtractString(attributes, 'brokle.span.version') CODEC(ZSTD(1)),

    -- ============================================
    -- Span Importance Level (Materialized from attributes for filtering/sorting)
    -- ============================================
    level LowCardinality(String) MATERIALIZED JSONExtractString(attributes, 'brokle.span.level') CODEC(ZSTD(1)),

    -- ============================================
    -- System Timestamps
    -- ============================================
    created_at DateTime64(3) DEFAULT now64(),
    updated_at DateTime64(3) DEFAULT now64(),

    deleted_at Nullable(DateTime64(3)) CODEC(ZSTD(1)),

    -- ============================================
    -- Indexes (Only on Materialized Columns)
    -- ============================================
    INDEX idx_span_id span_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_trace_id trace_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_project_id project_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_model model_name TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_provider provider_name TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_span_type span_type TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_span_version version TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_level level TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_total_cost total_cost TYPE minmax GRANULARITY 1
)
ENGINE = MergeTree()
PARTITION BY (toYYYYMM(start_time), project_id)
ORDER BY (project_id, start_time, span_id)
TTL start_time + INTERVAL 365 DAY
SETTINGS index_granularity = 8192;
