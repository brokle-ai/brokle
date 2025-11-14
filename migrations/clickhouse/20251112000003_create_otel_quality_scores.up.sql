-- ============================================================================
-- OTEL-Native: Quality Scores Table
-- ============================================================================
-- Purpose: Post-hoc evaluation scores for LLM quality monitoring
-- Design: Separate table (Langfuse pattern - not OTEL Events)
-- Rationale: Post-hoc evaluations, multiple evaluators, longer retention (90d)
-- Engine: MergeTree (immutable scores)
-- ============================================================================

CREATE TABLE IF NOT EXISTS quality_scores (
    -- ============================================
    -- IDENTITY
    -- ============================================
    score_id String CODEC(ZSTD(1)),
        -- Unique score identifier (ULID or UUID)

    -- ============================================
    -- LINKS (Denormalized for performance)
    -- ============================================
    project_id String CODEC(ZSTD(1)),
    trace_id String CODEC(ZSTD(1)),
    span_id String CODEC(ZSTD(1)),
        -- Links to specific span (or trace-level if NULL in future)

    -- ============================================
    -- SCORE DATA
    -- ============================================
    name String CODEC(ZSTD(1)),
        -- Score dimension
        -- Example: "relevance", "correctness", "hallucination", "coherence"

    value Nullable(Float64) CODEC(ZSTD(1)),
        -- Numeric score (0.0 - 1.0 normalized)

    string_value Nullable(String) CODEC(ZSTD(1)),
        -- Categorical score
        -- Example: "excellent", "good", "fair", "poor"

    data_type String CODEC(ZSTD(1)),
        -- Score type: NUMERIC, CATEGORICAL, BOOLEAN

    -- ============================================
    -- METADATA
    -- ============================================
    source String CODEC(ZSTD(1)),
        -- Score source: API, ANNOTATION, EVAL

    comment Nullable(String) CODEC(ZSTD(1)),
        -- Evaluator notes or reasoning

    -- ============================================
    -- EVALUATOR INFORMATION
    -- ============================================
    evaluator_name Nullable(String) CODEC(ZSTD(1)),
        -- Evaluator identifier
        -- Example: "gpt-4o", "human-reviewer-alice", "regex-validator"

    evaluator_version Nullable(String) CODEC(ZSTD(1)),
        -- Evaluator version for reproducibility

    evaluator_config JSON CODEC(ZSTD(1)),
        -- Evaluator configuration (JSON for flexibility)
        -- Example: {"threshold": 0.7, "criteria": "strict"}

    author_user_id Nullable(String) CODEC(ZSTD(1)),
        -- User who created score (for human annotations)

    -- ============================================
    -- TIMESTAMP
    -- ============================================
    timestamp DateTime64(3) DEFAULT now64(),
        -- When score was created (may be hours after span)

    -- ============================================
    -- INDEXES
    -- ============================================
    INDEX idx_trace_id trace_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_span_id span_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_project_id project_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_score_name name TYPE bloom_filter(0.01) GRANULARITY 1
)
ENGINE = MergeTree
    -- Immutable scores (no updates)

PARTITION BY (toYYYYMM(timestamp), project_id)
    -- Monthly partitions + project isolation

ORDER BY (project_id, timestamp, score_id)
    -- project_id FIRST: Multi-tenant isolation
    -- timestamp: Time-series ordering (no toDate for better compression)
    -- score_id: Unique identifier

TTL timestamp + INTERVAL 365 DAY
    -- Match spans TTL for consistency

SETTINGS index_granularity = 8192;
