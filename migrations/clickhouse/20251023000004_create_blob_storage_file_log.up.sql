-- Blob storage file log (S3 references for large payloads)
CREATE TABLE IF NOT EXISTS blob_storage_file_log (
    -- Identifiers
    id String,
    project_id String,

    -- Entity reference
    entity_type LowCardinality(String),
    entity_id String,
    event_id String,

    -- Storage location
    bucket_name String,
    bucket_path String,

    -- Metadata
    file_size_bytes Nullable(UInt64),
    content_type Nullable(String) DEFAULT 'text/plain',
    compression Nullable(String),

    -- Timestamps
    created_at DateTime64(3) DEFAULT now64(),
    updated_at DateTime64(3) DEFAULT now64(),

    -- ReplacingMergeTree
    version UInt32,
    event_ts DateTime64(3),
    is_deleted UInt8 DEFAULT 0,

    -- Indexes
    INDEX idx_entity_id entity_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_project_id project_id TYPE bloom_filter(0.001) GRANULARITY 1

) ENGINE = ReplacingMergeTree(version)
PARTITION BY toYYYYMM(created_at)
ORDER BY (project_id, entity_type, entity_id, created_at)
TTL toDateTime(created_at) + INTERVAL 365 DAY
SETTINGS index_granularity = 8192;
