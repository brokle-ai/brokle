-- Migration: rename_pricing_configs_to_plans
-- Created: 2026-01-06T15:31:35+05:30

-- Rename pricing_configs table to plans
ALTER TABLE pricing_configs RENAME TO plans;

-- Rename column first
ALTER TABLE organization_billings
  RENAME COLUMN pricing_config_id TO plan_id;

-- Update foreign key with new column name
ALTER TABLE organization_billings
  DROP CONSTRAINT organization_billings_pricing_config_id_fkey,
  ADD CONSTRAINT organization_billings_plan_id_fkey
    FOREIGN KEY (plan_id) REFERENCES plans(id);

-- Rename indexes
ALTER INDEX idx_pricing_configs_default RENAME TO idx_plans_default;
ALTER INDEX idx_organization_billings_pricing RENAME TO idx_organization_billings_plan;

-- Update comments
COMMENT ON TABLE plans IS 'Standard pricing tiers (Free, Pro, Enterprise)';
COMMENT ON COLUMN organization_billings.plan_id IS 'Base plan for pricing (can be overridden by contract)';
