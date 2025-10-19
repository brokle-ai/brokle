-- Drop old observability tables (clean slate for ClickHouse-first architecture)

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS llm_quality_scores CASCADE;
DROP TABLE IF EXISTS llm_observations CASCADE;
DROP TABLE IF EXISTS llm_traces CASCADE;

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_llm_traces_updated_at ON llm_traces;
DROP TRIGGER IF EXISTS trigger_llm_observations_updated_at ON llm_observations;
DROP TRIGGER IF EXISTS trigger_llm_observations_calculate_latency ON llm_observations;
DROP TRIGGER IF EXISTS trigger_llm_observations_calculate_total_tokens ON llm_observations;
DROP TRIGGER IF EXISTS trigger_llm_observations_calculate_total_cost ON llm_observations;
DROP TRIGGER IF EXISTS trigger_llm_quality_scores_updated_at ON llm_quality_scores;

-- Drop functions
DROP FUNCTION IF EXISTS calculate_observation_latency();
DROP FUNCTION IF EXISTS calculate_total_tokens();
DROP FUNCTION IF EXISTS calculate_total_cost();

-- Note: update_updated_at_column() function is used by other tables, so keep it
