-- Drop LLM observability tables in reverse order due to foreign key dependencies

-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_llm_observations_calculate_total_cost ON llm_observations;
DROP TRIGGER IF EXISTS trigger_llm_observations_calculate_total_tokens ON llm_observations;
DROP TRIGGER IF EXISTS trigger_llm_observations_calculate_latency ON llm_observations;
DROP TRIGGER IF EXISTS trigger_llm_quality_scores_updated_at ON llm_quality_scores;
DROP TRIGGER IF EXISTS trigger_llm_observations_updated_at ON llm_observations;
DROP TRIGGER IF EXISTS trigger_llm_traces_updated_at ON llm_traces;

-- Drop trigger functions (only the ones created in this migration)
DROP FUNCTION IF EXISTS calculate_total_cost();
DROP FUNCTION IF EXISTS calculate_total_tokens();
DROP FUNCTION IF EXISTS calculate_observation_latency();
-- NOTE: update_updated_at_column() is NOT dropped - it's a shared utility from initial_schema

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS llm_quality_scores;
DROP TABLE IF EXISTS llm_observations;
DROP TABLE IF EXISTS llm_traces;