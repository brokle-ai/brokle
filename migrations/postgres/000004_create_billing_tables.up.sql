-- Create billing tables for usage tracking and billing management
-- Migration: 000004_create_billing_tables.up.sql

-- Usage records table for individual billing entries
CREATE TABLE IF NOT EXISTS usage_records (
    id VARCHAR(26) PRIMARY KEY,
    organization_id VARCHAR(26) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    request_id VARCHAR(26) NOT NULL,
    provider_id VARCHAR(26) NOT NULL REFERENCES gateway_providers(id) ON DELETE CASCADE,
    model_id VARCHAR(26) NOT NULL REFERENCES gateway_models(id) ON DELETE CASCADE,
    request_type VARCHAR(50) NOT NULL,
    input_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    cost DECIMAL(12, 6) NOT NULL DEFAULT 0.0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    billing_tier VARCHAR(50) NOT NULL DEFAULT 'free',
    discounts DECIMAL(12, 6) NOT NULL DEFAULT 0.0,
    net_cost DECIMAL(12, 6) NOT NULL DEFAULT 0.0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,

    -- Indexes for performance
    INDEX idx_usage_records_org_created (organization_id, created_at DESC),
    INDEX idx_usage_records_org_tier (organization_id, billing_tier),
    INDEX idx_usage_records_request (request_id),
    INDEX idx_usage_records_provider (provider_id),
    INDEX idx_usage_records_model (model_id),
    INDEX idx_usage_records_processed (processed_at) WHERE processed_at IS NOT NULL
);

-- Usage quotas table for organization limits
CREATE TABLE IF NOT EXISTS usage_quotas (
    organization_id VARCHAR(26) PRIMARY KEY REFERENCES organizations(id) ON DELETE CASCADE,
    billing_tier VARCHAR(50) NOT NULL DEFAULT 'free',
    monthly_request_limit BIGINT NOT NULL DEFAULT 0,
    monthly_token_limit BIGINT NOT NULL DEFAULT 0,
    monthly_cost_limit DECIMAL(12, 6) NOT NULL DEFAULT 0.0,
    current_requests BIGINT NOT NULL DEFAULT 0,
    current_tokens BIGINT NOT NULL DEFAULT 0,
    current_cost DECIMAL(12, 6) NOT NULL DEFAULT 0.0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    reset_date TIMESTAMPTZ NOT NULL,
    last_updated TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Check constraints for limits
    CONSTRAINT chk_monthly_request_limit CHECK (monthly_request_limit >= 0),
    CONSTRAINT chk_monthly_token_limit CHECK (monthly_token_limit >= 0),
    CONSTRAINT chk_monthly_cost_limit CHECK (monthly_cost_limit >= 0),
    CONSTRAINT chk_current_requests CHECK (current_requests >= 0),
    CONSTRAINT chk_current_tokens CHECK (current_tokens >= 0),
    CONSTRAINT chk_current_cost CHECK (current_cost >= 0)
);

-- Billing records table for payment transactions
CREATE TABLE IF NOT EXISTS billing_records (
    id VARCHAR(26) PRIMARY KEY,
    organization_id VARCHAR(26) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    period VARCHAR(50) NOT NULL,
    amount DECIMAL(12, 6) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    transaction_id VARCHAR(255),
    payment_method VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,

    -- Indexes for performance
    INDEX idx_billing_records_org_period (organization_id, period),
    INDEX idx_billing_records_status (status),
    INDEX idx_billing_records_created (created_at DESC),
    INDEX idx_billing_records_transaction (transaction_id) WHERE transaction_id IS NOT NULL,

    -- Constraints
    CONSTRAINT chk_billing_amount CHECK (amount >= 0),
    CONSTRAINT chk_billing_status CHECK (status IN ('pending', 'paid', 'failed', 'cancelled', 'refunded'))
);

