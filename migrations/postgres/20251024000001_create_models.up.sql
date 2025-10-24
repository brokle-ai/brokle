-- Models table for pricing and model registry
CREATE TABLE IF NOT EXISTS models (
    id VARCHAR(255) PRIMARY KEY,
    project_id VARCHAR(255),

    -- Model identification
    model_name VARCHAR(255) NOT NULL,
    match_pattern VARCHAR(500) NOT NULL,
    provider VARCHAR(100) NOT NULL,

    -- Pricing (per 1k tokens)
    input_price DECIMAL(18,12),
    output_price DECIMAL(18,12),
    total_price DECIMAL(18,12),
    unit VARCHAR(50) DEFAULT 'TOKENS',

    -- Versioning
    start_date TIMESTAMP,
    is_deprecated BOOLEAN DEFAULT false,

    -- Additional config
    tokenizer_id VARCHAR(255),
    tokenizer_config JSONB,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT models_unique_key
        UNIQUE (project_id, model_name, start_date, unit)
);

-- Indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_models_model_name ON models(model_name);
CREATE INDEX IF NOT EXISTS idx_models_provider ON models(provider);
CREATE INDEX IF NOT EXISTS idx_models_project_id ON models(project_id);
CREATE INDEX IF NOT EXISTS idx_models_match_pattern ON models(match_pattern);

-- Sample data for common models
INSERT INTO models (id, project_id, model_name, match_pattern, provider, input_price, output_price, unit, start_date, is_deprecated) VALUES
    ('model_gpt4o_mini', NULL, 'gpt-4o-mini', '(?i)^(gpt-4o-mini)', 'openai', 0.00015, 0.0006, 'TOKENS', '2024-07-18', false),
    ('model_gpt4_turbo', NULL, 'gpt-4-turbo', '(?i)^(gpt-4-turbo)', 'openai', 0.01, 0.03, 'TOKENS', '2024-04-01', false),
    ('model_gpt35_turbo', NULL, 'gpt-3.5-turbo', '(?i)^(gpt-)(35|3.5)(-turbo)', 'openai', 0.0005, 0.0015, 'TOKENS', '2023-11-01', false),
    ('model_claude3_opus', NULL, 'claude-3-opus', '(?i)^(claude-3-opus)', 'anthropic', 0.015, 0.075, 'TOKENS', '2024-03-01', false),
    ('model_claude3_sonnet', NULL, 'claude-3-5-sonnet', '(?i)^(claude-3-5-sonnet)', 'anthropic', 0.003, 0.015, 'TOKENS', '2024-06-20', false),
    ('model_gemini_pro', NULL, 'gemini-pro', '(?i)^(gemini-pro)', 'google', 0.00125, 0.00375, 'TOKENS', '2024-02-01', false)
ON CONFLICT (project_id, model_name, start_date, unit) DO NOTHING;
