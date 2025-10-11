-- AI Gateway PostgreSQL Migration
-- Creates tables for provider management, model registry, configurations, and health tracking

-- Gateway Providers Table
-- Stores AI provider definitions (OpenAI, Anthropic, Cohere, etc.)
CREATE TABLE gateway_providers (
    id CHAR(26) PRIMARY KEY, -- ULID
    name VARCHAR(100) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL, -- 'openai', 'anthropic', 'cohere', 'google', etc.
    base_url TEXT NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    default_timeout_seconds INTEGER NOT NULL DEFAULT 30,
    max_retries INTEGER NOT NULL DEFAULT 3,
    health_check_url TEXT,
    supported_features JSONB NOT NULL DEFAULT '{}', -- {"streaming": true, "functions": true, "embeddings": true}
    rate_limits JSONB NOT NULL DEFAULT '{}', -- {"requests_per_minute": 1000, "tokens_per_minute": 150000}
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_gateway_providers_type ON gateway_providers(type);
CREATE INDEX idx_gateway_providers_enabled ON gateway_providers(is_enabled);

-- Gateway Models Table
-- Model registry with pricing, capabilities, and context limits
CREATE TABLE gateway_models (
    id CHAR(26) PRIMARY KEY, -- ULID
    provider_id CHAR(26) NOT NULL REFERENCES gateway_providers(id) ON DELETE CASCADE,
    model_name VARCHAR(100) NOT NULL, -- 'gpt-3.5-turbo', 'claude-3-sonnet', etc.
    display_name VARCHAR(200) NOT NULL,
    input_cost_per_1k_tokens DECIMAL(10,8) NOT NULL DEFAULT 0.0, -- Cost in USD
    output_cost_per_1k_tokens DECIMAL(10,8) NOT NULL DEFAULT 0.0,
    max_context_tokens INTEGER NOT NULL DEFAULT 4096,
    supports_streaming BOOLEAN NOT NULL DEFAULT false,
    supports_functions BOOLEAN NOT NULL DEFAULT false,
    supports_vision BOOLEAN NOT NULL DEFAULT false,
    model_type VARCHAR(50) NOT NULL DEFAULT 'text', -- 'text', 'embedding', 'image', 'audio'
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    quality_score DECIMAL(3,2) DEFAULT NULL, -- 0.0-1.0 quality rating
    speed_score DECIMAL(3,2) DEFAULT NULL, -- 0.0-1.0 speed rating
    metadata JSONB NOT NULL DEFAULT '{}', -- Additional model metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(provider_id, model_name)
);

-- Create indexes
CREATE INDEX idx_gateway_models_provider ON gateway_models(provider_id);
CREATE INDEX idx_gateway_models_name ON gateway_models(model_name);
CREATE INDEX idx_gateway_models_type ON gateway_models(model_type);
CREATE INDEX idx_gateway_models_enabled ON gateway_models(is_enabled);
CREATE INDEX idx_gateway_models_streaming ON gateway_models(supports_streaming);
CREATE INDEX idx_gateway_models_functions ON gateway_models(supports_functions);

-- Gateway Provider Configurations Table
-- Project-scoped provider API keys and settings
CREATE TABLE gateway_provider_configs (
    id CHAR(26) PRIMARY KEY, -- ULID
    project_id CHAR(26) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    provider_id CHAR(26) NOT NULL REFERENCES gateway_providers(id) ON DELETE CASCADE,
    api_key_encrypted TEXT NOT NULL, -- Encrypted API key
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    custom_base_url TEXT, -- Override provider base URL
    custom_timeout_seconds INTEGER, -- Override default timeout
    rate_limit_override JSONB, -- Override provider rate limits
    priority_order INTEGER NOT NULL DEFAULT 0, -- For fallback ordering (higher = higher priority)
    configuration JSONB NOT NULL DEFAULT '{}', -- Provider-specific config
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(project_id, provider_id)
);

-- Create indexes
CREATE INDEX idx_gateway_provider_configs_project ON gateway_provider_configs(project_id);
CREATE INDEX idx_gateway_provider_configs_provider ON gateway_provider_configs(provider_id);
CREATE INDEX idx_gateway_provider_configs_enabled ON gateway_provider_configs(is_enabled);
CREATE INDEX idx_gateway_provider_configs_priority ON gateway_provider_configs(priority_order DESC);

-- Provider Health Metrics Table
-- Tracks provider availability and performance
CREATE TABLE provider_health_metrics (
    id CHAR(26) PRIMARY KEY, -- ULID
    provider_id CHAR(26) NOT NULL REFERENCES gateway_providers(id) ON DELETE CASCADE,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL, -- 'healthy', 'degraded', 'unhealthy'
    avg_latency_ms INTEGER,
    success_rate DECIMAL(5,4), -- 0.0000 to 1.0000
    requests_per_minute INTEGER,
    errors_per_minute INTEGER,
    last_error TEXT,
    response_time_p95 INTEGER, -- 95th percentile response time in ms
    response_time_p99 INTEGER, -- 99th percentile response time in ms
    uptime_percentage DECIMAL(5,4), -- Last 24 hours uptime
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_provider_health_metrics_provider ON provider_health_metrics(provider_id);
CREATE INDEX idx_provider_health_metrics_timestamp ON provider_health_metrics(timestamp);
CREATE INDEX idx_provider_health_metrics_status ON provider_health_metrics(status);

-- Gateway Routing Rules Table
-- Project-specific routing configuration
CREATE TABLE gateway_routing_rules (
    id CHAR(26) PRIMARY KEY, -- ULID
    project_id CHAR(26) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    rule_name VARCHAR(100) NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 0, -- Higher number = higher priority
    conditions JSONB NOT NULL, -- {"model": "gpt-*", "user_tier": "premium", "time_range": "9-17"}
    routing_strategy VARCHAR(50) NOT NULL, -- 'cost_optimized', 'latency_optimized', 'quality_optimized', 'round_robin'
    target_providers JSONB NOT NULL, -- [{"provider_id": "...", "weight": 100}, {...}]
    fallback_providers JSONB DEFAULT '[]', -- Same format as target_providers
    rate_limits JSONB DEFAULT '{}', -- Per-rule rate limiting
    created_by CHAR(26) NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(project_id, rule_name)
);

-- Create indexes
CREATE INDEX idx_gateway_routing_rules_project ON gateway_routing_rules(project_id);
CREATE INDEX idx_gateway_routing_rules_enabled ON gateway_routing_rules(is_enabled);
CREATE INDEX idx_gateway_routing_rules_priority ON gateway_routing_rules(priority DESC);

-- Gateway Request Cache Table
-- For semantic caching (prepared for future caching implementation)
CREATE TABLE gateway_request_cache (
    id CHAR(26) PRIMARY KEY, -- ULID
    cache_key VARCHAR(64) NOT NULL UNIQUE, -- SHA-256 hash of normalized request
    project_id CHAR(26) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    model_name VARCHAR(100) NOT NULL,
    request_hash VARCHAR(64) NOT NULL, -- Hash of request content
    response_data JSONB NOT NULL, -- Cached response
    token_usage JSONB NOT NULL, -- {"input_tokens": 100, "output_tokens": 50, "total_tokens": 150}
    cost_usd DECIMAL(10,8) NOT NULL,
    hit_count INTEGER NOT NULL DEFAULT 0,
    last_accessed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_gateway_request_cache_key ON gateway_request_cache(cache_key);
CREATE INDEX idx_gateway_request_cache_project ON gateway_request_cache(project_id);
CREATE INDEX idx_gateway_request_cache_expires ON gateway_request_cache(expires_at);
CREATE INDEX idx_gateway_request_cache_accessed ON gateway_request_cache(last_accessed_at);

-- Update timestamps trigger function (reuse existing)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add update triggers
CREATE TRIGGER update_gateway_providers_updated_at BEFORE UPDATE ON gateway_providers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_gateway_models_updated_at BEFORE UPDATE ON gateway_models FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_gateway_provider_configs_updated_at BEFORE UPDATE ON gateway_provider_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_gateway_routing_rules_updated_at BEFORE UPDATE ON gateway_routing_rules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default providers
INSERT INTO gateway_providers (id, name, type, base_url, is_enabled, supported_features, rate_limits) VALUES
('01J9KXF0000000000000000001', 'OpenAI', 'openai', 'https://api.openai.com/v1', true, 
 '{"streaming": true, "functions": true, "embeddings": true, "vision": true}', 
 '{"requests_per_minute": 500, "tokens_per_minute": 150000}'),
('01J9KXF0000000000000000002', 'Anthropic', 'anthropic', 'https://api.anthropic.com/v1', false, 
 '{"streaming": true, "functions": true, "embeddings": false, "vision": true}', 
 '{"requests_per_minute": 1000, "tokens_per_minute": 200000}'),
('01J9KXF0000000000000000003', 'Cohere', 'cohere', 'https://api.cohere.ai/v1', false, 
 '{"streaming": true, "functions": false, "embeddings": true, "vision": false}', 
 '{"requests_per_minute": 10000, "tokens_per_minute": 1000000}'),
('01J9KXF0000000000000000004', 'Google AI', 'google', 'https://generativelanguage.googleapis.com/v1', false, 
 '{"streaming": true, "functions": true, "embeddings": true, "vision": true}', 
 '{"requests_per_minute": 1500, "tokens_per_minute": 32000}');

-- Insert default OpenAI models
INSERT INTO gateway_models (id, provider_id, model_name, display_name, input_cost_per_1k_tokens, output_cost_per_1k_tokens, max_context_tokens, supports_streaming, supports_functions, supports_vision, model_type, quality_score, speed_score) VALUES
-- GPT-4 models
('01J9KXF1000000000000000001', '01J9KXF0000000000000000001', 'gpt-4', 'GPT-4', 0.03, 0.06, 8192, true, true, false, 'text', 0.95, 0.60),
('01J9KXF1000000000000000002', '01J9KXF0000000000000000001', 'gpt-4-turbo', 'GPT-4 Turbo', 0.01, 0.03, 128000, true, true, true, 'text', 0.93, 0.75),
('01J9KXF1000000000000000003', '01J9KXF0000000000000000001', 'gpt-4o', 'GPT-4o', 0.005, 0.015, 128000, true, true, true, 'text', 0.92, 0.85),
('01J9KXF1000000000000000004', '01J9KXF0000000000000000001', 'gpt-4o-mini', 'GPT-4o Mini', 0.00015, 0.0006, 128000, true, true, true, 'text', 0.85, 0.95),
-- GPT-3.5 models
('01J9KXF1000000000000000005', '01J9KXF0000000000000000001', 'gpt-3.5-turbo', 'GPT-3.5 Turbo', 0.001, 0.002, 16384, true, true, false, 'text', 0.80, 0.90),
('01J9KXF1000000000000000006', '01J9KXF0000000000000000001', 'gpt-3.5-turbo-16k', 'GPT-3.5 Turbo 16K', 0.003, 0.004, 16384, true, true, false, 'text', 0.80, 0.85),
-- Embedding models
('01J9KXF1000000000000000007', '01J9KXF0000000000000000001', 'text-embedding-3-small', 'Text Embedding 3 Small', 0.00002, 0.0, 8191, false, false, false, 'embedding', 0.85, 0.95),
('01J9KXF1000000000000000008', '01J9KXF0000000000000000001', 'text-embedding-3-large', 'Text Embedding 3 Large', 0.00013, 0.0, 8191, false, false, false, 'embedding', 0.92, 0.85),
('01J9KXF1000000000000000009', '01J9KXF0000000000000000001', 'text-embedding-ada-002', 'Text Embedding Ada 002', 0.0001, 0.0, 8191, false, false, false, 'embedding', 0.80, 0.90);

-- Add comments for documentation
COMMENT ON TABLE gateway_providers IS 'AI provider definitions with capabilities and rate limits';
COMMENT ON TABLE gateway_models IS 'Model registry with pricing, capabilities, and performance metrics';
COMMENT ON TABLE gateway_provider_configs IS 'Project-scoped provider API keys and configurations';
COMMENT ON TABLE provider_health_metrics IS 'Provider availability and performance tracking';
COMMENT ON TABLE gateway_routing_rules IS 'Project-specific intelligent routing configuration';
COMMENT ON TABLE gateway_request_cache IS 'Semantic caching for AI responses (future implementation)';