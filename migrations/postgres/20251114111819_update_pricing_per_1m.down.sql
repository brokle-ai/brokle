-- Rollback pricing migration: revert to per-1K tokens

-- Step 1: Drop validation constraints
ALTER TABLE models DROP CONSTRAINT IF EXISTS models_valid_dates;
ALTER TABLE models DROP CONSTRAINT IF EXISTS models_positive_prices;

-- Step 2: Restore original unique constraint
ALTER TABLE models DROP CONSTRAINT IF EXISTS models_unique_key;
ALTER TABLE models
ADD CONSTRAINT models_unique_key
UNIQUE (project_id, model_name, start_date, unit);

-- Step 3: Drop temporal indexes
DROP INDEX IF EXISTS idx_models_active_pricing;
DROP INDEX IF EXISTS idx_models_historical_pricing;

-- Step 4: Drop new columns
ALTER TABLE models DROP COLUMN IF EXISTS input_price;
ALTER TABLE models DROP COLUMN IF EXISTS output_price;
ALTER TABLE models DROP COLUMN IF EXISTS end_date;
ALTER TABLE models DROP COLUMN IF EXISTS cache_write_multiplier;
ALTER TABLE models DROP COLUMN IF EXISTS cache_read_multiplier;
ALTER TABLE models DROP COLUMN IF EXISTS batch_discount_percentage;

-- Note: Original per-1K pricing columns (input_price, output_price) are preserved
-- No data loss on rollback
