-- Migration: add_missing_provider_models
-- Created: 2025-11-27T21:51:32+05:30

-- ============================================================================
-- Add Missing AI Provider Models for Cost Analytics
-- ============================================================================
-- Purpose: Add popular OpenAI, Anthropic, and Google models missing from initial seed
-- Research Date: November 27, 2025
-- Sources:
--   - OpenAI: https://openai.com/api/pricing/
--   - Anthropic: https://docs.claude.com/en/docs/about-claude/pricing
--   - Google: https://ai.google.dev/gemini-api/docs/pricing
-- ============================================================================

-- ============================================================================
-- OPENAI MODELS
-- ============================================================================

-- GPT-4 (original base model - deprecated but still in use)
-- Pricing: $30/1M input, $60/1M output
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01KB32GPT4BASE0000000001', 'gpt-4', '^gpt-4$', '2023-03-14', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01KB32P0001', '01KB32GPT4BASE0000000001', 'input', 30.000000000000),
('01KB32P0002', '01KB32GPT4BASE0000000001', 'output', 60.000000000000);

-- GPT-4 Turbo (current stable production model)
-- Pricing: $10/1M input, $30/1M output
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01KB32GPT4TURBO00000002', 'gpt-4-turbo', '^gpt-4-turbo$', '2024-04-09', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01KB32P0003', '01KB32GPT4TURBO00000002', 'input', 10.000000000000),
('01KB32P0004', '01KB32GPT4TURBO00000002', 'output', 30.000000000000);

-- GPT-4 Turbo Preview (legacy name, same pricing as gpt-4-turbo)
-- Pricing: $10/1M input, $30/1M output
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01KB32GPT4TPREV00000003', 'gpt-4-turbo-preview', '^gpt-4-turbo-preview$', '2023-11-06', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01KB32P0005', '01KB32GPT4TPREV00000003', 'input', 10.000000000000),
('01KB32P0006', '01KB32GPT4TPREV00000003', 'output', 30.000000000000);

-- GPT-4 0613 snapshot (June 2023 snapshot - same pricing as base GPT-4)
-- Pricing: $30/1M input, $60/1M output
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01KB32GPT40613000000004', 'gpt-4-0613', '^gpt-4-0613$', '2023-06-13', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01KB32P0007', '01KB32GPT40613000000004', 'input', 30.000000000000),
('01KB32P0008', '01KB32GPT40613000000004', 'output', 60.000000000000);

-- GPT-4 0125-preview (January 2025 preview - turbo pricing)
-- Pricing: $10/1M input, $30/1M output
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01KB32GPT40125000000005', 'gpt-4-0125-preview', '^gpt-4-0125-preview$', '2024-01-25', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01KB32P0009', '01KB32GPT40125000000005', 'input', 10.000000000000),
('01KB32P0010', '01KB32GPT40125000000005', 'output', 30.000000000000);

-- GPT-4 1106-preview (November 2023 preview - turbo pricing)
-- Pricing: $10/1M input, $30/1M output
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01KB32GPT41106000000006', 'gpt-4-1106-preview', '^gpt-4-1106-preview$', '2023-11-06', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01KB32P0011', '01KB32GPT41106000000006', 'input', 10.000000000000),
('01KB32P0012', '01KB32GPT41106000000006', 'output', 30.000000000000);

-- ============================================================================
-- ANTHROPIC MODELS (Claude 3 Family)
-- ============================================================================

-- Claude 3 Opus (most capable model)
-- Pricing: $15/1M input, $75/1M output
-- Note: Also supports batch API (50% discount), cache creation ($18.75), cache reads ($1.50)
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01KB32CLAUDE3OPUS00007', 'claude-3-opus', '^claude-3-opus', '2024-03-04', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01KB32P0013', '01KB32CLAUDE3OPUS00007', 'input', 15.000000000000),
('01KB32P0014', '01KB32CLAUDE3OPUS00007', 'output', 75.000000000000),
('01KB32P0015', '01KB32CLAUDE3OPUS00007', 'cache_creation_input_tokens', 18.750000000000),
('01KB32P0016', '01KB32CLAUDE3OPUS00007', 'cache_read_input_tokens', 1.500000000000),
('01KB32P0017', '01KB32CLAUDE3OPUS00007', 'batch_input', 7.500000000000),
('01KB32P0018', '01KB32CLAUDE3OPUS00007', 'batch_output', 37.500000000000);

-- Claude 3 Haiku (fastest, most compact model)
-- Pricing: $0.25/1M input, $1.25/1M output
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01KB32CLAUDE3HAIKU0008', 'claude-3-haiku', '^claude-3-haiku', '2024-03-13', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01KB32P0019', '01KB32CLAUDE3HAIKU0008', 'input', 0.250000000000),
('01KB32P0020', '01KB32CLAUDE3HAIKU0008', 'output', 1.250000000000),
('01KB32P0021', '01KB32CLAUDE3HAIKU0008', 'cache_creation_input_tokens', 0.300000000000),
('01KB32P0022', '01KB32CLAUDE3HAIKU0008', 'cache_read_input_tokens', 0.025000000000),
('01KB32P0023', '01KB32CLAUDE3HAIKU0008', 'batch_input', 0.125000000000),
('01KB32P0024', '01KB32CLAUDE3HAIKU0008', 'batch_output', 0.625000000000);

