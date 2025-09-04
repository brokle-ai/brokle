-- Create metrics table for real-time platform metrics
CREATE TABLE IF NOT EXISTS metrics (
    timestamp DateTime64(3) DEFAULT now64(),
    metric_name LowCardinality(String),
    metric_value Float64,
    labels Map(String, String),
    organization_id String,
    project_id String,
    user_id String,
    provider String DEFAULT '',
    model String DEFAULT '',
    environment String DEFAULT 'production',
    created_date Date DEFAULT toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, metric_name, organization_id)
TTL timestamp + INTERVAL 90 DAY;

-- Create events table for business and system events
CREATE TABLE IF NOT EXISTS events (
    timestamp DateTime64(3) DEFAULT now64(),
    event_id String,
    event_type LowCardinality(String),
    event_data String,
    user_id String,
    organization_id String,
    project_id String,
    source String DEFAULT 'api',
    severity LowCardinality(String) DEFAULT 'info',
    created_date Date DEFAULT toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, event_type, organization_id)
TTL timestamp + INTERVAL 180 DAY;

-- Create traces table for distributed tracing
CREATE TABLE IF NOT EXISTS traces (
    timestamp DateTime64(3) DEFAULT now64(),
    trace_id String,
    span_id String,
    parent_span_id String DEFAULT '',
    operation_name String,
    duration_ms UInt32,
    status_code UInt16,
    tags Map(String, String),
    user_id String,
    organization_id String,
    created_date Date DEFAULT toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, trace_id, span_id)
TTL timestamp + INTERVAL 30 DAY;

-- Create request_logs table for API request logging
CREATE TABLE IF NOT EXISTS request_logs (
    timestamp DateTime64(3) DEFAULT now64(),
    request_id String,
    method LowCardinality(String),
    path String,
    status_code UInt16,
    duration_ms UInt32,
    request_size UInt32,
    response_size UInt32,
    user_id String,
    organization_id String,
    ip_address IPv4,
    user_agent String,
    created_date Date DEFAULT toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, status_code, organization_id)
TTL timestamp + INTERVAL 60 DAY;

-- Create ai_routing_metrics table for AI provider routing decisions
CREATE TABLE IF NOT EXISTS ai_routing_metrics (
    timestamp DateTime64(3) DEFAULT now64(),
    request_id String,
    provider String,
    model String,
    prompt_tokens UInt32,
    completion_tokens UInt32,
    total_tokens UInt32,
    cost_usd Float64,
    latency_ms UInt32,
    quality_score Float32 DEFAULT 0,
    routing_confidence Float32,
    user_id String,
    organization_id String,
    project_id String,
    created_date Date DEFAULT toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, provider, model, organization_id)
TTL timestamp + INTERVAL 365 DAY;