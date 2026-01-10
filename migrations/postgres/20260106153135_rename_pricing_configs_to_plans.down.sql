-- Rollback: rename_pricing_configs_to_plans
-- Created: 2026-01-06T15:31:35+05:30

-- Revert comments
COMMENT ON TABLE plans IS 'Pricing configurations for billing plans';
COMMENT ON COLUMN organization_billings.plan_id IS 'Pricing configuration reference';

-- Revert indexes
ALTER INDEX idx_plans_default RENAME TO idx_pricing_configs_default;
ALTER INDEX idx_organization_billings_plan RENAME TO idx_organization_billings_pricing;

-- Revert foreign key first (table is still named 'plans' at this point)
ALTER TABLE organization_billings
  DROP CONSTRAINT organization_billings_plan_id_fkey,
  ADD CONSTRAINT organization_billings_pricing_config_id_fkey
    FOREIGN KEY (plan_id) REFERENCES plans(id);

-- Revert column rename
ALTER TABLE organization_billings
  RENAME COLUMN plan_id TO pricing_config_id;

-- Revert table rename
ALTER TABLE plans RENAME TO pricing_configs;
