-- Remove gateway domain tables
-- Note: models table is KEPT for observability cost calculation

-- Drop gateway tables in reverse dependency order
DROP TABLE IF EXISTS gateway_request_cache CASCADE;
DROP TABLE IF EXISTS gateway_routing_rules CASCADE;
DROP TABLE IF EXISTS provider_health_metrics CASCADE;
DROP TABLE IF EXISTS gateway_provider_configs CASCADE;
DROP TABLE IF EXISTS gateway_models CASCADE;
DROP TABLE IF EXISTS gateway_providers CASCADE;

-- Note: The "models" table is preserved for observability domain cost calculation
-- It is used by internal/core/domain/observability/model_pricing.go
-- and internal/infrastructure/repository/observability/model_pricing_repository.go
