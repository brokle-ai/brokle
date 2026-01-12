-- Rollback: create_usage_billing_tables
-- Created: 2026-01-05T01:13:12+05:30

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS usage_alerts;
DROP TABLE IF EXISTS usage_budgets;
DROP TABLE IF EXISTS organization_billings;
DROP TABLE IF EXISTS pricing_configs;
