CREATE TABLE IF NOT EXISTS traces (
    -- ============================================
    -- OTEL Core Identity
    -- ============================================
    trace_id String CODEC(ZSTD(1)),

    -- ============================================
    -- Multi-Tenancy
    -- ============================================
    project_id String CODEC(ZSTD(1)),

    -- ============================================
    -- Trace Metadata
    -- ============================================
    name String CODEC(ZSTD(1)),
    user_id Nullable(String) CODEC(ZSTD(1)),
    session_id Nullable(String) CODEC(ZSTD(1)),

    tags Array(String) CODEC(ZSTD(1)),
    environment LowCardinality(String) DEFAULT 'default' CODEC(ZSTD(1)),

    -- ============================================
    -- MODERN: JSON Metadata
    -- ============================================
    metadata JSON,

    -- ============================================
    -- Versioning & Release (Materialized from metadata)
    -- ============================================
    release LowCardinality(String) MATERIALIZED JSONExtractString(metadata, 'brokle.release') CODEC(ZSTD(1)),
    version LowCardinality(String) MATERIALIZED JSONExtractString(metadata, 'brokle.version') CODEC(ZSTD(1)),

    -- ============================================
    -- Timing
    -- ============================================
    start_time DateTime64(9) CODEC(Delta(8), ZSTD(1)),
    end_time Nullable(DateTime64(9)) CODEC(Delta(8), ZSTD(1)),
    duration_ms Nullable(UInt32) CODEC(ZSTD(1)),

    -- ============================================
    -- Status
    -- ============================================
    status_code UInt8 CODEC(ZSTD(1)),
    status_message Nullable(String) CODEC(ZSTD(1)),

    -- ============================================
    -- I/O
    -- ============================================
    input Nullable(String) CODEC(ZSTD(3)),
    output Nullable(String) CODEC(ZSTD(3)),

    -- ============================================
    -- Aggregations (Pre-computed from Spans)
    -- ============================================
    total_cost Nullable(Decimal(18,12)) CODEC(ZSTD(1)),
    total_tokens Nullable(UInt32) CODEC(ZSTD(1)),
    span_count Nullable(UInt32) CODEC(ZSTD(1)),

    -- ============================================
    -- Features
    -- ============================================
    bookmarked Bool DEFAULT false,
    public Bool DEFAULT false,

    -- ============================================
    -- System Timestamps
    -- ============================================
    created_at DateTime64(3) DEFAULT now64(),
    updated_at DateTime64(3) DEFAULT now64(),

    deleted_at Nullable(DateTime64(3)) CODEC(ZSTD(1)),

    -- ============================================
    -- Indexes
    -- ============================================
    INDEX idx_trace_id trace_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_project_id project_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_user_id user_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_session_id session_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_tags tags TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_release release TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_version version TYPE bloom_filter(0.01) GRANULARITY 1
)
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY (toYYYYMM(start_time), project_id)
ORDER BY (project_id, start_time, trace_id)
TTL start_time + INTERVAL 365 DAY
SETTINGS index_granularity = 8192;
