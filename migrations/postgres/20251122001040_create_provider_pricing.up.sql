-- ============================================================================
-- AI Provider Models + Pricing Architecture
-- ============================================================================
-- Purpose: Track AI provider pricing (OpenAI, Anthropic, Google) for cost analytics
-- Design: ProviderModel (metadata) + ProviderPrice (flexible usage types)
-- NOT FOR: User billing - Brokle doesn't charge based on these prices
-- FOR: Cost visibility - "You spent $50 with OpenAI this month"
-- Zero Users: Clean break, no backward compatibility
-- ============================================================================

-- Step 1: Drop existing tables (clean slate)
DROP TABLE IF EXISTS prices CASCADE;
DROP TABLE IF EXISTS models CASCADE;
DROP TABLE IF EXISTS provider_prices CASCADE;
DROP TABLE IF EXISTS provider_models CASCADE;

-- Step 2: Create AI provider models table
CREATE TABLE provider_models (
    id CHAR(26) PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Project association (NULL = global, non-NULL = project-specific override)
    project_id CHAR(26),

    -- Model identification
    model_name VARCHAR(255) NOT NULL,
    match_pattern VARCHAR(500) NOT NULL,

    -- Temporal versioning (pricing changes over time)
    start_date TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Unit specification
    unit VARCHAR(50) NOT NULL DEFAULT 'TOKENS',

    -- Tokenization
    tokenizer_id VARCHAR(100),
    tokenizer_config JSONB,

    -- Constraints
    CONSTRAINT provider_models_unique_version UNIQUE(project_id, model_name, start_date, unit),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Indexes for fast lookups
CREATE INDEX idx_provider_models_name ON provider_models(model_name);
CREATE INDEX idx_provider_models_lookup ON provider_models(project_id, model_name, start_date DESC);
CREATE INDEX idx_provider_models_pattern ON provider_models(match_pattern);

-- Step 3: Create AI provider pricing table (KEY INNOVATION)
CREATE TABLE provider_prices (
    id CHAR(26) PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    provider_model_id CHAR(26) NOT NULL,
    project_id CHAR(26),

    -- Flexible usage type (arbitrary string - no schema changes needed)
    usage_type VARCHAR(100) NOT NULL,

    -- Price per 1 million units (what OpenAI/Anthropic charges end users)
    price DECIMAL(20,12) NOT NULL,

    -- Constraints
    CONSTRAINT provider_prices_unique_usage UNIQUE(provider_model_id, usage_type),
    FOREIGN KEY (provider_model_id) REFERENCES provider_models(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Index for fast pricing lookups
CREATE INDEX idx_provider_prices_lookup ON provider_prices(provider_model_id, usage_type);
CREATE INDEX idx_provider_prices_project ON provider_prices(project_id);

-- Step 4: Seed global provider pricing (2025 production pricing)

-- OpenAI GPT-4o
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01JG0001GPT4O0000000000001', 'gpt-4o', '^gpt-4o$', '2024-05-13', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01JGP000000000000000000001', '01JG0001GPT4O0000000000001', 'input', 2.50),
('01JGP000000000000000000002', '01JG0001GPT4O0000000000001', 'output', 10.00),
('01JGP000000000000000000003', '01JG0001GPT4O0000000000001', 'cache_read_input_tokens', 1.25),
('01JGP000000000000000000004', '01JG0001GPT4O0000000000001', 'batch_input', 1.25),
('01JGP000000000000000000005', '01JG0001GPT4O0000000000001', 'batch_output', 5.00);

-- OpenAI GPT-4o Mini
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01JG0002MINI00000000000002', 'gpt-4o-mini', '^gpt-4o-mini', '2024-07-18', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01JGP000000000000000000006', '01JG0002MINI00000000000002', 'input', 0.150),
('01JGP000000000000000000007', '01JG0002MINI00000000000002', 'output', 0.600),
('01JGP000000000000000000008', '01JG0002MINI00000000000002', 'cache_read_input_tokens', 0.075),
('01JGP000000000000000000009', '01JG0002MINI00000000000002', 'batch_input', 0.075),
('01JGP000000000000000000010', '01JG0002MINI00000000000002', 'batch_output', 0.300);

-- Anthropic Claude 3.7 Sonnet
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01JG0003CLAUDE370000000003', 'claude-3-7-sonnet', '^claude-3-7-sonnet', '2025-01-01', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01JGP000000000000000000011', '01JG0003CLAUDE370000000003', 'input', 3.00),
('01JGP000000000000000000012', '01JG0003CLAUDE370000000003', 'output', 15.00),
('01JGP000000000000000000013', '01JG0003CLAUDE370000000003', 'cache_read_input_tokens', 0.30),
('01JGP000000000000000000014', '01JG0003CLAUDE370000000003', 'cache_creation_input_tokens', 3.75),
('01JGP000000000000000000015', '01JG0003CLAUDE370000000003', 'batch_input', 1.50),
('01JGP000000000000000000016', '01JG0003CLAUDE370000000003', 'batch_output', 7.50);

-- Anthropic Claude 3.5 Sonnet (current production)
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01JG0004CLAUDE350000000004', 'claude-3-5-sonnet', '^claude-3-5-sonnet', '2024-10-22', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01JGP000000000000000000017', '01JG0004CLAUDE350000000004', 'input', 3.00),
('01JGP000000000000000000018', '01JG0004CLAUDE350000000004', 'output', 15.00),
('01JGP000000000000000000019', '01JG0004CLAUDE350000000004', 'cache_read_input_tokens', 0.30),
('01JGP000000000000000000020', '01JG0004CLAUDE350000000004', 'cache_creation_input_tokens', 3.75),
('01JGP000000000000000000021', '01JG0004CLAUDE350000000004', 'batch_input', 1.50),
('01JGP000000000000000000022', '01JG0004CLAUDE350000000004', 'batch_output', 7.50);

-- OpenAI GPT-4o Realtime (Audio support)
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01JG0005AUDIO0000000000005', 'gpt-4o-realtime', '^gpt-4o-realtime', '2024-10-01', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01JGP000000000000000000023', '01JG0005AUDIO0000000000005', 'input', 5.00),
('01JGP000000000000000000024', '01JG0005AUDIO0000000000005', 'output', 20.00),
('01JGP000000000000000000025', '01JG0005AUDIO0000000000005', 'audio_input', 100.00),
('01JGP000000000000000000026', '01JG0005AUDIO0000000000005', 'audio_output', 200.00);

-- OpenAI GPT-3.5 Turbo
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01JG0006GPT350000000000006', 'gpt-3.5-turbo', '^gpt-3\\.5-turbo', '2023-11-01', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01JGP000000000000000000027', '01JG0006GPT350000000000006', 'input', 0.500),
('01JGP000000000000000000028', '01JG0006GPT350000000000006', 'output', 1.500);

-- Add comments
COMMENT ON TABLE provider_models IS 'AI provider model definitions (OpenAI, Anthropic, Google) - for cost analytics, NOT user billing';
COMMENT ON TABLE provider_prices IS 'AI provider pricing rates (per 1M tokens) - used to calculate user spending with providers, NOT what Brokle charges';
COMMENT ON COLUMN provider_prices.usage_type IS 'Arbitrary string: input, output, cache_read_input_tokens, audio_input, video_input_per_second, etc. - add new types without schema changes';
COMMENT ON COLUMN provider_prices.price IS 'Provider rate per 1M tokens (e.g., what OpenAI charges) - Brokle uses this for cost visibility dashboards';
