-- Update model pricing to per-1M tokens (industry standard 2025)
-- Migrate from per-1K tokens to per-1M tokens
-- Add advanced pricing features (caching, batch discounts, temporal versioning)

-- Step 1: Add new columns for per-1M pricing
ALTER TABLE models
ADD COLUMN IF NOT EXISTS input_price DECIMAL(18,12),
ADD COLUMN IF NOT EXISTS output_price DECIMAL(18,12),
ADD COLUMN IF NOT EXISTS end_date TIMESTAMP,
ADD COLUMN IF NOT EXISTS cache_write_multiplier DECIMAL(5,2) DEFAULT 1.0,
ADD COLUMN IF NOT EXISTS cache_read_multiplier DECIMAL(5,2) DEFAULT 0.1,
ADD COLUMN IF NOT EXISTS batch_discount_percentage DECIMAL(5,2) DEFAULT 0.0;

-- Step 2: Update sample data with latest 2025 pricing (per-1M tokens)
-- OpenAI pricing (as of Nov 2025)
UPDATE models SET
    input_price = 0.150,  -- $0.15/1M tokens
    output_price = 0.600   -- $0.60/1M tokens
WHERE model_name = 'gpt-4o-mini';

UPDATE models SET
    input_price = 10.00,   -- $10/1M tokens
    output_price = 30.00    -- $30/1M tokens
WHERE model_name = 'gpt-4-turbo';

UPDATE models SET
    input_price = 0.500,   -- $0.50/1M tokens
    output_price = 1.500    -- $1.50/1M tokens
WHERE model_name = 'gpt-3.5-turbo';

-- Anthropic pricing (as of Nov 2025)
UPDATE models SET
    input_price = 15.00,    -- $15/1M tokens
    output_price = 75.00,   -- $75/1M tokens
    cache_write_multiplier = 1.25, -- 1.25x for 5-min cache writes
    cache_read_multiplier = 0.1    -- 0.1x for cache reads (90% savings)
WHERE model_name = 'claude-3-opus';

UPDATE models SET
    input_price = 3.00,     -- $3/1M tokens
    output_price = 15.00,   -- $15/1M tokens
    cache_write_multiplier = 1.25,
    cache_read_multiplier = 0.1
WHERE model_name = 'claude-3-5-sonnet';

-- Google pricing (as of Nov 2025)
UPDATE models SET
    input_price = 1.25,     -- $1.25/1M tokens
    output_price = 5.00     -- $5/1M tokens
WHERE model_name = 'gemini-pro';

-- Step 4: Create indexes for temporal queries
CREATE INDEX IF NOT EXISTS idx_models_active_pricing
ON models(model_name, start_date, end_date)
WHERE end_date IS NULL;

CREATE INDEX IF NOT EXISTS idx_models_historical_pricing
ON models(model_name, start_date, end_date);

-- Step 5: Update unique constraint to support temporal versioning
ALTER TABLE models DROP CONSTRAINT IF EXISTS models_unique_key;
ALTER TABLE models
ADD CONSTRAINT models_unique_key
UNIQUE (project_id, model_name, start_date, unit);

-- Step 6: Add validation constraints
ALTER TABLE models
ADD CONSTRAINT models_valid_dates
CHECK (end_date IS NULL OR end_date > start_date);

ALTER TABLE models
ADD CONSTRAINT models_positive_prices
CHECK (
    (input_price IS NULL OR input_price >= 0) AND
    (output_price IS NULL OR output_price >= 0)
);

-- Step 7: Add comment for documentation
COMMENT ON COLUMN models.input_price IS 'Input token cost per 1 million tokens (industry standard)';
COMMENT ON COLUMN models.output_price IS 'Output token cost per 1 million tokens (industry standard)';
COMMENT ON COLUMN models.end_date IS 'Pricing validity end date (NULL = current active pricing)';
COMMENT ON COLUMN models.cache_write_multiplier IS 'Cost multiplier for cache writes (e.g., 1.25x for Anthropic 5-min cache)';
COMMENT ON COLUMN models.cache_read_multiplier IS 'Cost multiplier for cache reads (e.g., 0.1x = 90% savings)';
COMMENT ON COLUMN models.batch_discount_percentage IS 'Batch processing discount percentage (e.g., 50.0 = 50% off)';
