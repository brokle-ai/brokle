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
TTL toDateTime(timestamp) + INTERVAL 365 DAY;

