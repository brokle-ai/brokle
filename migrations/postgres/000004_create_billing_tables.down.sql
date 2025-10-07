-- Drop billing tables and related objects
-- Migration: 000004_create_billing_tables.down.sql

-- Drop triggers
DROP TRIGGER IF EXISTS update_payment_methods_updated_at ON payment_methods;
DROP TRIGGER IF EXISTS update_invoices_updated_at ON invoices;
DROP TRIGGER IF EXISTS update_discount_rules_updated_at ON discount_rules;
DROP TRIGGER IF EXISTS update_usage_quotas_updated_at ON usage_quotas;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS invoice_line_items;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS payment_methods;
DROP TABLE IF EXISTS discount_rules;
DROP TABLE IF EXISTS billing_summaries;
DROP TABLE IF EXISTS billing_records;
DROP TABLE IF EXISTS usage_quotas;
DROP TABLE IF EXISTS usage_records;