-- Billing summaries table for period summaries
CREATE TABLE IF NOT EXISTS billing_summaries (
    id VARCHAR(26) PRIMARY KEY,
    organization_id VARCHAR(26) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    period VARCHAR(50) NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    total_requests BIGINT NOT NULL DEFAULT 0,
    total_tokens BIGINT NOT NULL DEFAULT 0,
    total_cost DECIMAL(12, 6) NOT NULL DEFAULT 0.0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    provider_breakdown JSONB NOT NULL DEFAULT '{}',
    model_breakdown JSONB NOT NULL DEFAULT '{}',
    discounts DECIMAL(12, 6) NOT NULL DEFAULT 0.0,
    net_cost DECIMAL(12, 6) NOT NULL DEFAULT 0.0,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Unique constraint to prevent duplicate summaries
    UNIQUE (organization_id, period, period_start),
    
    -- Indexes for performance
    INDEX idx_billing_summaries_org_period (organization_id, period),
    INDEX idx_billing_summaries_period_start (period_start DESC),
    INDEX idx_billing_summaries_status (status),

    -- Constraints
    CONSTRAINT chk_billing_summary_cost CHECK (total_cost >= 0),
    CONSTRAINT chk_billing_summary_net_cost CHECK (net_cost >= 0),
    CONSTRAINT chk_billing_summary_dates CHECK (period_end > period_start),
    CONSTRAINT chk_billing_summary_status CHECK (status IN ('pending', 'finalized', 'invoiced', 'paid', 'cancelled'))
);

-- Discount rules table for promotional pricing
CREATE TABLE IF NOT EXISTS discount_rules (
    id VARCHAR(26) PRIMARY KEY,
    organization_id VARCHAR(26) REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    value DECIMAL(12, 6) NOT NULL,
    minimum_amount DECIMAL(12, 6) NOT NULL DEFAULT 0.0,
    maximum_discount DECIMAL(12, 6) NOT NULL DEFAULT 0.0,
    conditions JSONB NOT NULL DEFAULT '{}',
    valid_from TIMESTAMPTZ NOT NULL,
    valid_until TIMESTAMPTZ,
    usage_limit INTEGER,
    usage_count INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Indexes for performance
    INDEX idx_discount_rules_org (organization_id) WHERE organization_id IS NOT NULL,
    INDEX idx_discount_rules_active_priority (is_active, priority DESC) WHERE is_active = true,
    INDEX idx_discount_rules_valid_from (valid_from),
    INDEX idx_discount_rules_valid_until (valid_until) WHERE valid_until IS NOT NULL,
    INDEX idx_discount_rules_type (type),

    -- Constraints
    CONSTRAINT chk_discount_value CHECK (value >= 0),
    CONSTRAINT chk_discount_minimum_amount CHECK (minimum_amount >= 0),
    CONSTRAINT chk_discount_maximum_discount CHECK (maximum_discount >= 0),
    CONSTRAINT chk_discount_usage_limit CHECK (usage_limit IS NULL OR usage_limit > 0),
    CONSTRAINT chk_discount_usage_count CHECK (usage_count >= 0),
    CONSTRAINT chk_discount_type CHECK (type IN ('percentage', 'fixed', 'tiered')),
    CONSTRAINT chk_discount_dates CHECK (valid_until IS NULL OR valid_until > valid_from)
);

-- Invoices table for invoice management
CREATE TABLE IF NOT EXISTS invoices (
    id VARCHAR(26) PRIMARY KEY,
    invoice_number VARCHAR(255) NOT NULL UNIQUE,
    organization_id VARCHAR(26) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    organization_name VARCHAR(255) NOT NULL,
    billing_address JSONB,
    period VARCHAR(50) NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    issue_date TIMESTAMPTZ NOT NULL,
    due_date TIMESTAMPTZ NOT NULL,
    subtotal DECIMAL(12, 6) NOT NULL DEFAULT 0.0,
    tax_amount DECIMAL(12, 6) NOT NULL DEFAULT 0.0,
    discount_amount DECIMAL(12, 6) NOT NULL DEFAULT 0.0,
    total_amount DECIMAL(12, 6) NOT NULL DEFAULT 0.0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    payment_terms VARCHAR(255),
    notes TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMPTZ,

    -- Indexes for performance
    INDEX idx_invoices_org_period (organization_id, period),
    INDEX idx_invoices_number (invoice_number),
    INDEX idx_invoices_status (status),
    INDEX idx_invoices_due_date (due_date),
    INDEX idx_invoices_issue_date (issue_date DESC),
    INDEX idx_invoices_paid_at (paid_at) WHERE paid_at IS NOT NULL,

    -- Constraints
    CONSTRAINT chk_invoice_subtotal CHECK (subtotal >= 0),
    CONSTRAINT chk_invoice_tax_amount CHECK (tax_amount >= 0),
    CONSTRAINT chk_invoice_discount_amount CHECK (discount_amount >= 0),
    CONSTRAINT chk_invoice_total_amount CHECK (total_amount >= 0),
    CONSTRAINT chk_invoice_dates CHECK (period_end > period_start AND due_date >= issue_date),
    CONSTRAINT chk_invoice_status CHECK (status IN ('draft', 'sent', 'paid', 'overdue', 'cancelled', 'refunded'))
);

