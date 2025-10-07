-- AI Gateway PostgreSQL Down Migration
-- Drops all gateway-related tables in reverse order of dependencies

-- Drop triggers first
DROP TRIGGER IF EXISTS update_gateway_routing_rules_updated_at ON gateway_routing_rules;
DROP TRIGGER IF EXISTS update_gateway_provider_configs_updated_at ON gateway_provider_configs;
DROP TRIGGER IF EXISTS update_gateway_models_updated_at ON gateway_models;
DROP TRIGGER IF EXISTS update_gateway_providers_updated_at ON gateway_providers;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS gateway_request_cache;
DROP TABLE IF EXISTS gateway_routing_rules;
DROP TABLE IF EXISTS provider_health_metrics;
DROP TABLE IF EXISTS gateway_provider_configs;
DROP TABLE IF EXISTS gateway_models;
DROP TABLE IF EXISTS gateway_providers;

-- Drop the update function if it's not used elsewhere
-- Commented out as it might be used by other tables
-- DROP FUNCTION IF EXISTS update_updated_at_column();