-- Claude 3.5 Haiku (upgraded haiku model)
-- Pricing: $0.80/1M input, $4.00/1M output
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01KB32CLAUDE35HAIKU009', 'claude-3-5-haiku', '^claude-3-5-haiku', '2024-11-04', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01KB32P0025', '01KB32CLAUDE35HAIKU009', 'input', 0.800000000000),
('01KB32P0026', '01KB32CLAUDE35HAIKU009', 'output', 4.000000000000),
('01KB32P0027', '01KB32CLAUDE35HAIKU009', 'cache_creation_input_tokens', 1.000000000000),
('01KB32P0028', '01KB32CLAUDE35HAIKU009', 'cache_read_input_tokens', 0.080000000000),
('01KB32P0029', '01KB32CLAUDE35HAIKU009', 'batch_input', 0.400000000000),
('01KB32P0030', '01KB32CLAUDE35HAIKU009', 'batch_output', 2.000000000000);

-- ============================================================================
-- GOOGLE MODELS (Gemini Family)
-- ============================================================================

-- Gemini 2.5 Pro (standard context ≤200K tokens)
-- Pricing: $1.25/1M input, $10/1M output (up to 200K context)
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01KB32GEMINI25PRO0010', 'gemini-2.5-pro', '^gemini-2\\.5-pro', '2025-04-04', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01KB32P0031', '01KB32GEMINI25PRO0010', 'input', 1.250000000000),
('01KB32P0032', '01KB32GEMINI25PRO0010', 'output', 10.000000000000),
('01KB32P0033', '01KB32GEMINI25PRO0010', 'batch_input', 0.625000000000),
('01KB32P0034', '01KB32GEMINI25PRO0010', 'batch_output', 5.000000000000);

-- Gemini 2.5 Flash (cost-effective flash model)
-- Pricing: $0.30/1M input, $2.50/1M output
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01KB32GEMINI25FLASH011', 'gemini-2.5-flash', '^gemini-2\\.5-flash', '2025-04-04', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01KB32P0035', '01KB32GEMINI25FLASH011', 'input', 0.300000000000),
('01KB32P0036', '01KB32GEMINI25FLASH011', 'output', 2.500000000000),
('01KB32P0037', '01KB32GEMINI25FLASH011', 'batch_input', 0.150000000000),
('01KB32P0038', '01KB32GEMINI25FLASH011', 'batch_output', 1.250000000000);

-- Gemini 2.0 Flash (previous generation flash)
-- Pricing: $0.10/1M input, $0.40/1M output
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01KB32GEMINI20FLASH012', 'gemini-2.0-flash', '^gemini-2\\.0-flash', '2024-12-01', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01KB32P0039', '01KB32GEMINI20FLASH012', 'input', 0.100000000000),
('01KB32P0040', '01KB32GEMINI20FLASH012', 'output', 0.400000000000),
('01KB32P0041', '01KB32GEMINI20FLASH012', 'batch_input', 0.050000000000),
('01KB32P0042', '01KB32GEMINI20FLASH012', 'batch_output', 0.200000000000);

-- Gemini 1.5 Pro (widely used production model)
-- Pricing: $1.25/1M input, $5/1M output (up to 128K context)
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01KB32GEMINI15PRO0013', 'gemini-1.5-pro', '^gemini-1\\.5-pro', '2024-05-14', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01KB32P0043', '01KB32GEMINI15PRO0013', 'input', 1.250000000000),
('01KB32P0044', '01KB32GEMINI15PRO0013', 'output', 5.000000000000),
('01KB32P0045', '01KB32GEMINI15PRO0013', 'batch_input', 0.625000000000),
('01KB32P0046', '01KB32GEMINI15PRO0013', 'batch_output', 2.500000000000);

-- Gemini 1.5 Flash (cost-effective previous generation)
-- Pricing: $0.075/1M input, $0.30/1M output
INSERT INTO provider_models (id, model_name, match_pattern, start_date, unit) VALUES
('01KB32GEMINI15FLASH014', 'gemini-1.5-flash', '^gemini-1\\.5-flash', '2024-05-14', 'TOKENS');

INSERT INTO provider_prices (id, provider_model_id, usage_type, price) VALUES
('01KB32P0047', '01KB32GEMINI15FLASH014', 'input', 0.075000000000),
('01KB32P0048', '01KB32GEMINI15FLASH014', 'output', 0.300000000000),
('01KB32P0049', '01KB32GEMINI15FLASH014', 'batch_input', 0.037500000000),
('01KB32P0050', '01KB32GEMINI15FLASH014', 'batch_output', 0.150000000000);

-- Update table comment
COMMENT ON TABLE provider_models IS 'AI provider model definitions (OpenAI, Anthropic, Google) - includes legacy and current models for comprehensive cost tracking';