-- Invoice line items table for detailed billing breakdown
CREATE TABLE IF NOT EXISTS invoice_line_items (
    id VARCHAR(26) PRIMARY KEY,
    invoice_id VARCHAR(26) NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    quantity DECIMAL(12, 6) NOT NULL DEFAULT 1.0,
    unit_price DECIMAL(12, 6) NOT NULL DEFAULT 0.0,
    amount DECIMAL(12, 6) NOT NULL DEFAULT 0.0,
    provider_id VARCHAR(26) REFERENCES gateway_providers(id) ON DELETE SET NULL,
    provider_name VARCHAR(255),
    model_id VARCHAR(26) REFERENCES gateway_models(id) ON DELETE SET NULL,
    model_name VARCHAR(255),
    request_type VARCHAR(50),
    tokens BIGINT,
    requests BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Indexes for performance
    INDEX idx_invoice_line_items_invoice (invoice_id),
    INDEX idx_invoice_line_items_provider (provider_id) WHERE provider_id IS NOT NULL,
    INDEX idx_invoice_line_items_model (model_id) WHERE model_id IS NOT NULL,

    -- Constraints
    CONSTRAINT chk_line_item_quantity CHECK (quantity > 0),
    CONSTRAINT chk_line_item_unit_price CHECK (unit_price >= 0),
    CONSTRAINT chk_line_item_amount CHECK (amount >= 0),
    CONSTRAINT chk_line_item_tokens CHECK (tokens IS NULL OR tokens >= 0),
    CONSTRAINT chk_line_item_requests CHECK (requests IS NULL OR requests >= 0)
);

-- Payment methods table for organization payment information
CREATE TABLE IF NOT EXISTS payment_methods (
    id VARCHAR(26) PRIMARY KEY,
    organization_id VARCHAR(26) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL DEFAULT 'card',
    provider VARCHAR(100) NOT NULL DEFAULT 'stripe',
    external_id VARCHAR(255) NOT NULL,
    last_4 VARCHAR(4),
    expiry_month INTEGER,
    expiry_year INTEGER,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Unique constraint to prevent multiple defaults per organization
    UNIQUE (organization_id, is_default) WHERE is_default = true,

    -- Indexes for performance
    INDEX idx_payment_methods_org (organization_id),
    INDEX idx_payment_methods_external (provider, external_id),
    INDEX idx_payment_methods_default (organization_id, is_default) WHERE is_default = true,

    -- Constraints
    CONSTRAINT chk_payment_method_type CHECK (type IN ('card', 'bank_transfer', 'paypal', 'other')),
    CONSTRAINT chk_payment_method_last_4 CHECK (last_4 IS NULL OR length(last_4) = 4),
    CONSTRAINT chk_payment_method_expiry_month CHECK (expiry_month IS NULL OR (expiry_month >= 1 AND expiry_month <= 12)),
    CONSTRAINT chk_payment_method_expiry_year CHECK (expiry_year IS NULL OR expiry_year >= date_part('year', NOW()))
);

-- Add trigger to automatically update updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_usage_quotas_updated_at 
    BEFORE UPDATE ON usage_quotas 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_discount_rules_updated_at 
    BEFORE UPDATE ON discount_rules 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_invoices_updated_at 
    BEFORE UPDATE ON invoices 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payment_methods_updated_at 
    BEFORE UPDATE ON payment_methods 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE usage_records IS 'Individual usage records for billing and analytics';
COMMENT ON TABLE usage_quotas IS 'Organization usage quotas and limits';
COMMENT ON TABLE billing_records IS 'Payment transaction records';
COMMENT ON TABLE billing_summaries IS 'Billing period summaries';
COMMENT ON TABLE discount_rules IS 'Promotional discount rules and conditions';
COMMENT ON TABLE invoices IS 'Generated invoices for organizations';
COMMENT ON TABLE invoice_line_items IS 'Detailed line items for invoices';
COMMENT ON TABLE payment_methods IS 'Organization payment method information';

COMMENT ON COLUMN usage_records.request_id IS 'Reference to the original gateway request';
COMMENT ON COLUMN usage_records.billing_tier IS 'Billing tier at time of request (free, pro, business, enterprise)';
COMMENT ON COLUMN usage_quotas.reset_date IS 'When monthly quotas reset (typically first of month)';
COMMENT ON COLUMN discount_rules.conditions IS 'JSON conditions for discount application';
COMMENT ON COLUMN invoices.billing_address IS 'JSON billing address information';
COMMENT ON COLUMN invoices.metadata IS 'Additional invoice metadata (usage stats, etc.)';