-- Drop billing tables and related objects
-- Migration: 000004_create_billing_tables.down.sql

-- Drop triggers
DROP TRIGGER IF EXISTS update_payment_methods_updated_at ON payment_methods;
DROP TRIGGER IF EXISTS update_invoices_updated_at ON invoices;
DROP TRIGGER IF EXISTS update_discount_rules_updated_at ON discount_rules;
DROP TRIGGER IF EXISTS update_usage_quotas_updated_at ON usage_quotas;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes explicitly (PostgreSQL automatically drops indexes when tables are dropped,
-- but being explicit ensures clean rollback)
DROP INDEX IF EXISTS idx_payment_methods_org_default_unique;
DROP INDEX IF EXISTS idx_payment_methods_external;
DROP INDEX IF EXISTS idx_payment_methods_org;

DROP INDEX IF EXISTS idx_invoice_line_items_model;
DROP INDEX IF EXISTS idx_invoice_line_items_provider;
DROP INDEX IF EXISTS idx_invoice_line_items_invoice;

DROP INDEX IF EXISTS idx_invoices_paid_at;
DROP INDEX IF EXISTS idx_invoices_issue_date;
DROP INDEX IF EXISTS idx_invoices_due_date;
DROP INDEX IF EXISTS idx_invoices_status;
DROP INDEX IF EXISTS idx_invoices_number;
DROP INDEX IF EXISTS idx_invoices_org_period;

DROP INDEX IF EXISTS idx_discount_rules_type;
DROP INDEX IF EXISTS idx_discount_rules_valid_until;
DROP INDEX IF EXISTS idx_discount_rules_valid_from;
DROP INDEX IF EXISTS idx_discount_rules_active_priority;
DROP INDEX IF EXISTS idx_discount_rules_org;

DROP INDEX IF EXISTS idx_billing_summaries_status;
DROP INDEX IF EXISTS idx_billing_summaries_period_start;
DROP INDEX IF EXISTS idx_billing_summaries_org_period;

DROP INDEX IF EXISTS idx_billing_records_transaction;
DROP INDEX IF EXISTS idx_billing_records_created;
DROP INDEX IF EXISTS idx_billing_records_status;
DROP INDEX IF EXISTS idx_billing_records_org_period;

DROP INDEX IF EXISTS idx_usage_records_processed;
DROP INDEX IF EXISTS idx_usage_records_model;
DROP INDEX IF EXISTS idx_usage_records_provider;
DROP INDEX IF EXISTS idx_usage_records_request;
DROP INDEX IF EXISTS idx_usage_records_org_tier;
DROP INDEX IF EXISTS idx_usage_records_org_created;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS invoice_line_items;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS payment_methods;
DROP TABLE IF EXISTS discount_rules;
DROP TABLE IF EXISTS billing_summaries;
DROP TABLE IF EXISTS billing_records;
DROP TABLE IF EXISTS usage_quotas;
DROP TABLE IF EXISTS usage_records;
