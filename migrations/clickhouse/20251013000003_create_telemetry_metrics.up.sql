-- Create telemetry_metrics table for SDK observability metrics tracking
CREATE TABLE IF NOT EXISTS telemetry_metrics
(
    project_id      String COMMENT 'Project identifier',
    environment     String DEFAULT 'default' COMMENT 'Environment tag (production, staging, etc.)',
    metric_name     LowCardinality(String) COMMENT 'Metric name (e.g., batch_processing_time)',
    metric_type     LowCardinality(String) COMMENT 'Metric type (counter, gauge, histogram)',
    metric_value    Float64 COMMENT 'Metric value',
    labels          String COMMENT 'JSON labels for metric dimensions',
    metadata        String COMMENT 'JSON metadata for additional context',
    timestamp       DateTime64(3) COMMENT 'Metric timestamp',
    processed_at    DateTime64(3) COMMENT 'Processing timestamp',
    INDEX idx_project_id project_id TYPE bloom_filter() GRANULARITY 4,
    INDEX idx_environment environment TYPE bloom_filter() GRANULARITY 4,
    INDEX idx_metric_name metric_name TYPE bloom_filter() GRANULARITY 4,
    INDEX idx_metric_type metric_type TYPE bloom_filter() GRANULARITY 4,
    INDEX idx_timestamp timestamp TYPE minmax GRANULARITY 1
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (project_id, metric_name, timestamp)
TTL toDateTime(timestamp) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192
COMMENT 'Telemetry metrics for SDK observability performance tracking with 90-day retention';
