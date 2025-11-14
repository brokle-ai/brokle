-- ============================================================================
-- OTEL 1.38+ Native: Traces Table
-- ============================================================================
-- Purpose: Trace-level context for LLM workflows (stored ONCE, not per span)
-- Design: Langfuse pattern - prevents 98% context duplication
-- OTEL Compliance: Uses trace_id, standard status codes
-- Engine: ReplacingMergeTree (supports aggregation updates)
-- ============================================================================

CREATE TABLE IF NOT EXISTS traces (
    -- ============================================
    -- OTEL CORE IDENTITY
    -- ============================================
    trace_id String CODEC(ZSTD(1)),
        -- OTEL: Trace identifier (32 hex chars lowercase)
        -- Example: "4bf92f3577b34da6a3ce929d0e0e4736"

    -- ============================================
    -- MULTI-TENANCY (CRITICAL - First in ORDER BY)
    -- ============================================
    project_id String CODEC(ZSTD(1)),
        -- Brokle: Project/tenant identifier for data isolation

    -- ============================================
    -- TRACE METADATA
    -- ============================================
    name String CODEC(ZSTD(1)),
        -- Trace name / workflow description

    -- ============================================
    -- USER CONTEXT (Stored ONCE - not duplicated in spans)
    -- ============================================
    user_id Nullable(String) CODEC(ZSTD(1)),
        -- OTEL semantic: enduser.id
        -- All spans in trace inherit this context

    session_id Nullable(String) CODEC(ZSTD(1)),
        -- OTEL semantic: session.id
        -- Multi-turn conversation identifier

    -- ============================================
    -- A/B TESTING (Stored ONCE)
    -- ============================================
    version Nullable(String) CODEC(ZSTD(1)),
        -- Brokle: Application/experiment version
        -- Example: "v2-agents", "experiment-fast-mode"
        -- NOT in spans table (would duplicate 50x)

    -- ============================================
    -- CATEGORIZATION (Stored ONCE)
    -- ============================================
    tags Array(String) CODEC(ZSTD(1)),
        -- Brokle: Custom tags
        -- Example: ["production", "customer-support", "high-priority"]

    environment LowCardinality(String) DEFAULT 'default' CODEC(ZSTD(1)),
        -- OTEL semantic: deployment.environment.name
        -- Values: production, staging, development, test

    -- ============================================
    -- RESOURCE ATTRIBUTES (Service Context)
    -- ============================================
    resource_attributes JSON CODEC(ZSTD(1)),
        -- OTEL: Resource-level attributes
        -- Contains: service.*, deployment.*, host.*, cloud.*
        -- Example: {"service.name": "brokle-api", "service.version": "1.2.3"}

    -- Denormalized for fast filtering (extracted from resource_attributes)
    service_name Nullable(String) CODEC(ZSTD(1)),
    service_version Nullable(String) CODEC(ZSTD(1)),
    release Nullable(String) CODEC(ZSTD(1)),

    -- ============================================
    -- TIMING
    -- ============================================
    start_time DateTime64(3) CODEC(Delta(8), ZSTD(1)),
    end_time Nullable(DateTime64(3)) CODEC(Delta(8), ZSTD(1)),
    duration_ms Nullable(UInt32) CODEC(ZSTD(1)),

    -- ============================================
    -- OTEL STATUS
    -- ============================================
    status_code UInt8 CODEC(ZSTD(1)),
        -- OTEL: Status code enum (matches protobuf)
        -- 0 = STATUS_CODE_UNSET
        -- 1 = STATUS_CODE_OK
        -- 2 = STATUS_CODE_ERROR
        -- Set to 2 (ERROR) if ANY span in trace has error

    status_message Nullable(String) CODEC(ZSTD(1)),

    -- ============================================
    -- TRACE-LEVEL I/O (Full conversation context)
    -- ============================================
    input Nullable(String) CODEC(ZSTD(3)),
        -- High compression for large prompts

    output Nullable(String) CODEC(ZSTD(3)),
        -- High compression for large completions

    -- ============================================
    -- AGGREGATIONS (Calculated from spans)
    -- ============================================
    total_cost Nullable(Decimal(18, 9)) CODEC(ZSTD(1)),
        -- Sum of all span costs (9 decimals sufficient)

    total_tokens Nullable(UInt32) CODEC(ZSTD(1)),
        -- Sum of input + output tokens

    span_count Nullable(UInt32) CODEC(ZSTD(1)),
        -- Total spans in trace

    -- ============================================
    -- BROKLE FEATURES
    -- ============================================
    bookmarked Bool DEFAULT false,
    public Bool DEFAULT false,

    -- ============================================
    -- SYSTEM TIMESTAMPS
    -- ============================================
    created_at DateTime64(3) DEFAULT now64(),
    updated_at DateTime64(3) DEFAULT now64(),

    -- ============================================
    -- INDEXES
    -- ============================================
    INDEX idx_trace_id trace_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_project_id project_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_user_id user_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_session_id session_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_tags tags TYPE bloom_filter(0.01) GRANULARITY 1
)
ENGINE = ReplacingMergeTree(updated_at)
    -- Supports updates to aggregations as spans arrive

PARTITION BY (toYYYYMM(start_time), project_id)
    -- Monthly partitions + project isolation

ORDER BY (project_id, start_time, trace_id)
    -- project_id FIRST: Multi-tenant data locality
    -- start_time: Time-series optimization (no toDate for better compression)
    -- trace_id: Unique identifier

TTL start_time + INTERVAL 365 DAY
    -- Automatic cleanup after 1 year

SETTINGS index_granularity = 8192;
