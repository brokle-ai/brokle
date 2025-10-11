-- ClickHouse Migration: create_llm_quality_analytics
-- Created: 2025-10-11T23:36:00+05:30

-- Create llm_quality_analytics table for quality score analytics
CREATE TABLE IF NOT EXISTS llm_quality_analytics (
    timestamp DateTime64(3) DEFAULT now64(),
    trace_id String,
    observation_id String,
    organization_id String,
    project_id String,
    score_name LowCardinality(String),
    score_value Float64,
    string_value String DEFAULT '',
    data_type LowCardinality(String), -- 'NUMERIC', 'CATEGORICAL', 'BOOLEAN'
    score_source LowCardinality(String), -- 'auto', 'human', 'eval', 'api'
    evaluator_name String DEFAULT '',
    evaluator_version String DEFAULT '',
    provider String DEFAULT '',
    model String DEFAULT '',
    observation_type String DEFAULT '',
    created_date Date DEFAULT toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, score_name, organization_id, project_id)
TTL toDateTime(timestamp) + INTERVAL 180 DAY
SETTINGS index_granularity = 8192;
